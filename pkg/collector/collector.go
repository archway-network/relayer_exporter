package collector

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/config"
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const (
	successStatus                 = "success"
	errorStatus                   = "error"
	clientExpiryMetricName        = "cosmos_ibc_client_expiry"
	walletBalanceMetricName       = "cosmos_wallet_balance"
	channelStuckPacketsMetricName = "cosmos_ibc_stuck_packets"
)

var (
	clientExpiry = prometheus.NewDesc(
		clientExpiryMetricName,
		"Returns light client expiry in unixtime.",
		[]string{
			"host_chain_id",
			"client_id",
			"target_chain_id",
			"discord_ids",
			"status",
		},
		nil,
	)
	channelStuckPackets = prometheus.NewDesc(
		channelStuckPacketsMetricName,
		"Returns stuck packets for a channel.",
		[]string{
			"src_channel_id",
			"dst_channel_id",
			"src_chain_id",
			"dst_chain_id",
			"discord_ids",
			"status",
		},
		nil,
	)
	walletBalance = prometheus.NewDesc(
		walletBalanceMetricName,
		"Returns wallet balance for an address on a chain.",
		[]string{"account", "chain_id", "denom", "status"}, nil,
	)
)

type IBCCollector struct {
	RPCs  *map[string]config.RPC
	Paths []*config.IBCData
}

type WalletBalanceCollector struct {
	RPCs     *map[string]config.RPC
	Accounts []config.Account
}

func (cc IBCCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clientExpiry
	ch <- channelStuckPackets
}

func (cc IBCCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug(
		"Start collecting",
		zap.String(
			"metrics",
			fmt.Sprintf("%s, %s", clientExpiryMetricName, channelStuckPacketsMetricName),
		),
	)

	var wg sync.WaitGroup

	for _, p := range cc.Paths {
		wg.Add(1)

		go func(path *config.IBCData) {
			defer wg.Done()

			discordIDs := getDiscordIDs(path.Operators)

			// Client info
			ci, err := ibc.GetClientsInfo(path, cc.RPCs)
			status := successStatus

			if err != nil {
				status = errorStatus

				log.Error(err.Error())
			}

			ch <- prometheus.MustNewConstMetric(
				clientExpiry,
				prometheus.GaugeValue,
				float64(ci.ChainAClientExpiration.Unix()),
				[]string{
					(*cc.RPCs)[path.Chain1.ChainName].ChainID,
					path.Chain1.ClientID,
					(*cc.RPCs)[path.Chain2.ChainName].ChainID,
					discordIDs,
					status,
				}...,
			)

			ch <- prometheus.MustNewConstMetric(
				clientExpiry,
				prometheus.GaugeValue,
				float64(ci.ChainBClientExpiration.Unix()),
				[]string{
					(*cc.RPCs)[path.Chain2.ChainName].ChainID,
					path.Chain2.ClientID,
					(*cc.RPCs)[path.Chain1.ChainName].ChainID,
					discordIDs,
					status,
				}...,
			)

			// Stuck packets
			status = successStatus

			stuckPackets, err := ibc.GetChannelsInfo(path, cc.RPCs)
			if err != nil {
				status = errorStatus

				log.Error(err.Error())
			}

			if !reflect.DeepEqual(stuckPackets, ibc.ChannelsInfo{}) {
				for _, sp := range stuckPackets.Channels {
					ch <- prometheus.MustNewConstMetric(
						channelStuckPackets,
						prometheus.GaugeValue,
						float64(sp.StuckPackets.Source),
						[]string{
							sp.Source,
							sp.Destination,
							(*cc.RPCs)[path.Chain1.ChainName].ChainID,
							(*cc.RPCs)[path.Chain2.ChainName].ChainID,
							discordIDs,
							status,
						}...,
					)

					ch <- prometheus.MustNewConstMetric(
						channelStuckPackets,
						prometheus.GaugeValue,
						float64(sp.StuckPackets.Destination),
						[]string{
							sp.Destination,
							sp.Source,
							(*cc.RPCs)[path.Chain2.ChainName].ChainID,
							(*cc.RPCs)[path.Chain1.ChainName].ChainID,
							discordIDs,
							status,
						}...,
					)
				}
			}
		}(p)
	}

	wg.Wait()

	log.Debug("Stop collecting", zap.String("metric", clientExpiryMetricName))
}

func (wb WalletBalanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- walletBalance
}

func (wb WalletBalanceCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start collecting", zap.String("metric", walletBalanceMetricName))

	var wg sync.WaitGroup

	for _, a := range wb.Accounts {
		wg.Add(1)

		go func(account config.Account) {
			defer wg.Done()

			balance := 0.0
			status := successStatus

			err := account.GetBalance(wb.RPCs)
			if err != nil {
				status = errorStatus

				log.Error(err.Error(), zap.Any("account", account))
			} else {
				// Convert to a big float to get a float64 for metrics
				balance, _ = big.NewFloat(0.0).SetInt(account.Balance.BigInt()).Float64()
			}

			ch <- prometheus.MustNewConstMetric(
				walletBalance,
				prometheus.GaugeValue,
				balance,
				[]string{account.Address, (*wb.RPCs)[account.ChainName].ChainID, account.Denom, status}...,
			)
		}(a)
	}

	wg.Wait()

	log.Debug("Stop collecting", zap.String("metric", walletBalanceMetricName))
}

func getDiscordIDs(ops []config.Operator) string {
	var ids []string
	for _, op := range ops {
		ids = append(ids, op.Discord.ID)
	}

	return strings.Join(ids, ",")
}

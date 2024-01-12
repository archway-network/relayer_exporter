package collector

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/config"
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const (
	clientExpiryMetricName        = "cosmos_ibc_client_expiry"
	channelStuckPacketsMetricName = "cosmos_ibc_stuck_packets"
)

var (
	clientExpiry = prometheus.NewDesc(
		clientExpiryMetricName,
		"Returns light client expiry in unixtime.",
		[]string{
			"host_chain_id",
			"target_chain_id",
			"src_chain_name",
			"dst_chain_name",
			"client_id",
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
			"src_chain_name",
			"dst_chain_name",
			"discord_ids",
			"status",
		},
		nil,
	)
)

type IBCCollector struct {
	RPCs  *map[string]config.RPC
	Paths []*config.IBCData
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
			ci, err := ibc.GetClientsInfo(ctx, path, cc.RPCs)
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
					(*cc.RPCs)[path.Chain2.ChainName].ChainID,
					path.Chain1.ChainName,
					path.Chain2.ChainName,
					path.Chain1.ClientID,
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
					(*cc.RPCs)[path.Chain1.ChainName].ChainID,
					path.Chain2.ChainName,
					path.Chain1.ChainName,
					path.Chain2.ClientID,
					discordIDs,
					status,
				}...,
			)

			// Stuck packets
			status = successStatus

			stuckPackets, err := ibc.GetChannelsInfo(ctx, path, cc.RPCs)
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
							path.Chain1.ChainName,
							path.Chain2.ChainName,
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
							path.Chain2.ChainName,
							path.Chain1.ChainName,
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

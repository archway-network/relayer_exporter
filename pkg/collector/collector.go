package collector

import (
	"math/big"
	"sync"

	"github.com/archway-network/relayer_exporter/pkg/account"
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/cosmos/relayer/v2/relayer"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	clientExpiryMetricName  = "cosmos_ibc_client_expiry"
	walletBalanceMetricName = "cosmos_wallet_balance"
)

var (
	clientExpiry = prometheus.NewDesc(
		clientExpiryMetricName,
		"Returns light client expiry in unixtime.",
		[]string{"host_chain_id", "client_id", "target_chain_id"}, nil,
	)
	walletBalance = prometheus.NewDesc(
		walletBalanceMetricName,
		"Returns wallet balance for an address on a chain.",
		[]string{"account", "chain_id", "denom", "status"}, nil,
	)
)

type IBCClientsCollector struct {
	RPCs  map[string]string
	Paths []*relayer.IBCdata
}

type WalletBalanceCollector struct {
	RPCs     map[string]string
	Accounts []account.Account
}

func (cc IBCClientsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- clientExpiry
}

func (cc IBCClientsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start collecting", zap.String("metric", clientExpiryMetricName))

	clients := ibc.GetClientsInfos(cc.Paths, cc.RPCs)

	for _, c := range clients {
		ch <- prometheus.MustNewConstMetric(
			clientExpiry,
			prometheus.GaugeValue,
			float64(c.ChainAClientExpiration.Unix()),
			[]string{c.ChainA.ChainID(), c.ChainA.ClientID(), c.ChainB.ChainID()}...,
		)

		ch <- prometheus.MustNewConstMetric(
			clientExpiry,
			prometheus.GaugeValue,
			float64(c.ChainBClientExpiration.Unix()),
			[]string{c.ChainB.ChainID(), c.ChainB.ClientID(), c.ChainA.ChainID()}...,
		)
	}

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

		go func(account account.Account) {
			defer wg.Done()

			balance := 0.0

			err := account.GetBalance(wb.RPCs)
			if err != nil {
				log.Error(err.Error(), zap.Any("account", account))
			} else {
				// Convert to a big float to get a float64 for metrics
				balance, _ = big.NewFloat(0.0).SetInt(account.Balance.BigInt()).Float64()
			}

			ch <- prometheus.MustNewConstMetric(
				walletBalance,
				prometheus.GaugeValue,
				balance,
				[]string{account.Address, account.ChainID, account.Denom, account.Status}...,
			)
		}(a)
	}

	wg.Wait()

	log.Debug("Stop collecting", zap.String("metric", walletBalanceMetricName))
}

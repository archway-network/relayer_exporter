package collector

import (
	"context"
	"math/big"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	"github.com/archway-network/relayer_exporter/pkg/config"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const (
	walletBalanceMetricName = "cosmos_wallet_balance"
)

var walletBalance = prometheus.NewDesc(
	walletBalanceMetricName,
	"Returns wallet balance for an address on a chain.",
	[]string{"account", "chain_id", "denom", "status", "tags"}, nil,
)

type WalletBalanceCollector struct {
	RPCs     *map[string]config.RPC
	Accounts []*config.Account
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

			err := getBalance(ctx, &account, wb.RPCs)
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
				[]string{account.Address, (*wb.RPCs)[account.ChainName].ChainID, account.Denom, status, strings.Join(account.Tags, ",")}...,
			)
		}(*a)
	}

	wg.Wait()

	log.Debug("Stop collecting", zap.String("metric", walletBalanceMetricName))
}

func getBalance(ctx context.Context, a *config.Account, rpcs *map[string]config.RPC) error {
	chain, err := chain.PrepChain(ctx, chain.Info{
		ChainID: (*rpcs)[a.ChainName].ChainID,
		RPCAddr: (*rpcs)[a.ChainName].URL,
		Timeout: (*rpcs)[a.ChainName].Timeout,
	})
	if err != nil {
		return err
	}

	coins, err := chain.ChainProvider.QueryBalanceWithAddress(ctx, a.Address)
	if err != nil {
		return err
	}

	a.Balance = coins.AmountOf(a.Denom)

	return nil
}

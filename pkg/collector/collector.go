package collector

import (
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/archway-network/relayer_exporter/pkg/relayer"
	gorelayer "github.com/cosmos/relayer/v2/relayer"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	clientExpiryMetricName    = "cosmos_ibc_client_expiry"
	configuredChainMetricName = "cosmos_relayer_configured_chain"
	relayerUpMetricName       = "cosmos_relayer_up"
)

var (
	up = prometheus.NewDesc(
		relayerUpMetricName,
		"Was talking to relayer successful.",
		nil, nil,
	)

	clientExpiry = prometheus.NewDesc(
		clientExpiryMetricName,
		"Returns light client expiry in unixtime.",
		[]string{"chain_id", "client_id", "path"}, nil,
	)

	configuredChain = prometheus.NewDesc(
		configuredChainMetricName,
		"Returns configured chain.",
		[]string{"chain_id"}, nil,
	)
)

type IBCClientsCollector struct {
	Config ibc.Config
	Paths  []*gorelayer.IBCdata
}

func (cc IBCClientsCollector) Describe(_ chan<- *prometheus.Desc) {}

func (cc IBCClientsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start collecting", zap.String("metric", clientExpiryMetricName))

	clients := cc.Config.GetClientsInfos(cc.Paths)

	for _, c := range clients {
		path := c.PathName()

		ch <- prometheus.MustNewConstMetric(
			clientExpiry,
			prometheus.GaugeValue,
			float64(c.ChainAClientExpiration.Unix()),
			[]string{c.ChainA.ChainID(), c.ChainA.ClientID(), path}...,
		)

		ch <- prometheus.MustNewConstMetric(
			clientExpiry,
			prometheus.GaugeValue,
			float64(c.ChainBClientExpiration.Unix()),
			[]string{c.ChainB.ChainID(), c.ChainB.ClientID(), path}...,
		)
	}

	log.Debug("Stop collecting", zap.String("metric", clientExpiryMetricName))
}

type RelayerCollector struct {
	Rly string
}

func (rc RelayerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
}

func (rc RelayerCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start collecting", zap.String("metric", configuredChainMetricName))

	chains, err := relayer.GetConfiguredChains(rc.Rly)
	if err != nil {
		log.Error(err.Error())
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0)

		return
	}

	for _, chain := range chains {
		ch <- prometheus.MustNewConstMetric(
			configuredChain,
			prometheus.GaugeValue,
			1,
			chain,
		)
	}

	log.Debug("Stop collecting", zap.String("metric", configuredChainMetricName))

	log.Debug("Start collecting", zap.String("metric", clientExpiryMetricName))

	clients, err := relayer.GetClients(rc.Rly)
	if err != nil {
		log.Error(err.Error())
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0)

		return
	}

	for _, c := range clients {
		ch <- prometheus.MustNewConstMetric(
			clientExpiry,
			prometheus.GaugeValue,
			float64(c.ExpiresAt.Unix()),
			[]string{c.ChainID, c.Path}...,
		)
	}

	log.Debug("Stop collecting", zap.String("metric", clientExpiryMetricName))

	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1)
}

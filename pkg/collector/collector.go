package collector

import (
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/cosmos/relayer/v2/relayer"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	clientExpiryMetricName    = "cosmos_ibc_client_expiry"
	configuredChainMetricName = "cosmos_relayer_configured_chain"
	relayerUpMetricName       = "cosmos_relayer_up"
)

var clientExpiry = prometheus.NewDesc(
	clientExpiryMetricName,
	"Returns light client expiry in unixtime.",
	[]string{"chain_id", "client_id", "path"}, nil,
)

type IBCClientsCollector struct {
	RPCs  map[string]string
	Paths []*relayer.IBCdata
}

func (cc IBCClientsCollector) Describe(_ chan<- *prometheus.Desc) {}

func (cc IBCClientsCollector) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start collecting", zap.String("metric", clientExpiryMetricName))

	clients := ibc.GetClientsInfos(cc.Paths, cc.RPCs)

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

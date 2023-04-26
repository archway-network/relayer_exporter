package collector

import (
	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/archway-network/relayer_exporter/pkg/relayer"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	up = prometheus.NewDesc(
		"cosmos_relayer_up",
		"Was talking to relayer successful.",
		nil, nil,
	)

	clientExpiry = prometheus.NewDesc(
		"cosmos_relayer_client_expiry",
		"Returns light client expiry in unixtime.",
		[]string{"chain_id", "path"}, nil,
	)
)

type RelayerCollector struct {
	Rly string
}

func (rc RelayerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
}

func (rc RelayerCollector) Collect(ch chan<- prometheus.Metric) {
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

	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1)
}

package collector

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/config"
	"github.com/archway-network/relayer_exporter/pkg/ibc"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const (
	clientExpiryMetricName            = "cosmos_ibc_client_expiry"
	channelStuckPacketsMetricName     = "cosmos_ibc_stuck_packets"
	channelNewAckSinceStuckMetricName = "cosmos_ibc_new_ack_since_stuck"
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
		"Returns number of stuck packets for a channel.",
		[]string{
			"src_channel_id",
			"dst_channel_id",
			"src_chain_id",
			"dst_chain_id",
			"src_chain_height",
			"dst_chain_height",
			"src_chain_name",
			"dst_chain_name",
			"discord_ids",
			"status",
		},
		nil,
	)
	channelNewAckSinceStuck = prometheus.NewDesc(
		channelNewAckSinceStuckMetricName,
		"Returns 1 if new IBC ack was observed since last stuck packet detection, else returns 0.",
		[]string{
			"src_channel_id",
			"dst_channel_id",
			"src_chain_id",
			"dst_chain_id",
			"src_chain_height",
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
	ch <- channelNewAckSinceStuck
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

			channelsInfo, err := ibc.GetChannelsInfo(ctx, path, cc.RPCs)
			if err != nil {
				status = errorStatus

				log.Error(err.Error())
			}

			if !reflect.DeepEqual(channelsInfo, ibc.ChannelsInfo{}) {
				for _, chInfo := range channelsInfo.Channels {
					ch <- prometheus.MustNewConstMetric(
						channelStuckPackets,
						prometheus.GaugeValue,
						float64(len(chInfo.StuckPackets.Src)),
						[]string{
							chInfo.Source,
							chInfo.Destination,
							(*cc.RPCs)[path.Chain1.ChainName].ChainID,
							(*cc.RPCs)[path.Chain2.ChainName].ChainID,
							strconv.FormatInt(chInfo.StuckPackets.SrcHeight, 10),
							strconv.FormatInt(chInfo.StuckPackets.DstHeight, 10),
							path.Chain1.ChainName,
							path.Chain2.ChainName,
							discordIDs,
							status,
						}...,
					)

					ch <- prometheus.MustNewConstMetric(
						channelStuckPackets,
						prometheus.GaugeValue,
						float64(len(chInfo.StuckPackets.Dst)),
						[]string{
							chInfo.Destination,
							chInfo.Source,
							(*cc.RPCs)[path.Chain2.ChainName].ChainID,
							(*cc.RPCs)[path.Chain1.ChainName].ChainID,
							strconv.FormatInt(chInfo.StuckPackets.SrcHeight, 10),
							strconv.FormatInt(chInfo.StuckPackets.DstHeight, 10),
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

func (cc IBCCollector) MaybeStartNewAckPocessor(ch chan<- prometheus.Metric, path *config.IBCData, channelsInfo *ibc.ChannelsInfo, rpcs *map[string]config.RPC) {

}

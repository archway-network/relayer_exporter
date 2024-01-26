package collector

import (
	"context"
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
			"src_chain_id",
			"dst_chain_id",
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
		"Returns block height of new observed IBC Ack since last stuck packet detection, else returns 0.",
		[]string{
			"src_channel_id",
			"dst_channel_id",
			"src_chain_id",
			"dst_chain_id",
			"src_chain_name",
			"dst_chain_name",
			"status",
		},
		nil,
	)
)

type IBCCollector struct {
	RPCs          *map[string]config.RPC
	Paths         []*config.IBCData
	AckProcessors map[string]AckProcessor // map[ChainName]AckProcessor
}

type AckProcessor struct {
	ChainID      string
	ChannelID    string
	StartHeight  int64
	NewAckHeight chan uint64
}

func (cc IBCCollector) Describe(metricDesc chan<- *prometheus.Desc) {
	metricDesc <- clientExpiry
	metricDesc <- channelStuckPackets
	metricDesc <- channelNewAckSinceStuck
}

func (cc IBCCollector) Collect(metric chan<- prometheus.Metric) {
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

			// cosmos_ibc_client_expiry metric collection
			ci, err := ibc.GetClientsInfo(ctx, path, cc.RPCs)
			status := successStatus

			if err != nil {
				status = errorStatus

				log.Error(err.Error())
			}

			metric <- prometheus.MustNewConstMetric(
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

			metric <- prometheus.MustNewConstMetric(
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

			// cosmos_ibc_stuck_packets metric collection
			status = successStatus

			channelsInfo, err := ibc.GetChannelsInfo(ctx, path, cc.RPCs)
			if err != nil {
				status = errorStatus

				log.Error(err.Error())
			}

			if !reflect.DeepEqual(channelsInfo, ibc.ChannelsInfo{}) {
				for _, chInfo := range channelsInfo.Channels {
					metric <- prometheus.MustNewConstMetric(
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

					metric <- prometheus.MustNewConstMetric(
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

					// cosmos_ibc_new_ack_since_stuck metric collection
					err := cc.SetNewAckSinceStuckMetric(ctx, chInfo, path, metric)

					if err != nil {
						log.Error(err.Error())
					}
				}
			}
		}(p)
	}

	wg.Wait()

	log.Debug("Stop collecting", zap.String("metric", clientExpiryMetricName))
}

func (cc IBCCollector) SetNewAckSinceStuckMetric(
	ctx context.Context,
	chInfo ibc.Channel,
	path *config.IBCData,
	metric chan<- prometheus.Metric) error {

	// Note: cosmos sdk block height is uint64 but prometheus metric type expects float64
	var newAckHeight uint64

	// Hash key to search for AckProcessor in cc (<chainName_channelID>)
	srcHashKey := fmt.Sprintf("%s_%s", path.Chain1.ChainName, chInfo.Source)
	dstHashKey := fmt.Sprintf("%s_%s", path.Chain2.ChainName, chInfo.Destination)

	// Publish source chain metric
	// Start Src chain AckProcessors if stuck packets exist and no processor is running
	if _, ok := cc.AckProcessors[srcHashKey]; len(chInfo.StuckPackets.Src) > 0 && !ok {
		log.Info("Creating packet processor", zap.String("ChainName", path.Chain1.ChainName), zap.String("ChannelID", chInfo.Source))
		cc.AckProcessors[srcHashKey] = AckProcessor{
			ChainID:      (*cc.RPCs)[path.Chain1.ChainName].ChainID,
			ChannelID:    chInfo.Source,
			StartHeight:  chInfo.StuckPackets.SrcHeight,
			NewAckHeight: make(chan uint64),
		}
		// set default for newAckHeight
		newAckHeight = 0

		//TODO:start scan for new Acks from start height

		// read from NewAckHeight from AckProcessor
		select {
		case newAckHeight = <-cc.AckProcessors[srcHashKey].NewAckHeight:
			log.Debug("new ack found", zap.String("ChainName", path.Chain1.ChainName), zap.String("ChannelID", chInfo.Source),
				zap.Uint64("newAckHeight", newAckHeight))
		default:
			log.Debug("no new acks found", zap.String("ChainName", path.Chain1.ChainName), zap.String("ChannelID", chInfo.Source))
			newAckHeight = 0
		}

	} else if _, ok := cc.AckProcessors[srcHashKey]; ok {
		// AckProcessor already running, read from NewAckHeight
		// set default for newAckHeight
		newAckHeight = 0

		//TODO:start scan for new Acks from start height

		// read from NewAckHeight from AckProcessor
		select {
		case newAckHeight = <-cc.AckProcessors[srcHashKey].NewAckHeight:
			log.Debug("new ack found", zap.String("ChainName", path.Chain1.ChainName), zap.String("ChannelID", chInfo.Source),
				zap.Uint64("newAckHeight", newAckHeight))
		default:
			log.Debug("no new acks found", zap.String("ChainName", path.Chain1.ChainName), zap.String("ChannelID", chInfo.Source))
			newAckHeight = 0
		}
	} else {
		// no stuck packets, no processor running, no new acks
		newAckHeight = 0
	}

	// Publish source chain metric
	// channelNewAckSinceStuck = prometheus.NewDesc(
	// 	channelNewAckSinceStuckMetricName,
	// 	"Returns 1 if new IBC ack was observed since last stuck packet detection, else returns 0.",
	// 	[]string{
	// 		"src_channel_id",
	// 		"dst_channel_id",
	// 		"src_chain_id",
	// 		"dst_chain_id",
	// 		"src_chain_height",
	// 		"src_chain_name",
	// 		"dst_chain_name",
	// 		"status",
	// 	},
	// 	nil,
	// )
	metric <- prometheus.MustNewConstMetric(
		channelNewAckSinceStuck,
		prometheus.GaugeValue,
		float64(newAckHeight),
		[]string{
			chInfo.Source,
			chInfo.Destination,
			(*cc.RPCs)[path.Chain1.ChainName].ChainID,
			(*cc.RPCs)[path.Chain2.ChainName].ChainID,
			path.Chain1.ChainName,
			path.Chain2.ChainName,
			"success",
		}...,
	)

	// Publish destination chain metric
	if _, ok := cc.AckProcessors[dstHashKey]; len(chInfo.StuckPackets.Dst) > 0 && !ok {
		// Start Dst chain AckProcessors if stuck packets exist and no processor is running
		log.Info("Creating packet processor", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Destination))
		cc.AckProcessors[dstHashKey] = AckProcessor{
			ChainID:      (*cc.RPCs)[path.Chain2.ChainName].ChainID,
			ChannelID:    chInfo.Source,
			StartHeight:  chInfo.StuckPackets.DstHeight,
			NewAckHeight: make(chan uint64),
		}
		//start scan for new Acks from start height

		// read from NewAckHeight from AckProcessor
		log.Info("creating packet processor", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Destination))

		select {
		case newAckHeight = <-cc.AckProcessors[dstHashKey].NewAckHeight:
			log.Debug("new ack found", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Destination),
				zap.Uint64("newAckHeight", newAckHeight))
		default:
			log.Debug("no new acks found", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Destination))
			newAckHeight = 0
		}

		newAckHeight = <-cc.AckProcessors[path.Chain2.ChainName].NewAckHeight
	} else if _, ok := cc.AckProcessors[dstHashKey]; ok {
		// AckProcessor already running, read from NewAckHeight
		// set default for newAckHeight
		newAckHeight = 0

		//TODO:start scan for new Acks from start height

		// read from NewAckHeight from AckProcessor
		select {
		case newAckHeight = <-cc.AckProcessors[dstHashKey].NewAckHeight:
			log.Debug("new ack found", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Source),
				zap.Uint64("newAckHeight", newAckHeight))
		default:
			log.Debug("no new acks found", zap.String("ChainName", path.Chain2.ChainName), zap.String("ChannelID", chInfo.Source))
			newAckHeight = 0
		}
	} else {
		// no stuck packets, no processor running, no new acks
		newAckHeight = 0
	}

	metric <- prometheus.MustNewConstMetric(
		channelNewAckSinceStuck,
		prometheus.GaugeValue,
		float64(newAckHeight),
		[]string{
			chInfo.Source,
			chInfo.Destination,
			(*cc.RPCs)[path.Chain1.ChainName].ChainID,
			(*cc.RPCs)[path.Chain2.ChainName].ChainID,
			path.Chain1.ChainName,
			path.Chain2.ChainName,
			"fail",
		}...,
	)

	return nil
}

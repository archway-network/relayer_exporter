package ibc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/relayer/v2/relayer"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	"github.com/archway-network/relayer_exporter/pkg/config"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const stateOpen = 3

var (
	RtyAttNum = uint(5)
	RtyAtt    = retry.Attempts(RtyAttNum)
	RtyDel    = retry.Delay(time.Millisecond * 400)
	RtyErr    = retry.LastErrorOnly(true)

	defaultCoinType uint32 = 118
	defaultAlgo     string = string(hd.Secp256k1Type)
)

type ClientsInfo struct {
	ChainA                 *relayer.Chain
	ChainAClientInfo       relayer.ClientStateInfo
	ChainAClientExpiration time.Time
	ChainB                 *relayer.Chain
	ChainBClientInfo       relayer.ClientStateInfo
	ChainBClientExpiration time.Time
}

type ChannelsInfo struct {
	Channels []Channel
}

type Channel struct {
	Source          string
	Destination     string
	SourcePort      string
	DestinationPort string
	Ordering        string
	StuckPackets    struct {
		Source      int
		Destination int
	}
}

func GetClientsInfo(ctx context.Context, ibc *config.IBCData, rpcs *map[string]config.RPC) (ClientsInfo, error) {
	clientsInfo := ClientsInfo{}

	cdA := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain1.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain1.ChainName].URL,
		Timeout:  (*rpcs)[ibc.Chain1.ChainName].Timeout,
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := chain.PrepChain(ctx, cdA)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("%w for %v", err, cdA)
	}

	clientsInfo.ChainA = chainA

	cdB := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain2.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain2.ChainName].URL,
		Timeout:  (*rpcs)[ibc.Chain2.ChainName].Timeout,
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := chain.PrepChain(ctx, cdB)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("%w for %v", err, cdB)
	}

	clientsInfo.ChainB = chainB

	clientsInfo.ChainAClientExpiration, clientsInfo.ChainAClientInfo, err = relayer.QueryClientExpiration(
		ctx,
		chainA,
		chainB,
	)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("%w path %v <-> %v", err, cdA, cdB)
	}

	clientsInfo.ChainBClientExpiration, clientsInfo.ChainBClientInfo, err = relayer.QueryClientExpiration(
		ctx,
		chainB,
		chainA,
	)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("%w path %v <-> %v", err, cdB, cdA)
	}

	return clientsInfo, nil
}

func GetChannelsInfo(ctx context.Context, ibc *config.IBCData, rpcs *map[string]config.RPC) (ChannelsInfo, error) {
	channelInfo := ChannelsInfo{}

	// Init channel data
	for _, c := range ibc.Channels {
		var channel Channel
		channel.Source = c.Chain1.ChannelID
		channel.Destination = c.Chain2.ChannelID
		channel.SourcePort = c.Chain1.PortID
		channel.DestinationPort = c.Chain2.PortID
		channel.Ordering = c.Ordering
		channelInfo.Channels = append(channelInfo.Channels, channel)
	}

	cdA := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain1.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain1.ChainName].URL,
		Timeout:  (*rpcs)[ibc.Chain1.ChainName].Timeout,
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := chain.PrepChain(ctx, cdA)
	if err != nil {
		return ChannelsInfo{}, fmt.Errorf("error: %w for %+v", err, cdA)
	}

	cdB := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain2.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain2.ChainName].URL,
		Timeout:  (*rpcs)[ibc.Chain2.ChainName].Timeout,
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := chain.PrepChain(ctx, cdB)
	if err != nil {
		return ChannelsInfo{}, fmt.Errorf("error: %w for %+v", err, cdB)
	}

	// test that RPC endpoints are working
	if _, _, err := relayer.QueryLatestHeights(
		ctx, chainA, chainB,
	); err != nil {
		return channelInfo, fmt.Errorf("error: %w for %v", err, cdA)
	}

	for i, c := range channelInfo.Channels {
		var order chantypes.Order

		switch c.Ordering {
		case "none":
			order = chantypes.NONE
		case "unordered":
			order = chantypes.UNORDERED
		case "ordered":
			order = chantypes.ORDERED
		}

		ch := chantypes.IdentifiedChannel{
			State:    stateOpen,
			Ordering: order,
			Counterparty: chantypes.Counterparty{
				PortId:    c.DestinationPort,
				ChannelId: c.Destination,
			},
			PortId:    c.SourcePort,
			ChannelId: c.Source,
		}

		unrelayedSequences := UnrelayedSequences(ctx, chainA, chainB, &ch)

		channelInfo.Channels[i].StuckPackets.Source += len(unrelayedSequences.Src)
		channelInfo.Channels[i].StuckPackets.Destination += len(unrelayedSequences.Dst)
	}

	return channelInfo, nil
}

// UnrelayedSequences returns the unrelayed sequence numbers between two chains
func UnrelayedSequences(ctx context.Context, src, dst *relayer.Chain, srcChannel *chantypes.IdentifiedChannel) relayer.RelaySequences {
	var (
		srcPacketSeq = []uint64{}
		dstPacketSeq = []uint64{}
		rs           = relayer.RelaySequences{Src: []uint64{}, Dst: []uint64{}}
	)

	srch, dsth, err := relayer.QueryLatestHeights(ctx, src, dst)
	if err != nil {
		log.Error("Error querying latest heights", zap.Error(err))
		return rs
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		var (
			res *chantypes.QueryPacketCommitmentsResponse
			err error
		)
		if err = retry.Do(func() error {
			// Query the packet commitment
			res, err = src.ChainProvider.QueryPacketCommitments(ctx, uint64(srch), srcChannel.ChannelId, srcChannel.PortId)
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("no error on QueryPacketCommitments for %s, however response is nil", src.ChainID())
			default:
				return nil
			}
		}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr, retry.OnRetry(func(n uint, err error) {
			log.Info(
				"Failed to query packet commitments",
				zap.String("channel_id", srcChannel.ChannelId),
				zap.String("port_id", srcChannel.PortId),
				zap.Uint("attempt", n+1),
				zap.Uint("max_attempts", RtyAttNum),
				zap.Error(err),
			)
		})); err != nil {
			log.Error(
				"Failed to query packet commitments after max retries",
				zap.String("channel_id", srcChannel.ChannelId),
				zap.String("port_id", srcChannel.PortId),
				zap.Uint("attempts", RtyAttNum),
				zap.Error(err),
			)
			return
		}

		for _, pc := range res.Commitments {
			srcPacketSeq = append(srcPacketSeq, pc.Sequence)
		}
	}()

	go func() {
		defer wg.Done()
		var (
			res *chantypes.QueryPacketCommitmentsResponse
			err error
		)
		if err = retry.Do(func() error {
			res, err = dst.ChainProvider.QueryPacketCommitments(ctx, uint64(dsth), srcChannel.Counterparty.ChannelId, srcChannel.Counterparty.PortId)
			switch {
			case err != nil:
				return err
			case res == nil:
				return fmt.Errorf("no error on QueryPacketCommitments for %s, however response is nil", dst.ChainID())
			default:
				return nil
			}
		}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr, retry.OnRetry(func(n uint, err error) {
			log.Info(
				"Failed to query packet commitments",
				zap.String("channel_id", srcChannel.Counterparty.ChannelId),
				zap.String("port_id", srcChannel.Counterparty.PortId),
				zap.Uint("attempt", n+1),
				zap.Uint("max_attempts", RtyAttNum),
				zap.Error(err),
			)
		})); err != nil {
			log.Error(
				"Failed to query packet commitments after max retries",
				zap.String("channel_id", srcChannel.Counterparty.ChannelId),
				zap.String("port_id", srcChannel.Counterparty.PortId),
				zap.Uint("attempts", RtyAttNum),
				zap.Error(err),
			)
			return
		}

		for _, pc := range res.Commitments {
			dstPacketSeq = append(dstPacketSeq, pc.Sequence)
		}
	}()

	wg.Wait()

	var (
		srcUnreceivedPackets, dstUnreceivedPackets []uint64
	)

	if len(srcPacketSeq) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Query all packets sent by src that have not been received by dst.
			if err := retry.Do(func() error {
				var err error
				srcUnreceivedPackets, err = dst.ChainProvider.QueryUnreceivedPackets(ctx, uint64(dsth), srcChannel.Counterparty.ChannelId, srcChannel.Counterparty.PortId, srcPacketSeq)
				return err
			}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr, retry.OnRetry(func(n uint, err error) {
				log.Info(
					"Failed to query unreceived packets",
					zap.String("channel_id", srcChannel.Counterparty.ChannelId),
					zap.String("port_id", srcChannel.Counterparty.PortId),
					zap.Uint("attempt", n+1),
					zap.Uint("max_attempts", RtyAttNum),
					zap.Error(err),
				)
			})); err != nil {
				log.Error(
					"Failed to query unreceived packets after max retries",
					zap.String("channel_id", srcChannel.Counterparty.ChannelId),
					zap.String("port_id", srcChannel.Counterparty.PortId),
					zap.Uint("attempts", RtyAttNum),
					zap.Error(err),
				)
			}
		}()
	}

	if len(dstPacketSeq) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Query all packets sent by dst that have not been received by src.
			if err := retry.Do(func() error {
				var err error
				dstUnreceivedPackets, err = src.ChainProvider.QueryUnreceivedPackets(ctx, uint64(srch), srcChannel.ChannelId, srcChannel.PortId, dstPacketSeq)
				return err
			}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr, retry.OnRetry(func(n uint, err error) {
				log.Info(
					"Failed to query unreceived packets",
					zap.String("channel_id", srcChannel.ChannelId),
					zap.String("port_id", srcChannel.PortId),
					zap.Uint("attempt", n+1),
					zap.Uint("max_attempts", RtyAttNum),
					zap.Error(err),
				)
			})); err != nil {
				log.Error(
					"Failed to query unreceived packets after max retries",
					zap.String("channel_id", srcChannel.ChannelId),
					zap.String("port_id", srcChannel.PortId),
					zap.Uint("attempts", RtyAttNum),
					zap.Error(err),
				)
				return
			}
		}()
	}
	wg.Wait()

	// If this is an UNORDERED channel we can return at this point.
	if srcChannel.Ordering != chantypes.ORDERED {
		rs.Src = srcUnreceivedPackets
		rs.Dst = dstUnreceivedPackets
		return rs
	}

	// For ordered channels we want to only relay the packet whose sequence number is equal to
	// the expected next packet receive sequence from the counterparty.
	if len(srcUnreceivedPackets) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nextSeqResp, err := dst.ChainProvider.QueryNextSeqRecv(ctx, dsth, srcChannel.Counterparty.ChannelId, srcChannel.Counterparty.PortId)
			if err != nil {
				log.Error(
					"Failed to query next packet receive sequence",
					zap.String("channel_id", srcChannel.Counterparty.ChannelId),
					zap.String("port_id", srcChannel.Counterparty.PortId),
					zap.Error(err),
				)
				return
			}

			for _, seq := range srcUnreceivedPackets {
				if seq == nextSeqResp.NextSequenceReceive {
					rs.Src = append(rs.Src, seq)
					break
				}
			}
		}()
	}

	if len(dstUnreceivedPackets) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nextSeqResp, err := src.ChainProvider.QueryNextSeqRecv(ctx, srch, srcChannel.ChannelId, srcChannel.PortId)
			if err != nil {
				log.Error(
					"Failed to query next packet receive sequence",
					zap.String("channel_id", srcChannel.ChannelId),
					zap.String("port_id", srcChannel.PortId),
					zap.Error(err),
				)
				return
			}

			for _, seq := range dstUnreceivedPackets {
				if seq == nextSeqResp.NextSequenceReceive {
					rs.Dst = append(rs.Dst, seq)
					break
				}
			}
		}()
	}
	wg.Wait()

	return rs
}

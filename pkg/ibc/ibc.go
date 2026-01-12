package ibc

import (
	"context"
	"fmt"
	"strings"
	"time"

	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/relayer/v2/relayer"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	"github.com/archway-network/relayer_exporter/pkg/config"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const stateOpen = 3

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
	defer func() {
		if r := recover(); r != nil {
			log.GetLogger().Sugar().Debugf("Recovered from panic:", r)
		}
	}()

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
		// Ignore channels with wildcard ChannelID or PortID
		if strings.Contains(c.Chain1.ChannelID, "*") || strings.Contains(c.Chain1.PortID, "*") ||
			strings.Contains(c.Chain2.ChannelID, "*") || strings.Contains(c.Chain2.PortID, "*") {
			log.Info("Skipping channel with wildcard ChannelID or PortID", zap.Any("channel", c))
			continue
		}

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

		unrelayedSequences := relayer.UnrelayedSequences(ctx, chainA, chainB, &ch)

		channelInfo.Channels[i].StuckPackets.Source += len(unrelayedSequences.Src)
		channelInfo.Channels[i].StuckPackets.Destination += len(unrelayedSequences.Dst)
	}

	return channelInfo, nil
}

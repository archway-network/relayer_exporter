package ibc

import (
	"context"
	"fmt"
	"time"

	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/relayer/v2/relayer"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	"github.com/archway-network/relayer_exporter/pkg/config"
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
	Source       string
	Destination  string
	StuckPackets struct {
		Source      int
		Destination int
		Total       int
	}
}

func GetClientsInfo(ibc *relayer.IBCdata, rpcs *map[string]config.RPC) (ClientsInfo, error) {
	clientsInfo := ClientsInfo{}

	cdA := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain1.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain1.ChainName].URL,
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := chain.PrepChain(cdA)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w for %v", err, cdA)
	}

	clientsInfo.ChainA = chainA

	cdB := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain2.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain2.ChainName].URL,
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := chain.PrepChain(cdB)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w for %v", err, cdB)
	}

	clientsInfo.ChainB = chainB

	ctx := context.Background()

	clientsInfo.ChainAClientExpiration, clientsInfo.ChainAClientInfo, err = relayer.QueryClientExpiration(
		ctx,
		chainA,
		chainB,
	)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdA, cdB)
	}

	clientsInfo.ChainBClientExpiration, clientsInfo.ChainBClientInfo, err = relayer.QueryClientExpiration(
		ctx,
		chainB,
		chainA,
	)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdB, cdA)
	}

	return clientsInfo, nil
}

func GetChannelsInfo(ibc *relayer.IBCdata, rpcs *map[string]config.RPC) (ChannelsInfo, error) {
	ctx := context.Background()
	channelInfo := ChannelsInfo{}

	if (*rpcs)[ibc.Chain1.ChainName].ChainID == "" || (*rpcs)[ibc.Chain2.ChainName].ChainID == "" {
		return ChannelsInfo{}, fmt.Errorf(
			"Error: RPC data is missing, cannot retrieve channel data: %v",
			ibc.Channels,
		)
	}

	cdA := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain1.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain1.ChainName].URL,
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := chain.PrepChain(cdA)
	if err != nil {
		return ChannelsInfo{}, fmt.Errorf("Error: %w for %v", err, cdA)
	}

	cdB := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain2.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain2.ChainName].URL,
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := chain.PrepChain(cdB)
	if err != nil {
		return ChannelsInfo{}, fmt.Errorf("Error: %w for %v", err, cdB)
	}

	// test that RPC endpoints are working
	if _, _, err := relayer.QueryLatestHeights(
		ctx, chainA, chainB,
	); err != nil {
		return ChannelsInfo{}, fmt.Errorf("Error: %w for %v", err, cdA)
	}

	for _, c := range ibc.Channels {
		var order chantypes.Order

		var channel Channel

		channel.Source = c.Chain1.ChannelID
		channel.Destination = c.Chain2.ChannelID

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
				PortId:    c.Chain2.PortID,
				ChannelId: c.Chain2.ChannelID,
			},
			PortId:    c.Chain1.PortID,
			ChannelId: c.Chain2.ChannelID,
		}

		unrelayedSequences := relayer.UnrelayedSequences(ctx, chainA, chainB, &ch)

		channel.StuckPackets.Total += len(unrelayedSequences.Src) + len(unrelayedSequences.Dst)
		channel.StuckPackets.Source += len(unrelayedSequences.Src)
		channel.StuckPackets.Destination += len(unrelayedSequences.Dst)

		channelInfo.Channels = append(channelInfo.Channels, channel)
	}

	return channelInfo, nil
}

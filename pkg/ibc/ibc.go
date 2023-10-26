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

type ClientsInfo struct {
	ChainA                 *relayer.Chain
	ChainAClientInfo       relayer.ClientStateInfo
	ChainAClientExpiration time.Time
	ChainB                 *relayer.Chain
	ChainBClientInfo       relayer.ClientStateInfo
	ChainBClientExpiration time.Time
}

type ChannelInfo struct {
	StuckPackets struct {
		Source      int
		Destination int
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

func GetChannelInfo(ibc *relayer.IBCdata, rpcs *map[string]config.RPC) (ChannelInfo, error) {
	ctx := context.Background()
	channelInfo := ChannelInfo{}

	cdA := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain1.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain1.ChainName].URL,
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := chain.PrepChain(cdA)
	if err != nil {
		return ChannelInfo{}, fmt.Errorf("Error: %w for %v", err, cdA)
	}

	cdB := chain.Info{
		ChainID:  (*rpcs)[ibc.Chain2.ChainName].ChainID,
		RPCAddr:  (*rpcs)[ibc.Chain2.ChainName].URL,
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := chain.PrepChain(cdB)
	if err != nil {
		return ChannelInfo{}, fmt.Errorf("Error: %w for %v", err, cdB)
	}

	for _, c := range ibc.Channels {
		var order chantypes.Order

		switch c.Ordering {
		case "none":
			order = chantypes.NONE
		case "unordered":
			order = chantypes.UNORDERED
		case "ordered":
			order = chantypes.ORDERED
		}

		channel := chantypes.IdentifiedChannel{
			State:    3,
			Ordering: order,
			Counterparty: chantypes.Counterparty{
				PortId:    c.Chain2.PortID,
				ChannelId: c.Chain2.ChannelID,
			},
			PortId:    c.Chain1.PortID,
			ChannelId: c.Chain2.ChannelID,
		}

		unrelayedSequences := relayer.UnrelayedSequences(ctx, chainA, chainB, &channel)
		channelInfo.StuckPackets.Source += len(unrelayedSequences.Src)
		channelInfo.StuckPackets.Destination += len(unrelayedSequences.Dst)
	}

	return channelInfo, nil
}

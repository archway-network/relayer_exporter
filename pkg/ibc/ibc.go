package ibc

import (
	"context"
	"fmt"
	"time"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	"github.com/archway-network/relayer_exporter/pkg/config"
	"github.com/cosmos/relayer/v2/relayer"
)

type ClientsInfo struct {
	ChainA                 *relayer.Chain
	ChainAClientInfo       relayer.ClientStateInfo
	ChainAClientExpiration time.Time
	ChainB                 *relayer.Chain
	ChainBClientInfo       relayer.ClientStateInfo
	ChainBClientExpiration time.Time
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

	clientsInfo.ChainAClientExpiration, clientsInfo.ChainAClientInfo, err = relayer.QueryClientExpiration(ctx, chainA, chainB)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdA, cdB)
	}

	clientsInfo.ChainBClientExpiration, clientsInfo.ChainBClientInfo, err = relayer.QueryClientExpiration(ctx, chainB, chainA)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdB, cdA)
	}

	return clientsInfo, nil
}

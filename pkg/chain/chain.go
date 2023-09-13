package chain

import (
	"context"

	"github.com/cosmos/relayer/v2/relayer"
	"github.com/cosmos/relayer/v2/relayer/chains/cosmos"
)

const (
	rpcTimeout     = "10s"
	keyringBackend = "test"
)

type Info struct {
	ChainID  string
	RPCAddr  string
	ClientID string
}

func PrepChain(info Info) (*relayer.Chain, error) {
	chain := relayer.Chain{}
	providerConfig := cosmos.CosmosProviderConfig{
		ChainID:        info.ChainID,
		Timeout:        rpcTimeout,
		KeyringBackend: keyringBackend,
		RPCAddr:        info.RPCAddr,
	}

	provider, err := providerConfig.NewProvider(nil, "", false, info.ChainID)
	if err != nil {
		return nil, err
	}

	err = provider.Init(context.Background())
	if err != nil {
		return nil, err
	}

	chain.ChainProvider = provider

	err = chain.SetPath(&relayer.PathEnd{ClientID: info.ClientID})
	if err != nil {
		return nil, err
	}

	return &chain, nil
}

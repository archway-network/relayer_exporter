package chain

import (
	"context"

	"github.com/cosmos/relayer/v2/relayer"
	"github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"go.uber.org/zap"
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
	logger := zap.NewNop()
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

	chain := relayer.NewChain(logger, provider, false)

	err = chain.SetPath(&relayer.PathEnd{ClientID: info.ClientID})
	if err != nil {
		return nil, err
	}

	return chain, nil
}

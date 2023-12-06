package chain

import (
	"context"

	log "github.com/archway-network/relayer_exporter/pkg/logger"
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
	Timeout  string
}

func PrepChain(ctx context.Context, info Info) (*relayer.Chain, error) {
	timeout := rpcTimeout
	if info.Timeout != "" {
		timeout = info.Timeout
	}

	providerConfig := cosmos.CosmosProviderConfig{
		ChainID:        info.ChainID,
		Timeout:        timeout,
		KeyringBackend: keyringBackend,
		RPCAddr:        info.RPCAddr,
	}

	provider, err := providerConfig.NewProvider(nil, "", false, info.ChainID)
	if err != nil {
		return nil, err
	}

	err = provider.Init(ctx)
	if err != nil {
		return nil, err
	}

	chain := relayer.NewChain(log.GetLogger(), provider, false)

	err = chain.SetPath(&relayer.PathEnd{ClientID: info.ClientID})
	if err != nil {
		return nil, err
	}

	return chain, nil
}

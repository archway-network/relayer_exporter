package account

import (
	"context"

	"cosmossdk.io/math"

	"github.com/archway-network/relayer_exporter/pkg/chain"
)

type Account struct {
	Address string `yaml:"address"`
	Denom   string `yaml:"denom"`
	ChainID string `yaml:"chainId"`
	Balance math.Int
}

func (a *Account) GetBalance(rpcs map[string]string) error {
	chain, err := chain.PrepChain(chain.Info{
		ChainID: a.ChainID,
		RPCAddr: rpcs[a.ChainID],
	})
	if err != nil {
		return err
	}

	ctx := context.Background()

	coins, err := chain.ChainProvider.QueryBalanceWithAddress(ctx, a.Address)
	if err != nil {
		return err
	}

	a.Balance = coins.AmountOf(a.Denom)

	return nil
}

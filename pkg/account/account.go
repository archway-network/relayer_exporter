package account

import (
	"context"

	"cosmossdk.io/math"
	"go.uber.org/zap"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

type Account struct {
	Address string `yaml:"address"`
	Denom   string `yaml:"denom"`
	ChainID string `yaml:"chainId"`
	Balance math.Int
}

func GetBalances(accounts []Account, rpcs map[string]string) []Account {
	num := len(accounts)

	out := make(chan Account, num)
	defer close(out)

	for i := 0; i < num; i++ {
		go func(i int) {
			err := accounts[i].GetBalance(rpcs)
			if err != nil {
				out <- Account{}

				log.Error(err.Error(), zap.Any("account", accounts[i]))

				return
			}
			out <- accounts[i]
		}(i)
	}

	accountsWithBalance := []Account{}

	for i := 0; i < num; i++ {
		account := <-out
		if account.Address != "" {
			accountsWithBalance = append(accountsWithBalance, account)
		}
	}

	return accountsWithBalance
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

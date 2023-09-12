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
			account, err := GetBalance(accounts[i], rpcs)
			if err != nil {
				out <- Account{}

				log.Error(err.Error(), zap.Any("account", accounts[i]))

				return
			}
			out <- account
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

func GetBalance(account Account, rpcs map[string]string) (Account, error) {
	chain, err := chain.PrepChain(chain.Info{
		ChainID: account.ChainID,
		RPCAddr: rpcs[account.ChainID],
	})
	if err != nil {
		return Account{}, err
	}

	ctx := context.Background()

	coins, err := chain.ChainProvider.QueryBalanceWithAddress(ctx, account.Address)
	if err != nil {
		return Account{}, err
	}

	account.Balance = coins.AmountOf(account.Denom)

	return account, nil
}

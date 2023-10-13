package config

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"cosmossdk.io/math"
	"github.com/caarlos0/env/v9"
	"github.com/cosmos/relayer/v2/relayer"
	"github.com/google/go-github/v55/github"
	"gopkg.in/yaml.v3"

	"github.com/archway-network/relayer_exporter/pkg/chain"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const ibcPathSuffix = ".json"

var ErrGitHubClient = errors.New("GitHub client not provided")

type Account struct {
	Address   string `yaml:"address"`
	Denom     string `yaml:"denom"`
	ChainName string `yaml:"chainName"`
	Balance   math.Int
}

type RPC struct {
	ChainName string `yaml:"chainName"`
	ChainID   string `yaml:"chainId"`
	URL       string `yaml:"url"`
}

type Config struct {
	Accounts []Account `yaml:"accounts"`
	RPCs     []RPC     `yaml:"rpc"`
	GitHub   struct {
		Org            string `yaml:"org"`
		Repo           string `yaml:"repo"`
		IBCDir         string `yaml:"dir"`
		TestnetsIBCDir string `yaml:"testnetsDir"`
		Token          string `env:"GITHUB_TOKEN"`
	} `yaml:"github"`
}

func (a *Account) GetBalance(rpcs *map[string]RPC) error {
	chain, err := chain.PrepChain(chain.Info{
		ChainID: (*rpcs)[a.ChainName].ChainID,
		RPCAddr: (*rpcs)[a.ChainName].URL,
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

func (c *Config) GetRPCsMap() *map[string]RPC {
	rpcs := map[string]RPC{}

	for _, rpc := range c.RPCs {
		rpcs[rpc.ChainName] = rpc
	}

	return &rpcs
}

func (c *Config) IBCPaths() ([]*relayer.IBCdata, error) {
	client := github.NewClient(nil)

	if c.GitHub.Token != "" {
		log.Debug("Using provided GITHUB_TOKEN env var for GitHub client")

		client = github.NewClient(nil).WithAuthToken(c.GitHub.Token)
	}

	paths, err := c.getPaths(c.GitHub.IBCDir, client)
	if err != nil {
		return nil, err
	}

	testnetsPaths := []*relayer.IBCdata{}
	if c.GitHub.TestnetsIBCDir != "" {
		testnetsPaths, err = c.getPaths(c.GitHub.TestnetsIBCDir, client)
		if err != nil {
			return nil, err
		}
	}

	paths = append(paths, testnetsPaths...)

	return paths, nil
}

func (c *Config) getPaths(dir string, client *github.Client) ([]*relayer.IBCdata, error) {
	if client == nil {
		return nil, ErrGitHubClient
	}

	ctx := context.Background()

	_, ibcDir, _, err := client.Repositories.GetContents(ctx, c.GitHub.Org, c.GitHub.Repo, dir, nil)
	if err != nil {
		return nil, err
	}

	ibcs := []*relayer.IBCdata{}

	for _, file := range ibcDir {
		if strings.HasSuffix(*file.Path, ibcPathSuffix) {
			content, _, _, err := client.Repositories.GetContents(ctx, c.GitHub.Org, c.GitHub.Repo, *file.Path, nil)
			if err != nil {
				return nil, err
			}

			ibc := &relayer.IBCdata{}

			c, err := content.GetContent()
			if err != nil {
				return nil, err
			}

			if err = json.Unmarshal([]byte(c), &ibc); err != nil {
				return nil, err
			}

			ibcs = append(ibcs, ibc)
		}
	}

	return ibcs, nil
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(config); err != nil {
		return nil, err
	}

	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}

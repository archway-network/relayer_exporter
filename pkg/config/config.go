package config

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"cosmossdk.io/math"
	"github.com/caarlos0/env/v9"
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
	Timeout   string `yaml:"timeout"`
}

type Config struct {
	Accounts         []Account `yaml:"accounts"`
	GlobalRPCTimeout string    `env:"GLOBAL_RPC_TIMEOUT" envDefault:"5s"`
	RPCs             []RPC     `yaml:"rpc"`
	GitHub           struct {
		Org            string `yaml:"org"`
		Repo           string `yaml:"repo"`
		IBCDir         string `yaml:"dir"`
		TestnetsIBCDir string `yaml:"testnetsDir"`
		Token          string `env:"GITHUB_TOKEN"`
	} `yaml:"github"`
}

type IBCData struct {
	Schema string `json:"$schema"`
	Chain1 struct {
		ChainName    string `json:"chain_name"`
		ClientID     string `json:"client_id"`
		ConnectionID string `json:"connection_id"`
	} `json:"chain_1"`
	Chain2 struct {
		ChainName    string `json:"chain_name"`
		ClientID     string `json:"client_id"`
		ConnectionID string `json:"connection_id"`
	} `json:"chain_2"`
	Channels []struct {
		Chain1 struct {
			ChannelID string `json:"channel_id"`
			PortID    string `json:"port_id"`
		} `json:"chain_1"`
		Chain2 struct {
			ChannelID string `json:"channel_id"`
			PortID    string `json:"port_id"`
		} `json:"chain_2"`
		Ordering string `json:"ordering"`
		Version  string `json:"version"`
		Tags     struct {
			Status     string `json:"status"`
			Preferred  bool   `json:"preferred"`
			Dex        string `json:"dex"`
			Properties string `json:"properties"`
		} `json:"tags,omitempty"`
	} `json:"channels"`
	Operators []Operator `json:"operators"`
}

type Operator struct {
	Chain1 struct {
		Address string `json:"address"`
	} `json:"chain_1"`
	Chain2 struct {
		Address string `json:"address"`
	} `json:"chain_2"`
	Memo    string  `json:"memo"`
	Name    string  `json:"name"`
	Discord Discord `json:"discord"`
}

type Discord struct {
	Handle string `json:"handle"`
	ID     string `json:"id"`
}

func (a *Account) GetBalance(rpcs *map[string]RPC) error {
	chain, err := chain.PrepChain(chain.Info{
		ChainID: (*rpcs)[a.ChainName].ChainID,
		RPCAddr: (*rpcs)[a.ChainName].URL,
		Timeout: (*rpcs)[a.ChainName].Timeout,
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
		if rpc.Timeout == "" {
			rpc.Timeout = c.GlobalRPCTimeout
		}

		rpcs[rpc.ChainName] = rpc
	}

	return &rpcs
}

func (c *Config) IBCPaths() ([]*IBCData, error) {
	client := github.NewClient(nil)

	if c.GitHub.Token != "" {
		log.Debug("Using provided GITHUB_TOKEN env var for GitHub client")

		client = github.NewClient(nil).WithAuthToken(c.GitHub.Token)
	}

	paths, err := c.getPaths(c.GitHub.IBCDir, client)
	if err != nil {
		return nil, err
	}

	testnetsPaths := []*IBCData{}
	if c.GitHub.TestnetsIBCDir != "" {
		testnetsPaths, err = c.getPaths(c.GitHub.TestnetsIBCDir, client)
		if err != nil {
			return nil, err
		}
	}

	paths = append(paths, testnetsPaths...)

	return paths, nil
}

func (c *Config) getPaths(dir string, client *github.Client) ([]*IBCData, error) {
	if client == nil {
		return nil, ErrGitHubClient
	}

	ctx := context.Background()

	_, ibcDir, _, err := client.Repositories.GetContents(ctx, c.GitHub.Org, c.GitHub.Repo, dir, nil)
	if err != nil {
		return nil, err
	}

	ibcs := []*IBCData{}

	for _, file := range ibcDir {
		if strings.HasSuffix(*file.Path, ibcPathSuffix) {
			content, _, _, err := client.Repositories.GetContents(
				ctx,
				c.GitHub.Org,
				c.GitHub.Repo,
				*file.Path,
				nil,
			)
			if err != nil {
				return nil, err
			}

			ibc := &IBCData{}

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

package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v55/github"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const ibcPathSuffix = ".json"

var (
	ErrGitHubClient        = errors.New("GitHub client not provided")
	ErrMissingRPCConfigMsg = "missing RPC config for chain: %s"
)

type Account struct {
	Address   string   `yaml:"address" validate:"required"`
	Denom     []string `yaml:"denom" validate:"required"`
	ChainName string   `yaml:"chainName" validate:"required"`
	Balance   []math.Int
	Tags      []string `yaml:"tags,omitempty"`
}

type RPC struct {
	ChainName string `yaml:"chainName" validate:"required"`
	ChainID   string `yaml:"chainId" validate:"required"`
	URL       string `yaml:"url" validate:"required,http_url,has_port"`
	Timeout   string `yaml:"timeout"`
}

type GitHub struct {
	Org            string `yaml:"org" validate:"required"`
	Repo           string `yaml:"repo" validate:"required"`
	IBCDir         string `yaml:"dir" validate:"required"`
	TestnetsIBCDir string `yaml:"testnetsDir"`
	Token          string `env:"GITHUB_TOKEN"`
}

type Config struct {
	Accounts         []*Account `yaml:"accounts"`
	GlobalRPCTimeout string     `env:"GLOBAL_RPC_TIMEOUT" envDefault:"5s"`
	RPCs             []*RPC     `yaml:"rpc"`
	GitHub           *GitHub    `yaml:"github"`
}

type IBCChainMeta struct {
	ChainName    string `json:"chain_name"`
	ClientID     string `json:"client_id"`
	ConnectionID string `json:"connection_id"`
}

type Channel struct {
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

type IBCData struct {
	Schema    string       `json:"$schema"`
	Chain1    IBCChainMeta `json:"chain_1"`
	Chain2    IBCChainMeta `json:"chain_2"`
	Channels  []Channel    `json:"channels"`
	Operators []Operator   `json:"operators"`
}

type Discord struct {
	Handle string `json:"handle"`
	ID     string `json:"id"`
}

// GetRPCsMap uses the provided config file to return a map of chain
// chain_names to RPCs. It uses IBCData already extracted from
// github IBC registry to validate config for missing RPCs and raises
// an error if any are missing.
func (c *Config) GetRPCsMap() *map[string]RPC {
	rpcs := map[string]RPC{}

	for _, rpc := range c.RPCs {
		if rpc.Timeout == "" {
			rpc.Timeout = c.GlobalRPCTimeout
		}

		rpcs[rpc.ChainName] = *rpc
	}

	return &rpcs
}

func (c *Config) IBCPaths(ctx context.Context) ([]*IBCData, error) {
	client := github.NewClient(nil)

	if c.GitHub == nil {
		return nil, fmt.Errorf("GitHub configuration is required")
	}

	log.Info(
		fmt.Sprintf(
			"Github IBC registry: %s/%s",
			c.GitHub.Org,
			c.GitHub.Repo,
		),
		zap.String("Mainnet Directory", c.GitHub.IBCDir),
		zap.String("Testnet Directory", c.GitHub.TestnetsIBCDir),
	)

	if c.GitHub.Token != "" {
		log.Debug("Using provided GITHUB_TOKEN env var for GitHub client")

		client = github.NewClient(nil).WithAuthToken(c.GitHub.Token)
	}

	paths, err := c.getPaths(ctx, c.GitHub.IBCDir, client)
	if err != nil {
		return nil, err
	}

	testnetsPaths := []*IBCData{}
	if c.GitHub.TestnetsIBCDir != "" {
		testnetsPaths, err = c.getPaths(ctx, c.GitHub.TestnetsIBCDir, client)
		if err != nil {
			return nil, err
		}
	}

	paths = append(paths, testnetsPaths...)

	return paths, nil
}

func (c *Config) getPaths(ctx context.Context, dir string, client *github.Client) ([]*IBCData, error) {
	if client == nil {
		return nil, ErrGitHubClient
	}

	_, ibcDir, _, err := client.Repositories.GetContents(ctx, c.GitHub.Org, c.GitHub.Repo, dir, nil)
	if err != nil {
		return nil, err
	}

	ibcs := []*IBCData{}

	for _, file := range ibcDir {
		if strings.HasSuffix(*file.Path, ibcPathSuffix) {
			log.Debug(fmt.Sprintf("Fetching IBC data for %s/%s/%s", c.GitHub.Org, c.GitHub.Repo, *file.Path))

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

func (c *Config) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// register custom validation for http url as expected by go relayer i.e.
	// http_url must have port defined.
	// https://github.com/cosmos/relayer/blob/259b1278264180a2aefc2085f1b55753849c4815/cregistry/chain_info.go#L115
	err := validate.RegisterValidation("has_port", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		urlParsed, err := url.Parse(val)
		if err != nil {
			return false
		}

		port := urlParsed.Port()

		// Port must be a iny <= 65535.
		if portNum, err := strconv.ParseInt(
			port, 10, 32,
		); err != nil || portNum > 65535 || portNum < 1 {
			return false
		}

		return true
	})
	if err != nil {
		return err
	}

	// validate top level fields
	if err := validate.Struct(c); err != nil {
		return err
	}

	// validate RPCs
	for _, rpc := range c.RPCs {
		if err := validate.Struct(rpc); err != nil {
			return fmt.Errorf("%v for RPC config: %+v", err, rpc)
		}
	}

	// validate accounts
	for _, account := range c.Accounts {
		if err := validate.Struct(account); err != nil {
			return fmt.Errorf("%v for accounts config: %+v", err, account)
		}

		rpcMap := c.GetRPCsMap()
		if rpcMap != nil {
			_, ok := (*rpcMap)[account.ChainName]
			if !ok {
				return fmt.Errorf(ErrMissingRPCConfigMsg, account.ChainName)
			}
		}
	}

	return nil
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

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

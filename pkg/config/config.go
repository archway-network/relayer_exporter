package config

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/cosmos/relayer/v2/relayer"
	"github.com/google/go-github/v55/github"
	"gopkg.in/yaml.v3"

	"github.com/archway-network/relayer_exporter/pkg/account"
	log "github.com/archway-network/relayer_exporter/pkg/logger"
)

const ibcPathSuffix = ".json"

type RPC struct {
	ChainID string `yaml:"chainId"`
	URL     string `yaml:"url"`
}

type Config struct {
	Accounts []account.Account `yaml:"accounts"`
	RPCs     []RPC             `yaml:"rpc"`
	GitHub   struct {
		Org    string `yaml:"org"`
		Repo   string `yaml:"repo"`
		IBCDir string `yaml:"dir"`
		Token  string `env:"GITHUB_TOKEN"`
	} `yaml:"github"`
}

func (c *Config) GetRPCsMap() map[string]string {
	rpcs := map[string]string{}

	for _, rpc := range c.RPCs {
		rpcs[rpc.ChainID] = rpc.URL
	}

	return rpcs
}

func (c *Config) IBCPaths() ([]*relayer.IBCdata, error) {
	ctx := context.Background()

	client := github.NewClient(nil)

	if c.GitHub.Token != "" {
		log.Debug("Using provided GITHUB_TOKEN env var for GitHub client")

		client = github.NewClient(nil).WithAuthToken(c.GitHub.Token)
	}

	_, ibcDir, _, err := client.Repositories.GetContents(ctx, c.GitHub.Org, c.GitHub.Repo, c.GitHub.IBCDir, nil)
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

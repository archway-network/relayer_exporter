package config

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/cosmos/relayer/v2/relayer"
	"github.com/google/go-github/v54/github"
	"gopkg.in/yaml.v3"
)

const ibcPathSuffix = ".json"

type RPC struct {
	ChainID string `yaml:"chainId"`
	URL     string `yaml:"url"`
}

type Config struct {
	RPCs   []RPC `yaml:"rpc"`
	GitHub struct {
		Org    string `yaml:"org"`
		Repo   string `yaml:"repo"`
		IBCDir string `yaml:"dir"`
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
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

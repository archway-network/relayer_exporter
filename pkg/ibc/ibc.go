package ibc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/archway-network/relayer_exporter/pkg/logger"
	"github.com/cosmos/relayer/v2/relayer"
	"github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v54/github"
	"gopkg.in/yaml.v3"
)

const (
	ibcPathSuffix  = ".json"
	rpcTimeout     = "10s"
	keyringBackend = "test"
)

type ChainData struct {
	ChainID  string
	RPCAddr  string
	ClientID string
}

type ClientsInfo struct {
	ChainA                 *relayer.Chain
	ChainAClientInfo       relayer.ClientStateInfo
	ChainAClientExpiration time.Time
	ChainB                 *relayer.Chain
	ChainBClientInfo       relayer.ClientStateInfo
	ChainBClientExpiration time.Time
}

func (ci ClientsInfo) PathName() string {
	chainAID := ""
	if ci.ChainA != nil {
		chainAID = ci.ChainA.ChainID()
	}

	chainBID := ""
	if ci.ChainB != nil {
		chainBID = ci.ChainB.ChainID()
	}

	return fmt.Sprintf("%s<->%s", chainAID, chainBID)
}

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

func (c *Config) getRPCs() map[string]string {
	rpcs := map[string]string{}

	for _, rpc := range c.RPCs {
		rpcs[rpc.ChainID] = rpc.URL
	}

	return rpcs
}

func (c *Config) GetClientsInfos(ibcs []*relayer.IBCdata) []ClientsInfo {
	num := len(ibcs)

	out := make(chan ClientsInfo, num)
	defer close(out)

	for i := 0; i < num; i++ {
		go func(i int) {
			clientsInfo, err := c.GetClientsInfo(ibcs[i])
			if err != nil {
				out <- ClientsInfo{}

				log.Error(err.Error())

				return
			}
			out <- clientsInfo
		}(i)
	}

	clientsInfos := []ClientsInfo{}

	for i := 0; i < num; i++ {
		ci := <-out
		if !cmp.Equal(ci, ClientsInfo{}) {
			clientsInfos = append(clientsInfos, ci)
		}
	}

	return clientsInfos
}

func (c *Config) GetClientsInfo(ibc *relayer.IBCdata) (ClientsInfo, error) {
	clientsInfo := ClientsInfo{}

	rpcs := c.getRPCs()

	cdA := ChainData{
		ChainID:  ibc.Chain1.ChainName,
		RPCAddr:  rpcs[ibc.Chain1.ChainName],
		ClientID: ibc.Chain1.ClientID,
	}

	chainA, err := prepChain(cdA)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w for %v", err, cdA)
	}

	clientsInfo.ChainA = chainA

	cdB := ChainData{
		ChainID:  ibc.Chain2.ChainName,
		RPCAddr:  rpcs[ibc.Chain2.ChainName],
		ClientID: ibc.Chain2.ClientID,
	}

	chainB, err := prepChain(cdB)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w for %v", err, cdB)
	}

	clientsInfo.ChainB = chainB

	ctx := context.Background()

	clientsInfo.ChainAClientExpiration, clientsInfo.ChainAClientInfo, err = relayer.QueryClientExpiration(ctx, chainA, chainB)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdA, cdB)
	}

	clientsInfo.ChainBClientExpiration, clientsInfo.ChainBClientInfo, err = relayer.QueryClientExpiration(ctx, chainB, chainA)
	if err != nil {
		return ClientsInfo{}, fmt.Errorf("Error: %w path %v <-> %v", err, cdB, cdA)
	}

	return clientsInfo, nil
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

func prepChain(cd ChainData) (*relayer.Chain, error) {
	chain := relayer.Chain{}
	providerConfig := cosmos.CosmosProviderConfig{
		ChainID:        cd.ChainID,
		Timeout:        rpcTimeout,
		KeyringBackend: keyringBackend,
		RPCAddr:        cd.RPCAddr,
	}

	provider, err := providerConfig.NewProvider(nil, "", false, cd.ChainID)
	if err != nil {
		return nil, err
	}

	err = provider.Init(context.Background())
	if err != nil {
		return nil, err
	}

	chain.ChainProvider = provider

	err = chain.SetPath(&relayer.PathEnd{ClientID: cd.ClientID})
	if err != nil {
		return nil, err
	}

	return &chain, nil
}

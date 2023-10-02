package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRPCsMap(t *testing.T) {
	rpcs := []RPC{
		{
			ChainName: "archway",
			ChainID:   "archway-1",
			URL:       "https://rpc.mainnet.archway.io:443",
		},
		{
			ChainName: "archwaytestnet",
			ChainID:   "constantine-3",
			URL:       "https://rpc.constantine.archway.tech:443",
		},
	}

	cfg := Config{RPCs: rpcs}

	exp := map[string]RPC{
		"archway": {
			ChainName: "archway",
			ChainID:   "archway-1",
			URL:       "https://rpc.mainnet.archway.io:443",
		},
		"archwaytestnet": {
			ChainName: "archwaytestnet",
			ChainID:   "constantine-3",
			URL:       "https://rpc.constantine.archway.tech:443",
		},
	}

	res := cfg.GetRPCsMap()

	assert.Equal(t, &exp, res)
}

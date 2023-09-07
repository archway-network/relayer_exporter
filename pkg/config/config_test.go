package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRPCsMap(t *testing.T) {
	rpcs := []RPC{
		{
			ChainID: "archway-1",
			URL:     "https://rpc.mainnet.archway.io:443",
		},
		{
			ChainID: "constantine-3",
			URL:     "https://rpc.constantine.archway.tech:443",
		},
	}

	cfg := Config{RPCs: rpcs}

	exp := map[string]string{
		"archway-1":     "https://rpc.mainnet.archway.io:443",
		"constantine-3": "https://rpc.constantine.archway.tech:443",
	}

	res := cfg.GetRPCsMap()

	assert.Equal(t, exp, res)
}

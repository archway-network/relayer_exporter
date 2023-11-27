package config

import (
	"errors"
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
			Timeout:   "2s",
		},
	}

	cfg := Config{
		GlobalRPCTimeout: "5s",
		RPCs:             rpcs,
	}

	exp := map[string]RPC{
		"archway": {
			ChainName: "archway",
			ChainID:   "archway-1",
			URL:       "https://rpc.mainnet.archway.io:443",
			Timeout:   "5s",
		},
		"archwaytestnet": {
			ChainName: "archwaytestnet",
			ChainID:   "constantine-3",
			URL:       "https://rpc.constantine.archway.tech:443",
			Timeout:   "2s",
		},
	}

	res := cfg.GetRPCsMap()

	assert.Equal(t, &exp, res)
}

func TestGetPaths(t *testing.T) {
	cfg := Config{}

	expError := ErrGitHubClient

	_, err := cfg.getPaths("_IBC", nil)
	if err == nil {
		t.Fatalf("Expected error %q, got no error", expError)
	}

	if !errors.Is(err, expError) {
		t.Errorf("Expected error %q, got %q", expError, err)
	}
}

package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRPC(t *testing.T) {
	testCases := []struct {
		name  string
		rpcs  []RPC
		paths []*IBCData
		resp  map[string]RPC
		err   error
	}{
		{
			name: "No Missing or Invalid RPCs",
			rpcs: []RPC{
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
			},
			paths: []*IBCData{},
			resp: map[string]RPC{
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
			},
			err: nil,
		},
		{
			name: "Missing RPCs",
			rpcs: []RPC{
				{
					ChainName: "archway",
					ChainID:   "archway-1",
					URL:       "https://rpc.mainnet.archway.io:443",
				},
			},
			paths: []*IBCData{
				{
					Schema: "$schema",
					Chain1: IBCChainMeta{
						ChainName:    "archway",
						ClientID:     "Client1",
						ConnectionID: "Connection1",
					},
					Chain2: IBCChainMeta{
						ChainName:    "archwaytestnet",
						ClientID:     "Client2",
						ConnectionID: "Connection2",
					},
					Channels:  []Channel{},
					Operators: []Operator{},
				},
			},
			resp: map[string]RPC{
				"archway": {
					ChainName: "archway",
					ChainID:   "archway-1",
					URL:       "https://rpc.mainnet.archway.io:443",
					Timeout:   "5s",
				},
			},
			err: fmt.Errorf(ErrMissingRPCConfigMsg, "archwaytestnet"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				GlobalRPCTimeout: "5s",
				RPCs:             tc.rpcs,
			}
			res, err := cfg.GetRPCsMap(tc.paths)
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, &tc.resp, res)
		})
	}
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

package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/archway-network/relayer_exporter/pkg/config"
)

func TestDiscordIDs(t *testing.T) {
	testCases := []struct {
		name     string
		ops      []config.Operator
		expected string
	}{
		{
			name: "All Valid IDs",
			ops: []config.Operator{
				{
					Discord: config.Discord{ID: "123456"},
				},
				{
					Discord: config.Discord{ID: "12312387"},
				},
			},
			expected: "123456,12312387",
		},
		{
			name: "Some Invalid IDs",
			ops: []config.Operator{
				{
					Discord: config.Discord{ID: "123456"},
				},
				{
					Discord: config.Discord{ID: "ABCDEF"},
				},
				{
					Discord: config.Discord{ID: "789012"},
				},
			},
			expected: "123456,789012",
		},
		{
			name: "No Valid IDs",
			ops: []config.Operator{
				{Discord: config.Discord{ID: "ABCDEF"}},
				{Discord: config.Discord{ID: "GHIJKL"}},
			},
			expected: "",
		},
		{
			name:     "Empty Input",
			ops:      []config.Operator{{}, {}, {}},
			expected: "",
		},
	}
	for _, tc := range testCases {
		res := getDiscordIDs(tc.ops)
		assert.Equal(t, tc.expected, res)
	}
}

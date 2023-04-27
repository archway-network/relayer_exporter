package relayer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseChains(t *testing.T) {
	outStr := ` 1: theta-testnet-001    -> type(cosmos) key(✔) bal(✔) path(✔)
 2: osmo-test-5          -> type(cosmos) key(✔) bal(✔) path(✔)
 3: constantine-2        -> type(cosmos) key(✔) bal(✔) path(✔)
 4: axelar-testnet-lisbon-3 -> type(cosmos) key(✔) bal(✔) path(✔)
`
	exp := []string{
		"theta-testnet-001",
		"osmo-test-5",
		"constantine-2",
		"axelar-testnet-lisbon-3",
	}

	var b bytes.Buffer
	b.WriteString(outStr)

	res, err := parseChains(&b)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestParsePaths(t *testing.T) {
	outStr := ` 0: archwaytestnet-axelartestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>axelar-testnet-lisbon-3)
 1: archwaytestnet-cosmoshubtestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>theta-testnet-001)
 2: archwaytestnet-osmosistestnet -> chns(✔) clnts(✔) conn(✔) (constantine-2<>osmo-test-4)
`
	exp := []string{
		"archwaytestnet-axelartestnet",
		"archwaytestnet-cosmoshubtestnet",
		"archwaytestnet-osmosistestnet",
	}

	var b bytes.Buffer
	b.WriteString(outStr)

	res, err := parsePaths(&b)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestParseClientsForPath(t *testing.T) {
	path := "archwaytestnet-axelartestnet"
	outStr := `client 07-tendermint-102 (constantine-2) expires in 3h13m11s (14 Apr 23 18:16 UTC)
client 07-tendermint-401 (axelar-testnet-lisbon-3) expires in 3h13m52s (14 Apr 23 18:17 UTC)
`

	exp := []Client{
		{ChainID: "constantine-2", Path: path, ExpiresAt: time.Unix(1681496160, 0).UTC()},
		{ChainID: "axelar-testnet-lisbon-3", Path: path, ExpiresAt: time.Unix(1681496220, 0).UTC()},
	}

	var b bytes.Buffer
	b.WriteString(outStr)

	res, err := parseClientsForPath(path, &b)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}

func TestGetClients(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	exp := []Client{
		{ChainID: "constantine-2", Path: "archwaytestnet-cosmoshubtestnet", ExpiresAt: time.Unix(1682508960, 0).UTC()},
		{ChainID: "theta-testnet-001", Path: "archwaytestnet-cosmoshubtestnet", ExpiresAt: time.Unix(1682509020, 0).UTC()},
		{ChainID: "constantine-2", Path: "archwaytestnet-axelartestnet", ExpiresAt: time.Unix(1682503620, 0).UTC()},
		{ChainID: "axelar-testnet-lisbon-3", Path: "archwaytestnet-axelartestnet", ExpiresAt: time.Unix(1682503740, 0).UTC()},
	}

	rlyPath := filepath.Join(dir, "..", "..", "testdata", "rly.sh")
	res, err := GetClients(rlyPath)
	assert.NoError(t, err)

	assert.Equal(t, exp, res)
}

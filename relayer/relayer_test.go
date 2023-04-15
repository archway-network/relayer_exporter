package relayer

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestGetPaths(t *testing.T) {
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

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, exp) {
		t.Errorf("Expected %v, got %v instead.\n", exp, res)
	}
}

func TestGetClients(t *testing.T) {
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

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, exp) {
		t.Errorf("Expected %v, got %v instead.\n", exp, res)
	}
}

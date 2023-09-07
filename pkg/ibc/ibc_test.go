package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathName(t *testing.T) {
	clientsInfo := ClientsInfo{}
	assert.Equal(t, "<->", clientsInfo.PathName())
}

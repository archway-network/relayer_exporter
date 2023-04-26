package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	exp := "version: dev commit: none date: unknown"
	res := getVersion()

	assert.Equal(t, exp, res)
}

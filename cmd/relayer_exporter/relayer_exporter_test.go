package main

import "testing"

func TestGetVersion(t *testing.T) {
	exp := "version: dev commit: none date: unknown"
	res := getVersion()

	if res != exp {
		t.Errorf("Expected %q, got %q instead.\n", exp, res)
	}
}

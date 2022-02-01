package main

import (
	"os"
	"testing"
)

func TestEnvoke(t *testing.T) {
	if len(os.Args) > 1 {
		os.Args[1] = "testdata/envoy-config"
	} else {
		os.Args = append(os.Args, "testdata/envoy-config")
	}
	Envoke()
}

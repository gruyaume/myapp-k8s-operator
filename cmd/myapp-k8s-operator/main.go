package main

import (
	"myapp-k8s-operator/internal/charm"
	"os"

	"github.com/gruyaume/goops"
)

func main() {
	err := charm.Configure()
	if err != nil {
		goops.LogErrorf("Failed to configure charm: %v", err)
		os.Exit(1)
	}
}

// Package main is the entry point for the Open Targets deployment tool.
package main

import (
	"log"

	"github.com/opentargets/platform-deployment-standalone/cmd"
)

func main() {
	log.SetFlags(0)

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

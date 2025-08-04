// Package standalone is a command-line tool for configuring lightweight instances of the Open Targets platform.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/opentargets/lightweight/internal/config"
)

func main() {
	lipgloss.SetColorProfile(termenv.TrueColor)

	defaultsFilePath, err := filepath.Abs("./etc/defaults")
	if err != nil {
		log.Fatal(err)
	}

	// present the configuration form
	c := config.New(defaultsFilePath)
	f := config.Form(&c)
	err = f.Run()
	if err != nil {
		log.Fatal(err)
	}

	// set the deployment folder based on the configuration
	c.SetDeploymentFolder()

	// create the deployment folder and write the configuration files to it
	err = os.MkdirAll(c.DeploymentFolder.Value, 0755)
	if err != nil {
		log.Fatalf("failed to create deployment folder %s: %v", c.DeploymentFolder.Value, err)
	}
	c.WriteToFile(c.DeploymentFolder.Value + "/config")
	c.WriteSecrets(c.DeploymentFolder.Value)

	fmt.Print(c.DeploymentFolder.Value)
}

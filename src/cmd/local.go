package cmd

import (
	"log"

	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/housekeeping"
)

// RunLocal runs the local deployment setup.
func RunLocal(auto bool, configPath string) {
	// 1. Load defaults
	c, err := config.NewLocalDeploymentConfig(configPath)
	if err != nil {
		log.Fatalf("error loading config file: %v\n", err)
	}

	// 2. Parse env vars
	c.ReplaceFromEnv()

	// 3. If non-interactive mode, validate the config and exit if there are errors.
	// Otherwise, present the configuration form.
	if auto {
		err = c.Validate()
		if err != nil {
			log.Fatalf("bad configuration: %v\n", err)
		}
	} else {
		cf := config.LocalDeploymentForm(c)
		err = cf.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// 4. Prepare deployment directory
	housekeeping.PrepareDeploymentDir(c)
	housekeeping.WriteConfig(c)

	// 5. Run deployment
	housekeeping.DeployLocal(c)
}

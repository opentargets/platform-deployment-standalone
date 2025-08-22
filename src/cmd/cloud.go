package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/housekeeping"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// RunCloud runs the cloud deployment setup.
func RunCloud(auto bool, configPath string) {
	// 1. Load defaults
	c, err := config.NewCloudDeploymentConfig(configPath)
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
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
		cf := config.CloudDeploymentForm(c)
		err = cf.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// 4. Print the configuration to the console, and if interactive, request confirmation.
	log.Printf("%s\n", c.ToString())
	if !auto {
		var proceed bool
		pf := config.ConfirmationForm(&proceed)
		err = pf.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
		if !proceed {
			log.Fatal("exiting without deploying")
		}
	}

	// 5. Prepare deployment directory
	housekeeping.PrepareDeploymentDir(c)
	housekeeping.WriteConfig(c)

	// 6. Run deployment
	action := func() {
		housekeeping.DeployCloud(c)
	}
	tools.RunWithSpinner("deploying", action)

	// 8. Upload the configuration file to GCS
	err = housekeeping.UploadConfig(c)
	if err != nil {
		log.Printf("error uploading configuration file to ops uri: %v\n", err)
	}

	// 7. Show success message
	log.Println("Deployment completed successfully! Instance available at:")
	log.Printf("·  https://%s.%s\n", c.SubdomainName.Value, c.DomainName.Value)
	log.Printf("·  https://%s.%s/api\n", c.SubdomainName.Value, c.DomainName.Value)
}

// ListCloud lists cloud deployments.
func ListCloud(backend string) {
	ok := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ff00")).Render("✔")
	ko := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff0000")).Render("✘")
	em := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#777777")).Render(" — ")

	files := []string{}
	statuses := strings.Builder{}

	getCloudDeployments := func() {
		var err error
		files, err = tools.ListFilesInGCSPrefix(backend)
		if err != nil {
			log.Fatalf("error listing files in ops uri: %v\n", err)
		}
	}
	tools.RunWithSpinner("getting cloud deployments", getCloudDeployments)

	if len(files) == 0 {
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff0000")).Render("No deployments found.")
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f, ".tfstate") {
			configFilename := fmt.Sprintf("%s/%s", backend, f)

			var url, status string
			checkInstance := func() {
				url, status = housekeeping.CheckInstance(configFilename)
			}

			tools.RunWithSpinner(fmt.Sprintf("checking instance %s", f), checkInstance)

			parts := strings.Split(f, "/")
			name := lipgloss.NewStyle().Width(16).Align(lipgloss.Left).Bold(true).Render(parts[len(parts)-1])
			url = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#3366cc")).Render(url)

			if status == "live" {
				statuses.WriteString(ok)
			} else {
				statuses.WriteString(ko)
			}
			statuses.WriteString(em)
			statuses.WriteString(name)
			statuses.WriteString(em)
			if status == "live" {
				statuses.WriteString(url)
			} else {
				statuses.WriteString(status)
			}
			statuses.WriteString("\n")
		}
	}

	fmt.Printf("%s\n", statuses.String())
}

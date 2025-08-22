package config

import (
	"github.com/charmbracelet/huh"
)

// LocationForm creates a form for selecting the deployment location.
func LocationForm(location *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(
					huh.NewOption("Local deployment", "local"),
					huh.NewOption("Cloud deployment", "cloud"),
				).
				Value(location),
		).
			Title("Deployment location").
			Description("Select the type of deployment you want to configure. Local deployments are hosted on your machine, while cloud deployments are hosted on Google Cloud Platform."),
	)
}

// DeploymentForm creates a form for configuring a deployment.
func DeploymentForm(location string, c *DeploymentConfig) *huh.Form {
	switch location {
	case "local":
		return LocalDeploymentForm((*c).(*LocalDeploymentConfig))
	case "cloud":
		return CloudDeploymentForm((*c).(*CloudDeploymentConfig))
	}
	return nil
}

// ConfirmationForm creates a form for confirming the deployment configuration.
func ConfirmationForm(proceed *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Is the configuration correct?").
				Description("Inspect the config above and confirm the deployment.").
				Affirmative("Yes").
				Negative("No").
				Value(proceed),
		),
	)
}

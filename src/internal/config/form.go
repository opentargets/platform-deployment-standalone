package config

import (
	"github.com/charmbracelet/huh"
)

// Form creates a form for configuring a deployment.
func Form(c *Config) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(
					huh.NewOption("Local deployment", (string)(Local)),
					huh.NewOption("Cloud deployment", (string)(Cloud)),
				).
				Value((*string)(&c.Location.Value)),
		).
			Title("Deployment Type"),

		huh.NewGroup(
			huh.NewInput().
				Title("Data release").
				Description("The release should be in the form YY.MM, e.g. 25.06 for the June 2025 release.").
				Value(&c.Release.Value),
			huh.NewInput().
				Title("API docker image name").
				Value(&c.APIImage.Value),

			huh.NewInput().
				Title("API image tag").
				Description("Check available tags at https://github.com/opentargets/platform-api/pkgs/container/platform-api").
				Value(&c.APITag.Value),

			huh.NewInput().
				Title("OpenAI API image name").
				Value(&c.APIOpenAIImage.Value),

			huh.NewInput().
				Title("OpenAI API image tag").
				Description("Check available tags at https://github.com/opentargets/ot-ai-api/pkgs/container/ot-ai-api").
				Value(&c.APIOpenAITag.Value),

			huh.NewInput().
				Title("WebApp image name").
				Value(&c.WebAppImage.Value),

			huh.NewInput().
				Title("WebApp image tag").
				Description("Check available tags at https://github.com/opentargets/ot-ui-apps/pkgs/container/ot-ui-apps").
				Value(&c.WebAppTag.Value),

			huh.NewInput().
				Title("Opensearch image tag").
				Value(&c.OpensearchTag.Value),

			huh.NewInput().
				Title("ClickHouse image tag").
				Value(&c.ClickhouseTag.Value),
		).
			Title("Data and software versions").
			Description("Here you can choose the data release and software versions to deploy. Defaults are set to the latest release and stable versions of the different software components."),

		huh.NewGroup(
			huh.NewInput().
				Title("Release URL").
				Description("The base URL where release data is fetched from (without the actual release). Example sources: https://ftp.ebi.ac.uk/pub/databases/opentargets/platform and gs://open-targets-data-releases").
				Value(&c.ReleaseURL.Value).
				CharLimit(2048).
				Validate(huh.ValidateNotEmpty()).
				Validate(ValidateURL),

			huh.NewInput().
				Title("OpenAI API token").
				Value(&c.APIOpenAIToken.Value),
		).
			Title("Local Deployment Settings").
			WithHideFunc(func() bool {
				return Location(c.Location.Value) != Local
			}),

		huh.NewGroup(
			huh.NewInput().
				Title("Domain name").
				Value(&c.DomainName.Value),

			huh.NewInput().
				Title("Subdomain name").
				Value(&c.SubdomainName.Value).
				Validate(ValidateSubdomain(&c.DomainName.Value)),

			huh.NewInput().
				Title("Days to live").
				Description("The deployment will be destroyed after this many days.").
				Value(&c.DaysToLive.Value).
				Validate(ValidateDaysToLive(&c.SubdomainName.Value)),

			huh.NewInput().
				Title("GCP Project").
				Value(&c.GCPProject.Value).
				Validate(ValidateGCPResource()),

			huh.NewInput().
				Title("GCP Region").
				Value(&c.GCPRegion.Value).
				Validate(ValidateGCPResource()),

			huh.NewInput().
				Title("GCP Zone").
				Value(&c.GCPZone.Value).
				Validate(ValidateGCPResource()),

			huh.NewInput().
				Title("GCP Secret for OpenAI API token").
				Value(&c.GCPSecretOpenAIToken.Value).
				Validate(ValidateGCPResource()),

			huh.NewInput().
				Title("GCP Cloud DNS Zone").
				Value(&c.GCPCloudDNSZone.Value).
				Validate(ValidateGCPResource()),
		).
			Title("Cloud Deployment Settings").
			WithHideFunc(func() bool {
				return Location(c.Location.Value) != Cloud
			}),
	)
}

package config

import (
	"github.com/charmbracelet/huh"
)

// Form creates a form for configuring a deployment.
func Form(c *Config) *huh.Form {
	APIImage := huh.NewInput().
		Title("API docker image name").
		Value(&c.APIImage.Value)

	APITag := huh.NewInput().
		Title("API image tag").
		Description("Check available tags at https://github.com/opentargets/platform-api/pkgs/container/platform-api").
		Value(&c.APITag.Value)

	APIOpenAIImage := huh.NewInput().
		Title("OpenAI API image name").
		Value(&c.APIOpenAIImage.Value)

	APIOpenAITag := huh.NewInput().
		Title("OpenAI API image tag").
		Description("Check available tags at https://github.com/opentargets/ot-ai-api/pkgs/container/ot-ai-api").
		Value(&c.APIOpenAITag.Value)

	WebAppImage := huh.NewInput().
		Title("WebApp image name").
		Value(&c.WebAppImage.Value)

	WebAppTag := huh.NewInput().
		Title("WebApp image tag").
		Description("Check available tags at https://github.com/opentargets/ot-ui-apps/pkgs/container/ot-ui-apps").
		Value(&c.WebAppTag.Value)

	ClickhouseTag := huh.NewInput().
		Title("ClickHouse image tag").
		Value(&c.ClickhouseTag.Value)

	OpensearchTag := huh.NewInput().
		Title("Opensearch image tag").
		Value(&c.OpensearchTag.Value)

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

		// Local deployment software and data versions group.
		huh.NewGroup(
			huh.NewInput().
				Title("Release").
				Description("The data release name should be in the form YY.MM, e.g. 25.06 for the June 2025 release.").
				Value(&c.Release.Value),
			huh.NewInput().
				Title("Release URL").
				Description("The base URL where release data is fetched from (without the actual release). Example sources: https://ftp.ebi.ac.uk/pub/databases/opentargets/platform and gs://open-targets-data-releases").
				Value(&c.ReleaseURL.Value).
				CharLimit(2048).
				Validate(huh.ValidateNotEmpty()).
				Validate(ValidateURL),
			APIImage,
			APITag,
			APIOpenAIImage,
			APIOpenAITag,
			huh.NewInput().
				Title("OpenAI API token").
				Description("The OpenAI API token to use for the publication summarization feature.").
				Value(&c.APIOpenAIToken.Value),
			WebAppImage,
			WebAppTag,
			ClickhouseTag,
			OpensearchTag,
		).
			Title("Local deployment settings").
			Description("Defaults are set to the latest public release and stable versions of the different software components.").
			WithHideFunc(func() bool {
				return Location(c.Location.Value) != Local
			}),

		// Cloud deployment software and data versions group.
		huh.NewGroup(
			huh.NewInput().
				Title("ClickHouse data snapshot").
				Description("The name of the snapshot for the ClickHouse database.").
				Value(&c.SnapshotCH.Value).
				Validate(ValidateGCPSnapshot),
			huh.NewInput().
				Title("OpenSearch data snapshot").
				Description("The name of the snapshot for the OpenSearch database.").
				Value(&c.SnapshotOS.Value).
				Validate(ValidateGCPSnapshot),
			huh.NewSelect[string]().
				Options(
					huh.NewOption("Platform flavour", "platform"),
					huh.NewOption("PPP flavour", "ppp"),
				).
				Title("Deployment flavour").
				Description("The flavour of the deployment, either platform or ppp. This will determine the web application used.").
				Value(&c.Flavor.Value),
			APIImage,
			APITag,
			APIOpenAIImage,
			APIOpenAITag,
			WebAppImage,
			WebAppTag,
			ClickhouseTag,
			OpensearchTag,
		).
			Title("Cloud deployment settings for data and software versions").
			Description("Defaults are set to the latest stable release snapshots and stable versions of the different software components.").
			WithHideFunc(func() bool {
				return Location(c.Location.Value) != Cloud
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
				Description("The deployment will be destroyed after this many days (input 0 for no expiry)").
				Value(&c.DaysToLive.Value).
				Validate(ValidateDaysToLive),
			huh.NewInput().
				Title("GCP Secret for OpenAI API token").
				Description("The GCP Secret containing the OpenAI API token to use for the publication summarization feature.").
				Value(&c.GCPSecretOpenAIToken.Value).
				Validate(ValidateGCPSecret),
			huh.NewInput().
				Title("GCP Project").
				Value(&c.GCPProject.Value).
				Validate(ValidateGCPProject),
			huh.NewInput().
				Title("GCP Region").
				Value(&c.GCPRegion.Value).
				Validate(ValidateGCPResource),
			huh.NewInput().
				Title("GCP Zone").
				Value(&c.GCPZone.Value).
				Validate(ValidateGCPResource),
			huh.NewInput().
				Title("GCP Cloud DNS Zone").
				Value(&c.GCPCloudDNSZone.Value).
				Validate(ValidateGCPCloudDNSZone),
			huh.NewInput().
				Title("GCP Network").
				Description("The name of the GCP Network.").
				Value(&c.GCPNetwork.Value).
				Validate(ValidateGCPNetwork(&c.Flavor.Value)),
			huh.NewInput().
				Title("GCP Service Account").
				Description("The service account to use for the deployment. Refer to the README for more information.").
				Value(&c.GCPServiceAccount.Value),
		).
			Title("Cloud Deployment Settings").
			WithHideFunc(func() bool {
				return Location(c.Location.Value) != Cloud
			}),
	)
}

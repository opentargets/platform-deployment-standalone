package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// defaultsCloudPath is the default path to the cloud deployment configuration file.
const defaultsCloudPath = "./etc/defaults-cloud"

// CloudDeploymentMaxDaysToLive is the maximum number of days a cloud deployment can live.
const CloudDeploymentMaxDaysToLive = 14

// CloudDeploymentConfig holds the configuration for a cloud deployment.
type CloudDeploymentConfig struct {
	DeploymentType    Setting
	GCPProject        Setting
	GCPRegion         Setting
	GCPZone           Setting
	OpsURI            Setting
	DomainName        Setting
	SubdomainName     Setting
	DaysToLive        Setting
	WebAppFlavor      Setting
	SnapshotCH        Setting
	SnapshotOS        Setting
	APIImage          Setting
	APITag            Setting
	APIAIImage        Setting
	APIAITag          Setting
	WebAppImage       Setting
	WebAppTag         Setting
	ClickhouseTag     Setting
	OpensearchTag     Setting
	GCPSecretAIToken  Setting
	GCPCloudDNSZone   Setting
	GCPNetwork        Setting
	GCPServiceAccount Setting
	APICache          Setting
}

// NewCloudDeploymentConfig creates a new CloudDeploymentConfig with defaults.
func NewCloudDeploymentConfig(configPath string) (*CloudDeploymentConfig, error) {
	effectivePath := defaultsCloudPath
	if configPath != "" {
		effectivePath = configPath
	}

	env, err := tools.LoadEnvFromFile(effectivePath)
	if err != nil {
		return nil, err
	}
	if env["OT_DEPLOYMENT_TYPE"] == "" {
		return nil, fmt.Errorf("config file does not contain OT_DEPLOYMENT_TYPE setting")
	}
	if env["OT_DEPLOYMENT_TYPE"] != "cloud" {
		return nil, fmt.Errorf("config file is for deployment type '%s', not 'cloud'", env["OT_DEPLOYMENT_TYPE"])
	}

	config := &CloudDeploymentConfig{
		DeploymentType: Setting{
			Env:   "OT_DEPLOYMENT_TYPE",
			Value: "cloud",
		},

		// First form: GCP global settings
		GCPProject: Setting{
			Title:     "GCP Project",
			Env:       "TF_VAR_OT_GCP_PROJECT",
			Value:     env["TF_VAR_OT_GCP_PROJECT"],
			Validator: ValidateGCPProject,
		},
		GCPRegion: Setting{
			Title: "GCP Region",
			Env:   "TF_VAR_OT_GCP_REGION",
			Value: env["TF_VAR_OT_GCP_REGION"],
		},
		GCPZone: Setting{
			Title: "GCP Zone",
			Env:   "TF_VAR_OT_GCP_ZONE",
			Value: env["TF_VAR_OT_GCP_ZONE"],
		},
		OpsURI: Setting{
			Title:       "Ops URI",
			Description: "The URI where the deployment config and state will be persisted. This will be used as terraform backend.",
			Env:         "OT_OPS_URI",
			Value:       env["OT_OPS_URI"],
			Validator:   ValidateGCSBucket,
		},

		// Second form: Deployment settings
		DomainName: Setting{
			Title:     "Domain name",
			Env:       "TF_VAR_OT_DOMAIN_NAME",
			Value:     env["TF_VAR_OT_DOMAIN_NAME"],
			Validator: ValidateDomainName,
		},
		SubdomainName: Setting{
			Title:       "Subdomain name",
			Description: "Subdomains should be only one level deep and contain only lowercase letters, numbers, and hyphens.",
			Env:         "TF_VAR_OT_SUBDOMAIN_NAME",
			Value:       tools.Either(env["TF_VAR_OT_SUBDOMAIN_NAME"], tools.RandomString(4)),
			Validator:   ValidateSubdomainName,
		},
		DaysToLive: Setting{
			Title:       "Days to live",
			Description: "The deployment will be destroyed after this many days (0 for no expiry)",
			Env:         "TF_VAR_OT_DAYS_TO_LIVE",
			Value:       env["TF_VAR_OT_DAYS_TO_LIVE"],
			Validator:   ValidateDaysToLive,
		},
		WebAppFlavor: Setting{
			Title:       "Web App flavour",
			Description: "The flavor of the web application: `platform` or `ppp` partner preview (only available internally).",
			Env:         "OT_WEBAPP_FLAVOR",
			Value:       env["OT_WEBAPP_FLAVOR"],
			Validator:   ValidateWebAppFlavor,
		},

		// Third form: Data snapshots
		SnapshotCH: Setting{
			Title: "ClickHouse data snapshot",
			Env:   "TF_VAR_OT_SNAPSHOT_CH",
			Value: env["TF_VAR_OT_SNAPSHOT_CH"],
		},
		SnapshotOS: Setting{
			Title: "OpenSearch data snapshot",
			Env:   "TF_VAR_OT_SNAPSHOT_OS",
			Value: env["TF_VAR_OT_SNAPSHOT_OS"],
		},

		// Fourth form: Software versions
		APIImage: Setting{
			Title:     "API docker image name",
			Env:       "OT_API_IMAGE",
			Value:     env["OT_API_IMAGE"],
			Validator: ValidateImageName,
		},
		APITag: Setting{
			Title:       "API docker image tag",
			Description: "Check available tags at https://github.com/opentargets/platform-api/pkgs/container/platform-api",
			Env:         "OT_API_TAG",
			Value:       env["OT_API_TAG"],
		},
		APIAIImage: Setting{
			Title:     "AI API docker image name",
			Env:       "OT_API_AI_IMAGE",
			Value:     env["OT_API_AI_IMAGE"],
			Validator: ValidateImageName,
		},
		APIAITag: Setting{
			Title:       "AI API docker image tag",
			Description: "Check available tags at https://github.com/opentargets/ot-ai-api/pkgs/container/ot-ai-api",
			Env:         "OT_API_AI_TAG",
			Value:       env["OT_API_AI_TAG"],
		},
		WebAppImage: Setting{
			Title:     "WebApp docker image name",
			Env:       "OT_WEBAPP_IMAGE",
			Value:     env["OT_WEBAPP_IMAGE"],
			Validator: ValidateImageName,
		},
		WebAppTag: Setting{
			Title:       "WebApp docker image tag",
			Description: "Check available tags at at https://github.com/opentargets/ot-ui-apps/pkgs/container/ot-ui-apps",
			Env:         "OT_WEBAPP_TAG",
			Value:       env["OT_WEBAPP_TAG"],
		},
		ClickhouseTag: Setting{
			Title:     "ClickHouse docker image tag",
			Env:       "OT_CLICKHOUSE_TAG",
			Value:     env["OT_CLICKHOUSE_TAG"],
			Validator: ValidateNotEmpty,
		},
		OpensearchTag: Setting{
			Title:     "Opensearch docker image tag",
			Env:       "OT_OPENSEARCH_TAG",
			Value:     env["OT_OPENSEARCH_TAG"],
			Validator: ValidateNotEmpty,
		},

		// Fifth form: Additional settings
		GCPSecretAIToken: Setting{
			Title:       "GCP AI API token secret",
			Description: "The Google Cloud Secret Manager secret that contains the API token to use inside the AI API for the publication summarization feature.",
			Env:         "TF_VAR_OT_GCP_SECRET_AI_TOKEN",
			Value:       env["TF_VAR_OT_GCP_SECRET_AI_TOKEN"],
		},
		GCPCloudDNSZone: Setting{
			Title: "GCP Cloud DNS Zone",
			Env:   "TF_VAR_OT_GCP_CLOUD_DNS_ZONE",
			Value: env["TF_VAR_OT_GCP_CLOUD_DNS_ZONE"],
		},
		GCPNetwork: Setting{
			Title: "GCP Network",
			Env:   "TF_VAR_OT_GCP_NETWORK",
			Value: env["TF_VAR_OT_GCP_NETWORK"],
		},
		GCPServiceAccount: Setting{
			Title:       "GCP Service Account",
			Description: "Input in email form, e.g. `service-account@project.iam.gserviceaccount.com`.",
			Env:         "TF_VAR_OT_GCP_SA",
			Value:       env["TF_VAR_OT_GCP_SA"],
		},
		APICache: Setting{
			Title:       "API cache",
			Description: "Whether the API should use caching (recommended) or not. Disable for development purposes.",
			Env:         "PLATFORM_API_IGNORE_CACHE",
			Value:       tools.Either(env["PLATFORM_API_IGNORE_CACHE"], "false"),
		},
	}

	// Set callbacks for validators that require other settings' values.
	config.GCPRegion.Validator = ValidateGCPRegion(func() string { return config.GCPProject.Value })
	config.GCPZone.Validator = ValidateGCPZone(func() string { return config.GCPProject.Value })

	config.SnapshotCH.Validator = ValidateGCPSnapshot(func() string { return config.GCPProject.Value })
	config.SnapshotOS.Validator = ValidateGCPSnapshot(func() string { return config.GCPProject.Value })

	config.APITag.Validator = ValidateVersionTag(func() string { return config.APIImage.Value })
	config.APIAITag.Validator = ValidateVersionTag(func() string { return config.APIAIImage.Value })
	config.WebAppTag.Validator = ValidateVersionTag(func() string { return config.WebAppImage.Value })

	config.GCPSecretAIToken.Validator = ValidateGCPSecret(func() string { return config.GCPProject.Value })
	config.GCPCloudDNSZone.Validator = ValidateGCPCloudDNSZone(func() string { return config.GCPProject.Value })
	config.GCPNetwork.Validator = ValidateGCPNetwork(func() string { return config.WebAppFlavor.Value }, func() string { return config.GCPProject.Value })
	config.GCPServiceAccount.Validator = ValidateGCPServiceAccount(func() string { return config.GCPProject.Value })

	return config, nil
}

// GetDeploymentDir returns the directory where the cloud deployment files are stored.
func (c *CloudDeploymentConfig) GetDeploymentDir() string {
	return "deployment-cloud-" + c.SubdomainName.Value
}

// Validate validates all settings in CloudDeploymentSettings.
func (c *CloudDeploymentConfig) Validate() error {
	var errs []error
	tools.AppendIfErr(&errs, c.GCPProject.Validate())
	tools.AppendIfErr(&errs, c.GCPRegion.Validate())
	tools.AppendIfErr(&errs, c.GCPZone.Validate())
	tools.AppendIfErr(&errs, c.OpsURI.Validate())
	tools.AppendIfErr(&errs, c.DomainName.Validate())
	tools.AppendIfErr(&errs, c.SubdomainName.Validate())
	tools.AppendIfErr(&errs, c.DaysToLive.Validate())
	tools.AppendIfErr(&errs, c.WebAppFlavor.Validate())
	tools.AppendIfErr(&errs, c.SnapshotCH.Validate())
	tools.AppendIfErr(&errs, c.SnapshotOS.Validate())
	tools.AppendIfErr(&errs, c.APIImage.Validate())
	tools.AppendIfErr(&errs, c.APITag.Validate())
	tools.AppendIfErr(&errs, c.APIAIImage.Validate())
	tools.AppendIfErr(&errs, c.APIAITag.Validate())
	tools.AppendIfErr(&errs, c.WebAppImage.Validate())
	tools.AppendIfErr(&errs, c.WebAppTag.Validate())
	tools.AppendIfErr(&errs, c.ClickhouseTag.Validate())
	tools.AppendIfErr(&errs, c.OpensearchTag.Validate())
	tools.AppendIfErr(&errs, c.GCPSecretAIToken.Validate())
	tools.AppendIfErr(&errs, c.GCPCloudDNSZone.Validate())
	tools.AppendIfErr(&errs, c.GCPNetwork.Validate())
	tools.AppendIfErr(&errs, c.GCPServiceAccount.Validate())
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}
	return nil
}

// ReplaceFromEnv replaces the values of the CloudDeploymentConfig from environment.
func (c *CloudDeploymentConfig) ReplaceFromEnv() {
	c.GCPProject.ReplaceFromEnv()
	c.GCPRegion.ReplaceFromEnv()
	c.GCPZone.ReplaceFromEnv()
	c.OpsURI.ReplaceFromEnv()
	c.DomainName.ReplaceFromEnv()
	c.SubdomainName.ReplaceFromEnv()
	c.DaysToLive.ReplaceFromEnv()
	c.WebAppFlavor.ReplaceFromEnv()
	c.SnapshotCH.ReplaceFromEnv()
	c.SnapshotOS.ReplaceFromEnv()
	c.APIImage.ReplaceFromEnv()
	c.APITag.ReplaceFromEnv()
	c.APIAIImage.ReplaceFromEnv()
	c.APIAITag.ReplaceFromEnv()
	c.WebAppImage.ReplaceFromEnv()
	c.WebAppTag.ReplaceFromEnv()
	c.ClickhouseTag.ReplaceFromEnv()
	c.OpensearchTag.ReplaceFromEnv()
	c.GCPSecretAIToken.ReplaceFromEnv()
	c.GCPCloudDNSZone.ReplaceFromEnv()
	c.GCPNetwork.ReplaceFromEnv()
	c.GCPServiceAccount.ReplaceFromEnv()
	c.APICache.ReplaceFromEnv()
}

// ToString returns a string representation of the CloudDeploymentConfig.
func (c *CloudDeploymentConfig) ToString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Open Targets cloud deployment config for https://%s.%s\n", c.SubdomainName.Value, c.DomainName.Value))
	sb.WriteString(c.DeploymentType.ToString())
	sb.WriteString("\n# GCP global settings\n")
	sb.WriteString(c.GCPProject.ToString())
	sb.WriteString(c.GCPRegion.ToString())
	sb.WriteString(c.GCPZone.ToString())
	sb.WriteString(c.OpsURI.ToString())
	sb.WriteString("\n# Deployment settings\n")
	sb.WriteString(c.DomainName.ToString())
	sb.WriteString(c.SubdomainName.ToString())
	sb.WriteString(c.DaysToLive.ToString())
	sb.WriteString(c.WebAppFlavor.ToString())
	sb.WriteString("\n# Data snapshots\n")
	sb.WriteString(c.SnapshotCH.ToString())
	sb.WriteString(c.SnapshotOS.ToString())
	sb.WriteString("\n# Software versions\n")
	sb.WriteString(c.APIImage.ToString())
	sb.WriteString(c.APITag.ToString())
	sb.WriteString(c.APIAIImage.ToString())
	sb.WriteString(c.APIAITag.ToString())
	sb.WriteString(c.WebAppImage.ToString())
	sb.WriteString(c.WebAppTag.ToString())
	sb.WriteString(c.ClickhouseTag.ToString())
	sb.WriteString(c.OpensearchTag.ToString())
	sb.WriteString("\n# Additional settings\n")
	sb.WriteString(c.GCPSecretAIToken.ToString())
	sb.WriteString(c.GCPCloudDNSZone.ToString())
	sb.WriteString(c.GCPNetwork.ToString())
	sb.WriteString(c.GCPServiceAccount.ToString())
	sb.WriteString(c.APICache.ToString())
	return sb.String()
}

// GetSecretFields returns a slice of Settings that are secrets in the CloudDeploymentConfig.
func (c *CloudDeploymentConfig) GetSecretFields() []Setting {
	secrets := []Setting{}
	return secrets
}

// CloudDeploymentForm creates a form for the CloudDeploymentConfig.
func CloudDeploymentForm(c *CloudDeploymentConfig) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			c.GCPProject.Input(),
			c.GCPRegion.Input(),
			c.GCPZone.Input(),
			c.OpsURI.Input(),
		).
			Title("GCP global settings"),
		huh.NewGroup(
			c.DomainName.Input(),
			c.SubdomainName.Input(),
			c.DaysToLive.Input(),
			huh.NewSelect[string]().
				Options(
					huh.Option[string]{Value: "platform", Key: "platform"},
					huh.Option[string]{Value: "ppp", Key: "ppp"},
				).
				Title(c.WebAppFlavor.Title).
				Description(c.WebAppFlavor.Description).
				Value(&c.WebAppFlavor.Value).
				Validate(c.WebAppFlavor.Validator),
		).
			Title("Deployment settings"),
		huh.NewGroup(
			c.SnapshotCH.Input(),
			c.SnapshotOS.Input(),
		).
			Title("Data snapshots"),
		huh.NewGroup(
			c.APIImage.Input(),
			c.APITag.Input(),
			c.APIAIImage.Input(),
			c.APIAITag.Input(),
			c.WebAppImage.Input(),
			c.WebAppTag.Input(),
			c.ClickhouseTag.Input(),
			c.OpensearchTag.Input(),
		).
			Title("Software versions"),
		huh.NewGroup(
			c.GCPSecretAIToken.Input(),
			c.GCPCloudDNSZone.Input(),
			c.GCPNetwork.Input(),
			c.GCPServiceAccount.Input(),
			huh.NewSelect[string]().
				Options(
					// PLATFORM_API_IGNORE_CACHE=true means cache is disabled
					huh.Option[string]{Value: "false", Key: "yes"},
					huh.Option[string]{Value: "true", Key: "no"},
				).
				Title(c.APICache.Title).
				Description(c.APICache.Description).
				Value(&c.APICache.Value),
		).
			Title("Additional settings"),
	)
}

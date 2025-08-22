package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// defaultsLocalPath is the default path to the local deployment configuration file.
const defaultsLocalPath = "./etc/defaults-local"

// LocalDeploymentConfig represents the configuration for a local deployment.
type LocalDeploymentConfig struct {
	DeploymentType Setting
	APIImage       Setting
	APITag         Setting
	APIAIImage     Setting
	APIAITag       Setting
	WebAppImage    Setting
	WebAppTag      Setting
	ClickhouseTag  Setting
	OpensearchTag  Setting
	Release        Setting
	ReleaseURL     Setting
	APIAIToken     Setting
}

// NewLocalDeploymentConfig creates a new LocalDeploymentConfig from a configuration file.
func NewLocalDeploymentConfig(configPath string) (*LocalDeploymentConfig, error) {
	effectivePath := defaultsLocalPath
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
	if env["OT_DEPLOYMENT_TYPE"] != "local" {
		return nil, fmt.Errorf("config file is for deployment type '%s', not 'local'", env["OT_DEPLOYMENT_TYPE"])
	}

	config := &LocalDeploymentConfig{
		DeploymentType: Setting{
			Env:   "OT_DEPLOYMENT_TYPE",
			Value: "local",
		},

		// First form: Data release
		Release: Setting{
			Title:       "Data release",
			Description: "The data release name should be in the form YY.MM, e.g. 25.06 for the June 2025 release.",
			Env:         "OT_RELEASE",
			Value:       env["OT_RELEASE"],
			Validator:   ValidateRelease,
		},
		ReleaseURL: Setting{
			Title:       "Release URL",
			Description: "URL to the release tarball",
			Env:         "OT_RELEASE_URL",
			Value:       env["OT_RELEASE_URL"],
			Validator:   ValidateURL,
		},

		// Second form: Software versions
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

		// Third form: Additional settings
		APIAIToken: Setting{
			Title:          "AI API token",
			Description:    "The API token to use inside the AI API for the publication summarization feature.",
			Env:            "OT_API_AI_TOKEN",
			Value:          env["OT_API_AI_TOKEN"],
			Secret:         true,
			SecretFilename: "openai_token",
		},
	}

	config.APITag.Validator = ValidateVersionTag(func() string { return config.APIImage.Value })
	config.APIAITag.Validator = ValidateVersionTag(func() string { return config.APIAIImage.Value })
	config.WebAppTag.Validator = ValidateVersionTag(func() string { return config.WebAppImage.Value })

	return config, nil
}

// GetDeploymentDir returns the directory where the local deployment files are stored.
func (c *LocalDeploymentConfig) GetDeploymentDir() string {
	deploymentDir := fmt.Sprintf("deployment-local-%s", c.Release.Value)
	absDeploymentDir, err := filepath.Abs(deploymentDir)
	if err != nil {
		log.Fatalf("error getting absolute path for deployment directory: %v", err)
	}
	return absDeploymentDir
}

// Validate validates all settings in a LocalDeploymentConfig.
func (c *LocalDeploymentConfig) Validate() error {
	var errs []error
	tools.AppendIfErr(&errs, c.Release.Validate())
	tools.AppendIfErr(&errs, c.ReleaseURL.Validate())
	tools.AppendIfErr(&errs, c.APIImage.Validate())
	tools.AppendIfErr(&errs, c.APITag.Validate())
	tools.AppendIfErr(&errs, c.APIAIImage.Validate())
	tools.AppendIfErr(&errs, c.APIAITag.Validate())
	tools.AppendIfErr(&errs, c.WebAppImage.Validate())
	tools.AppendIfErr(&errs, c.WebAppTag.Validate())
	tools.AppendIfErr(&errs, c.ClickhouseTag.Validate())
	tools.AppendIfErr(&errs, c.OpensearchTag.Validate())
	tools.AppendIfErr(&errs, c.APIAIToken.Validate())
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}
	return nil
}

// ReplaceFromEnv replaces the values of the LocalDeploymentConfig from environment.
func (c *LocalDeploymentConfig) ReplaceFromEnv() {
	c.Release.ReplaceFromEnv()
	c.ReleaseURL.ReplaceFromEnv()
	c.APIImage.ReplaceFromEnv()
	c.APITag.ReplaceFromEnv()
	c.APIAIImage.ReplaceFromEnv()
	c.APIAITag.ReplaceFromEnv()
	c.WebAppImage.ReplaceFromEnv()
	c.WebAppTag.ReplaceFromEnv()
	c.ClickhouseTag.ReplaceFromEnv()
	c.OpensearchTag.ReplaceFromEnv()
	c.APIAIToken.ReplaceFromEnv()
}

// ToString returns a string representation of the LocalDeploymentConfig.
func (c *LocalDeploymentConfig) ToString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Open Targets local deployment config for release %s\n", c.Release.Value))
	sb.WriteString(c.DeploymentType.ToString())
	sb.WriteString("\n# Data release settings\n")
	sb.WriteString(c.Release.ToString())
	sb.WriteString(c.ReleaseURL.ToString())
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
	sb.WriteString(c.APIAIToken.ToString())
	return sb.String()
}

// GetSecretFields returns a slice of Settings that are secrets in the LocalDeploymentConfig.
func (c *LocalDeploymentConfig) GetSecretFields() []Setting {
	secrets := []Setting{
		c.APIAIToken,
	}
	return secrets
}

// LocalDeploymentForm creates a form for configuring a local deployment.
func LocalDeploymentForm(c *LocalDeploymentConfig) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			c.Release.Input(),
			c.ReleaseURL.Input(),
		).
			Title("Data release settings"),
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
			c.APIAIToken.Input(),
		).
			Title("Additional settings"),
	)
}

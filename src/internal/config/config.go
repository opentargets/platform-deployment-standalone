// Package config provides functionality for managing deployment configurations.
package config

import (
	"log"
	"math/rand"
	"path/filepath"

	"github.com/joho/godotenv"
)

// CloudDeploymentMaxDaysToLive is the maximum number of days a cloud deployment can live.
const CloudDeploymentMaxDaysToLive = 14

// Location represents the deployment location.
type Location string

const (
	// Local represents a deployment that is hosted in the machine.
	Local Location = "local"
	// Cloud represents a deployment that is hosted in the cloud.
	Cloud Location = "cloud"
)

// Setting represents a configuration setting with its environment variable and value.
type Setting struct {
	// Env is the environment variable that holds the value.
	Env string
	// Value is the value of the setting.
	Value string
	// Secret indicates if the setting is sensitive and should not be logged.
	Secret bool
	// SecretFile is the file where the secret value is stored.
	SecretFile string
}

// Config holds the configuration for the deployment.
type Config struct {
	// Deployment Type form group.
	// Location is the type of deployment, either local or cloud.
	Location Setting

	// Data and software versions form group.
	// Release is the version of Open Targets to deploy (e.g., "25.06"). Local-exclusive.
	Release Setting
	// ReleaseURL is the base URL (root path) of the release. Local-exclusive.
	ReleaseURL Setting
	// SnapshotOS is the name of the snapshot for the OpenSearch disk. Cloud-exclusive.
	SnapshotOS Setting
	// SnapshotCH is the name of the snapshot for the ClickHouse disk. Cloud-exclusive.
	SnapshotCH Setting
	// Flavor is the deployment flavor, either platform or ppp. Cloud-exclusive.
	Flavor Setting
	// APIImage is the name of the Docker image for the API.
	APIImage Setting
	// APITag is the tag of the Docker image for the API.
	APITag Setting
	// APIOpenAIImage is the name of the Docker image for the OpenAI API.
	APIOpenAIImage Setting
	// APIOpenAITag is the tag of the Docker image for the OpenAI API. Local-exclusive.
	APIOpenAITag Setting
	// APIOpenAIToken is the token for OpenAI API access.
	APIOpenAIToken Setting
	// WebAppImage is the name of the Docker image for the WebApp.
	WebAppImage Setting
	// WebAppTag is the tag of the Docker image for the WebApp.
	WebAppTag Setting
	// OpensearchTag is the tag of the Docker image for OpenSearch.
	OpensearchTag Setting
	// ClickHouseTag is the tag of the Docker image for ClickHouse.
	ClickhouseTag Setting

	// Cloud Deployment Settings form group.
	// DomainName is the domain name for the cloud deployment.
	DomainName Setting
	// SubdomainName is the subdomain name for the cloud deployment.
	SubdomainName Setting
	// DaysToLive is the time to live for the deployment, in days.
	DaysToLive Setting
	// GCPSecretOpenAIToken is the Google Cloud Platform secret name for the OpenAI API token.
	GCPSecretOpenAIToken Setting
	// GCPProject is the Google Cloud Platform project name for cloud deployments.
	GCPProject Setting
	// GCPRegion is the Google Cloud Platform region for cloud deployments.
	GCPRegion Setting
	// GCPZone is the Google Cloud Platform zone for cloud deployments.
	GCPZone Setting
	// GCPCloudDNSZone is the Cloud DNS zone for the deployment.
	GCPCloudDNSZone Setting
	// GCPNetwork is the Google Cloud Platform network name for the deployment.
	GCPNetwork Setting
	// GCPServiceAccount is the Google Cloud Platform service account used.
	GCPServiceAccount Setting

	// Settings that are not shown in the form.
	// DeploymentFolder is the folder where the deployment files are stored.
	DeploymentFolder Setting
}

func randomString(length int) string {
	letters := []rune("abcdefgh")
	randomString := make([]rune, length)

	for i := range randomString {
		randomString[i] = letters[rand.Intn(len(letters))]
	}

	return string(randomString)
}

// New creates a new Config instance with default values.
func New(defaultsFilePath string) Config {
	env, err := godotenv.Read(defaultsFilePath)
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Location:             Setting{Env: "OT_DEPLOYMENT_LOCATION", Value: env["OT_DEPLOYMENT_LOCATION"]},
		Release:              Setting{Env: "OT_RELEASE", Value: env["OT_RELEASE"]},
		SnapshotCH:           Setting{Env: "TF_VAR_OT_SNAPSHOT_CH", Value: env["TF_VAR_OT_SNAPSHOT_CH"]},
		SnapshotOS:           Setting{Env: "TF_VAR_OT_SNAPSHOT_OS", Value: env["TF_VAR_OT_SNAPSHOT_OS"]},
		Flavor:               Setting{Env: "OT_FLAVOR", Value: env["OT_WEBAPP_FLAVOR"]},
		APIImage:             Setting{Env: "OT_API_IMAGE", Value: env["OT_API_IMAGE"]},
		APITag:               Setting{Env: "OT_API_TAG", Value: env["OT_API_TAG"]},
		APIOpenAIImage:       Setting{Env: "OT_API_OPENAI_IMAGE", Value: env["OT_API_OPENAI_IMAGE"]},
		APIOpenAITag:         Setting{Env: "OT_API_OPENAI_TAG", Value: env["OT_API_OPENAI_TAG"]},
		WebAppImage:          Setting{Env: "OT_WEBAPP_IMAGE", Value: env["OT_WEBAPP_IMAGE"]},
		WebAppTag:            Setting{Env: "OT_WEBAPP_TAG", Value: env["OT_WEBAPP_TAG"]},
		OpensearchTag:        Setting{Env: "OT_OPENSEARCH_TAG", Value: env["OT_OPENSEARCH_TAG"]},
		ClickhouseTag:        Setting{Env: "OT_CLICKHOUSE_TAG", Value: env["OT_CLICKHOUSE_TAG"]},
		ReleaseURL:           Setting{Env: "OT_RELEASE_URL", Value: env["OT_RELEASE_URL"]},
		APIOpenAIToken:       Setting{Env: "OT_API_OPENAI_TOKEN", Value: "", Secret: true, SecretFile: "openai_token"},
		DomainName:           Setting{Env: "TF_VAR_OT_DOMAIN_NAME", Value: env["TF_VAR_OT_DOMAIN_NAME"]},
		SubdomainName:        Setting{Env: "TF_VAR_OT_SUBDOMAIN_NAME", Value: randomString(4)},
		DaysToLive:           Setting{Env: "TF_VAR_OT_DAYS_TO_LIVE", Value: env["TF_VAR_OT_DAYS_TO_LIVE"]},
		GCPSecretOpenAIToken: Setting{Env: "TF_VAR_OT_GCP_SECRET_OPENAI_TOKEN", Value: env["TF_VAR_OT_GCP_SECRET_OPENAI_TOKEN"]},
		GCPProject:           Setting{Env: "TF_VAR_OT_GCP_PROJECT", Value: env["TF_VAR_OT_GCP_PROJECT"]},
		GCPRegion:            Setting{Env: "TF_VAR_OT_GCP_REGION", Value: env["TF_VAR_OT_GCP_REGION"]},
		GCPZone:              Setting{Env: "TF_VAR_OT_GCP_ZONE", Value: env["TF_VAR_OT_GCP_ZONE"]},
		GCPCloudDNSZone:      Setting{Env: "TF_VAR_OT_GCP_CLOUD_DNS_ZONE", Value: env["TF_VAR_OT_GCP_CLOUD_DNS_ZONE"]},
		GCPNetwork:           Setting{Env: "TF_VAR_OT_GCP_NETWORK", Value: env["TF_VAR_OT_GCP_NETWORK"]},
		GCPServiceAccount:    Setting{Env: "TF_VAR_OT_GCP_SA", Value: env["TF_VAR_OT_GCP_SA"]},
	}
}

// SetDeploymentFolder sets the deployment folder based on the release and subdomain name.
func (c *Config) SetDeploymentFolder() {
	var deploymentFolder string
	if c.Location.Value == string(Local) {
		deploymentFolder = "deployment-local-" + c.Release.Value
	} else if c.Location.Value == string(Cloud) {
		deploymentFolder = "deployment-cloud-" + c.Release.Value + "-" + c.SubdomainName.Value
	}

	deploymentFolder, err := filepath.Abs(deploymentFolder)
	if err != nil {
		log.Fatal(err)
	}

	c.DeploymentFolder = Setting{
		Env:   "OT_DEPLOYMENT_FOLDER",
		Value: deploymentFolder,
	}
}

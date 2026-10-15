// Package housekeeping manages deployment directories and config files.
package housekeeping

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/joho/godotenv"
	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

const gcpContextTimeout = 10 * time.Second

// EnsureDir checks if the deployment directory exists and creates it if not.
func EnsureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("error creating deployment directory %s: %v", path, err)
		}
	}
}

// EnsureTools checks if required tools are installed and exits with an error if any are missing.
func EnsureTools() {
	requiredTools := []string{"docker", "docker-compose", "pigz"}

	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			log.Fatalf("required tool %q not found in PATH or is not executable: %v", tool, err)
		}
	}
}

// PrepareDeploymentDir creates the deployment directory and copies necessary files.
func PrepareDeploymentDir(c config.DeploymentConfig) {
	localDeploymentFiles := []string{
		"./etc/compose.yaml",
		"./etc/Dockerfile-opensearch",
		"./etc/.dockerignore",
	}

	cloudDeploymentFiles := []string{
		"./etc/cleanup.sh.tftpl",
		"./etc/compose.yaml",
		"./etc/config-watcher.service",
		"./etc/config-watcher.sh",
		"./etc/Dockerfile-opensearch",
		"./etc/google-startup-script.sh",
		"./etc/main.tf",
		"./etc/nginx.conf.tftpl",
		"./etc/ch-config.xml",
	}

	EnsureDir(c.GetDeploymentDir())

	var filesToCopy []string
	switch c.(type) {
	case *config.CloudDeploymentConfig:
		filesToCopy = cloudDeploymentFiles
	case *config.LocalDeploymentConfig:
		filesToCopy = localDeploymentFiles
	}

	for _, filename := range filesToCopy {
		srcPath, err := filepath.Abs(filename)
		if err != nil {
			log.Fatalf("error getting absolute path for %s: %v", filename, err)
		}

		dstPath := filepath.Join(c.GetDeploymentDir(), filepath.Base(filename))

		srcFile, err := os.Open(srcPath)
		if err != nil {
			log.Fatalf("error opening source file %s: %v", srcPath, err)
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			log.Fatalf("error creating destination file %s: %v", dstPath, err)
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			log.Fatalf("error copying file from %s to %s: %v", srcPath, dstPath, err)
		}
	}
}

// WriteConfig writes the deployment configuration to its deployment directory.
func WriteConfig(c config.DeploymentConfig) {
	deploymentDir := c.GetDeploymentDir()
	EnsureDir(deploymentDir)

	configFilePath := deploymentDir + "/config"
	if err := os.WriteFile(configFilePath, []byte(c.ToString()), 0644); err != nil {
		log.Fatalf("error writing config file %s: %v", configFilePath, err)
	}

	for _, s := range c.GetSecretFields() {
		if s.Secret {
			secretFilePath := deploymentDir + "/" + s.SecretFilename
			if err := os.WriteFile(secretFilePath, []byte(s.Value), 0600); err != nil {
				log.Fatalf("error writing secret file %s: %v", secretFilePath, err)
			}
		}
	}
}

// UploadConfig writes a cloud deployment configuration to a GCS uri.
func UploadConfig(c *config.CloudDeploymentConfig) error {
	configFileURI := fmt.Sprintf("%s/%s", c.OpsURI.Value, c.SubdomainName.Value)

	if err := tools.WriteFileToGCS(configFileURI, c.ToString()); err != nil {
		return err
	}
	return nil
}

// checkSpec describes how to probe an Open Targets instance: the HTTP
// request shape and the rule for turning the response into a status string.
type checkSpec struct {
	method  string
	path    string
	body    string
	headers map[string]string
	// validate returns the status string to display and whether the
	// instance counts as "live".
	validate func(status int, body []byte) (string, bool)
}

var (
	CheckInstanceLiveSpec = checkSpec{
		method:  http.MethodPost,
		path:    "/api/v4/graphql",
		body:    `{"query": "{ meta { name } }"}`,
		headers: map[string]string{"Content-Type": "application/json", "Accept": "application/json"},
		validate: func(status int, body []byte) (string, bool) {
			if status != http.StatusOK {
				return strconv.Itoa(status), false
			}
			if strings.Contains(string(body), "Open Targets") {
				return "live", true
			}
			return "error: unknown response", false
		},
	}

	CheckInstanceAuthSpec = checkSpec{
		method: http.MethodGet,
		path:   "/",
		validate: func(status int, body []byte) (string, bool) {
			if status >= 200 && status < 400 {
				return "live", true
			}
			return strconv.Itoa(status), false
		},
	}
)

// CheckInstance checks the state of an Open Targets instance using the
// request shape and validation rules described by spec.
func CheckInstance(configFilename string, spec checkSpec) (string, string) {
	config, err := tools.ReadFileFromGCS(configFilename)
	if err != nil {
		return "unknown url", fmt.Sprintf("error: unable to read config file %s: %v", configFilename, err)
	}
	env, err := godotenv.Unmarshal(config)
	if err != nil {
		return "unknown url", fmt.Sprintf("error: unable to parse config file %s: %v\n", configFilename, err)
	}

	rootURL := fmt.Sprintf("https://%s.%s", env["TF_VAR_OT_SUBDOMAIN_NAME"], env["TF_VAR_OT_DOMAIN_NAME"])
	url := rootURL + spec.path

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var body io.Reader = http.NoBody
	if spec.body != "" {
		body = bytes.NewBufferString(spec.body)
	}
	r, err := http.NewRequestWithContext(ctx, spec.method, url, body)
	if err != nil {
		return rootURL, fmt.Sprintf("error: unable to create request to %s: %v\n", url, err)
	}
	for k, v := range spec.headers {
		r.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return rootURL, "error: timeout"
		}
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound {
				return rootURL, "error: unknown host"
			}
			return rootURL, "error: dns error"
		}
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Timeout() {
				return rootURL, "error: network timeout"
			}
			return rootURL, "error: network error"
		}
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") {
			return rootURL, "error: connection refused"
		}

		return rootURL, "error: network error"
	}
	defer resp.Body.Close()

	var b = make([]byte, 1024)
	resp.Body.Read(b)

	status, _ := spec.validate(resp.StatusCode, b)
	return rootURL, status
}

// AccessGCPSecret retrieves the latest version payload of a GCP Secret Manager
// secret. The returned value is the raw secret payload as a string.
func AccessGCPSecret(projectID, secretName string) (string, error) {
	if projectID == "" {
		return "", errors.New("unable to determine GCP project: set GOOGLE_CLOUD_PROJECT, configure one with `gcloud config set project <id>`, or run on a GCP resource")
	}
	if secretName == "" {
		return "", errors.New("secret name is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), gcpContextTimeout)
	defer cancel()

	client, err := secretmanager.NewRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to access google cloud secret manager: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretName),
	}

	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("unable to access secret %q: %w", secretName, err)
	}

	return string(resp.GetPayload().GetData()), nil
}

// GetGCPProjectID returns the current GCP project ID. It tries, in order:
//  1. the GOOGLE_CLOUD_PROJECT environment variable;
//  2. the compute metadata server (so it works on GCE/GKE/Cloud Run);
//  3. the project configured in the gcloud CLI (`gcloud config get-value
//     project`), for developer laptops authenticated with `gcloud auth login`.
//
// Returns an empty string when no source yields a value.
func GetGCPProjectID() string {
	if pid := os.Getenv("GOOGLE_CLOUD_PROJECT"); pid != "" {
		return pid
	}

	metadataClient := metadata.NewClient(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if pid, err := metadataClient.ProjectIDWithContext(ctx); err == nil && pid != "" {
		return pid
	}

	if path, err := exec.LookPath("gcloud"); err == nil {
		out, err := exec.Command(path, "config", "get-value", "project").Output()
		if err == nil {
			if pid := strings.TrimSpace(string(out)); pid != "" && pid != "(unset)" {
				return pid
			}
		}
	}

	return ""
}

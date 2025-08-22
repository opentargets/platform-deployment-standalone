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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// EnsureDir checks if the deployment directory exists and creates it if not.
func EnsureDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("error creating deployment directory %s: %v", path, err)
		}
	}
}

// PrepareDeploymentDir creates the deployment directory and copies necessary files.
func PrepareDeploymentDir(c config.DeploymentConfig) {
	localDeploymentFiles := []string{
		"./etc/compose.yaml",
		"./etc/Dockerfile-opensearch",
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

// CheckInstance checks the state of an Open Targets instance.
func CheckInstance(configFilename string) (string, string) {
	config, err := tools.ReadFileFromGCS(configFilename)
	if err != nil {
		return "unknown url", fmt.Sprintf("error: unable to read config file %s: %v", configFilename, err)

	}
	env, err := godotenv.Unmarshal(config)
	if err != nil {
		return "unknown url", fmt.Sprintf("error: unable to parse config file %s: %v\n", configFilename, err)
	}

	rootURL := fmt.Sprintf("https://%s.%s", env["TF_VAR_OT_SUBDOMAIN_NAME"], env["TF_VAR_OT_DOMAIN_NAME"])
	url := fmt.Sprintf("%s/api/v4/graphql", rootURL)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	payload := `{"query": "{ meta { name } }"}`
	r, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(payload))
	if err != nil {
		return rootURL, fmt.Sprintf("error: unable to create request to %s: %v\n", url, err)
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
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
	if resp.StatusCode != 200 {
		return rootURL, strconv.Itoa(resp.StatusCode)
	}
	defer resp.Body.Close()

	var b = make([]byte, 1024)
	resp.Body.Read(b)

	if strings.Contains(string(b), "Open Targets") {
		return rootURL, "live"
	}

	return rootURL, "error: unknown response"
}

package housekeeping

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/joho/godotenv"
	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

var dataImagePaths = map[string]string{
	"clickhouse": "disk_images/clickhouse.tgz",
	"opensearch": "disk_images/opensearch.tgz",
}

// DeployLocal executes a local deployment command using Terraform.
func DeployLocal(c *config.LocalDeploymentConfig) {
	godotenv.Load(c.GetDeploymentDir() + "/config")

	downloadsDir, err := filepath.Abs("./downloads")
	if err != nil {
		log.Fatalf("error getting absolute path of downloads dir: %v\n", err)
	}

	clickhouseArchveSrc := fmt.Sprintf("%s/%s/%s", c.ReleaseURL.Value, c.Release.Value, dataImagePaths["clickhouse"])
	clickhouseArchiveDst := fmt.Sprintf("%s/clickhouse-%s.tgz", downloadsDir, c.Release.Value)
	clickhouseDataDir := fmt.Sprintf("%s/clickhouse", c.GetDeploymentDir())
	opensearchArchveSrc := fmt.Sprintf("%s/%s/%s", c.ReleaseURL.Value, c.Release.Value, dataImagePaths["opensearch"])
	opensearchArchiveDst := fmt.Sprintf("%s/opensearch-%s.tgz", downloadsDir, c.Release.Value)
	opensearchDataDir := fmt.Sprintf("%s/opensearch", c.GetDeploymentDir())

	downloadAction := func() {
		var wg sync.WaitGroup
		errCh := make(chan error, 2)

		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := tools.DownloadFile(clickhouseArchveSrc, clickhouseArchiveDst); err != nil {
				errCh <- fmt.Errorf("error downloading clickhouse data: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := tools.DownloadFile(opensearchArchveSrc, opensearchArchiveDst); err != nil {
				errCh <- fmt.Errorf("error downloading opensearch data: %v", err)
			}
		}()
		wg.Wait()
	}
	tools.RunWithSpinner("downloading data, this may take a while...", downloadAction)

	extractAction := func() {
		var wg sync.WaitGroup
		errCh := make(chan error, 2)

		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := tools.ExtractArchive(clickhouseArchiveDst, clickhouseDataDir); err != nil {
				errCh <- fmt.Errorf("error extracting clickhouse data: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := tools.ExtractArchive(opensearchArchiveDst, opensearchDataDir); err != nil {
				errCh <- fmt.Errorf("error extracting opensearch data: %v", err)
			}
		}()
		wg.Wait()
	}
	tools.RunWithSpinner("extracting data, this may take a while...", extractAction)

	startAction := func() {
		cmd := exec.Command("docker", "compose", "--file", fmt.Sprintf("%s/compose.yaml", c.GetDeploymentDir()), "up", "-d", "--quiet-build", "--quiet-pull", "--build", "--force-recreate")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalf("error running docker compose: %v\n", err)
		}
	}
	tools.RunWithSpinner("starting local deployment...", startAction)

	fmt.Println("deployment successful, check out http://localhost:8080")
}

// DeployCloud executes a cloud deployment command using Terraform.
func DeployCloud(c *config.CloudDeploymentConfig) {
	godotenv.Load(c.GetDeploymentDir() + "/config")

	logFilename := fmt.Sprintf("terraform-%s.log", time.Now().Format("2006-01-02-150405"))
	logFilepath := filepath.Join(c.GetDeploymentDir(), logFilename)
	logFile, err := os.Create(logFilepath)
	if err != nil {
		log.Fatalf("error opening log file %s: %v\n", logFilepath, err)
	}
	defer logFile.Close()

	installer := &releases.LatestVersion{
		Product: product.Terraform,
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing terraform: %v\n", err)
	}

	workingDir := c.GetDeploymentDir()
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error creating terraform instance: %v\n", err)
	}

	tf.SetStderr(logFile)
	tf.SetStdout(logFile)

	parts := strings.SplitN(strings.TrimPrefix(c.OpsURI.Value, "gs://"), "/", 2)
	if len(parts) < 2 {
		log.Fatalf("invalid ops uri: %s", c.OpsURI.Value)
	}
	bucket := fmt.Sprintf("bucket=%s", parts[0])
	prefix := fmt.Sprintf("prefix=%s", parts[1])

	err = tf.Init(
		context.Background(),
		tfexec.Upgrade(true),
		tfexec.BackendConfig(bucket),
		tfexec.BackendConfig(prefix),
	)
	if err != nil {
		log.Fatalf("error initializing terraform: %v\n", err)
	}

	err = tf.WorkspaceSelect(context.Background(), c.SubdomainName.Value)
	if err != nil {
		err = tf.WorkspaceNew(context.Background(), c.SubdomainName.Value)
		if err != nil {
			log.Fatalf("error selecting or creating workspace: %v\n", err)
		}
	}

	err = tf.Apply(context.Background())
	if err != nil {
		log.Fatalf("error applying terraform configuration: %v\n", err)
	}

	// We need to do this twice because if the change includes a new data volume,
	// the first apply will create the volume but not attach it to the instance.
	err = tf.Apply(context.Background())
	if err != nil {
		log.Fatalf("error applying terraform configuration: %v\n", err)
	}
}

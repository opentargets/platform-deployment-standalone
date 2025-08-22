package housekeeping

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/joho/godotenv"
	"github.com/opentargets/platform-deployment-standalone/internal/config"
	"github.com/opentargets/platform-deployment-standalone/internal/tools"
)

// Destroy destroys a deployment.
func Destroy(deploymentPath string) {
	action := func() {
		if strings.HasPrefix(deploymentPath, "gs://") {
			destroyCloudDeployment(deploymentPath)
		} else {
			env, err := godotenv.Read(deploymentPath + "/config")
			if err != nil {
				log.Fatalf("error reading config file in specified folder: %v", err)
			}

			switch env["OT_DEPLOYMENT_TYPE"] {
			case "local":
				destroyLocalDeployment(deploymentPath)
			case "cloud":
				destroyCloudDeployment(deploymentPath)
			default:
				log.Fatalf("unknown deployment type: %s", env["DEPLOYMENT_TYPE"])
			}
		}
	}

	tools.RunWithSpinner("destroying deployment", action)
}

func destroyLocalDeployment(deploymentPath string) {
	err := exec.Command("docker-compose", "-f", deploymentPath+"/compose.yaml", "down").Run()
	if err != nil {
		log.Fatalf("error destroying local deployment: %v", err)
	}
	log.Printf("local deployment %s destroyed", deploymentPath)
}

func destroyCloudDeployment(deploymentPath string) {
	c, err := config.NewCloudDeploymentConfig(deploymentPath)
	if err != nil {
		log.Fatalf("error loading cloud deployment config: %v", err)
	}
	PrepareDeploymentDir(c)

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

	err = tf.Destroy(context.Background())
	if err != nil {
		log.Fatalf("error destroying terraform deployment: %v\n", err)
	}

	log.Printf("cloud deployment %s destroyed, you can now safely delete the folder %s", c.SubdomainName.Value, c.GetDeploymentDir())
}

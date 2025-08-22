// Package cmd is the CLI entry point.
package cmd

import (
	"os"

	"github.com/opentargets/platform-deployment-standalone/internal/housekeeping"
	"github.com/spf13/cobra"
)

var (
	unattended bool
	configFile string
)

// RootCmd is the root command of the Open Targets Platform deployment tool.
var RootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "Open Targets Platform deployment tool",
	Long: `Open Targets Platform deployment tool allows you to create a deployment
either in a local environment or in the cloud.`,
}

var deployCmd = &cobra.Command{
	Use:   "deploy [local|cloud] [flags]",
	Short: "Create a deployment",
	Long:  "Create a deployment of the Open Targets Platform, either locally or in the cloud.",
}

var destroyCmd = &cobra.Command{
	Use:   "destroy <deployment-path-or-uri>",
	Short: "Destroy a deployment",
	Long: `Destroy a deployment of the Open Targets Platform, either locally or in the cloud.

You can pass either a local or cloud deployment folder or a remote Google Cloud
Storage URI (gs://bucket/path/to/folder).
`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		housekeeping.Destroy(args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list <backend-uri>",
	Short: "List cloud deployments",
	Long: `List cloud deployments of the Open Targets Platform in your Google Cloud project.

This command requires a backend URI as argument, where the deployment state is
stored.
`,
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		ListCloud(args[0])
	},
}

var localCmd = &cobra.Command{
	Use:     "local",
	Short:   "Create a local deployment",
	Long:    "Create a local deployment of the Open Targets Platform.",
	Example: "deploy local",
	Run: func(_ *cobra.Command, _ []string) {
		// TODO: Finish local deployment.
		RunLocal(false, "./etc/defaults-local")
	},
}

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "Create a cloud deployment",
	Long: `Create a cloud deployment of the Open Targets Platform.

This command deploys an Open Targets Platform instance in Google Cloud. You can
run it in interactive mode to configure the deployment manually, or in unattended
mode.

You can pass a configuration file with the --config flag. In interactive mode, the
form will be pre-filled with the the values inside. Configuration files can either
be local files or Google Cloud Storage URIs (gs://bucket/path/to/file).

The tool will upload both the terraform state and the configuration file to the
GCS URI specified in the OT_OPS_URI environment variable. Later on, it is possible
to use the configuration file from the cloud to update an instance.

If no configuration file is specified, the tool will use the defaults provided in
./etc/defaults-cloud.

Any environment variables that are set when running the tool will override the
values in the configuration file or the defaults. See examples below.
`,
	Example: `  $ deploy cloud
      shows a form to configure the deployment

  $ deploy cloud --unattended
      deploys an instance automatically, using default values

  $ deploy cloud --config ./config-2506
      shows a form to configure the deployment using values from ./config-2506

  $ deploy cloud --unattended --config ./config-2506
      deploys an instance automatically, using values from ./config-2506

  $ deploy cloud --config gs://open-targets-ops/terraform/devinstance/dev
      shows a form to configure the dev instance, which is stored in a Google
      Cloud Storage bucket.

  $ OT_API_IMAGE_TAG="another" deploy cloud --unattended --config ./myconfig
      deploys an instance automatically, using the configuration in ./myconfig,
      but overriding the API image tag to 'another'
`,
	Run: func(_ *cobra.Command, _ []string) {
		RunCloud(unattended, configFile)
	},
}

func init() {
	cloudCmd.Flags().BoolVarP(&unattended, "unattended", "u", false, "run in unattended mode")
	cloudCmd.Flags().StringVarP(&configFile, "config", "c", "", `Configuration file. This can be a local file or a Google
Cloud Storage URI (gs://bucket/path/to/file). If -c is not
specified, the tool will use the defaults values found in
./etc/defaults-cloud.`)

	RootCmd.AddGroup(&cobra.Group{
		ID:    "main",
		Title: "Main commands",
	})

	deployCmd.GroupID = "main"
	destroyCmd.GroupID = "main"
	listCmd.GroupID = "main"

	deployCmd.AddGroup(&cobra.Group{
		ID:    "deploy",
		Title: "Main commands",
	})

	localCmd.GroupID = "deploy"
	cloudCmd.GroupID = "deploy"

	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(destroyCmd)
	RootCmd.AddCommand(listCmd)
	deployCmd.AddCommand(localCmd)
	deployCmd.AddCommand(cloudCmd)
}

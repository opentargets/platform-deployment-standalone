# Standalone Deployment for Open Targets Platform
This application spins up a single-node deployment of [Open Targets
Platform](https://platform.opentargets.org/). It will fetch all the necessary
resources and components and provide sane defaults so that you can get started
quickly.

## Quick start
```bash
# clone the repo
git clone https://github.com/opentargets/standalone-deployment-platform.git
# navigate to the project
cd standalone-deployment-platform
# set the profile
./configure 2503-testrun-3
# start the platform
make
```

> [!NOTE]
>
> The first time you specify a data image in your deployment, it has to be
> downloaded. This will take a relatively long time, depending on the size of
> the image and the connection speed. Subsequent deployments with the same data
> image will not need to re-download, so they will be much faster.

Visit the platform in your browser http://localhost


## Requirements
- [docker](https://docs.docker.com/get-docker/)
- [gsutil](https://cloud.google.com/storage/docs/gsutil_install) (if downloading
  releases from Google Cloud Storage)
- at least 250GB local storage (as of release 25.03)

## Resources:
1. [Open Targets data release](https://ftp.ebi.ac.uk/pub/databases/opentargets/platform)
   1. OpenSearch data
   2. ClickHouse data
2. Open Targets software
   1. [Web app](https://github.com/opentargets/ot-ui-apps)
   2. [API](https://github.com/opentargets/platform-api)
   3. [OpenAI API](https://github.com/opentargets/ot-ai-api)
3. 3rd party software
   1. [OpenSearch](https://opensearch.org/)
   2. [ClickHouse](https://clickhouse.com/)


## Profiles
The `profiles` directory contains the [config files](profiles/2503-testrun-3).
These are small bash scripts exporting a set of environment variables that are
used to configure deployments. With them, you can combine standard software and
data releases with your own custom images for parts of the platform.

> [!WARNING]
>
> Keep in mind deploying a new profile will overwrite the current deployment!


### Fields
An example of a profile is shown below:
```bash
#!/bin/bash

export OT_RELEASE="25.03-testrun-3"

export OT_DATA_URL="gs://open-targets-pre-data-releases/$OT_RELEASE/disk_images"

export OT_OPENSEARCH_TAG="2.19.1"
export OT_CLICKHOUSE_TAG="23.3.1.2823"
export OT_API_TAG="25.0.0-alpha.14"
export OT_OPENAI_TAG="0.0.10"
export OT_WEBAPP_IMAGE="ghcr.io/javfg/ot-ui-apps/ot-ui-apps"
export OT_WEBAPP_TAG="0.11.11"

# if set to `true`, disable cache and set log level to debug
export OT_DEBUG=true

# if set to `true`, expose the services to the internet
export OT_EXPOSE=false
```

The available fields are:
| Field             | Required | Default                                     | Description                                                |
| ----------------- | -------- | ------------------------------------------- | ---------------------------------------------------------- |
| OT_RELEASE        | Yes      |                                             | The release of the Open Targets data to use                |
| OT_DATA_URL       | Yes      |                                             | The URL to the data release                                |
| OT_OPENSEARCH_TAG | No       | `latest`                                    | The version of OpenSearch to use                           |
| OT_CLICKHOUSE_TAG | No       | `latest`                                    | The version of ClickHouse to use                           |
| OT_API_IMAGE      | No       | `ghcr.io/opentargets/platform-api`          | The version of the Open Targets API to use                 |
| OT_API_TAG        | No       | `latest`                                    | The version of the Open Targets API to use                 |
| OT_OPENAI_IMAGE   | No       | `ghcr.io/opentargets/ot-ai-api`             | The version of the OpenAI API to use                       |
| OT_OPENAI_TAG     | No       | `latest`                                    | The version of the OpenAI API to use                       |
| OT_WEBAPP_IMAGE   | No       | `ghcr.io/opentargets/ot-ui-apps/ot-ui-apps` | The version of the Open Targets Webapp to use              |
| OT_WEBAPP_TAG     | No       | `latest`                                    | The version of the Open Targets Webapp to use              |
| OT_WEBAPP_PORT    | No       | `8080`                                      | The port to expose the webapp on                           |
| OT_DEBUG          | No       | `false`                                     | If set to `true`, disable cache and set log level to debug |
| OT_EXPOSE         | No       | `false`                                     | If set to `true`, expose the services on all interfaces    |


## Usage
The deployment is managed by a `configure` script to set a profile, and a
`Makefile` that provides a set of commands to manage the deployment. To see the
full list of commands, issue `make help`.

```
❯ make help
help             show this help message
clean            clean up
start            default target: deploy the standalone deployment
stop             stop the deployment
```

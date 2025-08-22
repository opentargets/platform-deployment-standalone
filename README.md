# Open Targets standalone deployment tool

Create your own Open Targets Platform instance hosted locally or in Google Cloud.

## System requirements

For the local deployment, you need a machine with at least **4 cores** and **16GB
of RAM.** About **300GB of hard drive space** is required for installation (as
of 25.06 release).
You can after reduce that amount by about 100GB by deleting the `downloads`
folder containing the disk image tarballs.

Cloud deployments use a [`n1-standard-4`](https://cloud.google.com/compute/docs/general-purpose-machines#n1_machine_types)
machine.

Regarding software, you will need:

* [Go](https://go.dev/doc/install), to compile the configurator
* [Docker](https://docs.docker.com/engine/install/), for local deployments
* [GCloud CLI](https://cloud.google.com/sdk/docs/install), and
* [Terraform](https://developer.hashicorp.com/terraform/install) for cloud deployments

## Usage

```
Open Targets Platform deployment tool allows you to create a deployment
either in a local environment or in the cloud.

Usage:
  ./platform [command]

Main commands
  deploy      Create a deployment
  destroy     Destroy a deployment
  list        List cloud deployments

Additional Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

Flags:
  -h, --help   help for ./platform

Use "./platform [command] --help" for more information about a command.
```

## Cloud deployments

You must be authenticated with the Google Cloud CLI, and have permissions to
create GCE vms, disks, firewall entries, modify networks and add record sets to
Cloud DNS zones.

> [!WARNING]
> People outside of the Open Targets Organisation interested in creating a cloud
> deployment must read carefully through the rest of this section, as there are
> some required additional steps.

The cloud deployments are intended to be disposable, and they are self-deleting
after a configurable timeframe. For this, the application assumes you have a
Service Account with the following roles:

* `roles/compute.instanceAdmin.v1` to delete a machine to host the platform
* `roles/compute.storageAdmin` to delete the disks holding the data
* `roles/compute.securityAdmin` to delete firewall rules that allow access to
the API and AI API
* `roles/compute.viewer` to watch operations related to disk destruction on
cleanup
* `roles/compute.networkAdmin` to allow modifying the network to delete firewall
rules
* `roles/dns.admin` to add and remove record sets for the deployment hostname and
from letsencrypt's certificate validation method from Cloud DNS.
* `secretmanager.secretAccessor` to access the secret holding your token for the
AI API

You can add those roles to an account in a restricted fashion by using these
commands:

``` bash
service_account_name="example"
project="example-project"
zone="europe-west1-d"

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/compute.instanceAdmin.v1" \
	--condition='expression=resource.name.startsWith("projects/'$project'/zones/europe-west1-d/instances/devinstance-"),title="Limited to devinstance instances"'

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/compute.storageAdmin" \
	--condition='expression=resource.name.startsWith("projects/'$project'/zones/europe-west1-d/disks/devinstance-"),title="Limited to devinstance disks"'

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/compute.viewer" \
	--condition='expression=resource.name.startsWith("projects/'$project'/zones/europe-west1-d/operations/operation-"),title="Limited to operations viewing"'

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/secretmanager.secretAccessor" \
	--condition='expression=resource.name.startsWith("projects/426265110888/secrets/openai-token"),title="Limited to openai token secret"'

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/compute.securityAdmin" \
	--condition='expression=resource.name.startsWith("projects/'$project'/global/firewalls/devinstance-"),title="Limited to devinstance firewall rules"'

gcloud projects add-iam-policy-binding $project \
  --member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
  --role="roles/compute.networkAdmin" \
  --condition='expression=resource.name=="projects/'$project'/global/networks/default",title="Limited to default network"'

gcloud projects add-iam-policy-binding $project \
	--member="serviceAccount:$service_account_name@$project.iam.gserviceaccount.com" \
	--role="roles/dns.admin"
```

You also need a secret named `openai_token`, which holds your OpenAI API Token,
used for the literature summarization feature. This is optional, but it must be
set to an empty string or the deployment will fail.

It is required to edit the `etc/default.tfbackend` file and change the `bucket`
and `prefix` values inside to ones you own.


# Copyright
Copyright 2014-2025 EMBL - European Bioinformatics Institute, Genentech, GSK,
MSD, Pfizer, Sanofi and Wellcome Sanger Institute

This software was developed as part of the Open Targets project. For more
information please see: http://www.opentargets.org.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the
License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.

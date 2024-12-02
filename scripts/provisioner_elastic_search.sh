#!/bin/bash
# Elastic Search Provisioner

# Bootstrap
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/config.sh"


# Main
logi "Provisioning Elastic Search"
# Check existing elastic search data volume
if [ -d "${otops_deployment_path_elastic_search_volume}" ]; then
    logw "ALREADY EXISTING Elastic Search deployment at '${otops_deployment_path_elastic_search_volume}', SKIP PROVISIONING"
    exit 0
fi

# Check-create elastic search data volume deployment folder
if [ ! -d "${otops_deployment_path_elastic_search_volume}" ]; then
    logi "Creating Elastic Search data volume deployment folder at '${otops_deployment_path_elastic_search_volume}'"
    mkdir -p "${otops_deployment_path_elastic_search_volume}"
fi

# Uncompress elastic search source data volume tarball into elastic search data volume deployment folder
logi "Uncompressing Elastic Search source data volume tarball '${otops_path_elastic_search_source_data_volume}' into Elastic Search data volume deployment folder '${otops_deployment_path_elastic_search_volume}'"
tar -xzf "${otops_path_elastic_search_source_data_volume}" -C "${otops_deployment_path_elastic_search_volume}"
# Update elastic search data volume deployment folder permissions, specially for cloud deployments
find "${otops_deployment_path_elastic_search_volume}" -type d -exec chmod 777 {} \;
find "${otops_deployment_path_elastic_search_volume}" -type f -exec chmod 666 {} \;
logi "[Elastic Search Provisioner] Done"

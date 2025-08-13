#! /bin/bash
set -e

download_file() {
  if [[ -f "$2" ]]; then
    echo "file $2 already exists, skipping download (delete the file if you want to re-download)"
    return 0
  fi

  mkdir -p downloads
  if [[ "$1" =~ ^https?.*|^ftp.* ]]; then
    exec wget "$1" -O "$2"
  elif [[ "$1" =~ ^gs:// ]]; then
    exec gsutil cp "$1" "$2"
  fi
}

extract_archive() {
  if [[ -d "$2" ]]; then
    echo "directory $2 already exists, skipping extraction (delete the directory if you want to re-extract)"
    return 0
  fi

  mkdir -p "$2"
  exec tar --use-compress-program=pigz -xf "$1" -C "$2"
}

cleanup() {
  echo "killing background tasks, you may have to manually delete partial downloads or deployment folders"
  pkill -P $$
  exit 1
}

trap cleanup SIGINT SIGTERM

# run configurator
OT_DEPLOYMENT_FOLDER=$(./bin/configure)
set -a
# shellcheck disable=SC1091
source "$OT_DEPLOYMENT_FOLDER/config"
set +a
cp etc/* "$OT_DEPLOYMENT_FOLDER"

# local deployment
if [ "$OT_DEPLOYMENT_LOCATION" = "local" ]; then
  clickhouse_archive="downloads/clickhouse-$OT_RELEASE.tgz"
  opensearch_archive="downloads/opensearch-$OT_RELEASE.tgz"

  echo "downloading data, this may take a while..."
  download_file "$OT_RELEASE_URL/$OT_RELEASE/disk_images/clickhouse.tgz" "$clickhouse_archive" &
  download_file "$OT_RELEASE_URL/$OT_RELEASE/disk_images/opensearch.tgz" "$opensearch_archive" &
  wait

  echo "extracting data, this may take a while..."
  extract_archive "$clickhouse_archive" "$OT_DEPLOYMENT_FOLDER/clickhouse" &
  extract_archive "$opensearch_archive" "$OT_DEPLOYMENT_FOLDER/opensearch" &
  wait

  echo "running docker compose"
  if docker compose --file "$OT_DEPLOYMENT_FOLDER/compose.yaml" up -d --quiet-build --quiet-pull --build --force-recreate; then
    echo "deployment successful, check out http://localhost:${OT_WEBAPP_PORT:-8080}"
  else
    echo "docker compose failed"
    exit 1
  fi

# cloud deployment
elif [ "$OT_DEPLOYMENT_LOCATION" = "cloud" ]; then
  touch "$OT_DEPLOYMENT_FOLDER/terraform.log"

  echo "initializing terraform"
  if ! terraform -chdir="$OT_DEPLOYMENT_FOLDER" init -no-color >>"$OT_DEPLOYMENT_FOLDER/terraform.log" 2>&1; then
    echo "terraform init failed"
    exit 1
  fi

  echo "applying terraform configuration, this takes about 5 minutes..."
  if terraform -chdir="$OT_DEPLOYMENT_FOLDER" apply -auto-approve -no-color >>"$OT_DEPLOYMENT_FOLDER/terraform.log" 2>&1; then
    echo "deployment successful, Open Targets Platform will be available at:"
    echo "  http://${TF_VAR_OT_SUBDOMAIN_NAME}.${TF_VAR_OT_DOMAIN_NAME}  "
    echo "it may take another 5 minutes for the deployment to be fully ready,"
    echo "if the web does not load or errors show up, please wait a bit longer"
    echo "and try again"
  else
    echo "terraform apply failed"
    exit 1
  fi

fi

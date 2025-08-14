#!/bin/bash

# mount data disks
mkdir -p /platform/clickhouse
mkdir -p /platform/opensearch
mount /dev/disk/by-id/google-datavolume-ch /platform/clickhouse
mount /dev/disk/by-id/google-datavolume-os /platform/opensearch
chown -R 1000:1000 /platform/clickhouse
chown -R 1000:1000 /platform/opensearch

# install dependencies
apt-get purge -y man-db
apt-get update
apt-get install -y ca-certificates curl at certbot python3-certbot-dns-google nginx
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
echo \
"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
systemctl start docker

# copy files from metadata server and parse config
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/compose-file > /platform/compose.yaml
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/dockerfile-opensearch > /platform/Dockerfile-opensearch
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/config > /platform/config
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/nginx-conf > /etc/nginx/sites-enabled/default
set -a
# shellcheck source=defaults
source /platform/config
set +a

# set auto-delete for disks manually because terraform does not let us
gcloud config set project $TF_VAR_OT_GCP_PROJECT
gcloud compute instances set-disk-auto-delete "devinstance-$TF_VAR_OT_SUBDOMAIN_NAME" --auto-delete --disk="devinstance-datavolume-os-$TF_VAR_OT_SUBDOMAIN_NAME" --zone="$TF_VAR_OT_GCP_ZONE"
gcloud compute instances set-disk-auto-delete "devinstance-$TF_VAR_OT_SUBDOMAIN_NAME" --auto-delete --disk="devinstance-datavolume-ch-$TF_VAR_OT_SUBDOMAIN_NAME" --zone="$TF_VAR_OT_GCP_ZONE"

# prepare secrets
gcloud secrets versions access latest --secret="$TF_VAR_OT_GCP_SECRET_AI_TOKEN" > /platform/openai_token
chmod 600 /platform/openai_token

# schedule cleanup script
cat > /platform/cleanup.sh <<-CLEANUP_EOF
  if [ "$TF_VAR_OT_GCP_NETWORK" == "default" ]; then
    gcloud compute firewall-rules delete "devinstance-allow-$TF_VAR_OT_SUBDOMAIN_NAME" --project="$TF_VAR_OT_GCP_PROJECT" --quiet
  fi
  gcloud dns record-sets delete "$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME." --type=A --zone="$TF_VAR_OT_GCP_CLOUD_DNS_ZONE" --project="$TF_VAR_OT_GCP_PROJECT" --quiet
  gcloud compute instances delete "devinstance-$TF_VAR_OT_SUBDOMAIN_NAME" --zone="$TF_VAR_OT_GCP_ZONE" --project="$TF_VAR_OT_GCP_PROJECT" --quiet
CLEANUP_EOF
chmod +x /platform/cleanup.sh
if [ $TF_VAR_OT_DAYS_TO_LIVE -ne 0 ]; then
  echo "/usr/bin/bash /platform/cleanup.sh" | at now + $TF_VAR_OT_DAYS_TO_LIVE days
fi

# prepare cert
certbot certonly \
  --dns-google \
  --non-interactive \
  --agree-tos \
  --email "admin@$TF_VAR_OT_DOMAIN_NAME" \
  --domains "$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME"

# restart nginx
nginx -s reload

# run platform
export OT_API_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME"
export OT_API_AI_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME"
export OT_DEPLOYMENT_FOLDER="/platform"
export OT_WEBAPP_FLAVOR="$OT_WEBAPP_FLAVOR"
cd /platform || exit 1
docker compose -f compose.yaml up -d

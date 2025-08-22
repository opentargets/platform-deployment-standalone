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
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/cleanup > /platform/cleanup.sh
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/config-watcher-script > /platform/config-watcher.sh
curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/config-watcher-service > /etc/systemd/system/config-watcher.service
set -a
# shellcheck source=/dev/null
source /platform/config
set +a
# prepare frontend env vars
export OT_WEBAPP_API_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME/api/v4/graphql"
export OT_WEBAPP_OPENAI_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME"

# prepare secrets
gcloud secrets versions access latest --secret="$TF_VAR_OT_GCP_SECRET_AI_TOKEN" > /platform/openai_token
chmod 600 /platform/openai_token

# schedule cleanup script

chmod +x /platform/cleanup.sh
if [ "$TF_VAR_OT_DAYS_TO_LIVE" -ne 0 ]; then
  echo "/usr/bin/bash /platform/cleanup.sh" | at now + "$TF_VAR_OT_DAYS_TO_LIVE" days
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

# run platform and the config watcher
cd /platform || exit 1
docker compose -f compose.yaml up --quiet-build --quiet-pull -d
chmod +x /platform/config-watcher.sh
systemctl enable --now config-watcher

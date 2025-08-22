#!/bin/bash

while true; do
  curl -sH "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/config > /platform/config-new
  if ! cmp -s /platform/config /platform/config-new; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') config updated"
    mv /platform/config-new /platform/config
    set -a
    # shellcheck source=/dev/null
    source /platform/config
    set +a
    export OT_WEBAPP_API_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME/api/v4/graphql"
    export OT_WEBAPP_OPENAI_URL="https://$TF_VAR_OT_SUBDOMAIN_NAME.$TF_VAR_OT_DOMAIN_NAME"

    umount /platform/clickhouse
    umount /platform/opensearch

    mount /dev/disk/by-id/google-datavolume-ch /platform/clickhouse
    mount /dev/disk/by-id/google-datavolume-os /platform/opensearch
    chown -R 1000:1000 /platform/clickhouse
    chown -R 1000:1000 /platform/opensearch

    docker compose -f /platform/compose.yaml up --quiet-build --quiet-pull --build --force-recreate -d
  else
    echo .
  fi
  sleep 60
done

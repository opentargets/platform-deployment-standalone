#!/bin/bash
set -e

# shellcheck source=/dev/null
source ./deployment/profile

if [ "$OT_EXPOSE" == "true" ]; then
  export OT_OS_PORT="0.0.0.0:9200"
  export OT_CH_PORT="0.0.0.0:8123"
  export OT_API_PORT="0.0.0.0:8081"
  export OT_OPENAI_PORT="0.0.0.0:8082"
  export OT_WEBAPP_PORT="0.0.0.0:${OT_WEBAPP_PORT:-8080}"
  OT_OPENSEARCH_UID="$(id -u)"
  export OT_OPENSEARCH_UID
  OT_OPENSEARCH_GID="$(id -g)"
  export OT_OPENSEARCH_GID
fi

docker compose up -d --build --force-recreate

echo "successfully started $OT_RELEASE deployment"

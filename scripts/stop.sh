#!/bin/bash
set -e

# shellcheck source=/dev/null
source ./deployment/profile

docker compose stop

echo "successfully stopped $OT_RELEASE deployment"

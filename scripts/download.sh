#!/bin/bash
set -e

# shellcheck source=/dev/null
source ./deployment/profile

mkdir -p downloads

if [ "$1" != "opensearch" ] && [ "$1" != "clickhouse" ]; then
    echo "usage: download.sh <opensearch|clickhouse>"
    exit 1
fi

echo "downloading $1 data"
if [[ "$OT_DATA_URL" =~ ^https?.*|^ftp.* ]]; then
    wget "$OT_DATA_URL/$1".tgz -O "./downloads/$1.tgz"
elif [[ "$OT_DATA_URL" =~ ^gs://* ]]; then
    gsutil cp "$OT_DATA_URL/$1.tgz" "./downloads/$1.tgz"
fi

echo "successfully downloaded $1 data"

#!/bin/bash
set -e

# shellcheck source=/dev/null
source ./deployment/profile

if [ "$1" != "opensearch" ] && [ "$1" != "clickhouse" ]; then
    echo "usage: download.sh <opensearch|clickhouse>"
    exit 1
fi

mkdir "deployment/$1"

echo "unpacking $1 data"
tar --use-compress-program=pigz -xf "./downloads/$1.tgz" -C "deployment/$1"

touch "deployment/$1/.unpacked"
echo "successfully unpacked $1 data"

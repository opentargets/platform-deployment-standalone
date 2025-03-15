### HOUSEKEEPING TARGETS ###
.PHONY: help clean clean_downloads .is_configured

MAKEFLAGS += -j8
.DEFAULT_GOAL := start

help:  ## show this help message
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

clean:  ## clean up
	@rm -rf deployment

.is_configured:
	@if [ ! -f "./deployment/.configured" ]; then \
		echo "you must set a profile with './configure <profile>' first"; \
		exit 1; \
	fi


### DEPLOYMENT TARGETS ###
downloads/opensearch.tgz:
	@./scripts/download.sh opensearch

downloads/clickhouse.tgz:
	@./scripts/download.sh clickhouse

deployment/opensearch/.unpacked: downloads/opensearch.tgz
	@./scripts/unpack.sh opensearch

deployment/clickhouse/.unpacked: downloads/clickhouse.tgz
	@./scripts/unpack.sh clickhouse

.opensearch_data: deployment/opensearch/.unpacked  ## download opensearch data
.clickhouse_data: deployment/clickhouse/.unpacked  ## download clickhouse data

start: .is_configured .opensearch_data .clickhouse_data  ## default target: deploy the standalone deployment
	@./scripts/start.sh

stop: .is_configured  ## stop the deployment
	@./scripts/stop.sh

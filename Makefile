# This Makefile helper automates bringing up and tearing down a local deployment of Open Targets Platform
.DEFAULT_GOAL := help

# Environment variables
ROOT_DIR_MAKEFILE:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
export OTOPS_PATH_PROFILES:=$(ROOT_DIR_MAKEFILE)/profiles
export OTOPS_PATH_SCRIPTS:=$(ROOT_DIR_MAKEFILE)/scripts
export OTOPS_PATH_RELEASE:=$(ROOT_DIR_MAKEFILE)/release
export OTOPS_PATH_CONFIG:=$(ROOT_DIR_MAKEFILE)/config
export OTOPS_PATH_CONFIG_WEBAPP:=$(OTOPS_PATH_CONFIG)/webapp
export OTOPS_PATH_DEPLOYMENT:=$(ROOT_DIR_MAKEFILE)/deployment
export OTOPS_PATH_DEPLOYMENT_CLICKHOUSE:=$(OTOPS_PATH_DEPLOYMENT)/clickhouse
export OTOPS_PATH_DEPLOYMENT_ELASTIC_SEARCH:=$(OTOPS_PATH_DEPLOYMENT)/elastic_search
export OTOPS_PROFILE_ELASTIC_SEARCH_UID=$(shell id -u)
export OTOPS_PROFILE_ELASTIC_SEARCH_GID=$(shell id -g)
export OTOPS_PATH_DEPLOYMENT_WEBAPP:=$(OTOPS_PATH_DEPLOYMENT)/webapp
export OTOPS_ACTIVE_PROFILE:=config_profile.sh
export OTOPS_FLAG_DEPLOYED:=.deployed
export OTOPS_PROVISIONER_CLICKHOUSE:=$(OTOPS_PATH_SCRIPTS)/provisioner_clickhouse.sh
export OTOPS_PROVISIONER_ELASTIC_SEARCH:=$(OTOPS_PATH_SCRIPTS)/provisioner_elastic_search.sh
export OTOPS_PROVISIONER_WEBAPP:=$(OTOPS_PATH_SCRIPTS)/provisioner_webapp.sh
export OTOPS_COLLECT_DATA:=$(OTOPS_PATH_SCRIPTS)/collect.sh
export OTOPS_CLEAN_DATA:=$(OTOPS_PATH_SCRIPTS)/clean_release.sh

# Targets
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-28s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

set_profile: ## Set an active configuration profile, e.g. "make set_profile profile='dev'" (see folder 'profiles')
	@echo "[OTOPS] Setting active profile '${profile}'"
	@ln -sf profiles/config.${profile} ${OTOPS_ACTIVE_PROFILE}

clean_profile: ## Clean the active configuration profile
	@echo "[OTOPS] Cleaning active profile"
	@rm -f ${OTOPS_ACTIVE_PROFILE}

.env: env.sh ${OTOPS_ACTIVE_PROFILE} ## Create a .env file from the active configuration profile
	@echo "[OTOPS] Creating .env file from active profile"
	@bash -c 'source ${OTOPS_ACTIVE_PROFILE} && ./env.sh' > .env

summary_environment: .env ## Print a summary of the configuration environment
	@echo "[OTOPS] Summary of the configuration environment"
	$(eval include .env)
	@env | grep -E '^(OTOPS_)' | sort

release: ## Create a release folder for the data images to be stored
	@echo "[OTOPS] Creating release folder at '${OTOPS_PATH_RELEASE}'"
	@mkdir -p ${OTOPS_PATH_RELEASE}

download_release: download_clickhouse download_elastic_search download_webapp ## Collect all the artifacts that make up an Open Targets Platform Release
	@echo "[OTOPS] Collecting all the artifacts that make up an Open Targets Platform Release"

download_clickhouse: .env release ## Collect the Clickhouse data image
	@echo "[OTOPS] Collecting Clickhouse data image"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_COLLECT_DATA}) && ./$(shell basename ${OTOPS_COLLECT_DATA}) clickhouse

download_elastic_search: .env release ## Collect the Elastic Search data image
	@echo "[OTOPS] Collecting Elastic Search data image"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_COLLECT_DATA}) && ./$(shell basename ${OTOPS_COLLECT_DATA}) elastic

download_webapp: .env release # Collect the web app bundle
	@echo "[OTOPS] Collecting web app bundle"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_COLLECT_DATA}) && ./$(shell basename ${OTOPS_COLLECT_DATA}) webapp

deployment: ## Create a deployment folder where Open Targets Platform provisioners will deposit their artifacts
	@echo "[OTOPS] Creating deployment folder at '${OTOPS_PATH_DEPLOYMENT}'"
	@mkdir -p ${OTOPS_PATH_DEPLOYMENT}

deploy_clickhouse: deployment download_clickhouse ## Deploy ClickHouse
	@echo "[OTOPS] Provisioning Clickhouse data store"
	@cd $(shell dirname ${OTOPS_PROVISIONER_CLICKHOUSE}) && ./$(shell basename ${OTOPS_PROVISIONER_CLICKHOUSE})

deploy_elastic_search: deployment download_elastic_search ## Deploy Elastic Search
	@echo "[OTOPS] Provisioning Elastic Search data store"
	@cd $(shell dirname ${OTOPS_PROVISIONER_ELASTIC_SEARCH}) && ./$(shell basename ${OTOPS_PROVISIONER_ELASTIC_SEARCH})

deploy_webapp: .env deployment download_webapp ## Deploy the Open Targets Platform Webapp
	@echo "[OTOPS] Provisioning Open Targets Platform Webapp"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_PROVISIONER_WEBAPP}) && ./$(shell basename ${OTOPS_PROVISIONER_WEBAPP})

deploy: deployment deploy_clickhouse deploy_elastic_search deploy_webapp ## Deploy an Open Targets Platform Release, according to the active configuration profile
	@echo "[OTOPS] Deploying an Open Targets Platform Release"

clean_clickhouse: ## Clean the ClickHouse deployment
	@echo "[OTOPS] Cleaning ClickHouse deployment at '${OTOPS_PATH_DEPLOYMENT_CLICKHOUSE}'"
	@rm -rf ${OTOPS_PATH_DEPLOYMENT_CLICKHOUSE}

clean_elastic_search: ## Clean the Elastic Search deployment
	@echo "[OTOPS] Cleaning Elastic Search deployment at '${OTOPS_PATH_DEPLOYMENT_ELASTIC_SEARCH}'"
	@rm -rf ${OTOPS_PATH_DEPLOYMENT_ELASTIC_SEARCH}

clean_webapp: ## Clean the Open Targets Platform Webapp deployment
	@echo "[OTOPS] Cleaning Open Targets Platform Webapp deployment at '${OTOPS_PATH_DEPLOYMENT_WEBAPP}'"
	@rm -rf ${OTOPS_PATH_DEPLOYMENT_WEBAPP}

clean_deployments: ## Clean all deployments stores
	@echo "[OTOPS] Cleaning all deployments stores"
	@rm -rf ${OTOPS_PATH_DEPLOYMENT}

clean_clickhouse_release: ## Clean the Clickhouse data image
	@echo "[OTOPS] Cleaning the Clickhouse data image"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_CLEAN_DATA}) && ./$(shell basename ${OTOPS_CLEAN_DATA}) clickhouse

clean_elastic_search_release: ## Clean the Elastic search data image
	@echo "[OTOPS] Cleaning the Elastic search data image"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_CLEAN_DATA}) && ./$(shell basename ${OTOPS_CLEAN_DATA}) elastic

clean_webapp_release: ## Clean the Open Targets Platform Webapp release bundle
	@echo "[OTOPS] Cleaning Open Targets Platform Webapp release bundle"
	$(eval include .env)
	@cd $(shell dirname ${OTOPS_CLEAN_DATA}) && ./$(shell basename ${OTOPS_CLEAN_DATA}) webapp

clean_release: clean_clickhouse_release clean_elastic_search_release clean_webapp_release ## Clean all data images and webapp bundle
	@echo "[OTOPS] Cleaning all data images and webapp bundle"
	@rm -rf ${OTOPS_PATH_RELEASE}

clean_env: ## Clean the .env file
	@echo "[OTOPS] Cleaning the .env file"
	@rm -f .env

platform_up: deploy ## Bring up an Open Targets Platform deployment
	@echo "[OTOPS] Bringing up an Open Targets Platform deployment"
	docker compose -f docker-compose.yml up -d --wait

platform_down: ## Tear down an Open Targets Platform deployment
	@echo "[OTOPS] Tearing down an Open Targets Platform deployment"
	docker compose -f docker-compose.yml down

clean: clean_profile clean_deployments clean_release clean_env ## Clean the active configuration profile, all deployments stores and data images
	@echo "[OTOPS] Cleaning the active configuration profile, all deployments stores and data images"

.PHONY: help clean_clickhouse_release clean_elastic_search_release clean_webapp_release download_release download_clickhouse download_elastic_search download_webapp set_profile clean clean_profile summary_environment deploy_clickhouse deploy_elastic_search deploy deploy_webapp deploy clean_clickhouse clean_elastic_search clean_webapp clean_deployments clean_env platform_up platform_down

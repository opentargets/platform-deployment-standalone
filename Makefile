.PHONY: run clean stop destroy

# default build target
run: ./bin/configure
	@./bin/run.sh

./bin/configure:
	@cd src && go build -o ../bin/configure cmd/main.go

clean:
	@rm -f bin/configure && cd src && go clean

stop:
	@for d in $$(ls -d deployment-local-*); do docker -l error compose -f $$d/compose.yaml down; done

destroy:
	@cd $(deployment) && set -a; . ./config; set +a; TF_VAR_OT_RELEASE="" terraform destroy

help:
	@echo "Usage:"
	@echo "  make         - Deploy an Open Targets platform instance."
	@echo "  make clean   - Clean up shop (leaves the deployments and downloads intact)."
	@echo "  make stop    - Stop local deployments."
	@echo "  make destroy - Destroy cloud deployments before the end of their lifetime. This is not
	@echo "                   usually needed, only in case you want to get rid of a deployment right
	@echo "                   away because you made a mistake. It requires the `deployment` argument
	@echo "                   to be passed like: `make destroy deployment=<path>` with the path being
	@echo "                   the path to the deployment directory. E.g.:"
	@echo "                   make destroy deployment=deployment-cloud-25.06-beef"
	@echo "  make help    - Show this help message"

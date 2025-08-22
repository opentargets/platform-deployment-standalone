.PHONY: clean

# default build target
build:
	@cd src && go build -o ../platform main.go

clean:
	@rm -f platform && cd src && go clean

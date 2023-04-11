APP      = ./bin/tui-dictionary
ARGS := test
##@ Run

.PHONY: run
run: ## run server
	go run main.go
search: ## single search
	go run client/my_prefer/main.go $(ARGS)
##@ Build
.PHONY: build build-windows

build: ## build server binary for linux
	GOOS=linux go build -race -o ${APP} main.go
	
build-windows: ## build server binary for windows
	GOOS=windows GOARCH=amd64 go build -race -o ${APP}.exe main.go

lint-install: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.52.2

lint:
	./bin/golangci-lint run ./...

trivy-install:
	curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b ./bin v0.39.1

scan:
	./bin/trivy fs .

gosec-install:
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s v2.15.0
	
gosec:
	./bin/gosec ./...

##@ Help

.PHONY: help

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

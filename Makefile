APP      = ./bin/dict
TARGET = $$(awk -F "=" '/target/ {print $$2}' app.ini)
##@ Run
.PHONY: run
run: ## run server
	go run main.go
##@ Build
.PHONY: build
build: ## build server binary
	go build -o ${APP} main.go

target:  ## Show version
	echo ${TARGET}

lint-install: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.40.1
	# curl -sSO - https://github.com/golangci/golangci-lint/releases/download/v1.40.1/golangci-lint-1.40.1-linux-amd64.tar.gz | tar -xaf

test-lint:
	./bin/golangci-lint run ./...
##@ Help

.PHONY: help

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

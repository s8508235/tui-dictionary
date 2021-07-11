APP      = ./bin/dict
TARGET = $$(awk -F "=" '/target/ {print $$2}' app.ini)
ARGS := test
##@ Run

.PHONY: run
run: ## run server
	go run main.go
search: ## single search
	go run client/main.go $(ARGS)
##@ Build
.PHONY: build build-windows

build: ## build server binary for linux
	GOOS=linux go build -o ${APP} main.go
	
# https://stackoverflow.com/questions/49078510/trouble-compiling-windows-dll-using-golang-1-10
build-windows: ## build server binary for windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o ${APP}.exe main.go

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

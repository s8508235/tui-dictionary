TARGET = $$(awk -F "=" '/target/ {print $$2}' app.ini)
##@ Run
.PHONY: run
run: ## run server
	go run main.go
##@ Build
.PHONY: build
build: ## build server binary
	go build -o ./bin/main main.go
##@ Search
.PHONY: search
search: ## search dictionary word definition
	while true ; do \
		echo "input: "; \
		read WORD; \
		curl "localhost:8087/search/$$WORD"; \
	done

target:  ## Show version
	echo ${TARGET}
##@ Help

.PHONY: help

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

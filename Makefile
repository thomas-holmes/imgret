.DEFAULT_GOAL := help
.PHONY: help build run clean test vet ci ci-build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-12s %s\\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the imgret binary
	go build -o imgret cmd/imgret/main.go

run: ## Run the application locally
	go run cmd/imgret/main.go

clean: ## Clean build artifacts and coverage files
	rm -f imgret imgret-* coverage.out

test: ## Run tests
	go test ./...

vet: ## Run go vet
	go vet ./...

ci: ## Run CI checks locally (verify, vet, test, build)
	go mod verify
	go vet ./...
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go build -o imgret cmd/imgret/main.go

ci-build: ## Build binaries for multiple platforms
	GOOS=linux GOARCH=amd64 go build -o imgret-linux-amd64 cmd/imgret/main.go
	GOOS=darwin GOARCH=amd64 go build -o imgret-darwin-amd64 cmd/imgret/main.go
	GOOS=windows GOARCH=amd64 go build -o imgret-windows-amd64.exe cmd/imgret/main.go

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

imgret is a Go web service that generates unique visual representations (images) based on SHA-256 hashes. It creates colorful bitmap patterns from any input string by hashing the input and converting the hash into a distinctive image pattern.

## Development Commands

### Running the Application
```bash
# Run locally (default port :30000)
go run cmd/imgret/main.go

# Run with custom bind address
go run cmd/imgret/main.go -bind=:8080

# Using Makefile
make run
```

### Building and Testing
```bash
# Build the application
go build -o imgret cmd/imgret/main.go
# Or using Makefile
make build

# Run the built binary
./imgret -bind=:8080

# Run tests
go test ./...
# Or using Makefile
make test

# Format code
go fmt ./...

# Vet code for issues
go vet ./...
# Or using Makefile
make vet

# Clean build artifacts
make clean
```

### Dependencies
The project uses Go modules for dependency management:
```bash
# Download dependencies
go mod download

# Update dependencies
go get -u ./...

# Tidy dependencies (remove unused, add missing)
go mod tidy

# View dependency graph
go mod graph
```

## Architecture

### Core Components

1. **HTTP Server** (`cmd/imgret/main.go`): Single-file web service with main handler
2. **Image Generation** (`bitPNG` struct): Custom image generator that creates 1024x1024 pixel images from SHA-256 hashes
3. **Caching Layer**: Redis-backed caching with fallback to no-cache mode
4. **Metrics Collection**: Librato integration for performance monitoring

### Key Patterns

- **Hash-to-Image Algorithm**: Uses SHA-256 to generate deterministic visual patterns
  - Each bit in the hash controls pixel color and positioning
  - Additional bit manipulation creates color variations (purple, blue, red channels)
  - Results in unique, reproducible images for identical inputs

- **Caching Strategy**: Two-tier caching implementation
  - `redisCache`: Redis-backed persistent cache
  - `noCache`: Fallback when Redis is unavailable
  - Cache key is the URL path (the input string)

- **Metrics Integration**: 
  - `encode.time`: Histogram tracking image generation duration
  - `cache.hit`/`cache.miss`: Counters for cache performance

### Environment Configuration

Required environment variables:
- `REDIS_URL`: Redis connection string for caching
- `LIBRATO_USER`: Librato username for metrics
- `LIBRATO_TOKEN`: Librato API token for metrics
- `PORT`: Server port (Heroku deployment)

### Heroku Deployment

- `Procfile`: Defines web process command (`./imgret -bind=:$PORT`)
- `go.mod`: Specifies Go version 1.23.0 with toolchain go1.23.12
- Uses Go modules for dependency management
- Build process: `go build -o imgret cmd/imgret/main.go`

### Image Generation Details

The `bitPNG` struct implements Go's `image.Image` interface:
- 16x16 grid scaled to 1024x1024 pixels (64x multiplier)
- Each hash byte controls 8 pixels through bit manipulation
- Color algorithm uses multiple bit positions for RGB channel variations
- Base colors: purple (128,0,128) with algorithmic modifications

## Code Conventions

- Single-file architecture in `cmd/imgret/main.go`
- Uses structured logging with logrus v1.9.3
- Error handling with panic for unrecoverable errors
- Interface-based design for cache abstraction with context support
- Environment-based configuration with `envdecode`
- Redis client upgraded to v8 with context-aware operations

## Major Dependencies

- `github.com/sirupsen/logrus` v1.9.3: Structured logging
- `github.com/go-redis/redis/v8` v8.11.5: Redis client with context support
- `github.com/go-kit/kit` v0.13.0: Microservice toolkit for metrics
- `github.com/heroku/x`: Heroku-specific utilities for Librato metrics
- `github.com/joeshaw/envdecode`: Environment variable decoding

## Continuous Integration

The project uses GitHub Actions for CI/CD:

- **Workflow**: `.github/workflows/ci.yml`
- **Testing**: Runs on Go versions 1.21.x, 1.22.x, and 1.23.x
- **Steps**: Downloads dependencies, runs `go vet`, executes tests with race detection and coverage
- **Cross-compilation**: Builds binaries for Linux, macOS, and Windows
- **Artifacts**: Uploads cross-platform binaries for each successful build
- **Coverage**: Reports test coverage (currently 0% - no unit tests exist)

### CI Commands
```bash
# Run the same checks locally as CI does
go mod verify
go vet ./...
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go build -o imgret cmd/imgret/main.go

# Or use the convenient Makefile target
make ci
```

### Additional Makefile Targets
```bash
# Show all available targets
make help

# Build cross-platform binaries
make ci-build

# Clean all artifacts
make clean
```
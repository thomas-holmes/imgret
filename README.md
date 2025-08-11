# imgret

[![CI](https://github.com/thomas-holmes/imgret/workflows/CI/badge.svg)](https://github.com/thomas-holmes/imgret/actions)

A Go web service that generates unique visual representations (images) based on SHA-256 hashes. Creates colorful bitmap patterns from any input string by hashing the input and converting the hash into a distinctive image pattern.

## Quick Start

```bash
# Run locally
go run cmd/imgret/main.go

# Build and run
make build
./imgret -bind=:8080
```

## Development

See [CLAUDE.md](CLAUDE.md) for detailed development instructions.

## Features

- Generates deterministic 1024x1024 pixel images from any input string
- Redis caching for improved performance
- Librato metrics integration
- Heroku deployment ready
- Cross-platform builds (Linux, macOS, Windows)
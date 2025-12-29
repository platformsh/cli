# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Upsun CLI is a Go-based command-line interface for Upsun (formerly Platform.sh). The CLI is a hybrid system that wraps a legacy PHP CLI while providing new Go-based commands. It supports multiple vendors through build tags and configuration files.

## Build and Test Commands

Build a single binary for your platform:
```bash
make single
```

Build a snapshot for all platforms:
```bash
make snapshot
```

Run tests:
```bash
make test
# or directly:
GOEXPERIMENT=jsonv2 go test -v -race -cover -count=1 ./...
```

Run linters:
```bash
make lint
# or individual linters:
make lint-gomod
make lint-golangci
```

Format code:
```bash
go fmt ./...
```

Tidy dependencies:
```bash
go mod tidy
```

Run a single test:
```bash
go test -v -run TestName ./path/to/package
```

## Architecture

### Hybrid CLI System

The CLI operates as a wrapper around a legacy PHP CLI:
- Go layer: Handles new commands (init, list, version, config:install, project:convert) and core infrastructure
- PHP layer: Legacy commands are proxied through `internal/legacy/CLIWrapper`
- The PHP CLI (platform.phar) is embedded at build time via go:embed

### Key Components

**Entry Point**: `cmd/platform/main.go`
- Loads configuration from YAML (embedded or external)
- Sets up Viper for environment variable handling
- Delegates to commands package

**Commands**: `commands/`
- `root.go`: Root command that sets up the Cobra CLI and delegates to legacy CLI when needed
- Native Go commands: init, list, version, config:install, project:convert, completion
- Unrecognized commands are passed to the legacy PHP CLI

**Configuration**: `internal/config/`
- `schema.go`: Config struct definition with validation tags
- Supports vendorization through embedded YAML configs (config_upsun.go, config_platformsh.go, config_vendor.go)
- Uses build tags to select which config is embedded
- Config can be loaded from external files for testing/development

**Legacy Integration**: `internal/legacy/`
- `legacy.go`: CLIWrapper that manages PHP binary and phar execution
- PHP binaries are embedded per platform via go:embed and build tags
- Uses file locking to prevent concurrent initialization
- Copies PHP binary and phar to cache directory on first run

**API Client**: `internal/api/`
- HTTP client for interacting with Platform.sh/Upsun API
- Handles authentication, organizations, and resource management

**Authentication**: `internal/auth/`
- JWT handling and OAuth2 flow
- Custom transport for API authentication

**Project Initialization**: `internal/init/`
- AI-powered project configuration generation
- Integrates with whatsun library for codebase analysis

### Build System

**Multi-Vendor Support**:
- Uses Go build tags (platform, upsun, vendor) to compile different binaries
- Configuration is embedded at compile time
- GoReleaser builds multiple variants (platform, upsun, vendor-specific)

**PHP Binary Handling**:
- PHP binaries are downloaded from [upsun/cli-php-builds](https://github.com/upsun/cli-php-builds) releases
- All platforms use static binaries built with [static-php-cli](https://github.com/crazywhalecc/static-php-cli)
- Supported platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Extensions included: curl, filter, openssl, pcntl (Unix), phar, posix (Unix), zlib
- Windows requires `cacert.pem` for OpenSSL (embedded separately)

**Downloading PHP Binaries**:
```bash
# Download PHP for current platform only (for development)
make php

# Download all PHP binaries (for release builds)
make download-php
```

**Upgrading PHP Version**:
1. Trigger the build workflow at [upsun/cli-php-builds](https://github.com/upsun/cli-php-builds/actions) with the new PHP version
2. Update `PHP_VERSION` in the Makefile
3. Run `make php` to download the new binary
4. Test and release

## Development Notes

### Testing

Tests use github.com/stretchr/testify for assertions. Table-driven tests are preferred with a "cases" slice containing simple test case structs.

### Configuration

The CLI uses Viper for configuration. Environment variables use the prefix defined in the config (UPSUN_CLI_ or PLATFORM_CLI_). The prefix is set in the config YAML.

### Legacy CLI Interaction

When the root command receives arguments it doesn't recognize, it passes them to the legacy PHP CLI via CLIWrapper.Exec(). The PHP binary and phar are extracted to a cache directory on first use.

### Vendorization

To build a vendor-specific CLI:
```bash
make vendor-snapshot VENDOR_NAME='Vendor Name' VENDOR_BINARY='vendorcli'
make vendor-release VENDOR_NAME='Vendor Name' VENDOR_BINARY='vendorcli'
```

This requires a config file at `internal/config/embedded-config.yaml` (downloaded at build time).

### Version Information

Version information is injected at build time via ldflags:
- `internal/config.Version`: Git tag/version
- `internal/config.Commit`: Git commit hash
- `internal/config.Date`: Build date
- `internal/legacy.PHPVersion`: PHP version embedded
- `internal/legacy.LegacyCLIVersion`: Legacy CLI version embedded

### Update Checks

The CLI checks for updates from GitHub releases (when Wrapper.GitHubRepo is set in config). This runs in a background goroutine and prints a message after command execution.

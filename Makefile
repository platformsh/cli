PHP_VERSION = 8.4.16

GORELEASER_ID ?= upsun

ifeq ($(GOOS), darwin)
	GORELEASER_ID=$(GORELEASER_ID)-macos
endif

GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH := $(shell uname -m)
ifeq ($(GOARCH), x86_64)
	GOARCH=amd64
endif
ifeq ($(GOARCH), aarch64)
	GOARCH=arm64
endif

VERSION := $(shell git describe --always)

# Tooling versions
GORELEASER_VERSION=v2.12.0

# PHP binaries are downloaded from cli-php-builds releases.
# See: https://github.com/upsun/cli-php-builds
PHP_BUILDS_REPO = upsun/cli-php-builds
PHP_RELEASE_URL = https://github.com/$(PHP_BUILDS_REPO)/releases/download/php-$(PHP_VERSION)

# Build the legacy CLI phar from the legacy/ subdirectory.
internal/legacy/archives/platform.phar: legacy/vendor/autoload.php
	mkdir -p internal/legacy/archives
	cd legacy && php -d phar.readonly=0 vendor/bin/box compile
	mv legacy/platform.phar internal/legacy/archives/platform.phar

legacy/vendor/autoload.php:
	cd legacy && composer install --no-dev --no-interaction

# Download PHP binary for the current platform.
internal/legacy/archives/php_darwin_$(GOARCH):
	mkdir -p internal/legacy/archives
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-darwin-$(GOARCH)" -o $@
	chmod +x $@

internal/legacy/archives/php_linux_$(GOARCH):
	mkdir -p internal/legacy/archives
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-linux-$(GOARCH)" -o $@
	chmod +x $@

internal/legacy/archives/php_windows_amd64: internal/legacy/archives/php_windows.exe internal/legacy/archives/cacert.pem

internal/legacy/archives/php_windows.exe:
	mkdir -p internal/legacy/archives
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-windows-amd64.exe" -o $@

.PHONY: internal/legacy/archives/cacert.pem
internal/legacy/archives/cacert.pem:
	mkdir -p internal/legacy/archives
	curl -fSL https://curl.se/ca/cacert.pem -o internal/legacy/archives/cacert.pem

# Download all PHP binaries (for release builds).
.PHONY: download-php
download-php:
	mkdir -p internal/legacy/archives
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-linux-amd64" -o internal/legacy/archives/php_linux_amd64
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-linux-arm64" -o internal/legacy/archives/php_linux_arm64
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-darwin-amd64" -o internal/legacy/archives/php_darwin_amd64
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-darwin-arm64" -o internal/legacy/archives/php_darwin_arm64
	curl -fSL "$(PHP_RELEASE_URL)/php-$(PHP_VERSION)-windows-amd64.exe" -o internal/legacy/archives/php_windows.exe
	curl -fSL https://curl.se/ca/cacert.pem -o internal/legacy/archives/cacert.pem
	chmod +x internal/legacy/archives/php_linux_* internal/legacy/archives/php_darwin_*

php: internal/legacy/archives/php_$(GOOS)_$(GOARCH)

.PHONY: goreleaser
goreleaser:
	command -v goreleaser >/dev/null || go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

.PHONY: single
single: goreleaser internal/legacy/archives/platform.phar php ## Build a single target release
	PHP_VERSION=$(PHP_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --clean

.PHONY: snapshot ## Build a snapshot release
snapshot: goreleaser internal/legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) goreleaser build --snapshot --clean

.PHONY: clean-phar
clean-phar: ## Clean up the legacy CLI phar
	rm -f internal/legacy/archives/platform.phar
	rm -rf legacy/vendor

.PHONY: release
release: goreleaser clean-phar internal/legacy/archives/platform.phar php ## Create and publish a release
	PHP_VERSION=$(PHP_VERSION) goreleaser release --clean
	VERSION=$(VERSION) bash cloudsmith.sh

.PHONY: test
# "We encourage users of encoding/json to test their programs with GOEXPERIMENT=jsonv2 enabled" (https://tip.golang.org/doc/go1.25)
test: ## Run unit tests
	GOEXPERIMENT=jsonv2 go test -v -race -cover -count=1 ./...

.PHONY: lint
lint: lint-gomod lint-golangci ## Run linters.

.PHONY: lint-gomod
lint-gomod:
	go mod tidy -diff

.PHONY: lint-golangci
lint-golangci:
	golangci-lint run --timeout=2m

.goreleaser.vendor.yaml: check-vendor ## Generate the goreleaser vendor config
	cat .goreleaser.vendor.yaml.tpl | envsubst > .goreleaser.vendor.yaml

.PHONY: check-vendor
check-vendor: ## Check that the vendor CLI variables are set
ifndef VENDOR_NAME
	$(error VENDOR_NAME is undefined)
endif
ifndef VENDOR_BINARY
	$(error VENDOR_BINARY is undefined)
endif

.PHONY: vendor-release
vendor-release:  check-vendor .goreleaser.vendor.yaml goreleaser clean-phar internal/legacy/archives/platform.phar php ## Release a vendor CLI
	PHP_VERSION=$(PHP_VERSION) VENDOR_BINARY="$(VENDOR_BINARY)" VENDOR_NAME="$(VENDOR_NAME)" goreleaser release --clean --config=.goreleaser.vendor.yaml

.PHONY: vendor-snapshot
vendor-snapshot: check-vendor .goreleaser.vendor.yaml goreleaser internal/legacy/archives/platform.phar php ## Build a vendor CLI snapshot
	PHP_VERSION=$(PHP_VERSION) VENDOR_BINARY="$(VENDOR_BINARY)" VENDOR_NAME="$(VENDOR_NAME)" goreleaser build --snapshot --clean --config=.goreleaser.vendor.yaml

.PHONY: goreleaser-check
goreleaser-check:  goreleaser ## Check the goreleaser configs
	PHP_VERSION=$(PHP_VERSION) goreleaser check --config=.goreleaser.yaml

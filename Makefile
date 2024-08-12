PHP_VERSION = 8.2.22
LEGACY_CLI_VERSION = 4.20.3

GORELEASER_ID ?= platform

ifeq ($(GOOS), darwin)
	GORELEASER_ID=$(GORELEASER_ID)-macos
endif

# The OpenSSL version must be compatible with the PHP version.
# See: https://www.php.net/manual/en/openssl.requirements.php
OPENSSL_VERSION = 1.1.1t

GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH := $(shell uname -m)
ifeq ($(GOARCH), x86_64)
	GOARCH=amd64
endif
ifeq ($(GOARCH), aarch64)
	GOARCH=arm64
endif

PHP_BINARY_PATH := internal/legacy/archives/php_$(GOOS)_$(GOARCH)
VERSION := $(shell git describe --always)

# Tooling versions
GORELEASER_VERSION=v1.26
GOLANGCI_LINT_VERSION=v1.59

internal/legacy/archives/platform.phar:
	curl -L https://github.com/platformsh/legacy-cli/releases/download/v$(LEGACY_CLI_VERSION)/platform.phar -o internal/legacy/archives/platform.phar

internal/legacy/archives/php_windows_amd64: internal/legacy/archives/php_windows.zip internal/legacy/archives/cacert.pem

internal/legacy/archives/php_darwin_$(GOARCH):
	bash build-php-brew.sh $(GOOS) $(PHP_VERSION) $(OPENSSL_VERSION)
	mv -f $(GOOS)/php-$(PHP_VERSION)/sapi/cli/php $(PHP_BINARY_PATH)
	rm -rf $(GOOS)

internal/legacy/archives/php_linux_$(GOARCH):
	cp ext/extensions.txt ext/static-php-cli/docker
	docker buildx build \
		--build-arg GOARCH=$(GOARCH) \
		--build-arg PHP_VERSION=$(PHP_VERSION) \
		--build-arg USE_BACKUP_ADDRESS=yes \
		--file=./Dockerfile.php \
		--platform=linux/$(GOARCH) \
		--output=type=local,dest=./internal/legacy/archives/ \
		--progress=plain \
		ext/static-php-cli/docker

internal/legacy/archives/php_windows.zip:
	mkdir -p internal/legacy/archives
	wget https://windows.php.net/downloads/releases/php-$(PHP_VERSION)-nts-Win32-vs16-x64.zip -O internal/legacy/archives/php_windows.zip

.PHONY: internal/legacy/archives/cacert.pem
internal/legacy/archives/cacert.pem:
	mkdir -p internal/legacy/archives
	wget https://curl.se/ca/cacert.pem -O internal/legacy/archives/cacert.pem

php: $(PHP_BINARY_PATH)

.PHONY: goreleaser
goreleaser:
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

.PHONY: single
single: goreleaser internal/legacy/archives/platform.phar php ## Build a single target release for Platform.sh or Upsun
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --clean

.PHONY: snapshot ## Build a snapshot release for Platform.sh and Upsun
snapshot: goreleaser internal/legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser build --snapshot --clean

.PHONY: clean-phar
clean-phar: ## Clean up the legacy CLI phar
	rm -f internal/legacy/archives/platform.phar

.PHONY: release
release: goreleaser clean-phar internal/legacy/archives/platform.phar php ## Release the Platform.sh and Upsun CLIs
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser release --clean
	VERSION=$(VERSION) bash cloudsmith.sh

.PHONY: test
test: ## Run unit tests
	go clean -testcache
	go test -v -race -mod=readonly -cover ./...

golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: golangci-lint ## Run linter
	golangci-lint run --timeout=10m --verbose

.goreleaser.vendor.yaml: check-vendor
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
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) VENDOR_BINARY="$(VENDOR_BINARY)" VENDOR_NAME="$(VENDOR_NAME)" goreleaser release --clean --config=.goreleaser.vendor.yaml

.PHONY: vendor-snapshot
vendor-snapshot: .goreleaser.vendor.yaml goreleaser internal/legacy/archives/platform.phar php ## Build a vendor CLI snapshot
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) VENDOR_BINARY="$(VENDOR_BINARY)" VENDOR_NAME="$(VENDOR_NAME)" goreleaser build --snapshot --clean --config=.goreleaser.vendor.yaml

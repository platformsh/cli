PHP_VERSION = 8.0.29
LEGACY_CLI_VERSION = 4.9.0

# Override these environment variables to build with alternative configuration.
CLI_CONFIG_FILE ?= config/platformsh-cli.yaml
GORELEASER_CONFIG_FILE ?= config/platformsh-cli-goreleaser.yaml

ifeq ($(GOOS), darwin)
	GORELEASER_ID=platform-macos
else
	GORELEASER_ID=platform
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

# Tooling versions
GORELEASER_VERSION=v1.20
GOLANGCI_LINT_VERSION=v1.52

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

internal/legacy/archives/cacert.pem:
	mkdir -p internal/legacy/archives
	wget https://curl.se/ca/cacert.pem -O internal/legacy/archives/cacert.pem

php: $(PHP_BINARY_PATH)

.PHONY: goreleaser
goreleaser:
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

.PHONY: copy-config-file
copy-config-file:
	cp "$(CLI_CONFIG_FILE)" internal/config/embedded-config.yaml

single: goreleaser internal/legacy/archives/platform.phar php copy-config-file
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser build --single-target --config="$(GORELEASER_CONFIG_FILE)" --id=$(GORELEASER_ID) --snapshot --clean

snapshot: goreleaser internal/legacy/archives/platform.phar php copy-config-file
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser build --snapshot --clean --config="$(GORELEASER_CONFIG_FILE)"

clean-phar:
	rm -f internal/legacy/archives/platform.phar

release: goreleaser clean-phar internal/legacy/archives/platform.phar php copy-config-file
	PHP_VERSION=$(PHP_VERSION) LEGACY_CLI_VERSION=$(LEGACY_CLI_VERSION) goreleaser release --clean --auto-snapshot --config="$(GORELEASER_CONFIG_FILE)"

.PHONY: test
test: ## Run unit tests
	go clean -testcache
	go test -v -race -mod=readonly -cover ./...

golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: golangci-lint ## Run linter
	golangci-lint run --timeout=10m --verbose

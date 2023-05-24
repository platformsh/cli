PHP_VERSION = 8.0.28
PSH_VERSION = 4.6.1
# The OpenSSL version must be compatible with the PHP version.
# See: https://www.php.net/manual/en/openssl.requirements.php
OPENSSL_VERSION = 1.1.1t
GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ifeq ($(GOOS), darwin)
	GORELEASER_ID=platform-macos
else
	GORELEASER_ID=platform
endif
GOARCH := $(shell uname -m)
ifeq ($(GOARCH), x86_64)
	GOARCH=amd64
endif
ifeq ($(GOARCH), aarch64)
	GOARCH=arm64
endif
PHP_BINARY_PATH := internal/legacy/archives/php_$(GOOS)_$(GOARCH)

internal/legacy/archives/platform.phar:
	curl -L https://github.com/platformsh/legacy-cli/releases/download/v$(PSH_VERSION)/platform.phar -o internal/legacy/archives/platform.phar

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

single: internal/legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --clean

snapshot: internal/legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --snapshot --clean

clean-phar:
	rm -f internal/legacy/archives/platform.phar

release: clean-phar internal/legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser release --clean --auto-snapshot

.PHONY: test
test: ## Run unit tests
	go clean -testcache
	go test -v -race -mod=readonly -cover ./...

.PHONY: lint
lint: ## Run linter
	command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=10m --verbose

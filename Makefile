PHP_VERSION = 8.0.23
PSH_VERSION = 3.82.2
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
PHP_BINARY_PATH := legacy/archives/php_$(GOOS)_$(GOARCH)

legacy/archives/platform.phar:
	wget https://github.com/platformsh/platformsh-cli/releases/download/v$(PSH_VERSION)/platform.phar -O legacy/archives/platform.phar

legacy/archives/php_windows_amd64: legacy/archives/php_windows.zip legacy/archives/cacert.pem

legacy/archives/php_darwin_$(GOARCH): php-unix

legacy/archives/php_linux_$(GOARCH): php-unix

php-unix:
	bash build-php-brew.sh $(PHP_VERSION) $(GOOS)
	mv -f $(GOOS)/php-$(PHP_VERSION)/sapi/cli/php $(PHP_BINARY_PATH)
	rm -rf $(GOOS)

legacy/archives/php_windows.zip:
	mkdir -p legacy/archives
	wget https://windows.php.net/downloads/releases/php-$(PHP_VERSION)-nts-Win32-vs16-x64.zip -O legacy/archives/php_windows.zip

legacy/archives/cacert.pem:
	wget https://curl.se/ca/cacert.pem -O legacy/archives/cacert.pem

php: $(PHP_BINARY_PATH)

snapshot: legacy/archives/platform.phar
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --snapshot --rm-dist --id=$(GORELEASER_ID)

single: php legacy/archives/platform.phar
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --rm-dist

release: legacy/archives/platform.phar
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser release --rm-dist

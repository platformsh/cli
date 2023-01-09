PHP_VERSION = 8.0.27
PSH_VERSION = 4.0.1
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
PHP_BINARY_PATH := legacy/archives/php_$(GOOS)_$(GOARCH)

legacy/archives/platform.phar:
	wget https://github.com/platformsh/legacy-cli/releases/download/v$(PSH_VERSION)/platform.phar -O legacy/archives/platform.phar

legacy/archives/php_windows_amd64: legacy/archives/php_windows.zip legacy/archives/cacert.pem

legacy/archives/php_darwin_$(GOARCH):
	bash build-php-brew.sh $(PHP_VERSION) $(GOOS)
	mv -f $(GOOS)/php-$(PHP_VERSION)/sapi/cli/php $(PHP_BINARY_PATH)
	rm -rf $(GOOS)

legacy/archives/php_linux_$(GOARCH):
	cp ext/extensions.txt ext/static-php-cli/docker
	cd ext/static-php-cli/docker ;\
	sed -i 's/alpine:latest/alpine:3.16/g' Dockerfile;\
	docker build --platform=linux/$(GOARCH) -t static-php . --build-arg USE_BACKUP_ADDRESS=yes --progress=plain
	mkdir -p legacy/archives
	docker run --rm --platform=linux/$(GOARCH) -v ${PWD}/legacy/archives:/dist -e USE_BACKUP_ADDRESS=yes static-php build-php no-mirror $(PHP_VERSION) all /dist
	mv -f legacy/archives/php legacy/archives/php_linux_$(GOARCH)
	rm -f legacy/archives/micro.sfx

legacy/archives/php_windows.zip:
	mkdir -p legacy/archives
	wget https://windows.php.net/downloads/releases/php-$(PHP_VERSION)-nts-Win32-vs16-x64.zip -O legacy/archives/php_windows.zip

legacy/archives/cacert.pem:
	mkdir -p legacy/archives
	wget https://curl.se/ca/cacert.pem -O legacy/archives/cacert.pem

php: $(PHP_BINARY_PATH)

single: legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --rm-dist

snapshot: legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --snapshot --rm-dist

release: legacy/archives/platform.phar php
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser release --rm-dist --auto-snapshot

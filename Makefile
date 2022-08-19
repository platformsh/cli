PHP_VERSION = 8.0.22
PSH_VERSION = 3.81.0
GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH := $(shell uname -m)
ifeq ($(GOOS), darwin)
	GORELEASER_ID=psh-go-macos
else
	GORELEASER_ID=psh-go
endif
PHP_BINARY_PATH := legacy/archives/php_$(GOOS)_$(GOARCH)

legacy/archives/platform.phar:
	wget https://github.com/platformsh/platformsh-cli/releases/download/v$(PSH_VERSION)/platform.phar -O legacy/archives/platform.phar

legacy/archives/php_linux_$(GOARCH):
	cp ext/extensions.txt ext/static-php-cli/docker
	cd ext/static-php-cli/docker ;\
	docker build -t static-php . --build-arg USE_BACKUP_ADDRESS=yes
	mkdir -p legacy/archives
	docker run --rm -v ${PWD}/legacy/archives:/dist -e USE_BACKUP_ADDRESS=yes static-php build-php no-mirror $(PHP_VERSION) all /dist
	mv -f legacy/archives/php legacy/archives/php_linux_$(GOARCH)
	rm -f legacy/archives/micro.sfx

legacy/archives/php_darwin_$(GOARCH):
	bash build-php-macos.sh $(PHP_VERSION) $(GOOS)
	mv -f $(GOOS)/php-$(PHP_VERSION)/sapi/cli/php $(PHP_BINARY_PATH)
	rm -rf $(GOOS)

legacy/archives/php_windows_$(GOARCH): legacy/archives/php_windows.zip legacy/archives/cacert.pem

legacy/archives/php_windows.zip:
	wget https://windows.php.net/downloads/releases/php-$(PHP_VERSION)-nts-Win32-vs16-x64.zip -O legacy/archives/php_windows.zip

legacy/archives/cacert.pem:
	wget https://curl.se/ca/cacert.pem -O legacy/archives/cacert.pem

php: $(PHP_BINARY_PATH)

snapshot: legacy/archives/platform.phar
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --snapshot --rm-dist --id=$(GORELEASER_ID)

single: php legacy/archives/platform.phar
	PHP_VERSION=$(PHP_VERSION) PSH_VERSION=$(PSH_VERSION) goreleaser build --single-target --id=$(GORELEASER_ID) --snapshot --rm-dist

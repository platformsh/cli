PHP_VERSION = 8.0.21
PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]')
PHP_BINARY_PATH := legacy/archives/php_$(PLATFORM)_$(uname -m)

php-linux:
	cp ext/extensions.txt ext/static-php-cli/docker
	cd ext/static-php-cli/docker ;\
	docker build -t static-php . --build-arg USE_BACKUP_ADDRESS=yes
	mkdir -p legacy/archives
	docker run --rm -v ${PWD}/legacy/archives:/dist -e USE_BACKUP_ADDRESS=yes static-php build-php no-mirror $(PHP_VERSION) all /dist
	mv -f legacy/archives/php legacy/archives/php_linux_`uname -m`
	rm -f legacy/archives/micro.sfx

legacy/archives/php_darwin_$(uname -m):
	bash build-php-macos.sh $(PHP_VERSION) $(PLATFORM)
	mv -f $(PLATFORM)/php-$(PHP_VERSION)/sapi/cli/php $(PHP_BINARY_PATH)
	rm -rf $(PLATFORM)

php: $(PHP_BINARY_PATH)

snapshot:
	goreleaser build --snapshot --rm-dist

single: php
	goreleaser build --single-target --snapshot --rm-dist

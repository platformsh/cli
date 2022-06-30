PHP_VERSION=8.0.20

php-linux:
	cp ext/extensions.txt ext/static-php-cli/docker
	cd ext/static-php-cli/docker ;\
	docker build -t static-php . --build-arg USE_BACKUP_ADDRESS=yes
	mkdir -p legacy/archives
	docker run --rm -v ${PWD}/legacy/archives:/dist -e USE_BACKUP_ADDRESS=yes static-php build-php no-mirror $(PHP_VERSION) all /dist
	mv -f legacy/archives/php legacy/archives/php_linux_`uname -m`
	rm -f legacy/archives/micro.sfx

snapshot:
	goreleaser release --snapshot --rm-dist

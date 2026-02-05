.PHONY: composer-dev
composer-dev:
	composer install --no-interaction

.PHONY: clean
clean:
	rm config/cache/container.php

.PHONY: lint-phpstan
lint-phpstan: composer-dev
	./vendor/bin/phpstan analyse

.PHONY: lint-php-cs-fixer
lint-php-cs-fixer: composer-dev
	./vendor/bin/php-cs-fixer check --config .php-cs-fixer.dist.php --diff

.PHONY: lint
lint: lint-php-cs-fixer lint-phpstan

.PHONY: test
test:
	./vendor/bin/phpunit --exclude-group slow

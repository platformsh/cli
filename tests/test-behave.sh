#!/usr/bin/env bash
set -e

echo ""
echo "Copying Go CLI to /usr/local/bin ..."

cp dist/$PATH_CLI /usr/local/bin/pshgo
cd tests/

echo ""
echo "All set. Running tests for legacy CLI..."

export cli="platform"
behave --color --tags=-gocli

echo ""
echo "Running tests for GO CLI..."

export cli="pshgo"
behave --color --tags=-legacycli

echo ""
echo "All good."
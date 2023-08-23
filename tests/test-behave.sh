#!/usr/bin/env bash
set -e

echo ""

echo "Copying Go CLI to $HOME/.local/bin ..."
mkdir -p $HOME/.local/bin
export PATH=$HOME/.local/bin:$PATH
cp dist/$PATH_CLI $HOME/.local/bin/pshgo

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

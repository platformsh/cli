#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -e
cd "$DIR/.."
make clean-phar platform
export data_dir="$(realpath ./tests/integration/data)"
export TEST_CLI_PATH="$(realpath ./platform)"
export CLI_CONFIG_FILE="$(realpath ./tests/integration/config.yaml)"
export TEST_CLI_NO_INTERACTION=1

# Clear the working directory to avoid this repository being detected as a "project".
cd /
echo "Using command: $TEST_CLI_PATH"
"$TEST_CLI_PATH" version
for c in help list 'help list' 'help create'; do
  export filename="$data_dir"/"$c".stdout.txt
  echo "Generating $filename"
  COLUMNS=120 "$TEST_CLI_PATH" $c > "$filename"
done

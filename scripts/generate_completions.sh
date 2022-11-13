set -e

rm -rf completion
mkdir -p completion/bash completion/zsh
go run main.go completion bash > completion/bash/platform.bash
go run main.go completion zsh > completion/zsh/_platform

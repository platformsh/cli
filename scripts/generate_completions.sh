set -e

rm -rf completion
mkdir -p completion/bash completion/zsh
go run cmd/platform/main.go completion bash > completion/bash/platform.bash
go run cmd/platform/main.go completion zsh > completion/zsh/_platform

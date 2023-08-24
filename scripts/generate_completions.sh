set -e

rm -rf completion
mkdir -p completion/bash completion/zsh
go run cmd/platform/main.go completion bash > completion/bash/platform.bash
go run cmd/platform/main.go completion zsh > completion/zsh/_platform

go run --tags=vendor,upsun cmd/platform/main.go completion bash > completion/bash/upsun.bash
go run --tags=vendor,upsun cmd/platform/main.go completion zsh > completion/zsh/_upsun

# if $VENDOR_BINARY is not empty
if [ -nz "$VENDOR_BINARY" ]; then
    go run --tags=vendor cmd/platform/main.go completion bash > completion/bash/$VENDOR_BINARY.bash
    go run --tags=vendor cmd/platform/main.go completion zsh > completion/zsh/_$VENDOR_BINARY
fi

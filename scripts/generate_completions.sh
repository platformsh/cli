set -e

rm -rf completion
mkdir -p completion/bash completion/zsh

# Upsun completions (default, no build tags)
go run cmd/platform/main.go completion bash > completion/bash/upsun.bash
go run cmd/platform/main.go completion zsh > completion/zsh/_upsun

# Platform.sh completions (requires platformsh build tag)
go run --tags=platformsh cmd/platform/main.go completion bash > completion/bash/platform.bash
go run --tags=platformsh cmd/platform/main.go completion zsh > completion/zsh/_platform

# Vendor completions (if $VENDOR_BINARY is not empty)
if [ -n "$VENDOR_BINARY" ]; then
    go run --tags=vendor cmd/platform/main.go completion bash > completion/bash/$VENDOR_BINARY.bash
    go run --tags=vendor cmd/platform/main.go completion zsh > completion/zsh/_$VENDOR_BINARY
fi

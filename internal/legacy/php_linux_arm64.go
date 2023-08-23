package legacy

import (
	_ "embed"
)

//go:embed archives/php_linux_arm64
var phpCLI []byte

//go:embed archives/php_linux_arm64.sha256
var phpCLIHash string

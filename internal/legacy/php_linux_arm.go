package legacy

import (
	_ "embed"
)

//go:embed archives/php_linux_arm
var phpCLI []byte

//go:embed archives/php_linux_arm.sha256
var phpCLIHash string

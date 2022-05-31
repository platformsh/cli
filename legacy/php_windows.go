package legacy

import (
	_ "embed"
)

//go:embed archives/php_windows.exe
var phpCLI []byte

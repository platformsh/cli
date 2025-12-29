//go:build vendor

package config

import _ "embed"

//go:embed embedded-config.yaml
var embedded []byte

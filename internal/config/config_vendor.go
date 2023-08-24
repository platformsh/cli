//go:build vendor && !upsun
// +build vendor,!upsun

package config

import _ "embed"

//go:embed embedded-config.yaml
var embedded []byte

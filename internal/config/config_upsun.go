//go:build vendor && upsun
// +build vendor,upsun

package config

import _ "embed"

//go:embed upsun-cli.yaml
var embedded []byte

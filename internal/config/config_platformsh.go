//go:build !vendor
// +build !vendor

package config

import _ "embed"

//go:embed platformsh-cli.yaml
var embedded []byte

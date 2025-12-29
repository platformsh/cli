//go:build !platformsh && !vendor

package config

import _ "embed"

//go:embed upsun-cli.yaml
var embedded []byte

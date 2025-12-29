//go:build platformsh && !vendor

package config

import _ "embed"

//go:embed platformsh-cli.yaml
var embedded []byte

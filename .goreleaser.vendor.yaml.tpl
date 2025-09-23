# GoReleaser configuration for the ${VENDOR_NAME} CLI.
version: 2
project_name: ${VENDOR_BINARY}

before:
  hooks:
    - go generate ./...
    - bash scripts/generate_completions.sh

builds:
  - binary: ${VENDOR_BINARY}
    id: ${VENDOR_BINARY}
    env:
      - CGO_ENABLED=0
    tags:
      - vendor
    goos:
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X "github.com/platformsh/cli/internal/legacy.PHPVersion={{.Env.PHP_VERSION}}"
      - -X "github.com/platformsh/cli/internal/legacy.LegacyCLIVersion={{.Env.LEGACY_CLI_VERSION}}"
      - -X "github.com/platformsh/cli/internal/config.Version={{.Version}}"
      - -X "github.com/platformsh/cli/internal/config.Commit={{.Commit}}"
      - -X "github.com/platformsh/cli/internal/config.Date={{.Date}}"
      - -X "github.com/platformsh/cli/internal/config.Vendor=${VENDOR_BINARY}"
      - -X "github.com/platformsh/cli/internal/config.BuiltBy=goreleaser"
    main: ./cmd/platform
  - binary: ${VENDOR_BINARY}
    id: ${VENDOR_BINARY}-macos
    env:
      - CGO_ENABLED=0
    tags:
      - vendor
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X "github.com/platformsh/cli/internal/legacy.PHPVersion={{.Env.PHP_VERSION}}"
      - -X "github.com/platformsh/cli/internal/legacy.LegacyCLIVersion={{.Env.LEGACY_CLI_VERSION}}"
      - -X "github.com/platformsh/cli/internal/config.Version={{.Version}}"
      - -X "github.com/platformsh/cli/internal/config.Commit={{.Commit}}"
      - -X "github.com/platformsh/cli/internal/config.Date={{.Date}}"
      - -X "github.com/platformsh/cli/internal/config.Vendor=${VENDOR_BINARY}"
      - -X "github.com/platformsh/cli/internal/config.BuiltBy=goreleaser"
    main: ./cmd/platform

checksum:
  name_template: checksums.txt

snapshot:
  version_template: '{{ incpatch .Version }}-{{ .Now.Format "2006-01-02" }}-{{ .ShortCommit }}-next'

universal_binaries:
  - id: ${VENDOR_BINARY}-macos
    name_template: ${VENDOR_BINARY}
    replace: true

archives:
  - name_template: "${VENDOR_BINARY}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - completion/*
    format_overrides:
      - goos: windows
        formats: [zip]

nfpms:
  - homepage: https://docs.upsun.com/anchors/fixed/cli/
    package_name: ${VENDOR_BINARY}-cli
    description: ${VENDOR_NAME} CLI
    maintainer: Antonis Kalipetis <antonis.kalipetis@upsun.com>
    license: MIT
    vendor: Platform.sh
    ids:
      - ${VENDOR_BINARY}
    formats:
      - apk
      - deb
      - rpm
    contents:
      - src: completion/bash/${VENDOR_BINARY}.bash
        dst: /etc/bash_completion.d/${VENDOR_BINARY}
      - src: completion/zsh/_${VENDOR_BINARY}
        dst: /usr/local/share/zsh/site-functions/_${VENDOR_BINARY}

release:
  disable: true

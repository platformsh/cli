# Upsun CLI

The **Upsun CLI** is the official command-line interface for [Upsun](https://upsun.com).

This repository hosts the source code and releases of the CLI.

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Install

To install the CLI, use either [Homebrew](https://brew.sh/) (on Linux, macOS, or the Windows Subsystem for Linux) or [Scoop](https://scoop.sh/) (on Windows):

### HomeBrew

```console
brew install platformsh/tap/upsun-cli
```

### Scoop

```console
scoop bucket add platformsh https://github.com/platformsh/homebrew-tap.git
scoop install upsun
```

### Bash installer

Use the bash installer for an automated installation, using the most preferable way for your system.

```console
curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | bash
```

The installer is configurable using the following environment variables:

* `INSTALL_LOG` - the install log file
* `INSTALL_METHOD` - force a specific installation method, possible values are `brew` and `raw`
* `INSTALL_DIR` - the installation directory for the `raw` installation method, for example you can use `INSTALL_DIR=$HOME/.local/bin` for a single user installation
* `VERSION` - the version of the CLI to install, if you need a version other than the latest one

#### Installation configuration examples

<details>
    <summary>Force the CLI to be installed using the raw method</summary>

    curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | INSTALL_METHOD=raw bash
</details>

<details>
    <summary>Install a specific version</summary>

    curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | VERSION=5.0.0 bash
</details>

<details>
    <summary>Install the CLI in a user owned directory</summary>

    curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | INSTALL_METHOD=raw INSTALL_DIR=$HOME/.local/bin bash
</details>

### Nix profile

Refer to the [Nix documentation if you are not on NixOS](https://nix.dev/manual/nix/2.24/installation/installing-binary.html).

```console
nix profile install nixpkgs#upsun
```

### Alpine

```console
sudo apk add --no-cache bash
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/upsun-cli/setup.alpine.sh' \
  | sudo -E bash
```

```console
apk add upsun-cli
```

### Ubuntu/Debian

```console
apt-get update
apt-get install -y apt-transport-https curl
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/upsun-cli/setup.deb.sh' \
  | sudo -E bash
```

```console
apt-get install -y upsun-cli
```

### CentOS/RHEL/Fedora

```console
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/upsun-cli/setup.rpm.sh' \
  | sudo -E bash

yum install -y upsun-cli
```

### Manual installation

For manual installation, you can also [download the latest binaries](https://github.com/upsun/cli/releases/latest).

## Upgrade

Upgrade using the same tool:

### HomeBrew

```console
brew update && brew upgrade platformsh/tap/upsun-cli
```

### Scoop

```console
scoop update upsun
```

### Bash installer

```console
curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | bash
```

### Alpine

```console
apk upgrade upsun-cli
```

### Ubuntu/Debian

```console
apt-get upgrade upsun-cli
```

### CentOS/RHEL/Fedora

```console
yum upgrade -y upsun-cli
```

## Platform.sh compatibility

For backwards compatibility with Platform.sh, a `platform` binary is also available:

```console
brew install platformsh/tap/platformsh-cli
```

Or with the bash installer:

```console
curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | VENDOR=platformsh bash
```

## Building

Build a single binary:

```console
make single
```

Build a snapshot:

```console
make snapshot
```

Build a snapshot for a vendor:

```console
# Download the config file at internal/config/embedded-config.yaml
make vendor-snapshot VENDOR_NAME='Vendor Name' VENDOR_BINARY='vendorcli'
```

Create a release:

```console
# First, create a new tag
git tag -m 'Release v5.0.0' 'v5.0.0'

# Create a release (requires GITHUB_TOKEN)
make release
```

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt).

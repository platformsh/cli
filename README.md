# Upsun CLI

The **Upsun CLI** is the official command-line interface for [Upsun](https://upsun.com) (formerly Platform.sh).

This repository hosts the source code and releases of the CLI.

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Install

To install the CLI, use either [Homebrew](https://brew.sh/) (on Linux, macOS, or the Windows Subsystem for Linux) or [Scoop](https://scoop.sh/) (on Windows):

### HomeBrew

```console
brew install platformsh/tap/platformsh-cli
```

> If you have issues with missing libraries on a Mac, see how to [troubleshoot CLI installation on M1/M2 Macs](https://community.platform.sh/t/troubleshoot-cli-installation-on-m1-macs/1202).

### Scoop

```console
scoop bucket add platformsh https://github.com/platformsh/homebrew-tap.git
scoop install platform
```

### Bash installer

Use the bash installer for an automated installation, using the most preferable way for your system.

```console
curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | bash
```

The installer is configurable using the following environment variables:

* `INSTALL_LOG` - the install log file
* `INSTALL_METHOD` - force a specific installation method, possible values are `brew` and `raw`
* `INSTALL_DIR` - the installation directory for the `raw` installation method, for example you can use `INSTALL_DIR=$HOME/.local/bin` for a single user installation
* `VERSION` - the version of the CLI to install, if you need a version other than the latest one


### Nix profile

Refer to the [Nix
documentation if you are not on NixOS](https://nix.dev/manual/nix/2.24/installation/installing-binary.html).

```console
nix profile install nixpkgs#platformsh
nix profile install nixpkgs#upsun
```

#### Installation configuration examples

<details>
    <summary>Force the CLI to be installed using the raw method</summary>

    curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | INSTALL_METHOD=raw bash
</details>

<details>
    <summary>Install a specific version</summary>

    curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | VERSION=4.0.1 bash
</details>

<details>
    <summary>Install the CLI in a user owned directory</summary>

    curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | INSTALL_METHOD=raw INSTALL_DIR=$HOME/.local/bin bash
</details>

### Alpine

```console
sudo apk add --no-cache bash
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/cli/setup.alpine.sh' \
  | sudo -E bash
```

<details>
    <summary>Manual setup</summary>

    apk add --no-cache curl
    curl -1sLf 'https://dl.cloudsmith.io/public/platformsh/cli/rsa.4F1C2AC5106DA770.key' > /etc/apk/keys/cli@platformsh-4F1C2AC5106DA770.rsa.pub
    curl -1sLf "https://dl.cloudsmith.io/public/platformsh/cli/config.alpine.txt" >> /etc/apk/repositories
    apk update

</details>

```console
# Install the CLI
apk add platformsh-cli
```

### Ubuntu/Debian

```console
apt-get update
apt-get install -y apt-transport-https curl
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/cli/setup.deb.sh' \
  | sudo -E bash
```

<details>
    <summary>Manual setup</summary>

    apt-get update

    # Only needed for Debian
    apt-get install -y debian-keyring debian-archive-keyring

    apt-get install -y apt-transport-https curl gnupg
    curl -1sLf 'https://dl.cloudsmith.io/public/platformsh/cli/gpg.6ED8A90E60ABD941.key' |  gpg --dearmor >> /usr/share/keyrings/platformsh-cli-archive-keyring.gpg
    # If you use an Ubuntu derivative distro, such as Linux Mint, you may need to use UBUNTU_CODENAME instead of VERSION_CODENAME below.
    curl -1sLf "https://dl.cloudsmith.io/public/platformsh/cli/config.deb.txt?distro=$(. /etc/os-release && echo "$ID")&codename=$(. /etc/os-release && echo "$VERSION_CODENAME")" > /etc/apt/sources.list.d/platformsh-cli.list
    apt-get update

</details>

```console
# Install the CLI
apt-get install -y platformsh-cli
```

### CentOS/RHEL/Fedora

```console
curl -1sLf \
  'https://dl.cloudsmith.io/public/platformsh/cli/setup.rpm.sh' \
  | sudo -E bash

# Install the CLI
yum install -y platformsh-cli
```

### Manual installation

For manual installation, you can also [download the latest binaries](https://github.com/platformsh/cli/releases/latest).

## Upgrade

Upgrade using the same tool:

### HomeBrew

```console
brew update && brew upgrade platformsh/tap/platformsh-cli
```

### Scoop

```console
scoop update platform
```

### Bash installer

```console
curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | bash
```

### Alpine

```console
apk add -l platformsh-cli
```

### Ubuntu/Debian

```console
apt-get upgrade platformsh-cli
```

### CentOS/RHEL/Fedora

```console
yum upgrade -y platformsh-cli
```

## Under the hood

## Building binaries, snapshots and releases

Build a single binary

```console
make single
```

Build a snapshot

```console
make snapshot
```

Build a snapshot for a vendor

```console
# Download the config file at internal/config/embedded-config.yaml
make vendor-snapshot VENDOR_NAME='Upsun staging' VENDOR_BINARY='upsunstg'
```

Create a release

```console
# First, create a new tag, if not already on a Git tag
git tag -f -m 'Release v5.0.0' '5.0.0'

# Create a release
# Expose a GITHUB_TOKEN to use for the release
make release

# Create a release for a vendor with the default .gorelease.vendor.yaml.tpl
# Download the config file at internal/config/embedded-config.yaml
# Optionally, download the .goreleaser.vendor.yaml file to use a custom one
make vendor-release VENDOR_NAME='Upsun staging' VENDOR_BINARY='upsunstg'
```

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt).

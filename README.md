# Upsun CLI

The **Upsun CLI** is the official command-line interface for [Upsun](https://upsun.com).

This repository hosts the source code and releases of the CLI.

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Install

To install the CLI, use either [Homebrew](https://brew.sh/) (on Linux, macOS, or the Windows Subsystem for Linux) or [Scoop](https://scoop.sh/) (on Windows):

### HomeBrew

```console
brew install upsun/tap/upsun-cli
```

### Scoop

```console
scoop bucket add upsun https://github.com/upsun/homebrew-tap.git
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
# Add the signing key and repository
sudo mkdir -p /etc/apk/keys
sudo curl -fsSL -o /etc/apk/keys/repositories-upsun-com.rsa.pub https://repositories.upsun.com/alpine/repositories-upsun-com.rsa.pub
echo "https://repositories.upsun.com/alpine" | sudo tee -a /etc/apk/repositories

# Install the CLI
sudo apk add upsun-cli
```

### Ubuntu/Debian

```console
# Add the signing key and repository
sudo mkdir -p /etc/apt/keyrings
sudo curl -fsSL https://repositories.upsun.com/gpg.key -o /etc/apt/keyrings/upsun.asc
echo "deb [signed-by=/etc/apt/keyrings/upsun.asc] https://repositories.upsun.com/debian stable main" | sudo tee /etc/apt/sources.list.d/upsun.list

# Install the CLI
sudo apt-get update
sudo apt-get install -y upsun-cli
```

### CentOS/RHEL/Fedora

```console
# Add the repository
sudo tee /etc/yum.repos.d/upsun.repo << 'EOF'
[upsun]
name=Upsun CLI
baseurl=https://repositories.upsun.com/fedora/$releasever/$basearch
enabled=1
gpgcheck=1
gpgkey=https://repositories.upsun.com/gpg.key
EOF

# Install the CLI (use yum on older systems)
sudo dnf install -y upsun-cli
```

### Manual installation

For manual installation, you can also [download the latest binaries](https://github.com/upsun/cli/releases/latest).

## Upgrade

Upgrade using the same tool:

### HomeBrew

```console
brew update && brew upgrade upsun/tap/upsun-cli
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
sudo apk update && sudo apk upgrade upsun-cli
```

### Ubuntu/Debian

```console
sudo apt-get update && sudo apt-get upgrade upsun-cli
```

### CentOS/RHEL/Fedora

```console
sudo dnf upgrade -y upsun-cli
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

## Creating a Release

Releases are automated via GitHub Actions. To create a new release:

1. Create and push a new tag:
   ```console
   git tag -m 'Release v5.0.0' 'v5.0.0'
   git push origin v5.0.0
   ```

2. The [Release workflow](.github/workflows/release.yml) will automatically:
   - Build binaries for all platforms
   - Sign packages (APK, DEB, RPM)
   - Create a GitHub release with all artifacts
   - Update package repositories at repositories.upsun.com

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt).

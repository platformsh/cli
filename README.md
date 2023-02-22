# Platform.sh CLI

The **Platform.sh CLI** is the official command-line interface for [Platform.sh](https://platform.sh). Use this tool to interact with your [Platform.sh](https://platform.sh) projects, and to build them locally for development purposes.

This repository hosts the source code and releases of the new CLI.

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

### Manual installation

For manual installation, you can also [download the latest binaries](https://github.com/platformsh/cli/releases/latest).

## Upgrade

Upgrade using the same tool:

### HomeBrew

```console
brew upgrade platformsh-cli
```

### Scoop

```console
scoop update platform
```

## Under the hood

The New Platform.sh CLI is built with backwards compatibility in mind. This is why we've embedded PHP, so that all Legacy PHP CLI commands can be executed in the exact same way, making sure that nothing breaks when you switch to it.

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt).


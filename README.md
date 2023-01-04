# Platform.sh CLI

This repository hosts the releases of the new Platform.sh CLI

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Install

To install the CLI, use either [Homebrew](https://brew.sh/) (on Linux, macOS, or the Windows Subsystem for Linux) or [Scoop](https://scoop.sh/) (on Windows):

### HomeBrew

```console
brew install platformsh/tap/platformsh-cli
```

### Scoop

```console
scoop bucket add platformsh https://github.com/platformsh/homebrew-tap.git
scoop install platform
```

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


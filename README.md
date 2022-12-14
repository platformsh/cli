# Platform.sh CLI

This repository hosts the releases of the new Platform.sh CLI

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Installation

We support installing the CLI either through the various options found below

### HomeBrew, LinuxBrew and Scoop

Currently we support [HomeBrew](https://docs.brew.sh/) for macOS, [LinuxBrew](https://docs.brew.sh/Homebrew-on-Linux) for Linux and Scoop for [Windows](https://scoop.sh/).

This is the preferred way of installation.

In order to install the package, you just need to run the following commands, after having installed the package manager for your OS:

#### macOS or Linux

```console
brew install platformsh/tap/platformsh-cli
```

#### Windows

```console
scoop bucket add platformsh https://github.com/platformsh/homebrew-tap.git
scoop install platform
```

### Distribution specific Linux packages

We are distributing the CLI using standard distribution channels for different Linux distributions, like APK, Deb, and RPM.

You can find all available packages in the [latest release](https://github.com/platformsh/cli/releases/latest).

### Binary installation

The binaries are included in each release, so installing the binary should be easy for all macOS, Linux and Windows systems.

_For macOS systems, it's better to install the CLI using HomeBrew, as you might end up with signing issues otherwise. Also, you need to make sure that the following dependencies exist in your system: `libssl` and `libonig`. You can install them with the following command._

```console
brew install oniguruma openssl@1.1
```

You can find all available binaries in the [latest release](https://github.com/platformsh/cli/releases/latest).

## Under the hood

The New Platform.sh CLI is built with backwards compatibility in mind. This is why we've embedded PHP, so that all Legacy PHP CLI commands can be executed in the exact same way, making sure that nothing breaks when you switch to it.

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt).


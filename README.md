# Platform.sh CLI

The **Platform.sh CLI** is the official command-line interface for [Platform.sh](https://platform.sh). Use this tool to interact with your [Platform.sh](https://platform.sh) projects, and to build them locally for development purposes.

This repository hosts the source code and releases of the new CLI.

> This product includes PHP software, freely available from [the PHP website](https://www.php.net/software)

## Install

To install the CLI, use either [Homebrew](https://brew.sh/) (on Linux, macOS, or the Windows Subsystem for Linux) or [Scoop](https://scoop.sh/) (on Windows):

### Homebrew

```console
brew install platformsh/tap/platformsh-cli
```

After installing or updating platformsh-cli via Homebrew, make sure you have the right `platform` in your path by issuing
```bash
hash -r && ls -l $(which platform)
```

You should see a path that involves homebrew, `/opt/homebrew/bin/platform` on an Apple Silicon Mac (M1/M2), `/usr/local/bin/platform` on an Intel Mac, `/home/linuxbrew/.linuxbrew/bin/platform` on Linux. If you don't see that, then delete the target that you are seeing with `sudo rm "$(which platform)" && hash -r` and then you should see the correct `platform` tool in `which platform`.

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


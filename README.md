# Platform.sh CLI - Go wrapper

This project intends to distribute the Platform.sh CLI as a single binary which embeds PHP.

## Installation

Currently we support installing the binary through [HomeBrew](https://docs.brew.sh/) for Mac and [LinuxBrew](https://docs.brew.sh/Homebrew-on-Linux) for Linux.

In order to install the package, you just need to run:

```console
brew install platformsh/tap/platformsh-cli
```

The HomeBrew packages are distributed with our own tap at https://github.com/platformsh/homebrew-tap

Alternatively, you can install the binary directly from the GitHub release.

### Future installers

Below are installers that do not currently exist, but will be created before we go GA.

* [Chocolatey](https://chocolatey.org/) for Windows
* [nfpm](https://github.com/goreleaser/nfpm) for `.dep`, `rpm` and `.apk` packages for all major Linux distributions

### Binary installation

Binaries are included in each release, so installing the binary should be easy for both Unix and Windows systems. For Unix systems, you need to make sure that the following dependencies exist in your system: `libssl` and `libonig`.

You can install these dependencies on Ubuntu with:

```console
apt install libonig5 libssl3
```

You can install these dependencies on MacOS with:

```console
brew install oniguruma openssl@1.1
```

## Building

The Go part is built using [Go Releaser](https://goreleaser.com/), while the PHP binary that is embedded relies on the PHP source and Brew dependencies.

The needed dependencies needed for the HomeBrew build are installed automatically in [`build-php-brew.sh`](./build-php-brew.sh) script.

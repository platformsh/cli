# Upsun CLI

> [!IMPORTANT]
> **This repository has been migrated to [upsun/cli](https://github.com/upsun/cli).**
>
> Please use the new repository for installations, updates, and issue reporting. Existing installations will continue to work, but we recommend migrating to the new repository.

The **Upsun CLI** is the official command-line interface for [Upsun](https://upsun.com) (formerly Platform.sh).

## Install

```console
curl -fsSL https://raw.githubusercontent.com/upsun/cli/main/installer.sh | bash
```

For other installation methods (Homebrew, Scoop, Alpine, Debian/Ubuntu, RHEL/Fedora, Nix), upgrade instructions, and documentation, see [upsun/cli](https://github.com/upsun/cli).

### The `platform` binary

The `platform` binary remains available for users who depend on it. It is functionally close to `upsun` but the binary name differs and a few behaviors are subtly different. To install it, use this repository's installer:

```console
curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | bash
```

## Licenses

This binary redistributes PHP in a binary form, which comes with the [PHP License](https://www.php.net/license/3_01.txt). PHP is freely available from [the PHP website](https://www.php.net/software).

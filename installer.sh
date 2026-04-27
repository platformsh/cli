#!/bin/sh
# Upsun CLI installer

set -eu

# Location of install log
: "${INSTALL_LOG:=/tmp/upsun-install-$(date '+%Y%m%d-%H%M%S').log}"

# Define this to force install method
: "${INSTALL_METHOD:=}"

: "${URL:=https://github.com/upsun/cli/releases/download}"

# Force Upsun CLI installation in this directory instead of system directory
: "${INSTALL_DIR:=}"

# Upsun CLI version to install
: "${VERSION:=}"

# macOS specifics
: "${BREW_TAP:=upsun/tap}"
: "${BREW_FORMULA:=upsun/tap/platformsh-cli}"

# GitHub token check
: "${GITHUB_TOKEN:=}"

# CI specifics
: "${CI:=}"
: "${BUILD_NUMBER:=}"
: "${RUN_ID:=}"

# The vendor to install. Defaults to platformsh for backward compatibility:
# users who pipe this script from platformsh/cli/installer.sh have historically
# received the `platform` binary unless they passed VENDOR=upsun.
: "${VENDOR:=platformsh}"

# global variables
binary="platform"
vendor_name="Upsun (formerly Platform.sh)"
cmd_shasum=""
cmd_sudo=""
dir_bin="/usr/bin"
footer_notes=""
has_sudo=""
kernel=""
machine=""
tag=""
version=""
package="platformsh-cli"
docs_url="https://docs.upsun.com"
support_url="https://upsun.com/contact"

if [ "$VENDOR" = "upsun" ]; then
    BREW_FORMULA="upsun/tap/upsun-cli"
    binary="upsun"
    vendor_name="Upsun"
    package="upsun-cli"
fi

# create a log file where every output will be pipe to
pipe=/tmp/$binary-install-$$.tmp
mkfifo $pipe
tee < $pipe $INSTALL_LOG &
exec 1>&-
exec 1>$pipe 2>&1
trap 'rm -f $pipe' EXIT

output() {
    style_start=""
    style_end=""
    if [ "${2:-}" != "" ]; then
    case $2 in
        "success")
            style_start="\033[0;32m"
            style_end="\033[0m"
            ;;
        "error")
            style_start="\033[31;31m"
            style_end="\033[0m"
            ;;
        "info"|"warning")
            style_start="\033[33m"
            style_end="\033[39m"
            ;;
        "heading")
            style_start="\033[1;33m"
            style_end="\033[22;39m"
            ;;
        "comment")
            style_start="\033[2m"
            style_end="\033[22;39m"
            ;;
    esac
    fi

    printf "%b%b%b\n" "${style_start}" "${1}" "${style_end}"
}

create_table_line() {
    local text="${1:-}"
    local width="${2:-}"
    local text_length
    local i
    local padding
    local right_padding

    if [ -z "$width" ]; then
        # Calculate width based on text length, minimum 50
        text_length=${#text}
        if [ $text_length -lt 50 ]; then
            width=50
        else
            width=$((text_length + 4))
        fi
    fi

    if [ "$text" = "border" ] || [ -z "$text" ]; then
        # Border line
        printf "+"
        i=0
        while [ $i -lt $((width - 2)) ]; do
            printf "-"
            i=$((i + 1))
        done
        printf "+"
    elif [ "$text" = "empty" ]; then
        # Empty line
        printf "|"
        i=0
        while [ $i -lt $((width - 2)) ]; do
            printf " "
            i=$((i + 1))
        done
        printf "|"
    else
        # Text line
        text_length=${#text}
        padding=$(((width - 2 - text_length) / 2))
        right_padding=$((width - 2 - text_length - padding))
        printf "|"
        i=0
        while [ $i -lt $padding ]; do
            printf " "
            i=$((i + 1))
        done
        printf "%s" "$text"
        i=0
        while [ $i -lt $right_padding ]; do
            printf " "
            i=$((i + 1))
        done
        printf "|"
    fi
    printf "\n"
}

create_table() {
    local text="$1"
    local min_width="${2:-50}"
    local text_length=${#text}
    local table_width
    if [ $text_length -lt $min_width ]; then
        table_width=$min_width
    else
        table_width=$((text_length + 4))
    fi

    create_table_line "border" "$table_width"
    create_table_line "empty" "$table_width"
    create_table_line "$text" "$table_width"
    create_table_line "empty" "$table_width"
    create_table_line "border" "$table_width"
}

exit_with_error() {
    title="Installation failed"

    output "$(create_table "$title")" "error"

    output "\nGet help with your $vendor_name CLI installation:" "heading"
    output "  Inspect the logs: ${INSTALL_LOG}"
    output "  Read the docs: $docs_url/administration/cli.html"
    output "  Get help: $support_url"

    exit 1
}

intro() {
    title="$vendor_name CLI Installer"

    output "$(create_table "$title")" "heading"

    output "\nChecking environment" "heading"
}

outro() {
    title="$vendor_name CLI has been installed successfully."

    output ""
    output "$(create_table "$title")" "success"

    output "\nWhat's next?" "heading"

    output "  To use the CLI, run: $binary" "output"

    output "\nUseful links:" "heading"
    output "  CLI introduction: $docs_url/get-started/introduction.html#cli"

    if [ ! -z "$footer_notes" ]; then
        output "\nWarning during installation:" "heading"
        output "$footer_notes" "warning"
    fi

    output "\nThank you for using $vendor_name!"
}

add_footer_note() {
    for var in "$@"; do
        if [ ! -z "$footer_notes" ]; then
            footer_notes="${footer_notes}\n${var}"
        else
            footer_notes="${footer_notes}${var}"
        fi
    done
}

indent() {
    OLDIFS=$IFS
    IFS='
'
    while read -r data; do
        line=$(printf "   | %s" "$data" | sed 's/\r/\r   | /g' | sed 's/\x1B\[[0-9;]*[A-Za-z]//g')
        output "$line" "comment"
    done
    IFS=$OLDIFS
}

# Check that cURL is installed
check_curl() {
    if command -v curl >/dev/null 2>&1; then
        output "  [*] cURL is installed" "success"
    else
        output "  [ ] ERROR: cURL is required for installation" "error"
        exit_with_error
    fi
}

setup_github_token() {
    if gh auth status >/dev/null 2>&1; then
        GITHUB_TOKEN="$(gh auth token)"
        if ! github_curl https://api.github.com/repos/upsun/cli/releases/latest >/dev/null 2>&1; then
            GITHUB_TOKEN=""
        else
            output "  [*] Using GitHub auth from the gh CLI" "success"
        fi
    elif [ ! -z "${GITHUB_TOKEN}" ]; then
        if ! github_curl https://api.github.com/repos/upsun/cli/releases/latest >/dev/null 2>&1; then
            GITHUB_TOKEN=""
        else
            output "  [*] Using GitHub auth from the GITHUB_TOKEN env variable" "success"
        fi
    fi
}

# Check that Gzip is installed
check_gzip() {
    if command -v gzip >/dev/null 2>&1; then
        output "  [*] Gzip is installed" "success"
    else
        output "  [ ] ERROR: Gzip is required for installation" "error"
        exit_with_error
    fi
}

# Check that ca-certificates is installed (Alpine)
check_ca_certificates() {
    if apk info -e ca-certificates >/dev/null 2>&1; then
        output "  [*] ca-certificates is installed" "success"
    else
        output "  [ ] ERROR: ca-certificates is required for installation" "error"
        exit_with_error
    fi
}

check_version() {
    if [ -z "${VERSION}" ]; then
        tag=$(github_curl https://api.github.com/repos/upsun/cli/releases/latest 2>/dev/null | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
    else
        tag=${VERSION}
    fi

    # Ensure the tag has the v prefix (for GitHub release URLs).
    case "$tag" in
        v*) ;;
        *) tag="v${tag}" ;;
    esac

    # The version without the v prefix (for asset filenames).
    version=${tag#v}

    if [ -z "$version" ]; then
        output "  [ ] ERROR: Could not determine CLI version" "error"
        exit_with_error
    fi

    if [ -z "${VERSION}" ]; then
        output "  [*] No version specified, using latest ($version)" "success"
    else
        output "  [*] Version ${version} specified" "success"
    fi
}

# Detect the kernel type
check_kernel() {
    kernel=$(uname -s 2>$pipe || /usr/bin/uname -s)
    case ${kernel} in
        "Linux"|"linux")
            kernel="linux"
            ;;
        "Darwin"|"darwin")
            kernel="darwin"
            dir_bin="/usr/local/bin"
            ;;
        "FreeBSD"|"freebsd")
            kernel="freebsd"
            ;;
        *)
            output "  [ ] Your OS (${kernel}) is currently not supported" "error"
            exit_with_error
            ;;
    esac

    output "  [*] Your kernel (${kernel}) is supported" "success"
}

check_architecture() {
    # Detect architecture
    machine=$(uname -m 2>$pipe || /usr/bin/uname -m)
    case ${machine} in
        aarch64*|armv8*|arm64*)
            machine="arm64"
            ;;
        i[36]86|x86)
            machine="386"
            ;;
        x86_64|amd64)
            machine="amd64"
            ;;
        *)
            output "  [ ] Your architecture (${machine}) is currently not supported" "error"
            exit_with_error
            ;;
    esac

    output "  [*] Your architecture (${machine}) is supported" "success"
}

check_install_method() {
    if [ "linux" = "$kernel" ]; then
        if [ -z "${INSTALL_METHOD}" ]; then
            if is_ci; then
                INSTALL_METHOD="raw"
            elif [ ! -z $VERSION ]; then
                INSTALL_METHOD="raw"
            elif command -v apt-get > /dev/null 2>&1; then
                INSTALL_METHOD="apt"
            elif command -v yum > /dev/null 2>&1; then
                INSTALL_METHOD="yum"
            elif command -v apk > /dev/null 2>&1; then
                INSTALL_METHOD="apk"
            else
                INSTALL_METHOD="raw"
            fi
        fi
    elif [ "darwin" = $kernel ]; then
        if [ -z "${INSTALL_METHOD}" ]; then
            if command -v brew > /dev/null 2>&1
            then
                INSTALL_METHOD="homebrew"
            else
                INSTALL_METHOD="raw"
            fi
        fi
        if [ "${INSTALL_METHOD}" = "raw" ]; then
            machine="all"
        fi
    elif [ "freebsd" = $kernel ]; then
        INSTALL_METHOD="raw"
    fi

    output "  [*] Using ${INSTALL_METHOD} install method" "success"
}

init_sudo() {
    if [ ! -z "${has_sudo}" ]; then
        return
    fi

    has_sudo=false
    # Are we running the installer as root?
    if [ "$(id -u)" = "0" ]; then
        has_sudo=true
        cmd_sudo=''

        return
    fi

    if command -v sudo > /dev/null 2>&1; then
        has_sudo=true
        cmd_sudo='sudo -E'
    fi
}

call_root() {
    init_sudo

    if ! ${has_sudo}; then
        output "  sudo is required to perform this operation" "error"
        exit_with_error
    fi

    local output_file
    output_file=$(mktemp)
    if $cmd_sudo sh -c "$1" >"$output_file" 2>&1; then
        cat "$output_file" | indent
        rm -f "$output_file"
        return 0
    else
        local exit_code=$?
        cat "$output_file" | indent
        rm -f "$output_file"
        return $exit_code
    fi
}

call_try_user() {
    if ! call_user "$1"; then
        output "  command failed, re-trying with sudo" "warning"
        if ! call_root "$1"; then
            output "  ${2:-command failed}" "error"
            exit_with_error
        fi
    fi
}

call_user() {
    sh -c "$1" 2>&1 | indent
}

check_shasum() {
    if command -v sha1sum > /dev/null 2>&1; then
        cmd_shasum="sha1sum"
        output "  [*] sha1sum is installed" "success"
    elif command -v shasum > /dev/null 2>&1; then
        cmd_shasum="shasum -a 1"
        output "  [*] shasum is installed" "success"
    else
        output "  [ ] No sha1sum or shasum available to verify binary" "error"
        exit_with_error
    fi
}

check_directories() {
    if [ ! -z "${INSTALL_DIR}" ]; then
        dir_bin="${INSTALL_DIR}"
        INSTALL_METHOD="raw"
    elif echo $PATH | grep "$HOME/.global/bin" > /dev/null; then
        dir_bin="$HOME/.global/bin"
    elif echo $PATH | grep "$HOME/.local/bin" > /dev/null; then
        dir_bin="$HOME/.local/bin"
    fi

    if ! echo $PATH | grep ${dir_bin} > /dev/null; then
        binary="${dir_bin}/$binary"

        output "  [ ] ${dir_bin} is not in \$PATH.\n" "warning"
        add_footer_note "  ⚠ The directory \"${dir_bin}\" is not in \$PATH"
        if echo $SHELL | grep '/bin/zsh' > /dev/null
        then
            add_footer_note \
                "    Run this command to add the directory to your PATH" \
                "    echo 'export PATH=\"${dir_bin}:\$PATH\"' >> \$HOME/.zshrc"
        elif echo $SHELL | grep '/bin/bash' > /dev/null
        then
            add_footer_note \
                "    Run this command to add the directory to your PATH" \
                "    echo 'export PATH=\"${dir_bin}:\$PATH\"' >> \$HOME/.bashrc"
        else
            add_footer_note \
                "    You can add it to your PATH by adding this line at the end of your shell configuration file" \
                "    export PATH=\"${dir_bin}:\$PATH\""
        fi
    else
        output "  [*] ${dir_bin} is in \$PATH" "success"
    fi

    output "\nTarget directories" "heading"
    output "  Binary will be installed in ${dir_bin}"
}

install_homebrew() {
    output "\nInstalling the $vendor_name brew tap" "heading"
    if ! call_user "brew tap ${BREW_TAP}"; then
        output "  could not add tap to brew" "error"
        exit_with_error
    fi

    output "\nInstalling the $vendor_name CLI formula" "heading"

    # running on Rosetta2?
    arch=''
    if [ "$(uname -m)" = "x86_64" ]; then
        if [ "$(sysctl -in sysctl.proc_translated)" = "1" ]; then
            arch="arch -arm64"
        fi
    fi

    if ! call_user "$arch brew install ${BREW_FORMULA}"; then
        output "  could not install Formula" "error"
        exit_with_error
    fi
}

install_yum() {
    output "\nSetting up the $vendor_name CLI yum repository" "heading"

    repo_file="/etc/yum.repos.d/upsun.repo"
    repo_content="[upsun]
name=Upsun CLI
baseurl=https://repositories.upsun.com/fedora/\\\$releasever/\\\$basearch
enabled=1
gpgcheck=1
gpgkey=https://repositories.upsun.com/gpg.key"

    cmd="printf '${repo_content}' > '${repo_file}'"

    if ! call_root "${cmd}"; then
        output "  could not setup the RPM repository the CLI" "error"
        exit_with_error
    fi

    output "\nInstalling the $vendor_name CLI" "heading"

    if ! call_root "yum install -y $package"; then
        output "  could not install the CLI" "error"
        exit_with_error
    fi
}

install_apt() {
    output "\nSetting up the $vendor_name CLI apt repository" "heading"

    # Install the GPG key
    if ! call_root "mkdir -p /etc/apt/keyrings"; then
        output "  could not create keyrings directory" "error"
        exit_with_error
    fi

    if ! call_root "curl -fsSL https://repositories.upsun.com/gpg.key -o /etc/apt/keyrings/upsun.asc"; then
        output "  could not download GPG key" "error"
        exit_with_error
    fi

    list_file="/etc/apt/sources.list.d/upsun.list"
    list_content="deb [signed-by=/etc/apt/keyrings/upsun.asc] https://repositories.upsun.com/debian stable main"

    cmd="printf '${list_content}' > '${list_file}'"

    if ! call_root "${cmd}"; then
        output "  could not setup the APT repository the CLI" "error"
        exit_with_error
    fi

    if ! call_root "apt-get update"; then
        output "  could not update apt cache" "error"
        exit_with_error
    fi

    output "\nInstalling the $vendor_name CLI" "heading"

    if ! call_root "apt-get install -y $package"; then
        output "  could not install the CLI" "error"
        exit_with_error
    fi
}

install_apk() {
    output "\nSetting up the $vendor_name CLI apk repository" "heading"

    # Repository configuration
    local repo_url="https://repositories.upsun.com"
    local rsa_key_name="repositories-upsun-com.rsa.pub"

    # Download and install the repository signing key
    output "  Installing repository signing key..."
    if ! call_root "mkdir -p /etc/apk/keys && curl -fsSL -o /etc/apk/keys/${rsa_key_name} ${repo_url}/alpine/${rsa_key_name}"; then
        output "  could not install the signing key" "error"
        exit_with_error
    fi

    # Add the repository (Alpine automatically appends /$arch to the URL)
    output "  Adding repository..."
    if ! call_root "grep -qF '${repo_url}/alpine' /etc/apk/repositories 2>/dev/null || echo '${repo_url}/alpine' >> /etc/apk/repositories"; then
        output "  could not add the repository" "error"
        exit_with_error
    fi

    output "\nInstalling the $vendor_name CLI" "heading"

    if ! call_root "apk add $package --update-cache"; then
        output "  could not install the CLI" "error"
        exit_with_error
    fi
}

github_curl() {
    if [ -z "${GITHUB_TOKEN}" ]; then
        curl -fsSL -H "Accept: application/vnd.github+json" "$1"
        return $?
    else
        curl -fsSL -H "Accept: application/vnd.github+json" -H "Authorization: Bearer ${GITHUB_TOKEN}" "$1"
        return $?
    fi
}

is_ci() {
    if [ ! -z "${CI}" ]; then # GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
        return 0
    elif [ ! -z "${BUILD_NUMBER}" ]; then # Jenkins, TeamCity
        return 0
    elif [ ! -z "${RUN_ID}" ]; then # TaskCluster, dsari
        return 0
    else
        return 1
    fi
}

install_raw() {
    # Start downloading the right version
    output "\nDownloading the $vendor_name CLI" "heading"

    url="${URL}/${tag}/${binary}_${version}_${kernel}_${machine}.tar.gz"
    output "  Downloading ${url}";
    tmp_dir=$(mktemp -d)
    tmp_name="$binary-"$(date +"%s")
    if ! github_curl "$url" > "${tmp_dir}/${tmp_name}.tgz"; then
        output "  the download failed" "error"
        exit_with_error
    fi

    output "  Uncompressing archive"
    tar -C "${tmp_dir}" -xzf "${tmp_dir}/${tmp_name}.tgz"

    output "  Making the binary executable"
    chmod 0755 "${tmp_dir}/$binary"

    if [ ! -d "$dir_bin" ]; then
        output "  Creating ${dir_bin} directory"
        call_try_user "mkdir -p ${dir_bin}" "Failed to create the ${dir_bin} directory"
    fi

    output "  Installing the binary under ${dir_bin}"
    call_try_user "mv '${tmp_dir}/$binary' '${dir_bin}/${binary}'" "Failed to move the binary ${binary}"
}

install() {
    if [ "homebrew" = "${INSTALL_METHOD}" ]; then
       install_homebrew
    elif [ "yum" = "${INSTALL_METHOD}" ]; then
       install_yum
    elif [ "apt" = "${INSTALL_METHOD}" ]; then
       install_apt
    elif [ "apk" = "${INSTALL_METHOD}" ]; then
       install_apk
    elif [ "raw" = "${INSTALL_METHOD}" ]; then
       install_raw
    fi
}

intro
check_kernel
check_architecture
check_install_method
if [ "raw" = "${INSTALL_METHOD}" ]; then
    check_curl
    if [ -z "${VERSION}" ]; then
        setup_github_token
    fi
    check_version
    check_gzip
    check_shasum
    check_directories
elif [ "apt" = "${INSTALL_METHOD}" ] || [ "yum" = "${INSTALL_METHOD}" ] || [ "apk" = "${INSTALL_METHOD}" ]; then
    check_curl
    if [ "apk" = "${INSTALL_METHOD}" ]; then
        check_ca_certificates
    fi
fi
install
outro

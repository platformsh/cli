#!/bin/bash
set -e

# Parse arguments
PACKAGE_MANAGER="${1:-all}"  # debian, rpm, alpine, or all (default)

if [[ ! "$PACKAGE_MANAGER" =~ ^(debian|rpm|alpine|all)$ ]]; then
    echo "Usage: $0 [debian|rpm|alpine|all]"
    echo "  Default: all"
    exit 1
fi

echo "Package manager filter: $PACKAGE_MANAGER"

# Required environment variables:
# - VERSION: Version of the CLI being released (auto-detected from dist/ if not set)
# - AWS credentials: Configured via AWS CLI or environment
# - GPG_PRIVATE_KEY_FILE: Path to GPG private key file for signing Debian/RPM packages
#   OR GPG_PRIVATE_KEY: Base64-encoded GPG private key (legacy)
# - RSA_PRIVATE_KEY_FILE: Path to RSA private key file for signing Alpine packages
#   OR RSA_PRIVATE_KEY: RSA private key content (legacy)
# - RSA_PUBLIC_KEY_FILE: Path to RSA public key file for Alpine users
#   OR RSA_PUBLIC_KEY: RSA public key content (legacy)

# Auto-detect VERSION from dist/ if not set
if [ -z "$VERSION" ]; then
    # Use glob instead of ls to avoid parsing ls output
    for deb_file in dist/*-cli_*_linux_*.deb; do
        if [ -f "$deb_file" ]; then
            VERSION=$(basename "$deb_file" | sed -E 's/.*-cli_(.+)_linux_.*/\1/')
            break
        fi
    done
    if [ -n "$VERSION" ]; then
        echo "Auto-detected VERSION: $VERSION"
    fi
fi

# Validate VERSION format (should be semver-like)
if [ -n "$VERSION" ] && ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+'; then
    echo "Warning: VERSION '$VERSION' does not look like a valid semver format"
fi

# Check required environment variables
MISSING_VARS=()

if [ -z "$VERSION" ]; then
    MISSING_VARS+=("VERSION")
fi

if [ -z "$GPG_PRIVATE_KEY_FILE" ] && [ -z "$GPG_PRIVATE_KEY" ]; then
    MISSING_VARS+=("GPG_PRIVATE_KEY_FILE or GPG_PRIVATE_KEY")
fi

if [ -z "$RSA_PRIVATE_KEY_FILE" ] && [ -z "$RSA_PRIVATE_KEY" ]; then
    MISSING_VARS+=("RSA_PRIVATE_KEY_FILE or RSA_PRIVATE_KEY")
fi

if [ -z "$RSA_PUBLIC_KEY_FILE" ] && [ -z "$RSA_PUBLIC_KEY" ]; then
    MISSING_VARS+=("RSA_PUBLIC_KEY_FILE or RSA_PUBLIC_KEY")
fi

if [ ${#MISSING_VARS[@]} -ne 0 ]; then
    echo "Error: The following required environment variables are not set:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    exit 1
fi

# AWS bucket for package repositories
export AWS_BUCKET="${AWS_BUCKET:-public-repogen-repositories-upsun-com}"

# Set AWS region
export AWS_DEFAULT_REGION="${AWS_REGION:-eu-west-1}"

# Create temporary directories for repository work
WORK_DIR=$(mktemp -d)

# Use a temporary GNUPGHOME to avoid polluting the user's keyring
export GNUPGHOME=$(mktemp -d)
trap "rm -rf \"$WORK_DIR\" \"$GNUPGHOME\"" EXIT

# Handle GPG key - from file or environment variable
if [ -n "$GPG_PRIVATE_KEY_FILE" ]; then
    GPG_KEY_FILE="$GPG_PRIVATE_KEY_FILE"
else
    GPG_KEY_FILE="${WORK_DIR}/gpg-private-key.asc"
    echo "$GPG_PRIVATE_KEY" | base64 -d > "$GPG_KEY_FILE"
fi

# Import GPG key
gpg --import "$GPG_KEY_FILE"

# Handle RSA private key - from file or environment variable
if [ -n "$RSA_PRIVATE_KEY_FILE" ]; then
    RSA_KEY_FILE="$RSA_PRIVATE_KEY_FILE"
else
    RSA_KEY_FILE="${WORK_DIR}/rsa-private-key.pem"
    echo "$RSA_PRIVATE_KEY" > "$RSA_KEY_FILE"
fi

# Handle RSA public key - from file or environment variable
if [ -n "$RSA_PUBLIC_KEY_FILE" ]; then
    RSA_PUB_KEY_FILE="$RSA_PUBLIC_KEY_FILE"
else
    RSA_PUB_KEY_FILE="${WORK_DIR}/repositories-upsun-com.rsa.pub"
    echo "$RSA_PUBLIC_KEY" > "$RSA_PUB_KEY_FILE"
fi

# Packages are expected to be pre-built by goreleaser in the dist/ directory
# goreleaser builds signed .deb, .rpm, and .apk packages via nfpm
# Check that packages exist
if [ ! -d "dist" ]; then
    echo "Error: dist/ directory not found. Run 'goreleaser release' or 'make snapshot' first."
    exit 1
fi

# Create shared staging directories for all packages
# Both platformsh and upsun packages go in the same directories so they're
# processed together and appear in the same repository indexes
PACKAGES_DEB_DIR="${WORK_DIR}/packages-deb"
PACKAGES_RPM_DIR="${WORK_DIR}/packages-rpm"
PACKAGES_APK_DIR="${WORK_DIR}/packages-apk"
REPO_DIR="${WORK_DIR}/repo"
mkdir -p "$PACKAGES_DEB_DIR" "$PACKAGES_RPM_DIR" "$PACKAGES_APK_DIR" "$REPO_DIR"

echo "=== Copying packages to staging directories ==="

# Copy Platform.sh packages
cp dist/platformsh-cli_${VERSION}_linux_amd64.deb "$PACKAGES_DEB_DIR/"
cp dist/platformsh-cli_${VERSION}_linux_arm64.deb "$PACKAGES_DEB_DIR/"
cp dist/platformsh-cli_${VERSION}_linux_amd64.apk "$PACKAGES_APK_DIR/"
cp dist/platformsh-cli_${VERSION}_linux_arm64.apk "$PACKAGES_APK_DIR/"
cp dist/platformsh-cli_${VERSION}_linux_amd64.rpm "$PACKAGES_RPM_DIR/"
cp dist/platformsh-cli_${VERSION}_linux_arm64.rpm "$PACKAGES_RPM_DIR/"

# Copy Upsun packages
cp dist/upsun-cli_${VERSION}_linux_amd64.deb "$PACKAGES_DEB_DIR/"
cp dist/upsun-cli_${VERSION}_linux_arm64.deb "$PACKAGES_DEB_DIR/"
cp dist/upsun-cli_${VERSION}_linux_amd64.apk "$PACKAGES_APK_DIR/"
cp dist/upsun-cli_${VERSION}_linux_arm64.apk "$PACKAGES_APK_DIR/"
cp dist/upsun-cli_${VERSION}_linux_amd64.rpm "$PACKAGES_RPM_DIR/"
cp dist/upsun-cli_${VERSION}_linux_arm64.rpm "$PACKAGES_RPM_DIR/"

echo "Packages staged:"
ls -la "$PACKAGES_DEB_DIR" "$PACKAGES_RPM_DIR" "$PACKAGES_APK_DIR" 2>/dev/null

# Check that at least some packages were found
pkg_count=$(find "$PACKAGES_DEB_DIR" "$PACKAGES_RPM_DIR" "$PACKAGES_APK_DIR" -type f 2>/dev/null | wc -l)
if [ "$pkg_count" -eq 0 ]; then
    echo "Error: No packages found in staging directories. Check that goreleaser produced the expected packages."
    exit 1
fi
echo "Found $pkg_count package(s) to process."

# --- Debian Repository ---
if [[ "$PACKAGE_MANAGER" == "all" || "$PACKAGE_MANAGER" == "debian" ]]; then
    echo "=== Processing Debian packages ==="
    DEB_REPO_DIR="${REPO_DIR}/debian"
    mkdir -p "$DEB_REPO_DIR"

    # Sync Debian metadata from S3 (if exists)
    aws s3 sync "s3://${AWS_BUCKET}/debian/dists" "${DEB_REPO_DIR}/dists" --delete 2>/dev/null || true

    # Run repogen for Debian packages
    repogen generate \
        --input-dir "$PACKAGES_DEB_DIR" \
        --output-dir "$DEB_REPO_DIR" \
        --incremental \
        --arch amd64,arm64 \
        --codename stable \
        --origin "Upsun" \
        --label "Upsun CLI" \
        --gpg-key "$GPG_KEY_FILE"

    # Sync Debian back to S3
    aws s3 sync "${DEB_REPO_DIR}" "s3://${AWS_BUCKET}/debian"
    echo "=== Done processing Debian packages ==="
fi

# --- RPM Repository ---
# Generates repos for multiple Fedora versions to support $releasever in yum/dnf configs
if [[ "$PACKAGE_MANAGER" == "all" || "$PACKAGE_MANAGER" == "rpm" ]]; then
    echo "=== Processing RPM packages ==="

    # Fedora versions to support (configurable via environment variable)
    FEDORA_VERSIONS="${FEDORA_VERSIONS:-40 41 42 43}"

    for FEDORA_VER in $FEDORA_VERSIONS; do
        echo "--- Processing Fedora $FEDORA_VER ---"

        # Use a separate output directory per version to avoid repogen's ParseExistingMetadata
        # scanning all version directories and detecting conflicts
        RPM_REPO_DIR="${REPO_DIR}/fedora-${FEDORA_VER}"
        mkdir -p "$RPM_REPO_DIR"

        # Sync only repodata metadata from S3 (not packages)
        for arch in x86_64 aarch64; do
            mkdir -p "${RPM_REPO_DIR}/${FEDORA_VER}/${arch}/repodata"
            aws s3 sync "s3://${AWS_BUCKET}/fedora/${FEDORA_VER}/${arch}/repodata" "${RPM_REPO_DIR}/${FEDORA_VER}/${arch}/repodata" --delete 2>/dev/null || true
        done

        # Run repogen for RPM packages with explicit version
        # repogen will create $version/$arch/Packages/ and $version/$arch/repodata/ inside output-dir
        repogen generate \
            --input-dir "$PACKAGES_RPM_DIR" \
            --output-dir "$RPM_REPO_DIR" \
            --incremental \
            --arch amd64,arm64 \
            --origin "Upsun" \
            --label "Upsun CLI" \
            --version "$FEDORA_VER" \
            --gpg-key "$GPG_KEY_FILE"

        # Sync this Fedora version back to S3 (without --delete to preserve existing packages)
        aws s3 sync "${RPM_REPO_DIR}/${FEDORA_VER}" "s3://${AWS_BUCKET}/fedora/${FEDORA_VER}"
    done

    echo "=== Done processing RPM packages ==="
fi

# --- Alpine Repository ---
if [[ "$PACKAGE_MANAGER" == "all" || "$PACKAGE_MANAGER" == "alpine" ]]; then
    echo "=== Processing Alpine packages ==="
    APK_REPO_DIR="${REPO_DIR}/alpine"
    mkdir -p "$APK_REPO_DIR"

    # Sync Alpine metadata from S3 (if exists)
    for arch in x86_64 aarch64; do
        mkdir -p "${APK_REPO_DIR}/${arch}"
        aws s3 cp "s3://${AWS_BUCKET}/alpine/${arch}/APKINDEX.tar.gz" "${APK_REPO_DIR}/${arch}/APKINDEX.tar.gz" 2>/dev/null || true
    done

    # Run repogen for Alpine packages with RSA signing
    repogen generate \
        --input-dir "$PACKAGES_APK_DIR" \
        --output-dir "$APK_REPO_DIR" \
        --incremental \
        --arch amd64,arm64 \
        --rsa-key "$RSA_KEY_FILE" \
        --key-name "repositories-upsun-com"

    # Copy RSA public key for users to download
    cp "$RSA_PUB_KEY_FILE" "${APK_REPO_DIR}/"

    # Sync Alpine back to S3
    aws s3 sync "${APK_REPO_DIR}" "s3://${AWS_BUCKET}/alpine"
    echo "=== Done processing Alpine packages ==="
fi

echo "All packages uploaded successfully!"

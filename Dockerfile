FROM ubuntu:24.04

# Install dependencies
RUN apt-get update && \
    apt-get install -y curl bash git ssh-client && \
    rm -rf /var/lib/apt/lists/*

# Install Upsun CLI
ARG VERSION=
RUN curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | VENDOR=upsun INSTALL_METHOD=raw VERSION=$VERSION bash

# Default command
ENTRYPOINT ["upsun"]

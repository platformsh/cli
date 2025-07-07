FROM ubuntu:latest

# Install dependencies
RUN apt-get update && \
    apt-get install -y curl bash && \
    rm -rf /var/lib/apt/lists/*

# Install Platform.sh CLI
ARG VERSION=
RUN curl -fsSL https://raw.githubusercontent.com/platformsh/cli/main/installer.sh | VENDOR=upsun INSTALL_METHOD=raw VERSION=$VERSION bash

# Default command
ENTRYPOINT ["upsun"]

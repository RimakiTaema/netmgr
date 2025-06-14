# Multi-stage build for NetMgr
FROM ubuntu:22.04 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    ninja-build \
    libjsoncpp-dev \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Build the application
RUN cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release && \
    cmake --build build --config Release

# Runtime stage
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libjsoncpp25 \
    iproute2 \
    iptables \
    net-tools \
    dnsutils \
    iputils-ping \
    traceroute \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Create netmgr user
RUN groupadd -r netmgr && useradd -r -g netmgr netmgr

# Copy binary from builder stage
COPY --from=builder /app/build/netmgr /usr/local/bin/netmgr

# Set capabilities for network operations
RUN setcap 'cap_net_admin,cap_net_raw+ep' /usr/local/bin/netmgr

# Create directories
RUN mkdir -p /etc/netmgr /var/log/netmgr && \
    chown netmgr:netmgr /etc/netmgr /var/log/netmgr

# Switch to netmgr user
USER netmgr

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/netmgr"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="NetMgr"
LABEL org.opencontainers.image.description="Cross-platform network management tool"
LABEL org.opencontainers.image.source="https://github.com/netmgr/netmgr"
LABEL org.opencontainers.image.licenses="MIT"

# Dockerfile for krakenv
# This Dockerfile is used by GoReleaser to build multi-arch images
# The binary is copied by GoReleaser during the build process

FROM alpine:3.21

# Install ca-certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Copy the binary (GoReleaser will copy the correct binary for each architecture)
COPY krakenv /usr/local/bin/krakenv

# Set working directory for mounted volumes
WORKDIR /workspace

# Run krakenv as the entrypoint
ENTRYPOINT ["krakenv"]


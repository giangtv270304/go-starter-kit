# syntax=docker/dockerfile:1.4

ARG GO_VERSION=1.25.4
ARG ALPINE_VERSION=3.22

########################
# --- Build Stage --- #
########################
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install build dependencies
RUN apk add --no-cache git gcc g++ make

WORKDIR /src

# Pre-copy go mod files for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Optional: Set build-time variables
ARG TARGETOS
ARG TARGETARCH

# Build the Go application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH make build

########################
# --- Final Stage --- #
########################
FROM alpine:${ALPINE_VERSION} AS final

# Add minimal runtime dependencies
RUN apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /src/go_starter_kit .

# Create user
RUN adduser -D appuser

# Create secrets directory and give it to appuser
RUN mkdir -p secrets && chown appuser:appuser secrets

# Add metadata labels
LABEL org.opencontainers.image.authors="Andy" \
  org.opencontainers.image.version="${GO_VERSION}" \
  org.opencontainers.image.description="Service built with Go ${GO_VERSION}"

# Use non-root user
USER appuser

# Use the entrypoint script
ENTRYPOINT ["/app/go_starter_kit"]

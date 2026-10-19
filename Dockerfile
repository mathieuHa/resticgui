# syntax=docker/dockerfile:1
# ResticGUI with restic, on Alpine. Build for several platforms with
#   docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 .
# The web UI and the Go binary are built on the build machine and
# cross-compiled; only the last stage runs on the target platform.
# The base images are written out, not build arguments, so that Dependabot
# can update them.

FROM --platform=$BUILDPLATFORM docker.io/library/node:24-alpine AS web
WORKDIR /front
COPY front/package.json front/package-lock.json ./
RUN npm ci
COPY front/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27-alpine AS go
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY internal internal
COPY --from=web /front/build internal/webui/ui
ARG TARGETOS TARGETARCH TARGETVARIANT
ARG VERSION=dev
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/resticgui ./cmd/resticgui

# restic's release binary (armv7 is linux_arm), checked against its SHA256SUMS
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.24 AS restic
ARG RESTIC_VERSION=0.19.1
ARG TARGETARCH
RUN apk add --no-cache curl bzip2
WORKDIR /dl
RUN set -eu; \
    file="restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2"; \
    base="https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}"; \
    curl -fsSLO "$base/$file"; \
    curl -fsSLO "$base/SHA256SUMS"; \
    grep " $file\$" SHA256SUMS | sha256sum -c -; \
    bunzip2 -c "$file" > /restic; \
    chmod 755 /restic

FROM docker.io/library/alpine:3.24
ARG VERSION=dev
# ca-certificates for the cloud backends, openssh-client for sftp, tzdata
# so that schedules follow TZ
RUN apk add --no-cache ca-certificates openssh-client tzdata
COPY --from=restic /restic /usr/bin/restic
COPY --from=go /out/resticgui /opt/resticgui
COPY cmd/resticgui/config-docker.yaml /opt/config.yaml
RUN mkdir -p /opt/db
WORKDIR /opt
ENV GIN_MODE=release
EXPOSE 3030
# the store and its key
VOLUME ["/opt/db"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD wget -q -O /dev/null http://127.0.0.1:3030/api/v2/version || exit 1
LABEL org.opencontainers.image.title="ResticGUI" \
      org.opencontainers.image.description="A web interface for restic backups" \
      org.opencontainers.image.source="https://github.com/mathieuHa/resticgui" \
      org.opencontainers.image.licenses="GPL-3.0-only" \
      org.opencontainers.image.version="${VERSION}"
CMD ["/opt/resticgui"]

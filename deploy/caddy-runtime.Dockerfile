# Temporary upstream-image gap: package verified Caddy releases in the existing
# public GMWE package. These runtime-only tags contain no application data.
FROM --platform=$BUILDPLATFORM docker.io/library/alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS verified
ARG TARGETARCH
WORKDIR /out
RUN case "$TARGETARCH" in \
      amd64) checksum=f032f8805e4df2ce5e1a8a624495c91e9ff763d8ff3ceddadd4784a07c472f40 ;; \
      arm64) checksum=6781a68e090fecda8c563b5a3b00115e18d7134084099ae804a017622471bed1 ;; \
      *) exit 1 ;; \
    esac && \
    wget -q "https://github.com/caddyserver/caddy/releases/download/v2.11.8/caddy_2.11.8_linux_${TARGETARCH}.tar.gz" -O caddy.tar.gz && \
    echo "$checksum  caddy.tar.gz" | sha256sum -c - && \
    tar -xzf caddy.tar.gz caddy && chmod 755 caddy

FROM docker.io/library/caddy:2.11.7-alpine@sha256:e70b164da81ef698b3ad19af799ea6086ce6d9e48ed4079eb9fe26f6fbf14ee5 AS standard
COPY --from=verified /out/caddy /usr/bin/caddy
LABEL org.opencontainers.image.title="Verified Caddy runtime" \
      org.opencontainers.image.version="v2.11.8"
RUN caddy version | grep -q '^v2.11.8 '

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS plugin
ARG TARGETARCH
ENV GOTOOLCHAIN=local GOMAXPROCS=2
RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@v0.4.7
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH xcaddy build v2.11.8 \
    --with github.com/caddy-dns/cloudflare@v0.2.4 --output /out/caddy

FROM standard AS cloudflare
COPY --from=plugin /out/caddy /usr/bin/caddy
LABEL org.opencontainers.image.title="Caddy runtime with Cloudflare v0.2.4"
RUN caddy version | grep -q '^v2.11.8 ' && caddy list-modules | grep -q '^dns.providers.cloudflare$'

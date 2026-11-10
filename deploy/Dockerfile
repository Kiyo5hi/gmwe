FROM docker.io/library/golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS api_build

RUN apk add --no-cache build-base

COPY . /src
WORKDIR /src
RUN go test ./api/... && mkdir -p /dist && \
    CGO_ENABLED=1 go build -o /dist/api ./api && \
    CGO_ENABLED=1 go build -o /dist/storage ./api/cmd/storage

FROM --platform=$BUILDPLATFORM docker.io/library/node:24.22.2@sha256:e7add2a490bb17d97b485b8641bd88455fff7839ade0f5eba493c3508124d413 AS client_build

RUN npm install --global npm@11.19.1

COPY . /src
WORKDIR /src/client
RUN npm ci && npm run check && npm run build && \
    cp -r dist /dist

FROM docker.io/library/caddy:2.11.8@sha256:4fcafa98afd1c2e642e353e3c5cf9ddbe892fcee14bc990e7ba2c5b78e51f9cf

ENV PORT=8080 ENV=release

WORKDIR /app
COPY --from=api_build /dist /app/api
COPY --from=client_build /dist /app/client
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY deploy/entrypoint.sh /app/entrypoint.sh

RUN apk add --no-cache libc6-compat

ENTRYPOINT [ "/app/entrypoint.sh" ]

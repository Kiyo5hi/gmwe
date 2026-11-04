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
RUN npm ci && npm run generate && \
    cp -r .output/public /dist

FROM ghcr.io/kiyo5hi/gmwe@sha256:a0c4ab539b935cd3fce294aa795619ac2bf3911e3e63d4950d4e29d4f8b55743

ENV PORT=8080 ENV=release

WORKDIR /app
COPY --from=api_build /dist /app/api
COPY --from=client_build /dist /app/client
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY deploy/entrypoint.sh /app/entrypoint.sh

RUN apk add --no-cache libc6-compat

ENTRYPOINT [ "/app/entrypoint.sh" ]

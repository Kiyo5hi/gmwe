FROM docker.io/library/golang:1.26-alpine AS api_build

RUN apk add --no-cache build-base

COPY . /src
WORKDIR /src
RUN go test ./api/... && mkdir -p /dist && \
    CGO_ENABLED=1 go build -o /dist/api ./api && \
    CGO_ENABLED=1 go build -o /dist/storage ./api/cmd/storage

FROM docker.io/library/node:lts AS client_build

COPY . /src
WORKDIR /src/client
RUN npm install && npm run generate && \
    cp -r .output/public /dist

FROM docker.io/library/caddy:2-alpine

ENV PORT=8080 ENV=release

WORKDIR /app
COPY --from=api_build /dist /app/api
COPY --from=client_build /dist /app/client
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY deploy/entrypoint.sh /app/entrypoint.sh

RUN apk add --no-cache libc6-compat

ENTRYPOINT [ "/app/entrypoint.sh" ]

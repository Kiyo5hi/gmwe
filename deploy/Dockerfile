FROM docker.io/library/golang:1.19 AS api_build

COPY . /src
WORKDIR /src
RUN mkdir -p /dist && go build -o /dist/api api/main.go

FROM docker.io/library/caddy:2-alpine

ENV PORT=8080 ENV=release

WORKDIR /app
COPY --from=api_build /dist /app/api
COPY client /app/client
COPY deploy/Caddyfile /etc/caddy/Caddyfile
COPY deploy/entrypoint.sh /app/entrypoint.sh

RUN apk add --no-cache libc6-compat

ENTRYPOINT [ "/app/entrypoint.sh" ]

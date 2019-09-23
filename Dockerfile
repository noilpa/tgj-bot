FROM golang:1.12-alpine AS build
RUN apk add --no-cache git gcc musl-dev sqlite && mkdir /app
WORKDIR /src
COPY go.mod go.sum ./ ./
RUN go mod download
RUN go build -o bin/tgj-bot ./cmd

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /src/bin/tgj-bot /usr/bin/tgj-bot
ENTRYPOINT ["tgj-bot", "/conf/conf.json"]
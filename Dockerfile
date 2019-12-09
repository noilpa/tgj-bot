FROM golang:1.12-alpine AS build
RUN apk add --no-cache git gcc && \
    mkdir /app
WORKDIR /src
COPY go.mod go.sum ./ ./
RUN go mod download && \
    go build -o bin/tgj-bot ./cmd

FROM alpine
RUN apk update && apk add ca-certificates && \
    rm -rf /var/cache/apk/*
RUN apk install -y gcc-c++ make
COPY --from=build /src/bin/tgj-bot /usr/bin/tgj-bot
ENTRYPOINT ["tgj-bot", "/conf/conf.json"]
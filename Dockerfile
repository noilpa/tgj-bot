FROM golang:1.12-alpine AS build

RUN apk add --no-cache git gcc musl-dev sqlite
WORKDIR /src
RUN mkdir /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY ./ ./
RUN go build -o bin/tgj-bot ./cmd

FROM alpine
COPY --from=build /src/bin/tgj-bot /usr/bin/tgj-bot
ENTRYPOINT ["tgj-bot", "/conf/conf.json"]
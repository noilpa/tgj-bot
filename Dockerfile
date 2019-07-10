FROM golang:1.12-alpine

RUN apk add --no-cache git gcc musl-dev

ENV GO111MODULE on

ADD . /go/src/tgj-bot/
WORKDIR /go/src/tgj-bot/
RUN go mod download
CMD ["tgj-bot"]

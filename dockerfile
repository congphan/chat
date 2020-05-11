FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /chat
RUN mkdir -p /chat/schema 

WORKDIR /chat

COPY ./schema/service.pb.go /chat/schema
COPY ./main.go /chat

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o chat .

CMD ./chat
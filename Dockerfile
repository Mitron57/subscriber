FROM golang:1.23 AS builder

WORKDIR /subscriber

COPY api api
COPY cmd cmd
COPY config config
COPY internal internal
COPY go.mod go.mod
COPY go.sum go.sum
COPY Makefile Makefile

RUN make server

FROM ubuntu:latest

COPY --from=builder /subscriber/subscriber /usr/bin

CMD ["subscriber"]
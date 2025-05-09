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

WORKDIR /subscriber

COPY --from=builder /subscriber/build/subscriber /subscriber

CMD ["./subscriber"]
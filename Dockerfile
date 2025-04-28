# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN go build -o server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["./server"]

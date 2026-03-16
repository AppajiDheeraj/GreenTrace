# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /out/greentrace ./

FROM alpine:3.19
WORKDIR /app
RUN adduser -D -g '' greentrace

COPY --from=builder /out/greentrace /app/greentrace

USER greentrace
ENTRYPOINT ["/app/greentrace"]

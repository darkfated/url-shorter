# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/url-shorter ./cmd/url-shorter

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates && adduser -D -g '' appuser
COPY --from=build /out/url-shorter /usr/local/bin/url-shorter

ENV HTTP_ADDR=:8080
ENV GIN_MODE=release
EXPOSE 8080

USER appuser
ENTRYPOINT ["/usr/local/bin/url-shorter"]

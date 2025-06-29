FROM golang:1.23-alpine AS builder
ENV CGO_ENABLED=0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /filesender ./cmd/filesender


FROM alpine:latest

RUN addgroup -S filesender && adduser -S -G filesender filesender
WORKDIR /app
ENV STATE_DIRECTORY=/app/data
RUN mkdir -p "$STATE_DIRECTORY" && chown -R filesender:filesender /app

COPY --from=builder /filesender /usr/local/bin/filesender

ENV FILESENDER_AUTH_METHOD=proxy MAX_UPLOAD_SIZE=2147483648
USER filesender
EXPOSE 8080

ENTRYPOINT ["filesender", "-listen", ":8080"]

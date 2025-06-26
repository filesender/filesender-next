FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
COPY . .
RUN go build -ldflags="-s -w" -o /filesender ./cmd/filesender


FROM alpine:3.20

RUN addgroup -S filesender && adduser -S -G filesender filesender
WORKDIR /app
ENV STATE_DIRECTORY=/app/data
RUN mkdir -p "$STATE_DIRECTORY" && chown -R filesender:filesender /app

COPY --from=builder /filesender /usr/local/bin/filesender

ENV FILESENDER_AUTH_METHOD=dummy MAX_UPLOAD_SIZE=2147483648
USER filesender
EXPOSE 8080

ENTRYPOINT ["filesender", "-listen", "0.0.0.0:8080"]

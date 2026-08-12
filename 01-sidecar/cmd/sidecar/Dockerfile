FROM golang:1.26 AS builder

WORKDIR /app

# we dont have any external lib's but still...
COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o sidecar ./cmd/sidecar

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/sidecar .

EXPOSE 8888

# default
ENV LOG_PATH="logs/2026/app.log"

CMD ["./sidecar"]

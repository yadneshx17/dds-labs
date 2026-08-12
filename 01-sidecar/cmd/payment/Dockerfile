FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# build context from project root
RUN go build -o payment ./cmd/payment

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/payment .

EXPOSE 8081

CMD ["./payment"]

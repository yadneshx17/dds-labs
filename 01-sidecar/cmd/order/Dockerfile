FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# build context from project root
RUN go build -o order ./cmd/order

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/order .

EXPOSE 8080

CMD ["./order"]

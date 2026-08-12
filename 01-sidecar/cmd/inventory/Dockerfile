FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

# build context from project root
RUN go build -o inventory ./cmd/inventory

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/inventory .

EXPOSE 8082

CMD ["./inventory"]

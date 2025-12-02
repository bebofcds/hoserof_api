FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 3000

CMD ["./server"]

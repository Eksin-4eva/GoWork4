# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

# Run stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/uploads

EXPOSE 8888

CMD ["./server"]

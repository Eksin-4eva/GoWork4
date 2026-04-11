FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /gateway ./gateway/gateway.go && \
    CGO_ENABLED=0 go build -o /chat ./chat/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /gateway /chat ./
COPY gateway/etc/gateway-api-docker.yaml etc/gateway-api.yaml
COPY chat/etc/chat-docker.yaml etc/chat.yaml
RUN mkdir -p uploads/avatars uploads/videos uploads/images

EXPOSE 8888 8889
CMD ./chat -f etc/chat.yaml & ./gateway -f etc/gateway-api.yaml

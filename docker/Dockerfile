# 多阶段构建：编译三个可执行文件，运行镜像里只留二进制与运行时依赖。
FROM golang:1.25-alpine AS builder

WORKDIR /src

# 先只拷依赖清单再下载，利用层缓存：只改代码时不会重新拉依赖
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api    ./cmd/api    && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/chat   ./cmd/chat   && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/user    ./cmd/user    && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/video   ./cmd/video   && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/comment ./cmd/comment

FROM alpine:3.20

# ffmpeg   worker 抽帧的运行时依赖
# tzdata   time.LoadLocation("Asia/Shanghai") 需要它，否则驱动解析时间会失败
# ca-certificates  将来访问 HTTPS 外部服务
RUN apk add --no-cache ffmpeg tzdata ca-certificates && \
    adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/ /app/bin/
COPY config/config.docker.yaml /app/config/config.yaml

# 不用 root 跑服务
USER app

# 8888 api / 8889 chat / 8881 user / 8882 video / 8883 comment
EXPOSE 8888 8889 8881 8882 8883

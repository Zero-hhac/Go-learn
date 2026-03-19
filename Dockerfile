# 使用官方 Go 镜像作为构建环境
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建可执行文件
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用轻量级的 alpine 镜像作为运行环境
FROM alpine:latest

WORKDIR /app

# 复制构建好的二进制文件
COPY --from=builder /app/main .

# 复制静态资源和模板
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/public ./public

# 暴露端口（默认 8080）
EXPOSE 8080

# 运行程序
CMD ["./main"]

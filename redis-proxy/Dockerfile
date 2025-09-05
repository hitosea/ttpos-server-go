# 使用官方Go语言镜像作为基础镜像
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

# 设置代理
ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖，这一步会生成缓存，在依赖有变化时才会重新下载
RUN go mod download

# 复制源代码
COPY . .

# 编译Go程序
RUN go build -o main ./main.go

# 添加执行权限
RUN chmod +x main

# 使用轻量级的镜像来运行应用
FROM ubuntu:alpine

# 安装必要的依赖（如有需要）
# RUN apk add --no-cache <必要的包>

# 设置工作目录
WORKDIR /var/www

# 复制编译好的二进制文件到新镜像
COPY --from=builder /app/main .

# 运行应用
CMD ["./main"]
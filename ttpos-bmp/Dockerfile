# 不要直接在 应用的目录下执行docker build ，通过最上层目录的 docker compose build bmp-takeout 来构建
# 使用官方Go语言镜像作为基础镜像
FROM golang:1.23 AS builder
#构建模块
ARG MODULE=takeout
ARG HTTP_PORT=14031
ARG RPC_PORT=14032
# 设置工作目录
WORKDIR /app/ttpos-bmp

# 设置代理
ENV GOPROXY=https://goproxy.cn,direct
RUN go install github.com/gogf/gf/cmd/gf/v2@latest
# 复制源代码
COPY . .

# 下载依赖，这一步会生成缓存，在依赖有变化时才会重新下载
RUN  go mod download

# 编译Go程序
RUN cd app/ttpos-$MODULE && gf build -a amd64 -s linux -ew -o bin/ttpos-$MODULE

# 添加执行权限
RUN chmod +x "app/ttpos-$MODULE/bin/ttpos-$MODULE"

# 使用轻量级的镜像来运行应用
FROM ubuntu:latest
#构建模块
ARG MODULE=takeout
ARG HTTP_PORT=14031
ARG RPC_PORT=14032

# 安装 ca-certificates 和 curl，并清除 apt 缓存
RUN apt-get update && apt-get install -y ca-certificates curl tzdata \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /app

# 复制编译好的二进制文件到新镜像
COPY --from=builder /app/ttpos-bmp/app/ttpos-$MODULE/bin/ttpos-$MODULE .
COPY manifest/deploy/entrypoint.sh ./

ENV GRPC_ENDPOINTS=""
# 生成 Core Dump 文件
ENV GOTRACEBACK=crash
ENV APP_NAME="ttpos-$MODULE"

# 暴露应用运行的端口
EXPOSE $HTTP_PORT
EXPOSE $RPC_PORT

# 运行应用
ENTRYPOINT ["./entrypoint.sh"]
#CMD ["./$APP_NAME"]
# 使用官方Go语言镜像作为基础镜像
FROM apache/skywalking-go:0.5.0-go1.23 AS builder

# 设置工作目录
WORKDIR /app

# 设置代理
# ENV GOPROXY=https://goproxy.cn,direct
RUN wget http://cmdb.keli.vip/static/assets/script/apache-skywalking-go-0.6.0-bin.tar.gz && tar xvf apache-skywalking-go-0.6.0-bin.tar.gz

#注意 context 要设置为上级目录,因为要拷贝别的模块的文件
#复制ttpos-bmp 模块 api
WORKDIR /ttpos-bmp
COPY  ttpos-bmp/go.mod ttpos-bmp/go.sum  ./
COPY  ttpos-bmp/app/ttpos-takeout/api  app/ttpos-takeout/api
COPY  ttpos-bmp/app/ttpos-erp/api  app/ttpos-erp/api
COPY  ttpos-bmp/app/ttpos-shop/api  app/ttpos-shop/api
COPY  ttpos-bmp/app/ttpos-manager/api  app/ttpos-manager/api

WORKDIR /app

# 复制go.mod和go.sum文件
COPY main/go.mod main/go.sum ./

# 下载依赖，这一步会生成缓存，在依赖有变化时才会重新下载
RUN go mod download
# 复制源代码
COPY main .

RUN  /app/apache-skywalking-go-0.6.0-bin/bin/skywalking-go-agent-0.6.0-linux-amd64   -inject ./main.go
RUN  go build -toolexec=/app/apache-skywalking-go-0.6.0-bin/bin/skywalking-go-agent-0.6.0-linux-amd64  -a -ldflags="-s -w"  -o main ./main.go


# 添加执行权限
RUN chmod +x main

# 使用轻量级的镜像来运行应用
FROM ubuntu:latest

ENV SW_AGENT_NAME=ttpos-server-main
ENV SW_AGENT_REPORTER_GRPC_BACKEND_SERVICE=103.63.136.110:30435
ENV SW_AGENT_REPORTER_GRPC_CHECK_INTERVAL=20


# 安装 ca-certificates 和 curl
RUN apt-get update && apt-get install -y ca-certificates curl golang

RUN apt-get clean && apt-get autoclean && apt-get autoremove --yes \
&& rm -rf /var/lib/{apt,dpkg,cache,log}/

# 设置工作目录
WORKDIR /var/www

# 复制编译好的二进制文件到新镜像
COPY --from=builder /app/main .

# 暴露应用运行的端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
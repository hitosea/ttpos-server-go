LOCAL_IP := $(shell ifconfig | grep "inet " | grep "192" | awk '{print $$2}' | head -n 1)
GO_PATH := $(shell go env GOPATH)

# 运行run
run:
	if [ ! -f "./main" ]; then \
		go build -o main ./main.go; \
	fi
	./main

# 启动开发模式 - 热重启
dev:
	${GO_PATH}/bin/fresh -c ./fresh.conf
	
# test	
test:
	go run ./main.go test
	
# 构建镜像	
build-image:
	docker build -t ttpos-server-websocket .
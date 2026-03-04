# 构建项目 - 生产
run:
	if [ ! -f "./main" ]; then \
		go build -o main ./main.go; \
	fi
	./main

# 构建镜像
build-image:
	docker build -t ttpos-server-go .
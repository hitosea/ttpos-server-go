# 运行run
run:
	if [ ! -f "./main" ]; then \
		go build -o main ./main.go; \
	fi
	./main

# 启动开发模式 - 热重启
dev:
	${HOME}/go/bin/fresh -c ./fresh.conf
	
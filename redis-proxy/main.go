package main

import (
	"fmt"
	"log"
	"redis-proxy/cache"
	"redis-proxy/config"
	"redis-proxy/proxy"

	"go.uber.org/zap"
)

// @BasePath /api/v1
func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化redis
	cache.Init(cache.Redis, cache.Config{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	// 创建并启动服务器
	server, err := proxy.NewServer(":16379")
	if err != nil {
		fmt.Println("Error creating Redis proxy", zap.Error(err))
		return
	}

	fmt.Println("Redis proxy starting on port 16379...")

	if err := server.Start(); err != nil {
		fmt.Println("Error starting Redis proxy", zap.Error(err))
		return
	}
}

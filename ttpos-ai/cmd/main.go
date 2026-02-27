package main

import (
	"fmt"
	"log"
	"ttpos-ai/config"
	"ttpos-ai/internal/handler"
	"ttpos-ai/internal/service"
	"ttpos-ai/pkg/database"
	"ttpos-ai/pkg/llm"
	"ttpos-ai/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer logger.Sync()

	// 设置Gin模式
	if config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 初始化数据库（可选）
	dbConnected := false
	if config.Database.User != "" {
		if err := database.Init(); err != nil {
			log.Printf("数据库连接失败（Agent功能将不可用）: %v", err)
		} else {
			dbConnected = true
			log.Println("数据库连接成功")
		}
	}

	// 静态文件服务
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"service":  "ttpos-ai",
			"llm":      llm.IsEnabled(),
			"database": dbConnected,
		})
	})

	// 注册API路由
	api := r.Group("/api/v1")

	if llm.IsEnabled() {
		// 聊天服务
		chatService, err := service.NewChatService()
		if err != nil {
			log.Fatalf("创建聊天服务失败: %v", err)
		}
		chatHandler := handler.NewChatHandler(chatService)
		api.POST("/chat", chatHandler.Chat)
		api.POST("/chat/simple", chatHandler.SimpleChat)

		// Agent服务（需要数据库）
		if dbConnected {
			agentService, err := service.NewAgentService()
			if err != nil {
				log.Printf("创建Agent服务失败: %v", err)
			} else {
				agentHandler := handler.NewAgentHandler(agentService)
				api.POST("/agent/query", agentHandler.Query)
				log.Println("Agent服务已启用")
			}
		}
	}

	// 启动服务
	addr := fmt.Sprintf(":%s", config.Server.Port)
	log.Printf("ttpos-ai 服务启动: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

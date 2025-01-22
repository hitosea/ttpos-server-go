package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"jjjshop-server-go/config"
	"jjjshop-server-go/i18n"
	"jjjshop-server-go/middleware"
	"jjjshop-server-go/pkg/cache"
	"jjjshop-server-go/pkg/database"
	"jjjshop-server-go/pkg/logger"
	"jjjshop-server-go/router"
)

// @title ttpos-server-go API
// @version 1.0
// @description 点餐系统服务端接口文档

// @securityDefinitions.apikey JwtToken
// @in header
// @name Authorization

// @host 127.0.0.1:8081
// @BasePath /
func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	// 初始化日志系统
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Logger.Sync()

	// 初始化国际化
	i18n.Init()
	// 初始化数据库管理器
	var dbm *database.DBManager = database.GetDBManager(config.Database)

	gin.SetMode(config.Server.Mode)
	// 创建Gin引擎
	r := gin.New()
	// 添加中间件
	r.Use(middleware.Cors())
	r.Use(middleware.Recovery(logger.Logger))
	// 注册 Swagger 路由
	setupSwagger(r)

	var cacheConfig cache.Config
	_ = copier.Copy(&cacheConfig, &config.Redis)
	// 注册路由
	router.Setup(r, dbm, cache.NewCache(cache.Redis, cacheConfig))

	// 启动服务器
	if err := r.Run(":" + config.Server.Port); err != nil {
		logger.Logger.Fatal("Error starting server", zap.Error(err))
	}
}

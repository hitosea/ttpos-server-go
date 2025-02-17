package command

import (
	"fmt"
	"log"
	"os"
	"ttpos-server-go/config"
	"ttpos-server-go/docs"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/router"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

var rootCommand = &cobra.Command{
	Use:   "ttpos",
	Short: "启动服务",
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}
		// 初始化国际化
		i18n.Init()

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)
		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化雪花ID生成器
		//database.InitSonyFlakeId()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()
		// 初始化数据库管理器
		var dbm *database.DBManager = database.GetDBManager(config.Database)
		// 初始化系统事件总线
		event.NewSystemBus()

		gin.SetMode(config.Server.Mode)
		// 创建Gin引擎
		r := gin.New()
		// 添加中间件
		r.Use(middleware.Cors())
		r.Use(middleware.Recovery(logger.Logger, config.Server.Mode))
		// 注册 Swagger 路由
		// 允许自定义Swagger文档链接
		docs.SwaggerInfo.BasePath = "/api/v1"
		// Swagger API 文档
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// 注册路由
		router.Setup(r, dbm, cache.Global)

		internalRouter := gin.Default()
		router.SetupInternal(internalRouter, dbm, cache.Global)
		go func() {
			// 启动内网服务
			if err := internalRouter.Run(":9000"); err != nil {
				fmt.Printf("Failed to start internal server: %v\n", err)
			}
		}()
		// 启动服务器
		if err := r.Run(":" + config.Server.Port); err != nil {
			logger.Logger.Fatal("Error starting server", zap.Error(err))
		}
	},
}

func Execute() {
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

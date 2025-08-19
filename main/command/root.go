package command

import (
	"fmt"
	"log"
	"os"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/queue"
	"ttpos-server-go/app/tasks"
	"ttpos-server-go/config"
	"ttpos-server-go/docs"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"
	"ttpos-server-go/router"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/robfig/cron/v3"
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

		// 自定义验证规则
		validator.Init()

		// 初始化id生成器
		utils.InitIdGenerator()

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化短信客户端
		sms.InitClient(config.SMS.APIKey, config.SMS.BaseURL, config.SMS.ProjectName)
		// 检查短信客户端配置
		if err := sms.GetSMSClient().CheckConfig(); err != nil {
			logger.Logger.Info("Failed to check SMS client config", zap.Error(err))
		}
		//初始化服务发现
		cloud.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 初始化数据库管理器
		var dbm *database.DBManager = database.GetDBManager(config.Database)

		// 初始化系统事件总线
		event.NewSystemBus()

		// 定时器
		initializeTimers(dbm, cache.Global)

		//初始化延迟消息
		initQueue()

		// 外网服务
		initializeExternalService(dbm, cache.Global)
	},
}

// Execute 执行命令
func Execute() {
	rootCommand.CompletionOptions.DisableDefaultCmd = true
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 外网服务
func initializeExternalService(dbm *database.DBManager, cache cache.Cache) {
	// 启动服务器
	gin.SetMode(config.Server.Mode)
	// 创建Gin引擎
	r := gin.New()
	// 添加中间件
	r.Use(middleware.Cors())
	// 添加请求参数日志中间件
	if config.Server.Mode == "debug" {
		r.Use(middleware.Recovery(logger.Logger, config.Server.Mode))
		r.Use(middleware.RequestLogger(logger.Logger))
	} else {
		r.Use(gin.Logger(), middleware.Recovery(logger.Logger, config.Server.Mode))
	}
	// 注册 Swagger 路由
	// 允许自定义Swagger文档链接
	docs.SwaggerInfo.BasePath = "/api/v1"
	// Swagger API 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 注册路由
	router.Setup(r, dbm, cache)
	if err := r.Run(":" + config.Server.Port); err != nil {
		logger.Logger.Fatal("Error starting server", zap.Error(err))
	}
}

// 初始化定时器任务
func initializeTimers(dbm *database.DBManager, cache cache.Cache) {
	c := cron.New(cron.WithSeconds())

	// 1秒检查打印
	// _, _ = c.AddFunc("*/1 * * * * *", func() {
	// 	printer_tasks.NewPrinterTask(dbm, cache).Execute()
	// })

	// NOTE: 舍弃5分钟自动切换
	// 1分钟检查Usb打印是否在线
	// _, _ = c.AddFunc("0 */2 * * * *", func() {
	// 	tasks.NewUsbPrintTask(dbm, cache).Execute()
	// })

	// 删除7天前的打印日志
	_, _ = c.AddFunc("0 6 * * *", func() {
		tasks.NewDelPrintTask(dbm, cache).Execute()
	})

	// 启动定时器
	c.Start()
}

// 初始化延迟消息队列
func initQueue() {
	queue.Init()
}

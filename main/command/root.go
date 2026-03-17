package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	objectStorageAdapter "ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/queue"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/tasks"
	"ttpos-server-go/config"
	"ttpos-server-go/docs"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/gormcache"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/metrics"
	"ttpos-server-go/pkg/otlp"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"
	"ttpos-server-go/router"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
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
			fmt.Printf("[FATAL] Failed to check SMS client config: %v\n", err)
			logger.Logger.Info("Failed to check SMS client config", zap.Error(err))
		}

		//初始化服务发现
		cloud.Init()

		// 初始化 OTLP 调用链跟踪
		if err := otlp.Init(context.Background(), config.Otlp); err != nil {
			fmt.Printf("[FATAL] Failed to initialize OpenTelemetry: %v\n", err)
			logger.Logger.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		}

		// 为 Redis 启用链路追踪（必须在 otlp.Init 之后）
		cache.EnableTracing()
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer logger.Logger.Sync()

		// 初始化数据库管理器
		var dbm *database.DBManager = database.GetDBManager(config.Database)

		// ==================== GORM 查询缓存（一键注释此区块可禁用）====================
		fmt.Println("[GORM Cache] 启用 GORM 查询缓存")
		gormcache.Init(&gormcache.Config{
			Easer:   true,            // 启用请求合并（相同查询只执行一次）
			TTL:     5 * time.Minute, // 缓存 5 分钟
			MaxRows: 1000,            // 超过 1000 行不缓存
			AllowDBs: []string{ // 内测商户白名单（仅这些商户启用缓存）
				"shop3290397937664000",
			},
			Tables: []string{ // 只缓存低频更新的表
				// ==================== 商品相关 ====================
				"ttpos_product",
				"ttpos_product_category",
				"ttpos_product_attribute",
				"ttpos_product_package",
				"ttpos_product_bom",
				"ttpos_product_label",
				"ttpos_product_flavor",        // 商品规格
				"ttpos_product_sauce",         // 商品小料
				"ttpos_product_unit",          // 商品单位
				"ttpos_product_bom_card",      // 商品成本卡
				"ttpos_product_package_group", // 商品套餐组

				// ==================== 桌位相关 ====================
				"ttpos_table_area",
				"ttpos_table",
				"ttpos_desk_region",     // 桌位区域
				"ttpos_desk_type",       // 桌位类型
				"ttpos_desk_map_layout", // 桌位地图布局

				// ==================== 支付相关 ====================
				"ttpos_payment_method",
				"ttpos_payment_order",

				// ==================== 会员相关 ====================
				"ttpos_member_point_log",
				"ttpos_member_level",
				"ttpos_member_card_type",

				// ==================== 自助餐相关 ====================
				"ttpos_buffet",
				"ttpos_buffet_package",
				"ttpos_buffet_customer_type",
				"ttpos_buffet_customer_type_price",

				// ==================== 订单属性表 ====================
				"ttpos_sale_order_buffet_customer_type",
				"ttpos_sale_order_buffet_delay_product",
				"ttpos_sale_order_product_attribute",
				"ttpos_sale_order_product_bom",
				"ttpos_sale_order_product_reason",
				"ttpos_sale_order_coupon",
				"ttpos_return_order_product",
				"ttpos_production_order_product",

				// ==================== 订单配置字典 ====================
				"ttpos_order_remark",       // 整单备注字典
				"ttpos_order_item_remark",  // 单品备注字典
				"ttpos_return_food_reason", // 退菜原因字典
				"ttpos_order_source",       // 外卖来源配置
				"ttpos_production_order",   // 生产订单

				// ==================== 原料物料 ====================
				"ttpos_material",      // 原料表
				"ttpos_material_unit", // 原料单位

				// ==================== 访问控制 ====================
				"ttpos_access", // 权限表
				"ttpos_role",   // 角色表

				// ==================== 财务配置 ====================
				"ttpos_tax",      // 税种表
				"ttpos_supplier", // 供应商表

				// ==================== 其他基础数据 ====================
				"ttpos_batch_tag",
				"ttpos_multi_language_name",
				"ttpos_company",
				"ttpos_company_setting",
				"ttpos_printer",
			},
		})
		database.SetOnDBReady(func(db *gorm.DB, dbName string) {
			gormcache.Enable(db, dbName, nil)
		})
		dbm.EnableCacheForAllDBs()
		// ==================== GORM 查询缓存结束 ====================

		// 初始化对象存储缓存配置的缓存层（使用三级缓存基础设施）
		objectStorageAdapter.InitObjectStorageCacheConfigCache(
			cache.Global,
			func() *gorm.DB {
				return dbm.GetDB(constant.DefaultDB)
			},
		)

		// 初始化缓存对象控制器（单例模式）
		repository.InitCacheObjectController(cache.Global, 5*time.Minute)

		// 初始化并启动缓存命中率快照保存器（已禁用）
		// objectStorageAdapter.InitSnapshotReporter(
		// 	dbm.GetDB,
		// 	utils.GetInstanceID(),
		// 	1*time.Minute,
		// )

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
	pprof.Register(r)
	// 添加中间件
	r.Use(middleware.Cors())

	// 开启 OTLP 调用链跟踪
	if config.Otlp.Enabled {
		r.Use(otlp.OtlpMiddleware(config.Otlp.ServiceName))
		// 添加记录特殊headers中间件
		r.Use(otlp.RecordSpecialHeaders())
	}
	// Prometheus metrics
	if config.Metrics.Enabled {
		metrics.Init()
		r.Use(metrics.Middleware())

		// Expose metrics on separate port
		metricsRouter := gin.New()
		metricsRouter.GET("/metrics", metrics.Handler())

		utils.Go(func() {
			logger.Logger.Info("Metrics server started", zap.String("port", config.Metrics.Port))
			if err := metricsRouter.Run(":" + config.Metrics.Port); err != nil {
				logger.Logger.Error("Metrics server failed", zap.Error(err))
			}
		})
	}

	// 添加请求参数日志中间件
	if config.Server.Mode == "debug" {
		r.Use(gin.Logger(), middleware.Recovery(logger.Logger, config.Server.Mode))
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

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":" + config.Server.Port,
		Handler: r,
	}

	// 启动服务器的 goroutine
	utils.Go(func() {
		logger.Logger.Info("HTTP 服务器启动", zap.String("port", config.Server.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("[FATAL] HTTP 服务器启动失败: %v\n", err)
			logger.Logger.Fatal("HTTP 服务器启动失败", zap.Error(err))
		}
	})

	// 等待中断信号来优雅关闭服务器
	gracefulShutdown(srv)
}

// gracefulShutdown 优雅关闭服务器和相关资源
func gracefulShutdown(srv *http.Server) {
	// 创建一个接收系统信号的通道
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT (Ctrl+C) 和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-quit
	logger.Logger.Info("收到关闭信号，开始优雅关闭...")

	// 创建一个带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	logger.Logger.Info("正在关闭 HTTP 服务器...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("HTTP 服务器关闭失败", zap.Error(err))
	} else {
		logger.Logger.Info("HTTP 服务器已关闭")
	}

	// 关闭 RocketMQ 消费者
	logger.Logger.Info("正在关闭 RocketMQ 消费者...")
	if err := queue.Shutdown(); err != nil {
		logger.Logger.Error("RocketMQ 消费者关闭失败", zap.Error(err))
	} else {
		logger.Logger.Info("RocketMQ 消费者已关闭")
	}

	// 关闭 Nacos 客户端
	logger.Logger.Info("正在关闭 Nacos 客户端...")
	if err := cloud.Shutdown(); err != nil {
		logger.Logger.Error("Nacos 客户端关闭失败", zap.Error(err))
	} else {
		logger.Logger.Info("Nacos 客户端已关闭")
	}

	//关闭 OTLP
	if err := otlp.Shutdown(ctx); err != nil {
		logger.Logger.Error("OTLP 关闭失败", zap.Error(err))
	} else {
		logger.Logger.Info("OTLP 已关闭")
	}

	// 关闭数据库连接
	logger.Logger.Info("正在关闭数据库连接...")
	database.CloseAll()
	logger.Logger.Info("数据库连接已关闭")

	// 同步日志
	logger.Logger.Sync()

	logger.Logger.Info("应用程序已优雅关闭")
}

// 初始化定时器任务
func initializeTimers(dbm *database.DBManager, cache cache.Cache) {
	c := cron.New(cron.WithSeconds())

	// 删除7天前的打印日志
	_, _ = c.AddFunc("0 6 * * *", func() {
		tasks.NewDelPrintTask(dbm, cache).Execute()
	})

	// // 每小时执行销售出库汇总任务
	_, _ = c.AddFunc("0 0 * * * *", func() {
		// 每5分钟执行一次
		// _, _ = c.AddFunc("*/5 * * * * *", func() {
		// _, _ = c.AddFunc("0 * * * * *", func() {
		tasks.NewDailySalesOutboundSummaryTask(dbm, cache).Execute()
	})

	// 每15分钟
	_, _ = c.AddFunc("0 */15 * * * *", func() {
		// _, _ = c.AddFunc("0 * * * * *", func() {
		tasks.NewKitchenEfficiencyAnalysisTask(dbm, cache).Execute()
	})

	// 每分钟执行翻译任务
	_, _ = c.AddFunc("15 * * * * *", func() {
		tasks.NewTranslateWhenSyncTask(dbm, cache).Execute()
	})

	// 每分钟执行检查物品库存任务
	_, _ = c.AddFunc("30 * * * * *", func() {
		tasks.NewCheckMaterialStockTask(dbm, cache).Execute()
	})

	// 每天凌晨0点执行 ERP Stock Entry 合并扣减任务
	_, _ = c.AddFunc("0 0 0 * * *", func() {
		tasks.NewErpStockEntryTask(dbm, cache).Execute()
	})

	// 每小时执行自动收货任务（时区前置过滤，仅处理到达本地午夜的门店）
	_, _ = c.AddFunc("0 0 * * * *", func() {
		tasks.NewAutoReceiptTask(dbm, cache).Execute()
	})

	// 启动定时器
	c.Start()
}

// 初始化延迟消息队列
func initQueue() {
	queue.Init()
}

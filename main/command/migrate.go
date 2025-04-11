package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCommand.AddCommand(migrateDataCmd)
}

// 数据迁移
var migrateDataCmd = &cobra.Command{
	Use:   "migrate-data",
	Short: "run data migration",
	Long:  `run data migration`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// // 初始化配置
		// if err := config.Init(); err != nil {
		// 	log.Fatalf("Failed to initialize config: %v", err)
		// }

		// // 初始化日志系统
		// if err := logger.Init(); err != nil {
		// 	log.Fatalf("Failed to initialize logger: %v", err)
		// }

		// // 初始化全局缓存引擎
		// var cacheConfig cache.Config
		// _ = copier.Copy(&cacheConfig, &config.Redis)
		// cache.Init(cache.Redis, cacheConfig)

		// // 初始化Redis分布式并发锁
		// lock.InitRedisLock(cacheConfig)
		// lock.NewSystemLock()

		// fmt.Println("Starting data migration...")

		// 打印所有命令行参数
		fmt.Println("命令行参数:")
		fmt.Printf(" args: %#v\n", args) // 使用%#v可以显示更详细的信息
		fmt.Printf(" args长度: %d\n", len(args))

		// 打印环境变量
		fmt.Println("环境变量:")
		fmt.Printf("  ARGS: %v\n", os.Getenv("ARGS"))

		// 打印所有标志
		fmt.Println("所有标志:")
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			fmt.Printf("  %s: %s\n", flag.Name, flag.Value.String())
		})

		// if company != "" {
		// 	fmt.Println("Company mode enabled")
		// }
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running data migration...")
		// 在这里添加您的数据迁移逻辑
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

// func NewMySQLConnection(conf config.DatabaseConf, dbName string) (*gorm.DB, error) {
// 	ignoreRecordNotFoundError := true // 忽略ErrRecordNotFound（记录未找到）错误
// 	logLevel := logger.Warn
// 	if config.Server.Mode == gin.DebugMode { // 调试模式
// 		ignoreRecordNotFoundError = false // 不忽略
// 		logLevel = logger.Info            // 详细日志
// 	}
// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&charset=utf8mb4&parseTime=True&loc=Local",
// 		"root",
// 		conf.RootPassword,
// 		conf.Host,
// 		conf.Port,
// 		dbName,
// 	)
// 	// 初始化会话
// 	return gorm.Open(mysql.Open(dsn), &gorm.Config{
// 		NamingStrategy: schema.NamingStrategy{
// 			TablePrefix:   conf.TablePrefix, // 表名前缀
// 			SingularTable: true,             // 使用单一表名, eg. `User` => `user`
// 		},
// 		DisableForeignKeyConstraintWhenMigrating: true, // 禁用自动创建外键约束
// 		SkipDefaultTransaction:                   true, // 禁用默认事务
// 		Logger: logger.New(
// 			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的地方）
// 			logger.Config{
// 				SlowThreshold:             time.Duration(conf.SlowQueryTime) * time.Second, // 慢查询阈值
// 				LogLevel:                  logLevel,                                        // 日志级别
// 				IgnoreRecordNotFoundError: ignoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
// 				Colorful:                  true,                                            // 彩色打印
// 			},
// 		),
// 	})
// }

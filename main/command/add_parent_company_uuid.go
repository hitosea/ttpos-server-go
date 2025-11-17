package command

import (
	"fmt"
	"log"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(addParentCompanyUuidCmd)
}

// 添加父级公司UUID路径
var addParentCompanyUuidCmd = &cobra.Command{
	Use:   "add-parent-company-uuid",
	Short: "add parent company uuid path",
	Long:  `add parent company uuid path, 添加上级公司UUID路径`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}
		config.Server.Mode = "release"

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化Redis分布式并发锁
		lock.InitRedisLock(cacheConfig)
		lock.NewSystemLock()

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 为命令行环境创建同时输出到控制台和文件的logger
		setupConsoleLogger()

		// 初始化id生成器
		utils.InitIdGenerator()

		//初始化服务发现
		cloud.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var siteCode string
		// 目标的公司UUID
		fmt.Printf("%s 输入要更新父级路径的site_code，默认为2: %s", blueColor, resetColor)
		fmt.Scanln(&siteCode)
		if siteCode == "" {
			siteCode = "2"
		}

		var dbm *database.DBManager = database.GetDBManager(config.Database)
		err := erp.NewIErpSrv(dbm).UpdateTtposCompanyParentUuids(siteCode)
		if err != nil {
			fmt.Printf("%s 更新父级路径失败 %s\n", redColor, resetColor)
			return
		}

		fmt.Printf("%s 数据更新完成 %s\n", greenColor, resetColor)
	},
}

package command

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/printer"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/trans/handler"

	ttposContext "ttpos-server-go/pkg/context"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/redis/go-redis/v9"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	companyIdStr    string
	companyUuid     uint64
	sourceCompanyId int
	sourceDB        *gorm.DB
	targetDB        *gorm.DB
	targetSaasDB    *gorm.DB
)

// 使用ANSI转义序列设置文本颜色
var (
	redColor    = "\033[31m" // 红色
	resetColor  = "\033[0m"  // 重置颜色
	greenColor  = "\033[32m" // 绿色
	yellowColor = "\033[33m" // 黄色
	blueColor   = "\033[34m" // 蓝色
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
		// 初始化配置
		if err := config.Init(); err != nil {
			log.Fatalf("%sFailed to initialize config: %v%s", redColor, err, resetColor)
		}

		// 初始化全局缓存引擎
		var cacheConfig cache.Config
		_ = copier.Copy(&cacheConfig, &config.Redis)
		cache.Init(cache.Redis, cacheConfig)

		// 初始化日志系统
		if err := logger.Init(); err != nil {
			log.Fatalf("Failed to initialize logger: %v", err)
		}

		// 初始化id生成器
		utils.InitIdGenerator()

		//确认
		fmt.Printf("%s 输入要迁移的旧公司ID继续: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 数据迁移已取消 %s\n", redColor, resetColor)
			return
		}
		if _, err := fmt.Sscanf(companyIdStr, "%d", &sourceCompanyId); err != nil {
			fmt.Printf("%s 错误: 公司ID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
			return
		}
		config.MigrateDatabase.MigrateOldDBDatabase = fmt.Sprintf("%s%d", constant.DBNamePrefix, sourceCompanyId)

		// 检查 MigrateDatabase 各字段是否有值
		var missingFields []string
		if config.MigrateDatabase.MigrateOldDBHost == "" {
			missingFields = append(missingFields, "MigrateOldDBHost")
		}
		if config.MigrateDatabase.MigrateOldDBPort == 0 {
			missingFields = append(missingFields, "MigrateOldDBPort")
		}
		if config.MigrateDatabase.MigrateOldDBUser == "" {
			missingFields = append(missingFields, "MigrateOldDBUser")
		}
		if config.MigrateDatabase.MigrateOldDBPassword == "" {
			missingFields = append(missingFields, "MigrateOldDBPassword")
		}
		if config.MigrateDatabase.MigrateOldDBDatabase == "" {
			missingFields = append(missingFields, "MigrateOldDBDatabase")
		}
		if config.MigrateDatabase.MigrateOldDBPrefix == "" {
			missingFields = append(missingFields, "MigrateOldDBPrefix")
		}
		if len(missingFields) > 0 {
			fmt.Printf("%s 错误: env配置缺少以下数据库配置字段: %s\n", redColor, resetColor)
			for _, field := range missingFields {
				fmt.Printf("%s  - %s%s\n", redColor, field, resetColor)
			}
			return
		}

		// 初始化源数据库连接
		source, err := database.NewMySQLConnection(config.DatabaseConf{
			Host:        config.MigrateDatabase.MigrateOldDBHost,
			Port:        config.MigrateDatabase.MigrateOldDBPort,
			User:        config.MigrateDatabase.MigrateOldDBUser,
			Password:    config.MigrateDatabase.MigrateOldDBPassword,
			Database:    config.MigrateDatabase.MigrateOldDBDatabase,
			TablePrefix: config.MigrateDatabase.MigrateOldDBPrefix,
		}, config.MigrateDatabase.MigrateOldDBDatabase)
		if err != nil {
			fmt.Printf("%s 错误: 连接源数据库失败 %s\n", redColor, resetColor)
			return
		}
		sourceDB = source
		fmt.Printf("%s 连接源数据库成功 %s\n", greenColor, resetColor)

		// 目标的公司UUID
		fmt.Printf("%s 输入要迁移目标的公司UUID继续: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 数据迁移已取消 %s\n", redColor, resetColor)
			return
		}
		if _, err := fmt.Sscanf(companyIdStr, "%d", &companyUuid); err != nil {
			fmt.Printf("%s 错误: 公司ID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
			return
		}

		// 初始化目标数据库连接
		targetSaas, err := database.NewMySQLConnection(config.Database, "saas")
		if err != nil {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 连接数据库失败, 检查是否正确配置了数据库信息，以及 -c 参数是否正确 %s\n", redColor, resetColor)
			return
		}
		targetSaasDB = targetSaas

		// 初始化数据库
		companyDB, err := database.NewMySQLConnection(config.Database, fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid))
		if err != nil {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 连接数据库失败, 检查是否正确配置了数据库信息，以及 -c 参数是否正确 %s\n", redColor, resetColor)
			return
		}
		fmt.Printf("%s 连接数据库成功 %s\n", greenColor, resetColor)

		// 将公司数据库实例保存到全局变量中，供 Run 函数使用
		targetDB = companyDB
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 确保数据库连接有效
		if sourceDB == nil {
			fmt.Printf("%s 错误: 源数据库连接无效 %s\n", redColor, resetColor)
			return
		}
		if targetDB == nil {
			fmt.Printf("%s 错误: 数据库连接无效 %s\n", redColor, resetColor)
			return
		}

		// 执行sql查询
		type Supplier struct {
			Name string `gorm:"default:'';comment:'名称'"`
		}
		query := fmt.Sprintf("SELECT * FROM %ssupplier limit 1", config.MigrateDatabase.MigrateOldDBPrefix)
		var supplier Supplier
		if err := sourceDB.Raw(query).Scan(&supplier).Error; err != nil {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 查询旧数据库公司信息失败 %s\n", redColor, resetColor)
			return
		}

		// 获取公司信息
		companyRepo := repository.NewCompanyRepo(targetDB)
		company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
		if err != nil && !strings.Contains(err.Error(), "record not found") {
			fmt.Printf("%s %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 错误: 获取新数据库公司信息失败 %s\n", redColor, resetColor)
			return
		}

		// 二次确认
		fmt.Printf("%s 将要从旧数据库 %s 迁移数据到新数据库 %s %s\n", redColor, config.MigrateDatabase.MigrateOldDBDatabase, fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid), resetColor)
		fmt.Printf("%s 旧数据库的商家名称为： %s %s\n", redColor, supplier.Name, resetColor)
		if company != nil {
			fmt.Printf("%s 目标数据库的商家名称为： %s %s\n", redColor, company.Name, resetColor)
		} else {
			fmt.Printf("%s 目标数据库的商家名称为空 %s\n", redColor, resetColor)
		}
		fmt.Printf("%s 注意：新数据库将会被清空，请确认是否要开始数据迁移？%s\n", redColor, resetColor)

		// 读取用户输入
		var confirmation string
		fmt.Printf("%s 输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s 数据迁移已取消 %s\n", redColor, resetColor)
			return
		}

		// 读取用户输入
		verificationCode := fmt.Sprintf("%05d", rand.Intn(90000)+10000) // 生成五位数字验证码
		fmt.Printf("%s 请输入验证码 '%s' 继续，输入其他内容取消: %s", yellowColor, verificationCode, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != verificationCode {
			fmt.Printf("%s 验证码错误，数据迁移已取消 %s\n", redColor, resetColor)
			return
		}

		// 开始数据迁移
		fmt.Printf("%s 开始数据迁移... %s\n", blueColor, resetColor)

		// 一. 清空目标数据库
		if !truncateTable(targetDB) {
			return
		}

		// 二.数据迁移
		err = handler.Run(sourceDB, targetDB, targetSaasDB, sourceCompanyId, companyUuid)

		if err != nil {
			fmt.Printf("%s 数据迁移失败 : %s %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 数据迁移完成 %s\n", greenColor, resetColor)

		// 三.清除缓存
		if err := clearCache(companyUuid); err != nil {
			fmt.Printf("%s 清除缓存失败 : %s %s\n", redColor, err, resetColor)
		} else {
			fmt.Printf("%s 清除缓存完成 %s\n", greenColor, resetColor)
		}
	},
}

// 清空目标数据库
func truncateTable(targetDB *gorm.DB) bool {
	// 1. 判断 desk 表 是否存在数据
	var deskCount int64
	targetDB.Raw("SELECT COUNT(*) FROM ttpos_desk").Scan(&deskCount)
	if deskCount > 0 {
		// 二次确认
		fmt.Printf("%s 目标数据库已经存在生产数据，是否确认清空？ %s %s\n", redColor, fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid), resetColor)
		var confirmation string
		fmt.Printf("%s 输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s 数据迁移已取消 %s\n", redColor, resetColor)
			return false
		}
	}

	// 2. 清空目标数据库
	fmt.Printf("%s 开始清空目标数据库... %s\n", blueColor, resetColor)
	var tables []string
	targetDB.Raw("SHOW TABLES").Scan(&tables)
	for _, table := range tables {
		if table == "ttpos_printer_type" {
			continue
		}
		if table == "ttpos_migrations" {
			continue
		}
		targetDB.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", table))
	}
	fmt.Printf("%s 清空目标数据库完成 %s\n", greenColor, resetColor)

	return true
}

func clearCache(companyUuid uint64) error {
	client := cache.Global.GetClient()
	clusterClient := cache.Global.GetClusterClient()
	if client == nil && clusterClient == nil {
		return errors.New("获取缓存对象失败")
	}

	type CacheStruct struct {
		Type    string   `json:"type"`
		Key     string   `json:"key"`
		Keys    string   `json:"keys"`
		Name    string   `json:"name"`
		DirPath []string `json:"dir_path"`
	}
	companyUuidStr := fmt.Sprintf("%d", companyUuid)
	cacheMap := map[string]CacheStruct{
		"category": {
			Type: "cache",
			Key:  "category_" + companyUuidStr,
			Name: "商品分类",
		},
		"setting": {
			Type: "cache",
			Key:  "setting:company_id:" + companyUuidStr,
			Keys: "setting:company_id:" + companyUuidStr,
			Name: "商城设置",
		},
		"app": {
			Type: "cache",
			Key:  "app_" + companyUuidStr,
			Name: "应用设置",
		},
		"agent": {
			Type: "cache",
			Key:  "agent_setting_" + companyUuidStr,
			Name: "分销设置",
		},
		"temp": {
			Type:    "file",
			Name:    "临时图片",
			DirPath: []string{},
		},
		"common": {
			Type: "common",
			Key:  "common_" + companyUuidStr,
			Name: "公共",
		},
	}

	fn := func(c redis.Cmdable) {
		ctx := context.Background()
		for s, cacheStruct := range cacheMap {
			switch {
			case cacheStruct.Type == "cache":
				if cacheStruct.Key != "" {
					c.Del(ctx, cacheStruct.Key)
				}
				if cacheStruct.Keys != "" {
					c.Del(ctx, cacheStruct.Keys)
				}
				if s == "app" {
					c.Del(ctx, "app_mp_"+companyUuidStr)
					c.Del(ctx, "app_wx_"+companyUuidStr)
				}
			case cacheStruct.Type == "common":
				c.Del(ctx, "tag:"+cryptor.Md5String(cacheStruct.Key))
				c.Del(ctx, "tag:"+cryptor.Md5String("firstshop"))
				c.Del(ctx, "tag:"+cryptor.Md5String("cashier"))
				c.Del(ctx, "tag:"+cryptor.Md5String("unit"))
			case s == "category":
				c.Del(ctx, "tag:"+cryptor.Md5String(fmt.Sprintf("category%d00", companyUuid)))
				c.Del(ctx, "tag:"+cryptor.Md5String(fmt.Sprintf("category%d01", companyUuid)))
				c.Del(ctx, "tag:"+cryptor.Md5String(fmt.Sprintf("category%d10", companyUuid)))
				c.Del(ctx, "tag:"+cryptor.Md5String(fmt.Sprintf("category%d11", companyUuid)))
			}
		}

		// 删除设置缓存
		c.Del(ctx, fmt.Sprintf("setting:company_id:%d", companyUuid))
		c.Del(ctx, "tag:"+cryptor.Md5String("common_get_settingLanguages"))
		c.Del(ctx, "sync_setting_"+constant.SettingCloudBasic)

		// 删除商品打印机列表缓存
		ctxs := ttposContext.NewContext()
		ctxs.SetCompanyUuid(companyUuid)
		printer.NewPrinterRepo(ctxs).DeleteProductPrinterListCache()
	}

	if client != nil {
		fn(client)
	} else {
		fn(clusterClient)
	}
	return nil
}

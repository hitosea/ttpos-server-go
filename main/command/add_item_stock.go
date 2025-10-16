package command

import (
	"fmt"
	"log"
	"strings"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func init() {
	rootCommand.AddCommand(addItemStockCmd)
}

var (
	warehouseErpCode string
	warehouseId      uint64
	materialCode     string
	actualNum        float64
)

// 添加物品库存
var addItemStockCmd = &cobra.Command{
	Use:   "add-item-stock",
	Short: "add item stock",
	Long:  `add item stock, 添加物品库存`,
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

		// 目标的公司UUID
		fmt.Printf("%s 输入要添加物品库存的公司UUID: %s", blueColor, resetColor)
		fmt.Scanln(&companyIdStr)
		if companyIdStr == "" {
			fmt.Printf("%s 添加物品库存已取消 %s\n", redColor, resetColor)
			return
		}
		if _, err := fmt.Sscanf(companyIdStr, "%d", &companyUuid); err != nil {
			fmt.Printf("%s 错误: 添加物品库存的公司UUID必须是有效的数字，当前值: %s%s\n", redColor, companyIdStr, resetColor)
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

		//初始化服务发现
		cloud.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if targetDB == nil {
			fmt.Printf("%s 错误: 数据库连接无效 %s\n", redColor, resetColor)
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
		fmt.Printf("%s 数据库的商家名称为：%s %s\n", redColor, company.Name, resetColor)

		// 先显示所有可用的仓库编码
		fmt.Printf("%s 查询可用的仓库编码... %s\n", yellowColor, resetColor)
		warehouseRepo := repository.NewWarehouseRepo(targetDB)
		warehouses, err := warehouseRepo.Get(warehouseRepo.WhereHeadquarterUuid(0))
		if err != nil {
			fmt.Printf("%s 错误: 获取仓库列表失败 %s %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 可用的仓库编码列表: %s\n", greenColor, resetColor)
		for _, wh := range warehouses {
			fmt.Printf(" %v : %s / %s\n", wh.ID, wh.ErpCode, language.JsonToLocaleResponse(wh.Name).ZH)
		}

		// 读取用户输入仓库ERP编码
		fmt.Printf("%s 请从上面列表中选择仓库ID: %s", yellowColor, resetColor)
		fmt.Scanln(&warehouseId)
		if warehouseId == 0 {
			fmt.Printf("%s 错误: 仓库不能为空 %s\n", redColor, resetColor)
			return
		}

		// 获取默认仓库ID
		warehouse, err := warehouseRepo.GetById(warehouseId)
		if err != nil {
			fmt.Printf("%s 仓库不存在: %s %s\n", redColor, err, resetColor)
			fmt.Printf("%s 请确保输入的仓库ID上面列表中的完全一致 %s\n", yellowColor, resetColor)
			return
		}
		fmt.Printf("%s 选择的仓库为：%s %s\n", redColor, language.JsonToLocaleResponse(warehouse.Name).ZH, resetColor)

		// 读取用户输入物料编码
		fmt.Printf("%s 输入物料编码: %s", yellowColor, resetColor)
		fmt.Scanln(&materialCode)
		if materialCode == "" {
			fmt.Printf("%s 错误: 物料编码不能为空 %s\n", redColor, resetColor)
			return
		}
		material, err := repository.NewMaterialRepo(targetDB).GetMaterialByErpCode(materialCode)
		if err != nil {
			fmt.Printf("%s 物料不存在: %s %s\n", redColor, err, resetColor)
			return
		}
		fmt.Printf("%s 选择的物料为：%s %s\n", redColor, language.JsonToLocaleResponse(material.Name).ZH, resetColor)

		// 读取用户输入实际数量
		fmt.Printf("%s 输入实际数量: %s", yellowColor, resetColor)
		fmt.Scanln(&actualNum)
		if actualNum <= 0 {
			fmt.Printf("%s 错误: 实际数量必须大于0 %s\n", redColor, resetColor)
			return
		}
		fmt.Printf("%s 选择的实际数量为：%f %s\n", redColor, actualNum, resetColor)

		// 二次确认
		fmt.Printf("%s 将要从数据库【%s】的仓库【%s】添加物品【%s】的库存 %f %s \n", redColor,
			fmt.Sprintf("%s%d", constant.DBNamePrefix, companyUuid), language.JsonToLocaleResponse(warehouse.Name).ZH, language.JsonToLocaleResponse(material.Name).ZH, float64(actualNum), resetColor)
		fmt.Printf("%s 数据库的商家名称为：%s %s\n", redColor, company.Name, resetColor)

		// 读取用户输入
		var confirmation string
		fmt.Printf("%s 输入 'yes' 继续，输入其他内容取消: %s", yellowColor, resetColor)
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Printf("%s 添加物品库存已取消 %s\n", redColor, resetColor)
			return
		}

		// 开始添加物品库存
		fmt.Printf("%s 开始更新库存... %s\n", blueColor, resetColor)
		warehouseItemRepo := repository.NewWarehouseItemRepo(targetDB)
		warehouseItem, err := warehouseItemRepo.GetByWarehouseAndMaterial(warehouse.Uuid, material.Uuid)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("%s 物品库存不存在，创建物品库存... %s\n", blueColor, resetColor)
				warehouseItem := &model.WarehouseItem{
					WarehouseUuid: warehouse.Uuid,
					MaterialUuid:  material.Uuid,
					Stock:         float64(actualNum),
					ReservedStock: 0,
					Valuation:     1.00,
				}
				err = warehouseItemRepo.Create(warehouseItem)
				if err != nil {
					fmt.Printf("%s 创建物品库存失败: %s %s\n", redColor, err, resetColor)
				}
				fmt.Printf("%s 添加成功 %s\n", greenColor, resetColor)
				return
			}
			fmt.Printf("%s 失败: %s %s\n", redColor, err, resetColor)
			return
		}

		// 增加物品库存
		err = warehouseItemRepo.AddStock(warehouseItem.Uuid, float64(actualNum))
		if err != nil {
			fmt.Printf("%s 增加物品库存失败: %s %s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s 添加成功 %s\n", greenColor, resetColor)

		// 当前库存
		warehouseItem, err = warehouseItemRepo.GetByUuid(warehouseItem.Uuid)
		if err != nil {
			fmt.Printf("%s 获取物品库存失败: %s %s\n", redColor, err, resetColor)
			return
		}
		fmt.Printf("%s 当前库存为：%f %s\n", redColor, warehouseItem.Stock, resetColor)
	},
}

package command

import (
	"fmt"
	"log"
	"ttpos-server-go/app/cloud"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	// 迁移命令参数
	migrateProductSelectionRangeCompanyUuidFlag uint64
)

func init() {
	migrateProductSelectionRangeCmd.Flags().Uint64Var(&migrateProductSelectionRangeCompanyUuidFlag, "company-uuid", 0, "指定商家UUID（不指定则迁移所有商家）")
	rootCommand.AddCommand(migrateProductSelectionRangeCmd)
}

// 商品选择范围旧数据迁移命令
var migrateProductSelectionRangeCmd = &cobra.Command{
	Use:   "migrate-product-selection-range",
	Short: "migrate product selection range data",
	Long:  `migrate product selection range data, 迁移商品选择范围旧数据（属性、加料、套餐分组）`,
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
		fmt.Printf("%s===== 商品选择范围数据迁移开始 =====%s\n", blueColor, resetColor)

		var dbm *database.DBManager = database.GetDBManager(config.Database)

		// 显示迁移范围
		if migrateProductSelectionRangeCompanyUuidFlag > 0 {
			fmt.Printf("%s迁移范围：指定商家 (UUID: %d)%s\n", blueColor, migrateProductSelectionRangeCompanyUuidFlag, resetColor)
		} else {
			fmt.Printf("%s迁移范围：所有商家%s\n", blueColor, resetColor)
		}

		// 确认操作
		fmt.Printf("\n%s警告：此操作将迁移商品选择范围的旧数据，包括：%s\n", yellowColor, resetColor)
		fmt.Printf("  1. 加料选择范围（sauce_required → sauce_min_selection）\n")
		fmt.Printf("  2. 属性选择范围（is_must → min_selection）\n")
		fmt.Printf("  3. 套餐分组可选范围（group_type → optional_min_count）\n")
		fmt.Printf("\n%s是否继续？(y/N): %s", yellowColor, resetColor)

		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Printf("%s操作已取消%s\n", yellowColor, resetColor)
			return
		}

		// 执行迁移
		var err error
		if migrateProductSelectionRangeCompanyUuidFlag > 0 {
			// 迁移指定商家
			err = migrateProductSelectionRangeDataForCompany(dbm, migrateProductSelectionRangeCompanyUuidFlag)
		} else {
			// 迁移所有商家
			err = migrateProductSelectionRangeDataForAllCompanies(dbm)
		}

		if err != nil {
			fmt.Printf("%s迁移失败: %v%s\n", redColor, err, resetColor)
			return
		}

		fmt.Printf("%s===== 数据迁移完成 =====%s\n", greenColor, resetColor)
	},
}

// migrateProductSelectionRangeDataForCompany 为指定商家执行数据迁移
func migrateProductSelectionRangeDataForCompany(dbm *database.DBManager, companyUuid uint64) error {
	// 从 saas 数据库获取指定公司信息
	saasDB := dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)
	company, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return fmt.Errorf("获取商家信息失败: %w", err)
	}
	if company == nil {
		return fmt.Errorf("商家不存在 (UUID: %d)", companyUuid)
	}

	fmt.Printf("%s正在迁移商家: %s (UUID: %d)%s\n\n", blueColor, company.Name, company.Uuid, resetColor)

	// 获取公司数据库
	companyDB := dbm.GetDB(company.Uuid)
	if companyDB == nil {
		return fmt.Errorf("获取数据库连接失败")
	}

	// 执行数据迁移
	if err := migrateProductSelectionRangeData(companyDB); err != nil {
		return fmt.Errorf("商家 %s (UUID: %d) 迁移失败: %w", company.Name, company.Uuid, err)
	}

	fmt.Printf("%s✓ 商家 %s (UUID: %d) 迁移成功%s\n", greenColor, company.Name, company.Uuid, resetColor)
	return nil
}

// migrateProductSelectionRangeDataForAllCompanies 为所有公司执行数据迁移
func migrateProductSelectionRangeDataForAllCompanies(dbm *database.DBManager) error {
	// 从 saas 数据库获取所有公司列表
	saasDB := dbm.GetDB(constant.DefaultDB)
	var companies []model.Company
	if err := saasDB.Scopes(repository.NotDeleted).Find(&companies).Error; err != nil {
		return fmt.Errorf("获取公司列表失败: %w", err)
	}

	if len(companies) == 0 {
		fmt.Printf("%s没有找到任何公司，跳过迁移%s\n", yellowColor, resetColor)
		return nil
	}

	fmt.Printf("%s共找到 %d 个公司，开始逐个迁移...%s\n\n", blueColor, len(companies), resetColor)

	successCount := 0
	failedCompanies := []string{}

	// 遍历每个公司执行迁移
	for i, company := range companies {
		fmt.Printf("%s[%d/%d] 正在迁移公司: %s (UUID: %d)%s\n",
			blueColor, i+1, len(companies), company.Name, company.Uuid, resetColor)

		// 获取公司数据库
		companyDB := dbm.GetDB(company.Uuid)
		if companyDB == nil {
			errMsg := fmt.Sprintf("公司 %s (UUID: %d): 获取数据库连接失败", company.Name, company.Uuid)
			fmt.Printf("%s  ✗ %s%s\n", redColor, errMsg, resetColor)
			failedCompanies = append(failedCompanies, errMsg)
			continue
		}

		// 执行该公司的数据迁移
		if err := migrateProductSelectionRangeData(companyDB); err != nil {
			errMsg := fmt.Sprintf("公司 %s (UUID: %d): %v", company.Name, company.Uuid, err)
			fmt.Printf("%s  ✗ %s%s\n", redColor, errMsg, resetColor)
			failedCompanies = append(failedCompanies, errMsg)
			continue
		}

		fmt.Printf("%s  ✓ 公司 %s (UUID: %d) 迁移成功%s\n\n", greenColor, company.Name, company.Uuid, resetColor)
		successCount++
	}

	// 输出统计结果
	fmt.Printf("%s========== 迁移统计 ==========%s\n", blueColor, resetColor)
	fmt.Printf("总公司数: %d\n", len(companies))
	fmt.Printf("%s成功: %d%s\n", greenColor, successCount, resetColor)

	if len(failedCompanies) > 0 {
		fmt.Printf("%s失败: %d%s\n", redColor, len(failedCompanies), resetColor)
		fmt.Printf("\n%s失败的公司列表：%s\n", redColor, resetColor)
		for _, msg := range failedCompanies {
			fmt.Printf("  - %s\n", msg)
		}
		return fmt.Errorf("部分公司迁移失败")
	}

	return nil
}

// migrateProductSelectionRangeData 执行单个数据库的数据迁移
func migrateProductSelectionRangeData(db *gorm.DB) error {

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	fmt.Printf("%s[1/5] 迁移加料必选状态（sauce_required → sauce_min_selection）...%s\n", blueColor, resetColor)
	// 1. 迁移加料：sauce_required → sauce_min_selection
	// 开启必选时，最小选择=1
	sql1 := `
		UPDATE ttpos_product_package 
		SET sauce_min_selection = CASE 
			WHEN sauce_required = 1 THEN 1 
			ELSE 0 
		END
		WHERE sauce_min_selection = 0
	`
	if err := tx.Exec(sql1).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("迁移加料必选状态失败: %w", err)
	}
	fmt.Printf("%s  ✓ 加料必选状态迁移完成%s\n", greenColor, resetColor)

	fmt.Printf("%s[2/5] 修正加料最大可选数量（sauce_max_selection）...%s\n", blueColor, resetColor)
	// 2. 修正加料：sauce_max_selection = 0 的情况
	// 如果未设置最大可选（为0），则设置为该商品的加料数量
	sql2 := `
		UPDATE ttpos_product_package pp
		SET pp.sauce_max_selection = (
			SELECT COUNT(DISTINCT pb.product_sauce_uuid)
			FROM ttpos_product_bom pb
			WHERE pb.product_package_uuid = pp.uuid
			AND pb.product_sauce_uuid > 0
			AND pb.delete_time = 0
		)
		WHERE pp.sauce_max_selection = 0
		AND EXISTS (
			SELECT 1
			FROM ttpos_product_bom pb
			WHERE pb.product_package_uuid = pp.uuid
			AND pb.product_sauce_uuid > 0
			AND pb.delete_time = 0
		)
	`
	if err := tx.Exec(sql2).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("修正加料最大可选数量失败: %w", err)
	}
	fmt.Printf("%s  ✓ 加料最大可选数量修正完成%s\n", greenColor, resetColor)

	fmt.Printf("%s[3/5] 迁移属性必选状态（is_must → min_selection）...%s\n", blueColor, resetColor)
	// 3. 迁移 is_must → min_selection
	sql3 := `
		UPDATE ttpos_product_package_attribute_group 
		SET min_selection = CASE 
			WHEN is_must = 1 THEN 1 
			ELSE 0 
		END
		WHERE min_selection = 0
	`
	if err := tx.Exec(sql3).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("迁移属性必选状态失败: %w", err)
	}
	fmt.Printf("%s  ✓ 属性必选状态迁移完成%s\n", greenColor, resetColor)

	fmt.Printf("%s[4/5] 修正属性最大可选数量（max_selection）...%s\n", blueColor, resetColor)
	// 4. 修正 max_selection = 0 的情况
	// 如果 max_selection 为 0，设置为属性值数量
	sql4 := `
		UPDATE ttpos_product_package_attribute_group ppag
		SET ppag.max_selection = (
			SELECT COUNT(*)
			FROM ttpos_product_package_attribute ppa
			WHERE ppa.product_package_attribute_group_uuid = ppag.uuid
			AND ppa.delete_time = 0
		)
		WHERE ppag.max_selection = 0
		AND EXISTS (
			SELECT 1
			FROM ttpos_product_package_attribute ppa
			WHERE ppa.product_package_attribute_group_uuid = ppag.uuid
			AND ppa.delete_time = 0
		)
	`
	if err := tx.Exec(sql4).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("修正属性最大可选数量失败: %w", err)
	}
	fmt.Printf("%s  ✓ 属性最大可选数量修正完成%s\n", greenColor, resetColor)

	fmt.Printf("%s[5/5] 迁移套餐分组可选范围（group_type → optional_min_count）...%s\n", blueColor, resetColor)
	// 5. 迁移套餐分组的可选范围
	// 可选分组：设置 optional_min_count = optional_count（旧的 optional_count 实际代表必须选择的数量）
	sql5 := `
		UPDATE ttpos_product_package_group 
		SET optional_min_count = optional_count
		WHERE group_type = 1 
		AND optional_min_count = 0
	`
	if err := tx.Exec(sql5).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("迁移套餐分组可选范围失败: %w", err)
	}
	fmt.Printf("%s  ✓ 套餐分组可选范围迁移完成%s\n", greenColor, resetColor)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

package handler

import (
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB, targetSassDB *gorm.DB, companyUuid uint64) error {
	userService := v1.NewUserGradeService(sourceDB, targetDB)
	err := userService.ConvertUserGrade()
	if err != nil {
		return err
	}

	// 桌台区域
	tableAreaService := v1.NewTableAreaService(sourceDB, targetDB)
	err = tableAreaService.ConvertTableArea()
	if err != nil {
		return err
	}

	// 商家表
	appService := v1.NewAppService(sourceDB, targetDB, v1.WithCompanyUuid(companyUuid))
	err = appService.ConvertApp()
	if err != nil {
		return err
	}

	// 钱箱
	shopAccountService := v1.NewShopAccountService(sourceDB, targetDB)
	err = shopAccountService.ConvertShopAccount()
	if err != nil {
		return err
	}

	// 月度统计
	erpMonthlyStatisticsService := v1.NewErpMonthlyStatisticsService(sourceDB, targetDB)
	err = erpMonthlyStatisticsService.ConvertERPMonthlyStatistics()
	if err != nil {
		return err
	}

	// 自助餐顾客类型
	customerTypeService := v1.NewCustomerTypeService(sourceDB, targetDB)
	err = customerTypeService.ConvertCustomerType()
	if err != nil {
		return err
	}

	// 商品分类
	categoryService := v1.NewCategoryService(sourceDB, targetDB)
	err = categoryService.ConvertCategory()
	if err != nil {
		return err
	}

	// 桌台类型
	tableTypeService := v1.NewTableTypeService(sourceDB, targetDB)
	err = tableTypeService.ConvertTableType()
	if err != nil {
		return err
	}

	// 退菜原因
	returnReasonService := v1.NewReturnReasonService(sourceDB, targetDB)
	err = returnReasonService.ConvertReturnReason()
	if err != nil {
		return err
	}

	// 赠菜免单原因
	freeTagService := v1.NewFreeTagService(sourceDB, targetDB)
	err = freeTagService.ConvertFreeTag()
	if err != nil {
		return err
	}

	// 支付方式
	payTypeService := v1.NewPayTypeService(sourceDB, targetDB)
	err = payTypeService.ConvertPayType()
	if err != nil {
		return err
	}

	// 打印机模板
	printerTemplateService := v1.NewPrinterTemplateService(sourceDB, targetDB)
	err = printerTemplateService.ConvertPrinterTemplate()
	if err != nil {
		return err
	}

	// 角色
	shopRoleService := v1.NewShopRoleService(sourceDB, targetDB)
	err = shopRoleService.ConvertShopRole()
	if err != nil {
		return err
	}

	// 权限
	shopAccessService := v1.NewShopAccessService(sourceDB, targetDB)
	err = shopAccessService.ConvertShopAccess()
	if err != nil {
		return err
	}

	// 角色权限
	shopRoleAccess := v1.NewShopRoleAccessService(sourceDB, targetDB)
	err = shopRoleAccess.ConvertShopRoleAccess()
	if err != nil {
		return err
	}

	// 设置
	settingService := v1.NewSettingService(sourceDB, targetDB)
	err = settingService.ConvertSetting()
	if err != nil {
		return err
	}

	// 员工
	shopUserService := v1.NewShopUserService(sourceDB, targetDB, v1.WithCompanyUuid(companyUuid))
	err = shopUserService.ConvertShopUser()
	if err != nil {
		return err
	}

	return nil
}

package handler

import (
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB, targetSassDB *gorm.DB, targetCompanyUuid uint64) error {
	userGradeService := v1.NewUserGradeService(sourceDB, targetDB)
	err := userGradeService.ConvertUserGrade()
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
	appService := v1.NewAppService(sourceDB, targetDB, targetCompanyUuid)
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
	shopUserService := v1.NewShopUserService(sourceDB, targetDB, targetSassDB, targetCompanyUuid)
	err = shopUserService.ConvertShopUser()
	if err != nil {
		return err
	}

	// 商品属性和属性组
	attributeService := v1.NewAttributeService(sourceDB, targetDB)
	err = attributeService.ConvertAttribute()
	if err != nil {
		return err
	}

	// 商品单位
	productUnitService := v1.NewProductUnitService(sourceDB, targetDB)
	err = productUnitService.ConvertProductUnit()
	if err != nil {
		return err
	}

	// 商品打印标签
	productPrintLabelService := v1.NewProductPrintLabelService(sourceDB, targetDB)
	err = productPrintLabelService.ConvertProductPrintLabel()
	if err != nil {
		return err
	}

	// 规格
	specService := v1.NewSpecService(sourceDB, targetDB)
	err = specService.ConvertSpec()
	if err != nil {
		return err
	}

	// 商品和材料
	productService := NewProductService(sourceDB, targetDB)
	err = productService.ConvertProduct()
	if err != nil {
		return err
	}

	// 商品小料库
	productSauceService := NewProductSauceService(sourceDB, targetDB)
	err = productSauceService.ConvertProductSauce()
	if err != nil {
		return err
	}

	// 供应商
	erpSupplierService := v1.NewErpSupplierService(sourceDB, targetDB)
	err = erpSupplierService.ConvertErpSupplier()
	if err != nil {
		return err
	}

	// 自助餐加钟
	buffetDelayService := v1.NewBuffetDelayService(sourceDB, targetDB)
	err = buffetDelayService.ConvertBuffetDelay()
	if err != nil {
		return err
	}

	// 自助餐
	buffetService := NewBuffetService(sourceDB, targetDB)
	err = buffetService.ConvertBuffet()
	if err != nil {
		return err
	}

	// 会员
	userService := v1.NewUserService(sourceDB, targetDB)
	err = userService.ConvertUser()
	if err != nil {
		return err
	}

	// 会员卡领取记录
	userCardRecordService := v1.NewUserCardRecordService(sourceDB, targetDB)
	err = userCardRecordService.ConvertUserCardRecord()
	if err != nil {
		return err
	}

	// 会员积分变动记录
	userPointsLogService := v1.NewUserPointsLogService(sourceDB, targetDB)
	err = userPointsLogService.ConvertUserPointsLog()
	if err != nil {
		return err
	}

	// 会员余额变动记录
	userBalanceLogService := v1.NewUserBalanceLogService(sourceDB, targetDB)
	err = userBalanceLogService.ConvertUserBalanceLog()
	if err != nil {
		return err
	}

	// 会员卡
	userCardService := v1.NewUserCardService(sourceDB, targetDB)
	err = userCardService.ConvertUserCard()
	if err != nil {
		return err
	}

	// 桌台
	tableService := v1.NewTableService(sourceDB, targetDB)
	err = tableService.ConvertTable()
	if err != nil {
		return err
	}

	return nil
}

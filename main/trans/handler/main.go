package handler

import (
	"ttpos-server-go/app/errors"
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB, targetSassDB *gorm.DB, sourceCompanyId int, targetCompanyUuid uint64) error {
	var err error
	// 桌台
	{
		// 桌台区域
		tableAreaService := v1.NewTableAreaService(sourceDB, targetDB)
		err = tableAreaService.ConvertTableArea()
		if err != nil {
			return err
		}

		// 桌台类型
		tableTypeService := v1.NewTableTypeService(sourceDB, targetDB)
		err = tableTypeService.ConvertTableType()
		if err != nil {
			return err
		}

		// 桌台
		tableService := v1.NewTableService(sourceDB, targetDB)
		err = tableService.ConvertTable()
		if err != nil {
			return err
		}
	}

	// 设备
	{
		shopBindRecordService := v1.NewShopBindRecordService(sourceDB, targetDB)
		err = shopBindRecordService.ConvertShopBindRecord()
		if err != nil {
			return err
		}
	}

	// 设置
	{
		// 设置
		settingService := v1.NewSettingService(sourceDB, targetDB, targetSassDB, targetCompanyUuid)
		err = settingService.ConvertSetting()
		if err != nil {
			return err
		}
	}

	// 商家、员工
	{
		// 商家表
		appService := v1.NewAppService(sourceDB, targetSassDB, targetDB, sourceCompanyId, targetCompanyUuid)
		err = appService.ConvertApp()
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

		// 员工
		shopUserService := v1.NewShopUserService(sourceDB, targetDB, targetSassDB, targetCompanyUuid)
		err = shopUserService.ConvertShopUser()
		if err != nil {
			return err
		}

		// 员工角色
		shopStaffRole := v1.NewShopUserRoleService(sourceDB, targetDB)
		err = shopStaffRole.ConvertShopUserRole()
		if err != nil {
			return err
		}

		// 员工交班记录
		shopUserShiftLog := v1.NewShopUserShiftLogService(sourceDB, targetDB)
		err = shopUserShiftLog.ConvertShopUserShiftLog()
		if err != nil {
			return err
		}

		// 员工交班快照
		shopUserShiftSnapshot := v1.NewShopUserShiftSnapshotService(sourceDB, targetDB)
		err = shopUserShiftSnapshot.ConvertShopUserShiftSnapshot()
		if err != nil {
			return err
		}
	}

	// 打印管理
	{
		// 商品打印标签 PrinterTag
		productPrintLabelService := v1.NewProductPrintLabelService(sourceDB, targetDB)
		err = productPrintLabelService.ConvertProductPrintLabel()
		if err != nil {
			return errors.WithMessage(err)
		}

		// 打印机 Printer
		printerService := v1.NewPrinterService(sourceDB, targetDB)
		err = printerService.ConvertPrinter()
		if err != nil {
			return err
		}

		// 打印机模板 PrinterTemplate
		printerTemplateService := v1.NewPrinterTemplateService(sourceDB, targetDB)
		err = printerTemplateService.ConvertPrinterTemplate()
		if err != nil {
			return err
		}

		// 商品打印 ProductPrinterProductItem ProductPrinterItem ProductPrinterRegion ProductPrinter
		supplierPrintingService := v1.NewSupplierPrintingService(sourceDB, targetDB)
		err = supplierPrintingService.ConvertSupplierPrinting()
		if err != nil {
			return err
		}
	}

	// 必点商品
	orderSchemeService := v1.NewOrderSchemeService(sourceDB, targetDB)
	err = orderSchemeService.ConvertOrderScheme()
	if err != nil {
		return err
	}

	// 会员
	{
		// 会员等级
		userGradeService := v1.NewUserGradeService(sourceDB, targetDB)
		err = userGradeService.ConvertUserGrade()
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
	}

	// 钱箱
	shopAccountService := v1.NewShopAccountService(sourceDB, targetDB)
	err = shopAccountService.ConvertShopAccount()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 月度统计
	erpMonthlyStatisticsService := v1.NewErpMonthlyStatisticsService(sourceDB, targetDB)
	err = erpMonthlyStatisticsService.ConvertERPMonthlyStatistics()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 自助餐顾客类型
	customerTypeService := v1.NewCustomerTypeService(sourceDB, targetDB)
	err = customerTypeService.ConvertCustomerType()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 商品分类
	categoryService := v1.NewCategoryService(sourceDB, targetDB)
	err = categoryService.ConvertCategory()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 退菜原因
	returnReasonService := v1.NewReturnReasonService(sourceDB, targetDB)
	err = returnReasonService.ConvertReturnReason()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 赠菜免单原因
	freeTagService := v1.NewFreeTagService(sourceDB, targetDB)
	err = freeTagService.ConvertFreeTag()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 支付方式
	payTypeService := v1.NewPayTypeService(sourceDB, targetDB)
	err = payTypeService.ConvertPayType()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 商品属性和属性组
	attributeService := v1.NewAttributeService(sourceDB, targetDB)
	err = attributeService.ConvertAttribute()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 商品单位
	productUnitService := v1.NewProductUnitService(sourceDB, targetDB)
	err = productUnitService.ConvertProductUnit()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 规格
	specService := v1.NewSpecService(sourceDB, targetDB)
	err = specService.ConvertSpec()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 商品和材料
	productService := NewProductService(sourceDB, targetDB)
	err = productService.ConvertProduct()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 商品小料库
	productSauceService := NewProductSauceService(sourceDB, targetDB)
	err = productSauceService.ConvertProductSauce()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 供应商
	erpSupplierService := v1.NewErpSupplierService(sourceDB, targetDB)
	err = erpSupplierService.ConvertErpSupplier()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 自助餐加钟
	buffetDelayService := v1.NewBuffetDelayService(sourceDB, targetDB)
	err = buffetDelayService.ConvertBuffetDelay()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 自助餐
	buffetService := NewBuffetService(sourceDB, targetDB)
	err = buffetService.ConvertBuffet()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 税种
	taxCategoryService := v1.NewTaxCategoryService(sourceDB, targetDB)
	err = taxCategoryService.ConvertTaxCategory()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 文件
	fileService := v1.NewUploadFileService(sourceDB, targetDB)
	err = fileService.ConvertUploadFile()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 文件组
	uploadGroupService := v1.NewUploadGroupService(sourceDB, targetDB)
	err = uploadGroupService.ConvertUploadGroup()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 用户充值订单
	userRechargeOrderService := v1.NewUserRechargeOrderService(sourceDB, targetDB)
	err = userRechargeOrderService.ConvertUserRechargeOrder()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 高峰时段
	orderPeakTimeService := v1.NewOrderPeakTimeService(sourceDB, targetDB)
	err = orderPeakTimeService.ConvertOrderPeakTime()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 采购单
	purchaseService := v1.NewPurchaseService(sourceDB, targetDB)
	err = purchaseService.ConvertPurchaseForm()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 入库记录
	err = purchaseService.ConvertWarehouseForm()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 出库记录
	stockService := v1.NewStockService(sourceDB, targetDB)
	err = stockService.ConvertWarehouseOut()
	if err != nil {
		return errors.WithMessage(err)
	}

	// 报损记录
	err = stockService.ConvertDemaged()
	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

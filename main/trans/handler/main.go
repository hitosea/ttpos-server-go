package handler

import (
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

func Run(sourceDB *gorm.DB, targetDB *gorm.DB) error {
	userService := v1.NewUserGradeService(sourceDB, targetDB)
	err := userService.ConvertUserGrade()
	if err != nil {
		return err
	}

	tableAreaService := v1.NewTableAreaService(sourceDB, targetDB)
	err = tableAreaService.ConvertTableArea()
	if err != nil {
		return err
	}

	appService := v1.NewAppService(sourceDB, targetDB)
	err = appService.ConvertApp()
	if err != nil {
		return err
	}

	shopAccountService := v1.NewShopAccountService(sourceDB, targetDB)
	err = shopAccountService.ConvertShopAccount()
	if err != nil {
		return err
	}

	erpMonthlyStatisticsService := v1.NewErpMonthlyStatisticsService(sourceDB, targetDB)
	err = erpMonthlyStatisticsService.ConvertERPMonthlyStatistics()
	if err != nil {
		return err
	}

	customerTypeService := v1.NewCustomerTypeService(sourceDB, targetDB)
	err = customerTypeService.ConvertCustomerType()
	if err != nil {
		return err
	}

	categoryService := v1.NewCategoryService(sourceDB, targetDB)
	err = categoryService.ConvertCategory()
	if err != nil {
		return err
	}

	tableTypeService := v1.NewTableTypeService(sourceDB, targetDB)
	err = tableTypeService.ConvertTableType()
	if err != nil {
		return err
	}

	return nil
}

package repository

import (
	"fmt"
	"testing"
	"time"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
)

func TestCreateCategoryProduct(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "QWERASDFQWE23421",
		RootPassword:  "QWERASDFQWE23421",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop1111000")
	if err != nil {
		panic(err)
	}
	productCategoryRepository := NewProductCategoryRepo(db)

	productCategories := []model.ProductCategory{
		{
			Name:                  "电子产品",
			ParentUuid:            0,
			MultiLanguageNameUuid: 1,
			Status:                true,
			OrderBy:               1,
			CreateTime:            time.Now().Unix(),
			UpdateTime:            time.Now().Unix(),
		},
		{
			Name:                  "家居用品",
			ParentUuid:            0,
			MultiLanguageNameUuid: 2,
			Status:                true,
			OrderBy:               2,
			CreateTime:            time.Now().Unix(),
			UpdateTime:            time.Now().Unix(),
		},
	}

	for _, productCategory := range productCategories {
		_, err := productCategoryRepository.CreateProductCategory(productCategory)
		if err != nil {
			panic(err)
		}
	}
}

func TestCreateCategory(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "QWERASDFQWE23421",
		RootPassword:  "QWERASDFQWE23421",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop1111000")
	if err != nil {
		panic(err)
	}
	categoryRepository := NewCategoryRepositoryService(db)
	productCategories := []req.CreateCategoryRequest{
		{
			Name: req.CategoryTranslation{
				TH:   "สินค้าอุปโภคบริโภค",
				ZH:   "商品",
				ZHTW: "商品",
				EN:   "Product",
			},
			ParentID: 0,
			Sort:     1,
		},
		{
			Name: req.CategoryTranslation{
				TH:   "เครื่องใช้ไฟฟ้า",
				ZH:   "电器",
				ZHTW: "電器",
				EN:   "Appliances",
			},
			ParentID: 0,
			Sort:     2,
		},
		{
			Name: req.CategoryTranslation{
				TH:   "运动器材",
				ZH:   "运动器材",
				ZHTW: "運動器材",
				EN:   "Sports Equipment",
			},
			ParentID: 0,
			Sort:     3,
		},
	}

	for _, category := range productCategories {
		_, err := categoryRepository.CreateCategory(category)
		if err != nil {
			panic(err)
		}
	}
}

func TestGetProductCategoryByIdWithMultiLanguageName(t *testing.T) {
	db, err := NewMySQLConnection(config.DatabaseConf{
		Host:          "localhost",
		Port:          3306,
		User:          "root",
		Password:      "QWERASDFQWE23421",
		RootPassword:  "QWERASDFQWE23421",
		TablePrefix:   "ttpos_",
		SlowQueryTime: 0,
	}, "shop1111000")
	if err != nil {
		panic(err)
	}
	productCategoryRepository := NewProductCategoryRepo(db)

	productCategory, err := productCategoryRepository.GetProductCategoryByIdWithMultiLanguageName(3)
	if err != nil {
		panic(err)
	}
	fmt.Println(fmt.Sprintf("%+v", productCategory))
}

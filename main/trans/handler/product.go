package handler

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	pkgUtils "ttpos-server-go/pkg/utils"
	oldModel "ttpos-server-go/trans/old_model"
	"ttpos-server-go/trans/v2/constant"
	newModel "ttpos-server-go/trans/v2/model"

	"gorm.io/gorm"
)

func testConvertProduct() error {
	InitializeSonyFlakeId()

	db, err := NewMySQLConnection(sourceConf, sourceDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	targetDB, err := NewMySQLConnection(targetConf, targetDBName)
	if err != nil {
		return errors.WithMessage(err, "NewMySQLConnection failed")
	}
	productService := ProductService{db: db, targetDB: targetDB}

	err = productService.ConvertProduct()
	if err != nil {
		return errors.WithMessage(err, "ConvertProduct failed")
	}
	return nil
}

type ProductInterface interface {
	GetProductList() ([]oldModel.Product, error)
	GetProductTax(productID uint, taxType uint) (oldModel.ProductTax, error)
	ConvertProduct() error
}

type ProductService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductService) GetProductList() ([]oldModel.Product, error) {
	var products []oldModel.Product
	err := s.db.Preload("ProductImage").Preload("ProductTax").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductService) GetProductTax(productID uint64, taxType uint) (oldModel.ProductTax, error) {
	var productTax oldModel.ProductTax
	err := s.db.Where("product_id = ? AND product_tax_type = ?", productID, taxType).First(&productTax).Error
	if err != nil {
		return oldModel.ProductTax{}, err
	}
	return productTax, nil
}

func (s *ProductService) ConvertProduct() error {
	products, err := s.GetProductList()
	if err != nil {
		return err
	}
	if err := s.targetDB.Transaction(func(tx *gorm.DB) error {
		fmt.Println(fmt.Sprintf("products: %v", len(products)))
		for index, _ := range products {
			product := &products[index]

			fmt.Println(fmt.Sprintf("index: %d, product: %+v", index, product))
			if product.Type == constant.ProductTypeProduct {
				// 成品
				fmt.Println(fmt.Sprintf("-----迁移成品：%+v", product))
				productPackage, err := newModel.NewProductPackage(product, s.db)
				if err != nil {
					fmt.Println(fmt.Sprintf("err: %v", err))
					return err
				}
				fmt.Println(fmt.Sprintf("productPackage: %+v", pkgUtils.ToJsonString(productPackage)))
				repo := repository.NewProductPackageRepo(tx)
				if err := repo.CreateProductPackage(productPackage); err != nil {
					return err
				}

			} else if product.Type == constant.ProductTypeMaterial {
				// 材料
				fmt.Println(fmt.Sprintf("-----迁移材料：%+v", product))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

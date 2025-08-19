package handler

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	oldModel "ttpos-server-go/trans/v1"
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
	GetProductTax(productID uint64, taxType uint) (oldModel.ProductTax, error)
	ConvertProduct() error
}

func NewProductService(db *gorm.DB, targetDB *gorm.DB) ProductInterface {
	return &ProductService{db: db, targetDB: targetDB}
}

type ProductService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductService) GetProductList() ([]oldModel.Product, error) {
	var products []oldModel.Product
	err := s.db.Preload("ProductImage").Preload("ProductTax").Preload("ProductSKUs").Find(&products).Error
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
		return errors.WithMessage(err)
	}
	if err := s.targetDB.Transaction(func(tx *gorm.DB) error {
		for index, _ := range products {
			product := &products[index]
			// 需要导入已删除的商品，因为入库记录中要显示商品名称
			// if product.IsDelete == 1 {
			// 	continue
			// }

			if product.Type == constant.ProductTypeProduct {
				// 成品
				fmt.Println(fmt.Sprintf("-----迁移成品：%+v", product))
				productPackage, err := newModel.NewProductPackage(product, s.db)
				if err != nil {
					return errors.WithMessage(err)
				}
				repo := repository.NewProductPackageRepo(tx)
				if err := repo.CreateProductPackage(productPackage); err != nil {
					return errors.WithMessage(err)
				}

			} else if product.Type == constant.ProductTypeMaterial {
				// 材料
				fmt.Println(fmt.Sprintf("-----迁移材料：%+v", product))
				languageName, err := newModel.NewMultiLanguageName(product.ProductName)
				if err != nil {
					return err
				}
				// 创建multi_language_name表数据
				multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
				if _, err := multiLanguageNameRepo.CreateMultiLanguageName(*languageName); err != nil {
					return errors.WithMessage(err)
				}

				// 材料的采购单价
				purchasePrice := 0.0
				productSKUs := product.ProductSKUs
				if len(productSKUs) > 0 {
					productSKU := productSKUs[0]
					purchasePrice = productSKU.PurchasePrice
				}
				// 材料的条形码
				barcodeValue := ""
				if len(productSKUs) > 0 {
					productSKU := productSKUs[0]
					barcodeValue = productSKU.Barcode
				}

				material := model.Material{
					BaseModel: model.BaseModel{
						Uuid:       uint64(product.ProductID),
						CreateTime: int64(product.CreateTime),
						UpdateTime: int64(product.UpdateTime),
						DeleteTime: int64(product.IsDelete),
					},
					Name:                  product.ProductName,
					MultiLanguageNameUuid: languageName.Uuid,
					CategoryUuid:          product.CategoryID,
					SupplierUuid:          uint64(product.ErpSupplierID),
					ImageUuid:             product.ProductImage.ImageID,
					ImageName:             product.ImgName,
					UnitUuid:              product.UnitID,
					Price:                 purchasePrice,
					StockNum:              usign(product.ProductMaterialStock),
					BarcodeValue:          barcodeValue,
					Status:                product.ProductStatus == 10,
				}
				if _, err := base.NewMaterialRepo(s.targetDB).CreateMaterial(material); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func usign(num float64) float64 {
	if num < 0 {
		return 0
	}
	return num
}

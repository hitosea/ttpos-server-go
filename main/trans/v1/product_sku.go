package v1

import (
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

// 产品规格材料关联表 `jjjfood_product_sku_material`
type ProductSkuMaterial struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement;comment:规格材料关联表主键"`
	SpecID         int     `gorm:"default:0;comment:规格id"`
	ProductSkuID   int     `gorm:"default:0;comment:产品规格id"`
	MaterialID     int     `gorm:"default:0;comment:材料id"`
	MaterialNum    float64 `gorm:"type:decimal(12,4);default:0.0000;comment:使用库存数量"`
	ShopSupplierID int     `gorm:"default:0;comment:门店id"`
	AppID          int     `gorm:"default:0;comment:应用id"`
	CreateTime     int64   `gorm:"autoCreateTime;comment:创建时间"`
}

// 产品售罄关联表 `jjjfood_product_sold_out`
type ProductSoldOut struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement;comment:Uuid"`
	ProductID      uint64 `gorm:"default:0;comment:产品id"`
	ProductSkuID   uint64 `gorm:"default:0;comment:产品规格id"`
	AppID          uint64 `gorm:"default:0;comment:应用ID"`
	ShopSupplierID uint64 `gorm:"default:0;comment:门店id"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:创建时间"`
}

// 产品规格表 `jjjfood_product_sku`
type ProductSKU struct {
	ProductSkuID  uint64  `gorm:"primaryKey;autoIncrement;comment:产品规格id"`
	ProductID     uint64  `gorm:"not null;default:0;comment:产品id"`
	SpecSkuID     int     `gorm:"default:0;comment:产品sku记录索引 (由规格id组成)"`
	ImageID       uint64  `gorm:"not null;default:0;comment:规格图片id"`
	ProductNo     string  `gorm:"size:100;not null;default:'';comment:产品编码"`
	ProductPrice  float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:产品价格"`
	PurchasePrice float64 `gorm:"type:decimal(12,2);default:0.00;comment:采购单价"`
	Barcode       string  `gorm:"size:255;default:'';comment:商品条码"`
	MaterialStock float64 `gorm:"type:decimal(12,4);default:0.0000;comment:规格材料库存（材料类型专用）"`
	LinePrice     float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:产品划线价"`
	LowPrice      float64 `gorm:"type:decimal(12,2);default:0.00;comment:产品底价"`
	StockNum      uint    `gorm:"not null;default:0;comment:当前库存数量"`
	ProductSales  float64 `gorm:"type:decimal(12,4);default:0.0000;comment:产品规格销量"`
	ProductWeight float64 `gorm:"not null;default:0;comment:产品重量(Kg)"`
	SupplierPrice float64 `gorm:"type:decimal(12,2);default:0.00;comment:供应商价格"`
	SpecName      string  `gorm:"size:2000;not null;default:'';comment:规格名"`
	BagPrice      float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:包装费"`
	CostPrice     float64 `gorm:"type:decimal(12,2);not null;default:0.00;comment:成本价"`
	IsFull        bool    `gorm:"not null;default:0;comment:是否次日置满0否1是"`
	BaseStockNum  uint    `gorm:"not null;default:0;comment:当前库存数量"`
	AppID         uint64  `gorm:"not null;default:0;comment:应用id"`
	CreateTime    int64   `gorm:"autoCreateTime;comment:创建时间"`
	UpdateTime    int64   `gorm:"autoUpdateTime;comment:更新时间"`

	Product             Product               `gorm:"foreignKey:ProductID;references:ProductID"`
	ProductSoldOut      *ProductSoldOut       `gorm:"foreignKey:ProductSkuID;references:ProductSkuID"`
	ProductSkuMaterials []*ProductSkuMaterial `gorm:"foreignKey:ProductSkuID;references:ProductSkuID"`
}

func (model *ProductSKU) GetIsSoldOut() int {
	if model.ProductSoldOut == nil {
		return 0
	}
	return 1
}

type ProductSKURepository interface {
	GetProductSKUService() ([]*ProductSKU, error)
	ConvertProductSKU() error
}

type ProductSKUService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductSKUService) GetProductSKUService() ([]*ProductSKU, error) {
	var productSKUs []*ProductSKU
	if err := s.db.Preload("ProductSoldOut").Preload("Product").Preload("ProductSkuMaterials").Find(&productSKUs).Error; err != nil {
		return nil, err
	}
	return productSKUs, nil
}

func (s *ProductSKUService) ConvertProductSKU() error {
	productSKUs, err := s.GetProductSKUService()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, productSKU := range productSKUs {
		fmt.Println(fmt.Sprintf("productSKU: %+v", productSKU))

		productBom := model.ProductBom{
			BaseModel: model.BaseModel{
				Uuid:       uint64(productSKU.ProductSkuID),
				CreateTime: productSKU.CreateTime,
				UpdateTime: productSKU.UpdateTime,
			},
			PurchasePrice:      productSKU.PurchasePrice,
			Price:              productSKU.ProductPrice,
			Name:               productSKU.SpecName,
			StockNum:           float64(productSKU.StockNum),
			BarcodeValue:       productSKU.Barcode,
			IsDefaultSelect:    0,
			Status:             int(productSKU.Product.GetProductStatus()),
			IsSoldOut:          productSKU.GetIsSoldOut(),
			ProductFlavorUuid:  uint64(productSKU.SpecSkuID),
			ProductSauceUuid:   0,
			ProductPackageUuid: uint64(productSKU.ProductID),
		}
		// fmt.Println(fmt.Sprintf("productBom: %+v", productBom))
		_, err := repository.NewProductBomRepo(s.targetDB).CreateProductBom(productBom)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

package v1

import (
	"encoding/json"
	"fmt"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	pkgUtils "ttpos-server-go/pkg/utils"
	"ttpos-server-go/trans/v2/utils"

	"gorm.io/gorm"
)

type Product struct {
	ProductID            uint64  `gorm:"primaryKey;autoIncrement;comment:'产品id'"`
	ProductName          string  `gorm:"default:'';comment:'产品名称'"`
	ImgName              string  `gorm:"comment:'图片名称（验证唯一性）'"`
	Type                 uint    `gorm:"comment:'类型 10-成品 20-材料'"`
	ErpSupplierID        uint    `gorm:"comment:'供应商id'"`
	ProductPrice         float64 `gorm:"default:0.00;comment:'产品一口价'"`
	LinePrice            float64 `gorm:"default:0.00;comment:'产品划线价'"`
	ProductNo            string  `gorm:"default:'';comment:'产品编码'"`
	ProductStock         int     `gorm:"default:0;comment:'产品总库存'"`
	ProductMaterialStock float64 `gorm:"default:0.00;comment:'材料总库存（材料类型专用）'"`
	SellingPoint         string  `gorm:"default:'';comment:'商品卖点'"`
	CategoryID           uint64  `gorm:"default:0;comment:'产品分类id'"`
	SpecType             uint8   `gorm:"default:0;comment:'产品规格(10单规格 20多规格)'"`
	DeductStockType      uint8   `gorm:"default:20;comment:'库存计算方式(10下单减库存 20付款减库存)'"`
	Content              string  `gorm:"default:'';comment:'产品详情'"`
	SalesInitial         uint    `gorm:"default:0;comment:'初始销量'"`
	SalesActual          float64 `gorm:"default:0.00;comment:'实际销量'"`
	ProductSort          uint    `gorm:"default:100;comment:'产品排序(数字越小越靠前)'"`
	IsPointsGift         uint8   `gorm:"default:0;comment:'是否开启积分赠送(1开启 0关闭)'"`
	IsPointsDiscount     uint8   `gorm:"default:0;comment:'是否允许使用积分抵扣(1允许 0不允许)'"`
	MaxPointsDiscount    uint    `gorm:"default:0;comment:'最大积分抵扣数量'"`
	IsEnableGrade        uint8   `gorm:"default:1;comment:'是否开启会员折扣(1开启 0关闭)'"`
	IsAloneGrade         uint8   `gorm:"default:0;comment:'会员折扣设置(0默认等级折扣 1单独设置折扣)'"`
	AloneGradeEquity     string  `gorm:"default:'';comment:'单独设置折扣的配置'"`
	AloneGradeType       uint8   `gorm:"default:10;comment:'折扣金额类型(10百分比 20固定金额)'"`
	ProductStatus        uint8   `gorm:"default:10;comment:'产品状态(10上架 20下架 )'"`
	ViewTimes            uint    `gorm:"default:0;comment:'访问次数'"`
	ShopSupplierID       uint64  `gorm:"default:0;comment:'供应商id'"`
	SupplierPrice        float64 `gorm:"default:0.00;comment:'供应商价格'"`
	LimitNum             uint    `gorm:"default:0;comment:'限购数量0为不限'"`
	GradeIDs             string  `gorm:"default:'';comment:'可购买会员等级id，逗号隔开'"`
	ProductType          uint8   `gorm:"default:0;comment:'0外卖1店内'"`
	ProductUnit          string  `gorm:"default:'';comment:'商品单位'"`
	ProductAttr          string  `gorm:"default:'';comment:'商品属性'"`
	ProductFeed          string  `gorm:"default:'';comment:'售卖加料'"`
	SpecialID            uint64  `gorm:"default:0;comment:'特殊分类id'"`
	CostPrice            float64 `gorm:"default:0.00;comment:'成本价'"`
	MinBuy               uint    `gorm:"default:1;comment:'最小购买量'"`
	BagPrice             float64 `gorm:"default:0.00;comment:'包装费'"`
	LabelID              uint64  `gorm:"default:0;comment:'打印标签id'"`
	UnitID               uint64  `gorm:"default:0;comment:'单位ID'"`
	IsShowCashier        uint8   `gorm:"default:1;comment:'是否显示在收银端 1-显示 2-不显示'"`
	IsShowTablet         uint8   `gorm:"default:1;comment:'是否显示在平板端 1-显示 2-不显示'"`
	IsShowKitchen        uint8   `gorm:"default:1;comment:'是否显示在送厨端 1-显示 2-不显示'"`
	IsShowAssistant      uint8   `gorm:"default:1;comment:'是否显示在点餐助手 1-显示 2-不显示'"`
	IsShowH5             uint8   `gorm:"default:1;comment:'是否显示在h5 1-显示 2-不显示'"`
	FeedRequired         uint    `gorm:"default:0;comment:'加料是否必填 0-否 1-是'"`
	FeedMaxSelect        uint    `gorm:"default:0;comment:'加料最多可选数量'"`
	FeedOpenMaxSelect    uint    `gorm:"default:0;comment:'加料开启最多可选数量 0-否 1-是'"`
	IsDelete             uint8   `gorm:"default:0;comment:'是否删除'"`
	AppID                uint    `gorm:"default:0;comment:'应用id'"`
	CreateTime           int64   `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime           int64   `gorm:"autoUpdateTime;comment:'更新时间'"`

	// 旧表中没有的字段。用于关联

	ProductImage ProductImage  `gorm:"foreignKey:product_id;references:product_id"`
	ProductTax   ProductTax    `gorm:"foreignKey:product_id;references:product_id"`
	ProductSKUs  []*ProductSKU `gorm:"foreignKey:product_id;references:product_id"`
}

func (model *Product) GetProductStatus() uint8 {
	if model.ProductStatus == 20 {
		return 0
	}
	return 1
}

func (model *Product) GetStockDeductMethod() uint {
	// 10下单减库存 20付款减库存
	if model.DeductStockType == 10 {
		// 1 下单减库存 0 付款减库存
		return 1
	}
	return 0
}

func (model *Product) IsShowCashierValue() uint {
	if model.IsShowCashier == 1 {
		return 1
	}
	return 0
}

func (model *Product) IsShowTabletValue() uint {
	if model.IsShowTablet == 1 {
		return 1
	}
	return 0
}

func (model *Product) IsShowKitchenValue() uint {
	if model.IsShowKitchen == 1 {
		return 1
	}
	return 0
}

func (model *Product) IsShowAssistantValue() uint {
	if model.IsShowAssistant == 1 {
		return 1
	}
	return 0
}

func (model *Product) IsShowH5Value() uint {
	if model.IsShowH5 == 1 {
		return 1
	}
	return 0
}

func (model *Product) OpenDiscount() uint {
	if model.IsEnableGrade == 1 {
		return 1
	}
	return 0
}

// 获取商品的属性组
func (model *Product) GetProductAttributeGroupList(db *gorm.DB, ids []uint64) ([]*Attribute, error) {
	var attributes []*Attribute
	err := db.Where("id IN ?", ids).Find(&attributes).Error
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

// 获取商品的属性组的属性
func (model *Product) GetProductAttributeList(db *gorm.DB, ids []uint64) ([]*Attribute, error) {
	var attributes []*Attribute
	err := db.Where("id IN ?", ids).Find(&attributes).Error
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

type ProductAttrGroup struct {
	ParentID               int      `json:"parent_id"`
	ParentName             string   `json:"parent_name"`
	AttributeName          string   `json:"attribute_name"`
	AttributeValue         []string `json:"attribute_value"`
	DefaultSelect          []int    `json:"default_select"`
	AttributeIDs           []int    `json:"attribute_ids"`
	AttributeMaxSelect     int      `json:"attribute_max_select"`
	AttributeMinSelect     int      `json:"attribute_min_select"`
	AttributeOpenMaxSelect int      `json:"attribute_open_max_select"`
	AttributeRequired      int      `json:"attribute_required"`
	AttributeNameText      string   `json:"attribute_name_text"`
}

func (model *ProductAttrGroup) GetIsMust() uint {
	if model.AttributeRequired == 1 {
		return 1
	}
	return 0
}

func (model *ProductAttrGroup) GetMaxSelection() uint {
	if model.AttributeMaxSelect == 0 {
		return 0
	}
	return uint(model.AttributeMaxSelect)
}

// 解析ProductAttr字段
func (model *Product) ParseProductAttr() ([]*ProductAttrGroup, error) {
	var attributes []*ProductAttrGroup
	fmt.Println(" model.ProductAttr ", model.ProductAttr)
	// model.ProductAttr 为"" 或 []
	if len(model.ProductAttr) < 10 {
		fmt.Println("无属性")
		return attributes, nil
	}
	err := json.Unmarshal([]byte(model.ProductAttr), &attributes)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

// 解析ProductFeed字段
func (model *Product) ParseProductFeed() ([]utils.ProductFeed, error) {
	// model.ProductFeed 为"" 或 []
	if len(model.ProductFeed) < 5 {
		fmt.Println("无加料")
		return nil, nil
	}
	productFeeds, err := utils.ParseFeedJson(model.ProductFeed)
	if err != nil {
		return nil, err
	}
	return productFeeds, nil
}

type ProductImage struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'主键id'"`
	ProductID  uint   `gorm:"default:0;comment:'商品id'"`
	ImageID    uint64 `gorm:"comment:'图片id(关联文件记录表)'"`
	AppID      uint   `gorm:"default:0;comment:'应用id'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间'"`
}

type ProductTax struct {
	ID             uint   `gorm:"primaryKey;autoIncrement;comment:'主键id'"`
	ProductID      uint   `gorm:"default:0;comment:'产品id'"`
	ProductTaxType uint   `gorm:"default:1;comment:'产品税类类型，1-堂食税类，2-外带税类'"`
	TaxCategoryID  uint64 `gorm:"default:0;comment:'税类id'"`
	AppID          uint   `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type ProductInterface interface {
	GetProductList() ([]Product, error)
	GetProductTax(productID uint, taxType uint) (ProductTax, error)
	ConvertProduct() error
}

type ProductService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ProductService) GetProductList() ([]Product, error) {
	var products []Product
	err := s.db.Preload("ProductImage").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductService) GetProductTax(productID uint64, taxType uint) (ProductTax, error) {
	var productTax ProductTax
	err := s.db.Where("product_id = ? AND product_tax_type = ?", productID, taxType).First(&productTax).Error
	if err != nil {
		return ProductTax{}, err
	}
	return productTax, nil
}

func (s *ProductService) ConvertProduct() error {
	products, err := s.GetProductList()
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, product := range products {
		names := Names{}
		err := names.GetNames(product.ProductName)
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("product_id: %d, product_name: %+v", product.ProductID, names))

		id, err := pkgUtils.GetID()
		if err != nil {
			return errors.WithMessage(err)
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)

		if product.Type == 10 {
			// 成品
			fmt.Println(fmt.Sprintf("-----迁移成品：%+v", product))
			StockDeductMethod := product.GetStockDeductMethod()
			Status := product.GetProductStatus()
			IsShowCashier := product.IsShowCashierValue()
			IsShowTablet := product.IsShowTabletValue()
			IsShowKitchen := product.IsShowKitchenValue()
			IsShowAssistant := product.IsShowAssistantValue()
			IsShowH5 := product.IsShowH5Value()
			OpenDiscount := product.OpenDiscount()

			productDineTax, err := s.GetProductTax(product.ProductID, 1)
			if err != nil {
				return errors.WithMessage(err)
			}
			productTakeoutTax, err := s.GetProductTax(product.ProductID, 2)
			if err != nil {
				return errors.WithMessage(err)
			}

			productPackage := model.ProductPackage{
				BaseModel: model.BaseModel{
					Uuid:       uint64(product.ProductID),
					CreateTime: product.CreateTime,
					UpdateTime: product.UpdateTime,
				},
				Name:                  names.Zh,
				MultiLanguageNameUuid: id,
				ImageName:             product.ImgName,
				ImageFileUuid:         product.ProductImage.ImageID,
				DeductStockType:       uint(StockDeductMethod),
				UnitUuid:              product.UnitID,
				DineTaxUuid:           productDineTax.TaxCategoryID,
				CategoryUuid:          product.CategoryID,
				TakeoutTaxUuid:        productTakeoutTax.TaxCategoryID,
				SpecialCategoryUuid:   product.SpecialID,
				PrinterTagUuid:        product.LabelID,
				SupplierUuid:          uint64(product.ErpSupplierID),
				Status:                uint(Status),
				IsShowCashier:         IsShowCashier,
				IsShowTablet:          IsShowTablet,
				IsShowKitchen:         IsShowKitchen,
				IsShowAssistant:       IsShowAssistant,
				IsShowH5:              IsShowH5,
				Sort:                  product.ProductSort,
				LimitNum:              product.LimitNum,
				Describe:              product.SellingPoint,
				OpenDiscount:          OpenDiscount,
				SauceRequired:         uint8(product.FeedRequired),
				SauceMaxSelection:     product.FeedMaxSelect,
				MultiLanguageName:     languageName,
			}
			_, err = base.NewProductPackageRepo(s.targetDB).CreateProductPackage(productPackage)
			if err != nil {
				return errors.WithMessage(err)
			}
		} else if product.Type == 20 {
			// 材料
			fmt.Println(fmt.Sprintf("-----迁移材料：%+v", product))
			material := model.Material{
				BaseModel: model.BaseModel{
					Uuid:       uint64(product.ProductID),
					CreateTime: product.CreateTime,
					UpdateTime: product.UpdateTime,
				},
				Name:                  names.Zh,
				MultiLanguageNameUuid: id,
				CategoryUuid:          product.CategoryID,
				SupplierUuid:          uint64(product.ErpSupplierID),
				ImageUuid:             product.ProductImage.ImageID,
				ImageName:             product.ImgName,
				UnitUuid:              product.UnitID,
				Price:                 product.ProductPrice,
				StockNum:              usign(product.ProductMaterialStock),
				BarcodeValue:          product.ProductNo,
				Status:                product.ProductStatus == 10,
				MultiLanguageName:     languageName,
			}
			_, err := base.NewMaterialRepo(s.targetDB).CreateMaterial(material)
			if err != nil {
				return errors.WithMessage(err)
			}
		}
	}
	return nil
}

func usign(num float64) float64 {
	if num < 0 {
		return 0
	}
	return num
}

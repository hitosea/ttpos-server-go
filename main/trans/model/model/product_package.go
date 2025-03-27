package model

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	pkgUtils "ttpos-server-go/pkg/utils"
	"ttpos-server-go/trans/model/constant"
	"ttpos-server-go/trans/model/utils"
	"ttpos-server-go/trans/old_model"
	"ttpos-server-go/trans/old_model/repository"

	"gorm.io/gorm"
)

// NewProductPackage 创建产品包
func NewProductPackage(product *old_model.Product, db *gorm.DB) (*model.ProductPackage, error) {
	productAttrGroups, err := product.ParseProductAttr()
	if err != nil {
		return nil, errors.WithMessage(err, "ParseProductAttr failed")
	}

	languageName, err := NewMultiLanguageName(product.ProductName)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMultiLanguageName failed")
	}

	StockDeductMethod := product.GetStockDeductMethod()
	Status := product.GetProductStatus()
	IsShowCashier := product.IsShowCashierValue()
	IsShowTablet := product.IsShowTabletValue()
	IsShowKitchen := product.IsShowKitchenValue()
	IsShowAssistant := product.IsShowAssistantValue()
	IsShowH5 := product.IsShowH5Value()
	OpenDiscount := product.OpenDiscount()

	productDineTax, err := repository.NewCommonRepo(db).GetProductTax(uint64(product.ProductID), constant.ProductTaxTypeDine)
	if err != nil {
		return nil, errors.WithMessage(err, "GetProductTax failed")
	}
	productTakeoutTax, err := repository.NewCommonRepo(db).GetProductTax(uint64(product.ProductID), constant.ProductTaxTypeTakeout)
	if err != nil {
		return nil, errors.WithMessage(err, "GetProductTax failed")
	}

	productBoms := make([]model.ProductBom, 0)

	// 创建产品包的规格
	flavorBoms, err := NewFlavorProductBom(db, uint64(product.ProductID))
	if err != nil {
		return nil, errors.WithMessage(err, "NewProductBom failed")
	}

	// 创建产品包的小料
	sauceBoms, err := NewSauceProductBom(db, product)
	if err != nil {
		return nil, errors.WithMessage(err, "NewSauceProductBom failed")
	}
	productBoms = append(productBoms, flavorBoms...)
	productBoms = append(productBoms, sauceBoms...)

	return &model.ProductPackage{
		BaseModel: model.BaseModel{
			Uuid:       uint64(product.ProductID),
			CreateTime: product.CreateTime,
			UpdateTime: product.UpdateTime,
		},
		Name:                          languageName.ZhName,
		MultiLanguageNameUuid:         languageName.Uuid,
		ImageName:                     product.ImgName,
		ImageFileUuid:                 product.ProductImage.ImageID,
		DeductStockType:               StockDeductMethod,
		UnitUuid:                      product.UnitID,
		DineTaxUuid:                   productDineTax.TaxCategoryID,
		CategoryUuid:                  product.CategoryID,
		TakeoutTaxUuid:                productTakeoutTax.TaxCategoryID,
		SpecialCategoryUuid:           product.SpecialID,
		PrinterTagUuid:                product.LabelID,
		SupplierUuid:                  product.ShopSupplierID,
		Status:                        uint(Status),
		IsShowCashier:                 IsShowCashier,
		IsShowTablet:                  IsShowTablet,
		IsShowKitchen:                 IsShowKitchen,
		IsShowAssistant:               IsShowAssistant,
		IsShowH5:                      IsShowH5,
		Sort:                          product.ProductSort,
		LimitNum:                      product.LimitNum,
		Describe:                      product.SellingPoint,
		OpenDiscount:                  OpenDiscount,
		SauceRequired:                 uint8(product.FeedRequired),
		SauceMaxSelection:             product.FeedMaxSelect,
		ProductPackageAttributeGroups: NewProductPackageAttributeGroup(uint64(product.ProductID), productAttrGroups),
		MultiLanguageName:             *languageName,
		ProductBoms:                   productBoms,
	}, nil
}

func NewProductPackageAttributeGroup(productPackageUuid uint64, productAttrGroup []*old_model.ProductAttrGroup) []model.ProductPackageAttributeGroup {
	productPackageAttributeGroups := make([]model.ProductPackageAttributeGroup, 0)
	for _, attrGroup := range productAttrGroup {
		uuid, err := pkgUtils.GetID()
		if err != nil {
			panic(errors.WithMessage(err, "获取uuid失败"))
		}
		productPackageAttributeGroups = append(productPackageAttributeGroups, model.ProductPackageAttributeGroup{
			BaseModel: model.BaseModel{
				Uuid: uuid,
			},
			IsMust:                    attrGroup.GetIsMust(),       // 是否必选
			MaxSelection:              attrGroup.GetMaxSelection(), // 最大选择数量
			ProductPackageUuid:        productPackageUuid,          // 产品包id
			ProductAttributeGroupUuid: uint64(attrGroup.ParentID),  // 产品属性组id
			ProductPackageAttributes:  NewProductPackageAttribute(uuid, *attrGroup),
		})
	}
	return productPackageAttributeGroups
}

func NewProductPackageAttribute(productPackageAttributeGroupUuid uint64, productAttrGroup old_model.ProductAttrGroup) []model.ProductPackageAttribute {
	productPackageAttributes := make([]model.ProductPackageAttribute, 0)
	for index, attrID := range productAttrGroup.AttributeIDs {
		uuid, err := pkgUtils.GetID()
		if err != nil {
			panic(errors.WithMessage(err, "获取uuid失败"))
		}
		productPackageAttributes = append(productPackageAttributes, model.ProductPackageAttribute{
			BaseModel: model.BaseModel{
				Uuid: uuid,
			},
			ProductPackageAttributeGroupUuid: productPackageAttributeGroupUuid,
			AttributeUuid:                    uint64(attrID),
			IsDefaultSelected:                uint(productAttrGroup.DefaultSelect[index]),
		})
	}
	return productPackageAttributes
}

func NewMultiLanguageName(nameJson string) (*model.MultiLanguageName, error) {
	names := utils.Names{}
	if err := names.GetNames(nameJson); err != nil {
		return nil, errors.WithMessage(err, "解析json失败")
	}
	uuid, err := pkgUtils.GetID()
	if err != nil {
		return nil, errors.WithMessage(err, "获取uuid失败")
	}
	multiLanguageName := names.CreateMultiLanguageName(uuid)
	return multiLanguageName, nil
}

func NewFlavorProductBom(db *gorm.DB, productID uint64) ([]model.ProductBom, error) {
	productSKUs, err := repository.NewCommonRepo(db).GetProductSKUs(productID)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品规格失败")
	}
	productBoms := make([]model.ProductBom, 0)
	for _, productSKU := range productSKUs {
		flavorMaterials := make([]*model.RelatedMaterial, 0)
		for _, productSkuMaterial := range productSKU.ProductSkuMaterials {
			flavorMaterials = append(flavorMaterials, &model.RelatedMaterial{
				BaseModel: model.BaseModel{
					Uuid: uint64(productSkuMaterial.ID),
				},
				MaterialUuid:     uint64(productSkuMaterial.MaterialID),
				RelatedUuid:      0,
				ProductBomUuid:   uint64(productSkuMaterial.ProductSkuID),
				ProductSauceUuid: 0,
				Num:              productSkuMaterial.MaterialNum,
			})
		}
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
			FlavorMaterials:    flavorMaterials,
		}
		productBoms = append(productBoms, productBom)
	}
	return productBoms, nil
}

func NewSauceProductBom(db *gorm.DB, product *old_model.Product) ([]model.ProductBom, error) {
	productFeeds, err := product.ParseProductFeed()
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品规格失败")
	}
	productBoms := make([]model.ProductBom, 0)
	for _, productFeed := range productFeeds {
		productBom := model.ProductBom{
			BaseModel: model.BaseModel{
				Uuid: uint64(productFeed.ProductFeedID),
			},
			PurchasePrice:      0,
			Price:              float64(productFeed.Price),
			Name:               productFeed.FeedName,
			StockNum:           float64(productFeed.StockNum),
			BarcodeValue:       "",
			IsDefaultSelect:    productFeed.DefaultSelect,
			Status:             1,
			IsSoldOut:          0,
			ProductFlavorUuid:  0,
			ProductSauceUuid:   uint64(productFeed.FeedID),
			ProductPackageUuid: uint64(product.ProductID),
		}
		productBoms = append(productBoms, productBom)
	}
	return productBoms, nil
}

func NewRelatedMaterial() *model.RelatedMaterial {
	return nil
}

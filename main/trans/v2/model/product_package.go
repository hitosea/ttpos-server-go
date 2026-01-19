package model

import (
	"fmt"
	"strings"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	pkgUtils "ttpos-server-go/pkg/utils"
	v1 "ttpos-server-go/trans/v1"
	"ttpos-server-go/trans/v1/repository"
	"ttpos-server-go/trans/v2/constant"
	"ttpos-server-go/trans/v2/utils"

	"gorm.io/gorm"
)

// NewProductPackage 创建产品包
func NewProductPackage(product *v1.Product, db *gorm.DB) (*model.ProductPackage, error) {
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
		// return nil, errors.WithMessage(err, "GetProductTax failed")
	}
	productTakeoutTax, err := repository.NewCommonRepo(db).GetProductTax(uint64(product.ProductID), constant.ProductTaxTypeTakeout)
	if err != nil {
		// return nil, errors.WithMessage(err, "GetProductTax failed")
	}

	productBoms := make([]model.ProductBom, 0)

	// 创建产品包的规格
	// flavorBoms := make([]model.ProductBom, 0)
	// if product.IsDelete == 1 {
	// 	flavorBoms, err = NewFlavorProductBom(db, uint64(product.ProductID), WithIsDeletedProduct())
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "NewProductBom failed")
	// 	}
	// } else {
	// 	flavorBoms, err = NewFlavorProductBom(db, uint64(product.ProductID))
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "NewProductBom failed")
	// 	}
	// }

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

	productPackageAttributeGroups, err := NewProductPackageAttributeGroup(db, product.ProductAttr, uint64(product.ProductID), productAttrGroups)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProductPackageAttributeGroup failed")
	}
	m := &model.ProductPackage{
		BaseModel: model.BaseModel{
			Uuid:       uint64(product.ProductID),
			CreateTime: product.CreateTime,
			UpdateTime: product.UpdateTime,
			DeleteTime: int64(product.IsDelete),
		},
		Name:                  languageName.ToJson(),
		MultiLanguageNameUuid: languageName.Uuid,
		ImageName:             product.ImgName,
		ImageFileUuid:         product.ProductImage.ImageID,
		DeductStockType:       StockDeductMethod,
		UnitUuid:              product.UnitID,
		DineTaxUuid: func() uint64 {
			if productDineTax != nil {
				return uint64(productDineTax.TaxCategoryID)
			}
			return 0
		}(),
		CategoryUuid: product.CategoryID,
		TakeoutTaxUuid: func() uint64 {
			if productTakeoutTax != nil {
				return uint64(productTakeoutTax.TaxCategoryID)
			}
			return 0
		}(),
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
		ActualSaleNum:                 product.SalesActual,
		OpenDiscount:                  OpenDiscount,
		SauceRequired:                 uint8(product.FeedRequired),
		SauceMaxSelection:             product.FeedMaxSelect,
		ProductPackageAttributeGroups: productPackageAttributeGroups,
		MultiLanguageName:             *languageName,
		ProductBoms:                   productBoms,
	}

	return m, nil
}

func NewProductPackageAttributeGroup(db *gorm.DB, productAttrJson string, productPackageUuid uint64, productAttrGroup []*v1.ProductAttrGroup) ([]model.ProductPackageAttributeGroup, error) {
	productPackageAttributeGroups := make([]model.ProductPackageAttributeGroup, 0)
	isOldVersion := false
	if len(productAttrJson) > 10 && !strings.Contains(productAttrJson, "attribute_ids") { // 保证只有这个特殊版本的数据才执行这个代码分支
		for _, attrGroup := range productAttrGroup {
			if len(attrGroup.AttributeIDs) == 0 {
				isOldVersion = true
			}
		}
	}

	// 兼容版本，如果productAttrGroup.AttributeIDs为空，则从jjjfood_product_attribute和jjjfood_product_attribute_group表查询数据
	if isOldVersion {
		// 通过商品id查询商品属性组
		productAttributeGroups, err := repository.NewCommonRepo(db).GetProductAttributeGroupListByProductID(productPackageUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "获取商品属性组失败")
		}

		productAttrGroupList := make([]*v1.ProductAttrGroup, 0)
		for _, attrGroup := range productAttributeGroups {
			attributeIDs := make([]int, 0)
			defaultSelect := make([]int, 0)
			for _, attr := range attrGroup.ProductAttributes {
				if attr.ProductID != uint(productPackageUuid) {
					continue
				}
				attributeIDs = append(attributeIDs, int(attr.AttributeID))
				defaultSelect = append(defaultSelect, int(attr.DefaultSelect))
			}

			productAttrGroupList = append(productAttrGroupList, &v1.ProductAttrGroup{
				ParentID:               int(attrGroup.AttributeID),
				DefaultSelect:          defaultSelect,
				AttributeIDs:           attributeIDs,
				AttributeMaxSelect:     int(attrGroup.AttributeMaxSelect),
				AttributeMinSelect:     int(attrGroup.AttributeMinSelect),
				AttributeOpenMaxSelect: int(attrGroup.AttributeOpenMaxSelect),
				AttributeRequired:      int(attrGroup.AttributeRequired),
			})
		}

		// 用新转换的数据覆盖就数据
		productAttrGroup = productAttrGroupList
	}

	for _, attrGroup := range productAttrGroup {
		uuid, err := pkgUtils.GetID()
		if err != nil {
			return nil, errors.WithMessage(err, "获取uuid失败")
		}
		productPackageAttributes, err := NewProductPackageAttribute(uuid, *attrGroup)
		if err != nil {
			return nil, errors.WithMessage(err, "NewProductPackageAttribute failed")
		}
		productPackageAttributeGroups = append(productPackageAttributeGroups, model.ProductPackageAttributeGroup{
			BaseModel: model.BaseModel{
				Uuid: uuid,
			},
			IsMust:                    attrGroup.GetIsMust(),       // 是否必选
			MaxSelection:              attrGroup.GetMaxSelection(), // 最大选择数量
			ProductPackageUuid:        productPackageUuid,          // 产品包id
			ProductAttributeGroupUuid: uint64(attrGroup.ParentID),  // 产品属性组id
			ProductPackageAttributes:  productPackageAttributes,
		})
	}
	return productPackageAttributeGroups, nil
}

func NewProductPackageAttribute(productPackageAttributeGroupUuid uint64, productAttrGroup v1.ProductAttrGroup) ([]model.ProductPackageAttribute, error) {
	productPackageAttributes := make([]model.ProductPackageAttribute, 0)
	for index, attrID := range productAttrGroup.AttributeIDs {
		uuid, err := pkgUtils.GetID()
		if err != nil {
			return nil, errors.WithMessage(err, "获取uuid失败")
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
	return productPackageAttributes, nil
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

// type NewFlavorProductBomOption struct {
// 	IsDeletedProduct bool // 是否是已经删除的商品
// }

// func WithIsDeletedProduct() func(*NewFlavorProductBomOption) {
// 	return func(option *NewFlavorProductBomOption) {
// 		option.IsDeletedProduct = true
// 	}
// }

func NewFlavorProductBom(db *gorm.DB, productID uint64) ([]model.ProductBom, error) {
	// func NewFlavorProductBom(db *gorm.DB, productID uint64, options ...func(*NewFlavorProductBomOption)) ([]model.ProductBom, error) {
	// option := &NewFlavorProductBomOption{}
	// for _, opt := range options {
	// 	opt(option)
	// }

	// // 如果商品已经删除，则返回一个空的productBom
	// if option.IsDeletedProduct {
	// 	// 获取商品规格.随便给个这个空的规格设置一个specId
	// 	specs, err := repository.NewCommonRepo(db).GetProductSpecList()
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "获取商品规格失败")
	// 	}
	// 	if len(specs) == 0 {
	// 		return nil, errors.WithMessage(err, "商品规格为空")
	// 	}
	// 	specId := specs[0].SpecID

	// 	productBomUuid, err := pkgUtils.GetID()
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "获取uuid失败")
	// 	}
	// 	return []model.ProductBom{
	// 		{
	// 			BaseModel: model.BaseModel{
	// 				Uuid:       productBomUuid,
	// 				CreateTime: time.Now().Unix(),
	// 				UpdateTime: time.Now().Unix(),
	// 			},
	// 			ProductPackageUuid: productID,
	// 			ProductFlavorUuid:  uint64(specId),
	// 		},
	// 	}, nil
	// }

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
					Uuid:       uint64(productSkuMaterial.ID),
					CreateTime: productSkuMaterial.CreateTime,
					UpdateTime: productSkuMaterial.CreateTime,
				},
				MaterialUuid: uint64(productSkuMaterial.MaterialID),
				RelatedUuid:  uint64(productSkuMaterial.ProductSkuID),
				Num:          productSkuMaterial.MaterialNum,
			})
		}
		productBom := model.ProductBom{
			BaseModel: model.BaseModel{
				Uuid:       uint64(productSKU.ProductSkuID),
				CreateTime: productSKU.CreateTime,
				UpdateTime: productSKU.UpdateTime,
			},
			PurchasePrice:   productSKU.PurchasePrice,
			Price:           productSKU.ProductPrice,
			Name:            productSKU.SpecName,
			StockNum:        float64(productSKU.StockNum),
			BarcodeValue:    productSKU.Barcode,
			IsDefaultSelect: 0,
			Status:          int(productSKU.Product.GetProductStatus()),
			IsSoldOut:       productSKU.GetIsSoldOut(),
			ActualSaleNum:   productSKU.ProductSales,
			ProductFlavorUuid: func() uint64 {
				if productSKU.SpecSkuID > 0 {
					return uint64(productSKU.SpecSkuID)
				}
				// 兼容v1.0版本，没有spec_sku_id字段也能点餐
				return 1971200000000000000
			}(),
			ProductSauceUuid:   0,
			ProductPackageUuid: uint64(productSKU.ProductID),
			FlavorMaterials:    flavorMaterials,
		}
		productBoms = append(productBoms, productBom)
	}
	return productBoms, nil
}

func NewSauceProductBom(db *gorm.DB, product *v1.Product) ([]model.ProductBom, error) {
	productFeeds, err := product.ParseProductFeed()
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品规格失败")
	}
	productBoms := make([]model.ProductBom, 0)
	for _, productFeed := range productFeeds {
		price, err := productFeed.GetPrice()
		if err != nil {
			return nil, errors.WithMessage(err, "productFeed.GetPrice() failed")
		}
		defaultSelect, err := productFeed.GetDefaultSelect()
		if err != nil {
			return nil, errors.WithMessage(err, "productFeed.GetDefaultSelect() failed")
		}

		// 兼容版本，如果productFeed.FeedName中不包含"feed_id"，则从jjjfood_product_feed表中取
		if !strings.Contains(productFeed.FeedName, "\"feed_id\"") && productFeed.FeedID == 0 {
			fmt.Println("兼容旧版本-------- productFeed.FeedName", productFeed.FeedName)
			feed, err := repository.NewCommonRepo(db).GetProductFeedIdByProductFeedID(productFeed.ProductFeedID)
			if err != nil {
				return nil, errors.WithMessage(err, "GetFeedIdByProductFeedID failed")
			}
			productFeed.FeedID = feed.FeedID
			productFeed.Price = feed.Price
			productFeed.StockNum = feed.StockNum
			productFeed.DefaultSelect = feed.DefaultSelect
		}

		// 生成新的雪花uuid
		feedUuid, err := NewFeedUuid(productFeed.FeedID)
		if err != nil {
			return nil, errors.WithMessage(err, "获取uuid失败")
		}

		productBomUuid, err := pkgUtils.GetID()
		if err != nil {
			return nil, errors.WithMessage(err, "获取uuid失败")
		}

		productBom := model.ProductBom{
			BaseModel: model.BaseModel{
				//Uuid: uint64(productFeed.ProductFeedID),
				Uuid: productBomUuid,
			},
			PurchasePrice:     0,
			Price:             price,
			Name:              productFeed.FeedName,
			StockNum:          float64(productFeed.StockNum),
			BarcodeValue:      "",
			IsDefaultSelect:   defaultSelect,
			Status:            1,
			IsSoldOut:         0,
			ProductFlavorUuid: 0,
			// ProductSauceUuid:   uint64(productFeed.FeedID),
			ProductSauceUuid:   feedUuid,
			ProductPackageUuid: uint64(product.ProductID),
		}
		productBoms = append(productBoms, productBom)
	}
	return productBoms, nil
}

func NewRelatedMaterial() *model.RelatedMaterial {
	return nil
}

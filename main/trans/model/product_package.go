package model

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/trans/old_model"
)

func NewProductPackage(product *old_model.Product) *model.ProductPackage {
	// productPackage := model.ProductPackage{
	// 	BaseModel: model.BaseModel{
	// 		Uuid:       uint64(product.ProductID),
	// 		CreateTime: product.CreateTime,
	// 		UpdateTime: product.UpdateTime,
	// 	},
	// 	Name:                  names.Zh,
	// 	MultiLanguageNameUuid: id,
	// 	ImageName:             product.ImgName,
	// 	ImageFileUuid:         product.ProductImage.ImageID,
	// 	DeductStockType:       uint(StockDeductMethod),
	// 	UnitUuid:              product.UnitID,
	// 	DineTaxUuid:           productDineTax.TaxCategoryID,
	// 	CategoryUuid:          product.CategoryID,
	// 	TakeoutTaxUuid:        productTakeoutTax.TaxCategoryID,
	// 	SpecialCategoryUuid:   product.SpecialID,
	// 	PrinterTagUuid:        product.LabelID,
	// 	SupplierUuid:          product.ShopSupplierID,
	// 	Status:                uint(Status),
	// 	IsShowCashier:         IsShowCashier,
	// 	IsShowTablet:          IsShowTablet,
	// 	IsShowKitchen:         IsShowKitchen,
	// 	IsShowAssistant:       IsShowAssistant,
	// 	IsShowH5:              IsShowH5,
	// 	Sort:                  product.ProductSort,
	// 	LimitNum:              product.LimitNum,
	// 	Describe:              product.SellingPoint,
	// 	OpenDiscount:          OpenDiscount,
	// 	SauceRequired:         uint8(product.FeedRequired),
	// 	SauceMaxSelection:     product.FeedMaxSelect,
	// 	MultiLanguageName:     languageName,
	// }

	return nil
}

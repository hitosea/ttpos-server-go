package product_resp

import (
	"ttpos-server-go/app/dto"
)

// ProductTakeoutShopDetailResp 外卖商品详情响应
type ProductTakeoutShopDetailResp struct {
	Uuid                uint64                         `json:"uuid"`                  // 外卖商品UUID
	ProductPackageUuid  uint64                         `json:"product_package_uuid"`  // 商品包UUID
	ProductType         uint                           `json:"product_type"`          // 商品类型 0-商品 1-套餐
	TakeoutType         int                            `json:"takeout_type"`          // 外卖类型 1-Grab 2-FoodPanda 3-其他
	LocaleName          dto.LocaleResponse             `json:"locale_name"`           // 商品名称（多语言）
	CategoryUuid        uint64                         `json:"category_uuid"`         // 外卖分类UUID
	CategoryName        dto.LocaleResponse             `json:"category_name"`         // 外卖分类名称
	SpecialCategoryUuid uint64                         `json:"special_category_uuid"` // 外卖特色分类UUID
	SpecialCategoryName dto.LocaleResponse             `json:"special_category_name"` // 外卖特色分类名称
	Status              int                            `json:"status"`                // 外卖状态 0-下架 1-上架
	ImageFileUuid       uint64                         `json:"image_file_uuid"`       // 外卖商品图片UUID
	ImageUrl            string                         `json:"image_url"`             // 外卖商品图片URL
	HeadquarterUuid     uint64                         `json:"headquarter_uuid"`      // 总部UUID,0表示不是总部商品
	Flavors             []ProductTakeoutShopFlavorResp `json:"flavors"`               // 外卖规格列表

	// 关联原商品的套餐信息
	PackageSubProductGroups ProductPackageSubProductGroupList `json:"package_sub_product_groups"` // 套餐子商品分组列表
}

// ProductTakeoutShopFlavorResp 外卖商品规格响应
type ProductTakeoutShopFlavorResp struct {
	BomUuid    uint64             `json:"bom_uuid"`    // 商品BOM UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 规格名称（多语言）
	Price      float64            `json:"price"`       // 外卖规格价格
}

// ProductTakeoutShopListResp 外卖商品列表响应
type ProductTakeoutShopListResp struct {
	Uuid               uint64             `json:"uuid"`                 // 外卖商品UUID
	ProductPackageUuid uint64             `json:"product_package_uuid"` // 商品包UUID
	TakeoutType        int                `json:"takeout_type"`         // 外卖类型 1-Grab 2-FoodPanda 3-其他
	LocaleName         dto.LocaleResponse `json:"locale_name"`          // 商品名称（多语言）
	CategoryName       dto.LocaleResponse `json:"category_name"`        // 外卖分类名称
	Status             int                `json:"status"`               // 外卖状态
	ImageUrl           string             `json:"image_url"`            // 外卖商品图片URL
	MinPrice           float64            `json:"min_price"`            // 最低价格
}

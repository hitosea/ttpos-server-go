package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// ProductTakeoutShopAddReq 外卖商品添加请求
type ProductTakeoutShopAddReq struct {
	ProductPackageUuid  uint64                           `json:"product_package_uuid" binding:"required"` // 商品包UUID（关联店内商品）
	TakeoutType         int                              `json:"takeout_type"`                            // 外卖类型 1-Grab 2-FoodPanda 3-其他，默认1
	LocaleName          dto.LocaleResponse               `json:"locale_name"`                             // 外卖商品名称（多语言），不填则使用店内商品名称
	CategoryUuid        uint64                           `json:"category_uuid"`                           // 外卖分类UUID
	SpecialCategoryUuid uint64                           `json:"special_category_uuid"`                   // 外卖特色分类UUID
	Flavors             []ProductTakeoutShopAddFlavorReq `json:"flavors"`                                 // 外卖规格列表（价格）
	Status              int                              `json:"status"`                                  // 外卖状态 0-下架 1-上架
	ImageFileUuid       uint64                           `json:"image_file_uuid"`                         // 外卖商品图片文件UUID
}

// ProductTakeoutShopAddFlavorReq 外卖商品规格添加请求
type ProductTakeoutShopAddFlavorReq struct {
	BomUuid uint64  `json:"bom_uuid" binding:"required"` // 商品BOM UUID（关联店内规格）
	Price   float64 `json:"price"`                       // 外卖规格价格
}

// ProductTakeoutShopEditReq 外卖商品编辑请求
type ProductTakeoutShopEditReq struct {
	Uuid                uint64                            `json:"uuid" binding:"required"` // 外卖商品UUID
	LocaleName          dto.LocaleResponse                `json:"locale_name"`             // 外卖商品名称（多语言）
	CategoryUuid        uint64                            `json:"category_uuid"`           // 外卖分类UUID
	SpecialCategoryUuid uint64                            `json:"special_category_uuid"`   // 外卖特色分类UUID
	Flavors             []ProductTakeoutShopEditFlavorReq `json:"flavors"`                 // 外卖规格列表（价格）
	Status              int                               `json:"status"`                  // 外卖状态 0-下架 1-上架
	ImageFileUuid       uint64                            `json:"image_file_uuid"`         // 外卖商品图片文件UUID
}

// Validate 验证外卖商品编辑请求
func (r *ProductTakeoutShopEditReq) Validate() error {
	if r.Uuid == 0 {
		return errors.WithMessage(errors.New("外卖商品UUID不能为空"))
	}
	return nil
}

// ProductTakeoutShopEditFlavorReq 外卖商品规格编辑请求
type ProductTakeoutShopEditFlavorReq struct {
	BomUuid uint64  `json:"bom_uuid" binding:"required"` // 商品BOM UUID
	Price   float64 `json:"price"`                       // 外卖规格价格
}

// ProductTakeoutShopDetailReq 外卖商品详情请求
type ProductTakeoutShopDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 外卖商品UUID
}

// ProductTakeoutShopListReq 外卖商品列表请求
type ProductTakeoutShopListReq struct {
	dto.PageReq
	ProductPackageUuid uint64 `form:"product_package_uuid" json:"product_package_uuid"` // 商品包UUID筛选
	TakeoutType        *int   `form:"takeout_type" json:"takeout_type"`                 // 外卖类型筛选
	Status             *int   `form:"status" json:"status"`                             // 外卖状态筛选
}

// ProductTakeoutShopDeleteReq 外卖商品删除请求
type ProductTakeoutShopDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 外卖商品UUID
}

// ProductTakeoutShopStatusReq 外卖商品状态修改请求
type ProductTakeoutShopStatusReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"`             // 外卖商品UUID
	Status *int   `json:"status" binding:"required,oneof=0 1"` // 外卖状态 0-下架 1-上架
}

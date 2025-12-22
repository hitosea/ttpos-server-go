package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// ProductTakeoutShopAddReq 外卖商品添加请求
type ProductTakeoutShopAddReq struct {
	ProductPackageUuid  uint64                                     `json:"product_package_uuid" binding:"required"` // 商品包UUID（关联店内商品）
	TakeoutType         int                                        `json:"takeout_type"`                            // 外卖类型 1-Grab 2-Lineman 3-其他，默认1
	LocaleName          dto.LocaleResponse                         `json:"locale_name"`                             // 外卖商品名称（多语言），不填则使用店内商品名称
	Describe            dto.LocaleResponse                         `json:"describe"`                                // 卖点描述（多语言），不填则使用店内商品卖点
	CategoryUuid        uint64                                     `json:"category_uuid"`                           // 外卖分类UUID
	SpecialCategoryUuid uint64                                     `json:"special_category_uuid"`                   // 外卖特色分类UUID
	Flavors             []ProductTakeoutShopAddFlavorReq           `json:"flavors"`                                 // 外卖规格列表（价格）
	Attributes          []ProductTakeoutShopAddAttributeReq        `json:"attributes"`                              // 外卖属性列表（价格）
	Price               float64                                    `json:"price"`                                   // 外卖商品价格 (套餐价格)
	PackageGroupItems   []ProductTakeoutShopAddPackageGroupItemReq `json:"package_group_items"`                     // 外卖套餐子商品列表（加价）
	Status              int                                        `json:"status"`                                  // 外卖状态 0-下架 1-上架
	ImageFileUuid       uint64                                     `json:"image_file_uuid"`                         // 外卖商品图片文件UUID
	Source              string                                     `json:"source"`                                  // 来源平台 (grab/lineman等 默认grab)
	SourceProductId     string                                     `json:"source_product_id"`                       // 来源平台商品ID (grab/lineman等 默认grab)
}

// ProductTakeoutShopAddFlavorReq 外卖商品规格添加请求
type ProductTakeoutShopAddFlavorReq struct {
	BomUuid        uint64  `json:"bom_uuid" binding:"required"` // 商品BOM UUID（关联店内规格）
	Price          float64 `json:"price"`                       // 外卖规格价格
	GrabModifierId string  `json:"grab_modifier_id"`            // Grab修饰符ID - 可选
}

// ProductTakeoutShopAddAttributeReq 外卖商品属性添加请求
type ProductTakeoutShopAddAttributeReq struct {
	ProductPackageAttributeUuid uint64  `json:"product_package_attribute_uuid" binding:"required"` // 商品属性 UUID（关联店内属性）
	Price                       float64 `json:"price"`                                             // 外卖属性价格
}

// ProductTakeoutShopAddPackageGroupItemReq 外卖套餐子商品添加请求
type ProductTakeoutShopAddPackageGroupItemReq struct {
	ProductPackageGroupItemUuid uint64  `json:"product_package_group_item_uuid" binding:"required"` // 套餐子商品 UUID（关联店内套餐子商品）
	AddPrice                    float64 `json:"add_price"`                                          // 外卖平台的加价金额（覆盖店内加价）
}

// ProductTakeoutShopEditReq 外卖商品编辑请求
type ProductTakeoutShopEditReq struct {
	Uuid                uint64                                      `json:"uuid" binding:"required"` // 外卖商品UUID
	LocaleName          dto.LocaleResponse                          `json:"locale_name"`             // 外卖商品名称（多语言）
	Describe            dto.LocaleResponse                          `json:"describe"`                // 卖点描述（多语言）
	CategoryUuid        uint64                                      `json:"category_uuid"`           // 外卖分类UUID
	SpecialCategoryUuid uint64                                      `json:"special_category_uuid"`   // 外卖特色分类UUID
	Flavors             []ProductTakeoutShopEditFlavorReq           `json:"flavors"`                 // 外卖规格列表（价格）
	Attributes          []ProductTakeoutShopEditAttributeReq        `json:"attributes"`              // 外卖属性列表（价格）
	Price               float64                                     `json:"price"`                   // 外卖商品价格 (套餐价格)
	PackageGroupItems   []ProductTakeoutShopEditPackageGroupItemReq `json:"package_group_items"`     // 外卖套餐子商品列表（加价）
	Status              int                                         `json:"status"`                  // 外卖状态 0-下架 1-上架
	ImageFileUuid       uint64                                      `json:"image_file_uuid"`         // 外卖商品图片文件UUID
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

// ProductTakeoutShopEditAttributeReq 外卖商品属性编辑请求
type ProductTakeoutShopEditAttributeReq struct {
	ProductPackageAttributeUuid uint64  `json:"product_package_attribute_uuid" binding:"required"` // 商品属性 UUID
	Price                       float64 `json:"price"`                                             // 外卖属性价格
}

// ProductTakeoutShopEditPackageGroupItemReq 外卖套餐子商品编辑请求
type ProductTakeoutShopEditPackageGroupItemReq struct {
	ProductPackageGroupItemUuid uint64  `json:"product_package_group_item_uuid" binding:"required"` // 套餐子商品 UUID
	AddPrice                    float64 `json:"add_price"`                                          // 外卖平台的加价金额
}

// ProductTakeoutShopDetailReq 外卖商品详情请求
type ProductTakeoutShopDetailReq struct {
	Uuid     uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品UUID
	Platform string `form:"platform" json:"platform"`            // 来源平台 grab/lineman等
}

// ProductTakeoutShopDeleteReq 外卖商品删除请求
type ProductTakeoutShopDeleteReq struct {
	Uuid     uint64 `json:"uuid" binding:"required"`                        // 商品UUID
	Platform string `json:"platform" binding:"required,oneof=grab lineman"` // 来源平台 grab/lineman
}

// ProductTakeoutShopStatusReq 外卖商品状态修改请求
type ProductTakeoutShopStatusReq struct {
	Uuid     uint64 `json:"uuid" binding:"required"`                        // 商品UUID
	Platform string `json:"platform" binding:"required,oneof=grab lineman"` // 来源平台 grab/lineman
	Status   *int   `json:"status" binding:"required,oneof=0 1"`            // 外卖状态 0-下架 1-上架
}

package req

import (
	"ttpos-server-go/app/dto"
)

// ProductListReq 商品列表查询
type ProductListReq struct {
	dto.PageReq // 分页参数
	// 以下字段不参与json序列化,内部方法使用
	RecommendProductPackageUuids []uint64 `json:"-"` // 推荐商品uuid列表
	IsMember                     bool     `json:"-"` // 是否是会员端查询商品列表
}

func (p *ProductListReq) ToPageReq() dto.PageReq {
	return dto.PageReq{
		PageNo:   p.PageNo,
		PageSize: p.PageSize,
	}
}

// ProductRecommendListReq 商品推荐列表查询
type ProductRecommendListReq struct {
}

// ProductSearchReq 商品搜索查询
type ProductSearchReq struct {
	Keyword  string `form:"keyword" json:"keyword" binding:"required"` // 搜索关键词
	IsMember bool   `json:"-"`                                         // 是否是会员端查询商品列表
}

// ProductUnitListReq 商品单位列表查询
type ProductUnitListReq struct {
	dto.PageReq // 分页参数
}

type ProductSauceListReq struct {
	dto.PageReq // 分页参数
}

type ProductSauceReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品加料UUID
}

type ProductUnitReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品单位UUID
}

type ProductUnitAddReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"` // 关联商品包UUID列表
}

type ProductUnitEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"` // 商品单位UUID
	LocaleName          dto.LocaleResponse `json:"locale_name"`             // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"`   // 关联商品包UUID列表
}

type ProductUnitSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品单位UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductUnitSortReq struct {
	List []ProductUnitSortItem `json:"list" binding:"required,dive"` // 商品单位排序列表
}

type ProductSauceAddReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品加料名称
	Price               float64            `json:"price"`                 // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"` // 关联商品包UUID列表
}

type ProductSauceEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"` // 商品加料UUID
	LocaleName          dto.LocaleResponse `json:"locale_name"`             // 商品加料名称
	Price               float64            `json:"price"`                   // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"`   // 关联商品包UUID列表
}

type ProductSauceSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品加料UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductSauceSortReq struct {
	List []ProductSauceSortItem `json:"list" binding:"required,dive"` // 商品加料排序列表
}

// ProductShopCategoryListReq 商品分类列表查询
type ProductShopCategoryListReq struct {
	Keyword    *string `form:"keyword"`     // 搜索关键词, 可选
	ParentUuid *uint64 `form:"parent_uuid"` // 父级分类UUID, 可选
	IsSpecial  *bool   `form:"is_special"`  // 是否特色分类, false-否 true-是, 可选
}

// ProductShopCategorySortItem 商品分类排序项
type ProductShopCategorySortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品分类UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

// ProductShopCategorySortReq 商品分类排序请求
type ProductShopCategorySortReq struct {
	List []ProductShopCategorySortItem `json:"list" binding:"required,dive"` // 商品分类排序列表
}

// ProductShopCategoryAddReq 商品分类添加请求
type ProductShopCategoryAddReq struct {
	IsSpecial  bool               `json:"is_special"`                     // 是否特殊分类, false-否 true-是
	ParentUuid uint64             `json:"parent_uuid"`                    // 父级分类UUID, 一级分类为0, 二级分类为一级分类的uuid
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品分类名称, 多语言
	Status     int                `json:"status"`                         // 商品分类状态 0-关闭 1-开启
}

// ProductShopCategoryEditReq 商品分类编辑请求
type ProductShopCategoryEditReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`        // 商品分类UUID
	ParentUuid uint64             `json:"parent_uuid"`                    // 父级分类UUID, 一级分类为0, 二级分类为一级分类的uuid
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品分类名称, 多语言
	Status     int                `json:"status"`                         // 商品分类状态 0-关闭 1-开启
}

// ProductShopCategoryReq 商品分类请求
type ProductShopCategoryReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品分类UUID
}

type ProductAttributeGroupListReq struct {
	dto.PageReq // 分页参数
}

type ProductAttributeGroupReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品属性分组UUID
}

type ProductAttributeGroupAddReq struct {
	LocaleName        dto.LocaleResponse                            `json:"locale_name" binding:"required"`             // 商品属性分组名称, 多语言
	ProductAttributes []ProductAttributeGroupAddProductAttributeReq `json:"product_attributes" binding:"required,dive"` // 商品属性
}

type ProductAttributeGroupAddProductAttributeReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品属性名称, 多语言
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductAttributeGroupEditReq struct {
	Uuid              uint64                                         `json:"uuid" binding:"required"`                    // 商品属性分组UUID
	LocaleName        dto.LocaleResponse                             `json:"locale_name" binding:"required"`             // 商品属性分组名称, 多语言
	ProductAttributes []ProductAttributeGroupEditProductAttributeReq `json:"product_attributes" binding:"required,dive"` // 商品属性
}

type ProductAttributeGroupEditProductAttributeReq struct {
	Uuid                uint64             `json:"uuid"`                           // 商品属性UUID, 可选，如果有，是编辑，没有是添加
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品属性名称, 多语言
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductAttributeGroupSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品属性分组UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductAttributeGroupSortReq struct {
	List []ProductAttributeGroupSortItem `json:"list" binding:"required,dive"` // 商品属性分组排序列表
}

type ProductAttributeSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品属性UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductAttributeSortReq struct {
	ProductAttributeGroupUuid uint64                     `json:"product_attribute_group_uuid" binding:"required"` // 商品属性分组UUID
	List                      []ProductAttributeSortItem `json:"list" binding:"required,dive"`                    // 商品属性排序列表
}

// ProductFlavorListReq 商品规格列表查询
type ProductFlavorListReq struct {
	dto.PageReq        // 分页参数
	Keyword     string `form:"keyword" json:"keyword"` // 搜索关键词
}

// ProductFlavorReq 商品规格请求
type ProductFlavorReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品规格UUID
}

// ProductFlavorAddReq 商品规格添加请求
type ProductFlavorAddReq struct {
	LocaleName dto.LocaleResponse               `json:"locale_name" binding:"required"` // 商品规格名称
	List       []ProductFlavorAddProductPackage `json:"list"`                           // 关联商品包列表
}

// ProductFlavorAddProductPackage 商品规格添加商品包
type ProductFlavorAddProductPackage struct {
	Uuid  uint64  `json:"uuid" binding:"required"`  // 商品包UUID
	Price float64 `json:"price" binding:"required"` // 商品包价格
}

// ProductFlavorDeleteReq 商品规格删除请求
type ProductFlavorDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品规格UUID
}

// ProductFlavorSortItem 商品规格排序项
type ProductFlavorSortItemReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品规格UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

// ProductFlavorSortReq 商品规格排序请求
type ProductFlavorSortReq struct {
	List []ProductFlavorSortItemReq `json:"list" binding:"required,dive"` // 商品规格排序列表
}

// ProdudctFlavorEditReq 商品规格编辑请求
type ProdudctFlavorEditReq struct {
	Uuid       uint64                               `json:"uuid" binding:"required"`        // 商品规格UUID
	LocaleName dto.LocaleResponse                   `json:"locale_name" binding:"required"` // 商品规格名称
	List       []ProductFlavorEditProductPackageReq `json:"list"`                           // 关联商品包列表
}

// ProductFlavorEditProductPackageReq 商品规格编辑商品包请求
type ProductFlavorEditProductPackageReq struct {
	Uuid     uint64  `json:"uuid" binding:"required"`  // 商品UUID
	BomUuid  uint64  `json:"bom_uuid"`                 // 商品BOM UUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	Price    float64 `json:"price" binding:"required"` // 商品价格
	IsDelete bool    `json:"is_delete"`                // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductImportListReq 导入商品列表请求
type ProductImportListReq struct {
	List []ProductImportListItemReq `json:"list" binding:"required,dive"` // 商品列表
}

// ProductImportItemReq 导入商品项请求
type ProductImportListItemReq struct {
	ProductName           string  `json:"product_name" binding:"required"`             // 商品名称
	CategoryName          string  `json:"category_name" binding:"required"`            // 分类名称
	ProductUnit           string  `json:"product_unit" binding:"required"`             // 商品单位
	SkuName               string  `json:"sku_name" binding:"required"`                 // 规格名称
	ProductPrice          float64 `json:"product_price" binding:"required"`            // 商品价格
	NumType               int     `json:"num_type" binding:"required"`                 // 数量计算方法, 1-整数 2-小数
	Barcode               string  `json:"barcode" binding:"required"`                  // 商品条码
	ProductStatus         int     `json:"product_status" binding:"required"`           // 商品状态, 1-上架 0-下架
	ProductRatingTaxType  string  `json:"product_rating_tax_type" binding:"required"`  // 堂食税类
	ProductTakeoutTaxType string  `json:"product_takeout_tax_type" binding:"required"` // 外带税类
	DeductStockType       int     `json:"deduct_stock_type" binding:"required"`        // 库存计算方式, 2-付款减库存 1-下单减库存
	Shows                 string  `json:"shows" binding:"required"`                    // 显示：123456
	IsEnableGrade         int     `json:"is_enable_grade" binding:"required"`          // 是否开启会员折扣(1开启 0关闭)
	OpenOverallDiscount   int     `json:"open_overall_discount" binding:"required"`    // 整单折扣(1开启 0关闭)
	Row                   int     `json:"row" binding:"required"`                      // excel表的行编号
}

// ProductImportReq 导入商品请求
type ProductImportReq struct {
	List []ProductImportItemReq `json:"list" binding:"required,dive"` // 商品列表
}

// ProductImportItemReq 导入商品项请求
type ProductImportItemReq struct {
	ProductName           string  `json:"product_name" binding:"required"`             // 商品名称
	CategoryName          string  `json:"category_name" binding:"required"`            // 分类名称
	ProductUnit           string  `json:"product_unit" binding:"required"`             // 商品单位
	SkuName               string  `json:"sku_name" binding:"required"`                 // 规格名称
	ProductPrice          float64 `json:"product_price" binding:"required"`            // 商品价格
	NumType               int     `json:"num_type" binding:"required"`                 // 数量计算方法, 1-整数 2-小数
	Barcode               string  `json:"barcode" binding:"required"`                  // 商品条码
	ProductStatus         int     `json:"product_status" binding:"required"`           // 商品状态, 1-上架 0-下架
	ProductRatingTaxType  string  `json:"product_rating_tax_type" binding:"required"`  // 堂食税类
	ProductTakeoutTaxType string  `json:"product_takeout_tax_type" binding:"required"` // 外带税类
	DeductStockType       int     `json:"deduct_stock_type" binding:"required"`        // 库存计算方式, 2-付款减库存 1-下单减库存
	Shows                 string  `json:"shows" binding:"required"`                    // 显示：123456
	IsEnableGrade         int     `json:"is_enable_grade" binding:"required"`          // 是否开启会员折扣(1开启 0关闭)
	OpenOverallDiscount   int     `json:"open_overall_discount" binding:"required"`    // 整单折扣(1开启 0关闭)
	Row                   int     `json:"row" binding:"required"`                      // excel表的行编号
	IsShowCashier         bool    `json:"is_show_cashier"`                             // 是否显示在收银端 1-显示 2-不显示
	IsShowTablet          bool    `json:"is_show_tablet"`                              // 是否显示在平板端 1-显示 2-不显示
	IsShowKitchen         bool    `json:"is_show_kitchen"`                             // 是否显示在送厨端 1-显示 2-不显示
	IsShowAssistant       bool    `json:"is_show_assistant"`                           // 是否显示在点餐助手 1-显示 2-不显示
	IsShowH5              bool    `json:"is_show_h5"`                                  // 是否显示在h5 1-显示 2-不显示
	IsShowDelivery        bool    `json:"is_show_delivery"`                            // 是否显示在外送 1-显示 2-不显示
	UnitId                uint64  `json:"unit_id"`                                     // 单位ID
	CategoryId            uint64  `json:"category_id"`                                 // 分类ID
	SkuId                 uint64  `json:"sku_id"`                                      // 规格ID
	RatingTaxId           uint64  `json:"ratin_tax_id"`                                // 堂食税类ID
	TakeoutTaxId          uint64  `json:"takeout_tax_id"`                              // 外带税类ID
}

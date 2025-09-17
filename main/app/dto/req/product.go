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
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductUnitEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"`        // 商品单位UUID
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品单位名称
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductUnitSortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品单位UUID
	Sort int    `json:"sort" binding:"required"` // 排序
}

type ProductUnitSortReq struct {
	List []ProductUnitSortItem `json:"list" binding:"required,dive"` // 商品单位排序列表
}

type ProductSauceAddReq struct {
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品加料名称
	Price               *float64           `json:"price" binding:"required"`       // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
}

type ProductSauceEditReq struct {
	Uuid                uint64             `json:"uuid" binding:"required"`        // 商品加料UUID
	LocaleName          dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品加料名称
	Price               *float64           `json:"price" binding:"required"`       // 商品加料价格
	ProductPackageUuids []uint64           `json:"product_package_uuids"`          // 关联商品包UUID列表
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

// ProductShopCategoryReq 商品分类请求
type ProductShopCategoryReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品分类UUID
}

// ProductShopCategorySortItem 商品分类排序项
type ProductShopCategorySortItem struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品分类UUID
	Sort int    `json:"sort"`                    // 排序
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
	Code       string             `json:"code"`                           // 分类编码
}

// ProductShopCategoryEditReq 商品分类编辑请求
type ProductShopCategoryEditReq struct {
	Uuid       uint64             `json:"uuid" binding:"required"`        // 商品分类UUID
	ParentUuid uint64             `json:"parent_uuid"`                    // 父级分类UUID, 一级分类为0, 二级分类为一级分类的uuid
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 商品分类名称, 多语言
	Status     int                `json:"status"`                         // 商品分类状态 0-关闭 1-开启
	Code       string             `json:"code"`                           // 分类编码
}

// ProductShopCategoryDeleteReq 商品分类请求
type ProductShopCategoryDeleteReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品分类UUID
}

type ProductAttributeGroupListReq struct {
	dto.PageReq // 分页参数
}

type ProductAttributeGroupReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品属性分组UUID
}

type ProductAttributeDeleteReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品属性UUID
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
	Sort                int                `json:"sort" binding:"required"`        // 排序
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
	Sort int    `json:"sort"`                    // 排序
}

// ProductFlavorSortReq 商品规格排序请求
type ProductFlavorSortReq struct {
	List []ProductFlavorSortItemReq `json:"list" binding:"required,dive"` // 商品规格排序列表
}

// ProductFlavorEditReq 商品规格编辑请求
type ProductFlavorEditReq struct {
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

// GetBarcodeDuplicateRows 获取条形码重复的行号
func (p *ProductImportListReq) GetBarcodeDuplicateRows() []int {
	barcodeMap := make(map[string]bool)
	duplicateRows := []int{}
	for _, item := range p.List {
		if _, ok := barcodeMap[item.Barcode]; ok {
			duplicateRows = append(duplicateRows, item.Row)
		}
		barcodeMap[item.Barcode] = true
	}
	return duplicateRows
}

// ProductImportItemReq 导入商品项请求
type ProductImportListItemReq struct {
	LocaleName            dto.LocaleResponse `json:"locale_name"`              // 商品名称
	CategoryName          string             `json:"category_name"`            // 分类名称
	ProductUnit           string             `json:"product_unit"`             // 商品单位
	SkuName               string             `json:"sku_name"`                 // 规格名称
	ProductPrice          float64            `json:"product_price"`            // 商品价格
	NumType               int                `json:"num_type"`                 // 数量计算方法, 1-整数 2-小数
	Barcode               string             `json:"barcode"`                  // 商品条码
	ProductStatus         int                `json:"product_status"`           // 商品状态, 1-上架 0-下架
	ProductRatingTaxType  string             `json:"product_rating_tax_type"`  // 堂食税类
	ProductTakeoutTaxType string             `json:"product_takeout_tax_type"` // 外带税类
	DeductStockType       int                `json:"deduct_stock_type"`        // 库存计算方式, 2-付款减库存 1-下单减库存
	Shows                 string             `json:"shows"`                    // 显示：123456
	IsEnableGrade         int                `json:"is_enable_grade"`          // 是否开启会员折扣(1开启 0关闭)
	OpenOverallDiscount   int                `json:"open_overall_discount"`    // 整单折扣(1开启 0关闭)
	Row                   int                `json:"row"`                      // excel表的行编号
}

// ProductImportReq 导入商品请求
type ProductImportReq struct {
	List []ProductImportItemReq `json:"list" binding:"required,dive"` // 商品列表
}

// GetBarcodeDuplicateRows 获取条形码重复的行号
func (p *ProductImportReq) GetBarcodeDuplicateRows() []int {
	barcodeMap := make(map[string]bool)
	duplicateRows := []int{}
	for _, item := range p.List {
		if item.Barcode == "" {
			continue
		}
		if _, ok := barcodeMap[item.Barcode]; ok {
			duplicateRows = append(duplicateRows, item.Row)
		}
		barcodeMap[item.Barcode] = true
	}
	return duplicateRows
}

// ProductImportItemReq 导入商品项请求
type ProductImportItemReq struct {
	LocaleName            dto.LocaleResponse `json:"locale_name"`              // 商品名称
	CategoryName          string             `json:"category_name"`            // 分类名称
	ProductUnit           string             `json:"product_unit"`             // 商品单位
	SkuName               string             `json:"sku_name"`                 // 规格名称
	ProductPrice          float64            `json:"product_price"`            // 商品价格
	NumType               int                `json:"num_type"`                 // 数量计算方法, 1-整数 2-小数
	Barcode               string             `json:"barcode"`                  // 商品条码
	ProductStatus         int                `json:"product_status"`           // 商品状态, 1-上架 0-下架
	ProductRatingTaxType  string             `json:"product_rating_tax_type"`  // 堂食税类
	ProductTakeoutTaxType string             `json:"product_takeout_tax_type"` // 外带税类
	DeductStockType       int                `json:"deduct_stock_type"`        // 库存计算方式, 2-付款减库存 1-下单减库存
	Shows                 string             `json:"shows"`                    // 显示：123456
	IsEnableGrade         int                `json:"is_enable_grade"`          // 是否开启会员折扣(1开启 0关闭)
	OpenOverallDiscount   int                `json:"open_overall_discount"`    // 整单折扣(1开启 0关闭)
	Row                   int                `json:"row"`                      // excel表的行编号
	IsShowCashier         int                `json:"is_show_cashier"`          // 是否显示在收银端 1-显示 0-不显示
	IsShowTablet          int                `json:"is_show_tablet"`           // 是否显示在平板端 1-显示 0-不显示
	IsShowKitchen         int                `json:"is_show_kitchen"`          // 是否显示在送厨端 1-显示 0-不显示
	IsShowAssistant       int                `json:"is_show_assistant"`        // 是否显示在点餐助手 1-显示 0-不显示
	IsShowH5              int                `json:"is_show_h5"`               // 是否显示在h5 1-显示 0-不显示
	IsShowDelivery        int                `json:"is_show_delivery"`         // 是否显示在外送 1-显示 0-不显示
	UnitUuid              uint64             `json:"unit_uuid"`                // 单位UUID
	CategoryUuid          uint64             `json:"category_uuid"`            // 分类UUID
	SkuUuid               uint64             `json:"sku_uuid"`                 // 规格UUID
	DineTaxUuid           uint64             `json:"dine_tax_uuid"`            // 堂食税类UUID
	TakeoutTaxUuid        uint64             `json:"takeout_tax_uuid"`         // 外带税类UUID
}

// ProductSingleListReq 单规格商品列表查询
type ProductSingleListReq struct {
	dto.PageReq          // 分页参数
	Keyword      *string `form:"keyword"`       // 搜索关键词, 可选
	CategoryUuid *uint64 `form:"category_uuid"` // 商品分类UUID, 可选
}

// ProductShopListReq 商品列表查询
type ProductShopListReq struct {
	dto.PageReq          // 分页参数
	Keyword      *string `form:"keyword"`       // 搜索商品名称（可选）
	Type         *string `form:"type"`          // 商品类型: 0-商品 1-套餐（可选）
	Tag          *string `form:"tag"`           // 商品标签: 0-多规格 1-属性 2-加料（可选）, 多个标签用逗号分隔: 0,1,2
	Status       *string `form:"status"`        // 商品状态: 0-下架 1-上架（可选）
	CategoryUuid *string `form:"category_uuid"` // 商品分类UUID（可选）
}

// SortProductShopListReq 商品排序请求
type SortProductShopListReq struct {
	CategoryUuid uint64                   `json:"category_uuid"`                // 商品分类UUID
	List         []SortProductShopItemReq `json:"list" binding:"required,dive"` // 商品排序列表
}

// SortProductShopItemReq 商品排序项请求
type SortProductShopItemReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品UUID
	Sort int    `json:"sort"`                    // 排序
}

// ProductDetailReq 商品详情请求
type ProductDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 商品UUID
}

// ProductShopStatusReq 商品状态请求
type ProductShopStatusReq struct {
	Uuid   uint64 `json:"uuid" binding:"required"`             // 商品UUID
	Status *int   `json:"status" binding:"required,oneof=0 1"` // 商品状态 0-下架 1-上架
}

// ProductShopAddReq 商品添加请求
type ProductShopAddReq struct {
	Type            int                               `json:"type"`              // 商品类型 0-商品 1-套餐
	LocaleName      dto.LocaleResponse                `json:"locale_name"`       // 商品名称
	CategoryUuid    uint64                            `json:"category_uuid"`     // 商品分类UUID
	UnitUuid        uint64                            `json:"unit_uuid"`         // 商品单位UUID
	Flavors         []ProductShopAddFlavorReq         `json:"flavors"`           // 商品规格列表
	Tax             ProductShopAddTaxReq              `json:"tax"`               // 商品税类
	Status          int                               `json:"status"`            // 商品状态 0-下架 1-上架
	ImageFileUuid   uint64                            `json:"image_file_uuid"`   // 商品图片文件UUID
	NumType         int                               `json:"num_type"`          // 数量计算方法 0-整数 1-小数
	Attributes      []ProductShopAddAttributeGroupReq `json:"attributes"`        // 商品属性列表
	Sauce           ProductShopAddSauceReq            `json:"sauce"`             // 商品加料列表
	DeductStockType int                               `json:"deduct_stock_type"` // 库存计算方式 0-付款减库存 1-下单减库存
	Show            ProductShopAddShowReq             `json:"show"`              // 商品显示设置
	Discount        ProductShopAddDiscountReq         `json:"discount"`          // 商品折扣设置
	Package         ProductShopAddPackageReq          `json:"package"`           // 商品套餐
	Row             int                               `json:"row"`               // 行号
}

// ProductShopAddFlavorReq 商品规格添加请求
type ProductShopAddFlavorReq struct {
	Uuid         uint64  `json:"uuid"`          // 商品规格UUID
	Price        float64 `json:"price"`         // 商品规格价格
	BarcodeValue string  `json:"barcode_value"` // 商品加料条码值, 可选
}

// ProductShopAddTaxReq 商品税类添加请求
type ProductShopAddTaxReq struct {
	DineUuid    uint64 `json:"dine_uuid"`    // 堂食税类UUID
	TakeoutUuid uint64 `json:"takeout_uuid"` // 外带税类UUID
}

// ProductShopAddSauceReq 商品加料添加请求
type ProductShopAddSauceReq struct {
	Sauces       []ProductShopAddSauceItemReq `json:"sauces"`        // 商品加料列表
	IsMust       int                          `json:"is_must"`       // 是否必选 0-否 1-是
	MaxSelection int                          `json:"max_selection"` // 最大选择数量
	IsOpenInput  bool                         `json:"is_open_input"` // 是否开启输入 0-否 1-是
}

// ProductShopAddSauceItemReq 商品加料项添加请求
type ProductShopAddSauceItemReq struct {
	Uuid              uint64 `json:"uuid"`                // 商品加料UUID
	IsDefaultSelected int    `json:"is_default_selected"` // 是否默认选中 0-否 1-是
}

// ProductShopAddAttributeGroupReq 商品属性组添加请求
type ProductShopAddAttributeGroupReq struct {
	Uuid         uint64                            `json:"uuid"`          // 商品属性组UUID
	IsMust       int                               `json:"is_must"`       // 商品属性组是否必选 0-否 1-是
	IsOpenInput  bool                              `json:"is_open_input"` // 商品属性组最大选择数量是否开启 false-关闭 true-开启
	MaxSelection int                               `json:"max_selection"` // 商品属性组最大选择数量
	Attributes   []ProductShopAddGroupAttributeReq `json:"attributes"`    // 商品属性列表
}

// ProductShopAddGroupAttributeReq 商品属性组-属性值添加请求
type ProductShopAddGroupAttributeReq struct {
	Uuid              uint64 `json:"uuid"`                // 商品属性组-属性值UUID
	IsDefaultSelected int    `json:"is_default_selected"` // 商品属性是否默认选中 0-否 1-是
}

// ProductShopAddShowReq 商品显示设置添加请求
type ProductShopAddShowReq struct {
	IsShowCashier   int `json:"is_show_cashier"`   // 是否显示在收银端 0-不显示 1-显示
	IsShowTablet    int `json:"is_show_tablet"`    // 是否显示在平板端 0-不显示 1-显示
	IsShowKitchen   int `json:"is_show_kitchen"`   // 是否显示在厨显端 0-不显示 1-显示
	IsShowAssistant int `json:"is_show_assistant"` // 是否显示在点餐助手 0-不显示 1-显示
	IsShowH5        int `json:"is_show_h5"`        // 是否显示在h5 0-不显示 1-显示
	IsShowDelivery  int `json:"is_show_delivery"`  // 是否显示在外送 0-不显示 1-显示
}

// ProductShopAddDiscountReq 商品折扣设置添加请求
type ProductShopAddDiscountReq struct {
	IsEnableMemberDiscount  int `json:"is_enable_member_discount"`  // 是否开启会员折扣 0-关闭 1-开启
	IsEnableOverallDiscount int `json:"is_enable_overall_discount"` // 是否开启整单折扣 0-关闭 1-开启
}

// ProductShopAddPackageReq 商品套餐添加请求
type ProductShopAddPackageReq struct {
	Price  float64                         `json:"price"`  // 套餐价格
	Groups []ProductShopAddPackageGroupReq `json:"groups"` // 套餐分组列表
}

// ProductShopAddPackageGroupReq 套餐分组添加请求
type ProductShopAddPackageGroupReq struct {
	LocaleName dto.LocaleResponse                     `json:"locale_name"` // 套餐分组名称
	Products   []ProductShopAddPackageGroupProductReq `json:"products"`    // 套餐分组商品列表
}

// ProductShopAddPackageGroupProductReq 套餐分组商品添加请求
type ProductShopAddPackageGroupProductReq struct {
	BomUuid uint64  `json:"bom_uuid"` // 商品BOM UUID
	Num     float64 `json:"num"`      // 商品数量
	Sort    int     `json:"sort"`     // 商品排序
}

// ProductShopEditReq 商品编辑请求
type ProductShopEditReq struct {
	Uuid            uint64                             `json:"uuid"`              // 商品UUID
	Type            int                                `json:"type"`              // 商品类型 0-商品 1-套餐
	LocaleName      dto.LocaleResponse                 `json:"locale_name"`       // 商品名称
	CategoryUuid    uint64                             `json:"category_uuid"`     // 商品分类UUID
	UnitUuid        uint64                             `json:"unit_uuid"`         // 商品单位UUID
	Flavors         []ProductShopEditFlavorReq         `json:"flavors"`           // 商品规格列表
	Tax             ProductShopEditTaxReq              `json:"tax"`               // 商品税类
	Status          int                                `json:"status"`            // 商品状态 0-下架 1-上架
	ImageFileUuid   uint64                             `json:"image_file_uuid"`   // 商品图片文件UUID
	NumType         int                                `json:"num_type"`          // 数量计算方法 0-整数 1-小数
	Attributes      []ProductShopEditAttributeGroupReq `json:"attributes"`        // 商品属性列表
	Sauce           ProductShopEditSauceReq            `json:"sauce"`             // 商品加料列表
	DeductStockType int                                `json:"deduct_stock_type"` // 库存计算方式 0-付款减库存 1-下单减库存
	Show            ProductShopEditShowReq             `json:"show"`              // 商品显示设置
	Discount        ProductShopEditDiscountReq         `json:"discount"`          // 商品折扣设置
	Package         ProductShopEditPackageReq          `json:"package"`           // 商品套餐
}

// ProductShopAddFlavorReq 商品规格添加请求
type ProductShopEditFlavorReq struct {
	Uuid         uint64  `json:"uuid"`          // 商品规格UUID
	Price        float64 `json:"price"`         // 商品规格价格
	BarcodeValue string  `json:"barcode_value"` // 商品加料条码值, 可选
	BomUuid      uint64  `json:"bom_uuid"`      // 商品BOM UUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete     bool    `json:"is_delete"`     // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopAddTaxReq 商品税类添加请求
type ProductShopEditTaxReq struct {
	DineUuid    uint64 `json:"dine_uuid"`    // 堂食税类UUID
	TakeoutUuid uint64 `json:"takeout_uuid"` // 外带税类UUID
}

// ProductShopAddSauceReq 商品加料添加请求
type ProductShopEditSauceReq struct {
	Sauces       []ProductShopEditSauceItemReq `json:"sauces"`        // 商品加料列表
	IsMust       int                           `json:"is_must"`       // 是否必选 0-否 1-是
	IsOpenInput  bool                          `json:"is_open_input"` // 是否开启输入 0-否 1-是
	MaxSelection int                           `json:"max_selection"` // 最大选择数量
}

// ProductShopAddSauceItemReq 商品加料项添加请求
type ProductShopEditSauceItemReq struct {
	Uuid              uint64 `json:"uuid"`                // 商品加料UUID
	IsDefaultSelected int    `json:"is_default_selected"` // 是否默认选中 0-否 1-是
	BomUuid           uint64 `json:"bom_uuid"`            // 商品BOM UUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete          bool   `json:"is_delete"`           // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopAddAttributeGroupReq 商品属性组添加请求
type ProductShopEditAttributeGroupReq struct {
	Uuid         uint64                             `json:"uuid"`          // 商品属性组UUID
	IsMust       int                                `json:"is_must"`       // 商品属性组是否必选 0-否 1-是
	IsOpenInput  bool                               `json:"is_open_input"` // 商品属性组最大选择数量是否开启 false-否 true-是
	MaxSelection int                                `json:"max_selection"` // 商品属性组最大选择数量
	Attributes   []ProductShopEditGroupAttributeReq `json:"attributes"`    // 商品属性列表
	IsDelete     bool                               `json:"is_delete"`     // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopAddGroupAttributeReq 商品属性组-属性值添加请求
type ProductShopEditGroupAttributeReq struct {
	Uuid              uint64 `json:"uuid"`                // 商品属性组-属性值UUID
	IsDefaultSelected int    `json:"is_default_selected"` // 商品属性是否默认选中 0-否 1-是
	IsDelete          bool   `json:"is_delete"`           // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopAddShowReq 商品显示设置添加请求
type ProductShopEditShowReq struct {
	IsShowCashier   int `json:"is_show_cashier"`   // 是否显示在收银端 0-不显示 1-显示
	IsShowTablet    int `json:"is_show_tablet"`    // 是否显示在平板端 0-不显示 1-显示
	IsShowKitchen   int `json:"is_show_kitchen"`   // 是否显示在厨显端 0-不显示 1-显示
	IsShowAssistant int `json:"is_show_assistant"` // 是否显示在点餐助手 0-不显示 1-显示
	IsShowH5        int `json:"is_show_h5"`        // 是否显示在h5 0-不显示 1-显示
	IsShowDelivery  int `json:"is_show_delivery"`  // 是否显示在外送 0-不显示 1-显示
}

// ProductShopAddDiscountReq 商品折扣设置添加请求
type ProductShopEditDiscountReq struct {
	IsEnableMemberDiscount  int `json:"is_enable_member_discount"`  // 是否开启会员折扣 0-关闭 1-开启
	IsEnableOverallDiscount int `json:"is_enable_overall_discount"` // 是否开启整单折扣 0-关闭 1-开启
}

// ProductShopEditPackageReq 商品套餐添加请求
type ProductShopEditPackageReq struct {
	Price  float64                          `json:"price"`  // 套餐价格
	Groups []ProductShopEditPackageGroupReq `json:"groups"` // 套餐分组列表
}

// ProductShopAddPackageGroupReq 套餐分组添加请求
type ProductShopEditPackageGroupReq struct {
	Uuid       uint64                                  `json:"uuid"`        // 套餐分组UUID
	LocaleName dto.LocaleResponse                      `json:"locale_name"` // 套餐分组名称
	Products   []ProductShopEditPackageGroupProductReq `json:"products"`    // 套餐分组商品列表
	IsDelete   bool                                    `json:"is_delete"`   // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopAddPackageGroupProductReq 套餐分组商品添加请求
type ProductShopEditPackageGroupProductReq struct {
	Uuid     uint64  `json:"uuid"`      // 套餐商品UUID
	BomUuid  uint64  `json:"bom_uuid"`  // 商品BOM UUID
	Num      float64 `json:"num"`       // 商品数量
	Sort     int     `json:"sort"`      // 商品排序
	IsDelete bool    `json:"is_delete"` // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

// ProductShopDeleteReq 商品删除请求
type ProductShopDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 商品UUID
}

// ProductShopChangePriceReq 商品改价请求
type ProductShopChangePriceReq struct {
	Uuid   uint64                          `json:"uuid"`   // 商品UUID
	Prices []ProductShopChangePriceItemReq `json:"prices"` // 商品价格列表
}

// ProductShopChangePriceItemReq 商品改价项请求
type ProductShopChangePriceItemReq struct {
	Uuid  uint64  `json:"uuid"`  // 商品BOM UUID
	Price float64 `json:"price"` // 商品价格
}

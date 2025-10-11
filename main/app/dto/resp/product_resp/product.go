package product_resp

import "ttpos-server-go/app/dto"

type ProductSearchResp struct {
	List []Product `json:"list"`
}

// Product 商品
type Product struct {
	Uuid                uint64                    `json:"uuid"`                  // 商品UUID
	LocaleName          dto.LocaleResponse        `json:"locale_name"`           // 商品名称
	Image               string                    `json:"image"`                 // 商品图片
	Unit                dto.LocaleResponse        `json:"unit"`                  // 商品单位
	NumType             uint                      `json:"num_type"`              // 商品数量计算方法 0-整数 1-小数
	Price               float64                   `json:"price"`                 // 商品价格
	CategoryUuid        uint64                    `json:"category_uuid"`         // 商品类别UUID
	FirstCategoryUuid   uint64                    `json:"first_category_uuid"`   // 商品一级类别UUID
	SpecialCategoryUuid uint64                    `json:"special_category_uuid"` // 商品特殊类别UUID
	LimitNum            uint                      `json:"limit_num"`             // 商品限购数量
	Flavors             ProductFlavorList         `json:"flavors"`               // 商品规格
	Sauces              ProductSauceList          `json:"sauces"`                // 商品小料
	AttributeGroups     ProductAttributeGroupList `json:"attribute_groups"`      // 商品属性组
	Describe            string                    `json:"describe"`              // 卖点，h5端显示
	IsShowKitchen       uint                      `json:"is_show_kitchen"`       // 是否在厨显端显示：1-是；0-否
	ProductType         uint                      `json:"product_type"`          // 商品类型 0-商品 1-套餐
	// 套餐分组
	PackageGroupList *ProductPackageGroupList `json:"package_group_list"`

	Sort int `json:"-"` // 商品排序，内部字段，用于推荐商品列表排序

}

// ProductPackageGroupList 套餐分组列表
type ProductPackageGroupList struct {
	List []ProductPackageGroup `json:"list"`
}

// ProductPackageGroup 套餐分组
type ProductPackageGroup struct {
	Uuid       uint64             `json:"uuid"`        // 套餐分组UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 套餐分组名称
	IsFull     bool               `json:"is_full"`     // 是否选满
	Num        int                `json:"num"`         // 套餐商品数量
	Products   ProductList        `json:"products"`    // 套餐商品列表
}

// 判断套餐分组是否选满
func (p *ProductPackageGroup) GetIsFull() bool {
	for _, product := range p.Products.List {
		// 如果套餐商品可以编辑，则套餐分组未选满
		if product.CanEdit {
			return false
		}
	}
	return true
}

// ProductList 商品列表
type ProductList struct {
	List []PackageProductDetail `json:"list"`
}

// PackageProductDetail 套餐商品详情
type PackageProductDetail struct {
	Detail  Product `json:"detail"`   // 商品详情
	Num     float64 `json:"num"`      // 商品数量，分组中item的数量
	CanEdit bool    `json:"can_edit"` // 是否可以编辑
}

// 判断是否可以编辑。无需选择属性的商品，不可以编辑
func (p *PackageProductDetail) GetCanEdit() bool {
	// 如果套餐商品没有属性，则不可以编辑
	if p.Detail.AttributeGroups.List == nil || len(p.Detail.AttributeGroups.List) == 0 {
		return false
	}
	// 如果套餐商品有属性，但必须且只有一个时，则不可以编辑
	if len(p.Detail.AttributeGroups.List) == 1 && p.Detail.AttributeGroups.List[0].IsMust {
		return false
	}

	return true
}

// ProductBatchType 分批类型
type BatchTag struct {
	Uuid       uint64             `json:"uuid"`        // 分批类型UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Color      string             `json:"color"`       // 颜色值，如#FF0000
	Sort       int                `json:"sort"`        // 排序，数字越小越靠前
}

// BatchTagList 分批类型列表
type BatchTagList struct {
	List []BatchTag `json:"list"`
}

// ProductBatchTypeDetail 分批类型详情
type BatchTagDetail struct {
	Uuid       uint64             `json:"uuid"`        // 分批类型UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 分批类型名称，多语言
	Color      string             `json:"color"`       // 颜色值，如#FF0000
	Sort       int                `json:"sort"`        // 排序，数字越小越靠前
}

// BatchTagList 分批类型列表
type BatchTagColorUsageList struct {
	List []BatchTagColorUsage `json:"list"`
}

// BatchTagColorUsage 色块被选择情况
type BatchTagColorUsage struct {
	Color        string `json:"color"`          // 颜色值
	IsUsed       bool   `json:"is_used"`        // 是否已被使用
	UsedBy       string `json:"used_by"`        // 被使用的分批类型名称（如果被使用的话）
	BatchTagUuid uint64 `json:"batch_tag_uuid"` // 分批类型UUID
}

// ProductFlavor 商品规格
type ProductFlavor struct {
	Uuid         uint64             `json:"uuid"`          // 商品规格UUID
	BomUuid      uint64             `json:"bom_uuid"`      // 商品BOM UUID
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 商品规格名称
	Price        float64            `json:"price"`         // 商品规格价格
	StockNum     float64            `json:"stock_num"`     // 商品库存数量
	Barcode      string             `json:"barcode"`       // 商品码。用于根据扫码枪扫码商品得到商品码在商品列表中搜索到商品
	InternalCode string             `json:"internal_code"` // 商品规格内部编码
}

// ProductSauce 商品小料
type ProductSauce struct {
	Uuid              uint64             `json:"uuid"`                // 商品小料UUID
	BomUuid           uint64             `json:"bom_uuid"`            // 商品BOM UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品小料名称
	Price             float64            `json:"price"`               // 商品小料价格
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
	StockNum          float64            `json:"stock_num"`           // 小料库存数量
}

// ProductAttributeGroup 商品属性组
type ProductAttributeGroup struct {
	Uuid       uint64                    `json:"uuid"`        // 商品属性组UUID
	LocaleName dto.LocaleResponse        `json:"locale_name"` // 商品属性组名称
	Attributes ProductAttributeValueList `json:"attributes"`  // 商品属性值列表
	IsMust     bool                      `json:"is_must"`     // 是否必选
	MaxSelect  uint                      `json:"max_select"`  // 最大可选数量
}

// ProductAttributeValue 商品属性值
type ProductAttributeValue struct {
	Uuid              uint64             `json:"uuid"`                // 商品属性UUID
	LocaleName        dto.LocaleResponse `json:"locale_name"`         // 商品属性名称
	IsDefaultSelected bool               `json:"is_default_selected"` // 是否默认选中
}

// ProductFlavorList 商品规格列表
type ProductFlavorList struct {
	List []ProductFlavor `json:"list"`
}

// ProductSauceList 商品小料列表
type ProductSauceList struct {
	List      []ProductSauce `json:"list"`
	IsMust    bool           `json:"is_must"`    // 是否必选小料
	MaxSelect int            `json:"max_select"` // 小料最大可选数量
}

// ProductAttributeGroupList 商品属性组列表
type ProductAttributeGroupList struct {
	List []ProductAttributeGroup `json:"list"`
}

// ProductAttributeValueList 商品属性值列表
type ProductAttributeValueList struct {
	List []ProductAttributeValue `json:"list"`
}

// ProductListWithPaginationResp 商品列表响应
type ProductListWithPaginationResp struct {
	List []Product        `json:"list"`
	Meta dto.PageResponse `json:"meta"`
}

// ProductRecommendListResp 商品推荐列表响应
type ProductRecommendListResp struct {
	List   []Product `json:"list"`    // 商品列表
	Title  string    `json:"title"`   // 推荐标题，如“商家推荐”
	IsOpen bool      `json:"is_open"` // 是否开启推荐, true-开启, false-关闭
}

// ProductCategory 商品类别（销售端）
type ProductCategory struct {
	Uuid        uint64                  `json:"uuid"`         // 商品类别UUID
	LocaleName  dto.LocaleResponse      `json:"locale_name"`  // 商品类别名称
	ParentUuid  uint64                  `json:"parent_uuid"`  // 父级类别UUID
	IsSpecial   bool                    `json:"is_special"`   // 是否特殊类别
	Children    ProductCategoryListResp `json:"children"`     // 子级类别
	CategoryKey string                  `json:"category_key"` // all-表示全部
}

// ProductCategoryListResp 商品类别列表响应（销售端）
type ProductCategoryListResp struct {
	List []ProductCategory `json:"list"`
}

// ProductUnit 商品单位
type ProductUnitItem struct {
	Uuid                uint64 `json:"uuid"`                  // 商品单位UUID
	Sort                int    `json:"sort"`                  // 商品单位排序
	ProductPackageCount int    `json:"product_package_count"` // 关联商品包数量
	Name                string `json:"name"`                  // 商品单位名称
	IsEditable          bool   `json:"is_editable"`           // 是否可编辑
}

// ProductUnitListResp 商品单位列表响应
type ProductUnitListResp struct {
	List []ProductUnitItem `json:"list"`
	Meta dto.PageResponse  `json:"meta"`
}

type ProductUnitProductPackage struct {
	Uuid uint64 `json:"uuid"` // 商品包UUID
	Name string `json:"name"` // 商品包名称
}

type ProductUnitProductPackageList struct {
	List []ProductUnitProductPackage `json:"list"`
}

type ProductUnitDetail struct {
	Uuid            uint64                        `json:"uuid"`             // 商品单位UUID
	LocaleName      dto.LocaleResponse            `json:"locale_name"`      // 商品单位名称
	ProductPackages ProductUnitProductPackageList `json:"product_packages"` // 商品包列表
	IsEditable      bool                          `json:"is_editable"`      // 是否可编辑
}

type ProductSauceListResp struct {
	List []ProductSauceItem `json:"list"`
	Meta dto.PageResponse   `json:"meta"`
}

type ProductSauceItem struct {
	Uuid                uint64             `json:"uuid"`                  // 商品加料UUID
	Name                string             `json:"name"`                  // 商品加料名称
	Price               float64            `json:"price"`                 // 商品加料价格
	Sort                int                `json:"sort"`                  // 商品加料排序
	ProductPackageCount int                `json:"product_package_count"` // 关联商品包数量
	ProductBomCardUuid  uint64             `json:"product_bom_card_uuid"` // 成本卡UUID，0表示没有成本卡
	ProductBomCardName  dto.LocaleResponse `json:"product_bom_card_name"` // 成本卡名称
	IsEditable          bool               `json:"is_editable"`           // 是否可编辑
}

type ProductSauceProductPackage struct {
	Uuid uint64 `json:"uuid"` // 商品包UUID
	Name string `json:"name"` // 商品包名称
}

type ProductSauceProductPackageList struct {
	List []ProductSauceProductPackage `json:"list"`
}

type ProductSauceDetail struct {
	Uuid            uint64                         `json:"uuid"`             // 商品单位UUID
	Price           float64                        `json:"price"`            // 商品加料价格
	LocaleName      dto.LocaleResponse             `json:"locale_name"`      // 商品单位名称
	ProductPackages ProductSauceProductPackageList `json:"product_packages"` // 商品包列表
	IsEditable      bool                           `json:"is_editable"`      // 是否可编辑
}

// ProductShopCategory 商品类别（商家端）
type ProductShopCategory struct {
	Uuid        uint64                      `json:"uuid"`         // 商品类别UUID
	Name        string                      `json:"name"`         // 商品类别名称
	ParentUuid  uint64                      `json:"parent_uuid"`  // 父级类别UUID
	IsSpecial   bool                        `json:"is_special"`   // 是否特色类别
	Sort        uint                        `json:"sort"`         // 商品类别排序
	Status      int                         `json:"status"`       // 商品类别状态 0-关闭 1-开启
	IsEditable  bool                        `json:"is_editable"`  // 是否可编辑
	CategoryKey string                      `json:"category_key"` // 商品类别关键字
	Children    ProductShopCategoryListResp `json:"children"`     // 子级类别
}

// ProductShopCategoryListResp 商品类别列表响应（商家端）
type ProductShopCategoryListResp struct {
	List []ProductShopCategory `json:"list"`
}

// ProductShopCategoryDetailResp 商品分类详情响应（商家端）
type ProductShopCategoryDetailResp struct {
	Uuid         uint64             `json:"uuid"`          // 商品类别UUID
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 商品类别名称
	ParentUuid   uint64             `json:"parent_uuid"`   // 父级类别UUID
	ParentName   string             `json:"parent_name"`   // 父级类别名称
	Sort         uint               `json:"sort"`          // 商品类别排序
	Status       int                `json:"status"`        // 商品类别状态 0-关闭 1-开启
	ProductCount int64              `json:"product_count"` // 商品数量
	ChildCount   int64              `json:"child_count"`   // 子级数量
	Code         string             `json:"code"`          // 分类编码
	IsEditable   bool               `json:"is_editable"`   // 是否可编辑
	CategoryKey  string             `json:"category_key"`  // 商品类别关键字
}

type ProductAttributeGroupItem struct {
	Uuid          uint64                               `json:"uuid"`           // 商品属性分组UUID
	Name          string                               `json:"name"`           // 商品属性分组名称
	Sort          int                                  `json:"sort"`           // 商品属性分组排序
	AttributeName string                               `json:"attribute_name"` // 商品属性名称
	Attributes    []ProductAttributeGroupAttributeItem `json:"attributes"`     // 属性值
	IsEditable    bool                                 `json:"is_editable"`    // 是否可编辑
}

type ProductAttributeGroupAttributeItem struct {
	Uuid       uint64             `json:"uuid"`        // 商品属性UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品属性名称
	IsEditable bool               `json:"is_editable"` // 是否可编辑
}

type ProductAttributeGroupListResp struct {
	List []ProductAttributeGroupItem `json:"list"`
	Meta dto.PageResponse            `json:"meta"`
}

type ProductAttributeGroupDetail struct {
	Uuid       uint64             `json:"uuid"`        // 商品属性分组UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品属性分组名称
	Attributes ProductAttributes  `json:"attributes"`  // 商品属性值列表
	IsEditable bool               `json:"is_editable"` // 是否可编辑
}

type ProductAttributes struct {
	List []ProductAttribute `json:"list"`
}

type ProductAttribute struct {
	Uuid            uint64                             `json:"uuid"`             // 商品属性UUID
	LocaleName      dto.LocaleResponse                 `json:"locale_name"`      // 商品属性名称
	ProductPackages ProductAttributeProductPackageList `json:"product_packages"` // 关联商品包列表
	Sort            int                                `json:"sort"`             // 商品属性排序
	IsEditable      bool                               `json:"is_editable"`      // 是否可编辑
}

type ProductAttributeProductPackage struct {
	Uuid uint64 `json:"uuid"` // 商品包UUID
	Name string `json:"name"` // 商品包名称
}

type ProductAttributeProductPackageList struct {
	List []ProductAttributeProductPackage `json:"list"`
}

// ProductFlavorListResp 商品规格列表响应
type ProductFlavorListResp struct {
	List []ProductFlavorItemResp `json:"list"`
	Meta dto.PageResponse        `json:"meta"`
}

// ProductFlavorItemResp 商品规格列表项
type ProductFlavorItemResp struct {
	Uuid                uint64 `json:"uuid"`                  // 商品规格UUID
	Name                string `json:"name"`                  // 商品规格名称
	Sort                int    `json:"sort"`                  // 商品规格排序
	ProductPackageCount int    `json:"product_package_count"` // 关联商品包数量
}

// ProductFlavorDetailResp 商品规格详情
type ProductFlavorDetailResp struct {
	Uuid                uint64                              `json:"uuid"`                  // 商品规格UUID
	LocaleName          dto.LocaleResponse                  `json:"locale_name"`           // 商品规格名称
	ProductPackageCount int                                 `json:"product_package_count"` // 关联商品包数量
	ProductPackageList  ProductFlavorProductPackageListResp `json:"product_package_list"`  // 关联商品包列表
}

// ProductFlavorProductPackage 商品规格关联商品包
type ProductFlavorProductPackageResp struct {
	Uuid    uint64  `json:"uuid"`     // 商品包UUID
	BomUuid uint64  `json:"bom_uuid"` // 商品BOM UUID
	Name    string  `json:"name"`     // 商品包名称
	Price   float64 `json:"price"`    // 商品包价格
}

// ProductFlavorProductPackageList 商品规格关联商品包列表
type ProductFlavorProductPackageListResp struct {
	List []ProductFlavorProductPackageResp `json:"list"`
}

// ProductImportResp 导入商品响应
type ProductImportResp struct {
	List         []ProductImportListItem     `json:"list" binding:"required,dive"`      // 商品列表
	CategoryList []ProductCategory           `json:"category_list" binding:"required"`  // 分类列表
	UnitList     []ProductImportUnitListItem `json:"unit_list" binding:"required,dive"` // 单位列表
	SkuList      []ProductImportSkuListItem  `json:"sku_list" binding:"required,dive"`  // 规格列表
	TaxList      []ProductImportTaxListItem  `json:"tax_list" binding:"required,dive"`  // 税类列表
}

// ProductImportListItem 导入商品列表项
type ProductImportListItem struct {
	LocaleName            dto.LocaleResponse `json:"locale_name" binding:"required"`              // 商品名称
	CategoryName          string             `json:"category_name" binding:"required"`            // 分类名称
	ProductUnit           string             `json:"product_unit" binding:"required"`             // 商品单位
	SpecName              string             `json:"spec_name" binding:"required"`                // 规格名称
	ProductPrice          float64            `json:"product_price" binding:"required"`            // 商品价格
	NumType               int                `json:"num_type" binding:"required"`                 // 数量计算方法, 1-整数 2-小数
	Barcode               string             `json:"barcode" binding:"required"`                  // 商品条码
	ProductStatus         int                `json:"product_status" binding:"required"`           // 商品状态, 1-上架 0-下架
	ProductRatingTaxType  string             `json:"product_rating_tax_type" binding:"required"`  // 堂食税类
	ProductTakeoutTaxType string             `json:"product_takeout_tax_type" binding:"required"` // 外带税类
	DeductStockType       int                `json:"deduct_stock_type" binding:"required"`        // 库存计算方式, 2-付款减库存 1-下单减库存
	Shows                 string             `json:"shows" binding:"required"`                    // 显示：123456
	IsEnableGrade         int                `json:"is_enable_grade" binding:"required"`          // 是否开启会员折扣(1开启 0关闭)
	OpenOverallDiscount   int                `json:"open_overall_discount" binding:"required"`    // 整单折扣(1开启 0关闭)
	Row                   int                `json:"row" binding:"required"`                      // excel表的行编号
	IsShowCashier         bool               `json:"is_show_cashier"`                             // 是否显示在收银端 1-显示 2-不显示
	IsShowTablet          bool               `json:"is_show_tablet"`                              // 是否显示在平板端 1-显示 2-不显示
	IsShowKitchen         bool               `json:"is_show_kitchen"`                             // 是否显示在送厨端 1-显示 2-不显示
	IsShowAssistant       bool               `json:"is_show_assistant"`                           // 是否显示在点餐助手 1-显示 2-不显示
	IsShowH5              bool               `json:"is_show_h5"`                                  // 是否显示在h5 1-显示 2-不显示
	IsShowDelivery        bool               `json:"is_show_delivery"`                            // 是否显示在外送 1-显示 2-不显示
	UnitUuid              uint64             `json:"unit_uuid"`                                   // 单位UUID
	CategoryUuid          uint64             `json:"category_uuid"`                               // 分类UUID
	SkuUuid               uint64             `json:"sku_uuid"`                                    // 规格UUID
	DineTaxUuid           uint64             `json:"dine_tax_uuid"`                               // 堂食税类UUID
	TakeoutTaxUuid        uint64             `json:"takeout_tax_uuid"`                            // 外带税类UUID
	LocaleNameIsExist     dto.LocaleResponse `json:"locale_name_is_exist"`                        // 商品名称是否存在，对应的key不为空则表示存在
	BarcodeIsExist        bool               `json:"barcode_is_exist"`                            // 条形码是否存在，存在则不保存
}

// ProductImportUnitListItem 导入商品单位列表项
type ProductImportUnitListItem struct {
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 单位名称
	Uuid       uint64             `json:"uuid" binding:"required"`        // 单位UUID
}

// ProductImportSkuListItem 导入商品规格列表项
type ProductImportSkuListItem struct {
	LocaleName dto.LocaleResponse `json:"locale_name" binding:"required"` // 规格名称
	Uuid       uint64             `json:"uuid" binding:"required"`        // 规格UUID
}

// ProductImportTaxListItem 导入商品税类列表项
type ProductImportTaxListItem struct {
	Name string `json:"name" binding:"required"` // 税类名称
	Uuid uint64 `json:"uuid" binding:"required"` // 税类UUID
}

// ProductSingleListResp 单规格商品列表响应
type ProductSingleListResp struct {
	List []ProductSingleListItemResp `json:"list"`
	Meta dto.PageResponse            `json:"meta"`
}

// ProductSingleListItemResp 单规格商品列表项响应
type ProductSingleListItemResp struct {
	Uuid               uint64             `json:"uuid"`                  // 商品规格UUID
	Name               dto.LocaleResponse `json:"locale_name"`           // 商品名称
	FlavorName         dto.LocaleResponse `json:"flavor_name"`           // 商品规格名称
	CategoryUuid       uint64             `json:"category_uuid"`         // 商品分类UUID
	ProductBomCardUuid uint64             `json:"product_bom_card_uuid"` // 成本卡UUID，0表示没有成本卡
	ProductBomCardName dto.LocaleResponse `json:"product_bom_card_name"` // 成本卡名称
}

type ProductPackageSubProduct struct {
	Uuid             uint64             `json:"uuid"`               // 套餐子商品UUID
	BomUuid          uint64             `json:"bom_uuid"`           // 商品BOM UUID
	ProductUuid      uint64             `json:"product_uuid"`       // 商品UUID
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 套餐子商品名称
	FlavorLocaleName dto.LocaleResponse `json:"flavor_locale_name"` // 商品规格名称
	Num              float64            `json:"num"`                // 套餐子商品数量
	Price            float64            `json:"price"`              // 套餐子商品价格
}

type ProductPackageSubProductList struct {
	List []ProductPackageSubProduct `json:"list"`
}

type ProductPackageSubProductGroup struct {
	Uuid       uint64                       `json:"uuid"`        // 套餐子商品分组UUID
	LocaleName dto.LocaleResponse           `json:"locale_name"` // 套餐子商品分组名称
	Products   ProductPackageSubProductList `json:"products"`    // 套餐子商品列表
}

type ProductPackageSubProductGroupList struct {
	List []ProductPackageSubProductGroup `json:"list"`
}

// ProductDetailResp 商品详情响应
type ProductDetailResp struct {
	ProductType  uint               `json:"product_type"`  // 商品类型 0-商品 1-套餐
	Uuid         uint64             `json:"uuid"`          // 商品UUID
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 商品名称
	CategoryUuid uint64             `json:"category_uuid"` // 商品分类UUID
	CategoryName string             `json:"category_name"` // 商品分类名称
	UnitUuid     uint64             `json:"unit_uuid"`     // 商品单位UUID
	UnitName     string             `json:"unit_name"`     // 商品单位名称
	Price        *float64           `json:"price"`         // 商品价格,套餐的价格

	Flavors                 ProductFlavorList                 `json:"flavors"`                    // 商品规格列表
	Sauces                  ProductSauceList                  `json:"sauces"`                     // 商品小料列表
	AttributeGroups         ProductAttributeGroupList         `json:"attribute_groups"`           // 商品属性组列表
	PackageSubProductGroups ProductPackageSubProductGroupList `json:"package_sub_product_groups"` // 套餐子商品分组列表

	TakeoutTaxUuid uint64 `json:"takeout_tax_uuid"` // 外带税类UUID
	TakeoutTaxName string `json:"takeout_tax_name"` // 外带税类名称
	DineTaxUuid    uint64 `json:"rating_tax_uuid"`  // 堂食税类UUID
	DineTaxName    string `json:"rating_tax_name"`  // 堂食税类名称

	Status          uint   `json:"status"`            // 商品状态 0-下架 1-上架
	ImageFileUuid   uint64 `json:"image_file_uuid"`   // 商品图片UUID
	Image           string `json:"image"`             // 商品图片。
	NumType         *uint  `json:"num_type"`          // 数量计算方法 0-整数 1-小数
	DeductStockType uint   `json:"deduct_stock_type"` // 库存计算方式,0-结账减库存 1-下单减库存

	IsShowCashier   bool `json:"is_show_cashier"`   // 是否显示在收银端 1-显示 0-不显示
	IsShowTablet    bool `json:"is_show_tablet"`    // 是否显示在平板端 1-显示 0-不显示
	IsShowKitchen   bool `json:"is_show_kitchen"`   // 是否显示在厨显端显示：1-显示；0-不显示
	IsShowAssistant bool `json:"is_show_assistant"` // 是否显示在点餐助手 1-显示 0-不显示
	IsShowH5        bool `json:"is_show_h5"`        // 是否显示在h5 1-显示 0-不显示
	IsShowDelivery  bool `json:"is_show_delivery"`  // 是否显示在外送 1-显示 0-不显示

	OpenDiscount        bool `json:"open_discount"`         // 是否开启会员折扣 1-开启 0-关闭
	OpenOverallDiscount bool `json:"open_overall_discount"` // 整单折扣 1-开启 0-关闭

	SauceRequired     bool `json:"sauce_required"`      // 是否必选小料 1-是 0-否
	SauceMaxSelection uint `json:"sauce_max_selection"` // 小料最大选择数量
}

// ProductListResp 商品列表响应
type ProductShopListResp struct {
	List []ProductShopListItemResp `json:"list"`
	Meta dto.PageResponse          `json:"meta"`
}

// ProductShopListItemResp 商品列表项响应
type ProductShopListItemResp struct {
	Uuid                uint64                            `json:"uuid"`                  // 商品UUID
	LocaleName          dto.LocaleResponse                `json:"locale_name"`           // 商品名称
	Image               string                            `json:"image"`                 // 商品图片
	Tag                 ProductShopListItemTagResp        `json:"tag"`                   // 商品标签列表
	MinPrice            float64                           `json:"min_price"`             // 最低价格
	MaxPrice            float64                           `json:"max_price"`             // 最高价格
	Unit                ProductShopListItemUnitResp       `json:"unit"`                  // 商品单位
	CategoryUuid        uint64                            `json:"category_uuid"`         // 商品分类UUID
	SpecialCategoryUuid uint64                            `json:"special_category_uuid"` // 商品特殊分类UUID
	Status              int                               `json:"status"`                // 商品状态 0-下架 1-上架
	IsSoldOut           bool                              `json:"is_sold_out"`           // 是否售罄 false-否 true-是
	ProductType         int                               `json:"product_type"`          // 商品类型 0-商品 1-套餐
	Sort                int                               `json:"sort"`                  // 商品排序
	Flavors             ProductShopListItemFlavorListResp `json:"flavors"`               // 商品规格列表
	NumType             uint                              `json:"num_type"`              // 商品数量计算方法 0-整数 1-小数
}

// ProductShopListItemTagResp 商品标签列表
type ProductShopListItemTagResp struct {
	IsMultipleSpec bool `json:"is_multiple_spec"` // 是否多规格
	IsAttribute    bool `json:"is_attribute"`     // 是否属性
	IsSauce        bool `json:"is_sauce"`         // 是否加料
}

type ProductShopListItemUnitResp struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品单位名称
}

type ProductShopListItemFlavorListResp struct {
	List []ProductShopListItemFlavorItemResp `json:"list"`
}

type ProductShopListItemFlavorItemResp struct {
	Uuid       uint64             `json:"uuid"`        // 商品Bom uuid
	LocaleName dto.LocaleResponse `json:"locale_name"` // 商品Bom名称
	Price      float64            `json:"price"`       // 商品Bom价格
}

// ProductTaxListResp 商品税类列表响应
type ProductTaxListResp struct {
	List []ProductTaxItemResp `json:"list"`
}

// ProductTaxItemResp 商品税类列表项响应
type ProductTaxItemResp struct {
	Uuid uint64  `json:"uuid"` // 税类UUID
	Name string  `json:"name"` // 税类名称
	Rate float64 `json:"rate"` // 税率
}

// ProductDeleteResp 商品删除响应
type ProductDeleteResp struct {
	List []string `json:"list"`
}

// ProductEditResp 商品编辑响应
type ProductEditResp struct {
	List []string `json:"list"`
}

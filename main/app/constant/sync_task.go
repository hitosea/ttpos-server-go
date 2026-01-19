package constant

// 同步任务状态
const (
	SyncTaskStatusRunning uint8 = 0 // 进行中
	SyncTaskStatusSuccess uint8 = 1 // 已完成
	SyncTaskStatusFailed  uint8 = 2 // 失败
)

// 同步任务明细状态
const (
	SyncTaskItemStatusPending uint8 = 0 // 待执行
	SyncTaskItemStatusRunning uint8 = 1 // 执行中
	SyncTaskItemStatusSuccess uint8 = 2 // 已完成
	SyncTaskItemStatusFailed  uint8 = 3 // 失败
)

// 同步任务类型
const (
	SyncTaskTypeProductCategory   = "product_category"   // 商品分类
	SyncTaskTypeMaterialCategory  = "material_category"  // 物品分类
	SyncTaskTypeTax               = "tax"                // 税类
	SyncTaskTypeUnit              = "unit"               // 单位
	SyncTaskTypeWarehouse         = "warehouse"          // 仓库
	SyncTaskTypeMaterial          = "material"           // 物品
	SyncTaskTypeFlavor            = "flavor"             // 规格
	SyncTaskTypeAttribute         = "attribute"          // 属性
	SyncTaskTypeSauce             = "sauce"              // 加料
	SyncTaskTypeProduct           = "product"            // 商品
	SyncTaskTypeProductStock      = "product_stock"      // 商品库存
	SyncTaskTypeBomCard           = "bom_card"           // 成本卡
	SyncTaskTypeSupplier          = "supplier"           // 供应商
	SyncTaskTypeWarehouseStock    = "warehouse_stock"    // 仓库物品库存
	SyncTaskTypePackageImage      = "package_image"      // 商品图片
	SyncTaskTypeMultiLanguage     = "multi_language"     // 多语言
	SyncTaskTypeCoupon            = "coupon"             // 优惠券
	SyncTaskTypeFullReduction     = "full_reduction"     // 满额减
	SyncTaskTypeProductLabel      = "product_label"      // 菜品标签
	SyncTaskTypeMarketingActivity = "marketing_activity" // 营销活动
	SyncTaskTypePaymentMethod     = "payment_method"     // 支付方式
)

// SyncTaskTypeNames 任务类型名称映射
var SyncTaskTypeNames = map[string]string{
	SyncTaskTypeProductCategory:   "商品分类",
	SyncTaskTypeMaterialCategory:  "物品分类",
	SyncTaskTypeTax:               "税类",
	SyncTaskTypeUnit:              "单位",
	SyncTaskTypeWarehouse:         "仓库",
	SyncTaskTypeMaterial:          "物品",
	SyncTaskTypeFlavor:            "规格",
	SyncTaskTypeAttribute:         "属性",
	SyncTaskTypeSauce:             "加料",
	SyncTaskTypeProduct:           "商品",
	SyncTaskTypeProductStock:      "商品库存",
	SyncTaskTypeBomCard:           "成本卡",
	SyncTaskTypeSupplier:          "供应商",
	SyncTaskTypeWarehouseStock:    "仓库物品库存",
	SyncTaskTypePackageImage:      "商品图片",
	SyncTaskTypeMultiLanguage:     "多语言",
	SyncTaskTypeCoupon:            "优惠券",
	SyncTaskTypeFullReduction:     "满额减",
	SyncTaskTypeProductLabel:      "菜品标签",
	SyncTaskTypeMarketingActivity: "营销活动",
	SyncTaskTypePaymentMethod:     "支付方式",
}

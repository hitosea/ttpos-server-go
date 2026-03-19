package constant

import "slices"

// HQ 控制模式
const (
	HqControlSeparate = 0 // 分开控制：子店可独立编辑
	HqControlUnified  = 1 // 统一控制：子店跟从总部
)

// HQ 推送实体类型
const (
	HqEntityProduct        = "product"
	HqEntityProductTakeout = "product_takeout"
	HqEntityMaterial       = "material"
)

// HQ 推送字段类型
const (
	HqFieldDineShelf     = "dine_shelf"     // 店内上下架
	HqFieldTakeoutShelf  = "takeout_shelf"  // 外卖上下架
	HqFieldTakeoutPrice  = "takeout_price"  // 外卖价格（含规格）
	HqFieldSafetyStock   = "safety_stock"   // 安全库存
	HqFieldNegativeStock = "negative_stock" // 负库存
)

// HQ 批量推送支持的类型
var HqBatchPushFieldTypes = []string{
	HqFieldDineShelf,
	HqFieldTakeoutShelf,
	HqFieldTakeoutPrice,
	HqFieldNegativeStock,
}

// IsValidBatchPushFieldType 检查是否为合法的批量推送字段类型
func IsValidBatchPushFieldType(fieldType string) bool {
	return slices.Contains(HqBatchPushFieldTypes, fieldType)
}

// IsValidHqFieldType 检查是否为合法的 HQ 字段类型
func IsValidHqFieldType(fieldType string) bool {
	switch fieldType {
	case HqFieldDineShelf, HqFieldTakeoutShelf, HqFieldTakeoutPrice, HqFieldSafetyStock, HqFieldNegativeStock:
		return true
	}
	return false
}

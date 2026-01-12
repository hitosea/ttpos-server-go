package value_object

// 外卖平台代码
const (
	TakeoutPlatformGrab    = "grab"    // Grab
	TakeoutPlatformLineman = "lineman" // LINE MAN
)

// 外卖平台名称映射
var TakeoutPlatformNames = map[string]string{
	TakeoutPlatformGrab:    "Grab",
	TakeoutPlatformLineman: "LINE MAN",
}

// 外卖类型
const (
	TakeoutTypeGrab    = 1 // Grab
	TakeoutTypeLineman = 2 // LINE MAN
)

// 外卖订单状态
const (
	TakeoutOrderStatePending         = 0  // 待接单
	TakeoutOrderStateAccepted        = 10 // 已接单配餐中
	TakeoutOrderStateRiderPending    = 20 // 待骑手接单
	TakeoutOrderStateRiderProcessing = 30 // 骑手配送中
	TakeoutOrderStateCompleted       = 40 // 已完成
	TakeoutOrderStateRejected        = 50 // 已拒单
	TakeoutOrderStateCanceled        = 60 // 已取消
)

// 库存状态
const (
	TakeoutStockStatusSufficient   = 1 // 充足
	TakeoutStockStatusInsufficient = 2 // 不足
)

// 库存数量
const (
	TakeoutStockUnlimited = 99999999 // 无限库存（用于表示库存充足或不限制库存）
)

// 接单类型
const (
	TakeoutOrderAcceptedTypeAuto   = "AUTO"   // 自动接单
	TakeoutOrderAcceptedTypeManual = "MANUAL" // 手动接单
)

// TTPOS 通用订单类型
const (
	TakeoutOrderTypeDelivery           = "DELIVERY"            // 平台配送（如 Grab 配送）
	TakeoutOrderTypeRestaurantDelivery = "RESTAURANT_DELIVERY" // 商家配送
	TakeoutOrderTypeTakeaway           = "TAKEAWAY"            // 自提/打包
	TakeoutOrderTypeDineIn             = "DINEIN"              // 店内就餐
)

// Grab 平台订单类型
const (
	GrabOrderTypeDeliveredByGrab       = "DeliveredByGrab"       // Grab 配送
	GrabOrderTypeDeliveredByRestaurant = "DeliveredByRestaurant" // 商家配送
	GrabOrderTypeTakeAway              = "TakeAway"              // 自提
	GrabOrderTypePickup                = "Pickup"                // 自提（别名）
	GrabOrderTypeDineIn                = "DineIn"                // 店内就餐
)

// ConvertGrabOrderTypeToTakeoutOrderType 转换 Grab 订单类型为 TTPOS 通用订单类型
func ConvertGrabOrderTypeToTakeoutOrderType(grabOrderType string) string {
	switch grabOrderType {
	case GrabOrderTypeDeliveredByGrab:
		return TakeoutOrderTypeDelivery
	case GrabOrderTypeDeliveredByRestaurant:
		return TakeoutOrderTypeRestaurantDelivery
	case GrabOrderTypeTakeAway, GrabOrderTypePickup:
		return TakeoutOrderTypeTakeaway
	case GrabOrderTypeDineIn:
		return TakeoutOrderTypeDineIn
	default:
		return TakeoutOrderTypeDelivery // 默认返回配送
	}
}

// ShouldActivateShopOnEnable 判断平台是否需要在启用时激活门店外卖渠道
// 参数：platform - 平台标识
// 返回：true - 需要激活；false - 不需要激活
func ShouldActivateShopOnEnable(platform string) bool {
	// Lineman 平台需要在启用时调用 ActivateShop 接口
	return platform == TakeoutPlatformLineman
}

// GetTakeoutTypeByPlatform 根据平台名称获取外卖类型
func GetTakeoutTypeByPlatform(platform string) int {
	switch platform {
	case TakeoutPlatformGrab:
		return TakeoutTypeGrab
	case TakeoutPlatformLineman:
		return TakeoutTypeLineman
	default:
		return TakeoutTypeGrab
	}
}

// GetPlatformName 根据平台代码获取平台显示名称
func GetPlatformName(platform string) string {
	if name, ok := TakeoutPlatformNames[platform]; ok {
		return name
	}
	return ""
}

package value_object

// 外卖平台代码
const (
	TakeoutPlatformGrab       = "grab"       // Grab
	TakeoutPlatformLineman    = "lineman"    // LINE MAN
	TakeoutPlatformShopeefood = "shopeefood" // ShopeeFood
)

// 外卖平台名称映射
var TakeoutPlatformNames = map[string]string{
	TakeoutPlatformGrab:       "Grab",
	TakeoutPlatformLineman:    "LINE MAN",
	TakeoutPlatformShopeefood: "ShopeeFood",
}

// 外卖订单状态
const (
	TakeoutOrderStatePending    = 1 // 待接单
	TakeoutOrderStateAccepted   = 2 // 已接单
	TakeoutOrderStateProcessing = 3 // 制作中
	TakeoutOrderStateCompleted  = 4 // 已完成
	TakeoutOrderStateRejected   = 5 // 已拒单
)

// 库存状态
const (
	TakeoutStockStatusSufficient   = 1 // 充足
	TakeoutStockStatusInsufficient = 2 // 不足
)

// 接单类型
const (
	TakeoutOrderAcceptedTypeAuto   = "AUTO"   // 自动接单
	TakeoutOrderAcceptedTypeManual = "MANUAL" // 手动接单
)

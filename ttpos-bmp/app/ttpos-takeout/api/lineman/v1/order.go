package v1

import "github.com/gogf/gf/v2/frame/g"

// ========== 嵌套结构体定义 ==========

// OrderItem 订单商品项
// 订单中的单个商品信息，包含商品ID、数量、价格、选项等
type OrderItem struct {
	Id          string              `json:"id" v:"required#菜单项ID不能为空" dc:"菜单项 ID，最大长度 255"`
	Quantity    int                 `json:"quantity" v:"required|min:1#商品数量不能为空|商品数量不能小于1" dc:"商品数量"`
	UnitPrice   float64             `json:"unitPrice" v:"required|min:0#商品单价不能为空|商品单价不能小于0" dc:"商品单价（THB），包含额外选项费用，已应用促销折扣"`
	Memo        string              `json:"memo" dc:"顾客备注"`
	PromotionId string              `json:"promotionId" dc:"促销活动 ID（如有），最大长度 255"`
	Discount    float64             `json:"discount" dc:"促销折扣金额（商户补贴）"`
	Properties  []OrderItemProperty `json:"properties" dc:"商品选项列表"`
}

// OrderItemProperty 商品属性
// 商品的自定义选项，如规格、加料等
type OrderItemProperty struct {
	Id     string                   `json:"id" v:"required#选项ID不能为空" dc:"选项 ID，最大长度 255"`
	Values []OrderItemPropertyValue `json:"values" v:"required|min-length:1#选项值列表不能为空|选项值列表至少包含一项" dc:"已选择的选项值"`
}

// OrderItemPropertyValue 商品属性值
// 商品选项的具体值和价格
type OrderItemPropertyValue struct {
	Id    string  `json:"id" v:"required#选项值ID不能为空" dc:"选项值 ID，最大长度 255"`
	Price float64 `json:"price" v:"required|min:0#选项值价格不能为空|选项值价格不能小于0" dc:"选项值价格（THB）"`
}

// OrderAdditionalItem 订单附加项
// 订单的额外信息，如备注、特殊要求等
type OrderAdditionalItem struct {
	Name string `json:"name" v:"required#附加信息不能为空" dc:"附加信息，如：ไม่รับช้อนส้อมพลาสติก，最大长度 1024"`
}

// ========== 订单创建 API ==========

// PlaceOrderReq 订单创建请求
// LINE MAN 向 TTPOS 发送新订单创建请求
type PlaceOrderReq struct {
	g.Meta            `path:"/partners/:partnerId/stores/:storeId/orders" method:"post" tags:"LINE MAN Order" summary:"接收订单创建通知"`
	PartnerId         string                `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID（路径参数）"`
	StoreId           string                `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID（路径参数）"`
	OrderId           string                `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度为1-20个字符" dc:"订单唯一 ID，格式：LMF-yyMMdd-{generated number}"`
	OrderShortCode    string                `json:"orderShortCode" v:"required|length:4,4#短订单ID不能为空|短订单ID长度必须为4个字符" dc:"短订单 ID，为 orderId 的后四位"`
	RestaurantRevenue float64               `json:"restaurantRevenue" v:"required|min:0#商户收入总额不能为空|商户收入总额不能小于0" dc:"商户收入总额（已扣除合作伙伴补贴折扣）"`
	OrderAcceptedTime string                `json:"orderAcceptedTime" v:"required#订单接受时间不能为空" dc:"订单接受时间，ISO 8601 格式，如：2022-11-01T13:08:06+07:00"`
	Items             []OrderItem           `json:"items" v:"required|min-length:1#订单商品列表不能为空|订单商品列表至少包含一项" dc:"订单商品列表"`
	AdditionalItems   []OrderAdditionalItem `json:"additionalItems" v:"required" dc:"订单附加项列表"`
	MemberId          string                `json:"memberId" dc:"绑定 LINE MAN 账号的会员 ID，最大长度 255"`
	CustomerType      string                `json:"customerType" v:"required|in:DELIVERY,PICKUP#订单类型不能为空|订单类型必须为DELIVERY或PICKUP" dc:"订单类型：DELIVERY（外送）或 PICKUP（自取）"`
}

// PlaceOrderRes 订单创建响应
// TTPOS 返回给 LINE MAN 的订单创建结果
type PlaceOrderRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}

// ========== 订单状态更新 API ==========

// OrderStatusUpdateReq 订单状态更新请求
// LINE MAN 通知 TTPOS 订单状态变化（完成或取消）
type OrderStatusUpdateReq struct {
	g.Meta      `path:"/partners/:partnerId/stores/:storeId/order/status" method:"post" tags:"LINE MAN Order" summary:"接收订单状态更新通知"`
	PartnerId   string `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID（路径参数）"`
	StoreId     string `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID（路径参数）"`
	OrderId     string `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度为1-20个字符" dc:"订单唯一 ID，格式：LMF-yyMMdd-{generated number}"`
	OrderStatus string `json:"orderStatus" v:"required|in:FINISH,CANCELED#订单状态不能为空|订单状态必须为FINISH或CANCELED" dc:"订单状态：FINISH（已完成）或 CANCELED（已取消）"`
}

// OrderStatusUpdateRes 订单状态更新响应
// TTPOS 返回给 LINE MAN 的订单状态更新结果
type OrderStatusUpdateRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}

// ========== 订单更新通知 API ==========

// OrderUpdateReq 订单更新通知请求
// LINE MAN 通知 TTPOS 订单信息已更新
type OrderUpdateReq struct {
	g.Meta            `path:"/partners/:partnerId/stores/:storeId/orders/notification" method:"put" tags:"LINE MAN Order" summary:"接收订单更新通知"`
	PartnerId         string                `json:"partnerId" v:"required#合作伙伴ID不能为空" dc:"合作伙伴唯一 ID（路径参数）"`
	StoreId           string                `json:"storeId" v:"required#门店ID不能为空" dc:"门店唯一 ID（路径参数）"`
	OrderId           string                `json:"orderId" v:"required|length:1,20#订单ID不能为空|订单ID长度为1-20个字符" dc:"订单唯一 ID，格式：LMF-yyMMdd-{generated number}"`
	OrderShortCode    string                `json:"orderShortCode" v:"required|length:4,4#短订单ID不能为空|短订单ID长度必须为4个字符" dc:"短订单 ID，为 orderId 的后四位"`
	RestaurantRevenue float64               `json:"restaurantRevenue" v:"required|min:0#商户收入总额不能为空|商户收入总额不能小于0" dc:"商户收入总额（已扣除合作伙伴补贴折扣）"`
	OrderAcceptedTime string                `json:"orderAcceptedTime" v:"required#订单接受时间不能为空" dc:"订单接受时间，ISO 8601 格式"`
	OrderUpdatedTime  string                `json:"orderUpdatedTime" v:"required#订单更新时间不能为空" dc:"订单更新时间，ISO 8601 格式"`
	Items             []OrderItem           `json:"items" v:"required|min-length:1#订单商品列表不能为空|订单商品列表至少包含一项" dc:"订单商品列表"`
	AdditionalItems   []OrderAdditionalItem `json:"additionalItems" v:"required" dc:"订单附加项列表"`
	MemberId          string                `json:"memberId" dc:"绑定 LINE MAN 账号的会员 ID，最大长度 255"`
	CustomerType      string                `json:"customerType" v:"required|in:DELIVERY,PICKUP#订单类型不能为空|订单类型必须为DELIVERY或PICKUP" dc:"订单类型：DELIVERY（外送）或 PICKUP（自取）"`
}

// OrderUpdateRes 订单更新通知响应
// TTPOS 返回给 LINE MAN 的订单更新结果
type OrderUpdateRes struct {
	g.Meta `mime:"application/json"`
	LinemanCommonResData
}

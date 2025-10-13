package event

import (
	"sync"
	"ttpos-server-go/pkg/eventbus"
)

// EventName 定义了事件名称的类型，用于标识不同的事件。
type EventName string

const (
	// =============================================================================
	// H5订单事件
	// =============================================================================
	EventAcceptH5Order EventName = "Event_Accept_H5_Order" // 接单事件
	EventRejectH5Order EventName = "Event_Reject_H5_Order" // 拒单事件

	// =============================================================================
	// 会员订单事件
	// =============================================================================
	EventCreateMemberSaleOrder     EventName = "Event_Create_Member_Sale_Order"      // 创建外送订单事件
	EventAcceptMemberSaleOrder     EventName = "Event_Accept_Member_Sale_Order"      // 外送订单商家接单事件
	EventRejectMemberSaleOrder     EventName = "Event_Reject_Member_Sale_Order"      // 拒单会员端销售订单事件
	EventCancelMemberOrder         EventName = "Event_Cancel_Member_Order"           // 整单取消事件
	EventCookFinishMemberSaleOrder EventName = "Event_Cook_Finish_Member_Sale_Order" // 外送订单备餐完成事件
	EventPayFinishMemberSaleOrder  EventName = "Event_Pay_Finish_Member_Sale_Order"  // 支付完成会员端销售订单事件

	// =============================================================================
	// 骑手配送事件
	// =============================================================================
	EventRiderAcceptMemberSaleOrder    EventName = "Event_Rider_Accept_Member_Sale_Order"    // 骑手接单事件
	EventRiderDeliveryMemberSaleOrder  EventName = "Event_Rider_Delivery_Member_Sale_Order"  // 骑手完成取餐事件
	EventRiderCompletedMemberSaleOrder EventName = "Event_Rider_Completed_Member_Sale_Order" // 骑手送达事件

	// =============================================================================
	// 桌台管理事件
	// =============================================================================
	EventOpenDesk              EventName = "Event_Open_Desk"                 // 开台事件
	EventChangeDesk            EventName = "Event_Change_Desk"               // 转台事件
	EventMergeDesk             EventName = "Event_Merge_Desk"                // 并台事件
	EventChangeMealNumSaleBill EventName = "Event_Change_Meal_Num_Sale_Bill" // 修改桌台就餐人数事件

	// =============================================================================
	// 订单操作事件
	// =============================================================================
	EventCheckoutSaleOrder     EventName = "Event_Checkout_Sale_Order"      // 结账事件
	EventCheckoutZeroSaleOrder EventName = "Event_Checkout_Zero_Sale_Order" // 结账手动抹零事件
	EventCancelOrder           EventName = "Event_Cancel_Order"             // 整单取消事件
	EventSplitOrder            EventName = "Event_Split_Order"              // 拆单事件
	EventCancelSplitOrder      EventName = "Event_Cancel_Split_Order"       // 取消拆单事件
	EventReturnOrder           EventName = "Event_Return_Order"             // 用餐订单退款事件
	EventOrderReverseSettle    EventName = "Event_Order_Reverse_Settle"     // 用餐订单反结账事件
	EventFreeSaleOrder         EventName = "Event_Free_Sale_Order"          // 免单事件

	// =============================================================================
	// 商品操作事件
	// =============================================================================
	EventCancelSaleOrderProduct       EventName = "Event_Cancel_Sale_Order_Product"        // 退菜订单商品事件
	EventGiftSaleOrderProduct         EventName = "Event_Gift_Sale_Order_Product"          // 赠菜事件
	EventCancelGiftSaleOrderProduct   EventName = "Event_Cancel_Gift_Sale_Order_Product"   // 取消赠菜事件
	EventChangeSaleOrderProductPrice  EventName = "Event_Change_Sale_Order_Product_Price"  // 修改销售订单产品价格事件
	EventCancelReturnSaleOrderProduct EventName = "Event_Cancel_Return_Sale_Order_Product" // 取消退菜事件
	EventChangeDeskSaleOrderProduct   EventName = "Event_Change_Desk_Sale_Order_Product"   // 转菜事件
	EventCancelDoingProduct           EventName = "Event_Cancel_Doing_Product"             // 取消制作中商品事件

	// =============================================================================
	// 厨房操作事件
	// =============================================================================
	EventSentCooking    EventName = "Event_Sent_Cooking"     // 送厨事件
	EventSentCookingPre EventName = "Event_Sent_Cooking_Pre" // 预送厨事件
	EventFinishMenu     EventName = "Event_Finish_Menu"      // 完成制作事件

	// =============================================================================
	// 打包相关事件
	// =============================================================================
	EventWrapSaleBill           EventName = "Event_Wrap_Sale_Bill"            // 打包事件
	EventUnwrapSaleBill         EventName = "Event_Unwrap_Sale_Bill"          // 取消整单打包事件
	EventWrapSaleOrderProduct   EventName = "Event_Wrap_Sale_Order_Product"   // 商品打包事件
	EventUnwrapSaleOrderProduct EventName = "Event_Unwrap_Sale_Order_Product" // 取消商品打包事件

	// =============================================================================
	// 销售账单事件
	// =============================================================================
	EventShowSaleBill EventName = "Event_Show_Sale_Bill" // 取单事件
	EventHideSaleBill EventName = "Event_Hide_Sale_Bill" // 挂单事件

	// =============================================================================
	// 折扣优惠事件
	// =============================================================================
	EventDiscountSaleOrder            EventName = "Event_Discount_Sale_Order"        // 优惠折扣事件
	EventDiscountZeroSaleOrder        EventName = "Event_Discount_Zero_Sale_Order"   // 订单抹零事件
	EventDiscountChangePriceSaleOrder EventName = "Event_Change_Price_Sale_Order"    // 改价事件
	EventCancelSaleOrderDiscount      EventName = "Event_Cancel_Sale_Order_Discount" // 取消优惠折扣事件

	// =============================================================================
	// 会员相关事件
	// =============================================================================
	EventChangeMemberBalance EventName = "Event_Change_Member_Balance" // 会员余额变动事件
	EventChangeMemberPoints  EventName = "Event_Change_Member_Points"  // 会员积分变动事件

	// =============================================================================
	// 库存管理事件
	// =============================================================================
	EventChangeStock EventName = "Event_Change_Stock" // 加库存事件

	// =============================================================================
	// 统计分析事件
	// =============================================================================
	EventStatisticsSale   EventName = "Event_Statistics_Sale"   // 统计销售事件
	EventStatisticsMember EventName = "Event_Statistics_Member" // 统计会员事件
)

// systemEventBus 是 SystemEventBus 的单例实例，用于全局访问。
// systemBusOnce 确保 systemEventBus 只被初始化一次。
var systemEventBus *SystemEventBus
var systemBusOnce sync.Once

// SystemEventBus 是一个包含事件总线的结构体。
// bus 是一个嵌入的 EventBus 实例，用于处理事件的发布和订阅。
type SystemEventBus struct {
	bus *eventbus.EventBus
}

// NewSystemBus 创建并返回 SystemEventBus 的单例实例。
// 使用 sync.Once 确保 SystemEventBus 只被初始化一次，避免多次创建实例。
// 初始化时，会创建一个新的 EventBus 实例并赋值给 SystemEventBus 的 bus 字段。
func NewSystemBus() *SystemEventBus {
	systemBusOnce.Do(func() {
		systemEventBus = &SystemEventBus{
			bus: eventbus.NewEventBus(),
		}
	})
	return systemEventBus
}

package event

import (
	member "ttpos-server-go/app/event/member"
	order "ttpos-server-go/app/event/order"
	order_h5 "ttpos-server-go/app/event/order_h5"
	product "ttpos-server-go/app/event/product"
	rider "ttpos-server-go/app/event/rider"
	statistics "ttpos-server-go/app/event/statistics"
	takeout "ttpos-server-go/app/event/takeout"
)

// init 初始化事件
func init() {
	// H5订单相关事件处理器
	// 自动注册"接单"事件处理器
	order_h5.AcceptH5OrderEventHandler()
	// 自动注册"拒单"事件处理器
	order_h5.RejectH5OrderEventHandler()

	// 会员相关事件处理器
	// 自动注册"会员余额变动"事件处理器
	member.ChangeMemberBalanceEventHandler()
	// 自动注册"会员积分变动"事件处理器
	member.ChangeMemberPointsEventHandler()
	// 自动注册"创建会员订单"事件处理器
	member.CreateMemberSaleOrderEventHandler()
	// 自动注册"取消会员订单"事件处理器
	member.CancelMemberOrderEventHandler()
	// 自动注册"会员订单制作完成"事件处理器
	member.CookFinishMemberSaleOrderEventHandler()
	// 自动注册"拒绝会员订单"事件处理器
	member.RejectMemberSaleOrderEventHandler()
	// 自动注册"会员订单支付完成"事件处理器
	member.PayFinishMemberSaleOrderEventHandler()
	// 自动注册"接受会员订单"事件处理器
	member.AcceptMemberSaleOrderEventHandler()

	// 订单操作相关事件处理器
	// 自动注册"结账"事件处理器
	order.CheckoutSaleOrderEventHandler()
	// 自动注册"零金额结账"事件处理器
	order.CheckoutZeroSaleOrderEventHandler()
	// 自动注册"取消订单"事件处理器
	order.CancelOrderEventHandler()
	// 自动注册"拆单"事件处理器
	order.SplitOrderEventHandler()
	// 自动注册"取消拆单"事件处理器
	order.CancelSplitOrderEventHandler()
	// 自动注册"预送厨"事件处理器
	order.SentCookingPreEventHandler()
	// 自动注册"开桌"事件处理器
	order.OpenDeskEventHandler()
	// 自动注册"换桌,转台"事件处理器
	order.ChangeDeskEventHandler()
	// 自动注册"并桌,并台"事件处理器
	order.MergeDeskEventHandler()
	// 自动注册"显示账单"事件处理器
	order.ShowSaleBillEventHandler()
	// 自动注册"隐藏账单,挂单"事件处理器
	order.HideSaleBillEventHandler()
	// 自动注册"打包账单"事件处理器
	order.WrapSaleBillEventHandler()
	// 自动注册"取消打包账单"事件处理器
	order.UnwrapSaleBillEventHandler()

	// 商品相关事件处理器
	// 自动注册"取消销售商品"事件处理器
	order.CancelSaleOrderProductEventHandler()
	// 自动注册"取消打包商品"事件处理器
	order.UnwrapSaleOrderProductEventHandler()
	// 自动注册"取消退货商品"事件处理器
	order.CancelReturnSaleOrderProductEventHandler()
	// 自动注册"取消赠送商品"事件处理器
	order.CancelGiftSaleOrderProductEventHandler()
	// 自动注册"打包商品"事件处理器
	order.WrapSaleOrderProductEventHandler()
	// 自动注册"改价商品"事件处理器
	order.ChangeSaleOrderProductPriceEventHandler()
	// 自动注册"赠送商品"事件处理器
	order.GiftSaleOrderProductEventHandler()
	// 自动注册"取消制作商品"事件处理器
	order.CancelDoingProductEventHandler()
	// 自动注册"换桌商品"事件处理器
	order.ChangeDeskSaleOrderProductEventHandler()

	// 优惠相关事件处理器
	// 自动注册"免单"事件处理器
	order.FreeSaleOrderEventHandler()
	// 自动注册"订单打折"事件处理器
	order.DiscountSaleOrderEventHandler()
	// 自动注册"满减活动"事件处理器
	order.ActivitySaleOrderEventHandler()
	// 自动注册"改价打折"事件处理器
	order.DiscountChangePriceSaleOrderEventHandler()
	// 自动注册"零元打折"事件处理器
	order.DiscountZeroSaleOrderEventHandler()
	// 自动注册"取消订单优惠"事件处理器
	order.CancelSaleOrderDiscountEventHandler()

	// 库存和制作相关事件处理器
	// 自动注册"库存变更"事件处理器
	order.ChangeStockEventHandler()
	// 自动注册"完成制作"事件处理器
	order.FinishMenuEventHandler()
	// 自动注册"送厨"事件处理器
	order.SentCookingEventHandler()

	// 商品沽清相关事件处理器
	// 自动注册"商品沽清"事件处理器
	product.ProductSoldOutEventHandler()

	// 其他操作事件处理器
	// 自动注册"修改就餐人数,修改注释"事件处理器
	order.ChangeMealNumSaleBillEventHandler()
	// 自动注册"退单,退款"事件处理器
	order.ReturnOrderEventHandler()
	// 自动注册"反结账"事件处理器
	order.OrderReverseSettleEventHandler()

	// 骑手相关事件处理器
	// 自动注册"骑手配送"事件处理器
	rider.RiderDeliveryMemberSaleOrderEventHandler()
	// 自动注册"骑手完成配送"事件处理器
	rider.RiderCompletedMemberSaleOrderEventHandler()
	// 自动注册"骑手接单"事件处理器
	rider.RiderAcceptMemberSaleOrderEventHandler()

	// 统计相关事件处理器
	// 自动注册"会员统计"事件处理器
	statistics.StatisticsMemberEventHandler()
	// 自动注册"销售统计"事件处理器
	statistics.StatisticsSaleEventHandler()

	// 外卖订单相关事件处理器
	// 自动注册"外卖订单创建"事件处理器
	takeout.TakeoutOrderCreatedEventHandler()
	// 自动注册"外卖订单接单"事件处理器
	takeout.TakeoutOrderAcceptEventHandler()
	// 自动注册"外卖订单取消"事件处理器
	takeout.TakeoutOrderCancelEventHandler()
	// 自动注册"外卖订单拒单"事件处理器
	takeout.TakeoutOrderRejectedEventHandler()
	// 自动注册"外卖订单准备完成"事件处理器
	takeout.TakeoutOrderReadyEventHandler()
	// 自动注册"外卖订单状态更新"事件处理器
	takeout.TakeoutOrderStatusUpdatedEventHandler()
}

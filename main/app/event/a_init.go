package event

// init 初始化事件
func init() {
	// H5订单相关事件处理器
	// 自动注册"接单"事件处理器
	acceptH5OrderEventHandler()
	// 自动注册"拒单"事件处理器
	rejectH5OrderEventHandler()

	// 会员相关事件处理器
	// 自动注册"会员余额变动"事件处理器
	changeMemberBalanceEventHandler()
	// 自动注册"会员积分变动"事件处理器
	changeMemberPointsEventHandler()
	// 自动注册"创建会员订单"事件处理器
	createMemberSaleOrderEventHandler()
	// 自动注册"取消会员订单"事件处理器
	cancelMemberOrderEventHandler()
	// 自动注册"会员订单制作完成"事件处理器
	cookFinishMemberSaleOrderEventHandler()
	// 自动注册"拒绝会员订单"事件处理器
	rejectMemberSaleOrderEventHandler()
	// 自动注册"会员订单支付完成"事件处理器
	payFinishMemberSaleOrderEventHandler()
	// 自动注册"接受会员订单"事件处理器
	acceptMemberSaleOrderEventHandler()

	// 订单操作相关事件处理器
	// 自动注册"结账"事件处理器
	checkoutSaleOrderEventHandler()
	// 自动注册"零金额结账"事件处理器
	checkoutZeroSaleOrderEventHandler()
	// 自动注册"取消订单"事件处理器
	cancelOrderEventHandler()
	// 自动注册"拆单"事件处理器
	splitOrderEventHandler()
	// 自动注册"取消拆单"事件处理器
	cancelSplitOrderEventHandler()
	// 自动注册"预送厨"事件处理器
	sentCookingPreEventHandler()
	// 自动注册"开桌"事件处理器
	openDeskEventHandler()
	// 自动注册"换桌,转台"事件处理器
	changeDeskEventHandler()
	// 自动注册"并桌,并台"事件处理器
	mergeDeskEventHandler()
	// 自动注册"显示账单"事件处理器
	showSaleBillEventHandler()
	// 自动注册"隐藏账单,挂单"事件处理器
	hideSaleBillEventHandler()
	// 自动注册"打包账单"事件处理器
	wrapSaleBillEventHandler()
	// 自动注册"取消打包账单"事件处理器
	unwrapSaleBillEventHandler()

	// 商品相关事件处理器
	// 自动注册"取消销售商品"事件处理器
	cancelSaleOrderProductEventHandler()
	// 自动注册"取消打包商品"事件处理器
	unwrapSaleOrderProductEventHandler()
	// 自动注册"取消退货商品"事件处理器
	cancelReturnSaleOrderProductEventHandler()
	// 自动注册"取消赠送商品"事件处理器
	cancelGiftSaleOrderProductEventHandler()
	// 自动注册"打包商品"事件处理器
	wrapSaleOrderProductEventHandler()
	// 自动注册"改价商品"事件处理器
	changeSaleOrderProductPriceEventHandler()
	// 自动注册"赠送商品"事件处理器
	giftSaleOrderProductEventHandler()
	// 自动注册"取消制作商品"事件处理器
	cancelDoingProductEventHandler()
	// 自动注册"换桌商品"事件处理器
	changeDeskSaleOrderProductEventHandler()

	// 优惠相关事件处理器
	// 自动注册"免单"事件处理器
	freeSaleOrderEventHandler()
	// 自动注册"订单打折"事件处理器
	discountSaleOrderEventHandler()
	// 自动注册"改价打折"事件处理器
	discountChangePriceSaleOrderEventHandler()
	// 自动注册"零元打折"事件处理器
	discountZeroSaleOrderEventHandler()
	// 自动注册"取消订单优惠"事件处理器
	cancelSaleOrderDiscountEventHandler()

	// 库存和制作相关事件处理器
	// 自动注册"库存变更"事件处理器
	changeStockEventHandler()
	// 自动注册"完成制作"事件处理器
	finishMenuEventHandler()
	// 自动注册"送厨"事件处理器
	sentCookingEventHandler()

	// 其他操作事件处理器
	// 自动注册"修改就餐人数,修改注释"事件处理器
	changeMealNumSaleBillEventHandler()
	// 自动注册"退单,退款"事件处理器
	returnOrderEventHandler()
	// 自动注册"反结账"事件处理器
	orderReverseSettleEventHandler()

	// 骑手相关事件处理器
	// 自动注册"骑手配送"事件处理器
	riderDeliveryMemberSaleOrderEventHandler()
	// 自动注册"骑手完成配送"事件处理器
	riderCompletedMemberSaleOrderEventHandler()
	// 自动注册"骑手接单"事件处理器
	riderAcceptMemberSaleOrderEventHandler()

	// 统计相关事件处理器
	// 自动注册"会员统计"事件处理器
	statisticsMemberEventHandler()
	// 自动注册"销售统计"事件处理器
	statisticsSaleEventHandler()
}

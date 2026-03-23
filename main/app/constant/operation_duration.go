package constant

// 操作耗时跟踪 Action 类型常量
// 用于 DurationTracker 记录各种操作的耗时
const (
	// 桌台操作
	ActionCreateDeskOrder = "create_desk_order" // 创建桌台订单
	ActionCloseDesk       = "close_desk"        // 关台
	ActionCompleteDesk    = "complete_desk"     // 完成桌台
	ActionChangeDesk      = "change_desk"       // 换台
	ActionMergeDesk       = "merge_desk"        // 并台
	ActionCancelDeskOrder = "cancel_desk_order" // 取消桌台订单

	// 点餐操作
	ActionCreateInstantOrder = "create_instant_order" // 创建点餐订单
	ActionCancelOrder        = "cancel_order"         // 取消订单
	ActionHideOrder          = "hide_order"           // 挂单
	ActionShowOrder          = "show_order"           // 取单
	ActionOrderTakeout       = "order_takeout"        // 打包

	// 订单设置
	ActionSetOrderSource = "set_order_source" // 设置订单来源
	ActionSetNationality = "set_nationality"  // 设置国籍

	// 订单商品操作
	ActionOrderProductDelete      = "order_product_delete"       // 删除订单商品
	ActionOrderProductChangePrice = "order_product_change_price" // 订单商品改价
	ActionOrderProductRemark      = "order_product_remark"       // 订单商品备注
	ActionOrderRemark             = "order_remark"               // 整单备注

	// 订单折扣
	ActionOrderDiscount       = "order_discount"        // 订单折扣
	ActionOrderDiscountCancel = "order_discount_cancel" // 取消订单折扣

	// 订单人数/自助餐
	ActionOrderChangePopulation  = "order_change_population"   // 修改人数
	ActionOrderChangeBuffet      = "order_change_buffet"       // 修改自助餐
	ActionOrderChangeBuffetClock = "order_change_buffet_clock" // 修改自助餐计时

	// 购物车商品操作
	ActionCartProductAdd          = "cart_product_add"              // 添加商品
	ActionCartPackageAdd          = "cart_package_add"              // 添加套餐
	ActionCartProductPackageAdd   = "cart_product_package_add"      // 添加套餐（桌台）
	ActionCartProductFlavorChange = "cart_product_flavor_change"    // 修改规格属性
	ActionCartProductNum          = "cart_product_num"              // 修改商品数量
	ActionCartProductCooking      = "cart_product_cooking"          // 送厨
	ActionCartProductBatchCooking = "cart_product_batch_cooking"    // 分批送厨
	ActionCartProductReturning    = "cart_product_returning"        // 退菜
	ActionCartProductCancelReturn = "cart_product_cancel_returning" // 取消退菜
	ActionCartProductGiving       = "cart_product_giving"           // 赠菜
	ActionCartProductCancelGiving = "cart_product_cancel_giving"    // 取消赠菜
	ActionCartProductChangeDesk   = "cart_product_change_desk"      // 商品换台
	ActionCartProductWrap         = "cart_product_wrap"             // 打包商品
	ActionCartProductUnwrap       = "cart_product_unwrap"           // 取消打包

	// 分批类型
	ActionChangeBatchTag = "change_batch_tag" // 更换分批类型

	// 必点方案
	ActionMustPlanConfirm      = "must_plan_confirm"       // 确认必点方案（桌台）
	ActionOrderMustPlanConfirm = "order_must_plan_confirm" // 确认必点方案（点餐）

	// 订单检查
	ActionOrderCheck = "order_check" // 订单检查

	// 支付相关
	ActionPaymentCoupon   = "payment_coupon"    // 选择优惠券
	ActionPaymentActivity = "payment_activity"  // 选择满减活动
	ActionPaymentPoints   = "payment_points"    // 设置抵扣积分
	ActionPaymentCreate   = "payment_create"    // 创建支付单
	ActionPaymentCancel   = "payment_cancel"    // 取消支付单
	ActionPaymentFinish   = "payment_finish"    // 完成支付
	ActionPaymentZeroRule = "payment_zero_rule" // 设置抹零规则
	ActionOrderFree       = "order_free"        // 免单

	// 拆单操作
	ActionSaleOrderCreate      = "sale_order_create"       // 创建子订单
	ActionSaleOrderMoveProduct = "sale_order_move_product" // 移动商品
	ActionSaleOrderDelete      = "sale_order_delete"       // 删除子订单
	ActionSaleOrderDeleteAll   = "sale_order_delete_all"   // 删除所有子订单

	// 会员操作
	ActionOrderUseMember    = "order_use_member"    // 使用会员
	ActionOrderMemberCancel = "order_member_cancel" // 取消会员

	// 打印操作
	ActionOrderPrint        = "order_print"         // 打印小票
	ActionOrderPrintInvoice = "order_print_invoice" // 打印发票

	// 退货/反结账
	ActionReturnOrder   = "return_order"    // 退货
	ActionReReturnOrder = "re_return_order" // 重新退款
	ActionReverseSettle = "reverse_settle"  // 反结账

	// 订单商品报价
	ActionOrderProductQuotation = "order_product_quotation" // 订单商品报价

	// 其他
	ActionOrderUnlock = "order_unlock" // 订单解锁

	// H5 扫码点餐操作
	ActionH5OpenDesk         = "h5_open_desk"          // H5 开台
	ActionH5Call             = "h5_call"               // H5 呼叫服务
	ActionH5ProductRemark    = "h5_product_remark"     // H5 商品备注
	ActionH5OrderRemark      = "h5_order_remark"       // H5 整单备注
	ActionH5CartProductAdd   = "h5_cart_product_add"   // H5 添加商品
	ActionH5CartPackageAdd   = "h5_cart_package_add"   // H5 添加套餐
	ActionH5CartFlavorChange = "h5_cart_flavor_change" // H5 修改规格
	ActionH5CartProductNum   = "h5_cart_product_num"   // H5 修改数量
	ActionH5ConfirmOrder     = "h5_confirm_order"      // H5 确认下单
	ActionH5MustPlanConfirm  = "h5_must_plan_confirm"  // H5 确认必点方案

	// Tablet 平板操作
	ActionTabletOpenDesk        = "tablet_open_desk"         // 平板开台
	ActionTabletAddAndCooking   = "tablet_add_and_cooking"   // 平板添加并送厨
	ActionTabletProductRemark   = "tablet_product_remark"    // 平板商品备注
	ActionTabletOrderRemark     = "tablet_order_remark"      // 平板整单备注
	ActionTabletMustPlanConfirm = "tablet_must_plan_confirm" // 平板确认必点方案
	ActionTabletCall            = "tablet_call"              // 平板呼叫服务

	// Member 会员端操作
	ActionMemberOrderCreate           = "member_order_create"              // 会员创建订单
	ActionMemberOrderAddress          = "member_order_address"             // 会员设置地址
	ActionMemberOrderPay              = "member_order_pay"                 // 会员支付
	ActionMemberOrderCancel           = "member_order_cancel"              // 会员取消订单
	ActionMemberDineInSetDiningMethod = "member_dine_in_set_dining_method" // 会员设置堂食用餐方式
	ActionMemberDineInPay             = "member_dine_in_pay"               // 会员堂食订单支付
	ActionMemberDineInCancel          = "member_dine_in_cancel"            // 会员取消堂食订单
	ActionMemberOrderCallback         = "member_order_callback"            // 会员订单回调

	// Shop 商户后台操作
	ActionShopOrderCancel         = "shop_order_cancel"           // 商户取消订单
	ActionShopOrderDelete         = "shop_order_delete"           // 商户删除订单
	ActionShopOrderReturn         = "shop_order_return"           // 商户退款
	ActionShopOrderReReturn       = "shop_order_re_return"        // 商户重新退款
	ActionShopOrderRejectAll      = "shop_order_reject_all"       // 商户全部拒单
	ActionShopMemberOrderReject   = "shop_member_order_reject"    // 商户拒绝会员订单
	ActionShopMemberOrderCancel   = "shop_member_order_cancel"    // 商户取消会员订单
	ActionShopMemberOrderReturn   = "shop_member_order_return"    // 商户退款会员订单
	ActionShopMemberOrderReReturn = "shop_member_order_re_return" // 商户重新退款会员订单

	// Cashier 收银端 H5 接单操作
	ActionCashierRejectH5Order = "cashier_reject_h5_order" // 收银端拒绝H5订单
	ActionCashierAcceptH5Order = "cashier_accept_h5_order" // 收银端接受H5订单

	// Cashier 收银端会员订单操作
	ActionCashierAcceptMemberOrder   = "cashier_accept_member_order"    // 收银端接受会员订单
	ActionCashierRejectMemberOrder   = "cashier_reject_member_order"    // 收银端拒绝会员订单
	ActionCashierCookFinishOrder     = "cashier_cook_finish_order"      // 收银端备餐完成
	ActionCashierCancelMemberOrder   = "cashier_cancel_member_order"    // 收银端取消会员订单
	ActionCashierMemberOrderReturn   = "cashier_member_order_return"    // 收银端会员订单退款
	ActionCashierMemberOrderReReturn = "cashier_member_order_re_return" // 收银端会员订单重新退款

	// Cashier 收银端充值订单操作
	ActionCashierCancelRechargeOrder        = "cashier_cancel_recharge_order"         // 取消充值订单
	ActionCashierRechargeOrderRefund        = "cashier_recharge_order_refund"         // 充值订单退款
	ActionCashierRechargeOrderReReturn      = "cashier_recharge_order_re_return"      // 充值订单重新退款
	ActionCashierRechargeOrderReverseSettle = "cashier_recharge_order_reverse_settle" // 充值订单反结账

	// Cashier 收银端外卖订单操作
	ActionCashierTakeoutAcceptOrder = "cashier_takeout_accept_order" // 收银端接受外卖订单
	ActionCashierTakeoutRejectOrder = "cashier_takeout_reject_order" // 收银端拒绝外卖订单
	ActionCashierTakeoutCancelOrder = "cashier_takeout_cancel_order" // 收银端取消外卖订单
	ActionCashierTakeoutPrintOrder  = "cashier_takeout_print_order"  // 收银端打印外卖订单

	// Assistant 助手端操作
	ActionAssistantCheckAuthorization = "assistant_check_authorization" // 助手端检查授权

	// Kitchen 厨房端操作
	ActionKitchenConfirmReturn    = "kitchen_confirm_return"     // 厨房确认退菜
	ActionKitchenConfirmReturnAll = "kitchen_confirm_return_all" // 厨房批量确认退菜
)

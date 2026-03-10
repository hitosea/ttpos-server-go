package constant

const (
	OrderOperateSource = "operate_source" // 操作来源
)

const (
	OrderSourceInstant      = "instant"        // 点餐
	OrderSourceDesk         = "desk"           // 桌台
	OrderSourceRecharge     = "recharge"       // 充值
	OrderSourceMember       = "member"         // 会员端-外送
	OrderSourceMemberDineIn = "member_dine_in" // 会员端-堂食
)

const (
	OrderDateTypeAll       = -1 // 全都
	OrderDateTypeToday     = 0  // 今天
	OrderDateTypeYesterday = 1  // 昨天
	OrderDateTypeWeek      = 2  // 本周
	OrderDateTypeMonth     = 3  // 本月
	OrderDateTypeYear      = 4  // 本年
	OrderDateTypeLastWeek  = 5  // 近7天
	OrderDateTypeLastMonth = 6  // 上个月
)

const (
	SaleBillTypeDesk    = 0 // 桌台
	SaleBillTypeInstant = 1 // 点餐
	SaleBillTypeTakeout = 2 // 外送，会员端订单
)

const (
	SaleBillDiningMethodDineIn  = 0 // 堂食
	SaleBillDiningMethodTakeout = 1 // 打包
)

var SaleBillDiningMethodMap = map[uint]uint{
	0: SaleBillDiningMethodDineIn,
	1: SaleBillDiningMethodTakeout,
}

// 会员端订单类型（用于商品列表查询）
const (
	MemberOrderTypeDelivery   = 0 // 外送（默认）- 商品价格应用外送折扣率
	MemberOrderTypeSelfPickup = 1 // 堂食/到店自取 - 商品价格与收银机相同
)

const (
	SaleBillStatusPending  = 0 // 待付款
	SaleBillStatusComplete = 1 // 已完成
	SaleBillStatusCanceled = 2 // 已取消
)

const (
	DiscountOperationLogTypeChangePriceSaleOrder = 1 // 订单改价、整单改价
	DiscountOperationLogTypeDiscountSaleOrder    = 2 // 订单折扣、整单折扣
	DiscountOperationLogTypeZeroSaleOrder        = 3 // 订单抹零、整单抹零
)

const (
	// 销售订单状态
	SaleOrderStatusPending  = 0 // 未结账
	SaleOrderStatusFinish   = 1 // 已结账
	SaleOrderStatusCanceled = 2 // 已取消
)

const (
	OrderProductIsAcceptOrderAccepted = Yes // 已接单
	OrderProductIsAcceptOrderUnAccept = No  // 未接单
)

const (
	// 税费类型
	TaxFeeTypeNone  = 0 // 不收取税费
	TaxFeeTypeNoTax = 1 // 商品未含税
	TaxFeeTypeTax   = 2 // 商品已含税
)

// 拆单
const (
	SaleBillIsSplitOrderYes = 1 // 拆单
	SaleBillIsSplitOrderNo  = 0 // 不拆单
)

const (
	// 是否开启服务费
	SaleBillSettingIsOpenServiceFeeNo  = "0" // 关闭服务费
	SaleBillSettingIsOpenServiceFeeYes = "1" // 开启服务费
	// 服务费类型
	SaleBillSettingServiceFeeFixed   = "1" // 固定服务费
	SaleBillSettingServiceFeePercent = "2" // 按比例
	// 是否开启税费
	SaleBillSettingIsOpenTaxNo  = "0" // 关闭税费
	SaleBillSettingIsOpenTaxYes = "1" // 开启税费
)

const (
	SaleBillSettingServiceFeeTypeNone       = 0 // 不收取服务费
	SaleBillSettingServiceFeeTypeFixed      = 1 // 固定服务费
	SaleBillSettingServiceFeeTypePercent    = 2 // 按比例收取服务费（服务费不收税）
	SaleBillSettingServiceFeeTypePercentTax = 3 // 按比例收取服务费（服务费收税）

	SaleBillSettingTaxFeeTypeNone       = 0 // 不收取税费
	SaleBillSettingTaxFeeTypePercent    = 1 // 商品未含税
	SaleBillSettingTaxFeeTypePercentTax = 2 // 商品已含税

	DiscountZeroRuleNone    = 0 // 实款实收
	DiscountZeroRulePercent = 1 // 抹分
	DiscountZeroRuleFixed   = 2 // 抹角
	DiscountZeroRuleRound   = 3 // 四舍五入保留一位小数
	DiscountZeroRuleInteger = 4 // 四舍五入保留整数

	SaleBillSettingCheckoutZeroingMethodNone        = 0 // 实款实收
	SaleBillSettingCheckoutZeroingMethodPercent     = 1 // 抹分
	SaleBillSettingCheckoutZeroingMethodFixed       = 2 // 抹角
	SaleBillSettingCheckoutZeroingMethodYuan        = 5 // 抹元。 为了整体无分歧，抹元使用5. 让在系统中各个抹零规格无歧义，0-5每个数字都代表不同含义
	SaleBillSettingCheckoutZeroingMethodYuanAbandon = 3 // 抹元。废弃，将在下个版本中删除

	SaleBillSettingIsStatGiftNone = 0 // 不计入总销售额、优惠折扣
	SaleBillSettingIsStatGiftYes  = 1 // 计入总销售额、优惠折扣

	SaleBillSettingIsStatFreeNone = 0 // 不计入总销售额、优惠折扣、服务费、税费
	SaleBillSettingIsStatFreeYes  = 1 // 计入总销售额、优惠折扣、服务费、税费

	SaleBillSettingDiscountTypePercent = 0 // 百分比打折%
	SaleBillSettingDiscountTypeOff     = 1 // 百分比直接减免% off
)

// OrderSourceMapToOrderNoType 订单来源映射到订单编号类型
var OrderSourceMapToOrderNoType = map[string]string{
	OrderSourceInstant:      "1", // 点餐
	OrderSourceDesk:         "2", // 桌台
	OrderSourceRecharge:     "3", // 充值
	OrderSourceMember:       "4", // 会员端-外送
	OrderSourceMemberDineIn: "5", // 会员端-堂食
}

// OrderSourceMapToBillType 订单来源映射到销售账单类型
var OrderSourceMapToBillType = map[string]uint{
	OrderSourceInstant:      SaleBillTypeInstant, // 点餐
	OrderSourceDesk:         SaleBillTypeDesk,    // 桌台
	OrderSourceMember:       SaleBillTypeTakeout, // 会员端-外送
	OrderSourceMemberDineIn: SaleBillTypeInstant, // 会员端-堂食（使用点餐类型，价格与收银机一致）
}

// 订单操作类型
const (
	OrderOpenTable           = "OPEN_TABLE"            // 开台
	OrderSendKitchen         = "SEND_KITCHEN"          // 送厨
	OrderRefundProduct       = "REFUND_PRODUCT"        // 退菜
	OrderCancelRefundProduct = "CANCEL_REFUND_PRODUCT" // 取消退菜
	OrderChangeTable         = "CHANGE_TABLE"          // 转台
	OrderChangePrice         = "CHANGE_PRICE"          // 改价
	OrderUpdateMealNum       = "UPDATE_MEAL_NUM"       // 修改桌台就餐人数
	OrderStayOrder           = "STAY_ORDER"            // 挂单
	OrderPickOrder           = "PICK_ORDER"            // 取单
	OrderProductFree         = "PRODUCT_FREE"          // 赠菜
	OrderProductWrap         = "PRODUCT_WRAP"          // 打包
	OrderProductUnwrap       = "PRODUCT_UNWRAP"        // 取消打包
	OrderWrapSaleBill        = "WRAP_SALE_BILL"        // 整单打包
	OrderUnwrapSaleBill      = "UNWRAP_SALE_BILL"      // 取消整单打包
	OrderCancelProductFree   = "CANCEL_PRODUCT_FREE"   // 取消赠菜
	OrderProductMove         = "PRODUCT_MOVE"          // 转菜
	OrderDiscount            = "DISCOUNT"              // 优惠折扣
	OrderCancelDiscount      = "CANCEL_DISCOUNT"       // 撤销优惠折扣
	OrderActivity            = "ACTIVITY"              // 满减活动
	OrderSettle              = "SETTLE"                // 结账
	OrderFreeSale            = "FREE_SALE"             // 免单
	OrderReverseSettle       = "REVERSE_SETTLE"        // 反结账
	OrderRefund              = "REFUND"                // 退款
	OrderOrderTaking         = "ORDER_TAKING"          // 接单
	OrderOrderReject         = "ORDER_REJECT"          // 拒单
	OrderMergeTable          = "MERGE_TABLE"           // 并台
	OrderOrderCancel         = "ORDER_CANCEL"          // 整单取消
	OrderCheckoutDiscount    = "CHECKOUT_DISCOUNT"     // 结账手动抹零
	OrderSplitOrder          = "SPLIT_ORDER"           // 拆单
	OrderCancelSplitOrder    = "CANCEL_SPLIT_ORDER"    // 撤销拆单
	// -----订单操作类型-----
	OrderProductRemark    = "PRODUCT_REMARK"     // 产品备注
	OrderOrderRemark      = "ORDER_REMARK"       // 整单备注
	OrderAddProduct       = "ADD_PRODUCT"        // 增加菜品
	OrderDeleteProduct    = "DELETE_PRODUCT"     // 删除菜品
	OrderUpdateProductNum = "UPDATE_PRODUCT_NUM" // 修改桌台就餐人数
	OrderClock            = "CLOCK"              // 加钟
	OrderTakeout          = "TAKEOUT"            // 打包
	OrderUnlock           = "UNLOCK"             // 解锁
	OrderQuery            = "QUERY"              // 查询。查询自助餐商品列表
	OrderH5Confirm        = "H5_CONFIRM"         // 下单扫码h5订单
)

// 结账抹零操作类型
const (
	OrderCheckoutDiscountAdd    = "add"    // 设置结账抹零
	OrderCheckoutDiscountCancel = "cancel" // 撤销结账抹零
)

const (
	ProductDown            = "PRODUCT_DOWN"               // 商品已下架
	ProductDelete          = "PRODUCT_DELETE"             // 删除已删除
	ProductStockZero       = "PRODUCT_STOCK_ZERO"         // 库存不足
	ProductPriceChange     = "PRODUCT_PRICE_CHANGE"       // 商品价格变化
	ProductLimitOut        = "PRODUCT_LIMIT_OUT"          // 商品限购超出
	ProductMustPlanNotPass = "PRODUCT_MUST_PLAN_NOT_PASS" // 商品还未满足限购要求
	ProductPass            = "PRODUCT_PASS"               // 商品检查通过
)

// 状态, 0-未下单 1-未接单 2-已接单 3-已拒单

const (
	H5OrderStatusChooseProduct = 0 // 未下单
	H5OrderStatusOrder         = 1 // 未接单
	H5OrderStatusAccepted      = 2 // 已接单
	H5OrderStatusRejected      = 3 // 已拒单
)

// H5订单类型
const (
	H5OrderTypeDesk         = 0 // 桌台扫码订单
	H5OrderTypeMemberDineIn = 1 // 会员端堂食订单
)

// 会员端堂食订单状态过滤
const (
	MemberDineInOrderStatusAll        = "all"        // 全部
	MemberDineInOrderStatusUnpaid     = "unpaid"     // 待支付
	MemberDineInOrderStatusInProgress = "inprogress" // 进行中（已支付，待接单/备餐中）
	MemberDineInOrderStatusCompleted  = "completed"  // 已完成（包括部分退款、全部退款）
	MemberDineInOrderStatusCancelled  = "cancelled"  // 已取消（包括已取消和已拒单）
)

// GetMemberDineInOrderStatusFilter 根据状态过滤参数返回 SaleBill 状态列表和 H5 订单状态列表
// 返回值：
//   - billStatuses: SaleBill 状态过滤条件
//   - h5OrderStatuses: H5Order 状态过滤条件（用于进行中和已取消状态的细分）
//   - isPaid: 是否已支付过滤条件
func GetMemberDineInOrderStatusFilter(status string) (billStatuses []uint, h5OrderStatuses []uint, isPaid *bool) {
	switch status {
	case MemberDineInOrderStatusUnpaid:
		// 待支付：账单未完成且未支付
		billStatuses = []uint{SaleBillStatusPending}
		isPaidFalse := false
		isPaid = &isPaidFalse
	case MemberDineInOrderStatusInProgress:
		// 进行中：已支付，待接单或备餐中（H5订单状态为待接单或已接单）
		billStatuses = []uint{SaleBillStatusComplete} // 会员端堂食订单支付完成后标记订单为已经完成
		h5OrderStatuses = []uint{H5OrderStatusOrder, H5OrderStatusAccepted}
		isPaidTrue := true
		isPaid = &isPaidTrue
	case MemberDineInOrderStatusCompleted:
		// 已完成：账单已结账（包括部分退款、全部退款，因为退款后账单状态仍为已完成）
		billStatuses = []uint{SaleBillStatusComplete}
	case MemberDineInOrderStatusCancelled:
		// 已取消：账单已取消 或 H5订单已拒单
		// 注意：已拒单时账单状态可能是已取消，所以只需过滤 SaleBillStatusCanceled
		billStatuses = []uint{SaleBillStatusCanceled}
	default:
		// 全部：不过滤状态
		billStatuses = nil
	}
	return
}

const (
	SaleBillIsLockYes   = 1 // 账单锁定
	SaleBillIsLockNo    = 0 // 账单未锁定
	SaleBillIsBuffetYes = 1 // 是自助餐账单
	SaleBillIsBuffetNo  = 0 // 不是自助餐账单
)

const (
	DiscountTypePercent = 0 // 百分比折扣
	DiscountTypeOff     = 1 // 百分比减免Off
)

// 退货类型
const (
	ReturnOrderRefundTypeTotal = 1 // 整单退
	ReturnOrderRefundTypePart  = 2 // 部分退
)

// 退货单关联订单类型
const (
	ReturnOrderRelatedOrderTypeSaleOrder     = 0 // 销售订单
	ReturnOrderRelatedOrderTypeRechargeOrder = 1 // 充值订单
	ReturnOrderRelatedOrderTypeMemberOrder   = 2 // 会员订单
)

// 退货单商品类型
const (
	ReturnOrderProductTypeSaleOrderProduct        = 1 // 销售订单商品
	ReturnOrderProductTypeSaleOrderBuffetCustomer = 2 // 销售订单顾客
	ReturnOrderProductTypeBuffetAddTimeProduct    = 3 // 自助餐加钟商品
)

const (
	RankTypeSaleNum    = 1 // 按销售数量
	RankTypeSaleAmount = 2 // 按销售金额
)

const (
	PosInvoiceItemCodeServiceFee           = "VP001" // 服务费
	PosInvoiceItemCodeMembershipRecharge   = "VP002" // 会员充值
	PosInvoiceItemCodeDeliveryFee          = "VP003" // 配送费
	PosInvoiceItemCodePaymentProcessingFee = "VP004" // 支付手续费
	PosInvoiceItemCodeSpareGoods           = "BY001" // 备用商品（Grab/LINE MAN 未映射商品兜底）
)

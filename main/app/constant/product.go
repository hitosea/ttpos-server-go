package constant

const (
	ProductStatusOnSale  = 1 // 上架
	ProductStatusOffSale = 0 // 下架
	ProductStatusSaleOut = 1 // 售罄、沽清
)

const (
	ProductMustPlanStatusOn  = 1 // 开启
	ProductMustPlanStatusOff = 0 // 关闭

	ProductMustPlanMustRuleAll = 0 // 全选，固定商品
	ProductMustPlanMustRuleAny = 1 // 任选，可选商品

	ProductMustPlanMustTypeEachPerson = 1 // 每人必选，每人必选1份
	ProductMustPlanMustTypeEachOrder  = 0 // 每单必选,每单必选1份
)

const (
	SaleOrderProductStatusNormal  = 0 // 未送厨
	SaleOrderProductStatusCooking = 1 // 送厨
)

const (
	SaleOrderCustomAmountCancel = -1 // 取消整单改价金额
)

const (
	NoDiscount = 1 // 没有折扣. 默认100%
)

const (
	SaleOrderIsFreeNo  = 0 // 订单是否免单
	SaleOrderIsFreeYes = 1 // 订单是否免单
)

// 生产单商品状态, 0-待制作 1-制作中 2-已完成 3-已退菜
const (
	ProductionOrderProductStatusWait     = 0 // 待制作
	ProductionOrderProductStatusCooking  = 1 // 制作中
	ProductionOrderProductStatusFinished = 2 // 已完成
	ProductionOrderProductStatusCancel   = 3 // 已退菜
)

const (
	CustomPriceOn  = 1 // 是, 商品已改价
	CustomPriceOff = 0 // 否, 商品未改价
)

const (
	// 商品是否开启会员打折
	ProductMemberDiscountOn  = 1 // 是, 商品开启会员打折
	ProductMemberDiscountOff = 0 // 否, 商品不开启会员打折
)

const (
	// 商品赠菜
	ProductGiftOn  = 1 // 是, 商品已赠菜
	ProductGiftOff = 0 // 否, 商品未赠菜
	// 商品退菜
	ProductCancelOn  = 1 // 是, 商品已退菜
	ProductCancelOff = 0 // 否, 商品未退菜
)

const (
	ProductMustPlanAutoCartOn  = 1 // 是, 商品已自动加入购物车
	ProductMustPlanAutoCartOff = 0 // 否, 商品未自动加入购物车

	ProductMustPlanCustomerCanChangeOn  = 1 // 是, 顾客可修改必点数量
	ProductMustPlanCustomerCanChangeOff = 0 // 否, 顾客不可修改必点数量

)

// 小料是否必选
const (
	ProductPackageSauceRequiredOn  = 1 // 是, 小料已必选
	ProductPackageSauceRequiredOff = 0 // 否, 小料未必选
)

// 小料是否默认勾选
const (
	ProductPackageSauceDefaultSelectionOn  = 1 // 是, 默认勾选
	ProductPackageSauceDefaultSelectionOff = 0 // 否, 默认不勾选
)

// 商品属性组是否必选
const (
	ProductAttributeGroupRequiredOn  = 1 // 是, 属性组已必选
	ProductAttributeGroupRequiredOff = 0 // 否, 属性组未必选
)

// 商品属性是否默认勾选
const (
	ProductAttributeDefaultSelectionOn  = 1 // 是, 默认勾选
	ProductAttributeDefaultSelectionOff = 0 // 否, 默认不勾选
)

// 必点方案使用渠道
const (
	ProductMustPlanUseChannelDining = "10" // 点餐方式
	ProductMustPlanUseChannelDesk   = "20" // 桌台方式
)

// 是否显示必点方案
const (
	SaleBillShowMustPlanYes = 1 // 是, 显示必点方案
	SaleBillShowMustPlanNo  = 0 // 否, 不显示必点方案
)

// 是否自动加购必点商品
const (
	AutoAddMustProductYes = 1 // 是, 自动加购必点商品
	AutoAddMustProductNo  = 0 // 否, 不自动加购必点商品
)

// 是否是必点的销售订单商品
const (
	IsMustProductYes = 1 // 是, 必点的销售订单商品。用于标记is_required=1的商品
	IsMustProductNo  = 0 // 否, 不是必点的销售订单商品。用于标记is_required=0的商品
)

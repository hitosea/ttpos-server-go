package constant

const (
	ProductStatusOnSale  = 1 // 上架
	ProductStatusOffSale = 0 // 下架
	ProductStatusSaleOut = 1 // 售罄、沽清
)

const (
	ProductMustPlanStatusOn  = 1 // 开启
	ProductMustPlanStatusOff = 0 // 关闭

	ProductMustPlanMustRuleAll = 1 // 全选
	ProductMustPlanMustRuleAny = 0 // 任选

	ProductMustPlanMustTypeEachPerson = 1 // 每人必选
	ProductMustPlanMustTypeEachOrder  = 0 // 每单必选
)

const (
	SaleOrderProductStatusNormal  = 0 // 未送厨
	SaleOrderProductStatusCooking = 1 // 送厨
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

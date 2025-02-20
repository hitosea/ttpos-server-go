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
	ProductStatusNormal  = 0 // 未送厨
	ProductStatusCooking = 1 // 送厨
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

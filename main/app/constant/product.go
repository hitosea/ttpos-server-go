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

package constant

// 商品数量不能超过999个
const (
	ProductNumMax = 999
)

const (
	ProductPackageDeductStockTypePay     = 0 // 结账减库存，付款减库存
	ProductPackageDeductStockTypeCooking = 1 // 下单减库存，送厨减库存
)

// 商品类型
const (
	ProductTypeProduct           = 0 // 商品
	ProductTypePackage           = 1 // 套餐
	ProductTypePackageSubProduct = 2 // 套餐子商品
)

// 库存相关常量
const (
	ProductBomInfiniteStock = 999999 // 无限库存（关闭库存检查时使用）
)

const (
	ProductBomTypeSauce  = 0 // 小料
	ProductBomTypeFlavor = 1 // 规格商品
)

const (
	ProductStatusOnSale  = 1 // 上架
	ProductStatusOffSale = 0 // 下架
	ProductStatusSaleOut = 1 // 售罄、沽清
)

const (
	ProductMustPlanStatusOn  = Yes // 开启
	ProductMustPlanStatusOff = No  // 关闭

	ProductMustPlanMustRuleAll = 0 // 全选，固定商品
	ProductMustPlanMustRuleAny = 1 // 任选，可选商品

	ProductMustPlanMustTypeEachPerson = 1 // 每人必选，每人必选1份
	ProductMustPlanMustTypeEachOrder  = 0 // 每单必选,每单必选1份
)

const (
	SaleOrderProductStatusNormal  = No  // 未送厨
	SaleOrderProductStatusCooking = Yes // 送厨
)

const (
	SaleOrderCustomAmountCancel = -1 // 取消整单改价金额
)

const (
	NoDiscount = 1 // 没有折扣. 默认100%
)

const (
	SaleOrderIsFreeNo  = No  // 订单是否免单
	SaleOrderIsFreeYes = Yes // 订单是否免单
)

// 生产单商品状态, 0-待制作 1-制作中 2-已完成 3-已退菜
const (
	ProductionOrderProductStatusWait     = 0 // 待制作
	ProductionOrderProductStatusCooking  = 1 // 制作中
	ProductionOrderProductStatusFinished = 2 // 已完成
	ProductionOrderProductStatusCancel   = 3 // 已退菜
)

const (
	CustomPriceOn  = Yes // 是, 商品已改价
	CustomPriceOff = No  // 否, 商品未改价
)

const (
	// 商品是否开启会员打折
	ProductMemberDiscountOn  = Yes // 是, 商品开启会员打折
	ProductMemberDiscountOff = No  // 否, 商品不开启会员打折
)

const (
	// 商品赠菜
	ProductGiftOn  = Yes // 是, 商品已赠菜
	ProductGiftOff = No  // 否, 商品未赠菜
	// 商品退菜
	ProductCancelOn  = Yes // 是, 商品已退菜
	ProductCancelOff = No  // 否, 商品未退菜
)

const (
	ProductMustPlanAutoCartOn  = Yes // 是, 商品已自动加入购物车
	ProductMustPlanAutoCartOff = No  // 否, 商品未自动加入购物车

	ProductMustPlanCustomerCanChangeOn  = Yes // 是, 顾客可修改必点数量
	ProductMustPlanCustomerCanChangeOff = No  // 否, 顾客不可修改必点数量

)

// 成本卡日志操作类型
const (
	ProductBomCardLogOperationTypeCreate = 1 // 创建
	ProductBomCardLogOperationTypeDelete = 2 // 删除
)

// 小料是否必选
const (
	ProductPackageSauceRequiredOn  = Yes // 是, 小料已必选
	ProductPackageSauceRequiredOff = No  // 否, 小料未必选
)

// 小料是否默认勾选
const (
	ProductPackageSauceDefaultSelectionOn  = Yes // 是, 默认勾选
	ProductPackageSauceDefaultSelectionOff = No  // 否, 默认不勾选
)

// 商品属性组是否必选
const (
	ProductAttributeGroupRequiredOn  = Yes // 是, 属性组已必选
	ProductAttributeGroupRequiredOff = No  // 否, 属性组未必选
)

// 商品属性是否默认勾选
const (
	ProductAttributeDefaultSelectionOn  = Yes // 是, 默认勾选
	ProductAttributeDefaultSelectionOff = No  // 否, 默认不勾选
)

// 必点方案使用渠道
const (
	ProductMustPlanUseChannelDining = "10" // 点餐方式
	ProductMustPlanUseChannelDesk   = "20" // 桌台方式
)

// 是否显示必点方案
const (
	SaleBillShowMustPlanYes = Yes // 是, 显示必点方案
	SaleBillShowMustPlanNo  = No  // 否, 不显示必点方案
)

// 是否自动加购必点商品
const (
	AutoAddMustProductYes = Yes // 是, 自动加购必点商品
	AutoAddMustProductNo  = No  // 否, 不自动加购必点商品
)

// 是否是必点的销售订单商品
const (
	IsMustProductYes = Yes // 是, 必点的销售订单商品。用于标记is_required=1的商品
	IsMustProductNo  = No  // 否, 不是必点的销售订单商品。用于标记is_required=0的商品
)

// 是否是自助餐商品
const (
	SaleOrderProductIsBuffetYes = Yes // 是, 自助餐商品
	SaleOrderProductIsBuffetNo  = No  // 否, 不是自助餐商品
)

// 是否为计量商品
const (
	SaleOrderProductIsMeteringYes = Yes // 是, 计价商品
	SaleOrderProductIsMeteringNo  = No  // 否, 不是计价商品
)

// 商品打印状态
const (
	ProductPrinterStatusOpen  = 1 // 开启
	ProductPrinterStatusClose = 2 // 关闭
)

const (
	ProductionOrderProductColumnSaleBill = "sale_bill_uuid"
	ProductionOrderProductColumnCategory = "first_category_uuid"
)

// 服务费计算基准
const (
	SaleBillSettingServiceFeeBasePrice  = 0 // 商品惠后价
	SaleBillSettingServiceFeeBaseAmount = 1 // 商品价格合计
)

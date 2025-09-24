package constant

// WarehouseOutFormScene 出库单场景
const (
	WarehouseOutFormSceneSales  = iota // 销售出库 0
	WarehouseOutFormSceneAdjust        // 调整出库 1
	WarehouseOutFormSceneLoss          // 损耗出库 2
	WarehouseOutFormSceneLost          // 丢失出库 3
	WarehouseOutFormSceneDelete        // 删除出库 4
)

// WarehouseOutFormStatus 出库单状态
const (
	WarehouseOutFormItemStatusPre     = iota // 预出库 0
	WarehouseOutFormItemStatusSuccess        // 已出库 1
)

// WarehouseOutFormItemReduceStock 出库单明细是否减库存
const (
	WarehouseOutFormItemReduceStockNotProcessed = iota // 未减库存 0
	WarehouseOutFormItemReduceStockSuccess             // 已减库存 1
)

// MemberPointLogProcessed 会员积分日志是否已处理
const (
	MemberPointLogOrBalanceProcessedNot     = iota // 未处理 0 未处理积分变动\会员余额变动
	MemberPointLogOrBalanceProcessedSuccess        // 已处理 1 已处理积分变动\会员余额变动
)

// WarehouseFormScene 入库单场景
const (
	WarehouseFormScenePurchase = iota // 采购入库 0
	WarehouseFormSceneAddStock        // 添加入库 1
	WarehouseFormSceneAdjust          // 调整入库 2
	WarehouseFormSceneReturn          // 退菜入库 3
)

// WarehouseOutFormStatus 出库单状态
const (
	WarehouseOutFormStatusSuccess  = iota // 已出库 0
	WarehouseOutFormStatusCanceled        // 撤销 1
)

// WarehouseMonthlyFormScene 月度报表场景
const (
	WarehouseMonthlyFormSceneStart = iota // 月初 0
	WarehouseMonthlyFormSceneEnd          // 月末 1
)

const (
	WarehouseTypeNormal  = "normal"
	WarehouseTypeTransit = "transit"
)

// WarehouseInOutLogType 出入库日志类型
const (
	WarehouseInOutLogTypePurchase = "purchase" // 采购入库 0
	WarehouseInOutLogTypeSale     = "sale"     // 销售出库 1
	WarehouseInOutLogTypeDelivery = "delivery" // 发货出库 2
)

const (
	WarehouseInOutLogTypePurchaseInt = 0 // 采购入库 0
	WarehouseInOutLogTypeSaleInt     = 1 // 销售出库 1
	WarehouseInOutLogTypeDeliveryInt = 2 // 发货出库 2
)

func WarehouseInOutLogTypeToInt(typ string) int {
	switch typ {
	case WarehouseInOutLogTypePurchase:
		return 0
	case WarehouseInOutLogTypeSale:
		return 1
	case WarehouseInOutLogTypeDelivery:
		return 2
	}
	return -1
}

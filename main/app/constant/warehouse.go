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

// WarehouseOutFormOrderType 出库单订单类型
const (
	WarehouseOutFormOrderTypeDineIn  = iota // 堂食订单 0
	WarehouseOutFormOrderTypeTakeout        // 外卖订单 1
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
	WarehouseInOutLogTypePurchase    = "purchase"     // 采购入库 0
	WarehouseInOutLogTypeSale        = "sale"         // 销售出库 1
	WarehouseInOutLogTypeDelivery    = "delivery"     // 发货出库 2
	WarehouseInOutLogTypeProfitIn    = "profit_in"    // 盘盈入库 3
	WarehouseInOutLogTypeLossOut     = "loss_out"     // 盘亏出库 4
	WarehouseInOutLogTypeTransferIn  = "transfer_in"  // 调拨入库 5
	WarehouseInOutLogTypeTransferOut = "transfer_out" // 调拨出库 6
)

const (
	WarehouseInOutLogTypePurchaseInt = 0 // 采购入库 0
	WarehouseInOutLogTypeSaleInt     = 1 // 销售出库 1
	WarehouseInOutLogTypeDeliveryInt = 2 // 发货出库 2
)

// WarehouseInOutLogLogType 出入库日志类型
const (
	WarehouseInOutLogLogTypeIn  = 0 // 入库
	WarehouseInOutLogLogTypeOut = 1 // 出库
)

// WarehouseInOutLogScene 出入库日志场景
const (
	WarehouseInOutLogScenePurchase    = 0 // 采购入库
	WarehouseInOutLogSceneSale        = 1 // 销售出库
	WarehouseInOutLogSceneDelivery    = 2 // 发货出库
	WarehouseInOutLogSceneProfitIn    = 3 // 盘盈入库
	WarehouseInOutLogSceneLossOut     = 4 // 盘亏出库
	WarehouseInOutLogSceneTransferIn  = 5 // 调拨入库
	WarehouseInOutLogSceneTransferOut = 6 // 调拨出库

	WarehouseInOutLogSceneTransitIn  = 20 // 在途入库
	WarehouseInOutLogSceneTransitOut = 21 // 在途出库
)

func WarehouseInOutLogTypeToInt(typ string) int {
	switch typ {
	case WarehouseInOutLogTypePurchase:
		return 0
	case WarehouseInOutLogTypeSale:
		return 1
	case WarehouseInOutLogTypeDelivery:
		return 2
	case WarehouseInOutLogTypeProfitIn:
		return 3
	case WarehouseInOutLogTypeLossOut:
		return 4
	case WarehouseInOutLogTypeTransferIn:
		return 5
	case WarehouseInOutLogTypeTransferOut:
		return 6
	}
	return -1
}

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
	WarehouseOutFormStatusPre     = iota // 预出库 0
	WarehouseOutFormStatusSuccess        // 已出库 1
)

// WarehouseOutFormItemReduceStock 出库单明细是否减库存
const (
	WarehouseOutFormItemReduceStockNotProcessed = iota // 未减库存 0
	WarehouseOutFormItemReduceStockSuccess             // 已减库存 1
)

// WarehouseFormScene 入库单场景
const (
	WarehouseFormScenePurchase = iota // 采购入库 0
	WarehouseFormSceneAddStock        // 添加入库 1
	WarehouseFormSceneAdjust          // 调整入库 2
	WarehouseFormSceneReturn          // 退菜入库 3
)

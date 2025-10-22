package constant

const (
	// StockReconciliationStatusSaved 盘点单状态-已保存
	StockReconciliationStatusSaved = 0
	// StockReconciliationStatusSubmitted 盘点单状态-已提交
	StockReconciliationStatusSubmitted = 1
	// StockReconciliationStatusApproved 盘点单状态-已审核
	StockReconciliationStatusApproved = 2
	// StockReconciliationStatusRejected 盘点单状态-已驳回
	StockReconciliationStatusRejected = 3
)

const (
	// StockReconciliationTypeSpecified 盘点类型-指定物品盘点
	StockReconciliationTypeSpecified = 1
	// StockReconciliationTypeAll 盘点类型-全部物品盘点
	StockReconciliationTypeAll = 2
)

const (
	// StockReconciliationPurposeInventory 盘点目的-库存盘点
	StockReconciliationPurposeInventory = 1
	// StockReconciliationPurposeInitial 盘点目的-期初盘点
	StockReconciliationPurposeInitial = 2
)

const (
	// StockReconciliationInventoryStatusProfit 库存状态-盘盈
	StockReconciliationInventoryStatusProfit = 1
	// StockReconciliationInventoryStatusLoss 库存状态-盘亏
	StockReconciliationInventoryStatusLoss = 2
	// StockReconciliationInventoryStatusNormal 库存状态-正常
	StockReconciliationInventoryStatusNormal = 3
)

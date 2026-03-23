package constant

type ItemGroup string

const (
	// ItemGroupRawMaterial 原材料
	ItemGroupRawMaterial ItemGroup = "Raw Material"
	// ItemGroupProducts 商品
	ItemGroupProducts ItemGroup = "Products"
	// ItemGroupOthers 其他
	ItemGroupOthers ItemGroup = ""
)

const (
	ErpHeadquartersSupplierCode = "Headquarters - Supplier"
	ErpRequestPageSize          = 999
	HeadquartersSupplierCode    = "HSP00001"
)

const (
	NormalWarehouseCodeContains  = "Normal-Default"
	TransitWarehouseCodeContains = "Transit-Transit"
)

const (
	NormalWarehouseCode  = "WH01"
	TransitWarehouseCode = "WH02"
)

const (
	ErpWarehouseTypeNormal1 = ""
	ErpWarehouseTypeNormal2 = "Normal"
	ErpWarehouseTypeTransit = "Transit"
)

// ErpSyncStatus ERP同步状态
const (
	ErpSyncStatusNotSynced  = 0
	ErpSyncStatusQueued     = 1
	ErpSyncStatusInProgress = 2
	ErpSyncStatusSuccess    = 3
	ErpSyncStatusFailed         = 4
	ErpSyncStatusExternalCancel = 5 // ERPNext 端外部取消（Webhook on_cancel）
)

// ErpPosInvoiceErrorScene 保存POS发票错误场景
const (
	// 物品库存不足
	ErpItemStockNotEnough = "ItemStockNotEnough"
)

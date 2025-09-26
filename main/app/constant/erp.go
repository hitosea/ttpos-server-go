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

// ErpPosInvoiceErrorScene 保存POS发票错误场景
const (
	// 物品库存不足
	ErpItemStockNotEnough = "ItemStockNotEnough"
)

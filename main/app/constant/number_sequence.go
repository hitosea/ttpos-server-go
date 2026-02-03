package constant

// 编号类型常量
const (
	NumberTypeInvoice         = "invoice"          // 发票编号
	NumberTypePurchaseReq     = "purchase_req"     // 采购申请（外部）
	NumberTypePurchaseReceipt = "purchase_receipt" // 采购收货（外部）
	NumberTypeBrandPurchase   = "brand_purchase"   // 品牌采购（内部）
	NumberTypeBrandReceipt    = "brand_receipt"    // 品采收货（内部）
	NumberTypeStockTake       = "stock_take"       // 盘点单
	NumberTypeStockLoss       = "stock_loss"       // 报损单
	NumberTypeTransfer        = "transfer"         // 调拨单
)

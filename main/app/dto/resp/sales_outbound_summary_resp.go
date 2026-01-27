package resp

// RegenerateSalesOutboundSummaryResp 重新生成销售出库汇总记录响应
type RegenerateSalesOutboundSummaryResp struct {
	DeletedCount   int   `json:"deleted_count"`   // 删除的记录数
	GeneratedCount int   `json:"generated_count"` // 生成的记录数
	DurationMs     int64 `json:"duration_ms"`     // 操作耗时（毫秒）
}

// RegenerateOrderMaterialResp 重新生成订单材料用料记录响应
type RegenerateOrderMaterialResp struct {
	DeletedCount  int   `json:"deleted_count"`  // 删除的记录数
	InsertedCount int   `json:"inserted_count"` // 插入的记录数
	DurationMs    int64 `json:"duration_ms"`    // 操作耗时（毫秒）
}

// RegenerateSaleBillMaterialOutboundResp 重新生成销售账单材料出库记录响应
type RegenerateSaleBillMaterialOutboundResp struct {
	DeletedCount  int   `json:"deleted_count"`  // 删除的记录数
	InsertedCount int   `json:"inserted_count"` // 插入的记录数
	DurationMs    int64 `json:"duration_ms"`    // 操作耗时（毫秒）
}

// RegenerateOrderPosInvoiceResp 重新生成订单POS发票响应
type RegenerateOrderPosInvoiceResp struct {
	ProductsInvoiceName string `json:"products_invoice_name"` // 商品发票名称
	MaterialInvoiceName string `json:"material_invoice_name"` // 材料发票名称
	DurationMs          int64  `json:"duration_ms"`           // 操作耗时（毫秒）
}

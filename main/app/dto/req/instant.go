package req

// GetInstantOrderInfoReq 获取点餐订单详情请求
type GetInstantOrderInfoReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

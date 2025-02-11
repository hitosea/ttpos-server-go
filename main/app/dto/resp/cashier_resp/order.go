package cashier_resp

// 创建点餐订单响应
type CreateInstantOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

// 创建桌台订单响应
type CreateDeskOrderResp struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID
}

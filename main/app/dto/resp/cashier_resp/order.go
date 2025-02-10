package cashier_resp

// 创建订单响应
type CreateOrderResp struct {
	Uuid uint64 `json:"uuid"` // 订单UUID
}

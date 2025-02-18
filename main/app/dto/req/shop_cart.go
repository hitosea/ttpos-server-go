package req

// OrderCartInfoReq 查询点餐购物车信息请求参数
type OrderCartInfoReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售订单ID
}

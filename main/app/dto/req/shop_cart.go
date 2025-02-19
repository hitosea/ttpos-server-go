package req

// OrderCartProductAddReq 向购物车添加商品请求参数
type OrderCartProductAddReq struct {
	SaleBillUuid      uint64   `json:"sale_bill_uuid"`  // 销售账单ID
	SaleOrderUuid     uint64   `json:"sale_order_uuid"` // 销售订单ID
	FlavorUuid        uint64   `json:"flavor_uuid"`     // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`      // 小料ID
	AttributeUuidList []uint64 `json:"attribute_uuid"`  // 规格ID
}

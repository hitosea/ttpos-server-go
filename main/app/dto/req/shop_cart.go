package req

// OrderCartProductAddReq 向购物车添加商品请求参数
type OrderCartProductAddReq struct {
	SaleBillUuid      uint64   `json:"sale_bill_uuid"`  // 销售账单ID。可选，参数不填时表示要新建销售账单，添加商品后创建点餐销售账单。
	SaleOrderUuid     uint64   `json:"sale_order_uuid"` // 销售订单ID。可选，参数不填时默认加购到第一个销售订单中
	FlavorUuid        uint64   `json:"flavor_uuid"`     // 某个规格商品ID
	SauceUuidList     []uint64 `json:"sauce_uuid"`      // 小料ID
	AttributeUuidList []uint64 `json:"attribute_uuid"`  // 规格ID
}

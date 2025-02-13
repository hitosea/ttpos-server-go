package req

// InstantOrderGetInfoReq 获取点餐订单详情请求
type InstantOrderGetInfoReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

// AddProduct 添加商品请求
type AddProduct struct {
	Uuid       uint64                `json:"uuid"`        // 商品UUID, 必填
	FlavorUuid uint64                `json:"flavor_uuid"` // 规格UUID, 必填
	SauceUuids []uint64              `json:"sauce_uuids"` // 小料UUID, 非必填
	Attributes []AddProductAttribute `json:"attributes"`  // 商品属性, 非必填
}

// AddProductAttribute 添加商品属性
type AddProductAttribute struct {
	GroupUuid  uint64   `json:"group_uuid"`  // 属性组UUID
	ValueUuids []uint64 `json:"value_uuids"` // 属性值UUID
}

// InstantOrderAddProductReq 添加商品请求
type InstantOrderAddProductReq struct {
	SaleBillUuid  uint64     `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64     `json:"sale_order_uuid"` // 销售订单UUID, 必填
	Product       AddProduct `json:"product"`         // 商品, 必填
}

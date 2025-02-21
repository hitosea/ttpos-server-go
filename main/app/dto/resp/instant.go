package resp

// 创建点餐订单响应
type CreateInstantOrderResp struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

// InstantOrder 点餐订单
type InstantOrder struct {
	SaleOrderUuid  uint64  `json:"sale_order_uuid"` // 销售订单UUID
	ProductCount   uint    `json:"product_count"`   // 产品数量
	ProductAmount  float64 `json:"product_amount"`  // 产品总金额
	ServiceFee     float64 `json:"service_fee"`     // 服务费
	TaxFee         float64 `json:"tax_fee"`         // 消费税
	DiscountAmount float64 `json:"discount_amount"` // 优惠折扣
	MemberDiscount float64 `json:"member_discount"` // 会员折扣
	TotalAmount    float64 `json:"total_amount"`    // 总金额
}

// InstantOrderList 点餐订单列表
type InstantOrderList struct {
	List []InstantOrder `json:"list"` // 点餐订单列表
}

// GetInstantOrderInfoResp 获取点餐订单详情响应
type GetInstantOrderInfoResp struct {
	SaleBillUuid  uint64             `json:"sale_bill_uuid"`  // 销售账单UUID
	IsKitchen     bool               `json:"is_kitchen"`      // 是否送厨: false-否 true-是
	SaleOrderList []InstantOrderList `json:"sale_order_list"` // 销售订单列表
}

// OrderMustPlanResp 点餐必点方案响应
type OrderMustPlanResp struct {
	MustPlanList []MustPlan `json:"must_plan_list"` // 必点方案列表
}

// MustPlan 必点方案
type MustPlan struct {
	ProductUuid uint64 `json:"product_uuid"` // 商品UUID
	ProductName string `json:"product_name"` // 商品名称
}

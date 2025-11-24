package setting

// GetDataManageInfoResp 获取数据管理信息响应
type GetDataManageResp struct {
	IsEnableDataManage bool                 `json:"is_enable_data_manage"` // 状态: false-关闭 true-开启
	StaffCount         int64                `json:"staff_count"`           // 操作人员数量
	OrderCount         int64                `json:"order_count"`           // 订单数量
	Statistics         DataManageStatistics `json:"statistics"`            // 统计信息
}

// DataManageStatistics 数据管理统计信息响应
type DataManageStatistics struct {
	SaleAmount     float64 `json:"sale_amount"`     // 总销售额
	ReceivedPrice  float64 `json:"received_price"`  // 实收金额
	ProductCount   float64 `json:"product_count"`   // 商品数量
	DiscountMember float64 `json:"discount_member"` // 会员折扣
	BusinessAmount float64 `json:"business_amount"` // 营业收入
	ServiceFee     float64 `json:"service_fee"`     // 服务费
	PaymentFee     float64 `json:"payment_fee"`     // 支付手续费
	Tax            float64 `json:"tax"`             // 税费
	RefundAmount   float64 `json:"refund_amount"`   // 退款金额
	Discount       float64 `json:"discount"`        // 优惠折扣
	DiscountRatio  float64 `json:"discount_ratio"`  // 优惠占比
	GiveAmount     float64 `json:"give_amount"`     // 赠菜金额
	GiveCount      float64 `json:"give_count"`      // 赠菜数量
	FreeAmount     float64 `json:"free_amount"`     // 免单总额
	FreeCount      float64 `json:"free_count"`      // 免单数量
}

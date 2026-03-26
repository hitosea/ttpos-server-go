package setting

import "ttpos-server-go/app/dto"

// GetDataManageInfoResp 获取数据管理信息响应
type GetDataManageResp struct {
	IsEnableDataManage bool                 `json:"is_enable_data_manage"` // 状态: false-关闭 true-开启
	StaffCount         int                  `json:"staff_count"`           // 操作人员数量
	OrderCount         int                  `json:"order_count"`           // 订单数量
	StaffUuids         []uint64             `json:"staff_uuids"`           // 操作人员UUID列表
	OrderUuids         []uint64             `json:"order_uuids"`           // 订单UUID列表
	Statistics         DataManageStatistics `json:"statistics"`            // 统计信息
}

// DataManageOrderListResp 已选订单列表响应
type DataManageOrderListResp struct {
	List []DataManageOrderItem   `json:"list"` // 订单列表
	Meta DataManageOrderListMeta `json:"meta"` // Meta信息
}

// DataManageOrderListMeta 已选订单列表Meta信息
type DataManageOrderListMeta struct {
	dto.PageResponse
	OrderCount int64   `json:"order_count"` // 已选订单数量（全量，不受搜索影响）
	PaidAmount float64 `json:"paid_amount"` // 实付金额合计（全量，不受搜索影响）
}

// DataManageOrderItem 数据管理订单项
type DataManageOrderItem struct {
	SaleBillUuid  uint64  `json:"sale_bill_uuid"` // 销售单UUID
	OrderNo       string  `json:"order_no"`       // 订单编号
	CreateTime    int64   `json:"create_time"`    // 创建时间
	Amount        float64 `json:"amount"`         // 订单金额
	PaymentAmount float64 `json:"payment_amount"` // 实付金额
	PaymentMethod string  `json:"payment_method"` // 支付方式
	IsSelected    bool    `json:"is_selected"`    // 是否已选中（用于可选订单列表）
}

// DataManageOrderSelectResp 可选订单列表响应
type DataManageOrderSelectResp struct {
	List []DataManageOrderItem     `json:"list"` // 订单列表
	Meta DataManageOrderSelectMeta `json:"meta"` // Meta信息
}

// DataManageOrderSelectMeta 可选订单列表Meta信息
type DataManageOrderSelectMeta struct {
	dto.PageResponse
}

// DataManageOrderSelectStatsResp 可选订单统计预览响应（不持久化，仅预览提交后的统计）
type DataManageOrderSelectStatsResp struct {
	SelectedCount int64    `json:"selected_count"` // 预览选中数量
	PaidAmount    float64  `json:"paid_amount"`    // 预览选中实付金额合计
	TotalCount    int64    `json:"total_count"`    // 筛选范围内订单总数
	IsSelectAll   bool     `json:"is_select_all"`  // 是否全选（所有筛选范围内的订单都已选中）
	SelectedUuids []uint64 `json:"selected_uuids"` // 最终选中的订单UUID列表
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

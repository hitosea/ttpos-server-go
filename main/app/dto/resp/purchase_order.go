package resp

import "ttpos-server-go/app/dto"

// PurchaseOrderListResp 采购订单列表响应
type PurchaseOrderListResp struct {
	List []*PurchaseOrderInfo `json:"list"` // 采购订单列表
	Meta dto.PageResponse     `json:"meta"` // 分页信息
}

// PurchaseOrderInfo 采购订单信息
type PurchaseOrderInfo struct {
	Uuid              uint64 `json:"uuid"`                // 采购订单ID
	OrderNo           string `json:"order_no"`            // 申请单编号（单据编号）
	ErpOrderNo        string `json:"erp_order_no"`        // ERP申请单编号（采购单号）
	Status            int    `json:"status"`              // 状态 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成)
	OrderTime         int64  `json:"order_time"`          // 单据日期
	Num               int    `json:"num"`                 // 物品数量
	OrderType         int    `json:"order_type"`          // 申请类型
	SupplierName      string `json:"supplier_name"`       // 供应商名称
	ExpectArrivalTime int64  `json:"expect_arrival_time"` // 期望到货日期
	ReceiptProgress   string `json:"receipt_progress"`    // 收货进度（百分比0.00%）前端直接显示
}

// PurchaseOrderDetailResp 采购订单详情响应
type PurchaseOrderDetailResp struct {
	PurchaseOrderInfo
	Items []PurchaseOrderItemInfo `json:"items"` // 采购明细
}

// PurchaseOrderItemInfo 采购订单商品明细信息
type PurchaseOrderItemInfo struct {
	Uuid               uint64  `json:"uuid"`                 // 明细ID
	MaterialCode       string  `json:"material_code"`        // 物品编码
	MaterialName       string  `json:"material_name"`        // 物品名称
	Num                float64 `json:"num"`                  // 申请数量
	ArrivalNum         float64 `json:"arrival_num"`          // 到货数量
	UnitName           string  `json:"unit_name"`            // 采购单位名称
	UnitConversionRate float64 `json:"unit_conversion_rate"` // 基准单位转换率
	BaseUnitName       string  `json:"base_unit_name"`       // 基准单位名称
}

// PurchaseOrderLogInfo 采购订单操作日志信息
type PurchaseOrderLogInfo struct {
	Uuid              uint64 `json:"uuid"`                // 日志ID
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid"` // 采购订单ID
	OperatorUuid      uint64 `json:"operator_uuid"`       // 操作人ID
	OperatorName      string `json:"operator_name"`       // 操作人姓名
	Action            string `json:"action"`              // 操作动作
	ActionDesc        string `json:"action_desc"`         // 操作描述
	OldStatus         int    `json:"old_status"`          // 操作前状态
	NewStatus         int    `json:"new_status"`          // 操作后状态
	Content           string `json:"content"`             // 操作内容详情
	Remark            string `json:"remark"`              // 备注
	CreateTime        int    `json:"create_time"`         // 创建时间
}

// PurchaseOrderCreateResp 创建采购订单响应
type PurchaseOrderCreateResp struct {
	Uuid    uint64 `json:"uuid"`     // 采购订单ID
	OrderNo string `json:"order_no"` // 采购订单编号
}

// PurchaseReceiptCreateResp 创建收货记录响应
type PurchaseReceiptCreateResp struct {
	Uuid      uint64 `json:"uuid"`       // 收货记录ID
	ReceiptNo string `json:"receipt_no"` // 收货单号
}

// PurchaseOrderStatisticsResp 采购订单统计响应
type PurchaseOrderStatisticsResp struct {
	StatusStats   []StatusStatItem   `json:"status_stats"`   // 状态统计
	PriorityStats []PriorityStatItem `json:"priority_stats"` // 优先级统计
	SupplierStats []SupplierStatItem `json:"supplier_stats"` // 供应商统计
	TimeStats     []TimeStatItem     `json:"time_stats"`     // 时间统计
	SummaryStats  SummaryStatItem    `json:"summary_stats"`  // 汇总统计
}

// StatusStatItem 状态统计项
type StatusStatItem struct {
	Status     int     `json:"status"`      // 状态
	StatusText string  `json:"status_text"` // 状态文本
	Count      int     `json:"count"`       // 数量
	Amount     float64 `json:"amount"`      // 金额
}

// PriorityStatItem 优先级统计项
type PriorityStatItem struct {
	Priority     int     `json:"priority"`      // 优先级
	PriorityText string  `json:"priority_text"` // 优先级文本
	Count        int     `json:"count"`         // 数量
	Amount       float64 `json:"amount"`        // 金额
}

// SupplierStatItem 供应商统计项
type SupplierStatItem struct {
	SupplierUuid uint64  `json:"supplier_uuid"` // 供应商ID
	SupplierName string  `json:"supplier_name"` // 供应商名称
	Count        int     `json:"count"`         // 数量
	Amount       float64 `json:"amount"`        // 金额
}

// TimeStatItem 时间统计项
type TimeStatItem struct {
	Date   string  `json:"date"`   // 日期
	Count  int     `json:"count"`  // 数量
	Amount float64 `json:"amount"` // 金额
}

// SummaryStatItem 汇总统计项
type SummaryStatItem struct {
	TotalCount         int     `json:"total_count"`           // 总订单数
	TotalAmount        float64 `json:"total_amount"`          // 总金额
	PendingCount       int     `json:"pending_count"`         // 待处理数量
	CompletedCount     int     `json:"completed_count"`       // 已完成数量
	CancelledCount     int     `json:"cancelled_count"`       // 已取消数量
	AverageAmount      float64 `json:"average_amount"`        // 平均金额
	CompletionRate     float64 `json:"completion_rate"`       // 完成率
	OnTimeDeliveryRate float64 `json:"on_time_delivery_rate"` // 按时交付率
}

// PurchaseReceiptOrderCreateResp 创建收货单响应
type PurchaseReceiptOrderCreateResp struct {
	Uuid      uint64 `json:"uuid"`       // 收货单ID
	ReceiptNo string `json:"receipt_no"` // 收货单号
}

// PurchaseReceiptOrderListResp 收货单列表响应
type PurchaseReceiptOrderListResp struct {
	List []PurchaseReceiptOrderInfo `json:"list"` // 收货单列表
	Meta dto.PageResponse           `json:"meta"` // 分页信息
}

// PurchaseReceiptOrderInfo 收货单信息
type PurchaseReceiptOrderInfo struct {
	Uuid              uint64 `json:"uuid"`                // 收货单ID
	Status            int    `json:"status"`              // 状态
	OrderNo           string `json:"order_no"`            // 订单编号
	Num               int    `json:"num"`                 // 物品数量
	ExpectArrivalTime int    `json:"expect_arrival_time"` // 期望到货日期
	ReceiveTime       int    `json:"receive_time"`        // 收货日期
}

// PurchaseReceiptItemInfo 收货明细信息
type PurchaseReceiptItemInfo struct {
	Uuid               uint64  `json:"uuid"`                 // 明细ID
	MaterialCode       string  `json:"material_code"`        // 物品编码
	MaterialName       string  `json:"material_name"`        // 物品名称
	Num                float64 `json:"num"`                  // 数量
	UnitName           string  `json:"unit_name"`            // 采购单位名称
	UnitConversionRate float64 `json:"unit_conversion_rate"` // 基准单位转换率
	BaseUnitName       string  `json:"base_unit_name"`       // 基准单位名称
}

// PurchaseReceiptOrderDetailResp 收货单详情响应
type PurchaseReceiptOrderDetailResp struct {
	PurchaseReceiptOrderInfo
	Items []PurchaseReceiptItemInfo `json:"items"` // 收货明细
}

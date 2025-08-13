package resp

// PageResponse 分页响应
type PageResponse struct {
	PageNo   int   `json:"page_no"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页大小
	Total    int64 `json:"total"`     // 总数
}

// PurchaseOrderListResp 采购订单列表响应
type PurchaseOrderListResp struct {
	List []PurchaseOrderInfo `json:"list"` // 采购订单列表
	Meta PageResponse        `json:"meta"` // 分页信息
}

// PurchaseOrderInfo 采购订单信息
type PurchaseOrderInfo struct {
	Uuid                 uint64  `json:"uuid"`                   // 采购订单ID
	OrderNo              string  `json:"order_no"`               // 采购订单编号
	Title                string  `json:"title"`                  // 采购单标题
	SupplierUuid         uint64  `json:"supplier_uuid"`          // 供应商ID
	SupplierName         string  `json:"supplier_name"`          // 供应商名称
	ApplicantUuid        uint64  `json:"applicant_uuid"`         // 申请人ID
	ApplicantName        string  `json:"applicant_name"`         // 申请人姓名
	ApproverUuid         uint64  `json:"approver_uuid"`          // 审批人ID
	ApproverName         string  `json:"approver_name"`          // 审批人姓名
	Status               int     `json:"status"`                 // 状态
	StatusText           string  `json:"status_text"`            // 状态文本
	Priority             int     `json:"priority"`               // 优先级
	PriorityText         string  `json:"priority_text"`          // 优先级文本
	TotalQuantity        float64 `json:"total_quantity"`         // 总数量
	TotalAmount          float64 `json:"total_amount"`           // 总金额
	ExpectedDeliveryTime int     `json:"expected_delivery_time"` // 预期交付时间
	ActualDeliveryTime   int     `json:"actual_delivery_time"`   // 实际交付时间
	ApproveTime          int     `json:"approve_time"`           // 审核时间
	Remark               string  `json:"remark"`                 // 备注
	CreateTime           int     `json:"create_time"`            // 创建时间
	UpdateTime           int     `json:"update_time"`            // 更新时间

	// 扩展字段
	ItemCount         int     `json:"item_count"`         // 明细数量
	CompletionRate    float64 `json:"completion_rate"`    // 完成率
	ReceivedQuantity  float64 `json:"received_quantity"`  // 已收货数量
	RemainingQuantity float64 `json:"remaining_quantity"` // 剩余数量
	IsEditable        bool    `json:"is_editable"`        // 是否可编辑
	CanApprove        bool    `json:"can_approve"`        // 是否可审核
	CanCancel         bool    `json:"can_cancel"`         // 是否可取消
}

// PurchaseOrderDetailResp 采购订单详情响应
type PurchaseOrderDetailResp struct {
	PurchaseOrderInfo
	Items    []PurchaseOrderItemInfo `json:"items"`    // 采购明细
	Logs     []PurchaseOrderLogInfo  `json:"logs"`     // 操作日志
	Receipts []PurchaseReceiptInfo   `json:"receipts"` // 收货记录
}

// PurchaseOrderItemInfo 采购订单商品明细信息
type PurchaseOrderItemInfo struct {
	Uuid              uint64  `json:"uuid"`                // 明细ID
	PurchaseOrderUuid uint64  `json:"purchase_order_uuid"` // 采购订单ID
	ProductUuid       uint64  `json:"product_uuid"`        // 商品ID
	ProductName       string  `json:"product_name"`        // 商品名称
	ProductSku        string  `json:"product_sku"`         // 商品SKU
	ProductUnit       string  `json:"product_unit"`        // 商品单位
	Specification     string  `json:"specification"`       // 规格型号
	Quantity          float64 `json:"quantity"`            // 采购数量
	ReceivedQuantity  float64 `json:"received_quantity"`   // 已收货数量
	UnitPrice         float64 `json:"unit_price"`          // 单价
	TotalPrice        float64 `json:"total_price"`         // 小计金额
	Sort              int     `json:"sort"`                // 排序
	Remark            string  `json:"remark"`              // 商品备注
	CreateTime        int     `json:"create_time"`         // 创建时间
	UpdateTime        int     `json:"update_time"`         // 更新时间

	// 扩展字段
	CompletionRate    float64 `json:"completion_rate"`    // 完成率
	RemainingQuantity float64 `json:"remaining_quantity"` // 剩余数量
	IsCompleted       bool    `json:"is_completed"`       // 是否完成
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

// PurchaseReceiptInfo 收货记录信息
type PurchaseReceiptInfo struct {
	Uuid              uint64                    `json:"uuid"`                // 收货记录ID
	ReceiptNo         string                    `json:"receipt_no"`          // 收货单号
	PurchaseOrderUuid uint64                    `json:"purchase_order_uuid"` // 采购订单ID
	ReceiverUuid      uint64                    `json:"receiver_uuid"`       // 收货人ID
	ReceiverName      string                    `json:"receiver_name"`       // 收货人姓名
	ReceiptTime       int                       `json:"receipt_time"`        // 收货时间
	TotalQuantity     float64                   `json:"total_quantity"`      // 收货总数量
	TotalAmount       float64                   `json:"total_amount"`        // 收货总金额
	Status            int                       `json:"status"`              // 状态
	StatusText        string                    `json:"status_text"`         // 状态文本
	Remark            string                    `json:"remark"`              // 收货备注
	CreateTime        int                       `json:"create_time"`         // 创建时间
	UpdateTime        int                       `json:"update_time"`         // 更新时间
	Items             []PurchaseReceiptItemInfo `json:"items,omitempty"`     // 收货明细
}

// PurchaseReceiptItemInfo 收货明细信息
type PurchaseReceiptItemInfo struct {
	Uuid                  uint64  `json:"uuid"`                     // 收货明细ID
	PurchaseReceiptUuid   uint64  `json:"purchase_receipt_uuid"`    // 收货记录ID
	PurchaseOrderItemUuid uint64  `json:"purchase_order_item_uuid"` // 采购订单商品明细ID
	ProductUuid           uint64  `json:"product_uuid"`             // 商品ID
	ProductName           string  `json:"product_name"`             // 商品名称
	OrderedQuantity       float64 `json:"ordered_quantity"`         // 订购数量
	ReceivedQuantity      float64 `json:"received_quantity"`        // 实收数量
	QualifiedQuantity     float64 `json:"qualified_quantity"`       // 合格数量
	UnitPrice             float64 `json:"unit_price"`               // 单价
	TotalPrice            float64 `json:"total_price"`              // 小计金额
	QualityStatus         int     `json:"quality_status"`           // 质检状态
	QualityStatusText     string  `json:"quality_status_text"`      // 质检状态文本
	QualityRemark         string  `json:"quality_remark"`           // 质检备注
	CreateTime            int     `json:"create_time"`              // 创建时间
	UpdateTime            int     `json:"update_time"`              // 更新时间

	// 扩展字段
	QualityRate       float64 `json:"quality_rate"`       // 质量合格率
	DefectiveQuantity float64 `json:"defective_quantity"` // 不合格数量
}

// PurchaseReceiptListResp 收货记录列表响应
type PurchaseReceiptListResp struct {
	List []PurchaseReceiptInfo `json:"list"` // 收货记录列表
	Meta PageResponse          `json:"meta"` // 分页信息
}

// PurchaseReceiptDetailResp 收货记录详情响应
type PurchaseReceiptDetailResp struct {
	PurchaseReceiptInfo
}

// PurchaseOrderCreateResp 创建采购订单响应
type PurchaseOrderCreateResp struct {
	Uuid    uint64 `json:"uuid"`     // 采购订单ID
	OrderNo string `json:"order_no"` // 采购订单编号
}

// PurchaseOrderUpdateResp 更新采购订单响应
type PurchaseOrderUpdateResp struct {
	Success bool `json:"success"` // 是否成功
}

// PurchaseOrderDeleteResp 删除采购订单响应
type PurchaseOrderDeleteResp struct {
	Success bool `json:"success"` // 是否成功
}

// PurchaseOrderApproveResp 审核采购订单响应
type PurchaseOrderApproveResp struct {
	Success bool `json:"success"` // 是否成功
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
	List []PurchaseReceiptInfo `json:"list"` // 收货单列表
	Meta PageResponse          `json:"meta"` // 分页信息
}

// PurchaseReceiptOrderDetailResp 收货单详情响应
type PurchaseReceiptOrderDetailResp struct {
	PurchaseReceiptInfo
}

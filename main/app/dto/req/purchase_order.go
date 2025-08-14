package req

import "ttpos-server-go/app/dto"

// PurchaseOrderListReq 采购订单列表请求
type PurchaseOrderListReq struct {
	dto.PageReq        // 分页参数
	OrderNo     string `json:"order_no" form:"order_no" binding:"omitempty,max=50"`        // 订单编号
	StatusIn    []int  `json:"status_in" form:"status_in" binding:"omitempty,min=0,max=5"` // 状态筛选: [0,1,2,3,4,5], 0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货
}

// PurchaseOrderCreateReq 创建采购订单请求
type PurchaseOrderCreateReq struct {
	BillDate             string                       `json:"bill_date" binding:"required,date"`                // 单据日期
	OrderType            int                          `json:"order_type" binding:"required,oneof=0"`            // 申请类型 0-仓库调拨
	ExpectedDeliveryTime int64                        `json:"expected_delivery_time" binding:"omitempty,min=0"` // 期望到货时间(时间戳)
	Items                []PurchaseOrderItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`      // 物品明细
}

// PurchaseOrderItemCreateReq 采购订单物品明细创建请求
type PurchaseOrderItemCreateReq struct {
	MaterialUuid uint64  `json:"material_uuid" binding:"required,min=1"` // 物品ID
	Num          float64 `json:"num" binding:"required,gt=0,lte=99999"`  // 数量
}

// PurchaseOrderUpdateReq 更新采购订单请求
type PurchaseOrderUpdateReq struct {
	Uuid                 uint64                       `json:"uuid" binding:"required,min=1"`                    // 采购订单ID
	ExpectedDeliveryTime int64                        `json:"expected_delivery_time" binding:"omitempty,min=0"` // 期望到货时间(时间戳)
	OrderType            int                          `json:"order_type" binding:"required,oneof=0"`            // 申请类型 0-仓库调拨
	Items                []PurchaseOrderItemUpdateReq `json:"items" binding:"omitempty,min=1,max=200,dive"`     // 采购商品明细
}

// PurchaseOrderItemUpdateReq 采购订单商品明细更新请求
type PurchaseOrderItemUpdateReq struct {
	MaterialUuid uint64  `json:"material_uuid" binding:"required,min=1"` // 物品ID
	Num          float64 `json:"num" binding:"required,gt=0,lte=99999"`  // 数量
}

// PurchaseOrderDetailReq 采购订单详情请求
type PurchaseOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 采购订单ID
}

// PurchaseOrderDeleteReq 删除采购订单请求
type PurchaseOrderDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 采购订单ID
}

// PurchaseOrderApproveReq 审核采购订单请求
type PurchaseOrderApproveReq struct {
	Uuid   uint64 `json:"uuid" binding:"required,min=1"`                  // 采购订单ID
	Action string `json:"action" binding:"required,oneof=approve reject"` // 审核动作：approve-通过，reject-驳回
}

// PurchaseOrderSubmitReq 提交采购订单请求
type PurchaseOrderSubmitReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 采购订单ID
}

// PurchaseReceiptCreateReq 创建收货记录请求
type PurchaseReceiptCreateReq struct {
	PurchaseOrderUuid uint64                         `json:"purchase_order_uuid" binding:"required,min=1"`    // 采购订单ID
	ReceiptTime       int64                          `json:"receipt_time" binding:"required,min=0"`           // 收货时间(时间戳)
	Items             []PurchaseReceiptItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`     // 收货明细
	IsConfirm         bool                           `json:"is_confirm" binding:"omitempty,oneof=true false"` // 是否确认收货
}

// PurchaseReceiptUpdateReq 更新收货记录请求
type PurchaseReceiptOrderUpdateReq struct {
	Uuid        uint64                         `json:"uuid" binding:"required,min=1"`                   // 收货记录ID
	ReceiptTime int                            `json:"receipt_time" binding:"required,min=0"`           // 收货时间(时间戳)
	Items       []PurchaseReceiptItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`     // 收货明细
	IsConfirm   bool                           `json:"is_confirm" binding:"omitempty,oneof=true false"` // 是否确认收货
}

// PurchaseReceiptItemCreateReq 收货明细创建请求
type PurchaseReceiptItemCreateReq struct {
	Uuid uint64  `json:"uuid" binding:"required,min=1"`          // 商品明细ID
	Num  float64 `json:"num" binding:"required,gte=0,lte=99999"` // 实收数量
}

// PurchaseReceiptListReq 收货记录列表请求
type PurchaseReceiptListReq struct {
	dto.PageReq              // 分页参数
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid" form:"purchase_order_uuid" binding:"omitempty,min=1"` // 采购订单ID
	ReceiptNo         string `json:"receipt_no" form:"receipt_no" binding:"omitempty,max=50"`                  // 收货单号
	QualityStatus     *int   `json:"quality_status" form:"quality_status" binding:"omitempty,min=1,max=3"`     // 质检状态筛选
	ReceiptTimeStart  int    `json:"receipt_time_start" form:"receipt_time_start" binding:"omitempty,min=0"`   // 收货时间开始
	ReceiptTimeEnd    int    `json:"receipt_time_end" form:"receipt_time_end" binding:"omitempty,min=0"`       // 收货时间结束
}

// PurchaseReceiptDetailReq 收货记录详情请求
type PurchaseReceiptDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 收货记录ID
}

// PurchaseOrderStatisticsReq 采购订单统计请求
type PurchaseOrderStatisticsReq struct {
	TimeStart int    `json:"time_start" form:"time_start" binding:"omitempty,min=0"`         // 统计开始时间
	TimeEnd   int    `json:"time_end" form:"time_end" binding:"omitempty,min=0"`             // 统计结束时间
	Type      string `json:"type" form:"type" binding:"omitempty,oneof=day week month year"` // 统计类型
}

// PurchaseReceiptOrderListReq 收货单列表请求
type PurchaseReceiptOrderListReq struct {
	dto.PageReq        // 分页参数
	OrderNo     string `json:"order_no" form:"order_no" binding:"omitempty,max=50"`        // 订单编号
	StatusIn    []int  `json:"status_in" form:"status_in" binding:"omitempty,min=0,max=5"` // 状态筛选: [0,1,2], 0-待收货 1-已收货 2-已取消
}

// PurchaseReceiptOrderDetailReq 收货单详情请求
type PurchaseReceiptOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 收货单ID
}

// PurchaseReceiptOrderCancelReq 取消收货单请求
type PurchaseReceiptOrderCancelReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 收货单ID
}

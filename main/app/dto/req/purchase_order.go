package req

import "ttpos-server-go/app/dto"

// PurchaseOrderListReq 采购订单列表请求
type PurchaseOrderListReq struct {
	dto.PageReq              // 分页参数
	OrderNo           string `json:"order_no" form:"order_no" binding:"omitempty,max=50"`                      // 订单编号
	Title             string `json:"title" form:"title" binding:"omitempty,max=255"`                           // 采购单标题
	Status            *int   `json:"status" form:"status" binding:"omitempty,min=0,max=7"`                     // 状态筛选
	Priority          *int   `json:"priority" form:"priority" binding:"omitempty,min=1,max=4"`                 // 优先级筛选
	SupplierUuid      uint64 `json:"supplier_uuid" form:"supplier_uuid" binding:"omitempty,min=1"`             // 供应商ID
	ApplicantUuid     uint64 `json:"applicant_uuid" form:"applicant_uuid" binding:"omitempty,min=1"`           // 申请人ID
	CreateTimeStart   int    `json:"create_time_start" form:"create_time_start" binding:"omitempty,min=0"`     // 创建时间开始
	CreateTimeEnd     int    `json:"create_time_end" form:"create_time_end" binding:"omitempty,min=0"`         // 创建时间结束
	DeliveryTimeStart int    `json:"delivery_time_start" form:"delivery_time_start" binding:"omitempty,min=0"` // 交付时间开始
	DeliveryTimeEnd   int    `json:"delivery_time_end" form:"delivery_time_end" binding:"omitempty,min=0"`     // 交付时间结束
}

// PurchaseOrderCreateReq 创建采购订单请求
type PurchaseOrderCreateReq struct {
	BillDate             string                       `json:"bill_date" binding:"required,date"`                // 单据日期
	OrderType            int                          `json:"order_type" binding:"required,oneof=1"`            // 申请类型 1-仓库调拨
	ExpectedDeliveryTime int64                        `json:"expected_delivery_time" binding:"omitempty,min=0"` // 预期交付时间(时间戳)
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
	Title                string                       `json:"title" binding:"required,max=255"`                 // 采购单标题
	SupplierUuid         uint64                       `json:"supplier_uuid" binding:"required,min=1"`           // 供应商ID
	SupplierName         string                       `json:"supplier_name" binding:"required,max=255"`         // 供应商名称
	Priority             int                          `json:"priority" binding:"omitempty,min=1,max=4"`         // 优先级 1-低 2-中 3-高 4-紧急
	ExpectedDeliveryTime int                          `json:"expected_delivery_time" binding:"omitempty,min=0"` // 预期交付时间(时间戳)
	Remark               string                       `json:"remark" binding:"omitempty,max=1000"`              // 备注
	Items                []PurchaseOrderItemUpdateReq `json:"items" binding:"required,min=1,max=200,dive"`      // 采购商品明细
}

// PurchaseOrderItemUpdateReq 采购订单商品明细更新请求
type PurchaseOrderItemUpdateReq struct {
	Uuid          uint64  `json:"uuid" binding:"omitempty,min=0"`            // 明细ID，0表示新增
	ProductUuid   uint64  `json:"product_uuid" binding:"required,min=1"`     // 商品ID
	ProductName   string  `json:"product_name" binding:"required,max=255"`   // 商品名称
	ProductSku    string  `json:"product_sku" binding:"omitempty,max=100"`   // 商品SKU
	ProductUnit   string  `json:"product_unit" binding:"required,max=50"`    // 商品单位
	Specification string  `json:"specification" binding:"omitempty,max=255"` // 规格型号
	Num           float64 `json:"num" binding:"required,gt=0,lte=99999"`     // 采购数量
	UnitPrice     float64 `json:"unit_price" binding:"required,gte=0"`       // 单价
	Remark        string  `json:"remark" binding:"omitempty,max=500"`        // 商品备注
	IsDeleted     bool    `json:"is_deleted" binding:"omitempty"`            // 是否删除
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
	Remark string `json:"remark" binding:"omitempty,max=500"`             // 审核备注
}

// PurchaseOrderStatusUpdateReq 更新采购订单状态请求
type PurchaseOrderStatusUpdateReq struct {
	Uuid   uint64 `json:"uuid" binding:"required,min=1"`         // 采购订单ID
	Status int    `json:"status" binding:"required,min=0,max=7"` // 新状态
	Remark string `json:"remark" binding:"omitempty,max=500"`    // 备注
}

// PurchaseReceiptCreateReq 创建收货记录请求
type PurchaseReceiptCreateReq struct {
	PurchaseOrderUuid uint64                         `json:"purchase_order_uuid" binding:"required,min=1"`   // 采购订单ID
	ReceiptTime       int                            `json:"receipt_time" binding:"required,min=0"`          // 收货时间(时间戳)
	QualityStatus     int                            `json:"quality_status" binding:"omitempty,min=1,max=3"` // 质检状态 1-合格 2-不合格 3-部分合格
	Remark            string                         `json:"remark" binding:"omitempty,max=1000"`            // 收货备注
	Items             []PurchaseReceiptItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`    // 收货明细
}

// PurchaseReceiptItemCreateReq 收货明细创建请求
type PurchaseReceiptItemCreateReq struct {
	PurchaseOrderItemUuid uint64  `json:"purchase_order_item_uuid" binding:"required,min=1"` // 采购订单商品明细ID
	ReceivedNum           float64 `json:"received_num" binding:"required,gte=0,lte=99999"`   // 实收数量
	QualifiedNum          float64 `json:"qualified_num" binding:"required,gte=0,lte=99999"`  // 合格数量
	QualityStatus         int     `json:"quality_status" binding:"omitempty,min=1,max=3"`    // 质检状态 1-合格 2-不合格 3-部分合格
	QualityRemark         string  `json:"quality_remark" binding:"omitempty,max=500"`        // 质检备注
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
	dto.PageReq              // 分页参数
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid" form:"purchase_order_uuid" binding:"omitempty,min=1"` // 采购订单ID
	ReceiptNo         string `json:"receipt_no" form:"receipt_no" binding:"omitempty,max=50"`                  // 收货单号
	ReceiptTimeStart  int    `json:"receipt_time_start" form:"receipt_time_start" binding:"omitempty,min=0"`   // 收货时间开始
	ReceiptTimeEnd    int    `json:"receipt_time_end" form:"receipt_time_end" binding:"omitempty,min=0"`       // 收货时间结束
}

// PurchaseReceiptOrderDetailReq 收货单详情请求
type PurchaseReceiptOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 收货单ID
}

// PurchaseReceiptOrderDeleteReq 删除收货单请求
type PurchaseReceiptOrderDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 收货单ID
}

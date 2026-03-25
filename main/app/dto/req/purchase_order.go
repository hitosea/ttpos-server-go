package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// PurchaseOrderListReq 采购订单列表请求
type PurchaseOrderListReq struct {
	dto.PageReq                     // 分页参数
	OrderNo                string   `json:"order_no" form:"order_no" binding:"omitempty,max=50"`                                  // 订单编号
	UuidIn                 []uint64 `json:"uuid_in" form:"uuid_in" binding:"omitempty,min=1"`                                     // 采购订单ID
	PurchaseType           int      `json:"purchase_type" form:"purchase_type" binding:"omitempty,min=0,max=2"`                   // V2.6 采购类型, 1-外部采购 2-内部采购
	StatusIn               []int    `json:"status_in" form:"status_in" binding:"omitempty,min=0,max=5"`                           // 状态筛选: [0,1,2,3,4], 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成) 5-待总部审核
	SupplierName           string   `json:"supplier_name" form:"supplier_name" binding:"omitempty,max=100"`                       // 供应商名称
	WarehouseErpCode       string   `json:"warehouse_erp_code" form:"warehouse_erp_code" binding:"omitempty,max=255"`             // 仓库编码
	CompanyUuid            uint64   `json:"company_uuid" form:"company_uuid" binding:"omitempty,min=1"`                           // 公司UUID
	CreateTimeStart        int      `json:"create_time_start" form:"create_time_start" binding:"omitempty,min=0"`                 // 创建时间开始
	CreateTimeEnd          int      `json:"create_time_end" form:"create_time_end" binding:"omitempty,min=0"`                     // 创建时间结束
	OrderTimeStart         int      `json:"order_time_start" form:"order_time_start" binding:"omitempty,min=0"`                   // 订单时间开始
	OrderTimeEnd           int      `json:"order_time_end" form:"order_time_end" binding:"omitempty,min=0"`                       // 订单时间结束
	ExpectArrivalTimeStart int      `json:"expect_arrival_time_start" form:"expect_arrival_time_start" binding:"omitempty,min=0"` // 期望到货时间开始
	ExpectArrivalTimeEnd   int      `json:"expect_arrival_time_end" form:"expect_arrival_time_end" binding:"omitempty,min=0"`     // 期望到货时间结束
	ReceiveTimeStart       int      `json:"receive_time_start" form:"receive_time_start" binding:"omitempty,min=0"`               // 收货时间开始
	ReceiveTimeEnd         int      `json:"receive_time_end" form:"receive_time_end" binding:"omitempty,min=0"`                   // 收货时间结束
}

// PurchaseOrderCreateReq 创建采购订单请求
type PurchaseOrderCreateReq struct {
	OrderTime            int64                        `json:"order_time" binding:"required,min=0"`              // 单据日期
	SupplierName         string                       `json:"supplier_name" binding:"required,min=1"`           // 供应商名称
	SupplierErpCode      string                       `json:"supplier_erp_code" binding:"omitempty,max=255"`    // V2.6 供应商编码
	ExpectedDeliveryTime int64                        `json:"expected_delivery_time" binding:"omitempty,min=0"` // 期望到货时间(时间戳)
	Items                []PurchaseOrderItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`      // 物品明细
	PurchaseType         int                          `json:"purchase_type" binding:"omitempty,min=0,max=2"`    // V2.6  采购类型, 1-外部采购 2-内部采购
	WarehouseErpCode     string                       `json:"warehouse_erp_code" binding:"omitempty,max=255"`   // V2.6 仓库编码
}

func (r *PurchaseOrderCreateReq) Validate() error {
	if r.SupplierName == "" {
		return errors.New("供应商名称不能为空")
	}
	if len(r.Items) == 0 {
		return errors.New("物品明细不能为空")
	}
	for _, item := range r.Items {
		if item.MaterialUuid == 0 {
			return errors.New("物品ID不能为空")
		}
	}
	return nil
}

type PurchaseOrderItemMaterialUnitReq struct {
	Uuid uint64  `json:"uuid" binding:"required"` // 单位UUID
	Num  float64 `json:"num"`                     // 数量
}

// PurchaseOrderItemCreateReq 采购订单物品明细创建请求
type PurchaseOrderItemCreateReq struct {
	MaterialUuid uint64                             `json:"material_uuid" binding:"required,min=1"` // 物品ID
	UnitList     []PurchaseOrderItemMaterialUnitReq `json:"unit_list"`                              // 单位列表,新增的非基准单位
}

// PurchaseOrderUpdateReq 更新采购订单请求
type PurchaseOrderUpdateReq struct {
	Uuid                 uint64                       `json:"uuid" binding:"required,min=1"`                    // 采购订单ID
	ExpectedDeliveryTime int64                        `json:"expected_delivery_time" binding:"omitempty,min=0"` // 期望到货时间(时间戳)
	SupplierName         string                       `json:"supplier_name" binding:"omitempty,min=1"`          // 供应商名称
	SupplierErpCode      string                       `json:"supplier_erp_code" binding:"omitempty,max=255"`    // V2.6 供应商编码
	WarehouseErpCode     string                       `json:"warehouse_erp_code" binding:"omitempty,max=255"`   // V2.6 仓库编码
	Items                []PurchaseOrderItemUpdateReq `json:"items" binding:"omitempty,min=1,max=200,dive"`     // 采购商品明细
}

func (r *PurchaseOrderUpdateReq) Validate() error {
	if r.Uuid == 0 {
		return errors.New("采购订单ID不能为空")
	}
	if r.SupplierName == "" {
		return errors.New("供应商名称不能为空")
	}
	if len(r.Items) == 0 {
		return errors.New("物品明细不能为空")
	}
	for _, item := range r.Items {
		if item.MaterialUuid == 0 {
			return errors.New("物品ID不能为空")
		}
	}
	return nil
}

// PurchaseOrderItemUpdateReq 采购订单商品明细更新请求
type PurchaseOrderItemUpdateReq struct {
	MaterialUuid uint64                             `json:"material_uuid" binding:"required,min=1"` // 物品ID
	UnitList     []PurchaseOrderItemMaterialUnitReq `json:"unit_list"`                              // 单位列表,新增的非基准单位
}

// PurchaseOrderDetailReq 采购订单详情请求
type PurchaseOrderDetailReq struct {
	Uuid            uint64 `json:"uuid" form:"uuid" binding:"required,min=1"`  // 采购订单ID
	WithReceiptList bool   `json:"with_receipt_list" form:"with_receipt_list"` // 是否查看收货清单
}

// PurchaseOrderDeleteReq 删除采购订单请求
type PurchaseOrderDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 采购订单ID
}

// PurchaseOrderApproveReq 审核采购订单请求
type PurchaseOrderApproveReq struct {
	Uuid         uint64 `json:"uuid" binding:"required,min=1"`                  // 采购订单ID
	Action       string `json:"action" binding:"required,oneof=approve reject"` // 审核动作：approve-通过，reject-驳回
	RejectReason string `json:"reject_reason"`                                  // 驳回备注，可以为空最大100个字符
	Remark       string `json:"remark"`                                         // 批注, 可以为空最大100个字符
	IsConfirm    bool   `json:"is_confirm"`                                     // 是否确认审核（跳过默认仓库校验提示）
}

// PurchaseOrderSubmitReq 提交采购订单请求
type PurchaseOrderSubmitReq struct {
	Uuid       uint64 `json:"uuid" binding:"required,min=1"` // 采购订单ID
	IsConfirm  bool   `json:"is_confirm"`                    // 是否确认提交（移除禁用物品）
	IsResubmit bool   `json:"is_resubmit"`                   // 是否重新提交
}

// PurchaseOrderSupplierSyncReq 供应商同步确认请求
type PurchaseOrderSupplierSyncReq struct {
	Uuid            uint64 `json:"uuid" binding:"required,min=1"`        // 采购订单UUID
	SupplierErpCode string `json:"supplier_erp_code" binding:"required"` // 供应商ERP编码
	ConfirmSync     bool   `json:"confirm_sync" binding:"required"`      // 是否确认同步
}

// PurchaseReceiptCreateReq 创建收货记录请求
type PurchaseReceiptCreateReq struct {
	PurchaseOrderUuid      uint64                         `json:"purchase_order_uuid" binding:"required,min=1"`     // 采购订单ID
	ReceiveTime            int64                          `json:"receive_time" binding:"required,min=0"`            // 收货时间(时间戳)
	ReceiptType            int                            `json:"receipt_type" binding:"required,min=1,max=2"`      // 收货类型 1-外部收货 2-内部收货
	Items                  []PurchaseReceiptItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"`      // 收货明细
	IsConfirm              bool                           `json:"is_confirm"`                                       // 是否确认收货
	FileUuids              []uint64                       `json:"file_uuids" binding:"omitempty,max=10"`            // 附件UUID列表，最多10个
	DeliveryNoteNo         string                         `json:"delivery_note_no" binding:"omitempty,max=255"`     // v2.16.0+ DN单号
	SourceSupplierCode     string                         `json:"source_supplier_code" binding:"omitempty,max=255"` // v2.16.0+ 供应商编码
	IsAutoReceipt          bool                           `json:"-"`                                                // 是否自动收货（仅内部使用，不接受外部传参）
	SourceWarehouseErpCode string                         `json:"-"`                                                // 发货仓库ERP编码（仅内部使用，DN收货时从DN获取）
}

// PurchaseReceiptUpdateReq 更新收货记录请求
type PurchaseReceiptOrderUpdateReq struct {
	Uuid        uint64                         `json:"uuid" binding:"required,min=1"`               // 收货记录ID
	ReceiveTime int64                          `json:"receive_time" binding:"required,min=0"`       // 收货时间(时间戳)
	Items       []PurchaseReceiptItemCreateReq `json:"items" binding:"required,min=1,max=200,dive"` // 收货明细
	IsConfirm   bool                           `json:"is_confirm"`                                  // 是否确认收货
	FileUuids   []uint64                       `json:"file_uuids" binding:"omitempty,max=10"`       // 附件UUID列表，最多10个
}

type PurchaseReceiptItemMaterialUnitReq struct {
	Uuid uint64  `json:"uuid" binding:"required"` // 单位UUID
	Num  float64 `json:"num"`                     // 数量
}

// PurchaseReceiptItemCreateReq 收货明细创建请求
type PurchaseReceiptItemCreateReq struct {
	Uuid                  uint64                               `json:"uuid"`                                        // 收货明细ID
	PurchaseOrderItemUuid uint64                               `json:"purchase_order_item_uuid" binding:"required"` // 采购订单明细ID
	UnitList              []PurchaseReceiptItemMaterialUnitReq `json:"unit_list"`                                   // 单位列表,新增的非基准单位
}

// PurchaseReceiptListReq 收货记录列表请求
type PurchaseReceiptListReq struct {
	dto.PageReq              // 分页参数
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid" form:"purchase_order_uuid" binding:"omitempty,min=1"` // 采购订单ID
	ReceiptNo         string `json:"receipt_no" form:"receipt_no" binding:"omitempty,max=50"`                  // 收货单号
	QualityStatus     *int   `json:"quality_status" form:"quality_status" binding:"omitempty,min=1,max=3"`     // 质检状态筛选
	ReceiveTimeStart  int    `json:"receive_time_start" form:"receive_time_start" binding:"omitempty,min=0"`   // 收货时间开始
	ReceiveTimeEnd    int    `json:"receive_time_end" form:"receive_time_end" binding:"omitempty,min=0"`       // 收货时间结束
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
	dto.PageReq               // 分页参数
	ReceiptType      int      `json:"receipt_type" form:"receipt_type" binding:"omitempty,min=0,max=2"`       // V2.6 收货类型, 1-外部收货 2-内部收货
	OrderNo          string   `json:"order_no" form:"order_no" binding:"omitempty,max=50"`                    // 订单编号
	UuidIn           []uint64 `json:"uuid_in" form:"uuid_in" binding:"omitempty,min=1"`                       // 收货单ID
	StatusIn         []int    `json:"status_in" form:"status_in" binding:"omitempty,min=0,max=5"`             // 状态筛选: [0,1,2], 0-待收货 1-已收货 2-已取消
	ReceiptTimeStart int64    `json:"receipt_time_start" form:"receipt_time_start" binding:"omitempty,min=0"` // V2.6 收货时间开始（时间戳）
	ReceiptTimeEnd   int64    `json:"receipt_time_end" form:"receipt_time_end" binding:"omitempty,min=0"`     // V2.6 收货时间结束（时间戳）
	CreateTimeStart  int64    `json:"create_time_start" form:"create_time_start" binding:"omitempty,min=0"`   // V2.6 创建时间开始（时间戳）
	CreateTimeEnd    int64    `json:"create_time_end" form:"create_time_end" binding:"omitempty,min=0"`       // V2.6 创建时间结束（时间戳）
}

// PurchaseReceiptOrderDetailReq 收货单详情请求
type PurchaseReceiptOrderDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required,min=1"` // 收货单ID
}

// PurchaseReceiptOrderCancelReq 取消收货单请求
type PurchaseReceiptOrderCancelReq struct {
	Uuid uint64 `json:"uuid" binding:"required,min=1"` // 收货单ID
}

// DeleteReceiptFileReq 删除收货单附件请求
type DeleteReceiptFileReq struct {
	ReceiptOrderUuid uint64 `json:"receipt_order_uuid" binding:"required,min=1"` // 收货单UUID
	FileUuid         uint64 `json:"file_uuid" binding:"required,min=1"`          // 文件UUID
}

// ReceiptPendingItemsReq 获取待收货物品请求 (v2.16.0+)
type ReceiptPendingItemsReq struct {
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid" form:"purchase_order_uuid" binding:"required,min=1"` // 采购订单UUID
	DeliveryNoteNo    string `json:"delivery_note_no" form:"delivery_note_no" binding:"omitempty,max=255"`    // DN单号（与SupplierErpCode二选一）
	SupplierErpCode   string `json:"supplier_erp_code" form:"supplier_erp_code" binding:"omitempty,max=255"`  // 供应商ERP编码（与DeliveryNoteNo二选一）
}

// PurchaseReceiptNewListReq 新收货单列表请求（按采购单维度）
type PurchaseReceiptNewListReq struct {
	dto.PageReq          // 分页参数
	ReceiptStatus *int   `json:"receipt_status" form:"receipt_status" binding:"omitempty,min=0,max=1"` // 采购单收货状态：0-待收货（未完全收货）1-已收货（已完全收货）
	OrderNo       string `json:"order_no" form:"order_no" binding:"omitempty,max=50"`                  // 单据编号（采购单OrderNo）
	ErpOrderNo    string `json:"erp_order_no" form:"erp_order_no" binding:"omitempty,max=50"`          // 采购单号（ErpOrderNo）
}

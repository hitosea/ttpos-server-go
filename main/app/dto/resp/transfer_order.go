package resp

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp/material_resp"
)

// TransferOrderListResp 调拨单列表响应
type TransferOrderListResp struct {
	List []*TransferOrderInfo `json:"list"` // 调拨单列表
	Meta dto.PageResponse     `json:"meta"` // 分页信息
}

// TransferOrderInfo 调拨单信息
type TransferOrderInfo struct {
	Uuid                     uint64             `json:"uuid"`                         // 调拨单UUID
	CompanyUuid              uint64             `json:"company_uuid"`                 // 所属公司UUID (发起门店)
	CompanyName              string             `json:"company_name"`                 // 所属公司名称 (发起门店)
	CompanyStoreCode         string             `json:"company_store_code"`           // 公司店铺编码
	OrderNo                  string             `json:"order_no"`                     // 单据编号
	ErpOrderNo               string             `json:"erp_order_no"`                 // ERP调拨单号
	TransferType             int                `json:"transfer_type"`                // 调拨类型: 1-调入 2-调出
	SenderCompanyUuid        uint64             `json:"sender_company_uuid"`          // 发货门店UUID
	SenderCompanyName        string             `json:"sender_company_name"`          // 发货门店名称
	SenderCompanyStoreCode   string             `json:"sender_company_store_code"`    // 发货门店编码
	ReceiverCompanyUuid      uint64             `json:"receiver_company_uuid"`        // 收货门店UUID
	ReceiverCompanyName      string             `json:"receiver_company_name"`        // 收货门店名称
	ReceiverCompanyStoreCode string             `json:"receiver_company_store_code"`  // 收货门店编码
	OutWarehouseErpCode      string             `json:"out_warehouse_erp_code"`       // 出库仓库ERP编码
	OutWarehouseName         dto.LocaleResponse `json:"out_warehouse_name"`           // 出库仓库名称
	InWarehouseErpCode       string             `json:"in_warehouse_erp_code"`        // 入库仓库ERP编码
	InWarehouseName          dto.LocaleResponse `json:"in_warehouse_name"`            // 入库仓库名称
	OrderTime                int64              `json:"order_time"`                   // 单据日期
	SubmitTime               int64              `json:"submit_time"`                  // 提交时间
	Status                   int                `json:"status"`                       // 状态: 0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成
	ApprovalProgress         string             `json:"approval_progress"`            // 审批中进度: wait-待审核 sender-待发货门店审批 sender_parent-待发货门店上级审批 receiver-待收货门店审批 receiver_parent-待收货门店上级审批 parent-待上级审批
	NextApprovalCompanyUuid  uint64             `json:"next_approval_company_uuid"`   // 下一个审批门店UUID
	NextApprovalCompanyName  string             `json:"next_approval_company_name"`   // 下一个审批门店名称
	Remark                   string             `json:"remark"`                       // 备注
	ItemCount                int                `json:"item_count"`                   // 物品种类数量
	CreateTime               int                `json:"create_time"`                  // 创建时间
	UpdateTime               int                `json:"update_time"`                  // 更新时间
	IsCanApprove             bool               `json:"is_can_approve"`               // 是否可审批
	IsCanReceive             bool               `json:"is_can_receive"`               // 是否可收货
	IsCanResubmit            bool               `json:"is_can_resubmit"`              // 是否可重新提交（已驳回状态且为发起人）
	IsNeedSelectOutWarehouse bool               `json:"is_need_select_out_warehouse"` // 是否需要选择出库仓库
	IsNeedSelectInWarehouse  bool               `json:"is_need_select_in_warehouse"`  // 是否需要选择入库仓库
}

// TransferOrderDetailResp 调拨单详情响应
type TransferOrderDetailResp struct {
	TransferOrderInfo
	Items       []TransferOrderItemInfo       `json:"items"`       // 调拨明细
	RejectInfo  TransferOrderRejectInfo       `json:"reject_info"` // 驳回信息
	Annotations []TransferOrderAnnotationItem `json:"annotations"` // 批注列表
	Files       []TransferOrderFileInfo       `json:"files"`       // 附件列表
}

// TransferOrderItemInfo 调拨单明细信息
type TransferOrderItemInfo struct {
	Uuid                 uint64                       `json:"uuid"`                   // 明细UUID
	MaterialUuid         uint64                       `json:"material_uuid"`          // 物品UUID
	MaterialCode         string                       `json:"material_code"`          // 物品编码
	MaterialName         dto.LocaleResponse           `json:"material_name"`          // 物品名称
	MaterialBarcode      string                       `json:"material_barcode"`       // 条形码值
	MaterialInternalCode string                       `json:"material_internal_code"` // 物品内部编码
	Valuation            float64                      `json:"valuation"`              // 估值单价（基准单位）
	AvailableNum         float64                      `json:"available_num"`          // 可用库存数量
	UnitLocaleName       dto.LocaleResponse           `json:"unit_locale_name"`       // 单位名称
	Units                []TransferOrderItemUnitInfo  `json:"units"`                  // 单位列表
	UnitList             []material_resp.MaterialUnit `json:"unit_list"`              // 单位列表
}

// TransferOrderItemUnitInfo 调拨单明细单位信息
type TransferOrderItemUnitInfo struct {
	Uuid               uint64             `json:"uuid"`                 // 明细单位UUID
	UnitUuid           uint64             `json:"unit_uuid"`            // 单位UUID
	UnitName           dto.LocaleResponse `json:"unit_name"`            // 单位名称
	UnitConversionRate float64            `json:"unit_conversion_rate"` // 单位转换率
	Num                float64            `json:"num"`                  // 调拨数量
	ErpnextUom         string             `json:"erpnext_uom"`          // ERP单位
}

// TransferOrderApprovalInfo 调拨单审批流程信息
type TransferOrderApprovalInfo struct {
	Uuid                uint64 `json:"uuid"`                  // 审批UUID
	TransferOrderUuid   uint64 `json:"transfer_order_uuid"`   // 调拨单UUID
	ApprovalType        string `json:"approval_type"`         // 审批类型: initiator/sender/sender_parent/receiver_parent/receiver
	ApprovalCompanyUuid uint64 `json:"approval_company_uuid"` // 审批门店UUID
	ApprovalCompanyName string `json:"approval_company_name"` // 审批门店名称
	Sequence            int    `json:"sequence"`              // 审批顺序
	Status              int    `json:"status"`                // 审批状态: 0-待审批 1-已通过 2-已驳回 3-已跳过
	ApproverUuid        uint64 `json:"approver_uuid"`         // 审批人UUID
	ApproverName        string `json:"approver_name"`         // 审批人姓名
	ApproveTime         int    `json:"approve_time"`          // 审批时间
	RejectReason        string `json:"reject_reason"`         // 驳回原因
	IsRequired          int    `json:"is_required"`           // 是否必须审批: 0-否 1-是
	Remark              string `json:"remark"`                // 备注
	CreateTime          int    `json:"create_time"`           // 创建时间
}

// TransferOrderRejectInfo 调拨单驳回信息
type TransferOrderRejectInfo struct {
	Title        string `json:"title"`         // 驳回节点标题（如：发货门店驳回）
	RejectReason string `json:"reject_reason"` // 驳回原因
	RejectTime   int64  `json:"reject_time"`   // 驳回时间
}

// TransferOrderLogInfo 调拨单操作日志信息
type TransferOrderLogInfo struct {
	Uuid              uint64 `json:"uuid"`                // 日志UUID
	TransferOrderUuid uint64 `json:"transfer_order_uuid"` // 调拨单UUID
	Action            string `json:"action"`              // 操作动作: create/submit/approve/reject/receive
	ActionDesc        string `json:"action_desc"`         // 操作描述
	OldStatus         int    `json:"old_status"`          // 操作前状态
	NewStatus         int    `json:"new_status"`          // 操作后状态
	OperatorUuid      uint64 `json:"operator_uuid"`       // 操作人UUID
	OperatorName      string `json:"operator_name"`       // 操作人姓名
	OperatorRole      string `json:"operator_role"`       // 操作人角色
	Content           string `json:"content"`             // 操作内容详情JSON
	Remark            string `json:"remark"`              // 备注
	CreateTime        int    `json:"create_time"`         // 创建时间
}

// TransferOrderLogListResp 调拨单操作日志列表响应
type TransferOrderLogListResp struct {
	List []*TransferOrderLogInfo `json:"list"` // 日志列表
	Meta dto.PageResponse        `json:"meta"` // 分页信息
}

// TransferOrderCreateResp 创建调拨单响应
type TransferOrderCreateResp struct {
	Uuid    uint64 `json:"uuid"`     // 调拨单UUID
	OrderNo string `json:"order_no"` // 调拨单编号
}

// TransferOrderApprovalListResp 调拨单审批流程列表响应
type TransferOrderApprovalListResp struct {
	List []*TransferOrderApprovalInfo `json:"list"` // 审批流程列表
}

// TransferOrderStatisticsResp 调拨单统计响应
type TransferOrderStatisticsResp struct {
	StatusStats       []TransferOrderStatusStatItem  `json:"status_stats"`        // 状态统计
	TransferTypeStats []TransferTypeStatItem         `json:"transfer_type_stats"` // 调拨类型统计
	CompanyStats      []TransferOrderCompanyStatItem `json:"company_stats"`       // 门店统计
}

// TransferOrderStatusStatItem 调拨单状态统计项
type TransferOrderStatusStatItem struct {
	Status int `json:"status"` // 状态
	Count  int `json:"count"`  // 数量
}

// TransferTypeStatItem 调拨类型统计项
type TransferTypeStatItem struct {
	TransferType int `json:"transfer_type"` // 调拨类型
	Count        int `json:"count"`         // 数量
}

// TransferOrderCompanyStatItem 调拨单门店统计项
type TransferOrderCompanyStatItem struct {
	CompanyUuid uint64 `json:"company_uuid"` // 门店UUID
	CompanyName string `json:"company_name"` // 门店名称
	Count       int    `json:"count"`        // 数量
}

// TransferOrderCompanyListResp 调拨单门店列表响应
type TransferOrderCompanyListResp struct {
	List []TransferOrderCompanyItem `json:"list"` // 门店列表
}

// TransferOrderCompanyItem 调拨单门店项
type TransferOrderCompanyItem struct {
	Uuid          uint64 `json:"uuid"`           // 门店UUID
	Name          string `json:"name"`           // 门店名称
	StoreCode     string `json:"store_code"`     // 门店编码
	IsHeadquarter bool   `json:"is_headquarter"` // 是否总部
}

// TransferOrderWarehouseListResp 调拨单仓库列表响应
type TransferOrderWarehouseListResp struct {
	List []TransferOrderWarehouseItem `json:"list"` // 仓库列表
}

// TransferOrderWarehouseItem 调拨单仓库项
type TransferOrderWarehouseItem struct {
	ErpCode string             `json:"erp_code"` // 仓库ERP编码
	Name    dto.LocaleResponse `json:"name"`     // 仓库名称（多语言）
	Type    string             `json:"type"`     // 仓库类型
}

// TransferOrderFileInfo 调拨单附件信息
type TransferOrderFileInfo struct {
	FileUuid   uint64 `json:"file_uuid"`   // 文件UUID
	FileName   string `json:"file_name"`   // 文件名
	FileSize   int64  `json:"file_size"`   // 文件大小（字节）
	FileType   string `json:"file_type"`   // 文件类型（image/document）
	Extension  string `json:"extension"`   // 文件扩展名
	FilePath   string `json:"file_path"`   // 文件访问路径
	SortOrder  int    `json:"sort_order"`  // 排序
	CreateTime int    `json:"create_time"` // 创建时间
}

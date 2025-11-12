package model

import (
	"encoding/json"
	"ttpos-bmp/app/ttpos-erp/api/material_transfer"
	"ttpos-server-go/app/constant"

	"github.com/shopspring/decimal"
)

// TransferOrder 调拨单主表 ttpos_transfer_order
type TransferOrder struct {
	BaseModel
	CompanyUuid     uint64 `gorm:"column:company_uuid;type:bigint;default:0;comment:所属公司UUID" json:"company_uuid"`
	CompanyName     string `gorm:"column:company_name;type:varchar(255);default:'';comment:所属公司名称" json:"company_name"`
	HeadquarterUuid uint64 `gorm:"column:headquarter_uuid;type:bigint;default:0;comment:总部UUID" json:"headquarter_uuid"`
	OrderNo         string `gorm:"column:order_no;type:varchar(255);default:'';comment:单据编号TR+12位数字" json:"order_no"`
	ErpOrderNo      string `gorm:"column:erp_order_no;type:varchar(255);default:'';comment:ERP调拨单号（销售单号）" json:"erp_order_no"`

	// 类型和方向
	TransferType int `gorm:"column:transfer_type;type:int(4);default:1;comment:调拨类型：1-调入 2-调出" json:"transfer_type"`

	// 门店信息
	SenderCompanyUuid   uint64 `gorm:"column:sender_company_uuid;type:bigint;default:0;comment:发货门店UUID" json:"sender_company_uuid"`
	SenderCompanyName   string `gorm:"column:sender_company_name;type:varchar(255);default:'';comment:发货门店名称" json:"sender_company_name"`
	ReceiverCompanyUuid uint64 `gorm:"column:receiver_company_uuid;type:bigint;default:0;comment:收货门店UUID" json:"receiver_company_uuid"`
	ReceiverCompanyName string `gorm:"column:receiver_company_name;type:varchar(255);default:'';comment:收货门店名称" json:"receiver_company_name"`

	// 仓库信息
	OutWarehouseErpCode string `gorm:"column:out_warehouse_erp_code;type:varchar(255);default:'';comment:出库仓库ERP编码" json:"out_warehouse_erp_code"`
	OutWarehouseName    string `gorm:"column:out_warehouse_name;type:text;comment:出库仓库名称" json:"out_warehouse_name"`
	InWarehouseErpCode  string `gorm:"column:in_warehouse_erp_code;type:varchar(255);default:'';comment:入库仓库ERP编码" json:"in_warehouse_erp_code"`
	InWarehouseName     string `gorm:"column:in_warehouse_name;type:text;comment:入库仓库名称" json:"in_warehouse_name"`

	// 时间记录
	OrderTime  int64 `gorm:"column:order_time;type:bigint;default:0;comment:单据日期（提交时间戳）" json:"order_time"`
	SubmitTime int64 `gorm:"column:submit_time;type:bigint;default:0;comment:提交时间" json:"submit_time"`

	// 状态
	Status int `gorm:"column:status;type:int(4);default:0;comment:状态：0-待提交 1-待审核 2-已驳回 3-待收货 4-已完成" json:"status"`

	// 创建人信息
	CreatorUuid uint64 `gorm:"column:creator_uuid;type:bigint;default:0;comment:创建人UUID" json:"creator_uuid"`
	CreatorName string `gorm:"column:creator_name;type:varchar(100);default:'';comment:创建人姓名" json:"creator_name"`

	// 下一个审批门店信息
	NextApprovalCompanyUuid uint64 `gorm:"column:next_approval_company_uuid;type:bigint;default:0;comment:下一个审批门店UUID" json:"next_approval_company_uuid"`
	NextApprovalCompanyName string `gorm:"column:next_approval_company_name;type:varchar(255);default:'';comment:下一个审批门店名称" json:"next_approval_company_name"`

	// 备注
	Remark string `gorm:"column:remark;type:text;comment:备注" json:"remark"`

	// 物品统计
	ItemCount int `gorm:"column:item_count;type:int;default:0;comment:物品种类数量" json:"item_count"`

	// ERP响应数据
	ErpResp             string `gorm:"column:erp_resp;type:text;comment:ERP响应数据" json:"erp_resp"`
	ReceiptOrderErpCode string `gorm:"column:receipt_order_erp_code;type:varchar(255);default:'';comment:收货单ERP编码" json:"receipt_order_erp_code"`
	ReceiptOrderErpResp string `gorm:"column:receipt_order_erp_resp;type:text;comment:收货单ERP响应数据" json:"receipt_order_erp_resp"`

	// 关联模型
	Items     []*TransferOrderItem     `gorm:"foreignKey:TransferOrderUuid;references:Uuid" json:"items,omitempty"`
	Approvals []*TransferOrderApproval `gorm:"foreignKey:TransferOrderUuid;references:Uuid" json:"approvals,omitempty"`
	Logs      []*TransferOrderLog      `gorm:"foreignKey:TransferOrderUuid;references:Uuid" json:"logs,omitempty"`
}

// TableName 指定表名
func (TransferOrder) TableName() string {
	return "ttpos_transfer_order"
}

// SetNil 设置为空
func (to *TransferOrder) SetNil() {
	to.Items = nil
	to.Approvals = nil
	to.Logs = nil
}

// getErpResp 获取ERP响应数据
func (to *TransferOrder) GetErpResp() *material_transfer.MaterialTransferResp {
	erpResp := &material_transfer.MaterialTransferResp{}
	if err := json.Unmarshal([]byte(to.ErpResp), erpResp); err != nil {
		return nil
	}
	return erpResp
}

// getErpResp 获取ERP响应数据
func (to *TransferOrder) GetErpOrderNos() []string {
	var erpOrderNos []string
	erpResp := to.GetErpResp()
	if erpResp == nil {
		return erpOrderNos
	}
	// 使用map去重
	uniqueMap := make(map[string]bool)
	soNos := []string{
		erpResp.FromReceipt.SoNo,
		erpResp.ToReceipt.SoNo,
		erpResp.AuditReceipt.SoNo,
	}
	for _, soNo := range soNos {
		// 过滤空字符串并去重
		if soNo != "" && !uniqueMap[soNo] {
			uniqueMap[soNo] = true
			erpOrderNos = append(erpOrderNos, soNo)
		}
	}
	return erpOrderNos
}

// 获取指定的审批流程
func (to *TransferOrder) GetApprovalByApprovalType(approvalType string) *TransferOrderApproval {
	for _, approval := range to.Approvals {
		if approval.ApprovalType == approvalType {
			return approval
		}
	}
	return nil
}

// 获取下一个审批流程（按 Sequence 顺序，在已审批的基础上找下一个）
func (to *TransferOrder) GetNextApproval() *TransferOrderApproval {
	if len(to.Approvals) == 0 {
		return nil
	}
	// 查找 Sequence 最小的待审批的、必须审批的审批流程
	var minSequence int = 99999
	var nextApproval *TransferOrderApproval
	for i := range to.Approvals {
		approval := to.Approvals[i]
		// 条件：状态为待审批、必须审批、Sequence 最小的
		if approval.Status == constant.TransferApprovalPending && approval.IsRequired == 1 &&
			approval.Sequence < minSequence {
			minSequence = approval.Sequence
			nextApproval = to.Approvals[i]
		}
	}
	return nextApproval
}

// TransferOrderItem 调拨单明细表 ttpos_transfer_order_item
type TransferOrderItem struct {
	BaseModel
	TransferOrderUuid uint64 `gorm:"column:transfer_order_uuid;type:bigint;default:0;comment:调拨单UUID" json:"transfer_order_uuid"`
	CompanyUuid       uint64 `gorm:"column:company_uuid;type:bigint;default:0;comment:所属公司UUID" json:"company_uuid"`
	HeadquarterUuid   uint64 `gorm:"column:headquarter_uuid;type:bigint;default:0;comment:总部UUID" json:"headquarter_uuid"`

	// 物品信息
	MaterialUuid         uint64 `gorm:"column:material_uuid;type:bigint;default:0;comment:物品UUID" json:"material_uuid"`
	MaterialCode         string `gorm:"column:material_code;type:varchar(255);default:'';comment:物品编码" json:"material_code"`
	MaterialName         string `gorm:"column:material_name;type:text;comment:物品名称JSON" json:"material_name"`
	MaterialInternalCode string `gorm:"column:material_internal_code;type:varchar(255);default:'';comment:物品内部编码" json:"material_internal_code"`

	// 价格
	Valuation float64 `gorm:"column:valuation;type:decimal(20,8);default:0.00000000;comment:估值单价（基准单位）" json:"valuation"`

	// 关联模型
	Material *Material                `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
	Units    []*TransferOrderItemUnit `gorm:"foreignKey:ItemUuid;references:Uuid" json:"units,omitempty"`
}

// TableName 指定表名
func (TransferOrderItem) TableName() string {
	return "ttpos_transfer_order_item"
}

// SetNil 设置为空
func (toi *TransferOrderItem) SetNil() {
	toi.Material = nil
	toi.Units = nil
}

// GetRelatedMaterialList 获取关联材料列表
func (item TransferOrderItem) GetUnitsTotalConversionRateNum() float64 {
	actualNum := decimal.NewFromFloat(0)
	if len(item.Units) > 0 {
		for _, unit := range item.Units {
			actualNum = actualNum.Add(decimal.NewFromFloat(unit.Num).Mul(decimal.NewFromFloat(unit.UnitConversionRate)))
		}
	}
	return actualNum.InexactFloat64()
}

// TransferOrderItemUnit 调拨单明细单位表 ttpos_transfer_order_item_unit
type TransferOrderItemUnit struct {
	BaseModel
	ItemUuid          uint64 `gorm:"column:item_uuid;type:bigint;default:0;comment:调拨单明细UUID" json:"item_uuid"`
	TransferOrderUuid uint64 `gorm:"column:transfer_order_uuid;type:bigint;default:0;comment:调拨单UUID" json:"transfer_order_uuid"`

	// 单位信息
	UnitUuid           uint64  `gorm:"column:unit_uuid;type:bigint;default:0;comment:单位UUID" json:"unit_uuid"`
	UnitName           string  `gorm:"column:unit_name;type:text;comment:单位名称JSON" json:"unit_name"`
	UnitConversionRate float64 `gorm:"column:unit_conversion_rate;type:decimal(12,4);default:1.0000;comment:单位转换率" json:"unit_conversion_rate"`

	// 数量
	Num float64 `gorm:"column:num;type:decimal(22,4);default:0.0000;comment:调拨数量" json:"num"`

	// ERP相关
	ErpnextUom string `gorm:"column:erpnext_uom;type:varchar(255);default:'';comment:ERP单位" json:"erpnext_uom"`
}

// TableName 指定表名
func (TransferOrderItemUnit) TableName() string {
	return "ttpos_transfer_order_item_unit"
}

// TransferOrderApproval 调拨单审批流程表 ttpos_transfer_order_approval
type TransferOrderApproval struct {
	BaseModel
	TransferOrderUuid uint64 `gorm:"column:transfer_order_uuid;type:bigint;default:0;comment:调拨单UUID" json:"transfer_order_uuid"`
	CompanyUuid       uint64 `gorm:"column:company_uuid;type:bigint;default:0;comment:所属公司UUID" json:"company_uuid"`
	HeadquarterUuid   uint64 `gorm:"column:headquarter_uuid;type:bigint;default:0;comment:总部UUID" json:"headquarter_uuid"`

	// 审批信息
	ApprovalType        string `gorm:"column:approval_type;type:varchar(50);default:'';comment:审批类型：sender/sender_parent/receiver/receiver_parent" json:"approval_type"`
	ApprovalCompanyUuid uint64 `gorm:"column:approval_company_uuid;type:bigint;default:0;comment:审批门店UUID" json:"approval_company_uuid"`
	ApprovalCompanyName string `gorm:"column:approval_company_name;type:varchar(255);default:'';comment:审批门店名称" json:"approval_company_name"`
	Sequence            int    `gorm:"column:sequence;type:int;default:0;comment:审批顺序，从1开始" json:"sequence"`

	// 审批状态
	Status       int    `gorm:"column:status;type:int(4);default:0;comment:审批状态：0-待审批 1-已通过 2-已驳回 3-已跳过" json:"status"`
	ApproverUuid uint64 `gorm:"column:approver_uuid;type:bigint;default:0;comment:审批人UUID" json:"approver_uuid"`
	ApproverName string `gorm:"column:approver_name;type:varchar(100);default:'';comment:审批人姓名" json:"approver_name"`
	ApproveTime  int64  `gorm:"column:approve_time;type:int;default:0;comment:审批时间" json:"approve_time"`
	RejectReason string `gorm:"column:reject_reason;type:text;comment:驳回原因" json:"reject_reason"`

	// 配置
	IsRequired            int    `gorm:"column:is_required;type:int(4);default:1;comment:是否必须审批：0-否 1-是" json:"is_required"`
	Remark                string `gorm:"column:remark;type:text;comment:备注" json:"remark"`
	IsViaCompanyWarehouse int    `gorm:"column:is_via_company_warehouse;type:int(4);default:0;comment:是否通过公司仓库：0-否 1-是" json:"is_via_company_warehouse"`
	ErpnextCompanyAbbr    string `gorm:"column:erpnext_company_abbr;type:varchar(255);default:'';comment:ERP公司简称" json:"erpnext_company_abbr"`
}

// TableName 指定表名
func (TransferOrderApproval) TableName() string {
	return "ttpos_transfer_order_approval"
}

// IsViaCompanyWarehouseBool 获取是否通过公司仓库
func (toa *TransferOrderApproval) IsViaCompanyWarehouseBool() bool {
	return toa.IsViaCompanyWarehouse == 1
}

// TransferOrderLog 调拨单操作日志表 ttpos_transfer_order_log
type TransferOrderLog struct {
	BaseModel
	TransferOrderUuid uint64 `gorm:"column:transfer_order_uuid;type:bigint;default:0;comment:调拨单UUID" json:"transfer_order_uuid"`
	CompanyUuid       uint64 `gorm:"column:company_uuid;type:bigint;default:0;comment:所属公司UUID" json:"company_uuid"`

	// 操作信息
	Action     string `gorm:"column:action;type:varchar(50);default:'';comment:操作动作：create/submit/approve/reject/receive" json:"action"`
	ActionDesc string `gorm:"column:action_desc;type:varchar(255);default:'';comment:操作描述" json:"action_desc"`
	OldStatus  int    `gorm:"column:old_status;type:int(4);default:0;comment:操作前状态" json:"old_status"`
	NewStatus  int    `gorm:"column:new_status;type:int(4);default:0;comment:操作后状态" json:"new_status"`

	// 操作人
	OperatorUuid uint64 `gorm:"column:operator_uuid;type:bigint;default:0;comment:操作人UUID" json:"operator_uuid"`
	OperatorName string `gorm:"column:operator_name;type:varchar(100);default:'';comment:操作人姓名" json:"operator_name"`
	OperatorRole string `gorm:"column:operator_role;type:varchar(50);default:'';comment:操作人角色：sender/sender_parent/receiver_parent/receiver" json:"operator_role"`

	// 详细内容
	Content string `gorm:"column:content;type:text;comment:操作内容详情JSON" json:"content"`
	Remark  string `gorm:"column:remark;type:text;comment:备注" json:"remark"`
}

// TableName 指定表名
func (TransferOrderLog) TableName() string {
	return "ttpos_transfer_order_log"
}

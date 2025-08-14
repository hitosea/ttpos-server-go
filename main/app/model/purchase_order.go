package model

import (
	"time"
	"ttpos-server-go/app/constant"
)

// PurchaseOrder 采购申请表 ttpos_purchase_order
type PurchaseOrder struct {
	BaseModel
	OrderNo           string  `gorm:"column:order_no;type:varchar(255);not null;default:'';comment:单号" json:"order_no"`
	OrderType         int     `gorm:"column:order_type;type:int(10);not null;default:0;comment:申请类型, 0-仓库调拨" json:"order_type"`
	Status            int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态, 0-待提交 1-待审核 2-已通过 3-已驳回 4-部分收货 5-全部收货" json:"status"`
	Num               float64 `gorm:"column:num;type:decimal(14,4);not null;default:0.0000;comment:物资数量，每种物品算一个" json:"num"`
	OrderTime         int64   `gorm:"column:order_time;type:int(10) unsigned;not null;default:0;comment:单据日期，采购单提交的时间（时间戳）" json:"order_time"`
	ApplicantUuid     uint64  `gorm:"column:applicant_uuid;type:bigint(20) unsigned;not null;default:0;comment:申请人ID" json:"applicant_uuid"`
	ApplicantName     string  `gorm:"column:applicant_name;type:varchar(255);not null;default:'';comment:申请人姓名" json:"applicant_name"`
	ApproverUuid      uint64  `gorm:"column:approver_uuid;type:bigint(20) unsigned;not null;default:0;comment:审批人ID" json:"approver_uuid"`
	ApproverName      string  `gorm:"column:approver_name;type:varchar(255);not null;default:'';comment:审批人姓名" json:"approver_name"`
	ExpectArrivalTime int64   `gorm:"column:expect_arrival_time;type:int(10) unsigned;not null;default:0;comment:期望到货日期（时间戳）" json:"expect_arrival_time"`
	PassTime          int64   `gorm:"column:pass_time;type:int(10) unsigned;not null;default:0;comment:通过时间（时间戳）" json:"pass_time"`
	RejectTime        int64   `gorm:"column:reject_time;type:int(10) unsigned;not null;default:0;comment:驳回时间（时间戳）" json:"reject_time"`
	FirstReceiveTime  int64   `gorm:"column:first_receive_time;type:int(10) unsigned;not null;default:0;comment:第一次收货时间（时间戳），从\"已通过\"状态变成\"部分收货\"状态的时间" json:"first_receive_time"`
	FinalReceiveTime  int64   `gorm:"column:final_receive_time;type:int(10) unsigned;not null;default:0;comment:最终收货时间（时间戳），从\"部分收货\"状态变成\"全部收货\"状态的时间" json:"final_receive_time"`

	// 关联关系
	Items    []PurchaseOrderItem    `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"items,omitempty"`
	Logs     []PurchaseOrderLog     `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"logs,omitempty"`
	Receipts []PurchaseReceiptOrder `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"receipts,omitempty"`
}

// TableName 指定表名
func (PurchaseOrder) TableName() string {
	return "ttpos_purchase_order"
}

// GetStatusText 获取状态文本
func (po *PurchaseOrder) GetStatusText() string {
	statusMap := map[int]string{
		0: "待提交",
		1: "待审核",
		2: "已通过",
		3: "已驳回",
		4: "部分收货",
		5: "全部收货",
	}
	if text, exists := statusMap[po.Status]; exists {
		return text
	}
	return "未知状态"
}

// IsEditable 判断是否可编辑
func (po *PurchaseOrder) IsEditable() bool {
	return po.Status == constant.PurchaseOrderStatusDraft
}

// CanApprove 判断是否可审核
func (po *PurchaseOrder) CanApprove() bool {
	return po.Status == constant.PurchaseOrderStatusPending
}

// CanReceive 判断是否可收货
func (po *PurchaseOrder) CanReceive() bool {
	// 已通过和部分收货状态可以收货
	return po.Status == constant.PurchaseOrderStatusApproved || po.Status == constant.PurchaseOrderStatusPartialReceived
}

// GetOrderDate 获取订单日期
func (po *PurchaseOrder) GetOrderDate() time.Time {
	if po.OrderTime > 0 {
		return time.Unix(int64(po.OrderTime), 0)
	}
	return time.Time{}
}

// GetExpectArrivalDate 获取期望到货日期
func (po *PurchaseOrder) GetExpectArrivalDate() time.Time {
	if po.ExpectArrivalTime > 0 {
		return time.Unix(int64(po.ExpectArrivalTime), 0)
	}
	return time.Time{}
}

// PurchaseOrderItem 采购申请物品表 ttpos_purchase_order_item
type PurchaseOrderItem struct {
	BaseModel
	PurchaseOrderUuid  uint64  `gorm:"column:purchase_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:采购申请ID;index" json:"purchase_order_uuid"`
	MaterialCode       string  `gorm:"column:material_code;type:varchar(255);not null;default:'';comment:物品编码, 提交采购时记录后不再修改" json:"material_code"`
	MaterialName       string  `gorm:"column:material_name;type:text;not null;default:'';comment:物品名称JSON, 提交采购时记录后不再修改" json:"material_name"`
	MaterialUuid       uint64  `gorm:"column:material_uuid;type:bigint(20) unsigned;not null;default:0;comment:物品ID" json:"material_uuid"`
	Num                float64 `gorm:"column:num;type:decimal(14,4);not null;default:0.0000;comment:申请数量" json:"num"`
	ArrivalNum         float64 `gorm:"column:arrival_num;type:decimal(14,4);not null;default:0.0000;comment:到货数量" json:"arrival_num"`
	UnitUuid           uint64  `gorm:"column:unit_uuid;type:bigint(20) unsigned;not null;default:0;comment:单位ID-采购单位ID" json:"unit_uuid"`
	UnitName           string  `gorm:"column:unit_name;type:text;not null;default:'';comment:单位名称JSON, 提交采购时记录后不再修改" json:"unit_name"`
	UnitConversionRate float64 `gorm:"column:unit_conversion_rate;type:decimal(12,4);not null;default:1;comment:单位转换率。申请数量*转换率=基准单位申请数量" json:"unit_conversion_rate"`
	BaseUnitUuid       uint64  `gorm:"column:base_unit_uuid;type:bigint(20) unsigned;not null;default:0;comment:基准单位ID" json:"base_unit_uuid"`
	BaseUnitName       string  `gorm:"column:base_unit_name;type:text;not null;default:'';comment:基准单位名称JSON, 提交采购时记录后不再修改" json:"base_unit_name"`

	// 关联关系
	PurchaseOrder PurchaseOrder `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"purchase_order,omitempty"`
	Material      *Material     `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
}

// TableName 指定表名
func (PurchaseOrderItem) TableName() string {
	return "ttpos_purchase_order_item"
}

// GetCompletionRate 获取到货完成率
func (poi *PurchaseOrderItem) GetCompletionRate() float64 {
	if poi.Num == 0 {
		return 0
	}
	return poi.ArrivalNum / poi.Num * 100
}

// IsCompleted 判断是否到货完成
func (poi *PurchaseOrderItem) IsCompleted() bool {
	return poi.ArrivalNum >= poi.Num
}

// GetRemainingQuantity 获取剩余待收货数量
func (poi *PurchaseOrderItem) GetRemainingQuantity() float64 {
	remaining := poi.Num - poi.ArrivalNum
	if remaining < 0 {
		return 0
	}
	return remaining
}

// PurchaseOrderLog 采购订单操作日志
type PurchaseOrderLog struct {
	BaseModel
	PurchaseOrderUuid uint64 `gorm:"column:purchase_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:采购订单ID;index" json:"purchase_order_uuid"`
	OperatorUuid      uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作人ID;index" json:"operator_uuid"`
	OperatorName      string `gorm:"column:operator_name;type:varchar(100);not null;default:'';comment:操作人姓名" json:"operator_name"`
	Action            string `gorm:"column:action;type:varchar(50);not null;default:'';comment:操作动作;index" json:"action"`
	ActionDesc        string `gorm:"column:action_desc;type:varchar(255);not null;default:'';comment:操作描述" json:"action_desc"`
	OldStatus         int    `gorm:"column:old_status;type:int(10);not null;default:0;comment:操作前状态" json:"old_status"`
	NewStatus         int    `gorm:"column:new_status;type:int(10);not null;default:0;comment:操作后状态" json:"new_status"`
	Content           string `gorm:"column:content;type:text;comment:操作内容详情" json:"content"`
	Remark            string `gorm:"column:remark;type:text;comment:备注" json:"remark"`

	// 关联关系
	PurchaseOrder PurchaseOrder `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"purchase_order,omitempty"`
}

// TableName 指定表名
func (PurchaseOrderLog) TableName() string {
	return "ttpos_purchase_order_log"
}

// ReceiptOrder 收货单表 ttpos_purchase_receipt_order
type PurchaseReceiptOrder struct {
	BaseModel
	OrderNo           string  `gorm:"column:order_no;type:varchar(255);not null;default:'';comment:单号" json:"order_no"`
	Status            int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态, 0-待收货 1-已收货 2-已取消" json:"status"`
	PurchaseOrderUuid uint64  `gorm:"column:purchase_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:采购申请ID" json:"purchase_order_uuid"`
	PurchaseOrderNo   string  `gorm:"column:purchase_order_no;type:varchar(255);not null;default:'';comment:采购申请单号" json:"purchase_order_no"`
	Num               float64 `gorm:"column:num;type:decimal(14,4);not null;default:0.0000;comment:物资数量，每种物品算一个" json:"num"`
	ExpectArrivalTime int64   `gorm:"column:expect_arrival_time;type:int(10) unsigned;not null;default:0;comment:期望到货日期（时间戳），与采购申请单的期望到货日期一致" json:"expect_arrival_time"`
	ReceiveTime       int64   `gorm:"column:receive_time;type:int(10) unsigned;not null;default:0;comment:收货时间（时间戳）" json:"receive_time"`
	CancelTime        int64   `gorm:"column:cancel_time;type:int(10) unsigned;not null;default:0;comment:取消时间（时间戳）" json:"cancel_time"`

	// 关联关系
	PurchaseOrder PurchaseOrder              `gorm:"foreignKey:PurchaseOrderUuid;references:Uuid" json:"purchase_order,omitempty"`
	Items         []PurchaseReceiptOrderItem `gorm:"foreignKey:ReceiptOrderUuid;references:Uuid" json:"items,omitempty"`
}

// TableName 指定表名
func (PurchaseReceiptOrder) TableName() string {
	return "ttpos_purchase_receipt_order"
}

// GetStatusText 获取状态文本
func (ro *PurchaseReceiptOrder) GetStatusText() string {
	statusMap := map[int]string{
		0: "待收货",
		1: "已收货",
		2: "已取消",
	}
	if text, exists := statusMap[ro.Status]; exists {
		return text
	}
	return "未知状态"
}

// GetReceiveDate 获取收货日期
func (ro *PurchaseReceiptOrder) GetReceiveDate() time.Time {
	if ro.ReceiveTime > 0 {
		return time.Unix(ro.ReceiveTime, 0)
	}
	return time.Time{}
}

// PurchaseReceiptOrderItem 收货单物品表 ttpos_purchase_receipt_order_item
type PurchaseReceiptOrderItem struct {
	BaseModel
	ReceiptOrderUuid      uint64  `gorm:"column:receipt_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:收货单ID;index" json:"receipt_order_uuid"`
	PurchaseOrderItemUuid uint64  `gorm:"column:purchase_order_item_uuid;type:bigint(20) unsigned;not null;default:0;comment:采购申请物品ID;index" json:"purchase_order_item_uuid"`
	MaterialCode          string  `gorm:"column:material_code;type:varchar(255);not null;default:'';comment:物品编码, 提交采购时记录后不再修改" json:"material_code"`
	MaterialName          string  `gorm:"column:material_name;type:text;not null;default:'';comment:物品名称JSON, 提交采购时记录后不再修改" json:"material_name"`
	MaterialUuid          uint64  `gorm:"column:material_uuid;type:bigint(20) unsigned;not null;default:0;comment:物品ID" json:"material_uuid"`
	Num                   float64 `gorm:"column:num;type:decimal(14,4);not null;default:0.0000;comment:收货数量" json:"num"`
	UnitUuid              uint64  `gorm:"column:unit_uuid;type:bigint(20) unsigned;not null;default:0;comment:单位ID" json:"unit_uuid"`
	UnitName              string  `gorm:"column:unit_name;type:varchar(255);not null;default:'';comment:单位名称, 提交采购时记录后不再修改" json:"unit_name"`
	UnitConversionRate    float64 `gorm:"column:unit_conversion_rate;type:decimal(12,4);not null;default:1;comment:单位转换率。收货数量*转换率=基准单位收货数量" json:"unit_conversion_rate"`
	BaseUnitUuid          uint64  `gorm:"column:base_unit_uuid;type:bigint(20) unsigned;not null;default:0;comment:基准单位ID" json:"base_unit_uuid"`
	BaseUnitName          string  `gorm:"column:base_unit_name;type:varchar(255);not null;default:'';comment:基准单位名称, 确认收货时记录后不再修改" json:"base_unit_name"`

	// 关联关系
	PurchaseReceiptOrder PurchaseReceiptOrder `gorm:"foreignKey:ReceiptOrderUuid;references:Uuid" json:"purchase_receipt_order,omitempty"`
	PurchaseOrderItem    PurchaseOrderItem    `gorm:"foreignKey:PurchaseOrderItemUuid;references:Uuid" json:"purchase_order_item,omitempty"`
	Material             *Material            `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
}

// TableName 指定表名
func (PurchaseReceiptOrderItem) TableName() string {
	return "ttpos_purchase_receipt_order_item"
}

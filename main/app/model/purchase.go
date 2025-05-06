package model

// 采购订单
type PurchaseForm struct {
	BaseModel
	FormNo        string  `gorm:"column:form_no;type:varchar(255);not null;default:'';comment:编号"`
	Name          string  `gorm:"column:name;type:varchar(255);not null;default:'';comment:名称"`
	ApplicantUuid uint64  `gorm:"column:applicant_uuid;type:bigint(20);unsigned;not null;default:0;comment:申请人id"`
	Remark        string  `gorm:"column:remark;type:varchar(255);default:null;comment:备注"`
	Num           float64 `gorm:"column:num;type:decimal(12,2);not null;default:0;comment:总数量"`
	Amount        float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:总金额"`
	Status        int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库"`
	ArrivalTime   int64   `gorm:"column:arrival_time;type:int(10);not null;default:0;comment:到达时间(时间戳)"`
}

// 采购订单明细
type PurchaseFormItem struct {
	BaseModel
	PurchaseFormUuid uint64  `gorm:"column:purchase_form_uuid;type:bigint(20);unsigned;not null;default:0;comment:采购订单uuid"`
	MaterialType     int     `gorm:"column:material_type;type:int(10);not null;default:0;comment:物料类型,0-商品 1-原料"`
	MaterialUuid     uint64  `gorm:"column:material_uuid;type:bigint(20);unsigned;not null;default:0;comment:物料ID"`
	EstimateNum      int     `gorm:"column:estimate_num;type:int(10);not null;default:0;comment:预计数量"`
	EstimatePrice    float64 `gorm:"column:estimate_price;type:decimal(12,2);not null;default:0;comment:预计单价"`
	EstimateAmount   float64 `gorm:"column:estimate_amount;type:decimal(12,2);not null;default:0;comment:预计金额"`
	Num              float64 `gorm:"column:num;type:decimal(20,4);not null;default:0;comment:数量"`
	Price            float64 `gorm:"column:price;type:decimal(12,2);not null;default:0;comment:实际单价"`
	Amount           float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0;comment:实际金额"`
}

// 采购订单明细日志
type PurchaseFormLog struct {
	BaseModel
	PurchaseFormUuid uint64 `gorm:"column:purchase_form_uuid;type:bigint(20);unsigned;not null;default:0;comment:采购订单uuid"`
	OperatorUuid     uint64 `gorm:"column:operator_uuid;type:bigint(20);unsigned;not null;default:0;comment:操作人uuid"`
	Username         string `gorm:"column:username;type:varchar(255);default:'';comment:操作人"`
	Status           int    `gorm:"column:status;type:int(10);not null;default:0;comment:操作状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库"`
	Operation        string `gorm:"column:operation;type:varchar(255);default:'';comment:操作动作"`
	Remark           string `gorm:"column:remark;type:varchar(255);default:'';comment:备注"`
}

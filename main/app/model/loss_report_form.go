package model

// 损耗记录
type LossReportForm struct {
	BaseModel
	FormNo         string  `gorm:"column:form_no;type:varchar(255);not null;default:'';comment:编号"`
	Scene          int     `gorm:"column:scene;type:int(10);not null;default:0;comment:报损类型,0-loss损耗 1-lost丢失"`
	Num            float64 `gorm:"column:num;type:decimal(20,4);not null;default:0;comment:数量"`
	Remark         string  `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注"`
	ProductBomUuid uint64  `gorm:"column:product_bom_uuid;type:bigint(20);unsigned;not null;default:0;comment:商品规格id"`
	MaterialUuid   uint64  `gorm:"column:material_uuid;type:bigint(20);unsigned;not null;default:0;comment:物料id"`
	ApplicantUuid  uint64  `gorm:"column:applicant_uuid;type:bigint(20);unsigned;not null;default:0;comment:申请人id"`
	RejectReason   string  `gorm:"column:reject_reason;type:varchar(255);not null;default:'';comment:驳回原因"`
	Status         int     `gorm:"column:status;type:int(10);not null;default:0;comment:状态,0-pending待审核 1-approved已通过 2-rejected已驳回"`
	OperatorUuid   uint64  `gorm:"column:operator_uuid;type:bigint(20);unsigned;not null;default:0;comment:操作人id"`
	ApprovedTime   int64   `gorm:"column:approved_time;type:int(10);not null;default:0;comment:审核通过时间"`
	RevokeTime     int64   `gorm:"column:revoke_time;type:int(10);not null;default:0;comment:撤销时间"`
}

package model

// FullReductionActivityRule 满减活动规则表
type FullReductionActivityRule struct {
	BaseModel
	FullReductionActivityUuid uint64  `gorm:"column:full_reduction_activity_uuid;type:bigint(20) unsigned;default:0;comment:'活动UUID';index" json:"full_reduction_activity_uuid"`
	Threshold                 float64 `gorm:"column:threshold;type:decimal(22,4);default:0.0000;comment:'阈值（满减条件）'" json:"threshold"`
	ReductionAmount           float64 `gorm:"column:reduction_amount;type:decimal(22,4);default:0.0000;comment:'扣减值（满减金额）'" json:"reduction_amount"`
}

func (*FullReductionActivityRule) TableName() string {
	return "ttpos_full_reduction_activity_rule"
}

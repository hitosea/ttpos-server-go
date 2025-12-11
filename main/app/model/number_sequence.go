package model

// NumberSequence 通用编号序列表（saas库）
// 用于管理各类编号的按日递增序列（发票、订单、收据等）
type NumberSequence struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CompanyUuid uint64 `gorm:"column:company_uuid;not null;index:idx_company_uuid_type_date,unique" json:"company_uuid"`                      // 商家UUID
	Type        string `gorm:"column:type;type:varchar(32);not null;index:idx_company_uuid_type_date,unique;index:idx_type_date" json:"type"` // 编号类型
	Date        string `gorm:"column:date;type:date;not null;index:idx_company_uuid_type_date,unique;index:idx_type_date" json:"date"`        // 日期 YYYY-MM-DD
	Sequence    uint64 `gorm:"column:sequence;not null;default:0" json:"sequence"`                                                            // 当日序列号
	CreateTime  int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`                                                      // 创建时间
	UpdateTime  int64  `gorm:"column:update_time;not null;default:0" json:"update_time"`                                                      // 更新时间
}

// TableName 指定表名
func (NumberSequence) TableName() string {
	return "ttpos_number_sequence"
}

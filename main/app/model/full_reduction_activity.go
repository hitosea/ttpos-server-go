package model

// FullReductionActivity 满减活动表
type FullReductionActivity struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(1000);default:'';comment:'活动名称（JSON格式）'" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:'多语言名称UUID'" json:"multi_language_name_uuid"`
	StartDate             int64  `gorm:"column:start_date;type:int(10);default:0;comment:'活动开始日期（时间戳，当天00:00:00）'" json:"start_date"`
	EndDate               int64  `gorm:"column:end_date;type:int(10);default:0;comment:'活动结束日期（时间戳，当天23:59:59）'" json:"end_date"`
	StartTime             string `gorm:"column:start_time;type:varchar(10);default:'';comment:'适用时间开始（格式：HH:mm）'" json:"start_time"`
	EndTime               string `gorm:"column:end_time;type:varchar(10);default:'';comment:'适用时间结束（格式：HH:mm）'" json:"end_time"`
	IsAllDay              int    `gorm:"column:is_all_day;type:tinyint(1);default:1;comment:'是否全天（1=全天，0=特定时段）'" json:"is_all_day"`
	ReductionType         int    `gorm:"column:reduction_type;type:tinyint(1);default:0;comment:'满减方式（0=阶梯满减，1=循环满减）'" json:"reduction_type"`
	IsDisabled            int    `gorm:"column:is_disabled;type:tinyint(1);default:0;comment:'是否失效（1=失效，0=未失效）'" json:"is_disabled"`

	// 关联
	MultiLanguageName MultiLanguageName                `gorm:"foreignKey:multi_language_name_uuid;references:uuid" json:"multi_language_name,omitempty"`
	Rules             []FullReductionActivityRule      `gorm:"foreignKey:full_reduction_activity_uuid;references:uuid" json:"rules,omitempty"`
}

func (*FullReductionActivity) TableName() string {
	return "ttpos_full_reduction_activity"
}

// GetStatus 获取活动状态（进行中/未开始/已结束）
func (m *FullReductionActivity) GetStatus(now int64, timezone string) string {
	// 如果已失效，返回已结束
	if m.IsDisabled == 1 {
		return "ended"
	}

	// 根据当前时间和活动日期判断状态
	if now < m.StartDate {
		return "not_started"
	}
	if now > m.EndDate {
		return "ended"
	}
	return "ongoing"
}


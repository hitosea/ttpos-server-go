package model

// BusinessStatusPeriod 测试营业时段记录 ttpos_business_status_period
// 仅记录测试营业时段，end_time=0 表示测试进行中
type BusinessStatusPeriod struct {
	BaseModel
	StartTime int64 `gorm:"column:start_time;type:int(10);default:0;comment:测试营业开始时间(时间戳);NOT NULL" json:"start_time"`
	EndTime   int64 `gorm:"column:end_time;type:int(10);default:0;comment:测试营业结束时间(时间戳),0表示进行中;NOT NULL" json:"end_time"`
}

// TableName 指定表名
func (BusinessStatusPeriod) TableName() string {
	return "ttpos_business_status_period"
}

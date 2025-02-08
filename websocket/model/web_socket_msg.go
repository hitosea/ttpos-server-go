package model

// Setting 设置表
type WebSocketMsg struct {
	ID           uint   `gorm:"default:0;comment:消息ID;NOT NULL" json:"id"`
	CompanyUuid  uint   `gorm:"default:0;comment:消息ID;NOT NULL" json:"company_uuid"`
	Uid          string `gorm:"default:'';comment:ID;NOT NULL" json:"uid"`
	Msg          string `gorm:"default:'';comment:消息内容;NOT NULL" json:"msg"`
	Type         string `gorm:"default:'';comment:消息类型;NOT NULL" json:"type"`
	Status       int    `gorm:"default:0;comment:状态 0-未读 1-已读;NOT NULL" json:"status"`
	SourceClient string `gorm:"default:'';comment:状态 0-未读 1-已读;NOT NULL" json:"source_client"`
	IsOffline    int    `gorm:"default:0;comment:是否离线消息 0-否 1-是;NOT NULL" json:"is_offline"`
	CreateTime   uint   `gorm:"default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime   uint   `gorm:"default:0;comment:更新时间;NOT NULL" json:"update_time"`
}

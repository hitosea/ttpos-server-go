package model

// 客户呼叫记录表
type CustomerCall struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint   `gorm:"default:0;comment:'客户呼叫记录ID'"`
	DeskUUID   uint   `gorm:"default:0;comment:'桌台ID'"`
	DeskNo     string `gorm:"default:'';comment:'桌台编号'"`
	Status     uint   `gorm:"default:0;comment:'状态,0-unhandled未处理 1-handled已处理'"`
	IsSend     uint   `gorm:"default:0;comment:'消息发送状态 0-否 1-是'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间'"`
}

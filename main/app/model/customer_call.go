package model

// 客户呼叫记录表
type CustomerCall struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"default:0;column:uuid;comment:'客户呼叫记录ID'"`
	DeskUUID   uint64 `gorm:"default:0;column:desk_uuid;comment:'桌台ID'"`
	DeskNo     string `gorm:"default:'';column:desk_no;comment:'桌台编号'"`
	Status     uint   `gorm:"default:0;column:status;comment:'状态,0-unhandled未处理 1-handled已处理'"`
	IsSend     uint   `gorm:"default:0;column:is_send;comment:'消息发送状态 0-否 1-是'"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;comment:'创建时间'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;comment:'更新时间'"`
	DeleteTime int64  `gorm:"default:0;column:delete_time;comment:'删除时间'"`
}

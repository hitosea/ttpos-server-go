package model

type QrcodeOrder struct {
	ID         uint    `gorm:"column:id;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64  `gorm:"column:uuid;default:0;comment:'扫码订单ID'"`
	Name       string  `gorm:"column:name;default:'';comment:'扫码订单名称'"`
	DeskUuid   uint64  `gorm:"column:desk_uuid;default:0;comment:'桌台ID'"`
	DeskNo     string  `gorm:"column:desk_no;default:'';comment:'桌台编号'"`
	Amount     float64 `gorm:"column:amount;type:decimal(12,2);default:0;comment:'订单金额'"`
	Status     uint    `gorm:"column:status;default:0;comment:'状态, 0-未接单 1-已接单 2-已拒单'"`
	HandleTime int64   `gorm:"column:handle_time;default:0;comment:'接单时间(时间戳)'"`
	CreateTime int64   `gorm:"autoCreateTime;column:create_time;comment:'创建时间(时间戳)'"`
	UpdateTime int64   `gorm:"autoUpdateTime;column:update_time;comment:'更新时间(时间戳)'"`
	DeleteTime int64   `gorm:"column:delete_time;default:0;comment:'删除时间(时间戳)'"`
}

package model

// CustomerCall 客户呼叫记录表 ttpos_customer_call
type CustomerCall struct {
	BaseModel
	DeskUUID uint64 `gorm:"default:0;column:desk_uuid;comment:'桌台ID'"`
	DeskNo   string `gorm:"default:'';column:desk_no;comment:'桌台编号,不随后台改变'"`
	Status   uint   `gorm:"default:0;column:status;comment:'状态,0-unhandled未处理 1-handled已处理'"`
	IsSend   uint   `gorm:"default:0;column:is_send;comment:'消息发送状态 0-否 1-是'"`
}

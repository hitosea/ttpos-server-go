package model

// PrinterCustomize 打印机定制结构体 `ttpos_printer_customize`
type PrinterCustomize struct {
	ID         uint64 `gorm:"column:id;primary_key;auto_increment" json:"id" form:"id"`                                 // 自增ID
	Uuid       uint64 `gorm:"column:uuid;unique;not null;default:0" json:"uuid" form:"uuid"`                            // ID
	Name       string `gorm:"column:name;type:varchar(255);default:''" json:"name" form:"name"`                         // 名称
	IsAdv      int    `gorm:"column:is_adv;type:int(11);default:0" json:"is_adv" form:"is_adv"`                         // 是否高级
	IsUse      int    `gorm:"column:is_use;type:int(11);default:0" json:"is_use" form:"is_use"`                         // 是否使用
	Data       string `gorm:"column:data;type:longtext" json:"data" form:"data"`                                        // 定制数据
	CreateTime int64  `gorm:"column:create_time;type:int(11);not null;default:0" json:"create_time" form:"create_time"` // 创建时间(时间戳)
	UpdateTime int64  `gorm:"column:update_time;type:int(11);not null;default:0" json:"update_time" form:"update_time"` // 更新时间(时间戳)
	DeleteTime int64  `gorm:"column:delete_time;type:int(10);not null;default:0" json:"delete_time" form:"delete_time"` // 删除时间(时间戳)
}

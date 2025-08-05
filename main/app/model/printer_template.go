package model

// PrinterTemplate 打印机模板结构体 `ttpos_printer_template`
type PrinterTemplate struct {
	ID         uint64 `gorm:"column:id;primary_key;auto_increment" json:"id" form:"id"`                                 // 自增ID
	Uuid       uint64 `gorm:"column:uuid;unique;not null;default:0" json:"uuid" form:"uuid"`                            // 打印机模板ID
	Name       string `gorm:"column:name;type:varchar(255);default:''" json:"name" form:"name"`                         // 打印名称
	Template   int    `gorm:"column:template;type:int(11);default:1" json:"template" form:"template"`                   // 模板选择
	IsShowSku  int    `gorm:"column:is_show_sku;type:int(1);default:1" json:"is_show_sku" form:"is_show_sku"`           // 是否显示SKU：0=不显示，1=显示
	CreateTime int64  `gorm:"column:create_time;type:int(11);not null;default:0" json:"create_time" form:"create_time"` // 创建时间(时间戳)
	UpdateTime int64  `gorm:"column:update_time;type:int(11);not null;default:0" json:"update_time" form:"update_time"` // 更新时间(时间戳)
	DeleteTime int64  `gorm:"column:delete_time;type:int(10);not null;default:0" json:"delete_time" form:"delete_time"` // 删除时间(时间戳)
}

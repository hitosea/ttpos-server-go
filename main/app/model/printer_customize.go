package model

// PrinterCustomize 打印机定制结构体 `ttpos_printer_customize`
type PrinterCustomize struct {
	BaseModel
	Name       string `gorm:"column:name;type:varchar(255);default:''" json:"name" form:"name"`                            // 名称
	IsAdv      int    `gorm:"column:is_adv;type:int(11);default:0" json:"is_adv" form:"is_adv"`                            // 是否高级
	TemplateId uint64 `gorm:"column:template_id;type:bigint(20);not null;default:0" json:"template_id" form:"template_id"` // 模板ID
	IsUse      int    `gorm:"column:is_use;type:int(11);default:0" json:"is_use" form:"is_use"`                            // 是否使用
	Data       string `gorm:"column:data;type:longtext" json:"data" form:"data"`                                           // 定制数据
}

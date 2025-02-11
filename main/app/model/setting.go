package model

// Setting 设置表 ttpos_setting
type Setting struct {
	Key         string `gorm:"column:key;type:varchar(30);comment:设置项标示;NOT NULL" json:"key"`
	Description string `gorm:"column:description;type:varchar(255);comment:设置项描述;NOT NULL" json:"description"`
	Values      string `gorm:"column:values;type:mediumtext;comment:设置内容（json格式）;NOT NULL" json:"values"`
	CreateTime  uint   `gorm:"column:create_time;type:int(10) unsigned;default:0;comment:创建时间;NOT NULL" json:"create_time"`
	UpdateTime  uint   `gorm:"column:update_time;type:int(10) unsigned;default:0;comment:更新时间;NOT NULL" json:"update_time"`
	DeleteTime  uint   `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间;NOT NULL" json:"delete_time"`
}

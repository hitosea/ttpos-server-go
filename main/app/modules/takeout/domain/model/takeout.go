package model

import "ttpos-server-go/app/model"

// Takeout 外卖平台状态表 ttpos_takeout
type Takeout struct {
	model.BaseModel
	Uuid        uint64      `gorm:"column:uuid;type:bigint(20);default:0;comment:唯一标识;NOT NULL" json:"uuid"`
	Platform    string      `gorm:"column:platform;type:varchar(50);comment:外卖平台(grab/lineman等);NOT NULL" json:"platform"`
	Enabled     bool        `gorm:"column:enabled;type:tinyint(1);default:1;comment:是否开启(1:开启 0:关闭);NOT NULL" json:"enabled"`
	IsBound     bool        `gorm:"column:is_bound;type:tinyint(1);default:0;comment:是否已经绑定平台(1:已绑定 0:未绑定);NOT NULL" json:"is_bound"`
	Skip        bool        `gorm:"column:skip;type:tinyint(1);default:0;comment:是否跳过绑定(1:跳过 0:不跳过);NOT NULL" json:"skip"`
	BindingLink string      `gorm:"column:binding_link;type:varchar(500);default:'';comment:平台绑定链接(缓存用);NOT NULL" json:"binding_link"`
	Menu        interface{} `gorm:"column:menu;type:json;comment:平台菜单数据(JSON格式)" json:"menu"`
}

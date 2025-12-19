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
	TtposMenu   interface{} `gorm:"column:ttpos_menu;type:json;comment:TTPOS导出的菜单数据(JSON格式)" json:"ttpos_menu"`

	// 导入进度相关字段
	ImportStatus int8 `gorm:"column:import_status;type:tinyint(3);default:0;comment:导入状态(0-未导入 1-导入中 2-导入成功 3-导入失败);NOT NULL;index:idx_import_status" json:"import_status"`
}

// ImportStatus 导入状态常量
const (
	ImportStatusNone       = 0 // 未导入
	ImportStatusInProgress = 1 // 导入中
	ImportStatusSuccess    = 2 // 导入成功
	ImportStatusFailed     = 3 // 导入失败
)

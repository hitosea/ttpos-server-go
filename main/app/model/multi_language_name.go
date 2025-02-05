package model

// MultiLanguageName 结构体表示多语言名称
type MultiLanguageName struct {
	Id         uint   `gorm:"column:id;primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	EnName     string `gorm:"column:en_name;not null;default:'';comment:'英文名称'"`
	ZhName     string `gorm:"column:zh_name;not null;default:'';comment:'中文名称'"`
	ZhTwName   string `gorm:"column:zh_tw_name;not null;default:'';comment:'繁体中文名称'"`
	ThName     string `gorm:"column:th_name;not null;default:'';comment:'泰语名称'"`
	MyName     string `gorm:"column:my_name;not null;default:'';comment:'缅甸语名称'"`
	JaName     string `gorm:"column:ja_name;not null;default:'';comment:'日语名称'"`
	KoName     string `gorm:"column:ko_name;not null;default:'';comment:'韩语名称'"`
	TrName     string `gorm:"column:tr_name;not null;default:'';comment:'土耳其语名称'"`
	CreateTime int    `gorm:"column:create_time;not null;default:0;comment:'创建时间（时间戳）'"`
	UpdateTime int    `gorm:"column:update_time;not null;default:0;comment:'更新时间（时间戳）'"`
	DeleteTime int    `gorm:"column:delete_time;not null;default:0;comment:'删除时间（时间戳）'"`
}

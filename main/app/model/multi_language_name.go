package model

// MultiLanguageName 结构体表示多语言名称
type MultiLanguageName struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:'记录唯一标识符'"`
	Uuid       uint64 `gorm:"default:0;comment:'唯一标识符'"`
	EnName     string `gorm:"default:'';comment:'英文名称'"`
	ZhName     string `gorm:"default:'';comment:'中文名称'"`
	ZhTwName   string `gorm:"default:'';comment:'繁体中文名称'"`
	ThName     string `gorm:"default:'';comment:'泰语名称'"`
	MyName     string `gorm:"default:'';comment:'缅甸语名称'"`
	JaName     string `gorm:"default:'';comment:'日语名称'"`
	KoName     string `gorm:"default:'';comment:'韩语名称'"`
	TrName     string `gorm:"default:'';comment:'土耳其语名称'"`
	CreateTime int64  `gorm:"autoCreateTime;comment:'创建时间（时间戳）'"`
	UpdateTime int64  `gorm:"autoUpdateTime;comment:'更新时间（时间戳）'"`
	DeleteTime int64  `gorm:"default:0;comment:'删除时间（时间戳）'"`
}

package model

// MultiLanguageName 结构体表示多语言名称 ttpos_multi_language_name
type MultiLanguageName struct {
	BaseModel
	EnName   string `gorm:"default:'';column:en_name;comment:'英文名称'"`
	ZhName   string `gorm:"default:'';column:zh_name;comment:'中文名称'"`
	ZhTwName string `gorm:"default:'';column:zh_tw_name;comment:'繁体中文名称'"`
	ThName   string `gorm:"default:'';column:th_name;comment:'泰语名称'"`
	MyName   string `gorm:"default:'';column:my_name;comment:'缅甸语名称'"`
	JaName   string `gorm:"default:'';column:ja_name;comment:'日语名称'"`
	KoName   string `gorm:"default:'';column:ko_name;comment:'韩语名称'"`
	TrName   string `gorm:"default:'';column:tr_name;comment:'土耳其语名称'"`
}

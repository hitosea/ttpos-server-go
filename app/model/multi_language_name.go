package model

// MultiLanguageName 结构体表示多语言名称
type MultiLanguageName struct {
	Id         uint   `json:"id"`          // ID
	EnName     string `json:"en_name"`     // 英文名称
	ZhName     string `json:"zh_name"`     // 中文名称
	ZhTwName   string `json:"zh_tw_name"`  // 繁体中文名称
	ThName     string `json:"th_name"`     // 泰语名称
	MyName     string `json:"my_name"`     // 缅甸语名称
	JaName     string `json:"ja_name"`     // 日语名称
	KoName     string `json:"ko_name"`     // 韩语名称
	TrName     string `json:"tr_name"`     // 土耳其语名称
	CreateTime int    `json:"create_time"` // 创建时间
	UpdateTime int    `json:"update_time"` // 更新时间
	DeleteTime int    `json:"delete_time"` // 删除时间
}

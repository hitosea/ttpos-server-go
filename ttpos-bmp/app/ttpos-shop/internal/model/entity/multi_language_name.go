// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MultiLanguageName is the golang structure for table multi_language_name.
type MultiLanguageName struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`      // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"多语言名称ID"`   // 多语言名称ID
	EnName     string `json:"enName"     orm:"en_name"     description:"英文名称"`      // 英文名称
	ZhName     string `json:"zhName"     orm:"zh_name"     description:"中文名称"`      // 中文名称
	ZhTwName   string `json:"zhTwName"   orm:"zh_tw_name"  description:"繁体中文名称"`    // 繁体中文名称
	ThName     string `json:"thName"     orm:"th_name"     description:"泰语名称"`      // 泰语名称
	MyName     string `json:"myName"     orm:"my_name"     description:"缅甸语名称"`     // 缅甸语名称
	JaName     string `json:"jaName"     orm:"ja_name"     description:"日语名称"`      // 日语名称
	KoName     string `json:"koName"     orm:"ko_name"     description:"韩语名称"`      // 韩语名称
	TrName     string `json:"trName"     orm:"tr_name"     description:"土耳其语名称"`    // 土耳其语名称
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"` // 删除时间(时间戳)
}

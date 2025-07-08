// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MultiLanguageName is the golang structure of table ttpos_multi_language_name for DAO operations like Where/Data.
type MultiLanguageName struct {
	g.Meta     `orm:"table:ttpos_multi_language_name, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 多语言名称ID
	EnName     interface{} // 英文名称
	ZhName     interface{} // 中文名称
	ZhTwName   interface{} // 繁体中文名称
	ThName     interface{} // 泰语名称
	MyName     interface{} // 缅甸语名称
	JaName     interface{} // 日语名称
	KoName     interface{} // 韩语名称
	TrName     interface{} // 土耳其语名称
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

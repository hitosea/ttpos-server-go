// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminAccess is the golang structure for table admin_access.
type AdminAccess struct {
	Id           uint   `json:"id"           orm:"id"            description:"自增ID"`           // 自增ID
	Name         string `json:"name"         orm:"name"          description:"权限名称"`           // 权限名称
	Path         string `json:"path"         orm:"path"          description:"路由地址"`           // 路由地址
	ApiPath      string `json:"apiPath"      orm:"api_path"      description:"后端路由地址"`         // 后端路由地址
	ParentId     int    `json:"parentId"     orm:"parent_id"     description:"父级ID"`           // 父级ID
	Sort         int    `json:"sort"         orm:"sort"          description:"排序(数字越小越靠前)"`    // 排序(数字越小越靠前)
	Icon         string `json:"icon"         orm:"icon"          description:"菜单图标"`           // 菜单图标
	RedirectName string `json:"redirectName" orm:"redirect_name" description:"重定向名称"`          // 重定向名称
	IsRoute      int    `json:"isRoute"      orm:"is_route"      description:"是否路由 0=不是1=是"`   // 是否路由 0=不是1=是
	IsMenu       int    `json:"isMenu"       orm:"is_menu"       description:"是否菜单 0=不是1=是"`   // 是否菜单 0=不是1=是
	IsShow       int    `json:"isShow"       orm:"is_show"       description:"是否显示 0=不是1=是"`   // 是否显示 0=不是1=是
	Remark       string `json:"remark"       orm:"remark"        description:"描述"`             // 描述
	IsSupplier   int    `json:"isSupplier"   orm:"is_supplier"   description:"是否门店菜单 0=不是1=是"` // 是否门店菜单 0=不是1=是
	CreateTime   int    `json:"createTime"   orm:"create_time"   description:"创建时间（时间戳）"`      // 创建时间（时间戳）
	UpdateTime   int    `json:"updateTime"   orm:"update_time"   description:"更新时间（时间戳）"`      // 更新时间（时间戳）
	DeleteTime   int    `json:"deleteTime"   orm:"delete_time"   description:"删除时间（时间戳）"`      // 删除时间（时间戳）
}

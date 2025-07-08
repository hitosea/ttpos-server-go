// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Access is the golang structure for table access.
type Access struct {
	Id               uint   `json:"id"               orm:"id"                 description:"自增ID"`          // 自增ID
	Uuid             uint64 `json:"uuid"             orm:"uuid"               description:"权限ID"`          // 权限ID
	Name             string `json:"name"             orm:"name"               description:"权限名称"`          // 权限名称
	Path             string `json:"path"             orm:"path"               description:"路由地址"`          // 路由地址
	ApiPath          string `json:"apiPath"          orm:"api_path"           description:"后端路由地址"`        // 后端路由地址
	ParentUuid       uint64 `json:"parentUuid"       orm:"parent_uuid"        description:"父级ID"`          // 父级ID
	Sort             int    `json:"sort"             orm:"sort"               description:"排序(数字越小越靠前)"`   // 排序(数字越小越靠前)
	Icon             string `json:"icon"             orm:"icon"               description:"菜单图标"`          // 菜单图标
	RedirectName     string `json:"redirectName"     orm:"redirect_name"      description:"重定向名称"`         // 重定向名称
	IsRoute          int    `json:"isRoute"          orm:"is_route"           description:"是否是路由 0=不是1=是"` // 是否是路由 0=不是1=是
	IsMenu           int    `json:"isMenu"           orm:"is_menu"            description:"是否是菜单 0不是 1是"`  // 是否是菜单 0不是 1是
	IsShow           int    `json:"isShow"           orm:"is_show"            description:"是否显示1=显示0=不显示"` // 是否显示1=显示0=不显示
	PlusCategoryUuid uint64 `json:"plusCategoryUuid" orm:"plus_category_uuid" description:"插件分类ID"`        // 插件分类ID
	Remark           string `json:"remark"           orm:"remark"             description:"描述"`            // 描述
	IsSupplier       int    `json:"isSupplier"       orm:"is_supplier"        description:"是否门店菜单0否1是"`    // 是否门店菜单0否1是
	CreateTime       uint   `json:"createTime"       orm:"create_time"        description:"创建时间(时间戳)"`     // 创建时间(时间戳)
	UpdateTime       uint   `json:"updateTime"       orm:"update_time"        description:"更新时间(时间戳)"`     // 更新时间(时间戳)
	DeleteTime       uint   `json:"deleteTime"       orm:"delete_time"        description:"删除时间(时间戳)"`     // 删除时间(时间戳)
}

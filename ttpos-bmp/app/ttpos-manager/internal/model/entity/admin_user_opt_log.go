// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AdminUserOptLog is the golang structure for table admin_user_opt_log.
type AdminUserOptLog struct {
	Id          uint   `json:"id"          orm:"id"            description:"自增ID"`      // 自增ID
	AdminUserId int    `json:"adminUserId" orm:"admin_user_id" description:"用户ID"`      // 用户ID
	Title       string `json:"title"       orm:"title"         description:"标题"`        // 标题
	Url         string `json:"url"         orm:"url"           description:"访问url"`     // 访问url
	RequestType string `json:"requestType" orm:"request_type"  description:"请求类型"`      // 请求类型
	Browser     string `json:"browser"     orm:"browser"       description:"浏览器"`       // 浏览器
	Agent       string `json:"agent"       orm:"agent"         description:"浏览器信息"`     // 浏览器信息
	Content     string `json:"content"     orm:"content"       description:"操作内容"`      // 操作内容
	Ip          string `json:"ip"          orm:"ip"            description:"登录ip"`      // 登录ip
	CreateTime  int    `json:"createTime"  orm:"create_time"   description:"创建时间（时间戳）"` // 创建时间（时间戳）
	UpdateTime  int    `json:"updateTime"  orm:"update_time"   description:"更新时间（时间戳）"` // 更新时间（时间戳）
	DeleteTime  int    `json:"deleteTime"  orm:"delete_time"   description:"删除时间（时间戳）"` // 删除时间（时间戳）
}

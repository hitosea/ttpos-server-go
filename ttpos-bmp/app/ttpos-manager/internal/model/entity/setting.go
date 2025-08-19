// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Setting is the golang structure for table setting.
type Setting struct {
	Key        string `json:"key"        orm:"key"         description:"设置项标示"`        // 设置项标示
	Describe   string `json:"describe"   orm:"describe"    description:"设置项描述"`        // 设置项描述
	Values     string `json:"values"     orm:"values"      description:"设置内容（json格式）"` // 设置内容（json格式）
	CreateTime int    `json:"createTime" orm:"create_time" description:"创建时间（时间戳）"`    // 创建时间（时间戳）
	UpdateTime int    `json:"updateTime" orm:"update_time" description:"更新时间（时间戳）"`    // 更新时间（时间戳）
	DeleteTime int    `json:"deleteTime" orm:"delete_time" description:"删除时间（时间戳）"`    // 删除时间（时间戳）
}

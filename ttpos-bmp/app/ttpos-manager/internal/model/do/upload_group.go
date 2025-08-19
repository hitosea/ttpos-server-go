// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UploadGroup is the golang structure of table ttpos_upload_group for DAO operations like Where/Data.
type UploadGroup struct {
	g.Meta     `orm:"table:ttpos_upload_group, do:true"`
	GroupId    interface{} // 自增ID
	GroupType  interface{} // 文件类型
	GroupName  interface{} // 分类名称
	Sort       interface{} // 分类排序(数字越小越靠前)
	CreateTime interface{} // 创建时间（时间戳）
	UpdateTime interface{} // 更新时间（时间戳）
	DeleteTime interface{} // 删除时间（时间戳）
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UploadFile is the golang structure of table ttpos_upload_file for DAO operations like Where/Data.
type UploadFile struct {
	g.Meta     `orm:"table:ttpos_upload_file, do:true"`
	FileId     interface{} // 自增ID
	Storage    interface{} // 存储方式
	GroupId    interface{} // 文件分组ID
	FileUrl    interface{} // 存储域名
	SaveName   interface{} // 保存路径
	FileName   interface{} // 文件路径
	FileSize   interface{} // 文件大小(字节)
	FileType   interface{} // 文件类型
	RealName   interface{} // 文件真实名
	UrlParam   interface{} // 签名参数
	Extension  interface{} // 文件扩展名
	IsUser     interface{} // 是否为c端用户上传
	IsRecycle  interface{} // 是否已回收
	CreateTime interface{} // 创建时间（时间戳）
	UpdateTime interface{} // 更新时间（时间戳）
	DeleteTime interface{} // 删除时间（时间戳）
}

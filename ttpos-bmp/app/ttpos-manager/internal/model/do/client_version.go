// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ClientVersion is the golang structure of table ttpos_client_version for DAO operations like Where/Data.
type ClientVersion struct {
	g.Meta         `orm:"table:ttpos_client_version, do:true"`
	Id             interface{} // 自增ID
	Type           interface{} // 类型： 1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
	Brand          interface{} // 品牌
	IsPublish      interface{} // 是否发布 0-否 1-是
	Md5Hash        interface{} // 谷歌云 md5-hash 值
	DownloadNum    interface{} // 下载数量
	VersionNumber  interface{} // 版本号
	VersionName    interface{} // 版本名称
	ApkVersionCode interface{} // Apk版本code
	ApkData        interface{} // apk数据
	ForcedUpdate   interface{} // 强制更新 0否 1是
	PackageUrl     interface{} // 包地址
	OriginalName   interface{} // 文件原名称
	UpdateLog      interface{} // 更新日志
	CreateTime     interface{} // 创建时间（时间戳）
	UpdateTime     interface{} // 更新时间（时间戳）
	DeleteTime     interface{} // 删除时间（时间戳）
}

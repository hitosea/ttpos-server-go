// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ClientVersion is the golang structure for table client_version.
type ClientVersion struct {
	Id             uint   `json:"id"             orm:"id"               description:"自增ID"`                             // 自增ID
	Type           int    `json:"type"           orm:"type"             description:"类型： 1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端"` // 类型： 1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端
	Brand          int    `json:"brand"          orm:"brand"            description:"品牌"`                               // 品牌
	IsPublish      int    `json:"isPublish"      orm:"is_publish"       description:"是否发布 0-否 1-是"`                     // 是否发布 0-否 1-是
	Md5Hash        string `json:"md5Hash"        orm:"md5_hash"         description:"谷歌云 md5-hash 值"`                   // 谷歌云 md5-hash 值
	DownloadNum    int    `json:"downloadNum"    orm:"download_num"     description:"下载数量"`                             // 下载数量
	VersionNumber  string `json:"versionNumber"  orm:"version_number"   description:"版本号"`                              // 版本号
	VersionName    string `json:"versionName"    orm:"version_name"     description:"版本名称"`                             // 版本名称
	ApkVersionCode string `json:"apkVersionCode" orm:"apk_version_code" description:"Apk版本code"`                        // Apk版本code
	ApkData        string `json:"apkData"        orm:"apk_data"         description:"apk数据"`                            // apk数据
	ForcedUpdate   int    `json:"forcedUpdate"   orm:"forced_update"    description:"强制更新 0否 1是"`                       // 强制更新 0否 1是
	PackageUrl     string `json:"packageUrl"     orm:"package_url"      description:"包地址"`                              // 包地址
	OriginalName   string `json:"originalName"   orm:"original_name"    description:"文件原名称"`                            // 文件原名称
	UpdateLog      string `json:"updateLog"      orm:"update_log"       description:"更新日志"`                             // 更新日志
	CreateTime     int    `json:"createTime"     orm:"create_time"      description:"创建时间（时间戳）"`                        // 创建时间（时间戳）
	UpdateTime     int    `json:"updateTime"     orm:"update_time"      description:"更新时间（时间戳）"`                        // 更新时间（时间戳）
	DeleteTime     int    `json:"deleteTime"     orm:"delete_time"      description:"删除时间（时间戳）"`                        // 删除时间（时间戳）
}

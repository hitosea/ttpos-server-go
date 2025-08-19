// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UploadFile is the golang structure for table upload_file.
type UploadFile struct {
	FileId     uint   `json:"fileId"     orm:"file_id"     description:"自增ID"`      // 自增ID
	Storage    string `json:"storage"    orm:"storage"     description:"存储方式"`      // 存储方式
	GroupId    uint   `json:"groupId"    orm:"group_id"    description:"文件分组ID"`    // 文件分组ID
	FileUrl    string `json:"fileUrl"    orm:"file_url"    description:"存储域名"`      // 存储域名
	SaveName   string `json:"saveName"   orm:"save_name"   description:"保存路径"`      // 保存路径
	FileName   string `json:"fileName"   orm:"file_name"   description:"文件路径"`      // 文件路径
	FileSize   uint   `json:"fileSize"   orm:"file_size"   description:"文件大小(字节)"`  // 文件大小(字节)
	FileType   string `json:"fileType"   orm:"file_type"   description:"文件类型"`      // 文件类型
	RealName   string `json:"realName"   orm:"real_name"   description:"文件真实名"`     // 文件真实名
	UrlParam   string `json:"urlParam"   orm:"url_param"   description:"签名参数"`      // 签名参数
	Extension  string `json:"extension"  orm:"extension"   description:"文件扩展名"`     // 文件扩展名
	IsUser     uint   `json:"isUser"     orm:"is_user"     description:"是否为c端用户上传"` // 是否为c端用户上传
	IsRecycle  uint   `json:"isRecycle"  orm:"is_recycle"  description:"是否已回收"`     // 是否已回收
	CreateTime int    `json:"createTime" orm:"create_time" description:"创建时间（时间戳）"` // 创建时间（时间戳）
	UpdateTime int    `json:"updateTime" orm:"update_time" description:"更新时间（时间戳）"` // 更新时间（时间戳）
	DeleteTime int    `json:"deleteTime" orm:"delete_time" description:"删除时间（时间戳）"` // 删除时间（时间戳）
}

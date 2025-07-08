// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// File is the golang structure for table file.
type File struct {
	Id            uint   `json:"id"            orm:"id"              description:"自增ID"`      // 自增ID
	Uuid          uint64 `json:"uuid"          orm:"uuid"            description:"文件ID"`      // 文件ID
	Storage       string `json:"storage"       orm:"storage"         description:"存储方式"`      // 存储方式
	GroupUuid     uint64 `json:"groupUuid"     orm:"group_uuid"      description:"文件分组UUID"`  // 文件分组UUID
	FileUrl       string `json:"fileUrl"       orm:"file_url"        description:"存储域名"`      // 存储域名
	SaveName      string `json:"saveName"      orm:"save_name"       description:"保存路径"`      // 保存路径
	FileName      string `json:"fileName"      orm:"file_name"       description:"文件路径"`      // 文件路径
	FileSize      int    `json:"fileSize"      orm:"file_size"       description:"文件大小(字节)"`  // 文件大小(字节)
	FileType      string `json:"fileType"      orm:"file_type"       description:"文件类型"`      // 文件类型
	RealName      string `json:"realName"      orm:"real_name"       description:"文件真实名"`     // 文件真实名
	UrlParam      string `json:"urlParam"      orm:"url_param"       description:"签名参数"`      // 签名参数
	IndexFileName string `json:"indexFileName" orm:"index_file_name" description:"文件唯一名"`     // 文件唯一名
	Extension     string `json:"extension"     orm:"extension"       description:"文件扩展名"`     // 文件扩展名
	IsUser        int    `json:"isUser"        orm:"is_user"         description:"是否为c端用户上传"` // 是否为c端用户上传
	IsRecycle     int    `json:"isRecycle"     orm:"is_recycle"      description:"是否已回收"`     // 是否已回收
	CreateTime    uint   `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"` // 创建时间(时间戳)
	UpdateTime    uint   `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"` // 更新时间(时间戳)
	DeleteTime    uint   `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"` // 删除时间(时间戳)
}

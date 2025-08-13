package req

import "ttpos-server-go/app/dto"

// FileListReq 文件列表请求
type FileListReq struct {
	GroupUuid   uint64 `form:"group_uuid,default=0"` // 分组UUID
	FileType    string `form:"file_type,default="`   // 文件类型：image、video
	dto.PageReq        // 分页参数
}

// DeleteFileReq 删除文件请求
type DeleteFileReq struct {
	Uuid uint64 `json:"uuid" binding:"required,gt=0"` // 文件UUID
}

// UploadImageReq 上传图片请求
type UploadImageReq struct {
	GroupId uint64 `form:"group_id,default=0"` // 分组ID
	Source  string `form:"source,default="`    // 来源标识
}

// UploadVideoReq 上传视频请求
type UploadVideoReq struct {
	GroupId uint64 `form:"group_id,default=0"` // 分组ID
	Size    int    `form:"size,default=30"`    // 最大文件大小(MB)
}

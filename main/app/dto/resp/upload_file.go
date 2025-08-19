package resp

import "ttpos-server-go/app/dto"

// UploadFileResp 上传文件响应
type UploadFileResp struct {
	Uuid          uint64 `json:"uuid"`            // 文件UUID
	GroupUuid     uint64 `json:"group_uuid"`      // 分组UUID
	Storage       string `json:"storage"`         // 存储引擎
	FileUrl       string `json:"file_url"`        // 文件访问域名
	FileName      string `json:"file_name"`       // 文件名称
	SaveName      string `json:"save_name"`       // 保存路径
	FileSize      int64  `json:"file_size"`       // 文件大小(字节)
	FileType      string `json:"file_type"`       // 文件类型
	Extension     string `json:"extension"`       // 文件扩展名
	RealName      string `json:"real_name"`       // 原始文件名
	IndexFileName string `json:"index_file_name"` // 索引文件名
	UrlParam      string `json:"url_param"`       // URL参数
	FilePath      string `json:"file_path"`       // 文件路径
	CreateTime    int    `json:"create_time"`     // 创建时间
}

// FileListResp 文件列表响应
type FileListResp struct {
	List []*UploadFileResp `json:"list"` // 文件列表
	Meta dto.PageResponse  `json:"meta"` // 分页信息
}

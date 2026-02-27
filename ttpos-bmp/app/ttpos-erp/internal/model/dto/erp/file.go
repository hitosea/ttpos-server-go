package erp

// FileUploadReq 文件上传请求参数
type FileUploadReq struct {
	// FileName 文件名
	FileName string `json:"file_name"`
	// FileContent 文件内容（字节数组）
	FileContent []byte `json:"-"`
	// DocType 关联的文档类型（可选）
	DocType string `json:"doctype,omitempty"`
	// DocName 关联的文档名称（可选）
	DocName string `json:"docname,omitempty"`
	// IsPrivate 是否私有文件（0: 公开, 1: 私有）
	IsPrivate int `json:"is_private,omitempty"`
	// Folder 文件夹路径（可选）
	Folder string `json:"folder,omitempty"`
}

// FileUploadBase64Req Base64 编码方式的文件上传请求
type FileUploadBase64Req struct {
	// FileName 文件名
	FileName string `json:"filename"`
	// FileData Base64 编码的文件内容
	FileData string `json:"filedata"`
	// DocType 关联的文档类型
	DocType string `json:"doctype,omitempty"`
	// DocName 关联的文档名称
	DocName string `json:"docname,omitempty"`
	// IsPrivate 是否私有文件（0: 公开, 1: 私有）
	IsPrivate int `json:"is_private,omitempty"`
	// FromForm 表单来源标记（固定为 1）
	FromForm int `json:"from_form,omitempty"`
}

// FileUploadUrlReq URL 方式的文件上传请求（ERPNext 会自动下载远程文件）
type FileUploadUrlReq struct {
	// FileUrl 远程文件 URL（必填）
	FileUrl string `json:"file_url"`
	// FileName 文件名（可选，不填则从 URL 提取）
	FileName string `json:"file_name,omitempty"`
	// DocType 关联的文档类型（可选）
	DocType string `json:"doctype,omitempty"`
	// DocName 关联的文档名称（可选）
	DocName string `json:"docname,omitempty"`
	// IsPrivate 是否私有文件（0: 公开, 1: 私有）
	IsPrivate int `json:"is_private,omitempty"`
	// Folder 文件夹路径（可选）
	Folder string `json:"folder,omitempty"`
}

// FileUploadResp 文件上传响应
type FileUploadResp struct {
	// Name 文件文档名称（File doctype 的 name 字段）
	Name string `json:"name,omitempty"`
	// FileName 文件名
	FileName string `json:"file_name,omitempty"`
	// FileUrl 文件访问 URL
	FileUrl string `json:"file_url,omitempty"`
	// IsPrivate 是否私有
	IsPrivate int `json:"is_private,omitempty"`
	// AttachedToDocType 关联的文档类型
	AttachedToDocType string `json:"attached_to_doctype,omitempty"`
	// AttachedToName 关联的文档名称
	AttachedToName string `json:"attached_to_name,omitempty"`
	// FileSize 文件大小（字节）
	FileSize int64 `json:"file_size,omitempty"`
	// ContentType 文件 MIME 类型
	ContentType string `json:"content_type,omitempty"`
}

// FileListReq 文件列表查询请求
type FileListReq struct {
	// DocType 关联的文档类型
	DocType string `json:"attached_to_doctype,omitempty"`
	// DocName 关联的文档名称
	DocName string `json:"attached_to_name,omitempty"`
	// IsPrivate 是否私有
	IsPrivate *int `json:"is_private,omitempty"`
	// Folder 文件夹
	Folder string `json:"folder,omitempty"`
	// LimitStart 分页起始位置
	LimitStart int `json:"limit_start,omitempty"`
	// Limit 分页数量
	Limit int `json:"limit,omitempty"`
}

// FileDeleteReq 文件删除请求
type FileDeleteReq struct {
	// Name 文件文档名称
	Name string `json:"name"`
}

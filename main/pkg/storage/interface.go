package storage

import (
	"io"
)

// StorageEngine 存储引擎接口
type StorageEngine interface {
	// Upload 上传文件
	// thumb: 缩略图尺寸，0为不生成缩略图
	Upload(thumb int) (string, error)

	// Delete 删除文件
	Delete(fileName string) error

	// SetUploadFile 设置上传的文件信息
	SetUploadFile(fileReader io.Reader, fileName string, contentType string, fileSize int64) error

	// SetLocalFile 设置本地文件路径
	SetLocalFile(filePath string) error

	// GetFileName 获取文件名
	GetFileName() string

	// GetError 获取错误信息
	GetError() string

	// GetUrlParam 获取URL参数（如Google云存储的签名参数）
	GetUrlParam() string

	// 设置companyUuid
	SetCompanyUuid(companyUuid uint64)
}

// FileInfo 文件信息结构体
type FileInfo struct {
	FileName    string `json:"file_name"`    // 文件名
	SaveName    string `json:"save_name"`    // 保存路径
	FileSize    int64  `json:"file_size"`    // 文件大小
	Extension   string `json:"extension"`    // 文件扩展名
	ContentType string `json:"content_type"` // 文件MIME类型
	RealName    string `json:"real_name"`    // 原始文件名
}

// StorageConfig 存储配置
type StorageConfig struct {
	Default string                    `json:"default"` // 默认存储引擎
	Engine  map[string]map[string]any `json:"engine"`  // 存储引擎配置
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	UploadPath string `json:"upload_path"` // 上传路径
}

// GoogleConfig Google云存储配置
type GoogleConfig struct {
	CredentialsFile  string `json:"credentials_file"`  // 凭证文件路径
	Bucket           string `json:"bucket"`            // 存储桶名称
	UploadsCatalogue string `json:"uploads_catalogue"` // 上传目录
	Domain           string `json:"domain"`            // 访问域名
}

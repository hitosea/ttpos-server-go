package service

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/storage"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// IUploadFileSrv 上传文件服务接口
type IUploadFileSrv interface {
	// UploadImage 上传图片
	UploadImage(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, source string) (*resp.UploadFileResp, error)
	// UploadVideo 上传视频
	UploadVideo(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, maxSize int) (*resp.UploadFileResp, error)
	// UploadDocument 上传文档（PDF、Word、Excel等）
	UploadDocument(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64) (*resp.UploadFileResp, error)
	// GetUploadFile 获取文件
	GetUploadFile(ctx context.Context, uuid uint64) (*resp.UploadFileResp, error)
}

// UploadFileSrvImpl 上传文件服务实现
type UploadFileSrvImpl struct {
	dbm            *database.DBManager
	storageFactory *storage.Factory
}

// NewUploadFileSrv 创建上传文件服务
func NewUploadFileSrv(dbm *database.DBManager) IUploadFileSrv {
	return NewUploadFileSrvImpl(dbm)
}

// NewUploadFileSrvImpl 创建上传文件服务实现
func NewUploadFileSrvImpl(dbm *database.DBManager) IUploadFileSrv {
	return &UploadFileSrvImpl{
		dbm: dbm,
		storageFactory: storage.NewFactory(&storage.StorageConfig{
			Default: viper.GetString("STORAGE_DRIVER"),
			Engine: map[string]map[string]any{
				"local": {
					"upload_path": "public/uploads",
				},
				"google": {
					"credentials_file":  config.GoogleBucket.GoogleApplicationCredentialsFileName,
					"bucket":            config.GoogleBucket.GoogleApplicationUploadsBucketName,
					"uploads_catalogue": config.GoogleBucket.GoogleApplicationUploadsCatalogueName,
					"domain":            fmt.Sprintf("https://storage.googleapis.com/%s/%s", config.GoogleBucket.GoogleApplicationUploadsBucketName, config.GoogleBucket.GoogleApplicationUploadsCatalogueName),
				},
			},
		}),
	}
}

// UploadImage 上传图片
func (s *UploadFileSrvImpl) UploadImage(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, source string) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"jpg", "jpeg", "png", "webp"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持JPG、JPEG、PNG、WEBP格式")
	}

	// 确定缩略图尺寸
	thumbSize := 500
	if source == "" {
		thumbSize = 5000
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, "image", thumbSize)
}

// UploadVideo 上传视频
func (s *UploadFileSrvImpl) UploadVideo(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, maxSize int) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"avi", "mpeg", "mov", "mp4"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持AVI、MPEG、MOV、MP4格式")
	}

	// 检查文件大小
	maxSizeBytes := int64(maxSize * 1024 * 1024)
	if fileSize > maxSizeBytes {
		if maxSize > 30 {
			return nil, fmt.Errorf("文件大小不能超过30MB")
		}
		return nil, fmt.Errorf("文件大小不能超过10MB")
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, "video", 0)
}

// GetUploadFile 获取文件
func (s *UploadFileSrvImpl) GetUploadFile(ctx context.Context, uuid uint64) (*resp.UploadFileResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	var uploadFile model.File
	err := db.Model(&model.File{}).Where("uuid = ? AND delete_time = 0", uuid).First(&uploadFile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在")
		}
		return nil, fmt.Errorf("获取文件失败: %v", err)
	}
	return &resp.UploadFileResp{
		Uuid:          uploadFile.Uuid,
		GroupUuid:     uploadFile.GroupUuid,
		Storage:       uploadFile.Storage,
		FileUrl:       uploadFile.FileUrl,
		FileName:      uploadFile.FileName,
		SaveName:      uploadFile.SaveName,
		FileSize:      int64(uploadFile.FileSize),
		FileType:      uploadFile.FileType,
		Extension:     uploadFile.Extension,
		RealName:      uploadFile.RealName,
		IndexFileName: uploadFile.IndexFileName,
		UrlParam:      uploadFile.UrlParam,
		FilePath:      uploadFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		CreateTime:    int(uploadFile.CreateTime),
	}, nil
}

// uploadFile 通用上传文件方法
func (s *UploadFileSrvImpl) uploadFile(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64, fileType string, thumbSize int) (*resp.UploadFileResp, error) {
	// 创建存储引擎
	engine, err := s.storageFactory.CreateEngine()
	if err != nil {
		return nil, fmt.Errorf("创建存储引擎失败: %v", err)
	}

	// 设置文件信息
	if err := engine.SetUploadFile(fileReader, fileName, fileType, fileSize); err != nil {
		return nil, fmt.Errorf("设置文件信息失败: %v", err)
	}

	// 设置公司ID
	engine.SetCompanyUuid(ctx.GetCompanyUuid())

	// 上传文件
	saveName, err := engine.Upload(thumbSize)
	if err != nil {
		return nil, fmt.Errorf("文件上传失败: %v", err)
	}

	if saveName == "" {
		errMsg := engine.GetError()
		if errMsg == "" {
			errMsg = "未知错误"
		}
		return nil, fmt.Errorf("上传失败: %s", errMsg)
	}

	// 标准化路径分隔符
	saveName = strings.ReplaceAll(saveName, "\\", "/")

	// 获取文件名和URL参数
	generatedFileName := engine.GetFileName()
	urlParam := engine.GetUrlParam()

	// 获取存储配置
	storageEngine := s.storageFactory.GetDefaultEngine()

	var fileUrl string
	if domain, ok := s.storageFactory.GetDefaultEngineConfig()["domain"]; ok {
		fileUrl = domain.(string)
	}

	// 创建文件记录
	uploadFile := model.File{
		BaseModel: model.BaseModel{
			Uuid: s.generateUuid(),
		},
		GroupUuid:     groupId,
		Storage:       storageEngine,
		FileUrl:       fileUrl,
		FileName:      generatedFileName,
		SaveName:      saveName,
		FileSize:      int(fileSize),
		FileType:      fileType,
		Extension:     s.getFileExtension(fileName),
		RealName:      fileName,
		IndexFileName: s.getFileNameWithoutExt(fileName),
		UrlParam:      urlParam,
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	err = db.Create(&uploadFile).Error
	if err != nil {
		return nil, fmt.Errorf("保存文件记录失败: %v", err)
	}

	// 返回结果
	return &resp.UploadFileResp{
		Uuid:          uploadFile.Uuid,
		GroupUuid:     uploadFile.GroupUuid,
		Storage:       uploadFile.Storage,
		FileUrl:       uploadFile.FileUrl,
		FileName:      uploadFile.FileName,
		SaveName:      uploadFile.SaveName,
		FileSize:      int64(uploadFile.FileSize),
		FileType:      uploadFile.FileType,
		Extension:     uploadFile.Extension,
		RealName:      uploadFile.RealName,
		IndexFileName: uploadFile.IndexFileName,
		UrlParam:      uploadFile.UrlParam,
		FilePath:      uploadFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		CreateTime:    int(uploadFile.CreateTime),
	}, nil
}

// isAllowedExtension 检查文件扩展名是否允许
func (s *UploadFileSrvImpl) isAllowedExtension(ext string, allowedExts []string) bool {
	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

// getFileExtension 获取文件扩展名
func (s *UploadFileSrvImpl) getFileExtension(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext != "" && ext[0] == '.' {
		return ext[1:]
	}
	return ext
}

// getFileNameWithoutExt 获取不含扩展名的文件名
func (s *UploadFileSrvImpl) getFileNameWithoutExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// UploadDocument 上传文档（PDF、Word、Excel等）
func (s *UploadFileSrvImpl) UploadDocument(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64, groupId uint64) (*resp.UploadFileResp, error) {
	// 验证文件类型
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension != "" && extension[0] == '.' {
		extension = extension[1:]
	}

	allowedExts := []string{"pdf", "doc", "docx", "xls", "xlsx", "jpg", "jpeg", "png", "gif"}
	if !s.isAllowedExtension(extension, allowedExts) {
		return nil, fmt.Errorf("仅支持PDF、Word、Excel、JPG、PNG、GIF格式")
	}

	// 检查文件大小（20MB = 20 * 1024 * 1024 bytes）
	maxSizeBytes := int64(20 * 1024 * 1024)
	if fileSize > maxSizeBytes {
		return nil, fmt.Errorf("文件大小不能超过20MB")
	}

	// 确定文件类型
	fileType := "document"
	imageExts := []string{"jpg", "jpeg", "png", "gif"}
	if s.isAllowedExtension(extension, imageExts) {
		fileType = "image"
	}

	return s.uploadFile(ctx, fileReader, fileName, fileSize, groupId, fileType, 0)
}

// generateUuid 生成UUID
func (s *UploadFileSrvImpl) generateUuid() uint64 {
	if uuid, err := utils.GetID(); err == nil {
		return uuid
	}
	// 如果生成失败，使用时间戳作为后备方案
	return uint64(time.Now().UnixNano())
}

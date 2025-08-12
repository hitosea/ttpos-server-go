package engines

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ttpos-server-go/pkg/storage"
)

// Local 本地文件存储引擎
type Local struct {
	*storage.Server
	config      *storage.LocalConfig
	companyUuid uint64
}

// NewLocal 创建本地存储引擎实例
func NewLocal(config *storage.LocalConfig) *Local {
	return &Local{
		Server: storage.NewServer(),
		config: config,
	}
}

// Upload 上传文件
func (l *Local) Upload(thumb int) (string, error) {
	// 确定上传路径
	uploadPath := l.getUploadPath()

	// 创建目录
	if err := l.createUploadDir(uploadPath); err != nil {
		l.SetError(fmt.Sprintf("创建上传目录失败: %v", err))
		return "", err
	}

	// 完整文件路径
	fullPath := filepath.Join(uploadPath, l.GetFileName())

	// 保存文件
	if err := l.saveFile(fullPath); err != nil {
		l.SetError(fmt.Sprintf("保存文件失败: %v", err))
		return "", err
	}

	// 返回相对路径
	relativePath := l.getRelativePath(fullPath)
	return relativePath, nil
}

// Delete 删除文件
func (l *Local) Delete(fileName string) error {
	// 构建完整文件路径
	fullPath := filepath.Join(l.config.UploadPath, fileName)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// 文件不存在，认为删除成功
		return nil
	}

	// 删除文件
	return os.Remove(fullPath)
}

// GetUrlParam 获取URL参数（本地存储不需要参数）
func (l *Local) GetUrlParam() string {
	return ""
}

// getUploadPath 获取上传路径
func (l *Local) getUploadPath() string {
	// 基础上传路径
	basePath := l.config.UploadPath
	if basePath == "" {
		basePath = "public/uploads"
	}

	// 根据应用ID确定子目录
	var subDir string
	if l.companyUuid > 0 {
		subDir = "shop" + strconv.FormatUint(l.companyUuid, 10)
	} else {
		subDir = "saas"
	}

	// 添加日期子目录
	dateDir := time.Now().Format("20060102")

	return filepath.Join(basePath, subDir, dateDir)
}

// createUploadDir 创建上传目录
func (l *Local) createUploadDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// saveFile 保存文件
func (l *Local) saveFile(filePath string) error {
	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if l.GetFile() != nil {
		// 从Reader复制数据
		_, err = io.Copy(dst, l.GetFile())
		return err
	} else if l.GetFilePath() != "" {
		// 从本地文件复制
		src, err := os.Open(l.GetFilePath())
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		return err
	}

	return fmt.Errorf("没有可用的文件源")
}

// getRelativePath 获取相对路径
func (l *Local) getRelativePath(fullPath string) string {
	// 移除基础路径，返回相对路径
	if rel, err := filepath.Rel(l.config.UploadPath, fullPath); err == nil {
		return filepath.ToSlash(rel) // 转换为Unix风格的路径分隔符
	}
	return fullPath
}

// SetCompanyUuid 设置公司ID
func (l *Local) SetCompanyUuid(companyUuid uint64) {
	l.companyUuid = companyUuid
}

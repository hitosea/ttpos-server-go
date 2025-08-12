package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ttpos-server-go/pkg/utils"
)

// Server 存储服务器基础类
type Server struct {
	file        io.Reader // 文件内容
	fileName    string    // 文件名
	filePath    string    // 文件路径（本地文件使用）
	contentType string    // 文件MIME类型
	fileSize    int64     // 文件大小
	realName    string    // 原始文件名
	extension   string    // 文件扩展名
	error       string    // 错误信息
}

// NewServer 创建服务器实例
func NewServer() *Server {
	return &Server{}
}

// SetUploadFile 设置上传的文件信息
func (s *Server) SetUploadFile(fileReader io.Reader, fileName string, contentType string, fileSize int64) error {
	s.file = fileReader
	s.realName = fileName
	s.contentType = contentType
	s.fileSize = fileSize

	// 获取文件扩展名
	s.extension = strings.ToLower(filepath.Ext(fileName))
	if s.extension != "" && s.extension[0] == '.' {
		s.extension = s.extension[1:]
	}

	// 生成唯一文件名
	s.fileName = s.generateFileName()

	return nil
}

// SetLocalFile 设置本地文件路径
func (s *Server) SetLocalFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	s.filePath = filePath
	s.realName = filepath.Base(filePath)

	// 获取文件扩展名
	s.extension = strings.ToLower(filepath.Ext(filePath))
	if s.extension != "" && s.extension[0] == '.' {
		s.extension = s.extension[1:]
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	s.fileSize = fileInfo.Size()

	// 生成唯一文件名
	s.fileName = s.generateFileName()

	return nil
}

// GetFileName 获取文件名
func (s *Server) GetFileName() string {
	return s.fileName
}

// GetError 获取错误信息
func (s *Server) GetError() string {
	return s.error
}

// GetRealPath 获取文件真实路径（仅用于本地文件）
func (s *Server) GetRealPath() string {
	return s.filePath
}

// GetFileInfo 获取文件信息
func (s *Server) GetFileInfo() *FileInfo {
	return &FileInfo{
		FileName:    s.fileName,
		FileSize:    s.fileSize,
		Extension:   s.extension,
		ContentType: s.contentType,
		RealName:    s.realName,
	}
}

// generateFileName 生成唯一文件名
func (s *Server) generateFileName() string {
	// 生成时间戳
	timestamp := time.Now().Format("20060102150405")

	// 生成随机字符串
	random := utils.RandomString(9, utils.Numbers, utils.LowerLetters)

	// 组合文件名
	if s.extension != "" {
		return fmt.Sprintf("%s%s.%s", timestamp, random, s.extension)
	}
	return fmt.Sprintf("%s%s", timestamp, random)
}

// SetError 设置错误信息
func (s *Server) SetError(err string) {
	s.error = err
}

// GetFile 获取文件读取器
func (s *Server) GetFile() io.Reader {
	return s.file
}

// GetFilePath 获取文件路径
func (s *Server) GetFilePath() string {
	return s.filePath
}

// GetContentType 获取文件MIME类型
func (s *Server) GetContentType() string {
	return s.contentType
}

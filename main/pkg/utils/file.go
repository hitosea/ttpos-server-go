package utils

import (
	"os"
	"path/filepath"
)

// IsFileExist 判断文件是否存在
func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// CreateFile 创建文件，如果父目录不存在则自动创建
// path: 文件路径
// content: 文件内容
// perm: 文件权限，例如 0644
func CreateFile(path string, content []byte, perm os.FileMode) error {
	// 创建父目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建并写入文件
	return os.WriteFile(path, content, perm)
}

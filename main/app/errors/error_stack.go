package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	pkgErr "github.com/pkg/errors"
)

func WithMessage(err error, messages ...string) error {
	var msg string
	msg = strings.Join(messages, " ")
	errorMessage := getErrorMessage(msg)
	return pkgErr.WithMessage(err, errorMessage)
}

func getErrorMessage(msg string) string {
	// 获取调用者的信息
	_, file, line, _ := runtime.Caller(2) // 1 表示获取调用 newErrorWithLocation 的位置
	// 格式化错误信息，包括文件名和行号
	file = getLastNParts(file, 3)
	return fmt.Sprintf("[%s:%d]%s", file, line, msg)
}

func getLastNParts(path string, n int) string {
	// 将路径标准化为使用正斜杠（/）
	path = filepath.ToSlash(path)

	// 将路径按分隔符（/）分割为片段
	parts := strings.Split(path, "/")

	// 确保我们不会尝试获取超过路径长度的部分
	if len(parts) <= n {
		return path // 如果路径部分少于或等于 n，返回整个路径
	}

	// 获取最后 n 个部分，并将它们合并为路径
	i := len(parts) - n
	if i < 0 {
		i = 0
	}
	return strings.Join(parts[i:], "/")
}

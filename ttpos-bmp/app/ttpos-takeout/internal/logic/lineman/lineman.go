// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sLineman LINE MAN 服务（统一管理 Token 和菜单同步）
type sLineman struct{}

func init() {
	service.RegisterLineman(New())
}

// New 创建 Lineman 服务实例
func New() *sLineman {
	return &sLineman{}
}

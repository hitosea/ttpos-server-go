// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	lineman_client "ttpos-bmp/app/ttpos-takeout/internal/client/lineman"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// sLineman LINE MAN 服务（统一管理菜单同步、订单处理、状态更新）
type sLineman struct {
	menuStatusClient     IMenuStatusClient
	modifierStatusClient IModifierStatusClient
}

func init() {
	service.RegisterLineman(New())
}

// New 创建 Lineman 服务实例
func New() *sLineman {
	return &sLineman{
		menuStatusClient:     lineman_client.NewMenuStatusClient(),
		modifierStatusClient: lineman_client.NewModifierStatusClient(),
	}
}

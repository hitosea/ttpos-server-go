package grab

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/grab/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// MenuSyncState 处理 Grab 菜单同步状态回调
// GrabFood 在菜单同步状态变化时调用此端点推送状态通知
// 参考: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/menu-sync-state-webhook
func (c *ControllerV1) MenuSyncState(ctx context.Context, req *v1.MenuSyncStateReq) (res *v1.MenuSyncStateRes, err error) {
	// 调用 Service 处理菜单同步状态
	if err := service.Grab().HandleMenuSyncState(ctx, req.MenuSyncWebhookRequest); err != nil {
		g.Log().Errorf(ctx, "[Grab] HandleMenuSyncState failed: %v", err)
		// 记录错误但返回成功，避免 Grab 重试
	}

	// 返回 200 OK
	return &v1.MenuSyncStateRes{}, nil
}

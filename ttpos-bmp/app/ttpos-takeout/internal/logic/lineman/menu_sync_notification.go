// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// HandleMenuSyncNotification 处理 LINE MAN 菜单同步通知并写入日志
// 参数:
//   - ctx: 上下文
//   - req: 菜单同步通知请求
//
// 返回:
//   - error: 错误信息
func (s *sLineman) HandleMenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) error {
	if req == nil {
		return gerror.NewCode(gcode.CodeInvalidParameter, "请求参数不能为空")
	}

	status := strings.ToUpper(req.Status)
	success := false
	switch status {
	case "SUCCESS":
		success = true
	case "FAILED":
		success = false
	default:
		return gerror.NewCode(gcode.CodeInvalidParameter, "status 必须为 SUCCESS 或 FAILED")
	}

	err := service.ChannelMenu().LogMenuSync(
		ctx,
		req.StoreId,
		string(consts.ProviderLineman),
		string(consts.MenuSyncTypeNotify),
		req.MenuSyncRequestId,
		success,
		"",
		req.ErrorMsg,
	)
	if err != nil {
		return gerror.NewCode(gcode.CodeInternalError, err.Error())
	}

	return nil
}

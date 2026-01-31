// Package lineman 提供 LINE MAN 平台集成服务
package lineman

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// HandleTriggerSyncMenu 处理 LINE MAN 触发菜单同步请求
// 参数:
//   - ctx: 上下文
//   - req: 触发同步请求
//
// 返回:
//   - error: 错误信息
func (s *sLineman) HandleTriggerSyncMenu(ctx context.Context, req *v1.TriggerSyncMenuReq) error {
	if req == nil {
		return gerror.NewCode(gcode.CodeInvalidParameter, "请求参数不能为空")
	}

	// Step 1: 解析 storeId（在 Lineman 场景中，storeId 就是 shopUUID）
	shopUUID, err := strconv.ParseUint(req.StoreId, 10, 64)
	if err != nil {
		g.Log().Errorf(ctx, "解析 storeId 失败: %v", err)
		return gerror.NewCode(gcode.CodeNotFound, "Invalid store ID")
	}

	// Step 2: 记录到 menu_log（sync_type=NOTIFY, status=QUEUED）
	err = service.ChannelMenu().LogMenuSync(
		ctx,
		req.StoreId,
		string(consts.ProviderLineman),
		string(consts.MenuSyncTypeNotify),
		"",       // request_id 由 LogMenuSync 自动生成
		false,    // success = false，表示尚未完成
		"QUEUED", // sync_status
		"",       // error_msg
	)
	if err != nil {
		g.Log().Errorf(ctx, "记录菜单同步日志失败: %v", err)
		return gerror.NewCode(gcode.CodeInternalError, "Failed to log menu sync")
	}

	// Step 3: 调用 service.Lineman().SyncMenu() 触发同步
	err = service.Lineman().SyncMenu(ctx, shopUUID)
	if err != nil {
		g.Log().Errorf(ctx, "触发菜单同步失败: %v", err)
		// 同步失败不影响响应，因为已经记录了日志
		// 实际同步会异步执行，这里只是触发
	}

	return nil
}

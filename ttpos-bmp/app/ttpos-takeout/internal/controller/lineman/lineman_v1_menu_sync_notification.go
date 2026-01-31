package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// MenuSyncNotification 接收 LINE MAN 菜单同步结果通知
func (c *ControllerV1) MenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) (res *v1.MenuSyncNotificationRes, err error) {
	type menuSyncNotifier interface {
		HandleMenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) error
	}

	notifier, ok := service.Lineman().(menuSyncNotifier)
	if !ok {
		return &v1.MenuSyncNotificationRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    "500",
				Message: "Lineman 服务未实现菜单同步通知处理",
			},
		}, nil
	}

	err = notifier.HandleMenuSyncNotification(ctx, req)
	if err != nil {
		code := "500"
		if gerror.Code(err) == gcode.CodeInvalidParameter {
			code = "400"
		}
		return &v1.MenuSyncNotificationRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    code,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.MenuSyncNotificationRes{
		LinemanCommonResData: v1.LinemanCommonResData{
			Status:  "ok",
			Code:    "200",
			Message: "Menu sync notification received",
		},
	}, nil
}

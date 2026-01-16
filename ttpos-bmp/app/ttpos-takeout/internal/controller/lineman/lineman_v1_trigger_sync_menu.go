package lineman

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "ttpos-bmp/app/ttpos-takeout/api/lineman/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// TriggerSyncMenu 接收 LINE MAN 触发菜单同步请求
func (c *ControllerV1) TriggerSyncMenu(ctx context.Context, req *v1.TriggerSyncMenuReq) (res *v1.TriggerSyncMenuRes, err error) {
	// 使用类型断言检查 service 是否实现了该方法
	type triggerSyncMenuHandler interface {
		HandleTriggerSyncMenu(ctx context.Context, req *v1.TriggerSyncMenuReq) error
	}

	handler, ok := service.Lineman().(triggerSyncMenuHandler)
	if !ok {
		return &v1.TriggerSyncMenuRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    "500",
				Message: "Lineman 服务未实现触发同步处理",
			},
		}, nil
	}

	err = handler.HandleTriggerSyncMenu(ctx, req)
	if err != nil {
		code := "500"
		if gerror.Code(err) == gcode.CodeInvalidParameter {
			code = "400"
		} else if gerror.Code(err) == gcode.CodeNotFound {
			code = "404"
		}
		return &v1.TriggerSyncMenuRes{
			LinemanCommonResData: v1.LinemanCommonResData{
				Status:  "fail",
				Code:    code,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.TriggerSyncMenuRes{
		LinemanCommonResData: v1.LinemanCommonResData{
			Status:  "ok",
			Code:    "200",
			Message: "Trigger sync is successfully trigger.",
		},
	}, nil
}

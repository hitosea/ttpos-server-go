package hello

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"ttpos-bmp/app/ttpos-manager/api/hello/v1"
	rpc "ttpos-bmp/app/ttpos-manager/api/rpc/v1"
	"ttpos-bmp/app/ttpos-manager/api/rpc/v1/svc"
	"ttpos-bmp/app/ttpos-manager/internal/logic"
)

func (c *ControllerV1) GetSetting(ctx context.Context, req *v1.GetSettingReq) (res *v1.GetSettingRes, err error) {
	var setting *rpc.Setting
	setting, err = logic.SettingClient.GetSetting(ctx, &svc.GetReq{
		Key: req.Key,
	})
	if err != nil {
		err = gerror.Wrap(err, "获取设置失败")
		return
	}
	res = &v1.GetSettingRes{
		Setting: *setting,
	}
	return
}

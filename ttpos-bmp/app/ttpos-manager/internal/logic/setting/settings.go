package setting

import (
	"context"
	rpc "ttpos-bmp/app/ttpos-manager/api/manager"
	"ttpos-bmp/app/ttpos-manager/api/svc"
	"ttpos-bmp/app/ttpos-manager/internal/dao"
	"ttpos-bmp/app/ttpos-manager/internal/service"
)

var (
	Setting = new(sSetting)
)

type sSetting struct {
}

func init() {
	service.RegisterSetting(Setting)
}

func (s *sSetting) GetSetting(ctx context.Context, req *svc.GetReq) (setting *rpc.Setting, err error) {
	err = dao.Setting.Ctx(ctx).Fields(setting).Where(dao.Setting.Columns().Key, req.Key).Scan(&setting)
	return
}

// Package grab 提供 GrabFood API 集成的业务逻辑
package grab

import (
	"context"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
	grabclient "ttpos-bmp/app/ttpos-takeout/internal/client/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

var (
	// Grab Grab服务实例
	Grab = new(sGrab)
)

type sGrab struct{}

func init() {
	service.RegisterGrab(Grab)
}

// MustConf 获取 Grab 配置
func (s *sGrab) MustConf() *conf.Grab {
	return grabclient.MustConf()
}

// GetShopProviderCfg 查询门店第三方配置
func (s *sGrab) GetShopProviderCfg(ctx context.Context, req *grab.GetShopProviderCfgReq) (*grab.GetShopProviderCfgResp, error) {
	return service.ShopProviderCfg().GetShopProviderCfgResp(ctx, req)
}

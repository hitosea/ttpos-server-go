package grab

import (
	"context"

	api "ttpos-bmp/app/ttpos-takeout/api/grab"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

// GetShopProviderCfg 查询门店第三方配置
func (c *Controller) GetShopProviderCfg(ctx context.Context, req *api.GetShopProviderCfgReq) (res *api.GetShopProviderCfgResp, err error) {
	return service.Grab().GetShopProviderCfg(ctx, req)
}

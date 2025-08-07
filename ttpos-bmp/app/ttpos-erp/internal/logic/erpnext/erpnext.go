package erpnext

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
)

func getRpcUrlWithName(req *dto.ErpReq) string {
	return fmt.Sprintf("%s/%s", rpcApiUrl, req.DocType)
}

func GetClient(ctx context.Context) *gclient.Client {
	var c = g.Client()
	m := grpcx.Ctx.IncomingMap(ctx)
	if m.Contains("erp_site_code") {
		var site *entity.Site
		dao.Site.Ctx(ctx).Limit(1).Where(do.Site{}.SiteCode, m.GetVar("erp_site_code")).Scan(&site)
		if site != nil {
			c.SetPrefix(site.SiteUrl)
			c.SetHeader("Authorization", fmt.Sprintf("token %s:%s", site.ApiKey, site.ApiSecret))
		}
	} else {
		c.SetPrefix(g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.serviceUrl").String())
	}
	return c
}

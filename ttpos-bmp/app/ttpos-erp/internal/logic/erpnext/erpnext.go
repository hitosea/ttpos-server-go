package erpnext

import (
	"context"
	"fmt"
	"net/http"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
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
	c.Use(func(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
		resp, err = c.Next(r)
		if resp != nil {
			//TODO 增加开关
			resp.RawDump()
		}
		return resp, err
	})
	return c
}

func detectError(resp *gvar.Var) error {
	if resp == nil {
		return gerror.New("调用erpnext接口返回空")
	} else {
		if j, err := gjson.DecodeToJson(resp); err == nil {
			if j.Contains("exc_type") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erpnext接口返回异常: %v", j)
				return gerror.Newf("调用erpnext接口返回异常,exc_type:%s,exception:%s", j.Get("exc_type").String(), j.Get("exception").String())
			}
			if j.Contains("errors") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erpnext接口返回异常: %v", j)
				return gerror.Newf("调用erpnext接口返回异常,error:%s", j.Get("errors").String())
			}
		} else {
			g.Log().Error(gctx.GetInitCtx(), "调用erpnext接口返回解析异常: %v", err)
			return gerror.Newf("调用erpnext接口返回解析异常:%s", err.Error())
		}
	}

	return nil
}

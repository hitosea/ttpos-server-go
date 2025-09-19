package erpnext

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
)

const rpcApiUrl = "/api/v2/method"

type sRpc struct {
}

var Rpc = new(sRpc)

func init() {
	service.RegisterRpc(Rpc)
}

func (s *sRpc) Execute(ctx context.Context, req *erp.ErpReq, params interface{}) (rst *g.Var, err error) {
	rst = GetClient(ctx).PostVar(ctx, fmt.Sprintf("%s%s", getRpcUrlWithName(req), req.Method), params)
	err = detectError(rst)
	return
}

func getRpcUrlWithName(req *erp.ErpReq) (url string) {
	url = fmt.Sprintf("%s/%s", rpcApiUrl, req.DocType)
	return
}

func GetClient(ctx context.Context) *gclient.Client {
	var c = g.Client()
	m := grpcx.Ctx.IncomingMap(ctx)
	if m.Contains(consts.ContextSiteCode) {
		var site *entity.Site
		dao.Site.Ctx(ctx).Where(dao.Site.Columns().SiteCode, m.GetVar(consts.ContextSiteCode)).Limit(1).Scan(&site)
		if site != nil {
			c.SetPrefix(site.SiteUrl)
			c.SetHeader("Authorization", fmt.Sprintf("token %s:%s", site.ApiKey, site.ApiSecret))
		} else {
			g.Log().Errorf(ctx, "根据站点编码[%s]查询站点信息失败", m.GetVar(consts.ContextSiteCode))
		}
	} else {
		c.SetPrefix(g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.serviceUrl").String())
	}
	//替代指定用户调用服务
	if ctx.Value(consts.ContextFakeUser) != nil {
		// 从上下文获取用户信息
		cashier := &entity.ShopCashier{}
		err := dao.ShopCashier.Ctx(ctx).Where(dao.ShopCashier.Columns().CashierEmail, ctx.Value(consts.ContextFakeUser)).Limit(1).Scan(&cashier)
		if err != nil || cashier == nil {
			g.Log().Errorf(ctx, "根据收银员邮箱[%s]查询收银员信息失败", ctx.Value(consts.ContextFakeUser))
		} else {
			//取不到收银员apikey的话会写空，请求会异常(算特性）
			c.SetHeader("Authorization", fmt.Sprintf("token %s:%s", cashier.ApiKey, cashier.ApiSecret))
		}
	}

	c.Use(func(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
		resp, err = c.Next(r)
		if resp != nil && g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.dump").Bool() {
			resp.RawDump()
		}
		return resp, err
	})
	return c
}

func detectError(resp *gvar.Var) error {
	if resp == nil || resp.IsEmpty() {
		return gerror.New("调用erp接口返回空")
	} else {
		if j, err := gjson.DecodeToJson(resp); err == nil {
			if j.Contains("exc_type") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回异常: %v", j)
				return gerror.Newf("调用erp接口返回异常,exc_type:%s,exception:%s", j.Get("exc_type").String(), j.Get("exception").String())
			}
			if j.Contains("errors") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回异常: %v", j)
				errMsgList := make([]string, 0)
				for _, errItem := range j.GetJsons("errors") {
					if errItem.Contains("message") {
						errMsgList = append(errMsgList, errItem.Get("message").String())
					} else {
						errMsgList = append(errMsgList, errItem.Get("exception").String())
					}
				}
				return gerror.Newf("调用erp接口返回异常,error:%s", strings.Join(errMsgList, ";"))
			}
		} else {
			g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回解析异常: %v", err)
			return gerror.Wrapf(err, "调用erp接口返回解析异常")
		}
	}

	return nil
}

func SetFakeUser(ctx context.Context, userEmail string) context.Context {
	ctx = context.WithValue(ctx, consts.ContextFakeUser, userEmail)
	return ctx
}

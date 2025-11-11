package erpnext

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
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

func (s *sRpc) Execute(ctx context.Context, req *erp.ErpReq, params interface{}) (rst *gjson.Json, err error) {
	resp := GetClient(ctx).PostVar(ctx, fmt.Sprintf("%s%s", getRpcUrlWithName(req), req.Method), params)
	rst, err = detectError(resp)
	return
}

// GetSiteCode 获取站点编码
func (*sRpc) GetSiteCode(ctx context.Context) string {
	m := grpcx.Ctx.IncomingMap(ctx)
	if m.Contains(consts.ContextSiteCode) {
		return m.GetVar(consts.ContextSiteCode).String()
	}
	//默认返回UAT
	return consts.SiteCodeUat
}

func getRpcUrlWithName(req *erp.ErpReq) (url string) {
	url = fmt.Sprintf("%s/%s", rpcApiUrl, req.DocType)
	return
}

func GetClient(ctx context.Context) *gclient.Client {
	var c = g.Client()
	m := grpcx.Ctx.IncomingMap(ctx)
	if m.Contains(consts.ContextSiteCode) {
		serviceAuthorization, err := Rpc.GetAndProcessSiteAuthorization(ctx, m.GetVar(consts.ContextSiteCode).String())
		if err != nil {
			g.Log().Errorf(ctx, "获取站点授权信息失败: %v", err)
		} else {
			c.SetPrefix(serviceAuthorization.SiteUrl)
			c.SetHeader("Authorization", serviceAuthorization.Authorization)
		}
	} else {
		c.SetPrefix(g.Cfg().MustGet(gctx.GetInitCtx(), "app.erpnext.serviceUrl").String())
	}
	//替代指定用户调用服务
	if ctx.Value(consts.ContextFakeUser) != nil {
		cashierAuthorization, err := Rpc.GetAndProcessCashierAuthorization(ctx, ctx.Value(consts.ContextFakeUser).(string))
		if err != nil {
			g.Log().Errorf(ctx, "获取收银员授权信息失败: %v", err)
		} else {
			c.SetHeader("Authorization", cashierAuthorization)
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

func detectError(resp *gvar.Var) (*gjson.Json, error) {
	if resp == nil || resp.IsEmpty() {
		return nil, gerror.New("调用erp接口返回空")
	} else {
		if j, err := gjson.DecodeToJson(resp); err == nil {
			if j.Contains("exc_type") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回异常: %v", j)
				return nil, gerror.Newf("调用erp接口返回异常,exc_type:%s,exception:%s", j.Get("exc_type").String(), j.Get("exception").String())
			}
			if j.Contains("errors") {
				g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回异常: %v", j)
				errMsgList := make([]string, 0)
				for _, errItem := range j.GetJsons("errors") {
					if errItem.Contains("type") {
						errMsgList = append(errMsgList, errItem.Get("type").String())
					}
					if errItem.Contains("message") {
						errMsgList = append(errMsgList, errItem.Get("message").String())
					} else {
						errMsgList = append(errMsgList, errItem.Get("exception").String())
					}
				}
				return nil, gerror.Newf("调用erp接口返回异常:%s", strings.Join(errMsgList, ";"))
			}
			// 成功时返回解析后的JSON数据
			return j, nil
		} else {
			g.Log().Errorf(gctx.GetInitCtx(), "调用erp接口返回解析异常: %v", err)
			return nil, gerror.Wrapf(err, "调用erp接口返回解析异常")
		}
	}
}

func SetFakeUser(ctx context.Context, userEmail string) context.Context {
	ctx = context.WithValue(ctx, consts.ContextFakeUser, userEmail)
	return ctx
}

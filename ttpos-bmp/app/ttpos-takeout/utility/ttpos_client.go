package utility

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/consts"
)

const defaultTimeout = 10 * time.Second

// GetTtposClient 获取预配置的 TTPOS HTTP Client
// 自动设置 endpoint prefix、超时、ContentJson
// 当 app.ttpos-client.dump 配置为 true 时，自动打印请求/响应详情
func GetTtposClient(ctx context.Context) *gclient.Client {
	c := g.Client()

	// 设置 prefix
	ttposEndpoint := g.Cfg().MustGet(gctx.GetInitCtx(), "app.ttposEndpoint").String()
	if ttposEndpoint != "" {
		c.SetPrefix(ttposEndpoint)
	}

	// 设置超时和 Content-Type
	c.Timeout(defaultTimeout)
	c.ContentJson()

	// 添加 dump 中间件
	c.Use(func(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
		resp, err = c.Next(r)
		if resp != nil && g.Cfg().MustGet(gctx.GetInitCtx(), "app.ttpos-client.dump").Bool() {
			resp.RawDump()
		}
		return resp, err
	})

	return c
}

// GetTtposClientWithAuth 获取带认证头的 TTPOS HTTP Client
// identifier: 用于生成认证头的标识符（如 shopUUID）
func GetTtposClientWithAuth(ctx context.Context, identifier string) (*gclient.Client, error) {
	c := GetTtposClient(ctx)

	// 生成并设置认证头
	auth, err := GenerateTtposAuth(identifier)
	if err != nil {
		return nil, err
	}
	c.SetHeader(consts.TTPOS_HEADER_SECRET, auth)

	return c, nil
}

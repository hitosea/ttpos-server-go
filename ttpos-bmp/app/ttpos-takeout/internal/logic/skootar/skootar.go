package skootar

import (
	"fmt"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"ttpos-bmp/app/ttpos-takeout/internal/model/conf"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

var (
	Skootar = new(sSkootar)
	config  *conf.Skootar
)

type sSkootar struct {
}

func init() {
	service.RegisterSkootar(Skootar)
}

func (s *sSkootar) MustConf() *conf.Skootar {
	// 判断 config 是否为空
	if config == nil {
		config = &conf.Skootar{}
		ctx := gctx.New()
		// 从 Cfg 中读取 app.provider 中 name = skootar 的配置
		if err := g.Cfg().MustGet(gctx.New(), "app.provider.skootar").Scan(config); err != nil {
			g.Log().Fatal(ctx, err)
			panic(gerror.Newf("获取Skootar配置失败"))
		}
	}
	return config
}

// GetUrl 获取Skootar API URL
func (s *sSkootar) GetUrl(apiPath string) string {
	return fmt.Sprintf("%v%v", s.MustConf().Endpoint, apiPath)
}

// ReqBase 获取请求基础参数
func (s *sSkootar) ReqBase() skootar.ReqBase {
	return skootar.ReqBase{
		ApiKey:   s.MustConf().ApiKey,
		UserName: s.MustConf().UserName,
		Channel:  s.MustConf().Channel,
	}
}

// getCallBackAuth 获取回调Auth
func (*sSkootar) getCallBackAuth(shopRefNo string) string {
	if res, err := gmd5.EncryptString(shopRefNo + g.Cfg().MustGet(gctx.GetInitCtx(), "app.callbackSecret").String()); err == nil {
		return res
	} else {
		g.Log().Error(gctx.GetInitCtx(), "获取回调Auth失败", err)
	}
	return ""
}

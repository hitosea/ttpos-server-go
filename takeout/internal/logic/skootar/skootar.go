package skootar

import (
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"takeout/internal/model/conf"
	"takeout/internal/model/input/skootar"
	"takeout/internal/service"
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

func (s *sSkootar) GetUrl(apiPath string) string {
	return fmt.Sprintf("%v%v", s.MustConf().Endpoint, apiPath)
}

func (s *sSkootar) ReqBase() skootar.ReqBase {
	return skootar.ReqBase{
		ApiKey:   s.MustConf().ApiKey,
		UserName: s.MustConf().UserName,
		Channel:  s.MustConf().Channel,
	}
}

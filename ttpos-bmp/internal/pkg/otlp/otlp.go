package otlp

import (
	"context"

	"github.com/gogf/gf/contrib/trace/otlphttp/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

var (
	globalCtx = gctx.GetInitCtx()
	config    = &OtlpConfig{}
)

type OtlpConfig struct {
	ServiceName string `json:"serviceName"`
	Endpoint    string `json:"endpoint"`
	Path        string `json:"path"`
}

func init() {
	// 初始化调用链监控
	// 从配置文件中读取调用链监控配置
	err := g.Cfg().MustGet(globalCtx, "otlp").Scan(&config)
	if err != nil {
		g.Log().Warningf(globalCtx, "otlp config error: %v", err)
		return
	}
	g.Log().Infof(globalCtx, "otlp config: %v", config)
}

func InitOtlp() func(context.Context) {
	shutdown, err := otlphttp.Init(config.ServiceName, config.Endpoint, config.Path)
	if err != nil {
		g.Log().Warningf(globalCtx, "otlp init error: %v", err)
		return nil
	}
	g.Log().Infof(globalCtx, "otlp init success")
	return shutdown
}

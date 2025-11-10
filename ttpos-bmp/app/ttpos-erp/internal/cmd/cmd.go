package cmd

import (
	"context"
	"os"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/controller/callback"
	"ttpos-bmp/internal/pkg/otlp"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"ttpos-bmp/app/ttpos-erp/internal/controller/hello"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

var (
	Main = &gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 初始化OTLP
			shutdownOtlp := otlp.InitOtlp()
			defer shutdownOtlp(ctx)
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
				group.ALL("/callback", callback.NewV1())
			})

			s.Run()
			return nil
		},
	}

	ErpMigrate = &gcmd.Command{
		Name:  "migrate",
		Usage: "migrate --siteCode 1 --dirBase ./manifest/erp-migrate/v2.5",
		Brief: "执行ERP数据迁移，初始化自定义字段、客户和支付方式",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始执行ERP数据迁移...")

			// 初始化自定义字段
			g.Log().Infof(ctx, "正在初始化自定义字段... %v", parser)
			siteCode := parser.GetOpt("siteCode", "1").String()
			dirBase := parser.GetOpt("dirBase", "./manifest/erp-migrate/v2.5").String()
			ctx = grpcx.Ctx.NewIncoming(ctx, g.Map{
				consts.ContextSiteCode: siteCode,
			})

			if err := service.Setup().InitErpDocTypeWithDirname(ctx, dirBase); err != nil {
				g.Log().Error(ctx, "初始化DocType失败", err)
				return err
			}

			g.Log().Info(ctx, "ERP数据迁移执行完成!")
			// 操作完成后退出
			os.Exit(0)
			return nil
		},
	}
)

type cMigrate struct {
	g.Meta `name:"migrate" brief:"执行ERP数据迁移"`
}

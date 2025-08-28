package cmd

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/consts"

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
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}

	ErpMigrate = &gcmd.Command{
		Name:  "erp-migrate",
		Usage: "erp-migrate",
		Brief: "执行ERP数据迁移，初始化自定义字段、客户和支付方式",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始执行ERP数据迁移...")

			// 初始化自定义字段
			g.Log().Infof(ctx, "正在初始化自定义字段... %v", parser)
			siteCode := parser.GetOpt("siteCode", "1").String()
			dirBase := parser.GetOpt("dirBase", "./manifest/erp-migrate/v2.5").String()
			ctx = context.WithValue(ctx, consts.ContextSiteCode, siteCode)
			if err := service.Setup().InitCustomFields(ctx, dirBase); err != nil {
				g.Log().Error(ctx, "初始化自定义字段失败", err)
				return err
			}
			g.Log().Info(ctx, "自定义字段初始化完成")

			// 初始化客户
			g.Log().Info(ctx, "正在初始化客户...")
			if err := service.Setup().InitCustomers(ctx, dirBase); err != nil {
				g.Log().Error(ctx, "初始化客户失败", err)
				return err
			}
			g.Log().Info(ctx, "客户初始化完成")

			// 初始化支付方式
			g.Log().Info(ctx, "正在初始化支付方式...")
			if err := service.Setup().InitModeOfPayment(ctx, dirBase); err != nil {
				g.Log().Error(ctx, "初始化支付方式失败", err)
				return err
			}
			g.Log().Info(ctx, "支付方式初始化完成")

			g.Log().Info(ctx, "ERP数据迁移执行完成!")
			return nil
		},
	}
)

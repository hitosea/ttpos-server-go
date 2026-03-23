package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"ttpos-bmp/app/ttpos-erp/internal/boot"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/controller/callback"
	"ttpos-bmp/app/ttpos-erp/internal/logic/erpnext"
	"ttpos-bmp/app/ttpos-erp/internal/middleware"
	"ttpos-bmp/internal/pkg/otlp"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
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
			// 初始化服务器组件（RPC、Consumer等）
			boot.InitServer(ctx)
			// 初始化OTLP
			shutdownOtlp := otlp.InitOtlp()
			defer shutdownOtlp(ctx)
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					hello.NewV1(),
				)
				// callback 路由（API path 已含 /callback 前缀，此处不重复）
				group.Group("/", func(group *ghttp.RouterGroup) {
					group.Middleware(middleware.WebhookAuth)
					group.Bind(callback.NewV1())
				})
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

	ErpAllMigrate = &gcmd.Command{
		Name:  "migrate-all",
		Usage: "migrate-all --siteCode 1 --dirBase ./manifest/erp-migrate",
		Brief: "遍历所有版本目录并执行ERP数据迁移",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始执行ERP全量迁移...")

			siteCode := parser.GetOpt("siteCode", "1").String()
			baseDir := parser.GetOpt("dirBase", "./manifest/erp-migrate").String()
			ctx = grpcx.Ctx.NewIncoming(ctx, g.Map{
				consts.ContextSiteCode: siteCode,
			})

			entries, readErr := os.ReadDir(baseDir)
			if readErr != nil {
				g.Log().Error(ctx, "读取迁移目录失败", readErr)
				return readErr
			}

			var dirs []string
			for _, e := range entries {
				if e.IsDir() {
					dirs = append(dirs, e.Name())
				}
			}
			sort.Strings(dirs)

			for _, d := range dirs {
				dirPath := fmt.Sprintf("%s/%s", baseDir, d)
				g.Log().Infof(ctx, "执行版本目录: %s", dirPath)
				if err := service.Setup().InitErpDocTypeWithDirname(ctx, dirPath); err != nil {
					g.Log().Errorf(ctx, "迁移失败: %s", err)
					return err
				}
			}

			g.Log().Info(ctx, "ERP全量迁移执行完成!")
			os.Exit(0)
			return nil
		},
	}

	// ErpSyncPrintFormat 同步 ERP 打印格式到本地
	ErpSyncPrintFormat = &gcmd.Command{
		Name:  "sync-print-format",
		Usage: "sync-print-format --siteCode 1 [--prefix Wallace] [--dirBase ./manifest/printformat/html/]",
		Brief: "从 ERP 同步打印格式到本地 JSON 文件",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始同步打印格式...")

			// 解析参数
			siteCode := parser.GetOpt("siteCode", "").String()
			prefix := parser.GetOpt("prefix", "Wallace").String()
			dirBase := parser.GetOpt("dirBase", "./manifest/printformat/html/").String()

			// 参数校验
			if siteCode == "" {
				g.Log().Error(ctx, "siteCode 参数必填")
				return gerror.New("siteCode 参数必填，请使用 --siteCode 指定站点编码")
			}

			// 设置 context，传递 siteCode
			ctx = grpcx.Ctx.NewIncoming(ctx, g.Map{
				consts.ContextSiteCode: siteCode,
			})

			g.Log().Infof(ctx, "站点: %s, 前缀: %s, 输出目录: %s", siteCode, prefix, dirBase)

			// 调用同步方法
			result, err := erpnext.PrintFormat.SyncToLocal(ctx, dirBase, prefix)
			if err != nil {
				g.Log().Error(ctx, "同步打印格式失败", err)
				return err
			}

			// 输出结果摘要
			g.Log().Info(ctx, "")
			g.Log().Info(ctx, "同步完成!")
			g.Log().Infof(ctx, "  成功: %d", result.Success)
			g.Log().Infof(ctx, "  失败: %d", result.Failed)
			g.Log().Infof(ctx, "  总计: %d", result.Total)

			if len(result.Errors) > 0 {
				g.Log().Info(ctx, "")
				g.Log().Info(ctx, "失败详情:")
				for _, errMsg := range result.Errors {
					g.Log().Infof(ctx, "  - %s", errMsg)
				}
			}

			os.Exit(0)
			return nil
		},
	}
)

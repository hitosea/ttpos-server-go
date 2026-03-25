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
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

// resolveSiteCodes 解析 siteCode 参数，"all" 时查库返回所有站点编码
func resolveSiteCodes(ctx context.Context, siteCodeOpt string) ([]string, error) {
	if siteCodeOpt != "all" {
		return []string{siteCodeOpt}, nil
	}
	var sites []*entity.Site
	if err := dao.Site.Ctx(ctx).Scan(&sites); err != nil {
		return nil, gerror.Wrap(err, "查询所有站点失败")
	}
	codes := make([]string, 0, len(sites))
	for _, s := range sites {
		if s.SiteCode != "" {
			codes = append(codes, s.SiteCode)
		}
	}
	if len(codes) == 0 {
		return nil, gerror.New("erp_site 表中无可用站点")
	}
	g.Log().Infof(ctx, "共发现 %d 个站点: %v", len(codes), codes)
	return codes, nil
}

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
		Usage: "migrate --siteCode all --dirBase ./manifest/erp-migrate/v2.5",
		Brief: "执行ERP数据迁移，支持 --siteCode all 对所有站点执行",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始执行ERP数据迁移...")

			siteCodeOpt := parser.GetOpt("siteCode", "all").String()
			dirBase := parser.GetOpt("dirBase", "./manifest/erp-migrate/v2.5").String()

			siteCodes, err := resolveSiteCodes(ctx, siteCodeOpt)
			if err != nil {
				return err
			}

			for _, sc := range siteCodes {
				g.Log().Infof(ctx, "========== 站点 [%s] 开始迁移 ==========", sc)
				siteCtx := grpcx.Ctx.NewIncoming(ctx, g.Map{
					consts.ContextSiteCode: sc,
				})
				if err := service.Setup().InitErpDocTypeWithDirname(siteCtx, dirBase); err != nil {
					g.Log().Errorf(ctx, "站点 [%s] 迁移失败: %v", sc, err)
					return err
				}
				g.Log().Infof(ctx, "========== 站点 [%s] 迁移完成 ==========", sc)
			}

			g.Log().Info(ctx, "ERP数据迁移执行完成!")
			os.Exit(0)
			return nil
		},
	}

	ErpAllMigrate = &gcmd.Command{
		Name:  "migrate-all",
		Usage: "migrate-all --siteCode all --dirBase ./manifest/erp-migrate",
		Brief: "遍历所有版本目录并执行ERP数据迁移，支持 --siteCode all",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			g.Log().Info(ctx, "开始执行ERP全量迁移...")

			siteCodeOpt := parser.GetOpt("siteCode", "all").String()
			baseDir := parser.GetOpt("dirBase", "./manifest/erp-migrate").String()

			siteCodes, err := resolveSiteCodes(ctx, siteCodeOpt)
			if err != nil {
				return err
			}

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

			for _, sc := range siteCodes {
				g.Log().Infof(ctx, "========== 站点 [%s] 开始全量迁移 ==========", sc)
				siteCtx := grpcx.Ctx.NewIncoming(ctx, g.Map{
					consts.ContextSiteCode: sc,
				})
				for _, d := range dirs {
					dirPath := fmt.Sprintf("%s/%s", baseDir, d)
					g.Log().Infof(ctx, "站点 [%s] 执行版本目录: %s", sc, dirPath)
					if err := service.Setup().InitErpDocTypeWithDirname(siteCtx, dirPath); err != nil {
						g.Log().Errorf(ctx, "站点 [%s] 迁移失败: %s", sc, err)
						return err
					}
				}
				g.Log().Infof(ctx, "========== 站点 [%s] 全量迁移完成 ==========", sc)
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

package main

import (
	"os"

	_ "ttpos-bmp/app/ttpos-erp/internal/packed"

	_ "ttpos-bmp/app/ttpos-erp/internal/boot"
	_ "ttpos-bmp/app/ttpos-erp/internal/logic"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-erp/internal/cmd"
)

func main() {
	ctx := gctx.GetInitCtx()

	// 获取命令行参数（跳过程序名称）
	args := os.Args[1:]

	// 根据第一个参数判断执行哪个命令
	if len(args) > 0 {
		commandName := args[0]
		switch commandName {
		case "migrate":
			// 执行迁移命令，不启动 HTTP 服务器
			// 重新构造命令行参数，移除命令名称，保留选项参数
			os.Args = append([]string{os.Args[0]}, args[1:]...)
			cmd.ErpMigrate.Run(ctx)
			return
		case "migrate-all":
			// 执行全量迁移命令，不启动 HTTP 服务器
			// 重新构造命令行参数，移除命令名称，保留选项参数
			os.Args = append([]string{os.Args[0]}, args[1:]...)
			cmd.ErpAllMigrate.Run(ctx)
			return
		}
	}

	// 默认执行 Main 命令（HTTP 服务器）
	cmd.Main.Run(ctx)
}

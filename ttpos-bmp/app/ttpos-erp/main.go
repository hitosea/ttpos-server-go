package main

import (
	_ "ttpos-bmp/app/ttpos-erp/internal/packed"

	_ "ttpos-bmp/app/ttpos-erp/internal/boot"
	_ "ttpos-bmp/app/ttpos-erp/internal/logic"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-erp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

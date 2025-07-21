package main

import (
	_ "ttpos-bmp/app/ttpos-erp/internal/packed"

	_ "ttpos-bmp/app/ttpos-erp/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-erp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

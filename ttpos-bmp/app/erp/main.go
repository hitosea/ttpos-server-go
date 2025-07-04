package main

import (
	_ "ttpos-bmp/app/erp/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/erp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

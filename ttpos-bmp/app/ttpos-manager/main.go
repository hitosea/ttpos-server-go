package main

import (
	_ "ttpos-bmp/app/ttpos-manager/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-manager/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

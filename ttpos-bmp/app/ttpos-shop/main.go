package main

import (
	_ "ttpos-bmp/app/ttpos-shop/internal/packed"

	//_ "ttpos-bmp/app/ttpos-shop/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-shop/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

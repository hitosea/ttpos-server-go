package main

import (
	_ "ttpos-bmp/app/ttpos-message/internal/packed"

	_ "ttpos-bmp/app/ttpos-message/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-message/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

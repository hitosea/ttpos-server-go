package main

import (
	_ "ttpos-bmp/app/ttpos-manager/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	_ "ttpos-bmp/app/ttpos-manager/internal/logic"

	_ "ttpos-bmp/app/ttpos-manager/internal/boot"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-manager/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

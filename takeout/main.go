package main

import (
	_ "takeout/internal/packed"

	_ "takeout/internal/logic"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"takeout/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

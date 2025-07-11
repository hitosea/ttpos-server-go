package main

import (
	_ "takeout/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"takeout/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

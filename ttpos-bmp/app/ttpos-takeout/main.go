package main

import (
	_ "ttpos-bmp/app/ttpos-takeout/internal/packed"

	_ "ttpos-bmp/app/ttpos-takeout/internal/logic"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	_ "ttpos-bmp/app/ttpos-takeout/internal/boot"

	"github.com/gogf/gf/v2/os/gctx"

	"ttpos-bmp/app/ttpos-takeout/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

package boot

import (
	"context"

	"ttpos-bmp/internal/pkg/cache"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/os/gctx"
)

var (
	ctx = gctx.GetInitCtx()
)

func init() {
	// 基础初始化（不启动服务器）
	// 指定应用类型为 ERP，避免跨应用 ID 冲突
	uuid.InitIdGenerator(ctx, uuid.AppTypeERP)
	// 设置缓存
	cache.SetAdapter(ctx)
}

// InitServer 初始化服务器相关组件（RPC、Consumer等）
// 只在需要启动服务器的命令中调用（如 Main 命令）
func InitServer(ctx context.Context) {
	InitRpc(ctx)
	InitConsumer(ctx)
}

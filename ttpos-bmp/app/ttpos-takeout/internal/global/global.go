package global

import (
	"context"
	"ttpos-bmp/internal/pkg/cache"
)

// 在这里可以配置一些全局公用的变量

func Init(ctx context.Context) {
	//设置缓存
	cache.SetAdapter(ctx)
}

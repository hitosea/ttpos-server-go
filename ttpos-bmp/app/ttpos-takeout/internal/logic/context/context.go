package context

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/internal/model"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

var (
	Context = new(sContextService)
)

type sContextService struct{}

func init() {
	// 注册服务
	service.RegisterContextService(Context)
}

//func (s *sContextService) Init(ctx context.Context, ctxData *model.Context) context.Context {
//	//TODO 完善逻辑
//	return context.WithValue(ctx, model.ContextKey, ctxData)
//}

func (s *sContextService) Get(ctx context.Context) *model.Context {
	value := ctx.Value(model.ContextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*model.Context); ok {
		return localCtx
	}

	return nil
}

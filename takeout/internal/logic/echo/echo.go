package echo

import (
	"context"
	"fmt"
	"takeout/internal/dao"
	"takeout/internal/model"
	"takeout/internal/model/entity"
	"takeout/internal/service"
)

type sEcho struct {
}

func init() {
	// 注册服务
	service.RegisterEcho(New())
}

func New() *sEcho {
	return &sEcho{}
}

func (s *sEcho) Msg(ctx context.Context, in *model.EchoMsgInput) (out *model.EcoMsgOutput, err error) {
	var echoEntity = &entity.Echo{}
	dao.Echo.Ctx(ctx).Scan(echoEntity)
	out = &model.EcoMsgOutput{
		Message: fmt.Sprintf("Hello %v: %v", in.Message, echoEntity.Msg),
	}
	return
}

package echo

import (
	"context"
	"fmt"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
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

func (s *sEcho) Msg(ctx context.Context, in *dto.EchoMsgInput) (out *dto.EcoMsgOutput, err error) {
	g.Log().Infof(ctx, "获取随路数据:%v", grpcx.Ctx.IncomingMap(ctx))
	var echoEntity = &entity.Echo{}
	dao.Echo.Ctx(ctx).Scan(echoEntity)
	out = &dto.EcoMsgOutput{
		Message: fmt.Sprintf("Hello %v: %v", in.Message, echoEntity.Msg),
	}
	return
}

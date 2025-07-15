package echo

import (
	"context"
	"fmt"
	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/os/glog"
	"takeout/internal/dao"
	"takeout/internal/model/entity"
	"takeout/internal/model/input"
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

func (s *sEcho) Msg(ctx context.Context, in *input.EchoMsgInput) (out *input.EcoMsgOutput, err error) {
	glog.Infof(ctx, "获取随路数据:%v", grpcx.Ctx.IncomingMap(ctx))
	var echoEntity = &entity.Echo{}
	dao.Echo.Ctx(ctx).Scan(echoEntity)
	out = &input.EcoMsgOutput{
		Message: fmt.Sprintf("Hello %v: %v", in.Message, echoEntity.Msg),
	}
	return
}

package order

import (
	"context"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"

	api "ttpos-bmp/app/ttpos-takeout/api/order"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/service"
)

type Controller struct {
	api.UnimplementedOrderServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterOrderServiceServer(s.Server, &Controller{})
}

// GetOrderInfo 获取订单信息
func (c *Controller) GetOrderInfo(ctx context.Context, req *api.GetOrderInfoReq) (*takeout.ApiResponse, error) {
	res, err := service.Order().GetOrderInfo(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 将 res 转换为 anypb.Any
	dataAny, err := anypb.New(res)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Debugf(ctx, "GetOrderInfo success: %+v", res)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}


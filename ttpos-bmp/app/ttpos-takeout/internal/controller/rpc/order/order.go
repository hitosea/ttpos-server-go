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

// PrepareOrder 准备订单（接受/拒绝）
func (c *Controller) PrepareOrder(ctx context.Context, req *api.PrepareOrderReq) (*takeout.ApiResponse, error) {
	// 参数验证
	if req.TakeoutOrderUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "订单UUID不能为空",
		}, nil
	}

	if req.ToState != string(consts.OrderPrepareStateAccepted) && req.ToState != string(consts.OrderPrepareStateRejected) {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "目标状态必须为Accepted或Rejected",
		}, nil
	}

	// 调用 Service 层
	res, err := service.Order().PrepareOrder(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "准备订单失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 将 res 转换为 anypb.Any
	dataAny, err := anypb.New(res)
	if err != nil {
		g.Log().Errorf(ctx, "序列化响应失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	g.Log().Infof(ctx, "准备订单成功: orderUuid=%s, toState=%s", req.TakeoutOrderUuid, req.ToState)
	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

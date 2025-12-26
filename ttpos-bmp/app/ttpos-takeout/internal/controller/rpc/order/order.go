package order

import (
	"context"
	"errors"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/grab/grabfood-api-sdk-go"
	"google.golang.org/protobuf/types/known/anypb"

	"ttpos-bmp/app/ttpos-takeout/api/grab"
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
		// 如果是 grab 订单，返回异常的详情
		var apiErr *grabfood.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			dataAny, _ := anypb.New(&grab.ErrorDetail{
				Body: string(apiErr.Body()),
			})
			return &takeout.ApiResponse{
				Code:    string(consts.CodeServiceError),
				Message: err.Error(),
				Data:    dataAny,
			}, nil

		}
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

// MarkOrderReady 标记订单准备完成
func (c *Controller) MarkOrderReady(ctx context.Context, req *api.MarkOrderReadyReq) (*takeout.ApiResponse, error) {
	// 1. 参数验证
	if req.TakeoutOrderUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "takeout_order_uuid 不能为空",
		}, nil
	}

	// 2. 记录请求日志
	g.Log().Infof(ctx, "接收到 MarkOrderReady 请求: order_uuid=%s, request_id=%s",
		req.TakeoutOrderUuid, req.RequestId)

	// 3. 调用 Service 层
	orderUuid, err := service.Order().MarkOrderReady(ctx, req.TakeoutOrderUuid, req.RequestId)
	if err != nil {
		g.Log().Errorf(ctx, "MarkOrderReady 失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 4. 构建响应
	resp := &api.MarkOrderReadyResp{
		OrderUuid: orderUuid,
	}

	// 5. 将 resp 转换为 anypb.Any
	dataAny, err := anypb.New(resp)
	if err != nil {
		g.Log().Errorf(ctx, "序列化响应失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	// 6. 记录成功日志
	g.Log().Infof(ctx, "MarkOrderReady 成功: order_uuid=%s", orderUuid)

	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

// CancelOrder 取消订单（包含预检查机制）
func (c *Controller) CancelOrder(ctx context.Context, req *api.CancelOrderReq) (*takeout.ApiResponse, error) {
	// 1. 参数验证
	if req.TakeoutOrderUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "订单UUID不能为空",
		}, nil
	}

	// 2. 记录请求日志
	g.Log().Infof(ctx, "接收到 CancelOrder 请求: order_uuid=%s, cancel_code=%s, request_id=%s",
		req.TakeoutOrderUuid, req.CancelCode, req.RequestId)

	// 3. 调用 Service 层
	res, err := service.Order().CancelOrder(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "CancelOrder 失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 4. 构建响应
	dataAny, err := anypb.New(res)
	if err != nil {
		g.Log().Errorf(ctx, "序列化响应失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	// 6. 记录成功日志
	g.Log().Infof(ctx, "CancelOrder 成功: order_uuid=%s", res.OrderUuid)

	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: "订单已成功取消",
		Data:    dataAny,
	}, nil
}

// CheckOrderCancelable 检查订单是否可取消
func (c *Controller) CheckOrderCancelable(ctx context.Context, req *api.CheckOrderCancelableReq) (*takeout.ApiResponse, error) {
	// 1. 参数验证
	if req.TakeoutOrderUuid == "" {
		return &takeout.ApiResponse{
			Code:    string(consts.CodeInvalidParam),
			Message: "订单UUID不能为空",
		}, nil
	}

	// 2. 记录请求日志
	g.Log().Infof(ctx, "接收到 CheckOrderCancelable 请求: order_uuid=%s, request_id=%s",
		req.TakeoutOrderUuid, req.RequestId)

	// 3. 调用 Service 层
	res, err := service.Order().CheckOrderCancelable(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "CheckOrderCancelable 失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeServiceError),
			Message: err.Error(),
		}, nil
	}

	// 4. 构建响应
	dataAny, err := anypb.New(res)
	if err != nil {
		g.Log().Errorf(ctx, "序列化响应失败: %v", err)
		return &takeout.ApiResponse{
			Code:    string(consts.CodeSerializeError),
			Message: consts.MsgSerializeFailed,
		}, nil
	}

	// 5. 记录成功日志
	g.Log().Infof(ctx, "CheckOrderCancelable 成功: order_uuid=%s, can_cancel=%t", res.OrderUuid, res.CanCancel)

	return &takeout.ApiResponse{
		Code:    string(consts.CodeSuccess),
		Message: consts.MsgSuccess,
		Data:    dataAny,
	}, nil
}

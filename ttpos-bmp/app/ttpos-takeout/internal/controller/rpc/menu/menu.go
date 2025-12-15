package menu

import (
	"context"
	api "ttpos-bmp/app/ttpos-takeout/api/menu"
	"ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/types/known/anypb"
)

type Controller struct {
	api.UnimplementedMenuServiceServer
}

func Register(s *grpcx.GrpcServer) {
	api.RegisterMenuServiceServer(s.Server, &Controller{})
}

func (c *Controller) GetMenuSnapshot(ctx context.Context, req *api.GetMenuSnapshotReq) (*takeout.ApiResponse, error) {
	res, err := service.ChannelMenu().GetMenuSnapshot(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: err.Error(),
		}, nil
	}

	// 将 res 转换为 anypb.Any
	dataAny, err := anypb.New(res)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: "序列化响应数据失败",
		}, nil
	}

	g.Log().Debugf(ctx, "查询菜单快照成功:%+v", res)
	return &takeout.ApiResponse{
		Code:    "0",
		Message: "success",
		Data:    dataAny,
	}, nil
}

func (c *Controller) SaveMenuSnapshot(ctx context.Context, req *api.SaveMenuSnapshotReq) (*takeout.ApiResponse, error) {
	res, err := service.ChannelMenu().SaveMenuSnapshot(ctx, req)
	if err != nil {
		return &takeout.ApiResponse{
			Code:    "5001",
			Message: err.Error(),
		}, nil
	}

	// SaveMenuSnapshotResp 是空结构体，可以只返回 code 和 message
	// 如果需要，也可以将 res 转换为 anypb.Any
	var dataAny *anypb.Any
	if res != nil {
		dataAny, _ = anypb.New(res)
	}

	g.Log().Debugf(ctx, "保存菜单快照成功:%+v", res)
	return &takeout.ApiResponse{
		Code:    "0",
		Message: "success",
		Data:    dataAny,
	}, nil
}

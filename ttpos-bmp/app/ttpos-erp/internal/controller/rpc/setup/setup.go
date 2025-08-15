package setup

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 设置服务控制器
type Controller struct {
	setup.UnimplementedSetupServiceServer
}

// Register 注册设置服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	setup.RegisterSetupServiceServer(s.Server, &Controller{})
}

// InitShop 初始化商店
// 参数：ctx 上下文，req 初始化商店请求
// 返回：响应信息和错误
func (c *Controller) InitShop(ctx context.Context, req *setup.InitShopReq) (*api.ResponseInfo, error) {
	// 参数校验
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}
	// 调用服务层初始化商店
	branchName, err := service.Setup().InitShop(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("初始化商店成功", &setup.InitShopResp{BranchName: branchName}), nil
}

// CreatePosUser 创建POS用户
// 参数：ctx 上下文，req 创建用户请求
// 返回：响应信息和错误
func (c *Controller) CreatePosUser(ctx context.Context, req *setup.CreatePosUserReq) (*api.ResponseInfo, error) {
	// 参数校验
	if req == nil {
		return rpc.ApiError("请求参数不能为空"), nil
	}
	// 调用服务层创建用户
	userEmail, err := service.Setup().CreateUser(ctx, &erp.CreateUserInp{
		UserEmail: req.UserEmail,
		FirstName: req.UserName,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("创建POS用户成功", &setup.CreatePosUserResp{
		UserEmail: userEmail,
		UserName:  req.UserName,
	}), nil
}

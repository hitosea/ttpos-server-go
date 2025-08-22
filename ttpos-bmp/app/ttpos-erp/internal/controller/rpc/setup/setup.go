package setup

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	setup2 "ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
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
	resp, err := service.Setup().InitShop(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("初始化商店成功", resp), nil
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
	err := service.Setup().CreateUser(ctx, &setup2.CreateUserInp{
		UserEmail: req.UserEmail,
		FirstName: req.UserName,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("创建POS用户成功", &setup.CreatePosUserResp{
		UserEmail: req.UserEmail,
		UserName:  req.UserName,
	}), nil
}

// CreateDefaultPosProfile 创建默认POS配置文件
// 参数：ctx 上下文，req 创建默认POS配置文件请求
// 返回：响应信息和错误
func (c *Controller) CreateDefaultPosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateCreateDefaultPosProfileReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层创建POS配置文件
	resp, err := service.Setup().CreateDefaultPosProfile(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("创建默认POS配置文件成功", &setup.CreateDefaultPosProfileResp{
		Name: resp,
	}), nil
}

// validateCreateDefaultPosProfileReq 验证创建默认POS配置文件请求参数
func (c *Controller) validateCreateDefaultPosProfileReq(req *setup.CreateDefaultPosProfileReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.Name) == "" {
		return gerror.New("POS配置文件名称不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写编码不能为空")
	}
	if strings.TrimSpace(req.Branch) == "" {
		return gerror.New("分店名称不能为空")
	}
	return nil
}

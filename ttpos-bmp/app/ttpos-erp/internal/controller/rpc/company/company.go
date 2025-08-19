package company

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 公司服务控制器
type Controller struct {
	company.UnimplementedCompanyServiceServer
}

// Register 注册公司服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	company.RegisterCompanyServiceServer(s.Server, &Controller{})
}

// GetCompanyList 获取公司列表
// 参数：ctx 上下文，req 获取公司列表请求
// 返回：响应信息和错误
func (c *Controller) GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (*api.ResponseInfo, error) {

	// 调用服务层获取数据
	resp, err := service.Company().GetCompanyList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取公司列表成功", resp), nil
}

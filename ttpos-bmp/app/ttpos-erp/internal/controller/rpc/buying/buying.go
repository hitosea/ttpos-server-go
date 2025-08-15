package buying

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
)

type Controller struct {
	buying.UnimplementedBuyingServiceServer
}

func Register(s *grpcx.GrpcServer) {
	buying.RegisterBuyingServiceServer(s.Server, &Controller{})
}

// GetSupplierList 获取供应商列表
// 参数：ctx 上下文，req 获取供应商列表请求
// 返回：响应信息和错误
func (c *Controller) GetSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetSupplierListReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	resp, err := service.Buying().GetSupplierList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("查询供应商列表成功", resp), nil
}

// validateGetSupplierListReq 验证获取供应商列表请求参数
func (c *Controller) validateGetSupplierListReq(req *buying.GetSupplierListReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司缩写编码不能为空")
	}
	return nil
}

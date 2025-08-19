package warehouse

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 仓库服务控制器
type Controller struct {
	warehouse.UnimplementedWarehouseServiceServer
}

// Register 注册仓库服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	warehouse.RegisterWarehouseServiceServer(s.Server, &Controller{})
}

// CreateWarehouse 创建仓库
// 参数：ctx 上下文，req 仓库信息
// 返回：响应信息和错误
func (c *Controller) CreateWarehouse(ctx context.Context, req *warehouse.WarehouseInfo) (*api.ResponseInfo, error) {
	// 调用服务层创建仓库
	warehouseName, err := service.Warehouse().CreateWarehouse(ctx, &setup.CreateWarehouseInp{
		Company:     req.Company,
		CompanyAbbr: req.CompanyAbbr,
		Branch:      req.Branch,
		AliasName:   req.AliasName,
		WhType:      req.WarehouseType,
	})
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("创建仓库成功", &warehouse.WarehouseInfo{
		Company:       req.Company,
		CompanyAbbr:   req.CompanyAbbr,
		Branch:        req.Branch,
		AliasName:     warehouseName,
		WarehouseType: req.WarehouseType,
	}), nil
}

// GetWarehouseList 获取仓库列表
// 参数：ctx 上下文，req 获取仓库列表请求
// 返回：响应信息和错误
func (c *Controller) GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (*api.ResponseInfo, error) {
	// 调用服务层获取数据
	dataList, err := service.Warehouse().GetWarehouseList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取仓库列表成功", dataList), nil
}

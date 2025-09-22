package warehouse

import (
	"context"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/warehouse"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/errors/gerror"
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
// 参数：
//   - ctx: 上下文对象
//   - req: 仓库信息请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) CreateWarehouse(ctx context.Context, req *warehouse.WarehouseInfo) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateCreateWarehouseReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

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
		Name:          warehouseName + " - " + req.Branch,
		WarehouseType: req.WarehouseType,
	}), nil
}

// validateCreateWarehouseReq 验证创建仓库请求参数
func (c *Controller) validateCreateWarehouseReq(req *warehouse.WarehouseInfo) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司简称不能为空")
	}
	if strings.TrimSpace(req.Branch) == "" {
		return gerror.New("分店名称不能为空")
	}
	if strings.TrimSpace(req.AliasName) == "" {
		return gerror.New("仓库别名不能为空")
	}
	if strings.TrimSpace(req.WarehouseType) == "" {
		return gerror.New("仓库类型不能为空")
	}
	return nil
}

// GetWarehouseList 获取仓库列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取仓库列表请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) GetWarehouseList(ctx context.Context, req *warehouse.GetWarehouseListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetWarehouseListReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	dataList, err := service.Warehouse().GetWarehouseList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取仓库列表成功", dataList), nil
}

// validateGetWarehouseListReq 验证获取仓库列表请求参数
func (c *Controller) validateGetWarehouseListReq(req *warehouse.GetWarehouseListReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.CompanyAbbr) == "" {
		return gerror.New("公司简称不能为空")
	}
	return nil
}

// GetWarehouse 获取仓库详情
// 参数：
//   - ctx: 上下文对象
//   - req: 获取仓库详情请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) GetWarehouse(ctx context.Context, req *warehouse.GetWarehouseReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetWarehouseReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	warehouseData, err := service.Warehouse().GetWarehouse(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取仓库详情成功", warehouseData), nil
}

// UpdateWarehouse 更新仓库信息
// 参数：
//   - ctx: 上下文对象
//   - req: 更新仓库请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) UpdateWarehouse(ctx context.Context, req *warehouse.UpdateWarehouseReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateUpdateWarehouseReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层更新数据
	updatedData, err := service.Warehouse().UpdateWarehouse(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("更新仓库成功", updatedData), nil
}

// DeleteWarehouse 删除仓库
// 参数：
//   - ctx: 上下文对象
//   - req: 删除仓库请求参数
//
// 返回：
//   - *api.ResponseInfo: 响应信息
//   - error: 错误信息
func (c *Controller) DeleteWarehouse(ctx context.Context, req *warehouse.DeleteWarehouseReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateDeleteWarehouseReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层删除数据
	result, err := service.Warehouse().DeleteWarehouse(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData(result.Message, result), nil
}

// validateGetWarehouseReq 验证获取仓库详情请求参数
func (c *Controller) validateGetWarehouseReq(req *warehouse.GetWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.Name) == "" {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

// validateUpdateWarehouseReq 验证更新仓库请求参数
func (c *Controller) validateUpdateWarehouseReq(req *warehouse.UpdateWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if req.Warehouse == nil {
		return gerror.New("仓库信息不能为空")
	}
	if strings.TrimSpace(req.Warehouse.Name) == "" {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

// validateDeleteWarehouseReq 验证删除仓库请求参数
func (c *Controller) validateDeleteWarehouseReq(req *warehouse.DeleteWarehouseReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.Name) == "" {
		return gerror.New("仓库名称不能为空")
	}
	return nil
}

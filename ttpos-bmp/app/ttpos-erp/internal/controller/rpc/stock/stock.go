package stock

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api"
	"ttpos-bmp/app/ttpos-erp/api/stock"
	"ttpos-bmp/app/ttpos-erp/internal/controller/rpc"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
)

// Controller 库存服务控制器
type Controller struct {
	stock.UnimplementedStockServiceServer
}

// Register 注册库存服务到gRPC服务器
func Register(s *grpcx.GrpcServer) {
	stock.RegisterStockServiceServer(s.Server, &Controller{})
}

func (*Controller) SaveMaterialRequest(ctx context.Context, req *stock.SaveMaterialRequestReq) (res *api.ResponseInfo, err error) {
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if len(req.Branch) == 0 {
		return rpc.ApiError("分公司不能为空"), nil
	}
	if len(req.Items) == 0 {
		return rpc.ApiError("物品列表不能为空"), nil
	}
	if req.RequiredBy == 0 {
		return rpc.ApiError("采购日期不能为空"), nil
	}
	if req.TransactionDate == 0 {
		return rpc.ApiError("单据日期不能为空"), nil
	}
	//采购时，供应商不可为空
	if (len(req.Purpose) == 0 || req.Purpose == erp.StockEntryTypePurchase) && len(req.Supplier) == 0 {
		return rpc.ApiError("采购时，供应商不能为空"), nil
	}
	// 调用服务层处理业务逻辑
	resp, err := service.Stock().CreateMaterialRequest(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	if len(req.Purpose) == 0 || req.Purpose == erp.StockEntryTypePurchase {
		// 创建并提交采购申请
		purchaseOrder, err := service.Buying().CreatePurchaseFromMq(ctx, &dto.CreatePurchaseFromMqReq{
			SourceName: resp.MaterialRequestName,
			Supplier:   req.Supplier,
		})
		if err != nil {
			return rpc.ApiError(err.Error()), nil
		}
		resp.PurchaseOrder = purchaseOrder.Name
	}
	// 返回成功响应
	return rpc.ApiSuccessWithData("保存物品申请单成功", resp), nil
}

func (*Controller) GetMaterialRequestList(ctx context.Context, req *stock.GetMaterialRequestListReq) (*api.ResponseInfo, error) {
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}

	// 调用服务层获取数据
	resp, err := service.Stock().GetMaterialRequestList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取物品申请单列表成功", resp), nil
}

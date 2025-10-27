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
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
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

	requiredBy := service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.RequiredBy)).Format("Y-m-d")

	//note 下面的交易时间是系统生成的默认是改不了的 transactionDate
	if len(req.Purpose) == 0 || req.Purpose == erp.StockEntryTypePurchase {
		// 创建并提交采购申请, 设置目标仓库为门店在途仓
		purchaseOrder, err := service.Buying().CreatePurchaseFromMq(ctx, &dto.CreatePurchaseFromMqReq{
			SourceName:      resp.MaterialRequestName,
			Supplier:        req.Supplier,
			RequiredBy:      requiredBy,
			TargetWarehouse: req.TargetWarehouse,
		})
		if err != nil {
			//需要回退之前创建的采购申请
			g.Log().Warningf(ctx, "创建采购订单失败，回退之前创建的材料申请:%s\n %v", resp.MaterialRequestName, err)
			_, err2 := service.Document().ChangeDocStatus(ctx, erp.DocTypeMaterialRequest, resp.MaterialRequestName, erp.DocstatusCancelled)
			if err2 != nil {
				g.Log().Warningf(ctx, "回退之前创建的材料申请失败:%s\n %v", resp.MaterialRequestName, err2)
			}
			return rpc.ApiError(err.Error()), nil
		}
		resp.PurchaseOrder = purchaseOrder.Name

		saleOrder := &erp.SaleOrder{}
		if saleOrder, err = service.Buying().CreateInnerSaleOrderFromPurchaseOrder(ctx, &dto.CreateInnerSaleOrderFromPurchaseOrderReq{
			SourceName:      purchaseOrder.Name,
			DeliveryDate:    requiredBy,
			SourceWarehouse: req.SourceWarehouse,
		}); err != nil {
			//需要回退之前创建的材料申请，erp会自动取消采购订单
			g.Log().Warningf(ctx, "创建内部销售订单失败，回退之前创建的材料申请:%s\n %v", resp.MaterialRequestName, err)
			service.Document().ChangeDocStatus(ctx, erp.DocTypeMaterialRequest, resp.MaterialRequestName, erp.DocstatusCancelled)
			return rpc.ApiError(err.Error()), nil
		}

		//直接创建发货单，后续接入物流方

		_, err = service.Buying().CreateDeliveryNoteFromInnerSaleOrder(ctx, &dto.CreateDeliveryNoteFromInnerSaleOrderReq{
			SourceName:      saleOrder.Name,
			SourceWarehouse: req.SourceWarehouse,
			//TargetWarehouse: "", // TODO 取在途仓
		})
		if err != nil {
			g.Log().Warningf(ctx, "创建发货单失败，回退之前创建的内部销售订单及材料申请:%s, %s\n %v", saleOrder.Name, resp.MaterialRequestName, err)
			service.Document().ChangeDocStatus(ctx, erp.DocTypeSaleOrder, saleOrder.Name, erp.DocstatusCancelled)
			service.Document().ChangeDocStatus(ctx, erp.DocTypePurchaseOrder, purchaseOrder.Name, erp.DocstatusCancelled)
			service.Document().ChangeDocStatus(ctx, erp.DocTypeMaterialRequest, resp.MaterialRequestName, erp.DocstatusCancelled)

			return rpc.ApiError(err.Error()), nil
		}
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

// GetStockLedger 获取库存分类账
// 参数：ctx 上下文，req 查询条件
// 返回：库存分类账列表和操作结果
func (*Controller) GetStockLedger(ctx context.Context, req *stock.GetStockLedgerReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if len(req.FromDate) == 0 {
		return rpc.ApiError("开始日期不能为空"), nil
	}
	if len(req.ToDate) == 0 {
		return rpc.ApiError("结束日期不能为空"), nil
	}

	// 调用服务层获取库存分类账数据
	resp, err := service.Stock().GetStockLedger(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取库存分类账成功", resp), nil
}

// SaveStockReconciliation 保存库存盘点
// 参数：ctx 上下文，req 保存库存盘点请求
// 返回：保存结果和操作信息
func (*Controller) SaveStockReconciliation(ctx context.Context, req *stock.SaveStockReconciliationReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}
	if len(req.PostingDate) == 0 {
		return rpc.ApiError("过账日期不能为空"), nil
	}
	if len(req.Items) == 0 {
		return rpc.ApiError("盘点明细不能为空"), nil
	}

	// 验证明细项目
	for i, item := range req.Items {
		if len(item.ItemCode) == 0 {
			return rpc.ApiError(g.I18n().Tf(ctx, "第{0}项物品编码不能为空", i+1)), nil
		}
		if item.Qty <= 0 {
			return rpc.ApiError(g.I18n().Tf(ctx, "第{0}项盘点数量必须大于0", i+1)), nil
		}
	}

	// 调用服务层处理业务逻辑
	resp, err := service.Stock().SaveStockReconciliation(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("保存库存盘点成功", resp), nil
}

// GetStockReconciliationList 获取库存盘点列表
// 参数：ctx 上下文，req 查询条件
// 返回：库存盘点列表和操作结果
func (*Controller) GetStockReconciliationList(ctx context.Context, req *stock.GetStockReconciliationListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.CompanyAbbr) == 0 {
		return rpc.ApiError("公司简称不能为空"), nil
	}

	// 调用服务层获取数据
	resp, err := service.Stock().GetStockReconciliationList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData("获取库存盘点列表成功", resp), nil
}

// SubmitStockReconciliation 提交库存盘点
// 参数：ctx 上下文，req 提交库存盘点请求
// 返回：提交结果和操作信息
func (*Controller) SubmitStockReconciliation(ctx context.Context, req *stock.SubmitStockReconciliationReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.StockReconciliationName) == 0 {
		return rpc.ApiError("库存盘点单号不能为空"), nil
	}

	// 调用服务层处理业务逻辑
	resp, err := service.Stock().SubmitStockReconciliation(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData(resp.Message, resp), nil
}

// CancelStockReconciliation 取消库存盘点
// 参数：ctx 上下文，req 取消库存盘点请求
// 返回：取消结果和操作信息
func (*Controller) CancelStockReconciliation(ctx context.Context, req *stock.CancelStockReconciliationReq) (*api.ResponseInfo, error) {
	// 参数验证
	if len(req.StockReconciliationName) == 0 {
		return rpc.ApiError("库存盘点单号不能为空"), nil
	}

	// 调用服务层处理业务逻辑
	resp, err := service.Stock().CancelStockReconciliation(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 返回成功响应
	return rpc.ApiSuccessWithData(resp.Message, resp), nil
}

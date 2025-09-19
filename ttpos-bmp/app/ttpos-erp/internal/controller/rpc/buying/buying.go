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
	"github.com/gogf/gf/v2/util/gconv"
)

type Controller struct {
	buying.UnimplementedBuyingServiceServer
}

func Register(s *grpcx.GrpcServer) {
	buying.RegisterBuyingServiceServer(s.Server, &Controller{})
}

// GetSupplierList 获取供应商列表,只返回内部供应商列表
// 参数：ctx 上下文，req 获取供应商列表请求
// 返回：响应信息和错误
func (c *Controller) GetSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetSupplierListReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	resp, err := service.Supplier().GetInnerSupplierList(ctx, req)
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

// GetPurchaseOrder 获取采购订单
// 参数：ctx 上下文，req 获取采购订单请求
// 返回：响应信息和错误
func (c *Controller) GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateGetPurchaseOrderReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层获取数据
	resp, err := service.Buying().GetPurchaseOrder(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 转换响应数据格式
	purchaseOrderItems := make([]*buying.PurchaseOrderItem, 0)
	for _, item := range resp.Items {
		purchaseOrderItems = append(purchaseOrderItems, &buying.PurchaseOrderItem{
			ItemName: item.ItemName,
			ItemCode: item.ItemCode,
			StockUom: item.StockUom,
			Qty:      item.Qty,
		})
	}

	return rpc.ApiSuccessWithData("查询采购订单成功", &buying.GetPurchaseOrderResp{
		PurchaseOrder: &buying.PurchaseOrderInfo{
			PurchaseOrderName: resp.Name,
			SupplierName:      resp.Supplier,
			PerReceived:       resp.PerReceived,
			ScheduleDate:      resp.ScheduleDate,
			Items:             purchaseOrderItems,
		},
	}), nil
}

// validateGetPurchaseOrderReq 验证获取采购订单请求参数
func (c *Controller) validateGetPurchaseOrderReq(req *buying.GetPurchaseOrderReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.PurchaseOrderName) == "" {
		return gerror.New("采购订单名称不能为空")
	}
	return nil
}

// SavePurchaseReceipt 保存采购收货单
// 参数：ctx 上下文，req 保存采购收货单请求
// 返回：响应信息和错误
func (c *Controller) SavePurchaseReceipt(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*api.ResponseInfo, error) {
	// 参数验证
	if err := c.validateSavePurchaseReceiptReq(req); err != nil {
		return rpc.ApiError(err.Error()), nil
	}

	// 调用服务层保存数据
	res, err := service.Buying().CreatePurchaseReceiptFromOrder(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	purchaseReceiptItems := make([]*buying.PurchaseReceiptItem, 0)
	for _, item := range res.Items {
		purchaseReceiptItems = append(purchaseReceiptItems, &buying.PurchaseReceiptItem{
			ItemName: item.ItemName,
			ItemCode: item.ItemCode,
			StockUom: item.StockUom,
			Qty:      item.Qty,
		})
	}
	return rpc.ApiSuccessWithData("保存采购收货成功", &buying.SavePurchaseReceiptResp{
		PurchaseReceipt: &buying.PurchaseReceiptInfo{
			PurchaseReceiptName: res.Name,
			Items:               purchaseReceiptItems,
		},
	}), nil
}

// validateSavePurchaseReceiptReq 验证保存采购收货单请求参数
func (c *Controller) validateSavePurchaseReceiptReq(req *buying.SavePurchaseReceiptReq) error {
	if req == nil {
		return gerror.New("请求参数不能为空")
	}
	if strings.TrimSpace(req.PurchaseOrderName) == "" {
		return gerror.New("采购订单名称不能为空")
	}
	if req.Items == nil || len(req.Items) == 0 {
		return gerror.New("采购订单物品列表不能为空")
	}

	// 验证每个物品项的参数
	for i, item := range req.Items {
		if item == nil {
			return gerror.Newf("第%d个物品项不能为空", i+1)
		}
		if strings.TrimSpace(item.ItemCode) == "" {
			return gerror.Newf("第%d个物品项的物品编码不能为空", i+1)
		}
		if item.Qty <= 0 {
			return gerror.Newf("第%d个物品项的数量必须大于0", i+1)
		}
	}

	return nil
}

func (*Controller) CreateSupplier(ctx context.Context, req *buying.CreateSupplierReq) (*api.ResponseInfo, error) {
	resp, err := service.Supplier().CreateSupplier(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("创建供应商成功", resp), nil
}

func (*Controller) GetSupplier(ctx context.Context, req *buying.GetSupplierReq) (*api.ResponseInfo, error) {
	resp, err := service.Supplier().GetSupplier(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	supplier := &buying.SupplierData{}
	gconv.Scan(resp, &supplier)
	return rpc.ApiSuccessWithData("获取供应商成功", &buying.GetSupplierResp{
		Supplier: supplier,
	}), nil
}

func (*Controller) UpdateSupplier(ctx context.Context, req *buying.UpdateSupplierReq) (*api.ResponseInfo, error) {
	resp, err := service.Supplier().UpdateSupplier(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("更新供应商成功", resp), nil
}

func (*Controller) DeleteSupplier(ctx context.Context, req *buying.DeleteSupplierReq) (*api.ResponseInfo, error) {
	resp, err := service.Supplier().DeleteSupplier(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("删除供应商成功", resp), nil
}

func (*Controller) ListSuppliers(ctx context.Context, req *buying.ListSuppliersReq) (*api.ResponseInfo, error) {
	resp, err := service.Supplier().ListSuppliers(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("获取供应商列表成功", resp), nil
}

// GetPurchaseOrderList 获取采购订单列表
func (*Controller) GetPurchaseOrderList(ctx context.Context, req *buying.GetPurchaseOrderListReq) (*api.ResponseInfo, error) {
	resp, err := service.Buying().GetPurchaseOrderList(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("获取采购订单列表成功", resp), nil
}

// GetPurchaseOrderCount 获取采购订单数量
func (*Controller) GetPurchaseOrderCount(ctx context.Context, req *buying.GetPurchaseOrderCountReq) (*api.ResponseInfo, error) {
	resp, err := service.Buying().GetPurchaseOrderCount(ctx, req)
	if err != nil {
		return rpc.ApiError(err.Error()), nil
	}
	return rpc.ApiSuccessWithData("获取采购订单数量成功", resp), nil
}

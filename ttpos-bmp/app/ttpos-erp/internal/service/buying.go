// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/buying"
	dto "ttpos-bmp/app/ttpos-erp/internal/model/dto/buying"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	IBuying interface {
		// CreatePurchaseFromMq 根据材料请求创建采购订单
		CreatePurchaseFromMq(ctx context.Context, req *dto.CreatePurchaseFromMqReq) (res *erp.PurchaseOrder, err error)
		// CreateInnerSaleOrderFromPurchaseOrder 创建内部销售订单
		CreateInnerSaleOrderFromPurchaseOrder(ctx context.Context, req *dto.CreateInnerSaleOrderFromPurchaseOrderReq) (res *erp.SaleOrder, err error)
		// GetPurchaseOrder 获取采购订单
		GetPurchaseOrder(ctx context.Context, req *buying.GetPurchaseOrderReq) (*erp.PurchaseOrder, error)
		// CreatePurchaseReceiptFromOrder 创建采购收货订单
		CreatePurchaseReceiptFromOrder(ctx context.Context, req *buying.SavePurchaseReceiptReq) (*erp.PurchaseReceipt, error)
		// GetPurchaseOrderList 获取采购订单列表
		// 根据查询条件过滤并返回采购订单信息列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取采购订单列表请求参数
		//
		// 返回：
		//   - res: 采购订单列表响应
		//   - err: 错误信息
		GetPurchaseOrderList(ctx context.Context, req *buying.GetPurchaseOrderListReq) (res *buying.GetPurchaseOrderListResp, err error)
		// GetPurchaseOrderCount 获取采购订单数量
		// 根据查询条件统计采购订单数量
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取采购订单数量请求参数
		//
		// 返回：
		//   - res: 采购订单数量响应
		//   - err: 错误信息
		GetPurchaseOrderCount(ctx context.Context, req *buying.GetPurchaseOrderCountReq) (res *buying.GetPurchaseOrderCountResp, err error)
	}
	ISupplier interface {
		// GetInnerSupplierList 获取内部供应商列表
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 获取供应商列表请求参数，包含公司缩写编码
		//
		// 返回：
		//   - res: 获取供应商列表响应，包含供应商基本信息列表
		//   - err: 操作过程中产生的错误（若有）
		GetInnerSupplierList(ctx context.Context, req *buying.GetSupplierListReq) (*buying.GetSupplierListResp, error)
		// AddSupplerTransactCompany 为供应商添加允许交易的公司
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 添加供应商交易公司请求参数，包含供应商名称和公司缩写
		//
		// 返回：
		//   - err: 操作过程中产生的错误（若有）
		AddSupplerTransactCompany(ctx context.Context, req *dto.AddSupplerTransactCompanyReq) error
		// CreateSupplier 创建供应商
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 创建供应商请求参数，包含供应商详细信息
		//
		// 返回：
		//   - res: 创建供应商响应，包含创建结果
		//   - err: 操作过程中产生的错误（若有）
		CreateSupplier(ctx context.Context, req *buying.CreateSupplierReq) (*buying.CreateSupplierResp, error)
		// GetSupplier 获取供应商详情
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 获取供应商请求参数，包含供应商名称或ID
		//
		// 返回：
		//   - res: 获取供应商响应，包含供应商详细信息
		//   - err: 操作过程中产生的错误（若有）
		GetSupplier(ctx context.Context, req *buying.GetSupplierReq) (*erp.Supplier, error)
		// UpdateSupplier 更新供应商
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 更新供应商请求参数，包含供应商更新信息
		//
		// 返回：
		//   - res: 更新供应商响应，包含更新结果
		//   - err: 操作过程中产生的错误（若有）
		UpdateSupplier(ctx context.Context, req *buying.UpdateSupplierReq) (*buying.UpdateSupplierResp, error)
		// DeleteSupplier 删除供应商
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 删除供应商请求参数，包含供应商名称或ID
		//
		// 返回：
		//   - res: 删除供应商响应，包含删除结果
		//   - err: 操作过程中产生的错误（若有）
		DeleteSupplier(ctx context.Context, req *buying.DeleteSupplierReq) (*buying.DeleteSupplierResp, error)
		// ListSuppliers 获取供应商列表
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 获取供应商列表请求参数，包含分页和过滤条件
		//
		// 返回：
		//   - res: 获取供应商列表响应，包含供应商列表和分页信息
		//   - err: 操作过程中产生的错误（若有）
		ListSuppliers(ctx context.Context, req *buying.ListSuppliersReq) (*buying.ListSuppliersResp, error)
	}
)

var (
	localBuying   IBuying
	localSupplier ISupplier
)

func Buying() IBuying {
	if localBuying == nil {
		panic("implement not found for interface IBuying, forgot register?")
	}
	return localBuying
}

func RegisterBuying(i IBuying) {
	localBuying = i
}

func Supplier() ISupplier {
	if localSupplier == nil {
		panic("implement not found for interface ISupplier, forgot register?")
	}
	return localSupplier
}

func RegisterSupplier(i ISupplier) {
	localSupplier = i
}

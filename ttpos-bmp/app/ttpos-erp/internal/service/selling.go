// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
)

type (
	IAsyncSelling interface {
		CancelPosInvoice(ctx context.Context, req *selling.CancelPosInvoiceReq) (asyncRecordId string, err error)
		SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error)
		ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
		ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
		GetLatestReceivePosInvoice(ctx context.Context, req *do.ReceivePosInvoice) (*entity.ReceivePosInvoice, error)
	}
	ISelling interface {
		// CreateSalesOrder 创建销售订单
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 销售订单信息
		//
		// 返回：
		//   - *dtoSelling.SalesOrder: 创建后的销售订单信息
		//   - error: 错误信息
		CreateSalesOrder(ctx context.Context, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error)
		// UpdateSalesOrder 更新销售订单
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 销售订单名称
		//   - req: 更新的销售订单信息
		//
		// 返回：
		//   - *dtoSelling.SalesOrder: 更新后的销售订单信息
		//   - error: 错误信息
		UpdateSalesOrder(ctx context.Context, name string, req *dtoSelling.SalesOrder) (*dtoSelling.SalesOrder, error)
		// GetSalesOrder 获取销售订单信息
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 销售订单名称
		//
		// 返回：
		//   - *dtoSelling.SalesOrder: 销售订单信息
		//   - error: 错误信息
		GetSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error)
		// GetSalesOrderList 获取销售订单列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 查询参数
		//
		// 返回：
		//   - []*dtoSelling.SalesOrder: 销售订单列表
		//   - error: 错误信息
		GetSalesOrderList(ctx context.Context, req *dtoSelling.SalesOrderListReq) ([]*dtoSelling.SalesOrder, error)
		// CountSalesOrder 统计销售订单数量
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 查询参数
		//
		// 返回：
		//   - int: 销售订单数量
		//   - error: 错误信息
		CountSalesOrder(ctx context.Context, req *dtoSelling.SalesOrderListReq) (int, error)
		// CancelSalesOrder 删除销售订单（取消订单）
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 销售订单名称
		//
		// 返回：
		//   - error: 错误信息
		CancelSalesOrder(ctx context.Context, name string) error
		// SubmitSalesOrder 提交销售订单
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 销售订单名称
		//
		// 返回：
		//   - *dtoSelling.SalesOrder: 提交后的销售订单信息
		//   - error: 错误信息
		SubmitSalesOrder(ctx context.Context, name string) (*dtoSelling.SalesOrder, error)
		// GetPosProfileList 查询POS配置文件列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 查询请求参数
		//
		// 返回：
		//   - *selling.PosProfileListResp: POS配置文件列表响应
		//   - error: 错误信息
		GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error)
		// CreateModePaymentAccount 创建支付方式账户
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 创建支付方式账户请求参数
		//
		// 返回：
		//   - error: 错误信息
		CreateModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) error
		// CreatePosProfile 创建默认的POS配置文件
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 创建POS配置文件请求参数
		//
		// 返回：
		//   - *erp.POSProfile: POS配置文件信息
		//   - error: 错误信息
		CreatePosProfile(ctx context.Context, req *setup.CreatePosProfileInp) (*erp.POSProfile, error)
		// OpenPosEntry 开帐
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 开帐请求参数
		//
		// 返回：
		//   - *selling.OpenPosEntryResp: 开帐响应信息
		//   - error: 错误信息
		OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error)
		// ClosePosEntry 关帐
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 关帐请求参数
		//
		// 返回：
		//   - *selling.ClosePosEntryResp: 关帐响应信息
		//   - error: 错误信息
		ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
		// IsProfileOpening 查询POS配置文件是否开帐
		// 参数：
		//   - ctx: 上下文对象
		//   - posProfile: POS配置文件名称
		//
		// 返回：
		//   - bool: 是否已开帐
		//   - error: 错误信息
		IsProfileOpening(ctx context.Context, posProfile string, user string) (bool, error)
		// GetPosInvoiceList 获取POS发票列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取POS发票列表请求参数
		//
		// 返回：
		//   - []dtoSelling.SimplePosInvoice: POS发票列表
		//   - error: 错误信息
		GetPosInvoiceList(ctx context.Context, req *dtoSelling.GetPosInvoiceListReq) ([]dtoSelling.SimplePosInvoice, error)
		// SavePosInvoice 保存POS发票
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 保存POS发票请求参数
		//
		// 返回：
		//   - *selling.SavePosInvoiceResp: 保存POS发票响应信息
		//   - error: 错误信息
		SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error)
		// SavePosInvoiceStep 保存POS发票步骤
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 保存POS发票请求参数
		//   - openingEntry: POS开帐记录
		//   - isMaterialItem: 是否为物品发票
		//
		// 返回：
		//   - string: POS发票名称
		//   - error: 错误信息
		SavePosInvoiceStep(ctx context.Context, req *selling.SavePosInvoiceReq, openingEntry *erp.POSOpeningEntry, isMaterialItem bool) (string, error)
		// GetPosOpeningEntry 获取POS开帐记录
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 开帐记录名称
		//
		// 返回：
		//   - *erp.POSOpeningEntry: POS开帐记录信息
		//   - error: 错误信息
		GetPosOpeningEntry(ctx context.Context, name string) (*erp.POSOpeningEntry, error)
		// ReturnPosInvoice 退货POS发票
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 退货POS发票请求参数
		//
		// 返回：
		//   - *selling.ReturnPosInvoiceResp: 退货POS发票响应参数
		//   - error: 错误信息
		//
		// 功能：
		//   - 退货指定名称的POS发票
		ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
		// CancelPosInvoice 取消POS发票
		// 参数：
		//   - ctx: 上下文对象
		//   - invoiceName: 发票名称
		//
		// 返回：
		//   - error: 错误信息
		//
		// 功能：
		//   - 取消指定名称的POS发票
		CancelPosInvoice(ctx context.Context, invoiceName string) error
		// GetModeOfPaymentList 获取支付方式列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取支付方式列表请求参数
		//
		// 返回：
		//   - *selling.GetModeOfPaymentListResp: 获取支付方式列表响应参数
		//   - error: 错误信息
		//
		// 功能：
		//   - 获取支付方式列表
		GetModeOfPaymentList(ctx context.Context, req *selling.GetModeOfPaymentListReq) (*selling.GetModeOfPaymentListResp, error)
		// CountCustomer 统计客户数量
		// 参数：
		//   - ctx: 上下文对象
		//   - filter: 客户过滤条件，可选
		//
		// 返回：
		//   - int: 客户数量
		//   - error: 错误信息
		CountCustomer(ctx context.Context, filter *erp.Customer) (int, error)
		// CreateCustomer 创建客户
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 客户信息
		//
		// 返回：
		//   - *erp.Customer: 创建后的客户信息
		//   - error: 错误信息
		CreateCustomer(ctx context.Context, req *erp.Customer) (*erp.Customer, error)
		// UpdateCustomer 更新客户
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 客户名称
		//   - req: 更新的客户信息
		//
		// 返回：
		//   - *erp.Customer: 更新后的客户信息
		//   - error: 错误信息
		UpdateCustomer(ctx context.Context, name string, req *erp.Customer) (*erp.Customer, error)
		// GetCustomer 获取客户信息
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 客户名称
		//
		// 返回：
		//   - *erp.Customer: 客户信息
		//   - error: 错误信息
		GetCustomer(ctx context.Context, name string) (*erp.Customer, error)
		// AddCompanyToCustomer 将公司添加到客户的允许交易公司列表
		// 参数：
		//   - ctx: 上下文对象
		//   - customer: 客户信息
		//   - companyName: 要添加的公司名称
		//
		// 返回：
		//   - error: 错误信息
		AddCompanyToCustomer(ctx context.Context, customer *erp.Customer, companyName string) error
		// ListCustomers 获取客户列表
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 获取客户列表请求参数，包含分页和过滤条件
		//
		// 返回：
		//   - []*erp.Customer: 客户列表
		//   - error: 错误信息
		ListCustomers(ctx context.Context, req *dtoSelling.ListCustomersReq) ([]*erp.Customer, error)
	}
)

var (
	localAsyncSelling IAsyncSelling
	localSelling      ISelling
)

func AsyncSelling() IAsyncSelling {
	if localAsyncSelling == nil {
		panic("implement not found for interface IAsyncSelling, forgot register?")
	}
	return localAsyncSelling
}

func RegisterAsyncSelling(i IAsyncSelling) {
	localAsyncSelling = i
}

func Selling() ISelling {
	if localSelling == nil {
		panic("implement not found for interface ISelling, forgot register?")
	}
	return localSelling
}

func RegisterSelling(i ISelling) {
	localSelling = i
}

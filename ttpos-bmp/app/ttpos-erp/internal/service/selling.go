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

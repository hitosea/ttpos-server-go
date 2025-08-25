// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
)

type (
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
		// GetPosOpeningEntry 获取POS开帐记录
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 开帐记录名称
		//
		// 返回：
		//   - *erp.POSOpeningEntry: POS开帐记录信息
		//   - error: 错误信息
		GetPosOpeningEntry(ctx context.Context, name string) (*erp.POSOpeningEntry, error)
		ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error)
	}
)

var (
	localSelling ISelling
)

func Selling() ISelling {
	if localSelling == nil {
		panic("implement not found for interface ISelling, forgot register?")
	}
	return localSelling
}

func RegisterSelling(i ISelling) {
	localSelling = i
}

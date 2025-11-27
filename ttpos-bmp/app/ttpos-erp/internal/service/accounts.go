// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	IAccounts interface {
		// GetDefaultPurchaseTaxTemplate 查询公司默认采购税费模板名称
		// 参数：
		//   - ctx: 上下文对象
		//   - company: 公司名称
		//
		// 返回：
		//   - 默认采购税费模板名称，如果不存在则返回空字符串
		GetDefaultPurchaseTaxTemplate(ctx context.Context, company string) string
		// GetDefaultSalesTaxTemplate 查询公司默认销售税费模板名称
		// 参数：
		//   - ctx: 上下文对象
		//   - company: 公司名称
		//
		// 返回：
		//   - 默认销售税费模板名称，如果不存在则返回空字符串
		GetDefaultSalesTaxTemplate(ctx context.Context, company string) string
		// GetPurchaseTaxTemplateDetails 根据模板名称获取采购税费明细
		// 参数：
		//   - ctx: 上下文对象
		//   - templateName: 采购税费模板名称
		//
		// 返回：
		//   - 采购税费明细列表，如果不存在则返回 nil
		GetPurchaseTaxTemplateDetails(ctx context.Context, templateName string) []*erp.PurchaseTaxesAndCharges
		// GetSalesTaxTemplateDetails 根据模板名称获取销售税费明细
		// 参数：
		//   - ctx: 上下文对象
		//   - templateName: 销售税费模板名称
		//
		// 返回：
		//   - 销售税费明细列表，如果不存在则返回 nil
		GetSalesTaxTemplateDetails(ctx context.Context, templateName string) []*erp.SalesTaxesAndCharges
	}
)

var (
	localAccounts IAccounts
)

func Accounts() IAccounts {
	if localAccounts == nil {
		panic("implement not found for interface IAccounts, forgot register?")
	}
	return localAccounts
}

func RegisterAccounts(i IAccounts) {
	localAccounts = i
}

// Package accounts 提供 ERPNext 账务相关服务
package accounts

import (
	"context"

	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/service"
)

// GetDefaultPurchaseTaxTemplate 查询公司默认采购税费模板名称
// 参数：
//   - ctx: 上下文对象
//   - company: 公司名称
//
// 返回：
//   - 默认采购税费模板名称，如果不存在则返回空字符串
func (s *sAccounts) GetDefaultPurchaseTaxTemplate(ctx context.Context, company string) string {
	if company == "" {
		return ""
	}

	filters := [][]string{
		{"company", "=", company},
		{"is_default", "=", "1"},
		{"disabled", "=", "0"},
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseTaxesTemplate,
	}, &erp.RequestParams{Filters: filters, Limit: 1})

	if err != nil || resp == nil || resp.IsNil() {
		return ""
	}

	dataArray := resp.GetJsons("data")
	if len(dataArray) == 0 {
		return ""
	}

	return dataArray[0].Get("name").String()
}

// GetDefaultSalesTaxTemplate 查询公司默认销售税费模板名称
// 参数：
//   - ctx: 上下文对象
//   - company: 公司名称
//
// 返回：
//   - 默认销售税费模板名称，如果不存在则返回空字符串
func (s *sAccounts) GetDefaultSalesTaxTemplate(ctx context.Context, company string) string {
	if company == "" {
		return ""
	}

	filters := [][]string{
		{"company", "=", company},
		{"is_default", "=", "1"},
		{"disabled", "=", "0"},
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSalesTaxesTemplate,
	}, &erp.RequestParams{Filters: filters, Limit: 1})

	if err != nil || resp == nil || resp.IsNil() {
		return ""
	}

	dataArray := resp.GetJsons("data")
	if len(dataArray) == 0 {
		return ""
	}

	return dataArray[0].Get("name").String()
}

// GetPurchaseTaxTemplateDetails 根据模板名称获取采购税费明细
// 参数：
//   - ctx: 上下文对象
//   - templateName: 采购税费模板名称
//
// 返回：
//   - 采购税费明细列表，如果不存在则返回 nil
func (s *sAccounts) GetPurchaseTaxTemplateDetails(ctx context.Context, templateName string) []*erp.PurchaseTaxesAndCharges {
	if templateName == "" {
		return nil
	}

	// 获取模板详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypePurchaseTaxesTemplate,
		Name:    templateName,
	}, nil)

	if err != nil || resp == nil || resp.IsNil() {
		return nil
	}

	// 解析 taxes 子表
	taxesJson := resp.GetJson("data.taxes")
	if taxesJson == nil || taxesJson.IsNil() {
		return nil
	}

	var taxes []*erp.PurchaseTaxesAndCharges
	if err := taxesJson.Scan(&taxes); err != nil {
		return nil
	}

	return taxes
}

// GetSalesTaxTemplateDetails 根据模板名称获取销售税费明细
// 参数：
//   - ctx: 上下文对象
//   - templateName: 销售税费模板名称
//
// 返回：
//   - 销售税费明细列表，如果不存在则返回 nil
func (s *sAccounts) GetSalesTaxTemplateDetails(ctx context.Context, templateName string) []*erp.SalesTaxesAndCharges {
	if templateName == "" {
		return nil
	}

	// 获取模板详情
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeSalesTaxesTemplate,
		Name:    templateName,
	}, nil)

	if err != nil || resp == nil || resp.IsNil() {
		return nil
	}

	// 解析 taxes 子表
	taxesJson := resp.GetJson("data.taxes")
	if taxesJson == nil || taxesJson.IsNil() {
		return nil
	}

	var taxes []*erp.SalesTaxesAndCharges
	if err := taxesJson.Scan(&taxes); err != nil {
		return nil
	}

	return taxes
}

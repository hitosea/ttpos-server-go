// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/company"
)

type (
	ICompany interface {
		GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (res *company.GetCompanyListResp, err error)
		GetCompanyWithAbbr(ctx context.Context, abbr string) (res *company.CompanyInfo, err error)
		// GetCompanyNameWithAbbr  根据公司简称获取公司名称
		GetCompanyNameWithAbbr(ctx context.Context, companyAbbr string) (string, error)
		// HasSubCompany 判断公司是否有子公司
		// 通过 parent_company 字段关联查询，判断是否存在子公司
		// 参数：ctx 上下文，companyName 公司名称
		// 返回：是否有子公司，错误信息
		HasSubCompany(ctx context.Context, companyName string) (bool, error)
	}
)

var (
	localCompany ICompany
)

func Company() ICompany {
	if localCompany == nil {
		panic("implement not found for interface ICompany, forgot register?")
	}
	return localCompany
}

func RegisterCompany(i ICompany) {
	localCompany = i
}

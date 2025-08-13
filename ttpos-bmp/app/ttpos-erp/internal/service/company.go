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

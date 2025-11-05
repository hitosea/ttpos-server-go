// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/company"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
)

type (
	ICompany interface {
		GetCompanyList(ctx context.Context, req *company.GetCompanyListReq) (res *company.GetCompanyListResp, err error)
		// GetCompany 根据公司名称获取公司信息
		// 参数：
		//   - ctx: 上下文对象
		//   - name: 公司名称
		//
		// 返回：
		//   - res: 公司信息
		//   - err: 错误信息
		GetCompany(ctx context.Context, name string) (res *erp.Company, err error)
		GetCompanyWithAbbr(ctx context.Context, abbr string) (res *company.CompanyInfo, err error)
		// GetCompanyNameWithAbbr  根据公司简称获取公司名称
		GetCompanyNameWithAbbr(ctx context.Context, companyAbbr string) (string, error)
		// HasSubCompany 判断公司是否有子公司
		// 通过 parent_company 字段关联查询，判断是否存在子公司
		// 参数：ctx 上下文，companyName 公司名称
		// 返回：是否有子公司，错误信息
		HasSubCompany(ctx context.Context, companyName string) (bool, error)
		// GetAllSubCompanies 递归查询指定公司下的所有子公司
		// 通过 parent_company 字段递归查询，获取所有层级的子公司
		// 参数：
		//   - ctx: 上下文对象
		//   - companyName: 父公司名称
		//
		// 返回：
		//   - []*company.CompanyInfo: 所有子公司列表（包括所有层级的子公司）
		//   - error: 错误信息
		GetAllSubCompanies(ctx context.Context, companyName string) ([]*company.CompanyInfo, error)
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

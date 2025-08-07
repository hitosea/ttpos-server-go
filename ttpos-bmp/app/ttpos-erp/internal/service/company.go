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
		// CreateBranch 创建分店
		// 参数：店铺名称和公司缩写编码
		// 返回：ERP用户名和创建结果
		CreateBranch(ctx context.Context, req *company.CreateBranchReq) (res *company.CreateBranchResp, err error)
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

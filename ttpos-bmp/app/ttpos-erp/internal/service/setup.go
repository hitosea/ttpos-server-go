// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/setup"
	setup2 "ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
)

type (
	ISetup interface {
		// CreateBranch 创建分店
		// 参数：店铺名称和公司缩写编码
		// 返回：ERP用户名和创建结果
		CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, err error)
		// CreateUser 创建网站用户
		CreateUser(ctx context.Context, req *setup2.CreateUserInp) (userEmail string, err error)
		// CreatePosProfile CreatePosFile 创建 默认 pos profile  配置默认
		CreatePosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posFileId string, err error)
		// InitShop 初始化店铺
		// 参数：ctx 上下文，req 包含 shop_name、company_abbr、shop_uuid
		// 返回：是否成功，错误信息
		InitShop(ctx context.Context, req *setup.InitShopReq) (resp *setup.InitShopResp, err error)
	}
)

var (
	localSetup ISetup
)

func Setup() ISetup {
	if localSetup == nil {
		panic("implement not found for interface ISetup, forgot register?")
	}
	return localSetup
}

func RegisterSetup(i ISetup) {
	localSetup = i
}

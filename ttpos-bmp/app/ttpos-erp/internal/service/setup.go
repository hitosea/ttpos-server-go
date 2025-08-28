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
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 初始化店铺请求参数
		//
		// 返回：
		//   - branchName: 分店名称
		//   - err: 错误信息
		CreateBranch(ctx context.Context, req *setup.InitShopReq) (branchName string, err error)
		// CreateUser 创建网站用户，门店收银账户
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 创建用户请求参数
		//
		// 返回：
		//   - err: 错误信息
		CreateUser(ctx context.Context, req *setup2.CreateUserInp) error
		// CreateDefaultPosProfile 创建默认的POS配置文件
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 创建默认POS配置文件请求参数
		//
		// 返回：
		//   - posProfileName: POS配置文件名称
		//   - err: 错误信息
		CreateDefaultPosProfile(ctx context.Context, req *setup.CreateDefaultPosProfileReq) (posProfileName string, err error)
		// InitShop 初始化店铺
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 初始化店铺请求参数
		//
		// 返回：
		//   - resp: 初始化店铺响应结果
		//   - err: 错误信息
		InitShop(ctx context.Context, req *setup.InitShopReq) (resp *setup.InitShopResp, err error)
		// GetUserApiKeySecret 获取用户的API密钥和密钥
		// 参数：
		//   - ctx: 上下文对象
		//   - userEmail: 用户邮箱
		//
		// 返回：
		//   - apiKey: API密钥
		//   - apiSecret: API密钥
		//   - err: 错误信息
		GetUserApiKeySecret(ctx context.Context, userEmail string) (apiKey string, apiSecret string, err error)
		// InitCustomFields 初始化自定义字段
		// 遍历manifest/erp-migrate/v2.5/custom_fields目录下所有JSON文件，创建Custom Field文档
		// 参数：
		//   - ctx: 上下文对象
		//
		// 返回：
		//   - err: 错误信息
		InitCustomFields(ctx context.Context, dirBase string) error
		// InitCustomers 初始化客户
		// 遍历manifest/erp-migrate/v2.5/new_customer目录下所有JSON文件，创建Customer文档
		// 参数：
		//   - ctx: 上下文对象
		//
		// 返回：
		//   - err: 错误信息
		InitCustomers(ctx context.Context, dirBase string) error
		// InitModeOfPayment 初始化支付方式
		// 遍历manifest/erp-migrate/v2.5/mode_of_payment目录下所有JSON文件，创建Mode Of Payment文档
		// 参数：
		//   - ctx: 上下文对象
		//
		// 返回：
		//   - err: 错误信息
		InitModeOfPayment(ctx context.Context, dirBase string) error
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

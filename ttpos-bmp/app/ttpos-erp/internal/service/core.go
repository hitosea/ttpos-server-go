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
	IPosPriceList interface {
		// CreatePosPriceList 创建 POS 价格表规则
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 创建价格表规则请求参数
		//
		// 返回：
		//   - res: 创建的价格表规则数据
		//   - err: 操作过程中产生的错误（若有）
		CreatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error)
		// GetPosPriceList 获取 POS 价格表规则详情
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - name: 价格表规则名称
		//
		// 返回：
		//   - res: 价格表规则详细信息
		//   - err: 操作过程中产生的错误（若有）
		GetPosPriceList(ctx context.Context, name string) (*erp.PosPriceList, error)
		// UpdatePosPriceList 更新 POS 价格表规则
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 更新价格表规则请求参数
		//
		// 返回：
		//   - res: 更新后的价格表规则数据
		//   - err: 操作过程中产生的错误（若有）
		UpdatePosPriceList(ctx context.Context, req *erp.PosPriceList) (*erp.PosPriceList, error)
		// DeletePosPriceList 删除 POS 价格表规则
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - name: 价格表规则名称
		//
		// 返回：
		//   - err: 操作过程中产生的错误（若有）
		DeletePosPriceList(ctx context.Context, name string) error
		// ListPosPriceLists 获取 POS 价格表规则列表
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - req: 查询价格表规则列表请求参数
		//
		// 返回：
		//   - list: 价格表规则列表
		//   - err: 操作过程中产生的错误（若有）
		ListPosPriceLists(ctx context.Context, req *erp.ListPosPriceListsReq) ([]*erp.PosPriceList, error)
		// GetDefaultPosPriceList 从配置文件获取默认价格表
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//
		// 返回：
		//   - res: 默认价格表配置
		//   - err: 操作过程中产生的错误（若有）
		GetDefaultPosPriceList(ctx context.Context) (*erp.PosPriceList, error)
		// GetPosPriceListByCompany 根据公司获取默认价格表规则
		// 参数：
		//   - ctx: 上下文对象，用于传递请求范围的元数据
		//   - company: 公司名称
		//
		// 返回：
		//   - res: 价格表规则数据
		//   - err: 操作过程中产生的错误（若有）
		GetPosPriceListByCompany(ctx context.Context, company string) (*erp.PosPriceList, error)
	}
	IUser interface {
		// GetUserByUsername 根据用户名获取用户信息
		// 参数：
		//   - ctx: 上下文
		//   - userEmail: 用户名
		//
		// 返回：
		//   - *erp.User: 用户信息
		//   - error: 错误信息
		GetUserByUsername(ctx context.Context, userEmail string) (*erp.User, error)
		// MustGetUserTimeZone 获取用户时区
		// 参数：
		//   - ctx: 上下文
		//   - userEmail: 用户邮箱
		//
		// 返回：
		//   - string: 时区
		MustGetUserTimeZone(ctx context.Context, userEmail string) string
	}
)

var (
	localPosPriceList IPosPriceList
	localUser         IUser
)

func PosPriceList() IPosPriceList {
	if localPosPriceList == nil {
		panic("implement not found for interface IPosPriceList, forgot register?")
	}
	return localPosPriceList
}

func RegisterPosPriceList(i IPosPriceList) {
	localPosPriceList = i
}

func User() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser, forgot register?")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}

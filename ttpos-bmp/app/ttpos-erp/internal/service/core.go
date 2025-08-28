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
	localUser IUser
)

func User() IUser {
	if localUser == nil {
		panic("implement not found for interface IUser, forgot register?")
	}
	return localUser
}

func RegisterUser(i IUser) {
	localUser = i
}

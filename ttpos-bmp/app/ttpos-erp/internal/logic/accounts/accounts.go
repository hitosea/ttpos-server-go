// Package accounts 提供 ERPNext 账务相关服务
// 包含税费模板查询等功能
package accounts

import "ttpos-bmp/app/ttpos-erp/internal/service"

var Accounts = new(sAccounts)

type sAccounts struct{}

func init() {
	service.RegisterAccounts(Accounts)
}

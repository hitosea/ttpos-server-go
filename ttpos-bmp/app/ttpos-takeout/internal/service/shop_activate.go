// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"ttpos-bmp/app/ttpos-takeout/api/shop"
)

type (
	IShopActivate interface {
		// ActivateShop 激活门店外卖渠道（多渠道路由）
		// 参数：
		//   - ctx: 上下文对象
		//   - req: 激活请求参数，包含 provider_name、shop_uuid、request_id
		// 返回：
		//   - res: 激活结果，包含 shop_uuid、provider_name、updated_at，Grab 渠道额外返回 self_serve_url
		//   - err: 操作过程中产生的错误（若有）
		ActivateShop(ctx context.Context, req *shop.ActivateShopReq) (*shop.ActivateShopResp, error)
	}
)

var (
	localShopActivate IShopActivate
)

func ShopActivate() IShopActivate {
	if localShopActivate == nil {
		panic("implement not found for interface IShopActivate, forgot register?")
	}
	return localShopActivate
}

func RegisterShopActivate(i IShopActivate) {
	localShopActivate = i
}

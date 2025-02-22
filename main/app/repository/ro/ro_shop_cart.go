package ro

import "ttpos-server-go/app/model"

type ShopCartRepo struct {
	SaleBill *model.SaleBill
}

// 判断是否时桌台购物车。如果是桌台购物车，肯定desk不为nil
func (ro *ShopCartRepo) IsDeskShopCart() bool {
	return ro.SaleBill.Desk != nil
}

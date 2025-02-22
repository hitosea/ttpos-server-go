package ro

import "ttpos-server-go/app/model"

type ShopCartRepo struct {
	SaleBill *model.SaleBill
}

func (ro *ShopCartRepo) IsDeskShopCart() bool {
	if ro.SaleBill == nil {
		return false
	}
	if ro.SaleBill.Desk == nil {
		return false
	}
	return ro.SaleBill.Desk.Uuid != 0 // 如果是桌台购物车的话肯定会查询有桌台信息
}

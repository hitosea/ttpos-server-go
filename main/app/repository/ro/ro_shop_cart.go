package ro

import (
	"ttpos-server-go/app/model"
)

type ShopCartRepo struct {
	SaleBill *model.SaleBill
}

// 判断是否时桌台购物车。如果是桌台购物车，肯定desk不为nil
func (ro *ShopCartRepo) IsDeskShopCart() bool {
	if ro.SaleBill == nil {
		return false
	}
	if ro.SaleBill.Desk == nil {
		return false
	}
	return ro.SaleBill.Desk.Uuid != 0 // 如果是桌台购物车的话肯定会查询有桌台信息
}

// 获取购物车中的必点商品信息
func (ro *ShopCartRepo) GetMustPlanProductInfo() map[uint64]map[uint64]uint {
	// MustPlanUuid => ProductPackageUuid => num
	dataMap := make(map[uint64]map[uint64]uint)
	for _, saleOrder := range ro.SaleBill.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			mustPlanUuid := saleOrderProduct.MustPlanUuid
			if mustPlanUuid == 0 {
				continue
			}
			productPackageUuid := saleOrderProduct.ProductPackageUuid
			if num, ok := dataMap[mustPlanUuid][productPackageUuid]; ok {
				// 累加必点商品数量
				dataMap[mustPlanUuid][productPackageUuid] = num + saleOrderProduct.Num
			} else {
				dataMap[mustPlanUuid][productPackageUuid] = saleOrderProduct.Num
			}
		}
	}
	return dataMap
}

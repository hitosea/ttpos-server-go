package ro

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"
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

// 描述购物车中某个必点方案的某个商品已经选购了多少个
type MustPlanProductInfo map[uint64]map[uint64]uint // MustPlanUuid => ProductPackageUuid => num

// 获取购物车中的必点商品信息
func (ro *ShopCartRepo) GetMustPlanProductInfo() MustPlanProductInfo {
	// MustPlanUuid => ProductPackageUuid => num
	dataMap := make(map[uint64]map[uint64]uint)
	for _, saleOrder := range ro.SaleBill.SaleOrders {
		for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
			mustPlanUuid := saleOrderProduct.MustPlanUuid
			if mustPlanUuid == 0 {
				continue
			}
			productPackageUuid := saleOrderProduct.ProductPackageUuid
			if _, ok := dataMap[mustPlanUuid]; ok {
				if num, exist := dataMap[mustPlanUuid][productPackageUuid]; exist {
					// 累加必点商品数量
					dataMap[mustPlanUuid][productPackageUuid] = num + saleOrderProduct.Num
				} else {
					dataMap[mustPlanUuid][productPackageUuid] = saleOrderProduct.Num
				}
			} else {
				dataMap[mustPlanUuid] = make(map[uint64]uint)
				dataMap[mustPlanUuid][productPackageUuid] = saleOrderProduct.Num
			}
		}
	}
	fmt.Println("GetMustPlanProductInfo dataMap", utils.ToJsonString(dataMap))
	return dataMap
}

package utils

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
)

type OrderCheck struct {
	IsSoldOut            bool // 有商品下架或删除
	SoldOutProductList   []model.SaleOrderProduct
	IsStockZero          bool // 有商品库存不足
	StockZeroProductList []model.SaleOrderProduct
	IsPriceChange        bool // 有商品价格变动
	PriceChangeProduct   []model.SaleOrderProduct
	IsLimitOut           bool // 有商品超出限购
	LimitOutProduct      []model.SaleOrderProduct
	IsNotPassMustPlan    bool // 桌台未满足必点要求
	NotPassMustPlan      ProductMustPlanCheckResultTips
}

func (o *OrderCheck) Result() (bool, string, string) {
	if o.IsSoldOut {
		tips := `` // 提示信息
		return false, constant.ProductDown, tips
	}
	if o.IsStockZero {
		tips := ``
		return false, constant.ProductStockZero, tips
	}
	if o.IsPriceChange {
		tips := ``
		return false, constant.ProductPriceChange, tips
	}
	if o.IsLimitOut {
		tips := ``
		return false, constant.ProductLimitOut, tips
	}
	if o.IsNotPassMustPlan {
		tips := ``
		return false, constant.ProductMustPlanNotPass, tips
	}
	return true, constant.ProductPass, ``

}

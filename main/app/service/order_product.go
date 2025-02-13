package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/database"
)

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface {
	CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error                      // 检查商品规格
	CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error                     // 检查商品加料
	CheckProductOrderAttribute(productPackage model.ProductPackage, attributes []model.ProductAttribute) error // 检查商品属性
	CheckProductOrderFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error               // 检查商品规格库存
	CheckProductOrderSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error                // 检查商品加料库存
}

// orderProductSrv 订单商品服务结构体
type orderProductSrv struct {
	dbm *database.DBManager
}

// NewOrderProductSrv 创建商品服务
func NewOrderProductSrv(dbm *database.DBManager) IOrderProductSrv {
	return NewOrderProductSrvImpl(dbm)
}

// NewOrderProductSrvImpl 创建商品服务实现
func NewOrderProductSrvImpl(dbm *database.DBManager) IOrderProductSrv {
	return &orderProductSrv{
		dbm: dbm,
	}
}

func (o orderProductSrv) CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o orderProductSrv) CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o orderProductSrv) CheckProductOrderAttribute(productPackage model.ProductPackage, attributes []model.ProductAttribute) error {
	//TODO implement me
	panic("implement me")
}

func (o orderProductSrv) CheckProductOrderFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o orderProductSrv) CheckProductOrderSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

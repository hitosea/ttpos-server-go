package service

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface {
	CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error                      // 检查商品规格
	CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error                     // 检查商品加料
	CheckProductOrderAttribute(productPackage model.ProductPackage, attributes []model.ProductAttribute) error // 检查商品属性
	CheckProductOrderFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error               // 检查商品规格库存
	CheckProductOrderSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error                // 检查商品加料库存
	GetInvalidProductList(companyId uint64, saleOrderUuid uint64) ([]model.SaleOrderProduct, error)            // 获取所有已经失效的订单商品，检查销售订单的商品是否都是上架状态且未删除

	//CheckOderProductStock(productPackage model.ProductPackage) (bool, error)                                   // 检查订单商品库存是否都是
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

func (o *orderProductSrv) CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderProductSrv) CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderProductSrv) CheckProductOrderAttribute(productPackage model.ProductPackage, attributes []model.ProductAttribute) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderProductSrv) CheckProductOrderFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderProductSrv) CheckProductOrderSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

// CheckProductOrderBomStatus 检查销售订单的商品是否都是上架状态且未删除
func (o *orderProductSrv) GetInvalidProductList(companyId uint64, saleOrderUuid uint64) ([]model.SaleOrderProduct, error) {
	var invalidProductList []model.SaleOrderProduct
	// 查询销售订单商品组合表
	bomList, err := repository.NewOrderRepo(o.dbm.GetDB(companyId)).GetSaleOrderBomList(saleOrderUuid)
	if err != nil {
		return nil, err
	}
	for _, bom := range bomList {
		if bom.ProductBom.IsDelete() || bom.ProductBom.ProductPackage.IsDown() || bom.ProductBom.ProductPackage.IsDelete() {
			invalidProductList = append(invalidProductList, bom.SaleOrderProduct)
		}
	}
	return invalidProductList, nil
}

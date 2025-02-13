package service

import (
	"errors"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
)

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface {
	CheckProduct(dbId uint64, productUuid uint64) (*model.ProductPackage, error)                            // 检查商品
	CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error                   // 检查商品规格
	CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error                  // 检查商品加料
	CheckProductOrderAttribute(productPackage model.ProductPackage, attributeMap map[uint64][]uint64) error // 检查商品属性
	CheckProductOrderFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error            // 检查商品规格库存
	CheckProductOrderSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error             // 检查商品加料库存
	GetInvalidProductList(companyId uint64, saleOrderUuid uint64) ([]model.SaleOrderProduct, error)
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

// CheckProduct 检查商品
func (o *orderProductSrv) CheckProduct(dbId uint64, productUuid uint64) (*model.ProductPackage, error) {
	db := o.dbm.GetDB(dbId)
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackage, _ := productRepo.GetProduct(
		commonRepo.WhereByUuid(productUuid),
		commonRepo.WhereBySoftDelete(),
		productRepo.WithProductBoms(),
		productRepo.WithProductPackageAttributeGroup(),
		productRepo.WithProductPackageAttributeGroupProductPackageAttributes(),
	)
	if productPackage.Uuid == 0 {
		return nil, errors.New("商品不存在")
	}
	if productPackage.Status == constant.ProductStatusOffSale {
		return nil, errors.New("商品已下架")
	}

	return &productPackage, nil
}

// CheckProductOrderFlavor 检查商品规格
// productPackage需要包含ProductBoms
func (o orderProductSrv) CheckProductOrderFlavor(productPackage model.ProductPackage, flavorUuid uint64) error {
	if slices.ContainsFunc(productPackage.ProductBoms, func(productBom model.ProductBom) bool {
		return productBom.ProductFlavorUuid == flavorUuid
	}) {
		return nil
	}
	return errors.New("商品规格不存在")
}

// CheckProductOrderSauce 检查商品加料
// productPackage需要包含ProductBoms
func (o *orderProductSrv) CheckProductOrderSauce(productPackage model.ProductPackage, sauceUuids []uint64) error {
	var count = uint(len(sauceUuids))

	if productPackage.SauceRequired == 1 && count == 0 {
		return errors.New("商品加料不能为空")
	}
	if count > productPackage.SauceMaxSelection {
		return errors.New("商品加料超出最大可选数量")
	}

	for _, sauceUuid := range sauceUuids {
		if !slices.ContainsFunc(productPackage.ProductBoms, func(productBom model.ProductBom) bool {
			return productBom.ProductSauceUuid == sauceUuid
		}) {
			return errors.New("商品加料不存在")
		}
	}

	return nil
}

// CheckProductOrderAttribute 检查商品属性,
// productPackage需要包含ProductPackageAttributeGroups
// ProductPackageAttributeGroups需要包含ProductPackageAttributes
func (o *orderProductSrv) CheckProductOrderAttribute(productPackage model.ProductPackage, attributeMap map[uint64][]uint64) error {
	var groups = productPackage.ProductPackageAttributeGroups
	for _, group := range groups {
		var count = uint(len(attributeMap[group.Uuid]))
		if group.IsMust == 1 && count == 0 {
			return errors.New("商品属性不能为空")
		}
		if count > group.MaxSelection {
			return errors.New("商品属性超出最大可选数量")
		}
		for _, valueUuid := range attributeMap[group.Uuid] {
			if !slices.ContainsFunc(group.ProductPackageAttributes, func(attributeValue model.ProductPackageAttribute) bool {
				return attributeValue.Uuid == valueUuid
			}) {
				return errors.New("商品属性不存在")
			}
		}
	}

	return nil
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

package service

import (
	"errors"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"

	"github.com/shopspring/decimal"
)

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface {
	CheckProduct(dbId uint64, productUuid uint64) (model.ProductPackage, error)                             // 检查商品
	CheckOrderProductFlavor(productPackage model.ProductPackage, flavorUuid uint64) error                   // 检查商品规格
	CheckOrderProductSauce(productPackage model.ProductPackage, sauceUuids []uint64) error                  // 检查商品加料
	CheckOrderProductAttribute(productPackage model.ProductPackage, attributeMap map[uint64][]uint64) error // 检查商品属性
	CheckOrderProductFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error            // 检查商品规格库存
	CheckOrderProductSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error             // 检查商品加料库存
	GetInvalidProductList(companyId uint64, saleOrderUuid uint64) ([]model.SaleOrderProduct, error)
	//CheckOderProductStock(productPackage model.ProductPackage) (bool, error)                                   // 检查订单商品库存是否都是
	CreateOrderProduct(dbId uint64, req CreateOrderProductReq) (*model.SaleOrderProduct, error) // 创建订单商品
	CalcAmount(boms []model.ProductBom, num uint) CalcAmountResp                                // 计算单价,商品原价+小料价
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
func (o *orderProductSrv) CheckProduct(dbId uint64, productUuid uint64) (model.ProductPackage, error) {
	db := o.dbm.GetDB(dbId)
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackage, _ := productRepo.GetProduct(
		commonRepo.WhereByUuid(productUuid),
		commonRepo.WhereBySoftDelete(),
		productRepo.WithMultiLanguageName(),
		productRepo.WithProductBoms(),
		productRepo.WithProductBomsProductFlavor(),
		productRepo.WithProductBomsProductFlavorMultiLanguageName(),
		productRepo.WithProductBomsProductSauce(),
		productRepo.WithProductBomsProductSauceMultiLanguageName(),
		productRepo.WithProductPackageAttributeGroup(),
		productRepo.WithProductPackageAttributeGroupProductPackageAttributes(),
		productRepo.WithProductPackageAttributeGroupProductPackageAttributesAttribute(),
		productRepo.WithProductPackageAttributeGroupProductPackageAttributesAttributeMultiLanguageName(),
	)
	if productPackage.Uuid == 0 {
		return model.ProductPackage{}, errors.New("商品不存在")
	}
	if productPackage.Status == constant.ProductStatusOffSale {
		return model.ProductPackage{}, errors.New("商品已下架")
	}

	return productPackage, nil
}

// CheckOrderProductFlavor 检查商品规格
// productPackage需要包含ProductBoms
func (o *orderProductSrv) CheckOrderProductFlavor(productPackage model.ProductPackage, flavorUuid uint64) error {
	if slices.ContainsFunc(productPackage.ProductBoms, func(productBom model.ProductBom) bool {
		return productBom.ProductFlavorUuid == flavorUuid
	}) {
		return nil
	}

	return errors.New("商品规格不存在")
}

// CheckOrderProductSauce 检查商品加料
// productPackage需要包含ProductBoms
func (o *orderProductSrv) CheckOrderProductSauce(productPackage model.ProductPackage, sauceUuids []uint64) error {
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

// CheckOrderProductAttribute 检查商品属性,
// productPackage需要包含ProductPackageAttributeGroups
// ProductPackageAttributeGroups需要包含ProductPackageAttributes
func (o *orderProductSrv) CheckOrderProductAttribute(productPackage model.ProductPackage, attributeMap map[uint64][]uint64) error {
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

func (o *orderProductSrv) CheckOrderProductFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

func (o *orderProductSrv) CheckOrderProductSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error {
	//TODO implement me
	panic("implement me")
}

// GetInvalidProductList 检查销售订单的商品是否都是上架状态且未删除
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

// CreateOrderProductReq 创建订单商品请求
type CreateOrderProductReq struct {
	Lang           string
	SaleOrder      model.SaleOrder
	ProductPackage model.ProductPackage
	ProductFlavor  model.ProductFlavor
	ProductBoms    []model.ProductBom
	Num            uint
}

// CreateOrderProduct 创建订单商品
func (o *orderProductSrv) CreateOrderProduct(dbId uint64, req CreateOrderProductReq) (*model.SaleOrderProduct, error) {

	return nil, nil
}

type CalcAmountResp struct {
	UnitPrice      float64 // 单价: productBom.Price累加
	Price          float64 // 最终单价: 折扣和优惠后的UnitPrice
	ServiceFee     float64 // 服务费,按比例收取时有 todo
	TaxFee         float64 // 税费 todo
	OriginalAmount float64 // 原价销售额: (UnitPrice + TaxFee) * 数量
}

// CalcAmount 计算单价,商品原价+小料价
func (o *orderProductSrv) CalcAmount(boms []model.ProductBom, num uint) CalcAmountResp {
	var unitPrice float64
	var TaxFee float64
	var originalAmount float64
	for _, bom := range boms {
		unitPrice = decimal.NewFromFloat(unitPrice).Add(decimal.NewFromFloat(bom.Price)).InexactFloat64()
	}

	// 原价销售额 = (UnitPrice + TaxFee) * 数量
	originalAmount = decimal.NewFromFloat(unitPrice).Add(decimal.NewFromFloat(TaxFee)).Mul(decimal.NewFromInt(int64(num))).InexactFloat64()

	return CalcAmountResp{
		UnitPrice:      unitPrice,
		Price:          unitPrice,
		OriginalAmount: originalAmount,
	}
}

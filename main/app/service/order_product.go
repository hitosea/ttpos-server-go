package service

import (
	"errors"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// IOrderProductSrv 定义订单商品服务接口
type IOrderProductSrv interface {
	CheckProduct(dbId uint64, productUuid uint64) (model.ProductPackage, error)                                 // 检查商品
	CheckOrderProductFlavor(productPackage model.ProductPackage, flavorUuid uint64) error                       // 检查商品规格
	CheckOrderProductSauce(productPackage model.ProductPackage, sauceUuids []uint64) error                      // 检查商品加料
	CheckOrderProductAttribute(productPackage model.ProductPackage, attributes []req.AddProductAttribute) error // 检查商品属性
	CheckOrderProductFlavorStock(productPackage model.ProductPackage, sauceUuids []uint64) error                // 检查商品规格库存
	CheckOrderProductSauceStock(productPackage model.ProductPackage, sauceUuids []uint64) error                 // 检查商品加料库存
	GetInvalidProductList(companyId uint64, saleOrderUuid uint64) ([]model.SaleOrderProduct, error)
	//CheckOderProductStock(productPackage model.ProductPackage) (bool, error)                                   // 检查订单商品库存是否都是
	CheckCreateOrderProduct(dbId uint64, product req.AddProduct) (*model.ProductPackage, error) // 检查创建订单商品
	CreateOrderProduct(dbId uint64, req CreateOrderProductReq) error                            // 创建订单商品
	GenerateOrderProduct(req GenerateOrderProductReq) model.SaleOrderProduct                    // 生成订单商品
	UpdateOrderProductAmount(db *gorm.DB, req UpdateOrderProductAmountReq) error                // 更新订单商品金额
}

// orderProductSrv 订单商品服务结构体
type orderProductSrv struct {
	dbm          *database.DBManager
	orderCalcSrv IOrderCalcSrv
}

// NewOrderProductSrv 创建商品服务
func NewOrderProductSrv(dbm *database.DBManager, orderCalcSrv IOrderCalcSrv) IOrderProductSrv {
	return NewOrderProductSrvImpl(dbm, orderCalcSrv)
}

// NewOrderProductSrvImpl 创建商品服务实现
func NewOrderProductSrvImpl(dbm *database.DBManager, orderCalcSrv IOrderCalcSrv) IOrderProductSrv {
	return &orderProductSrv{
		dbm:          dbm,
		orderCalcSrv: orderCalcSrv,
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
		productRepo.WithDineTax(),
		productRepo.WithTakeoutTax(),
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
func (o *orderProductSrv) CheckOrderProductAttribute(productPackage model.ProductPackage, attributes []req.AddProductAttribute) error {
	var attributeMap = make(map[uint64][]uint64)
	for _, attribute := range attributes {
		attributeMap[attribute.GroupUuid] = append(attributeMap[attribute.GroupUuid], attribute.ValueUuids...)
	}
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

// CheckCreateOrderProduct 检查创建订单商品
func (o *orderProductSrv) CheckCreateOrderProduct(dbId uint64, product req.AddProduct) (*model.ProductPackage, error) {
	// 检查商品
	productPackage, err := o.CheckProduct(dbId, product.Uuid)
	if err != nil {
		return nil, err
	}
	// 检查商品规格
	if err = o.CheckOrderProductFlavor(productPackage, product.FlavorUuid); err != nil {
		return nil, err
	}
	// 检查商品属性
	if err = o.CheckOrderProductAttribute(productPackage, product.Attributes); err != nil {
		return nil, err
	}
	// 检查商品加料
	if err = o.CheckOrderProductSauce(productPackage, product.SauceUuids); err != nil {
		return nil, err
	}

	// todo 检查商品规格库存

	// todo 是否必选商品

	return &productPackage, nil
}

// CreateOrderProductReq 创建订单商品请求
type CreateOrderProductReq struct {
	Lang           string
	SaleBill       model.SaleBill
	SaleOrder      model.SaleOrder
	ProductPackage model.ProductPackage
	SauceUuids     []uint64
	Num            uint
}

// CreateOrderProduct 创建订单商品
func (o *orderProductSrv) CreateOrderProduct(dbId uint64, req CreateOrderProductReq) error {
	db := o.dbm.GetDB(dbId)
	err := db.Transaction(func(tx *gorm.DB) error {
		// 生成销售订单商品
		orderProductData := o.GenerateOrderProduct(GenerateOrderProductReq{
			Lang:           req.Lang,
			ProductPackage: req.ProductPackage,
			SaleBill:       req.SaleBill,
			SaleOrder:      req.SaleOrder,
			SauceUuids:     req.SauceUuids,
			Num:            req.Num,
		})

		// 判断销售订单商品签名是否存在, 存在则更新, 不存在则创建
		orderProduct, err := repository.NewOrderProductRepo(tx).GetProductInfo(
			repository.CommonRepo.WhereBySign(orderProductData.Sign),
			repository.CommonRepo.WhereBySaleBillUuid(orderProductData.SaleBillUuid),
			repository.CommonRepo.WhereBySaleOrderUuid(orderProductData.SaleOrderUuid),
			repository.CommonRepo.WhereBySoftDelete(),
		)
		if err != nil {
			return err
		}
		if orderProduct.Uuid == 0 {
			// 创建销售订单商品
			orderProduct, err = repository.NewOrderProductRepo(tx).Create(orderProductData)
			if err != nil {
				return err
			}
			// 创建销售订单商品bom
			for key, bom := range orderProductData.SaleOrderProductBoms {
				bom.SaleOrderProductUuid = orderProduct.Uuid
				orderProductData.SaleOrderProductBoms[key] = bom
			}
			if err := repository.NewOrderProductBomRepo(tx).CreateBatch(orderProductData.SaleOrderProductBoms); err != nil {
				return err
			}
			// 创建销售订单商品属性
			if len(orderProductData.SaleOrderProductAttributes) > 0 {
				for key, attribute := range orderProductData.SaleOrderProductAttributes {
					attribute.SaleOrderProductUuid = orderProduct.Uuid
					orderProductData.SaleOrderProductAttributes[key] = attribute
				}
				if err := repository.NewOrderProductAttributeRepo(tx).CreateBatch(orderProductData.SaleOrderProductAttributes); err != nil {
					return err
				}
			}
		} else {
			// 更新销售订单商品
			if err := repository.NewOrderProductRepo(tx).Update(
				map[string]interface{}{
					"num": repository.NewCommonRepo().IncrementNum(req.Num),
				},
				repository.NewCommonRepo().WhereByUuid(orderProduct.Uuid),
			); err != nil {
				return err
			}
		}

		// 计算销售订单商品相关金额
		err = o.UpdateOrderProductAmount(tx, UpdateOrderProductAmountReq{
			SaleBill:     req.SaleBill,
			SaleOrder:    req.SaleOrder,
			OrderProduct: orderProduct,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// GenerateOrderProductReq 生成订单商品请求
type GenerateOrderProductReq struct {
	Lang           string
	ProductPackage model.ProductPackage
	SaleBill       model.SaleBill
	SaleOrder      model.SaleOrder
	SauceUuids     []uint64
	Num            uint
}

// GenerateOrderProduct 生成订单商品
func (o *orderProductSrv) GenerateOrderProduct(req GenerateOrderProductReq) model.SaleOrderProduct {
	// 获取商品规格名称
	flavorName := ""
	flavor := req.ProductPackage.GetFlavor()
	if flavor.Uuid > 0 {
		flavorName = flavor.MultiLanguageName.GetNameByLang(req.Lang)
	}

	// 生成商品原始价格
	flavorPrice, saucePrice, productPrice, salePrice := req.ProductPackage.GenerateOriginalAmount(req.SauceUuids)

	// 获取消费税率
	taxRate := req.ProductPackage.DineTax.TaxRate
	if req.SaleBill.DiningMethod == constant.SaleBillDiningMethodTakeout {
		taxRate = req.ProductPackage.TakeoutTax.TaxRate
	}

	// 构建销售订单商品模型
	orderProduct := model.SaleOrderProduct{
		Name:                  req.ProductPackage.MultiLanguageName.GetNameByLang(req.Lang),
		FlavorName:            flavorName,
		Num:                   req.Num,
		FlavorPrice:           flavorPrice,
		SaucePrice:            saucePrice,
		ProductPrice:          productPrice,
		SalePrice:             salePrice,
		DeductStockType:       req.ProductPackage.DeductStockType,
		MultiLanguageNameUuid: req.ProductPackage.MultiLanguageNameUuid,
		ImageFileUuid:         req.ProductPackage.ImageFileUuid,
		ProductPackageUuid:    req.ProductPackage.Uuid,
		SaleBillUuid:          req.SaleBill.Uuid,
		SaleOrderUuid:         req.SaleOrder.Uuid,
		IsOpenMemberDiscount:  req.ProductPackage.OpenDiscount,
		TaxRate:               taxRate,
	}

	// 构建销售订单商品BOM
	var orderProductBoms []model.SaleOrderProductBom
	for _, bom := range req.ProductPackage.ProductBoms {
		var name string
		var isFlavorBom uint
		if bom.ProductFlavorUuid > 0 {
			name = bom.ProductFlavor.MultiLanguageName.GetNameByLang(req.Lang)
			isFlavorBom = 1
		}
		if bom.ProductSauceUuid > 0 {
			name = bom.ProductSauce.MultiLanguageName.GetNameByLang(req.Lang)
		}
		orderProductBoms = append(orderProductBoms, model.SaleOrderProductBom{
			Name:           name,
			Price:          bom.Price,
			IsFlavorBom:    isFlavorBom,
			SaleOrderUuid:  req.SaleOrder.Uuid,
			ProductBomUuid: bom.Uuid,
		})
	}

	// 构建销售订单商品属性
	var orderProductAttributes []model.SaleOrderProductAttribute
	for _, productPackageGroup := range req.ProductPackage.ProductPackageAttributeGroups {
		for _, productPackageAttribute := range productPackageGroup.ProductPackageAttributes {
			orderProductAttributes = append(orderProductAttributes, model.SaleOrderProductAttribute{
				Name:                 productPackageAttribute.Attribute.MultiLanguageName.GetNameByLang(req.Lang),
				SaleOrderUuid:        req.SaleOrder.Uuid,
				ProductAttributeUuid: productPackageAttribute.AttributeUuid,
			})
		}
	}

	orderProduct.SaleOrderProductBoms = orderProductBoms
	orderProduct.SaleOrderProductAttributes = orderProductAttributes
	orderProduct.Sign = orderProduct.GenerateProductSign()

	return orderProduct
}

// UpdateOrderProductAmountReq 更新订单商品金额请求
type UpdateOrderProductAmountReq struct {
	SaleBill     model.SaleBill
	SaleOrder    model.SaleOrder
	OrderProduct model.SaleOrderProduct
}

// UpdateOrderProductAmount 更新订单商品金额
func (o *orderProductSrv) UpdateOrderProductAmount(db *gorm.DB, req UpdateOrderProductAmountReq) error {
	orderProductAmount := o.orderCalcSrv.CalcOrderProductAmount(req.SaleBill, req.SaleOrder, req.OrderProduct)
	if err := repository.NewOrderProductRepo(db).Update(
		map[string]interface{}{
			"price":                     orderProductAmount.DiscountAmount.Price,
			"discount_fee":              orderProductAmount.DiscountAmount.DiscountFee,
			"member_discount_fee":       orderProductAmount.DiscountAmount.MemberDiscountFee,
			"custom_discount_fee":       orderProductAmount.DiscountAmount.CustomDiscountFee,
			"member_discount_rate":      orderProductAmount.DiscountAmount.MemberDiscountRate,
			"member_card_discount_rate": orderProductAmount.DiscountAmount.MemberCardDiscountRate,
			"custom_discount_rate":      orderProductAmount.DiscountAmount.CustomDiscountRate,
			"tax_fee":                   orderProductAmount.TaxAmount.TaxFee,
			"service_fee":               orderProductAmount.ServiceAmount.ServiceFee,
			"service_tax_fee":           orderProductAmount.ServiceAmount.ServiceTaxFee,
			"total_price":               orderProductAmount.TotalPrice,
		},
		repository.NewCommonRepo().WhereByUuid(req.OrderProduct.Uuid),
		repository.NewCommonRepo().WhereBySoftDelete(),
	); err != nil {
		return err
	}

	return nil
}

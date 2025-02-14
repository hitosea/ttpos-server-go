package service

import (
	"errors"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"

	"gorm.io/gorm"
)

// IInstantSrv 点餐订单服务接口
type IInstantSrv interface {
	CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error)                                                         // 创建点餐订单
	GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error)                       // 获取点餐订单详情
	AddProductToInstantOrder(dbId uint64, lang string, req req.InstantOrderAddProductReq) (*resp.GetInstantOrderInfoResp, error) // 添加商品
}

// instantSrv 点餐订单服务结构体
type instantSrv struct {
	dbm             *database.DBManager // 数据库管理器
	orderSrv        IOrderSrv           // 订单服务
	orderProductSrv IOrderProductSrv    // 订单商品服务
}

// NewInstantSrv 创建点餐订单服务
func NewInstantSrv(dbm *database.DBManager, orderSrv IOrderSrv, orderProductSrv IOrderProductSrv) IInstantSrv {
	return NewInstantSrvImpl(dbm, orderSrv, orderProductSrv)
}

// NewInstantSrvImpl 创建点餐订单服务实现
func NewInstantSrvImpl(dbm *database.DBManager, orderSrv IOrderSrv, orderProductSrv IOrderProductSrv) IInstantSrv {
	return &instantSrv{
		dbm:             dbm,
		orderSrv:        orderSrv,
		orderProductSrv: orderProductSrv,
	}
}

// CreateInstantOrder 创建点餐订单
func (s *instantSrv) CreateInstantOrder(dbId uint64) (resp.CreateInstantOrderResp, error) {
	return s.orderSrv.CreateInstantOrder(dbId)
}

func (s *instantSrv) GetInstantOrderInfo(dbId uint64, req req.InstantOrderGetInfoReq) (resp.GetInstantOrderInfoResp, error) {
	return resp.GetInstantOrderInfoResp{}, nil
}

// AddProductToInstantOrder 添加商品
func (s *instantSrv) AddProductToInstantOrder(dbId uint64, lang string, req req.InstantOrderAddProductReq) (*resp.GetInstantOrderInfoResp, error) {
	// 禁止并发操作
	lock.NewSystemLock().LockUuid(req.SaleBillUuid)
	defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)

	db := s.dbm.GetDB(dbId)

	// 验证参数
	if req.SaleBillUuid == 0 || req.SaleOrderUuid == 0 {
		return nil, errors.New("销售账单uuid或销售订单uuid不能为空")
	}
	if req.Product.Uuid == 0 {
		return nil, errors.New("商品uuid不能为空")
	}
	if req.Product.FlavorUuid == 0 {
		return nil, errors.New("商品规格uuid不能为空")
	}

	// 检查销售账单或销售订单
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillInfo(req.SaleBillUuid, req.SaleOrderUuid)
	if err != nil {
		return nil, errors.New("销售账单不存在")
	}
	if len(saleBill.SaleOrders) == 0 {
		return nil, errors.New("销售订单不存在")
	}
	// 检查订单是否可操作
	if err = saleBill.ValidateOrderStatus(constant.OrderAddProduct); err != nil {
		return nil, err
	}

	// 检查商品
	productPackage, err := s.orderProductSrv.CheckProduct(dbId, req.Product.Uuid)
	if err != nil {
		return nil, err
	}
	// 检查商品规格
	if err = s.orderProductSrv.CheckOrderProductFlavor(productPackage, req.Product.FlavorUuid); err != nil {
		return nil, err
	}
	// 检查商品属性
	var attributeMap = make(map[uint64][]uint64)
	for _, attribute := range req.Product.Attributes {
		attributeMap[attribute.GroupUuid] = append(attributeMap[attribute.GroupUuid], attribute.ValueUuids...)
	}
	if err = s.orderProductSrv.CheckOrderProductAttribute(productPackage, attributeMap); err != nil {
		return nil, err
	}
	// 检查商品加料
	if err = s.orderProductSrv.CheckOrderProductSauce(productPackage, req.Product.SauceUuids); err != nil {
		return nil, err
	}

	// todo 检查是否已选择必填商品

	// todo 检查商品规格库存

	err = db.Transaction(func(tx *gorm.DB) error {
		// 生成销售订单商品
		orderProductData := s.generateOrderProduct(productPackage, lang, req)

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
					"num": repository.NewCommonRepo().IncrementNum(1),
				},
				repository.NewCommonRepo().WhereByUuid(orderProduct.Uuid),
			); err != nil {
				return err
			}
		}

		// 计算销售订单商品相关金额

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &resp.GetInstantOrderInfoResp{}, nil
}

// generateOrderProduct 生成销售订单商品
func (s *instantSrv) generateOrderProduct(productPackage model.ProductPackage, lang string, req req.InstantOrderAddProductReq) model.SaleOrderProduct {
	// 获取商品规格名称
	flavorName := ""
	flavor := productPackage.GetFlavor()
	if flavor.Uuid > 0 {
		flavorName = flavor.MultiLanguageName.GetNameByLang(lang)
	}

	// 生成商品原始价格
	flavorPrice, saucePrice, productPrice, salePrice := productPackage.GenerateOriginalAmount(req.Product.SauceUuids)

	// 构建销售订单商品模型
	orderProduct := model.SaleOrderProduct{
		Name:                  productPackage.MultiLanguageName.GetNameByLang(lang),
		FlavorName:            flavorName,
		Num:                   1,
		FlavorPrice:           flavorPrice,
		SaucePrice:            saucePrice,
		ProductPrice:          productPrice,
		SalePrice:             salePrice,
		DeductStockType:       productPackage.DeductStockType,
		MultiLanguageNameUuid: productPackage.MultiLanguageNameUuid,
		ImageFileUuid:         productPackage.ImageFileUuid,
		ProductPackageUuid:    productPackage.Uuid,
		SaleBillUuid:          req.SaleBillUuid,
		SaleOrderUuid:         req.SaleOrderUuid,
	}

	// 构建销售订单商品BOM
	var orderProductBoms []model.SaleOrderProductBom
	for _, bom := range productPackage.ProductBoms {
		var name string
		var isFlavorBom uint
		if bom.ProductFlavorUuid > 0 {
			name = bom.ProductFlavor.MultiLanguageName.GetNameByLang(lang)
			isFlavorBom = 1
		}
		if bom.ProductSauceUuid > 0 {
			name = bom.ProductSauce.MultiLanguageName.GetNameByLang(lang)
		}
		orderProductBoms = append(orderProductBoms, model.SaleOrderProductBom{
			Name:           name,
			Price:          bom.Price,
			IsFlavorBom:    isFlavorBom,
			SaleOrderUuid:  req.SaleOrderUuid,
			ProductBomUuid: bom.Uuid,
		})
	}

	// 构建销售订单商品属性
	var orderProductAttributes []model.SaleOrderProductAttribute
	for _, productPackageGroup := range productPackage.ProductPackageAttributeGroups {
		for _, productPackageAttribute := range productPackageGroup.ProductPackageAttributes {
			orderProductAttributes = append(orderProductAttributes, model.SaleOrderProductAttribute{
				Name:                 productPackageAttribute.Attribute.MultiLanguageName.GetNameByLang(lang),
				SaleOrderUuid:        req.SaleOrderUuid,
				ProductAttributeUuid: productPackageAttribute.AttributeUuid,
			})
		}
	}

	orderProduct.SaleOrderProductBoms = orderProductBoms
	orderProduct.SaleOrderProductAttributes = orderProductAttributes
	orderProduct.Sign = orderProduct.GenerateProductSign()

	return orderProduct
}

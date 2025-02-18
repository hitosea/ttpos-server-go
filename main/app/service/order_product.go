package service

import (
	"errors"
	"fmt"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/utils"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/shopspring/decimal"

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
	GetInvalidProductList(ctx context.Context, saleOrderUuid uint64) ([]model.SaleOrderProduct, error)          // 检查销售订单的商品是否都是上架状态且未删除
	GetMustPlanRuleByDeskUuid(ctx context.Context, deskUuid uint64) (*utils.Rule, *utils.Check, error)          // 查询某个桌台的必点商品规则
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
	orderSrv     IOrderSrv
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

// 查询某个桌台的必点商品规则
func (o *orderProductSrv) GetDeskMustProductRule(ctx context.Context) (*utils.ProductMustPlanCheck, error) {
	// 获取到桌台的必点商品规则
	productMustPlanCheck := &utils.ProductMustPlanCheck{}
	return productMustPlanCheck, nil
}

// 检查桌台的商品是否已经满足必点规则
func (o *orderProductSrv) CheckDeskMustProductRule(ctx context.Context, deskUuid uint64) (bool, *utils.ProductMustPlanCheckResult, error) {
	// 获取到桌台的必点商品规则
	rule, err := o.GetDeskMustProductRule(ctx)
	if err != nil {
		return false, nil, err
	}
	result := rule.CheckResult()
	if result.IsPassed() {
		return true, nil, nil
	}
	return false, result, nil
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
	if !productPackage.IsUp() {
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
func (srv *orderProductSrv) GetInvalidProductList(ctx context.Context, saleOrderUuid uint64) ([]model.SaleOrderProduct, error) {
	companyUuid := ctx.GetCompanyUuid()

	companyId := ctx.GetDbId()
	var downOrSaleOutProductList []model.SaleOrderProduct    // 下架或沽清的商品
	var priceChangeProductList []model.SaleOrderProduct      // 价格变动的商品
	var invalidProductBomList []model.ProductBom             // 库存不足的ProductBom
	var exceedLimitProductPackageList []model.ProductPackage // 超过限购数量的ProductPackage

	productBomSubStockMap := make(map[uint64]float64)          // 各个ProductBom的需要消耗的库存
	productBomMap := make(map[uint64]model.ProductBom)         // 各个ProductBom的库存数
	productPackageMap := make(map[uint64]model.ProductPackage) // 我需要知道销售订单购买了哪些ProductPackage
	saleOrderProductUuidMap := make(map[uint64]bool)           // 记录该订单下单了哪些商品
	productPackageBuyNumMap := make(map[uint64]float64)        // 各个ProductPackage购买的数量

	// 查询销售订单商品组合表
	saleOrderProductBomList, err := repository.NewOrderRepo(srv.dbm.GetDB(companyId)).GetSaleOrderBomList(saleOrderUuid)
	if err != nil {
		return nil, err
	}
	// 得到该销售订单各个ProductBom的需要消耗的库存、各个ProductBom的库存
	for _, saleOrderProductBom := range saleOrderProductBomList {
		productBomUuid := saleOrderProductBom.ProductBomUuid
		num := float64(1) // 每个SaleOrderProduct的消耗ProductBom数量都是1
		pre := productBomSubStockMap[productBomUuid]
		// 累加
		productBomSubStockMap[productBomUuid] = decimal.NewFromFloat(pre).Add(decimal.NewFromFloat(num)).Round(4).InexactFloat64()

		// 记录该ProductBom的库存。仅记录一次，因为后面的库存数是一样的
		if _, ok := productBomMap[productBomUuid]; !ok {
			productBomMap[productBomUuid] = saleOrderProductBom.ProductBom
		}
		// 记录该订单商品购买了哪些ProductPackage
		if _, ok := productPackageMap[saleOrderProductBom.ProductBom.ProductPackage.Uuid]; !ok {
			productPackageMap[saleOrderProductBom.ProductBom.ProductPackage.Uuid] = saleOrderProductBom.ProductBom.ProductPackage
		}
		// 记录该订单下单了哪些商品
		if _, ok := saleOrderProductUuidMap[saleOrderProductBom.SaleOrderProduct.Uuid]; !ok {
			saleOrderProductUuidMap[saleOrderProductBom.SaleOrderProduct.Uuid] = true
		}
	}
	if len(productBomMap) == len(productBomSubStockMap) {
		return nil, errors.New("业务数据异常，请联系管理员")
	}
	// 判断各个ProductBom的库存是否足够
	for productBomUuid, _ := range productBomSubStockMap {
		// 如果库存不足
		productBom := productBomMap[productBomUuid]
		if productBom.StockNum < productBomSubStockMap[productBomUuid] {
			invalidProductBomList = append(invalidProductBomList, productBom) // 记录库存不足的ProductBom
		}
	}
	// 逐个订单商品判断
	for _, saleOrderProductBom := range saleOrderProductBomList {
		// 判断是否已经下架/软删除
		if saleOrderProductBom.ProductBom.ProductPackage.IsDown() || saleOrderProductBom.ProductBom.IsDown() {
			downOrSaleOutProductList = append(downOrSaleOutProductList, saleOrderProductBom.SaleOrderProduct)
		}
		// 判断sale_order_product价格是否有变动
		if saleOrderProductBom.Price != productBomMap[saleOrderProductBom.ProductBomUuid].Price {
			priceChangeProductList = append(priceChangeProductList, saleOrderProductBom.SaleOrderProduct) // 记录priceChangeProductList
		}
	}
	// 判断哪些ProductPackage超过了限购数量
	// 获取订单的订单商品列表
	saleOrderProductUuids := make([]uint64, 0)
	for saleOrderProductUuid, _ := range saleOrderProductUuidMap {
		saleOrderProductUuids = append(saleOrderProductUuids, saleOrderProductUuid)
	}
	orderProducts, err := repository.NewOrderRepo(srv.dbm.GetDB(companyId)).GetSaleOrderProductListBySaleOrderProductUuids(saleOrderProductUuids)
	if err != nil {
		return nil, err
	}
	for _, orderProduct := range orderProducts {
		// 累计该ProductPackage购买的数量
		pre := productPackageBuyNumMap[orderProduct.ProductPackageUuid]
		productPackageBuyNumMap[orderProduct.ProductPackageUuid] = decimal.NewFromFloat(pre).Add(decimal.NewFromFloat(float64(orderProduct.Num))).Round(4).InexactFloat64()
	}
	// 判断哪些ProductPackage超过了限购数量
	for productPackageUuid, buyNum := range productPackageBuyNumMap {
		if uint(buyNum) > productPackageMap[productPackageUuid].LimitNum {
			exceedLimitProductPackageList = append(exceedLimitProductPackageList, productPackageMap[productPackageUuid])
		}
	}
	// 判断必点商品是否满足
	isPass, resultTips, err := srv.GetMustProductStat(ctx, companyUuid)
	if err != nil {
		return nil, err
	}
	if !isPass {
		//todo 怎么返回
		fmt.Println(resultTips)
		return nil, nil
	}
	return downOrSaleOutProductList, nil

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

// GetRuleByDeskUuid 获取桌台的必点商品规则
func (o *orderProductSrv) GetMustPlanRuleByDeskUuid(ctx context.Context, deskUuid uint64) (*utils.Rule, *utils.Check, error) {
	// 查询桌台所属的区域
	companyUuid := ctx.GetCompanyUuid()
	desk, err := repository.NewDeskRepo(o.dbm.GetDB(companyUuid)).GetDeskAndSaleBillByDeskUuid(deskUuid)
	if err != nil {
		return nil, nil, err
	}
	// 桌台人数
	//mealNum := desk.SaleBill.MealNum

	regionUuid := desk.RegionUuid
	// 用regionUuid在product_must_plan_region表中查询得到该桌台关联的必选方案
	productMustPlanList, err := repository.NewProductMustPlanRepo(o.dbm.GetDB(companyUuid)).GetProductMustPlanByRegionUuid(regionUuid)
	if err != nil {
		return nil, nil, err
	}

	rule := utils.Rule{
		EachPersonProductPlan: make(map[uint64]map[uint64]uint),
		EachOrderProductPlan:  make(map[uint64]uint),
	}

	productMustPlanMap := make(map[uint64]model.ProductMustPlan) // 必点方案列表
	// product_must_plan_uuid -> product_package_uuid -> true  记录某个必点方案1有哪些商品
	//                        -> product_package_uuid -> true
	// product_must_plan_uuid -> product_package_uuid -> true  记录某个必点方案2有哪些商品
	//                        -> product_package_uuid -> true
	//                        -> product_package_uuid -> true
	//                        -> product_package_uuid -> true
	prodcutMustPlanProductMap := make(map[uint64]map[uint64]bool) // 必点方案商品列表

	// 遍历必选方案列表
	for _, productMustPlan := range productMustPlanList {
		// 判断必选方案是否开启
		if productMustPlan.Status == constant.ProductMustPlanStatusOn {
			// 判断必选方案是全选还是任选
			if productMustPlan.MustRule == constant.ProductMustPlanMustRuleAll {
				// 全选
				for _, productMustPlanItem := range productMustPlan.ProductMustPlanItem {
					num := uint(0)
					if productMustPlan.MustType == constant.ProductMustPlanMustTypeEachPerson {
						//num = mealNum
					} else if productMustPlan.MustType == constant.ProductMustPlanMustTypeEachOrder {
						num = 1
					}
					rule.EachPersonProductPlan[productMustPlan.Uuid][productMustPlanItem.ProductPackageUuid] = uint(num)
					prodcutMustPlanProductMap[productMustPlan.Uuid][productMustPlanItem.ProductPackageUuid] = true
				}
			} else if productMustPlan.MustRule == constant.ProductMustPlanMustRuleAny {
				for _, productMustPlanItem := range productMustPlan.ProductMustPlanItem {
					prodcutMustPlanProductMap[productMustPlan.Uuid][productMustPlanItem.ProductPackageUuid] = true
				}
				// 任选
				num := uint(0)
				if productMustPlan.MustType == constant.ProductMustPlanMustTypeEachPerson {
					//num = mealNum
				} else if productMustPlan.MustType == constant.ProductMustPlanMustTypeEachOrder {
					num = 1
				}
				rule.EachOrderProductPlan[productMustPlan.Uuid] = uint(num)
			}
			// 记录必点方案
			productMustPlanMap[productMustPlan.Uuid] = productMustPlan
		}
	}

	check, err := o.getMustPlanCheck(prodcutMustPlanProductMap, productMustPlanMap)
	if err != nil {
		return nil, nil, err
	}

	return &rule, check, nil
}

// 获取桌台的每个必点方案的点餐情况
func (o *orderProductSrv) getMustPlanCheck(prodcutMustPlanProductMap map[uint64]map[uint64]bool, productMustPlanMap map[uint64]model.ProductMustPlan) (*utils.Check, error) {

	check := utils.Check{
		PerProduct:         make(map[uint64]map[uint64]uint),
		CombinationProduct: make(map[uint64]uint),
	}

	// 获取销售订单商品列表
	var saleOrderProductList []model.SaleOrderProduct

	// 遍历销售订单商品列表
	for _, saleOrderProduct := range saleOrderProductList {
		// 判断销售订单商品是否属于某个必点方案
		productPackageUuid := saleOrderProduct.ProductPackageUuid
		for productMustPlanUuid, productPackageMap := range prodcutMustPlanProductMap {
			if _, ok := productPackageMap[productPackageUuid]; ok {
				// 判断是全选还是任选
				if productMustPlanMap[productMustPlanUuid].MustRule == constant.ProductMustPlanMustRuleAll {
					// 全选
					pre := check.PerProduct[productMustPlanUuid][productPackageUuid]
					check.PerProduct[productMustPlanUuid][productPackageUuid] = pre + saleOrderProduct.Num
				} else if productMustPlanMap[productMustPlanUuid].MustRule == constant.ProductMustPlanMustRuleAny {
					// 任选
					pre := check.CombinationProduct[productMustPlanUuid]
					check.CombinationProduct[productMustPlanUuid] = pre + saleOrderProduct.Num
				}
			}
		}
	}

	return &check, nil
}

// 获取桌台的必点商品统计，判断是否通过必点校验。若不通过校验，返回还需加购的数量
func (o *orderProductSrv) GetMustProductStat(ctx context.Context, deskUuid uint64) (bool, *utils.ProductMustPlanCheckResultTips, error) {
	rule, check, err := o.GetMustPlanRuleByDeskUuid(ctx, deskUuid)
	if err != nil {
		return false, nil, err
	}

	obj := utils.ProductMustPlanCheck{
		Rule:  *rule,
		Check: *check,
	}
	result := obj.CheckResult()

	resp, err := utils.Tips(o.dbm.GetDB(ctx.GetCompanyUuid()), result)
	if err != nil {
		return false, nil, err
	}

	return resp.IsPass(), resp, nil
}

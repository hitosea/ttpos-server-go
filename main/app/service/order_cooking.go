package service

import (
	"fmt"
	"sort"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"gorm.io/gorm"
)

// InstantOrderMustPlan 获取点餐必点方案
func (s *orderSrv) InstantOrderMustPlan(ctx context.Context, deviceSn string) (*resp.InstantProductMustPlanResp, bool, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 通过deviceSn获取saleBillUuid
	saleBillUuid, errUuid := s.getSaleBillUuidByDeviceSn(ctx)
	if errUuid != nil {
		return nil, false, errors.WithMessage(errUuid, "无法找到销售账单")
	}
	ctx.Log().Debug("查询必点方案列表", zap.Any("saleBillUuid", saleBillUuid), zap.Any("deviceSn", deviceSn))

	mustPlanList := make([]resp.InstantProductMustPlan, 0)
	// must_plan_uuid => autoFlavorProduct
	planAutoFlavorProduct := make(map[uint64]InstanceAutoFlavorProduct) // 必点方案ID => 自动加购的商品列表. 用于记录每个必点方案的自动加购商品

	// 查询到购物车信息
	shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(saleBillUuid)
	if err != nil {
		return nil, false, errors.WithMessage(err, "repository.NewOrderRepo(db).GetOrderCartInfo failed", fmt.Sprintf("saleBillUuid:%d", saleBillUuid))
	}

	planList, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
	if errMustPlan != nil {
		ctx.Log().Info("获取必点列表失败", zap.Error(errMustPlan))
		return nil, false, errors.New("获取必点列表失败")
	}
	mustPlanList = planList
	ctx.Log().Debug("构建好必点方案列表", zap.Any("数量", len(mustPlanList)))

	// 遍历得到要自动加购的商品
	for i, plan := range mustPlanList {
		// product_bom_uuid => *resp.InstantMustPlanProduct
		autoFlavorProduct := make(map[uint64]resp.ProductAutoAddReq) // 有自动加购的必选计划，且能自动加购的商品列表。要求只有一个规格，没有的商品才会自动加购
		for j, product := range plan.Products.List {
			if product.IsAutoAdd {
				planProduct := mustPlanList[i].Products.List[j].Product
				productFlavorBomUuid := planProduct.Flavors.List[0].Uuid
				productFlavorStockNum := planProduct.Flavors.List[0].StockNum
				if productFlavorStockNum > 0 {
					autoFlavorProduct[productFlavorBomUuid] = product.ProductAutoAddReq
				}
			}
		}
		if len(autoFlavorProduct) > 0 {
			planAutoFlavorProduct[plan.Uuid] = autoFlavorProduct
		}
	}

	var shopCart *resp.ShopCart
	var isAutoAdd bool
	// 判断是否需要给点餐账单自动加购商品。当map列表中有商品时，表示需要自动加购
	if len(planAutoFlavorProduct) > 0 && shopCartInfo.SaleBill.IsAutoAddMustProduct() {
		errTx := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
			// 通过上下文中的device_sn找到该收银机的点餐账单，若没有点餐账单则新建一个点餐账单并加购这些自动加购商品
			shopCart, err = autoAddSaleOrderProduct(ctx, tx, s, planAutoFlavorProduct)
			if err != nil {
				return errors.WithMessage(err, "自动添加必点商品失败")
			}
			// 设置已经完成自动加购
			if err := repository.NewSaleBillRepo(tx).UpdateSaleBillAutoAddMustProduct(saleBillUuid); err != nil {
				return errors.WithMessage(err, "标记自动加购完成失败")
			}
			return nil
		})
		if errTx != nil {
			return nil, false, errors.WithMessage(errTx, "自动添加必点商品失败")
		}
		isAutoAdd = true
	}

	var cartInfo *resp.InstantShopCart
	if shopCart != nil {
		cartInfo = &resp.InstantShopCart{
			SaleBillUuid:  shopCart.SaleBillUuid,
			DiningMethod:  shopCart.DiningMethod,
			SaleOrderList: shopCart.SaleOrderList,
		}
	}

	list := make([]resp.InstantProductMustPlan, 0)
	if shopCartInfo.SaleBill.IsShowMustPlan() {
		list = mustPlanList
	}
	mustPlan := &resp.InstantProductMustPlanResp{List: list, ShopCartInfo: cartInfo}
	return mustPlan, isAutoAdd, nil
}

// 通过itemCode获取订单中库存不足的商品名称
func (s *orderSrv) GetProductNameByItemCode(ctx context.Context, itemCode string, saleOrderUuid uint64) ([]ProductInfo, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取物品
	material, err := repository.NewMaterialRepo(db).GetMaterialByErpCode(itemCode)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取成本卡uuid列表
	productBomCardUuids, err := repository.NewRelatedMaterialRepo(db).GetProductBomCardUuidsByMaterialUuid(material.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 通过成本卡uuid列表获取product_bom列表
	productBomUuids, err := repository.NewProductBomRepo(db).GetFlavorProductBomUuidsByCardUuids(productBomCardUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 通过product_bom uuid列表查询sale_order_product_bom表获取sale_order_product_uuid列表
	saleOrderProductUuids, err := repository.NewSaleOrderProductRepo(db).GetSaleOrderProductUuidsByProductBomUuids(saleOrderUuid, productBomUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取商品信息
	products, err := repository.NewSaleOrderProductRepo(db).GetSaleOrderProductBySaleOrderProductUuids(saleOrderProductUuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	productInfos := make([]ProductInfo, 0)
	for _, product := range products {
		localeName := product.GetNameAndFlavorName()
		// 如果商品包没有多语言名称，则查询数据库获取
		if product.MultiLanguageName.IsNullName() {
			uuid := product.MultiLanguageNameUuid
			names, err := repository.NewMultiLanguageNameRepositoryImpl(db).GetMultiLanguageNameByUuid(uuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			bom, _ := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.SaleOrderProductBoms[0].ProductBomUuid)
			localeName = product.GetNameAndFlavorNameFrom(bom, &names)
		}
		productInfos = append(productInfos, ProductInfo{
			ProductName: localeName,
		})
	}

	return productInfos, nil
}

// InstantOrderMustPlanConfirm 确认必点商品
func (s *orderSrv) InstantOrderMustPlanConfirm(ctx context.Context, req req.InstantOrderMustPlanConfirmReq, opts ...func(option *MustPlanConfirmOption)) (bool, *resp.InstantProductMustPlan, error) {
	option := &MustPlanConfirmOption{}
	for _, opt := range opts {
		opt(option)
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	var mustPlanList []resp.InstantProductMustPlan

	// 点餐页面未创建销售账单时，但前端已显示出必点弹窗
	if req.SaleBillUuid == 0 {
		// 获取点餐的必选方案列表
		mustPlanLists, err := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, make(ro.MustPlanProductInfo))
		if err != nil {
			return false, nil, errors.WithMessage(err, "GetInstantMustPlanList failed")
		}
		if len(mustPlanLists) == 0 {
			return false, nil, errors.New("没有必点方案")
		}
		mustPlanList = mustPlanLists
	} else {
		// 获取销售账单信息
		saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
		if errSaleBill != nil {
			ctx.Log().Error("获取销售账单信息失败", zap.Error(errSaleBill))
			return false, nil, errors.WithMessage(errSaleBill, "获取销售账单信息失败")
		}

		// 查询到购物车信息
		var shopCartInfo *ro.ShopCartRepo
		var err error
		if option.IsH5Order {
			// 如果是h5订单时要查询出未下单的商品
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid, repository.WithNotDeleted())
		} else {
			shopCartInfo, err = repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
		}
		if err != nil {
			ctx.Log().Error("获取购物车信息失败", zap.Error(err))
			return false, nil, errors.WithMessage(err, "获取购物车信息失败")
		}

		if saleBill.IsDeskSaleBill() {
			mustPlan, errMustPlan := s.mustPlanSrv.GetDeskMustPlanList(ctx, shopCartInfo.SaleBill.MealNum, shopCartInfo.GetMustPlanProductInfo(), saleBill.DeskUuid)
			if errMustPlan != nil {
				return false, nil, errMustPlan
			}
			mustPlanList = mustPlan
		} else {
			mustPlan, errMustPlan := s.mustPlanSrv.GetInstantMustPlanList(ctx, db, shopCartInfo.GetMustPlanProductInfo())
			if errMustPlan != nil {
				return false, nil, errMustPlan
			}
			mustPlanList = mustPlan
		}
	}

	// 判断必点商品是否售罄
	if mustPlanList != nil && len(mustPlanList) > 0 {
		for _, plan := range mustPlanList {
			if plan.NeedNum > 0 {
				// 判断商品是否售罄。如果售罄，则允许"确认必点"
				isSoldOut, err := planProductSoldOut(ctx, &plan)
				if err != nil {
					return false, nil, errors.WithMessage(err)
				}
				if isSoldOut {
					break // 所有的未满足必点的商品都没有库存时，允许"确认必点"
				} else {
					ctx.Log().Info("确认必点商品失败，必点商品未点", zap.Any("plan name", plan.Name))
					return false, &plan, nil
				}
			}
		}
	}

	// 点餐订单需要创建
	if req.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return false, nil, errors.WithMessage(err)
		}
		if billInfo != nil && hasInstantOrder {
			req.SaleBillUuid = billInfo.Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return false, nil, errors.WithMessage(err)
			}
			req.SaleBillUuid = order.SaleBillUuid
		}
	}

	// 修改sale_bill表的show_must_plan
	if err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(req.SaleBillUuid); err != nil {
		return false, nil, errors.WithMessage(err, "确认必点商品失败")
	}

	return true, nil, nil
}

// OrderCheck 订单检查
func (s *orderSrv) OrderCheck(ctx context.Context, req req.InstantOrderCheckReq) (*resp.OrderCheckServiceRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	// 助手拆单之后不能操作
	if ctx.GetSource() == constant.SourceAssistant && len(saleBill.SaleOrders) > 1 {
		return nil, errors.NewWithCode(constant.CodeOrderCheckSplit, "当前订单已经拆单，请前去收银机操作")
	}

	ctx.Log().Debug("获取销售账单信息")

	// 从http的header中获取h5_order_uuid
	h5OrderProductUnAccept := make([]*model.SaleOrderProduct, 0) // 未接单的h5订单商品
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		h5OrderProductUnAccept = saleBill.GetH5OrderProductUnAccept(h5OrderUuid)
	}

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()

	// 获取所有商品,用于检查限购
	var saleOrderProductAll []*model.SaleOrderProduct
	if h5OrderUuid != 0 {
		// 如果从接到"进入桌台"时操作结账检查，要加入本h5订单的商品
		saleOrderProductAll = saleBill.GetSaleOrderProductAll(model.WithH5CheckLimit(), model.WithH5OrderUuid(h5OrderUuid))
	} else {
		saleOrderProductAll = saleBill.GetSaleOrderProductAll()
	}
	// 结账检查时，只检查未送厨的商品是否超过限购
	saleOrderProductUuids := make([]uint64, 0)
	for _, saleOrderProduct := range unCookingSaleOrderProducts {
		saleOrderProductUuids = append(saleOrderProductUuids, saleOrderProduct.Uuid)
	}
	for _, saleOrderProduct := range h5OrderProductUnAccept {
		saleOrderProductUuids = append(saleOrderProductUuids, saleOrderProduct.Uuid)
	}

	// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
	var deskUuid uint64
	if saleBill.IsDeskSaleBill() {
		deskUuid = saleBill.DeskUuid
	}

	var checkServiceRes *resp.OrderCheckServiceRes
	var errCheck error
	if h5OrderUuid != 0 {
		// 接单场景下，检查h5订单商品金额
		checkServiceRes, errCheck = s.checkOrder(ctx, req.IgnoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithSaleOrderProductUuid(saleOrderProductUuids...), WithH5OrderUuid(h5OrderUuid))
	} else {
		checkServiceRes, errCheck = s.checkOrder(ctx, req.IgnoreMust, db, req.SaleBillUuid, deskUuid, saleOrderProductAll, WithCheckTypeCheckout(), WithSaleOrderProductUuid(saleOrderProductUuids...))
	}
	if errCheck != nil {
		return nil, errors.WithMessage(errCheck, "订单检查失败")
	}
	if checkServiceRes != nil {
		return checkServiceRes, nil
	}

	// 检查自助餐顾客类型价格是否变动
	res := s.checkBuffetCustomerTypePriceChanged(ctx, saleBill)
	if res != nil {
		{
			// 如果价格变化，更新销售订单顾客的价格。都是后台更新价格而未立即更新已选购商品的价格引起的
			// 价格变化包括：
			// 1. 自助餐顾客类型价格变化
			shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			saleBill := shopCartInfo.SaleBill
			s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice())
		}
		return res, nil
	}

	// 检查含税未含税是否改变、服务费类型是否改变、服务费是否改变、服务费比例是否改变
	res, newSetting, err := s.checkSaleBillSettingChanged(ctx, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res != nil {
		{
			// 如果账单快照设置变化，更新销售订单的金额。都是后台更新后而未立即更新账单设置引起的
			// 账单快照设置变化包括：
			// 1. 含税未含税是否改变
			// 2. 服务费类型是否改变
			// 3. 固定服务费是否改变
			// 4. 服务费比例是否改变
			shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(req.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			saleBill := shopCartInfo.SaleBill
			s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting))
		}
		return res, nil
	}

	// 检查是否有未送厨的商品
	if len(unCookingSaleOrderProducts) > 0 || len(h5OrderProductUnAccept) > 0 {
		products := make([]resp.Product, 0)
		for _, product := range unCookingSaleOrderProducts {
			if product.IsPackageSubProduct() {
				continue
			}
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.DiscountFee,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsCancel:      product.IsCancelProduct(),
			})
		}
		for _, product := range h5OrderProductUnAccept {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.DiscountFee,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsCancel:      product.IsCancelProduct(),
			})
		}

		// 获取预送厨的商品
		preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
		for _, product := range preCookingSaleOrderProducts {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.DiscountFee,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsCancel:      product.IsCancelProduct(),
			})
		}

		res := &resp.OrderCheckServiceRes{
			Code:          constant.CodeOrderCheckProductUnCooking,
			OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: products}},
		}
		return res, nil
	} else {
		products := make([]resp.Product, 0)
		// 获取预送厨的商品
		preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
		for _, product := range preCookingSaleOrderProducts {
			products = append(products, resp.Product{
				Uuid:          product.Uuid,
				LocaleName:    product.GetNameAndFlavorName(),
				Num:           product.Num,
				SalePrice:     product.SalePrice,
				DiscountPrice: product.DiscountFee,
				Status:        int(product.Status),
				Remark:        product.Remark,
				IsMust:        product.IsMustProduct(),
				IsGift:        product.IsGiftProduct(),
				IsCancel:      product.IsCancelProduct(),
			})
		}
		if len(products) > 0 {
			res := &resp.OrderCheckServiceRes{
				Code:          constant.CodeOrderCheckProductUnCooking,
				OrderCheckRes: resp.OrderCheckRes{Products: &resp.CartProductList{List: products}},
			}
			return res, nil
		}
	}

	return nil, nil
}
func (s *orderSrv) CalcAndSaveSaleBill(ctx context.Context, db *gorm.DB, saleBill *model.SaleBill, options ...func(option *model.CalcOption)) error {
	// 保存到数据库
	if db == nil {
		db = s.dbm.GetDB(ctx.GetDbId())
	}
	return CalcAndSaveSaleBill(ctx, db, saleBill, options...)
}

// GetMustPlanList 点餐助手、平板端获取必点商品方案列表
func (s *orderSrv) GetMustPlanList(ctx context.Context, saleBillUuid uint64) (resp.ProductMustPlanList, error) {
	saleBillRepo := repository.NewSaleBillRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.ErrInternal
	}
	list, err := s.mustPlanSrv.GetDeskMustPlanList(ctx, saleBill.MealNum, make(map[uint64]map[uint64]float64), saleBill.DeskUuid)
	if err != nil {
		return resp.ProductMustPlanList{}, errors.WithMessage(err)
	}
	return resp.ProductMustPlanList{
		List: list,
	}, nil
}

// GetUnsentKitchen 未送厨商品列表
func (s *orderSrv) GetUnsentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart, opts ...repository.OrderCartInfoOptionFunc) (resp.UnsentKitchen, error) {
	// 初始化返回结果
	res := resp.UnsentKitchen{
		Products:   resp.CartProductList{List: make([]resp.Product, 0)},
		AmountInfo: resp.SimpleAmountInfo{},
	}

	if shopCart == nil {
		var err error
		// 获取购物车信息
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
		if err != nil {
			return res, errors.WithMessage(err, "获取点餐购物车信息: "+err.Error())
		}
	}

	// 按商品签名分组并合并相同商品
	signProduct := make(map[string]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			// 只处理未送厨且非赠菜的商品
			if product.Status != constant.SaleOrderProductStatusNormal {
				continue
			}
			if product.IsGift {
				product.DiscountPrice = 0
			}
			res.AmountInfo.ProductNum += product.Num
			// 合并相同商品的数量和折扣价格
			if p, exists := signProduct[product.Sign]; exists {
				product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
				product.Num = p.Num + product.Num
			}
			signProduct[product.Sign] = product
		}
	}

	// 将合并后的商品添加到结果列表
	res.Products.List = make([]resp.Product, 0, len(signProduct))
	for _, product := range signProduct {
		res.Products.List = append(res.Products.List, product)
	}

	// 按送厨时间排序
	sort.Slice(res.Products.List, func(i, j int) bool {
		if res.Products.List[i].CreateTime == res.Products.List[j].CreateTime {
			return res.Products.List[i].Uuid < res.Products.List[j].Uuid
		}
		return res.Products.List[i].CreateTime < res.Products.List[j].CreateTime
	})

	// 获取销售账单信息并计算未送厨商品总金额
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return res, errors.WithMessage(errors.New("获取销售账单所有信息"), err.Error())
	}

	for _, order := range saleBill.SaleOrders {
		res.AmountInfo.ProductAmount = utils.DecimalAdd(res.AmountInfo.ProductAmount, order.GetUnCookingProductAmount())
	}

	return res, nil
}

// GetSentKitchen 已送厨商品列表
func (s *orderSrv) GetSentKitchen(ctx context.Context, saleBillUuid uint64, shopCart *resp.ShopCart) (resp.SentKitchen, error) {
	if shopCart == nil {
		var err error
		shopCart, err = s.GetOrderCartInfo(ctx, saleBillUuid)
		if err != nil {
			return resp.SentKitchen{}, errors.WithMessage(err, "获取点餐购物车信息: "+err.Error())
		}
	}

	// 统计已送厨商品并按送厨时间分组
	productNum := float64(0)
	productGroup := make(map[int64]map[string]resp.Product)
	for _, saleOrder := range shopCart.SaleOrderList {
		for _, product := range saleOrder.ProductList {
			if product.Status != constant.SaleOrderProductStatusCooking {
				continue
			}

			productNum += product.Num

			// 按送厨时间分组,并在组内按商品签名合并
			if _, exists := productGroup[product.SendKitchenTime]; !exists {
				productGroup[product.SendKitchenTime] = make(map[string]resp.Product)
			}

			if p, exists := productGroup[product.SendKitchenTime][product.Sign]; exists {
				product.DiscountPrice = utils.DecimalAdd(p.DiscountPrice, product.DiscountPrice)
				product.Num = p.Num + product.Num
			}
			productGroup[product.SendKitchenTime][product.Sign] = product
		}
	}

	// 构建分组列表
	groups := make([]resp.SentKitchenProductGroup, 0, len(productGroup))
	for sendKitchenTime, signProducts := range productGroup {
		products := make([]resp.Product, 0, len(signProducts))
		for _, product := range signProducts {
			products = append(products, product)
		}

		groups = append(groups, resp.SentKitchenProductGroup{
			SendKitchenTime: sendKitchenTime,
			Products: resp.GroupProductList{
				List: products,
			},
		})
	}

	// 按送厨时间倒序排序
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SendKitchenTime > groups[j].SendKitchenTime
	})

	// 获取销售账单并计算金额
	saleBill, err := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return resp.SentKitchen{}, errors.WithMessage(errors.New("获取销售账单所有信息"), err.Error())
	}

	amount := resp.AmountInfo{ProductNum: productNum}
	for _, order := range saleBill.SaleOrders {
		calc := order.CalcCookingSaleOrder(*saleBill.SaleBillSetting)
		amount.ProductOriginalAmount = utils.DecimalAdd(amount.ProductOriginalAmount, calc.ProductOriginalAmount)
		amount.ProductAmount = utils.DecimalAdd(amount.ProductAmount, calc.ProductAmount)
		amount.ServiceAmount = utils.DecimalAdd(amount.ServiceAmount, calc.ServiceFee)
		amount.TaxAmount = utils.DecimalAdd(amount.TaxAmount, calc.TaxFee)
		amount.DiscountAmount = utils.DecimalAdd(amount.DiscountAmount, calc.CustomDiscountFee)
		amount.MemberDiscountAmount = utils.DecimalAdd(amount.MemberDiscountAmount, calc.MemberDiscountFee)
		amount.Amount = utils.DecimalAdd(amount.Amount, utils.IfFloat64(order.CustomAmount >= 0, order.CustomAmount, calc.Amount))
	}

	return resp.SentKitchen{
		Groups:     resp.GroupList{List: groups},
		AmountInfo: amount,
	}, nil
}
func (s *orderSrv) GetOrderCartProductBatchCookingList(ctx context.Context, req req.GetOrderCartProductBatchCookingListReq) (*resp.OrderCartProductBatchCookingRes, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	batchCookingSaleOrderProducts := saleBill.GetSaleOrderProductBatchCooking()

	batchCookingSaleOrderProductsList := make([]resp.OrderCartProductBatchCooking, 0)
	for _, saleOrderProduct := range batchCookingSaleOrderProducts {
		baseURL := utils.GetBaseURL(ctx.GetGin().Request)
		batchCookingSaleOrderProductsList = append(batchCookingSaleOrderProductsList, resp.OrderCartProductBatchCooking{
			Uuid:       saleOrderProduct.Uuid,
			LocaleName: saleOrderProduct.MultiLanguageName.GetNames(),
			// LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			LocaleAttributeName: saleOrderProduct.GetFlavorName(),
			Image: func() string {
				if saleOrderProduct.ImageFile != nil {
					url := saleOrderProduct.ImageFile.GetUrl(baseURL)
					return url
				}
				return ""
			}(),
			BatchTagUuid:    saleOrderProduct.BatchTagUuid,
			BatchTime:       saleOrderProduct.BatchTime,
			SendKitchenTime: saleOrderProduct.SendKitchenTime,
			CreateTime:      saleOrderProduct.CreateTime,
		})
	}

	tagMap := make(map[uint64]int)
	for _, batchCookingSaleOrderProduct := range batchCookingSaleOrderProducts {
		tagMap[batchCookingSaleOrderProduct.BatchTagUuid]++
	}

	// 获取分批类型列表
	batchTags, errBatchTags := repository.NewBatchTagRepo(db).GetBatchTagList()
	if errBatchTags != nil {
		return nil, errors.WithMessage(errBatchTags)
	}
	batchCookingSaleOrderProductsTags := make([]resp.OrderCartProductBatchCookingTag, 0)
	for _, batchTag := range batchTags {
		batchCookingSaleOrderProductsTags = append(batchCookingSaleOrderProductsTags, resp.OrderCartProductBatchCookingTag{
			Uuid:       batchTag.Uuid,
			LocaleName: batchTag.MultiLanguageName.GetNames(),
			Color:      batchTag.Color,
			Sort:       uint(batchTag.Sort),
			Count:      uint(tagMap[batchTag.Uuid]),
		})
	}

	// 排序
	batchCookingSaleOrderProductsList = sortBatchCookingSaleOrderProducts(batchCookingSaleOrderProductsList)

	return &resp.OrderCartProductBatchCookingRes{List: batchCookingSaleOrderProductsList, Tags: batchCookingSaleOrderProductsTags}, nil
}

// 分批送厨
func (s *orderSrv) OrderCartProductBatchCooking(ctx context.Context, req req.OrderCartProductBatchCookingReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	// 获取saleBill中的分批商品
	saleBillProducts := saleBill.GetSaleOrderProductBatchCookingBySaleOrderUuid(req.SaleOrderProductUuids)

	// 标记分批送厨
	batchTime := time.Now().Unix()
	for _, saleBillProduct := range saleBillProducts {
		saleBillProduct.SetCookingBatch(req.BatchTagUuid, batchTime)
	}

	// 更新saleBill中的最新分批类型的颜色。（每次分批送厨后，都改变一次）
	saleBill.BatchTagUuid = req.BatchTagUuid

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新销售订单商品状态从预送厨变为已送厨
		if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(saleBillProducts); err != nil {
			return errors.WithMessage(err)
		}
		// 更新送厨单商品的batch_time、batch_tag_uuid
		// 将product_order_product中的分批商品标记为已送厨
		if err := repository.NewProductionRepo(tx).UpdateProductionOrderProductBatchTimeAndBatchTagUuid(req.SaleBillUuid, req.SaleOrderProductUuids, batchTime, req.BatchTagUuid); err != nil {
			return errors.WithMessage(err)
		}
		// 更新销售账单中的分批类型UUID
		if err := repository.NewSaleBillRepo(tx).UpdateSaleBillBatchTagUuid(req.SaleBillUuid, req.BatchTagUuid); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发起“预送厨”操作的事件
	utils.Go(func() {
		s.bus.PublishSentCookingPreEvent(event.SentCookingPrePayload{
			BasePayload: event.BasePayload{ // 送厨
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  saleBill.Uuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Products: func() event.ProductsPre {
				products := make(event.ProductsPre, 0)
				for _, unCookingSaleOrderProduct := range saleBillProducts {
					unCookingSaleOrderProduct.BatchTagUuid = req.BatchTagUuid
					products = append(products, s.convertToEventOrderProductPre(
						unCookingSaleOrderProduct,
						saleBill,
					))
				}
				return products
			}(),
		})
	})
	shopCart, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return shopCart, nil
}

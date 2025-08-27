package service

import (
	"strconv"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/websocket"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ActionCookingOption struct {
	CalcAndSaveSaleBill      bool
	SelectedMustPlanProducts *ro.MustPlanProductInfo // 桌台已经选择的必点商品。使用场景仅用于平板加购并送厨时，将新加购的商品构建为该对象
	OnlyCheckCooking         bool                    // 是否是仅检查送厨，不进行实际送厨。场景：助手端开启下单校验高级密码时，先检查送厨，再实际送厨。检查送厨时不进行实际送厨
}

func withCalcAndSaveSaleBill() func(option *ActionCookingOption) {
	return func(option *ActionCookingOption) {
		option.CalcAndSaveSaleBill = true
	}
}

func WithSelectedMustPlanProductsActionCookingOption(selectedMustPlanProducts *ro.MustPlanProductInfo) func(option *ActionCookingOption) {
	return func(option *ActionCookingOption) {
		option.SelectedMustPlanProducts = selectedMustPlanProducts
	}
}

func WithOnlyCheckCooking() func(option *ActionCookingOption) {
	return func(option *ActionCookingOption) {
		option.OnlyCheckCooking = true
	}
}

// 转换器
func (s *orderSrv) convertToEventOrderProduct(saleOrderProduct *model.SaleOrderProduct, saleBill *model.SaleBill, saleOrder *model.SaleOrder) event.OrderProduct {
	orderProduct := event.OrderProduct{
		OrderProductId:        saleOrderProduct.Uuid,
		ProductId:             saleOrderProduct.ProductPackageUuid,
		ProductName:           saleOrderProduct.MultiLanguageName.GetNames(),
		ProductAttr:           saleOrderProduct.GetAttributeName(),
		ProductAttrList:       saleOrderProduct.GetAttributeNameList(),
		ProductSauceNamesList: saleOrderProduct.GetSauceNamesList(),
		TotalNum:              saleOrderProduct.Num,
		NumType:               saleOrderProduct.NumType,
		IsBuffet:              saleOrderProduct.IsBuffet == 1,
		IsWrap:                saleOrderProduct.CalculateIsWrap(saleBill),
		Remark:                saleOrderProduct.Remark,
	}

	// 如果是套餐主商品，添加子商品
	if saleOrderProduct.IsPackageProduct() && saleOrder != nil {
		subProducts := make([]event.OrderProduct, 0)
		for _, subProduct := range saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid) {
			subProducts = append(subProducts, s.convertToEventOrderProduct(subProduct, saleBill, nil))
		}
		orderProduct.SubProducts = subProducts
	}

	return orderProduct
}

// ActionCooking 送厨
func (s *orderSrv) ActionCooking(ctx context.Context, ignoreMust bool, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct, h5OrderUuid uint64, isAutoOrder bool, options ...func(option *ActionCookingOption)) (*resp.OrderCheckServiceRes, error) {
	option := &ActionCookingOption{}
	for _, opt := range options {
		opt(option)
	}
	var productionOrder *model.ProductionOrder
	var warehouseOutForms []*model.WarehouseOutForm

	if ctx.NoLock() {
		s.lock.LockUuid(saleBill.Uuid)
		defer s.lock.UnlockUuid(saleBill.Uuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	if len(unCookingSaleOrderProducts) == 0 {
		return nil, errors.New("没有未送厨的商品")
	}
	saleOrderUuid := unCookingSaleOrderProducts[0].SaleOrderUuid

	// 送厨相关
	{
		// 获取所有商品,用于检查限购
		var saleOrderProductAll []*model.SaleOrderProduct
		// 如果是接单场景
		if h5OrderUuid != 0 {
			saleOrderProductAll = saleBill.GetSaleOrderProductAll(model.WithH5CheckLimit(), model.WithH5OrderUuid(h5OrderUuid))
		} else {
			saleOrderProductAll = saleBill.GetSaleOrderProductAll()
		}

		// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
		var checkServiceRes *resp.OrderCheckServiceRes
		var errCheck error
		if option.SelectedMustPlanProducts != nil {
			// 平板端加购并送厨时，将平板加购的商品注入到checkOrder中
			checkServiceRes, errCheck = s.checkOrder(ctx, ignoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithSelectedMustPlanProducts(option.SelectedMustPlanProducts))
		} else {
			// 限购检查只检查未送厨的商品
			uuids := make([]uint64, 0)
			for _, saleOrderProduct := range unCookingSaleOrderProducts {
				uuids = append(uuids, saleOrderProduct.Uuid)
			}
			if h5OrderUuid != 0 {
				// 接单场景下，检查h5订单商品金额
				checkServiceRes, errCheck = s.checkOrder(ctx, ignoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithSaleOrderProductUuid(uuids...), WithH5OrderUuid(h5OrderUuid))
			} else {
				if ctx.GetScene() == constant.SceneMemberOrder {
					// 会员端订单不检查
				} else {
					checkServiceRes, errCheck = s.checkOrder(ctx, ignoreMust, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking(), WithSaleOrderProductUuid(uuids...))
				}
			}
		}
		if errCheck != nil {
			ctx.Log().Error("检查商品失败", zap.Error(errCheck))
			return nil, errors.WithMessage(errCheck)
		}
		if checkServiceRes != nil {
			if checkServiceRes.Code == constant.CodeOrderCheckProductMust && ignoreMust {
				// 必点方案未选择，且忽略必点方案
			} else {
				return checkServiceRes, nil
			}
		}

		// 如果只检查送厨，则在此直接返回return nil, nil
		if option.OnlyCheckCooking {
			return nil, nil
		}

		// 构建送厨单
		productionOrder = newProductionOrder(ctx, saleOrderUuid, saleBill.Uuid, saleBill.DeskUuid, unCookingSaleOrderProducts)

		// 修改商品状态为已送厨
		for _, product := range unCookingSaleOrderProducts {
			product.SetCooking(productionOrder.Uuid)
		}

		// 修改账单状态为已送厨
		if !saleBill.IsCookingStatus() {
			saleBill.SetCookingStatus()
		}
	}

	// 出库相关
	{
		// 获取减库存的清单信息
		decreaseStockList, err := s.GetProductDecreaseStockList(ctx, unCookingSaleOrderProducts)
		if err != nil {
			return nil, errors.WithMessage(err, "s.GetProductDecreaseStockList failed")
		}
		// 构建出库单
		warehouseOutForms = model.NewWarehouseOutForm(decreaseStockList, false, saleBill.Uuid, ctx.GetStaffUuid())
	}

	ctx.Log().Debug("准备开始更新")
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 计算账单. 只有加购并送厨时，才计算账单
		if option.CalcAndSaveSaleBill {
			// 注意要先执行CalcAndSaveSaleBill，因为saleOrder中可能有要新建的商品
			if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 送厨相关
		{
			// 修改订单商品状态为已送厨
			if errUpdateSaleProductStatus := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(unCookingSaleOrderProducts); errUpdateSaleProductStatus != nil {
				ctx.Log().Debug("商品状态更新失败", zap.Error(errUpdateSaleProductStatus))
				return errors.WithMessage(errUpdateSaleProductStatus, "商品状态更新失败")
			}
			ctx.Log().Debug("商品状态成功")
			if errCreateProduction := repository.NewProductionRepo(tx).CreateProductionOrder(productionOrder); errCreateProduction != nil {
				ctx.Log().Debug("创建送厨单失败", zap.Error(errCreateProduction))
				return errors.WithMessage(errCreateProduction, "创建送厨单失败")
			}

			// 如果账单有更新，则更新账单
			if saleBill.GetUpdate() {
				if err := repository.NewSaleBillRepo(tx).UpdateSaleBillRecord(*saleBill); err != nil {
					ctx.Log().Debug("更新账单失败", zap.Error(err))
					return errors.WithMessage(err, "更新账单失败")
				}
			} else {
				go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_ORDER, map[string]interface{}{
					"sale_bill_uuid": saleBill.Uuid,
					"desk_uuid":      saleBill.DeskUuid,
					"update_time":    saleBill.UpdateTime,
				})
			}

		}
		// 出库相关
		{
			for _, warehouseOutForm := range warehouseOutForms {
				// 如果出库单明细不为空，则创建出库单
				if len(warehouseOutForm.WarehouseOutFormItems) > 0 {
					// 创建出库单
					if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormRecord(*warehouseOutForm); err != nil {
						return errors.WithMessage(err)
					}
					// 创建出库单记录
					if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseOutFormItemRecords(warehouseOutForm.WarehouseOutFormItems); err != nil {
						return errors.WithMessage(err)
					}
				}
			}
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新数据失败")
	}

	// 送厨成功后，重新获取账单信息。为了补全新加购商品中缺失的信息
	if option.CalcAndSaveSaleBill { // 只有加购并送厨时，才需要补全新加购商品中缺失的信息
		saleBill, _ = repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(saleBill.Uuid)
		uuids := make(map[uint64]bool)
		for _, saleOrderProduct := range unCookingSaleOrderProducts {
			uuids[saleOrderProduct.Uuid] = true
		}
		unCookingSaleOrderProducts = saleBill.GetSaleOrderProductUnCookingByUuids(uuids)
	}

	// 操作记录相关
	{
		// 接单后送厨，需要先发布接单事件
		if h5OrderUuid > 0 {
			s.bus.PublishAcceptH5OrderEvent(event.AcceptH5OrderPayload{
				BasePayload: event.BasePayload{ // 接单
					Ctx:          ctx,
					CompanyUuid:  ctx.GetCompanyUuid(),
					Source:       ctx.GetSource(),
					SaleBillUuid: saleBill.Uuid,
					H5OrderUuid:  h5OrderUuid,
					OperatorUuid: int64(ctx.GetStaffUuid()),
				},
				IsAutoOrder: isAutoOrder,
			})
		}

		// 发起“送厨”操作的事件
		go s.bus.PublishSentCookingEvent(event.SentCookingPayload{
			BasePayload: event.BasePayload{ // 送厨
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  saleBill.Uuid,
				SaleOrderUuid: saleOrderUuid,
				H5OrderUuid:   h5OrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			Products: func() event.Products {
				products := make(event.Products, 0)
				for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {
					// 套餐子商品不显示送厨记录
					if unCookingSaleOrderProduct.IsPackageSubProduct() {
						continue
					}
					products = append(products, s.convertToEventOrderProduct(
						unCookingSaleOrderProduct,
						saleBill,
						saleBill.GetSaleOrder(saleOrderUuid),
					))
				}
				return products
			}(),
		})
		// 送厨成功后，推送更新订单
		go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
			"update_time": time.Now().Unix(),
		})
	}
	return nil, nil
}

// ActionAdd 加购
func (s *orderSrv) ActionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill) error {
	db := ctx.GetDB()

	var err error
	if request.IsMemberAdd {
		saleBill, err = s.actionAdd(ctx, request, saleBill, WithIsMemberAdd())
		if err != nil {
			return errors.WithMessage(err)
		}
	} else {
		saleBill, err = s.actionAdd(ctx, request, saleBill)
		if err != nil {
			return errors.WithMessage(err)
		}
	}

	if saleBill.IsTakeout() {
		// 如果是打包订单，则需要更新所有商品为打包商品
		saleBill.SetTakeoutSaleBill(constant.SaleBillDiningMethodTakeout)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// ActionAddAndCooking 加购并送厨
func (s *orderSrv) ActionAddAndCooking(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill, ignoreMust bool) (*resp.OrderCheckServiceRes, error) {

	// 加购相关
	_, err := s.actionAdd(ctx, request, saleBill, WithIsTableAdd(), WithSkipLimit())
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// MustPlanUuid => ProductPackageUuid => num
	selectedMustPlanProducts := make(ro.MustPlanProductInfo)
	for _, product := range request.Products {
		if selectedMustPlanProducts[product.MustPlanUuid] == nil {
			selectedMustPlanProducts[product.MustPlanUuid] = make(map[uint64]float64)
		}
		flavorProductBom, err := repository.NewProductBomRepo(ctx.GetDB()).GetFlavorProductBomByUuid(product.FlavorProductBomUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if _, ok := selectedMustPlanProducts[product.MustPlanUuid][flavorProductBom.ProductPackageUuid]; ok {
			selectedMustPlanProducts[product.MustPlanUuid][flavorProductBom.ProductPackageUuid] += product.Num // 如果已经存在该商品，则累计增加数量
		} else {
			selectedMustPlanProducts[product.MustPlanUuid][flavorProductBom.ProductPackageUuid] = product.Num // 如果不存在该商品，则新增
		}
	}
	// 送厨相关
	{
		// 获取未送厨的商品列表
		unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()
		// 只送厨本次加购的商品。排除掉其他端未送厨的商品
		unCookingSaleOrderProducts = filterOtherClientProducts(unCookingSaleOrderProducts)
		if len(unCookingSaleOrderProducts) == 0 {
			return nil, errors.New("没有未送厨的商品")
		}

		// 送厨
		checkServiceRes, err := s.ActionCooking(ctx, ignoreMust, saleBill, unCookingSaleOrderProducts, 0, false, withCalcAndSaveSaleBill(), WithSelectedMustPlanProductsActionCookingOption(&selectedMustPlanProducts)) // 平板端加购并送厨
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if checkServiceRes != nil {
			return checkServiceRes, nil
		}
	}

	return nil, nil
}

// 只送厨本次加购的商品。排除掉其他端未送厨的商品
func filterOtherClientProducts(unCookingSaleOrderProducts []*model.SaleOrderProduct) []*model.SaleOrderProduct {
	products := make([]*model.SaleOrderProduct, 0)
	for _, saleOrderProduct := range unCookingSaleOrderProducts {
		// 判断是否是在平板端新加购的商品
		isAddProduct := saleOrderProduct.CreateTime == 0 // 其他端的未送厨商品都是有创建时间的，只有平板端新加购的商品没有创建时间
		if isAddProduct {
			products = append(products, saleOrderProduct)
		}
	}
	return products
}

type TabletAddAndCookingRes struct {
	resp.OrderCheckRes
	Value           *int64             `json:"value"`             // 限制时间或限制数量。当触发错误：限制时间或限制数量时，返回限制时间或限制数量。当不限制时，返回nil
	NewProductInfos NewProductInfoList `json:"new_product_infos"` // 本次加购的商品列表的最新商品信息，结构与商品列表接口的商品结构相同
}

type NewProductInfoList struct {
	List []*product_resp.Product `json:"list"`
}

// 判断价格是否有变动。判断前端计算的单价是否与后台设置的最新价格一致。若不一致，则返回最新价格。（不考虑打折）
func (s *orderSrv) getInfo(ctx context.Context, product req.ProductParams, db *gorm.DB) (*product_resp.Product, error) {
	uuids := make([]uint64, 0)
	uuids = append(uuids, product.FlavorProductBomUuid)
	uuids = append(uuids, product.SauceProductBomUuidList...)
	productBoms, err := repository.NewProductBomRepo(db).GetProductBomsByUuids(uuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	lastestPrice := decimal.NewFromFloat(0)
	for _, productBom := range productBoms {
		// 如果是自助餐商品，则不记商品价格
		if product.IsBuffet != nil && *product.IsBuffet {
			if productBom.Uuid == product.FlavorProductBomUuid {
				continue
			}
		}
		lastestPrice = lastestPrice.Add(decimal.NewFromFloat(productBom.Price))
	}
	if lastestPrice.Cmp(decimal.NewFromFloat(*product.Price)) != 0 {
		// 获取最新的商品详情
		productPackageUuid := productBoms[0].ProductPackageUuid
		productInfo, err := s.GetProductDetail(ctx, productPackageUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		// 如果是自助餐商品，则价格改为0元
		if product.IsBuffet != nil && *product.IsBuffet {
			productInfo.Price = 0
			for index, _ := range productInfo.Flavors.List {
				productInfo.Flavors.List[index].Price = 0
			}
		}
		return &productInfo, nil
	}
	return nil, nil
}

// TabletAddAndCooking 平板端加购并送厨
func (s *orderSrv) TabletAddAndCooking(ctx context.Context, request req.TabletOrderCartProductAddReq) (*TabletAddAndCookingRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	// 判断商品价格是否与后台设置的最新价格不一致
	// 查询商品规格的最新价格
	// 查询所选加料的最新价格
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	productInfos := make([]*product_resp.Product, 0)
	for _, product := range request.Products {
		if product.Price != nil {
			if product.ProductType == constant.ProductTypePackage {
				continue
			}
			productInfo, err := s.getInfo(ctx, product, db)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			if productInfo != nil {
				productInfos = append(productInfos, productInfo)
			}
		}
	}
	if len(productInfos) > 0 {
		return &TabletAddAndCookingRes{
			NewProductInfos: NewProductInfoList{List: productInfos},
		}, errors.ErrProductPriceChanged
	}

	// 如果加购并送厨中的商品中有必点方案uuid，说明用户已经在平板端完成了必点，则关闭该销售账单的必点弹窗
	if request.Products != nil {
		for _, product := range request.Products {
			if product.MustPlanUuid != 0 {
				if err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(request.SaleBillUuid); err != nil {
					return nil, errors.WithMessage(err)
				}
				// 设置已经完成自动加购
				if err := repository.NewSaleBillRepo(db).UpdateSaleBillAutoAddMustProduct(request.SaleBillUuid); err != nil {
					return nil, errors.WithMessage(err, "标记自动加购完成失败")
				}
				break
			}
		}
	}

	saleBill, _ := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if saleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeDeskOrderEnd, "桌台订单结束"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	getProductNum := func(products []req.ProductParams) float64 {
		num := decimal.NewFromFloat(0)
		for _, product := range products {
			num = num.Add(decimal.NewFromFloat(product.Num))
		}
		return num.Truncate(2).InexactFloat64()
	}
	// 平板端下单限制
	tabletSetting, _ := s.settingSrv.GetTabletSetting(ctx, []dto.LanguageItem{})
	if tabletSetting.IsBuffetOrderLimit == "1" || tabletSetting.IsOrderLimit == "1" {
		// 读取上一单下单时间
		productionRepo := repository.NewProductionRepo(db)
		productionOrder, _ := productionRepo.GetProductionOrder(
			productionRepo.WhereSaleBillUuid(request.SaleBillUuid), productionRepo.WhereSource(ctx.GetSource()))
		var lastOrderTime int64 = 0
		if productionOrder != nil {
			lastOrderTime = productionOrder.CreateTime
		}
		if tabletSetting.IsBuffetOrderLimit == "1" && saleBill.IsBuffetSaleBill() { // 自助餐下单限制
			if tabletSetting.BuffetOrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(tabletSetting.BuffetOrderLimit.LimitTime)
				if err != nil {
					return nil, errors.WithMessage(err, "解析平板端设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					value := nextTime - now
					return &TabletAddAndCookingRes{Value: &value}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if tabletSetting.BuffetOrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(tabletSetting.BuffetOrderLimit.LimitNum)
				if err != nil {
					return nil, errors.WithMessage(err, "解析平板端设置失败")
				}
				if getProductNum(request.Products) > float64(numLimit) {
					value := int64(numLimit)
					return &TabletAddAndCookingRes{Value: &value}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
		if tabletSetting.IsOrderLimit == "1" && !saleBill.IsBuffetSaleBill() { // 非自助餐下单限制
			if tabletSetting.OrderLimit.IsLimitTime == "1" && lastOrderTime > 0 { // 限制下单间隔
				interval, err := strconv.Atoi(tabletSetting.OrderLimit.LimitTime)
				if err != nil {
					return nil, errors.WithMessage(err, "解析平板端设置失败")
				}
				// 小于间隔时间，不可下单
				nextTime := time.Unix(lastOrderTime, 0).Add(time.Duration(interval) * time.Minute).Unix()
				now := time.Now().Unix()
				if nextTime-now > 0 {
					value := nextTime - now
					return &TabletAddAndCookingRes{Value: &value}, errors.NewWithCode(constant.CodeH5OrderTimeLimit, "时间限制")
				}
			}
			if tabletSetting.OrderLimit.IsLimitNum == "1" { // 限制下单最大商品总数
				numLimit, err := strconv.Atoi(tabletSetting.OrderLimit.LimitNum)
				if err != nil {
					return nil, errors.WithMessage(err, "解析平板端设置失败")
				}
				if getProductNum(request.Products) > float64(numLimit) {
					value := int64(numLimit)
					return &TabletAddAndCookingRes{Value: &value}, errors.NewWithCode(constant.CodeH5OrderNumLimit, "数量限制")
				}
			}
		}
	}

	// 暂时不使用这两个字段
	for index, _ := range request.Products {
		request.Products[index].Price = nil
		request.Products[index].IsBuffet = nil
	}

	for index, _ := range request.Products {
		if request.Products[index].ProductType == constant.ProductTypePackage {
			request.Products[index].FlavorProductBomUuid = request.Products[index].ProductPackageUuid // 套餐商品规格uuid改为套餐商品uuid
		}
	}

	// 记录相关的子商品。
	for index, _ := range request.Products {
		productReq := request.Products[index]
		// 如果该商品是套餐的话,添加子商品的属性
		if productReq.ProductType == 1 {
			subProductParams := make([]req.ProductParams, 0)
			for _, subProductParam := range productReq.Products {
				params := req.ProductParams{
					FlavorProductBomUuid:            subProductParam.EditProductReq.FlavorUuid,
					Num:                             subProductParam.Num,
					ProductPackageAttributeUuidList: subProductParam.EditProductReq.AttributeUuidList,
					ProductPackageGroupUuid:         subProductParam.ProductPackageGroupUuid,
					Operation:                       "add",
				}
				subProductParams = append(subProductParams, params)
			}
			productReq.SetIsPackageProduct(subProductParams) // 设置为套餐商品
		}
		request.Products[index] = productReq
	}

	checkServiceRes, err := s.ActionAddAndCooking(ctx, req.ProductAddReq{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products:      request.Products,
		IsH5Product:   false,
	}, saleBill, request.IgnoreMust)
	if checkServiceRes != nil {
		if checkServiceRes.Code == constant.CodeOrderCheckProductMust && request.IgnoreMust {
			// 必点方案未选择，且忽略必点方案
		} else {
			return &TabletAddAndCookingRes{OrderCheckRes: checkServiceRes.OrderCheckRes}, errors.WithMessage(errors.New(constant.ParseCodeOrderCheck(checkServiceRes.Code, constant.WithIsTablet())))
		}
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type ActionAddOption struct {
	IsTableAdd  bool // 是否是平板端加购
	IsMemberAdd bool // 是否是会员端加购
	skipLimit   bool // 是否跳过加购时的限购检查。使用场景是加购并送厨时，因为送厨会在坚持一次限购，所有加购时不检查
}

func WithIsMemberAdd() func(option *ActionAddOption) {
	return func(option *ActionAddOption) {
		option.IsMemberAdd = true
	}
}

func WithIsTableAdd() func(option *ActionAddOption) {
	return func(option *ActionAddOption) {
		option.IsTableAdd = true
	}
}

func WithSkipLimit() func(option *ActionAddOption) {
	return func(option *ActionAddOption) {
		option.skipLimit = true
	}
}

// 加购。内部方法复用
func (s *orderSrv) actionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill, options ...func(option *ActionAddOption)) (*model.SaleBill, error) {
	option := &ActionAddOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	// 检查销售订单商品数量是否超过1000项
	if err := saleBill.CheckSaleOrderProductNum(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}
	// 录入订单商品数据
	saleOrderProducts, err := s.newSaleOrderProduct(ctx, CreateSaleOrderProductParams{
		IsH5Product: request.IsH5Product,
		Setting:     *saleBill.SaleBillSetting,
		SaleBill:    saleBill,
		SaleOrder:   saleOrder,
		Products:    request.Products,
	}, options...)
	if err != nil {
		return nil, errors.WithMessage(err, "构建商品失败")
	}

	if !option.skipLimit {
		// 检查限购
		if request.IsH5Product == true {
			// 如果是h5端加购的话
			if err := s.checkLimitPurchase(ctx, saleBill, saleOrderProducts, model.WithH5CheckLimit()); err != nil {
				return nil, errors.WithMessage(err)
			}
		} else {
			if err := s.checkLimitPurchase(ctx, saleBill, saleOrderProducts); err != nil {
				return nil, errors.WithMessage(err)
			}
		}
	}
	// 检查超时不能加购
	if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, saleOrderProducts); err != nil {
		return nil, errors.WithMessage(err)
	}
	// saleBill已经加入了新的商品，并且重新计算了价格
	return saleBill, nil
}

// 检查限制购 checkLimitPurchase
func (s *orderSrv) checkLimitPurchase(ctx context.Context, saleBill *model.SaleBill, saleOrderProducts []*model.SaleOrderProduct, options ...func(option *model.CalcOption)) error {
	option := &model.CalcOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	limitProducts := make(map[uint64]uint) // product_package_uuid => limit_num
	var err error

	if saleBill.IsBuffetSaleBill() { // 只有自助餐账单才查询自助餐商品限购信息
		// 获取自助餐商品的限购数量 limitProducts map[uint64]uint product_package_uuid => limit_num
		limitProducts, err = s.getBuffetProductLimitList(ctx, saleBill.Uuid)
		if err != nil {
			return errors.WithMessage(err)
		}
	}

	// 仅检查本次加购的商品是否超过限购
	uuids := make([]uint64, 0)
	for _, saleOrderProduct := range saleOrderProducts {
		uuids = append(uuids, saleOrderProduct.Uuid)
	}

	var overLimitProducts []*model.SaleOrderProduct
	if option.H5CheckLimit {
		// 如果是h5端加购检查限购
		overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithH5CheckLimit(), model.WithSaleOrderProductUuid(uuids...))
	} else {
		overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithSaleOrderProductUuid(uuids...))
	}
	if len(overLimitProducts) > 0 {
		return errors.New("商品超过限购")
	}
	return nil
}

// 检查超时不能加购 checkTimeoutAndCannotAddPurchase
func (s *orderSrv) checkTimeoutAndCannotAddPurchase(ctx context.Context, saleBill *model.SaleBill, saleOrderProducts []*model.SaleOrderProduct) error {
	// 检查超时不能加购
	if saleBill.IsBuffetSaleBill() {
		// 获取自助餐的剩余时长
		if saleBill.GetTotalRemainingSeconds() == 0 {
			// 获取自助餐设置
			companySetting, err := s.settingSrv.GetCompanySetting(ctx)
			if err != nil {
				return err
			}
			buffetSetting, buffetErr := s.settingSrv.GetBuffetSetting(ctx, companySetting)
			if buffetErr != nil {
				return buffetErr
			}
			// 如果自助餐设置为非自助餐商品到时不能继续选购，则不能加购
			if buffetSetting.IsBuyContinue == "0" {
				return errors.New("用餐时间已到，无法继续下单")
			}
			// 自助餐已结束，不能加购自助餐商品。但可以根据设置，继续选购非自助餐商品
			for _, saleOrderProduct := range saleOrderProducts {
				if saleOrderProduct.IsBuffetProduct() {
					return errors.New("自助餐时间已到达，自助餐商品不可继续下单")
				}
			}
		}
	}
	return nil
}

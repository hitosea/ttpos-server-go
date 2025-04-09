package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ActionCookingOption struct {
	CalcAndSaveSaleBill bool
}

func withCalcAndSaveSaleBill() func(option *ActionCookingOption) {
	return func(option *ActionCookingOption) {
		option.CalcAndSaveSaleBill = true
	}
}

// ActionCooking 送厨
func (s *orderSrv) ActionCooking(ctx context.Context, ignoreMust bool, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct, h5OrderUuid uint64, options ...func(option *ActionCookingOption)) (*resp.OrderCheckServiceRes, error) {
	option := &ActionCookingOption{}
	for _, opt := range options {
		opt(option)
	}
	var productionOrder *model.ProductionOrder
	var warehouseOutForm *model.WarehouseOutForm

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
		saleOrderProductAll := saleBill.GetSaleOrderProductAll()

		// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
		checkServiceRes, errCheck := s.checkOrder(ctx, false, db, saleBill.Uuid, saleBill.DeskUuid, saleOrderProductAll, WithCheckTypeCooking())
		if errCheck != nil {
			ctx.Log().Error("检查商品失败", zap.Error(errCheck))
			return nil, errors.New("检查商品失败")
		}
		if checkServiceRes != nil {
			if checkServiceRes.Code == constant.CodeOrderCheckProductMust && ignoreMust {
				// 必点方案未选择，且忽略必点方案
			} else {
				return checkServiceRes, nil
			}
		}

		// 构建送厨单
		productionOrder = newProductionOrder(ctx, saleOrderUuid, saleBill.Uuid, unCookingSaleOrderProducts)

		// 修改商品状态为已送厨
		for index, _ := range unCookingSaleOrderProducts {
			product := unCookingSaleOrderProducts[index]
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
		warehouseOutForm = model.NewWarehouseOutForm(decreaseStockList, false, saleBill.Uuid, ctx.GetStaffUuid())
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
			}
		}
		// 出库相关
		{
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
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新数据失败")
	}

	// 操作记录相关
	{
		// 发起“送厨”操作的事件
		products := make(event.Products, 0)
		for _, unCookingSaleOrderProduct := range unCookingSaleOrderProducts {
			products = append(products, event.OrderProduct{
				OrderProductId:  unCookingSaleOrderProduct.Uuid,
				ProductId:       unCookingSaleOrderProduct.ProductPackageUuid,
				ProductName:     unCookingSaleOrderProduct.MultiLanguageName.GetNames(),
				ProductAttr:     unCookingSaleOrderProduct.GetAttributeName(),
				ProductAttrList: unCookingSaleOrderProduct.GetAttributeNameList(),
				TotalNum:        unCookingSaleOrderProduct.Num,
				IsBuffet:        unCookingSaleOrderProduct.IsBuffet == 1,
				Remark:          unCookingSaleOrderProduct.Remark,
			})
		}
		go func() {
			s.bus.PublishSentCookingEvent(event.SentCookingPayload{
				BasePayload: event.BasePayload{
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  saleBill.Uuid,
					SaleOrderUuid: saleOrderUuid,
					H5OrderUuid:   h5OrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				Products: products,
			})
		}()
	}
	return nil, nil
}

// 加购
func (s *orderSrv) ActionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill) error {
	db := ctx.GetDB()

	saleBill, err := s.actionAdd(ctx, request, saleBill)
	if err != nil {
		return errors.WithMessage(err)
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
func (s *orderSrv) ActionAddAndCooking(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill) (*resp.OrderCheckServiceRes, error) {

	// 加购相关
	_, err := s.actionAdd(ctx, request, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 送厨相关
	{
		// 获取未送厨的商品列表
		unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()
		if len(unCookingSaleOrderProducts) == 0 {
			return nil, errors.New("没有未送厨的商品")
		}

		// 送厨
		checkServiceRes, err := s.ActionCooking(ctx, true, saleBill, unCookingSaleOrderProducts, 0, withCalcAndSaveSaleBill()) // 平板端加购并送厨
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if checkServiceRes != nil {
			return checkServiceRes, nil
		}
	}

	return nil, nil
}

// TabletAddAndCooking 平板端加购并送厨
func (s *orderSrv) TabletAddAndCooking(ctx context.Context, request req.TabletOrderCartProductAddReq) error {
	saleBill, _ := repository.NewOrderRepo(ctx.GetDB()).GetSaleBillAllInfo(request.SaleBillUuid)
	if saleBill.IsEndStatus() {
		return errors.WithMessage(errors.NewWithCode(constant.CodeDeskOrderEnd, "桌台订单结束"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid); err != nil {
		return errors.WithMessage(err)
	}
	_, err := s.ActionAddAndCooking(ctx, req.ProductAddReq{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products:      request.Products,
		IsH5Product:   false,
	}, saleBill)
	return err
}

// 加购。内部方法复用
func (s *orderSrv) actionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill) (*model.SaleBill, error) {

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
	})
	if err != nil {
		return nil, errors.WithMessage(err, "构建商品失败")
	}

	// 检查限购
	{
		overLimitProducts := saleBill.GetSaleOrderProductOverLimit()
		if len(overLimitProducts) > 0 {
			return nil, errors.New("商品超过限购")
		}
	}
	// 检查超时不能加购
	{
		if saleBill.IsBuffetSaleBill() {
			// 获取自助餐的剩余时长
			if saleBill.GetTotalRemainingSeconds() == 0 {
				// 自助餐已结束，不能加购自助餐商品。但可以根据设置，继续选购非自助餐商品
				for _, saleOrderProduct := range saleOrderProducts {
					if saleOrderProduct.IsBuffetProduct() {
						return nil, errors.New("自助餐已结束")
					}
				}
				// 获取自助餐设置
				companySetting, err := s.settingSrv.GetCompanySetting(ctx)
				if err != nil {
					return nil, err
				}
				buffetSetting, buffetErr := s.settingSrv.GetBuffetSetting(ctx, companySetting)
				if buffetErr != nil {
					return nil, buffetErr
				}
				// 如果自助餐设置为非自助餐商品到时不能继续选购，则不能加购
				if buffetSetting.IsBuyContinue == "0" {
					return nil, errors.New("自助餐已结束")
				}
			}
		}
	}
	// 检查库存 todo 检查商品库存，避免前端商品列表未更新导致库存不足的商品被成功加购
	// for _, productParam := range request.Products {
	// }

	// saleBill已经加入了新的商品，并且重新计算了价格
	return saleBill, nil
}

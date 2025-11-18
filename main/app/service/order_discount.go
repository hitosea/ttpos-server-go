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
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderProductChangePrice  修改订单商品价格
func (s *orderSrv) OrderProductChangePrice(ctx context.Context, req req.OrderProductChangePriceReq) (*resp.ShopCart, error) {
	if req.Price < 0 || req.Price > 100000000 {
		return nil, errors.New("请输入0-100000000间的价格")
	}

	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取操作的销售账单信息
	saleBill, saleOrder, saleOrderProduct, err := getSaleOrderFromDB(ctx, db, req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询销售订单信息失败")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderChangePrice, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	ctx.Log().Debug("改价前", zap.Any("SalePrice", saleOrderProduct.SalePrice))
	ctx.Log().Debug("改价前", zap.Any("saleOrderProduct calc", saleOrderProduct.BeforeCalc()))
	// 改价
	saleOrderProduct.ChangeProductPrice(req.Price)

	ctx.Log().Debug("改价后", zap.Any("SalePrice", saleOrderProduct.SalePrice))

	// 计算商品数据。折扣、税费、服务
	afterCalc := saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
	ctx.Log().Debug("改价后", zap.Any("saleOrderProduct calc", afterCalc))

	// 计算订单金额
	ctx.Log().Debug("改价前,销售订单信息", zap.Any("saleOrder calc", saleOrder.BeforeCalc()))
	afterSaleOrderCalc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("改价后,销售订单信息", zap.Any("saleOrder calc", afterSaleOrderCalc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdateProduct := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*saleOrderProduct); errUpdateProduct != nil {
			return errUpdateProduct
		}
		// 更新套餐商品的子商品的签名
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.UpdateSign()
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProductRecord(*subProduct); errUpdate != nil {
					return errUpdate
				}
			}
		}
		// 更新完整个销售订单
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errUpdate
		}
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errUpdateSaleBill
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "更新销售订单失败")
	}

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"改价"事件
	utils.Go(func() {
		event.NewSystemBus().PublishChangeSaleOrderProductPriceEvent(event.ChangeSaleOrderProductPricePayload{
			BasePayload: event.BasePayload{ // 改价
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.OrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			TotalNum:       saleOrderProduct.Num,
			Price:          req.Price,
		})
	})

	return info, nil
}

// OrderAmountChange  修改订单金额，整单改价
func (s *orderSrv) OrderAmountChange(ctx context.Context, req req.OrderAmountChangeReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 设置整单改价金额
	saleOrder.SetCustomAmount(req.Price)
	oldPrice := saleOrder.GetOriginAmount()

	// 整单改价后，整单折扣会取消，需要重新计算订单商品的金额
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"改价"事件
	utils.Go(func() {
		event.NewSystemBus().PublishDiscountChangePriceSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 改价
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        oldPrice, // 旧价格为订单的原始应收金额
			NewPrice:        req.Price,
			DiscountType:    constant.DiscountOperationLogTypeChangePriceSaleOrder, // 整单改价的类型值
			SpecialDiscount: decimal.NewFromFloat(oldPrice).Sub(decimal.NewFromFloat(req.Price)).InexactFloat64(),
		})
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderDiscount  修改订单折扣
func (s *orderSrv) OrderDiscount(ctx context.Context, req req.OrderDiscountReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 取消会员优惠券
	saleOrder.SetAllCouponCancel()
	if err := repository.NewSaleOrderCouponRepo(db).UpdateSaleOrderCouponCancelAll(saleOrder.Uuid); err != nil {
		return nil, errors.WithMessage(err, "取消销售订单优惠券失败")
	}

	// 在折扣之前计算会员打折后金额。必须在设置折扣之前获取，否则amount值已经改变了
	memberDiscountAmount := saleOrder.GetMemberDiscountAmount()
	// 设置整单折扣率
	saleOrder.SetCustomDiscount(req.GetDiscount())

	// 获取最新的设置
	newSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, saleBill.DeskUuid, false)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting)); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"整单打折"事件
	utils.Go(func() {
		discountAmount := saleOrder.CustomDiscountFee
		event.NewSystemBus().PublishDiscountSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 整单打折
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OldPrice:        memberDiscountAmount, // 旧价格为订单的会员折扣后的金额。如果没有会员折扣，则旧价格为订单应收金额
			NewPrice:        saleOrder.GetAmount(),
			DiscountType:    constant.DiscountOperationLogTypeDiscountSaleOrder,
			RoundingRate:    req.GetPercentDiscount(),
			SpecialDiscount: discountAmount,
		})
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderZeroRule  修改订单抹零规则
func (s *orderSrv) OrderZeroRule(ctx context.Context, req req.OrderZeroRuleReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 设置整单改价金额
	saleOrder.SetZeroRule(req.ZeroRule)

	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布"订单抹零"事件
	utils.Go(func() {
		event.NewSystemBus().PublishDiscountZeroSaleOrderEvent(event.DiscountSaleOrderPayload{
			BasePayload: event.BasePayload{ // 订单抹零
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			DiscountType:    constant.DiscountOperationLogTypeZeroSaleOrder,
			RoundingType:    req.ZeroRule,
			SpecialDiscount: saleOrder.ZeroFee, // ZeroFee这个字段是算好的抹零优惠金额。先计算好订单应付金额，再根据抹零规格进行抹零得到的结果
		})
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderDiscountCancel  取消点餐订单所有优惠折扣，包括改价、打折、抹零。撤销优惠折扣
func (s *orderSrv) OrderDiscountCancel(ctx context.Context, req req.OrderDiscountCancelReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 当前销售账单数据
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errSaleBill
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderDiscount, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 撤销订单的优惠折扣
	saleOrder.SetAllDiscountCancel()

	// 获取最新的设置
	newSetting, err := s.NewSaleBillSetting(ctx, saleBill.Uuid, saleBill.DeskUuid, false)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 计算并保存销售账单
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting)); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 发布取消优惠折扣事件
	utils.Go(func() {
		event.NewSystemBus().PublishCancelSaleOrderDiscountEvent(event.CancelSaleOrderDiscountPayload{
			BasePayload: event.BasePayload{ // 取消折扣
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			ParentId:  req.SaleBillUuid,
			OrderName: req.SaleOrderUuid,
		})
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

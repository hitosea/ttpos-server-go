package service

import (
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderProductDelete 删除订单商品
func (s *orderSrv) OrderProductDelete(ctx context.Context, dbId uint64, staffUuid uint64, source string, req req.OrderProductDeleteReq) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	return s.orderProductDelete(ctx, dbId, staffUuid, source, req)
}

// OrderProductRemark  修改订单商品备注
func (s *orderSrv) OrderProductRemark(ctx context.Context, req req.OrderProductRemarkReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取信息源
	orderRepo := repository.NewOrderRepo(s.dbm.GetDB(dbId))

	// 获取订单信息
	billInfo, err := orderRepo.GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if ctx.GetSource() == constant.SourceAssistant && billInfo.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := billInfo.ValidateOrderStatus(ctx.GetSource(), constant.OrderProductRemark, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单
	saleOrder := billInfo.GetSaleOrder(req.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 修改订单商品备注
	var saleOrderProduct *model.SaleOrderProduct
	for _, product := range saleOrder.SaleOrderProducts {
		if product.Uuid == req.OrderProductUuid {
			saleOrderProduct = product
			break
		}
	}
	if saleOrderProduct == nil {
		return nil, errors.New("订单商品不存在")
	}

	// 更新订单商品备注
	repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		saleOrderProduct.Remark = req.Remark
		saleOrderProduct.UpdateSign()
		sign := saleOrderProduct.Sign
		if err := repository.NewOrderRepo(db).ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, req.OrderProductUuid, req.Remark, sign); err != nil {
			return errors.WithMessage(err)
		}
		// 更新套餐商品的子商品的签名
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				subProduct.UpdateSign()
				if err := repository.NewOrderRepo(db).ChangeProductRemark(req.SaleBillUuid, req.SaleOrderUuid, subProduct.Uuid, req.Remark, subProduct.Sign); err != nil {
					return errors.WithMessage(err)
				}
			}
		}
		return nil
	})

	// 获取新的数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderCartProductAdd 向购物车添加商品
func (s *orderSrv) OrderCartProductAdd(ctx context.Context, request req.ProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 判断订单状态
	if ctx.GetSource() == constant.SourceAssistant {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid, model.WithIsAssistant()); err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 设置添加来源
	saleBill.SetOperateSource(ctx.GetSource())

	// 加购
	if err := s.ActionAdd(ctx, request, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// OrderCartProductNum 修改购物车商品数量
func (s *orderSrv) OrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}
	option := &repository.OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.Log().Info("修改购物车商品数量", zap.Any("request", request))
	// 商品数量不能超过999个
	if request.Num > 999 {
		ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
		return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
	}

	// 检查商品销售库存是否充足
	ctx.Log().Debug("获取账单信息")
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	ctx.Log().Debug("获取到账单信息成功")

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateProductNum, request.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	ctx.Log().Debug("获取到订单信息成功")

	// 获取销售订单商品信息
	saleOrderProduct, index, errSaleOrderProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
	if errSaleOrderProduct != nil {
		return nil, errors.WithMessage(errSaleOrderProduct)
	}
	ctx.Log().Debug("获取到订单商品信息成功")

	if saleOrderProduct.IsCookingProduct() {
		return nil, errors.WithMessage(errors.New("商品已送厨，不能修改数量"))
	}
	// 数量为0删除商品
	if request.Num == 0 {
		res, err := s.OrderProductDelete(ctx, ctx.GetDbId(), ctx.GetStaffUuid(), ctx.GetSource(), req.OrderProductDeleteReq{
			SaleBillUuid:     request.SaleBillUuid,
			SaleOrderUuid:    request.SaleOrderUuid,
			OrderProductUuid: request.SaleOrderProductUuid,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "删除商品失败")
		}
		return res, nil
	}

	// 修改销售订单商品数量
	beforeNum := saleOrderProduct.Num
	saleOrderProduct.Num = request.Num
	ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

	// 检查商品销售库存是否充足
	if request.Num > beforeNum {
		status, message := saleOrderProduct.CheckCookingProduct(ctx.GetLanguage())
		if status != constant.CodeSuccess {
			return nil, errors.WithMessage(errors.New(message))
		}
	}
	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了商品金额", zap.Any("saleOrderProduct salePrice", saleOrderProduct.SalePrice))
	saleOrder.SaleOrderProducts[index] = saleOrderProduct

	// 计算订单金额
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	// 检查限购
	{
		// 如果是减数量，则不检查限购. 只有加数量时，才检查限购
		if request.Num > beforeNum {
			limitProducts, err := s.getBuffetProductLimitList(ctx, request.SaleBillUuid)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			var overLimitProducts []*model.SaleOrderProduct
			if option.UnorderedH5Product == repository.UnorderedH5Product {
				overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithH5CheckLimit(), model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
			} else {
				overLimitProducts = saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
			}
			if len(overLimitProducts) > 0 {
				return nil, errors.WithMessage(errors.New("商品超过限购"))
			}
		}
	}

	// 商品数量不能超过999个
	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购999个
	if request.Num > beforeNum {
		if request.Num > constant.ProductNumMax {
			ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
			return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
		}
	}

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
			subProduct.Num = decimal.NewFromFloat(saleOrderProduct.Num).Mul(unitNum).Mul(decimal.NewFromFloat(subProduct.CopyNum)).Round(3).InexactFloat64()
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单商品成功")
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(subProduct); errUpdate != nil {
					return errors.WithMessage(errUpdate)
				}
				ctx.Log().Debug("更新销售订单套餐子商品成功")
			}
		}
		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单成功")
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errors.WithMessage(errUpdateSaleBill)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "修改商品数量时，保存数据失败")
	}
	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	ctx.Log().Debug("获取新的账单数据")
	return info, nil
}

// AssistantOrderCartProductNum 助手端修改购物车商品数量
func (s *orderSrv) AssistantOrderCartProductNum(ctx context.Context, request req.OrderCartProductNumReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.Log().Info("修改购物车商品数量", zap.Any("request", request))
	// 商品数量不能超过999个
	if request.Num > 999 {
		ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
		return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	ctx.Log().Debug("获取到账单信息成功")

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderUpdateProductNum, 0, model.WithIsAssistant()); err != nil {
		return nil, errors.WithMessage(err)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	ctx.Log().Debug("获取到订单信息成功")

	// 获取销售订单商品信息
	saleOrderProduct := saleBill.GetSaleOrderProductByUuid(request.SaleOrderProductUuid)
	if saleOrderProduct == nil {
		return nil, errors.New("销售订单商品不存在")
	}
	ctx.Log().Debug("获取到订单商品信息成功")

	// 商品的签名
	sign := saleOrderProduct.Sign
	saleOrderProductList := saleBill.GetSaleOrderProductBySign(sign)
	productNum := float64(0)
	for _, saleOrderProduct := range saleOrderProductList {
		productNum += saleOrderProduct.Num
	}
	operation := "add"
	if productNum > request.Num {
		operation = "sub"
	}

	// 排序. 从子单开始减商品
	slices.SortFunc(saleBill.SaleOrders, func(a, b *model.SaleOrder) int {
		return int(b.CreateTime) - int(a.CreateTime)
	})
	for _, saleOrder := range saleBill.SaleOrders {
		for _, product := range saleOrderProductList {
			if product.SaleOrderUuid == saleOrder.Uuid {
				saleOrderProduct = product
				break
			}
		}
	}

	if saleOrderProduct.IsCookingProduct() {
		return nil, errors.WithMessage(errors.New("商品已送厨，不能修改数量"))
	}

	// 修改销售订单商品数量
	beforeNum := saleOrderProduct.Num

	// 检查商品销售库存是否充足
	if operation == "add" {
		saleOrderProduct.Num = saleOrderProduct.Num + 1
		request.Num = saleOrderProduct.Num
		ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))

		// FIXME 暂时废弃，只在送厨和结账时检查
		// status, message := saleOrderProduct.CheckCookingProduct(ctx.GetLanguage())
		// if status != constant.CodeSuccess {
		// 	return nil, errors.WithMessage(errors.New(message))
		// }
	} else if operation == "sub" {
		num := saleOrderProduct.Num - 1
		// 数量为0删除商品
		if num == 0 {
			res, err := s.OrderProductDelete(ctx, ctx.GetDbId(), ctx.GetStaffUuid(), ctx.GetSource(), req.OrderProductDeleteReq{
				SaleBillUuid:     request.SaleBillUuid,
				SaleOrderUuid:    saleOrderProduct.SaleOrderUuid,
				OrderProductUuid: saleOrderProduct.Uuid,
			})
			if err != nil {
				return nil, errors.WithMessage(err, "删除商品失败")
			}
			return res, nil
		}
		saleOrderProduct.Num = num
		request.Num = saleOrderProduct.Num
		ctx.Log().Debug("修改商品数量", zap.Any("num", saleOrderProduct.Num))
	}
	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
			subProduct.Num = decimal.NewFromFloat(saleOrderProduct.Num).Mul(unitNum).Round(3).InexactFloat64()
			subProduct.CalcSaleOrderProduct(*saleBill.SaleBillSetting)
		}
	}

	// 计算订单金额
	calc := saleOrder.CalcSaleOrder(*saleBill.SaleBillSetting)
	ctx.Log().Debug("重新计算了订单金额", zap.Any("calc", calc))
	// 计算账单金额
	saleBill.CalcSaleBill()

	// FIXME 暂时废弃，只在送厨和结账时检查
	// 检查限购
	// {
	// 	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购
	// 	if request.Num > beforeNum {
	// 		limitProducts, err := s.getBuffetProductLimitList(ctx, request.SaleBillUuid)
	// 		if err != nil {
	// 			return nil, errors.WithMessage(err)
	// 		}
	// 		overLimitProducts := saleBill.GetSaleOrderProductOverLimit(limitProducts, model.WithSaleOrderProductUuid(request.SaleOrderProductUuid))
	// 		if len(overLimitProducts) > 0 {
	// 			return nil, errors.WithMessage(errors.New("商品超过限购"))
	// 		}
	// 	}
	// }

	// 商品数量不能超过999个
	// 如果是减数量，则不检查限购. 只有加数量时，才检查限购999个
	if request.Num > beforeNum {
		if request.Num > constant.ProductNumMax {
			ctx.Log().Error("商品数量不能超过999个", zap.Any("request", request))
			return nil, errors.WithMessage(errors.New("商品数量不能超过999个"))
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(saleOrderProduct); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单商品成功")
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				if errUpdate := repository.NewSaleOrderProductRepo(db).UpdateSaleOrderProduct(subProduct); errUpdate != nil {
					return errors.WithMessage(errUpdate)
				}
			}
			ctx.Log().Debug("更新销售订单套餐子商品成功")
		}

		if errUpdate := repository.NewSaleOrderRepo(db).UpdateSaleOrderRecord(*saleOrder); errUpdate != nil {
			return errors.WithMessage(errUpdate)
		}
		ctx.Log().Debug("更新销售订单成功")
		if errUpdateSaleBill := repository.NewSaleBillRepo(db).UpdateSaleBillRecord(*saleBill); errUpdateSaleBill != nil {
			return errors.WithMessage(errUpdateSaleBill)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "修改商品数量时，保存数据失败")
	}

	// 获取新的桌台数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	ctx.Log().Debug("获取新的账单数据")

	return info, nil
}

// InstantOrderCartProductCooking 送厨购物车商品
func (s *orderSrv) InstantOrderCartProductCooking(ctx context.Context, req req.OrderCartProductCookingReq) (*resp.ShopCart, *resp.OrderCheckServiceRes, error) {
	defer func() { // 送厨结束后，执行分批送厨
		// 助手端前置模式：分批送厨（每次点击下单都送优先级最高的分批类型）
		if ctx.GetSource() == constant.SourceAssistant {
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				ctx.Log().Info("获取业务设置失败,导致不能分批送厨", zap.Error(err))
			} else if businessSetting.BatchCookingMode == constant.BatchCookingModePre {
				// 异步执行分批送厨，不阻塞流程
				utils.Go(func() {
					ctx := ctx.Copy()
					if err := s.AutoSendCookingByPriority(ctx, req.SaleBillUuid); err != nil {
						ctx.Log().Error("分批送厨失败", zap.Error(err))
					}
				})
			}
		}
	}()

	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())

	// 从http的header中获取h5_order_uuid
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		req.H5OrderUuid = h5OrderUuid
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, nil, errSaleBill
	}
	ctx.Log().Debug("获取销售账单信息")

	h5OrderProductUnAccept := make([]*model.SaleOrderProduct, 0)
	if req.H5OrderUuid != 0 {
		h5OrderProductUnAccept = saleBill.GetH5OrderProductUnAccept(req.H5OrderUuid)
	}

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := saleBill.GetSaleOrderProductUnCooking()
	nonNeedBatchSendCooking := false
	// 没有未送厨商品时，判断是否需要弹出分批送厨弹窗。只有收银机和助手端需要判断
	if ctx.GetSource() == constant.SourceAssistant || ctx.GetSource() == constant.SourceCashier {
		// 获取门店业务设置
		businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
		if err != nil {
			return nil, nil, errors.WithMessage(err)
		}
		if businessSetting.OpenIsBatch() {
			if saleBill.IsNeedBatchSendCooking() {
				nonNeedBatchSendCooking = true
			}
		}
	}

	if len(unCookingSaleOrderProducts) == 0 && len(h5OrderProductUnAccept) == 0 && !nonNeedBatchSendCooking {
		return nil, nil, errors.New("没有未送厨的商品")
	} else if len(unCookingSaleOrderProducts) == 0 && len(h5OrderProductUnAccept) == 0 && nonNeedBatchSendCooking {
		if !req.IgnoreMust { // 如果不是在结账检查时送厨，才返回-209，否则直接不分批送厨而是使用正常送厨
			return nil, &resp.OrderCheckServiceRes{
				Code: constant.CodeOrderCheckProductBatch,
			}, nil
		} else {
			// 获取预送厨的商品
			preCookingSaleOrderProducts := saleBill.GetSaleOrderProductPreCooking()
			if len(preCookingSaleOrderProducts) > 0 {
				nonBatchUuids := make([]uint64, 0) // 预送厨的商品uuid列表
				for _, saleOrderProduct := range preCookingSaleOrderProducts {
					saleOrderProduct.IsBatch = 0 // 临时在内存中将该商品的is_batch设置为0,让打印出该商品的送厨单
					nonBatchUuids = append(nonBatchUuids, saleOrderProduct.Uuid)
				}
				if len(nonBatchUuids) > 0 {
					// 将预送厨的商品变为未分批商品
					db := s.dbm.GetDB(ctx.GetDbId())
					// 从业务设置中获取分批送厨模式
					businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
					if err != nil {
						return nil, nil, errors.WithMessage(err)
					}
					if err := s.updateProductBatchFlagToZero(db, nonBatchUuids, businessSetting.BatchCookingMode); err != nil {
						return nil, nil, errors.WithMessage(err)
					}
					// 场景一: 全是预送厨商品时
					// 发起“送厨”操作的事件
					utils.Go(func() {
						saleOrderUuid := saleBill.GetFirstSaleOrder().Uuid
						s.bus.PublishSentCookingEvent(event.SentCookingPayload{
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
								for _, unCookingSaleOrderProduct := range preCookingSaleOrderProducts {
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
					})

				}
			}
		}
	}

	// 获取某个h5订单的已下单但未接单的商品
	if req.H5OrderUuid != 0 {
		checkRes, err := s.AcceptH5Order(ctx, req.H5OrderUuid, false)
		if err != nil {
			return nil, nil, err
		}
		if checkRes != nil {
			return nil, checkRes, nil
		}
	}

	// 送厨
	if len(unCookingSaleOrderProducts) > 0 {
		var checkServiceRes *resp.OrderCheckServiceRes
		var err error
		// 助手端，仅检查送厨时
		if ctx.GetSource() == constant.SourceAssistant && req.IsCheckCooking {
			checkServiceRes, err = s.ActionCooking(ctx, req.IgnoreMust, saleBill, unCookingSaleOrderProducts, 0, false, WithOnlyCheckCooking()) // 购物车送厨检查
		} else {
			checkServiceRes, err = s.ActionCooking(ctx, req.IgnoreMust, saleBill, unCookingSaleOrderProducts, 0, false, WithIsBatch()) // 购物车送厨商品
		}
		if err != nil {
			return nil, nil, err
		}
		if checkServiceRes != nil {
			return nil, checkServiceRes, nil
		}
	}

	// 会员端接单，不返回购物车信息
	if req.IsMemberOrderAccept {
		return nil, nil, nil
	}

	ctx.Log().Debug("获取新的购物车信息")
	cartInfo, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取购物车信息失败")
	}
	return cartInfo, nil, nil
}

// InstantOrderCartProductReturning 退菜购物车商品
func (s *orderSrv) InstantOrderCartProductReturning(ctx context.Context, req req.OrderCartProductReturningReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 验证高级密码
	if err := s.settingSrv.VerifyAdvancedPassword(ctx, req.Password); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 校验是否合规
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsSendKitchen():
		return nil, errors.New("商品未送厨")
	case req.Num == 0:
		return nil, errors.New("退菜数量不能为0")
	case req.Num > saleOrderProduct.Num:
		return nil, errors.New("退菜数量不能大于当前商品数量")
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	returnFoodReason, err := base.NewReturnFoodReasonRepo(db).GetReturnFoodReasonListByUuids(req.ReturnIds)
	if err != nil {
		return nil, errors.WithMessage(err, "params:", utils.ToJson(req.ReturnIds))
	}
	// 如果查到的原因数量跟提交的原因数量不一致，提示退菜原因不存在
	if len(returnFoodReason) != len(req.ReturnIds) {
		return nil, errors.WithMessage(fmt.Errorf("退菜原因不存在: %v", req.ReturnIds))
	}
	// 构建订单商品退菜原因列表
	returnFoodReasonList := saleOrderProduct.NewSaleOrderProductReasonList(returnFoodReason)

	var returnSaleOrderProduct *model.SaleOrderProduct
	// var returnSubSaleOrderProducts []*model.SaleOrderProduct // 套餐子商品退菜列表
	var keepNum float64

	// 如果退菜数量等于该商品的数量，则标记该商品为退菜并在商品的退菜原因列表中添加退菜原因
	if saleOrderProduct.Num == req.Num {
		// 尝试获取相同签名的商品，如果有，将两个商品合并。该商品可能已经退过菜
		newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderProduct.SaleOrderUuid)
		newSaleOrderProduct.SetNum(req.Num)
		newSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
		if sameSignSaleOrderProduct != nil {
			// 有相同签名的商品。将两个商品合并，数量相加
			sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
			returnSaleOrderProduct = sameSignSaleOrderProduct
			saleOrderProduct.SetDelete() // 标记该商品为删除
			saleOrderProduct.SetUpdate() // 标记该商品需要更新
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					subProduct.SetDelete() // 标记该商品为删除
					subProduct.SetUpdate() // 标记该商品需要更新
				}
			}
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if sameSignSaleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(sameSignSaleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
					addNum := decimal.NewFromFloat(req.Num).Mul(unitNum).Round(3).InexactFloat64()
					subProduct.SetNum(subProduct.Num + addNum)
				}
			}

		} else {
			saleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
			returnSaleOrderProduct = saleOrderProduct
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					subProduct.SetCancelInfo(req.Reason, nil)
				}
			}
		}
	} else {
		// 如果退菜数量小于该商品的数量，则新建一个销售订单商品并在新商品的退菜原因列表中添加退菜原因
		// 1. 修改原商品的数量
		// 2. 新建一个销售订单商品，该商品数量为退菜数量.
		// 3. 判断新建的销售订单商品是否要合并到已有的退菜商品中。当两个退菜商品的签名一致时，将两个商品合并，数量相加
		// 修改原商品的数量
		keepNum = saleOrderProduct.Num - req.Num
		saleOrderProduct.SetNum(keepNum)
		// 新建一个销售订单商品，该商品数量为退菜数量
		newSaleOrderProduct := saleOrderProduct.CopyOrderProduct(saleOrderProduct.SaleOrderUuid)
		newSaleOrderProduct.SetNum(req.Num)
		newSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
		sameSignSaleOrderProduct := saleOrder.GetSaleOrderProductBySign(newSaleOrderProduct.Sign)
		if sameSignSaleOrderProduct != nil {
			// 有相同签名的商品。将两个商品合并，数量相加
			sameSignSaleOrderProduct.SetNum(sameSignSaleOrderProduct.Num + req.Num)
			returnSaleOrderProduct = sameSignSaleOrderProduct
			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(sameSignSaleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
					num := decimal.NewFromFloat(sameSignSaleOrderProduct.Num).Mul(unitNum).Round(3).InexactFloat64()
					subProduct.SetNum(num)
				}
			}
		} else {
			// 没有相同签名的商品。将新建的商品添加到销售订单商品列表中
			// CalcAndSaveSaleBill 方法会检查到newSaleOrderProduct没有主键ID，会创建新记录。所以不用另外创建该订单商品，否则会重复创建
			saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, newSaleOrderProduct)
			returnSaleOrderProduct = newSaleOrderProduct
			// 补全newSaleOrderProduct对象的ProductBom对象
			for _, saleOrderProductBom := range newSaleOrderProduct.SaleOrderProductBoms {
				productBom, err := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(saleOrderProductBom.ProductBomUuid)
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				saleOrderProductBom.ProductBom = *productBom
			}

			// 如果是套餐商品，则更新套餐子商品退菜信息
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					newSubSaleOrderProduct := subProduct.CopyOrderProduct(subProduct.SaleOrderUuid)
					newSubSaleOrderProduct.PackageUuid = newSaleOrderProduct.Uuid // 设置为新的套餐商品uuid
					unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
					num := decimal.NewFromFloat(req.Num).Mul(unitNum).Round(3).InexactFloat64()
					newSubSaleOrderProduct.SetNum(num)
					newSubSaleOrderProduct.SetCancelInfo(req.Reason, returnFoodReasonList)
					saleOrder.SaleOrderProducts = append(saleOrder.SaleOrderProducts, newSubSaleOrderProduct)
				}
			}
		}
		// 更新旧的子商品的数量
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
				num := decimal.NewFromFloat(keepNum).Mul(unitNum).Round(3).InexactFloat64()
				subProduct.SetNum(num)
			}
		}
	}

	// 如果退菜商品是下单减库存的商品，则需要创建入库单
	var warehouseForm *model.WarehouseForm
	if returnSaleOrderProduct.DeductStockType == constant.ProductPackageDeductStockTypeCooking {
		currentReturnSaleOrderProduct := *returnSaleOrderProduct
		currentReturnSaleOrderProduct.Num = req.Num // 退菜数量,仅入库本次退菜的数量

		cookingDeductSaleOrderProducts := []*model.SaleOrderProduct{&currentReturnSaleOrderProduct}

		// 如果是套餐商品，则加上子商品的退菜信息
		if currentReturnSaleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(currentReturnSaleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				product := *subProduct
				product.Num = decimal.NewFromFloat(req.Num).Mul(decimal.NewFromFloat(subProduct.GetUnitNum())).Round(3).InexactFloat64() // 退菜数量,仅入库本次退菜的数量
				cookingDeductSaleOrderProducts = append(cookingDeductSaleOrderProducts, &product)
			}
		}

		productList, err := s.getDecreaseStockList(ctx, cookingDeductSaleOrderProducts)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		warehouseForm = model.NewWarehouseForm(productList, req.SaleBillUuid)
	}

	// 新建一个销售订单商品，该商品数量为移动数量
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 如果退菜商品还关联着分批标签，则解除
		if returnSaleOrderProduct.IsBatchBool() {
			returnSaleOrderProduct.BatchTagUuid = 0 // 手动置0，否则可能又被updateSaleOrderProduct方法更新回非零值
			returnSaleOrderProduct.BatchTime = 0
			if err := tx.Model(&model.SaleOrderProduct{}).Where("uuid = ?", returnSaleOrderProduct.Uuid).Updates(map[string]any{
				"batch_tag_uuid": 0,
				"batch_time":     0,
			}).Error; err != nil {
				return errors.WithMessage(err)
			}
		}
		// 创建退菜记录
		if len(returnFoodReasonList) > 0 {
			if err := repository.NewSaleOrderProductReasonRepo(tx).CreateSaleOrderProductReasons(returnFoodReasonList); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}

		if warehouseForm != nil && len(warehouseForm.WarehouseFormItems) > 0 {
			// 创建入库单
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseFormRecord(*warehouseForm); err != nil {
				return errors.WithMessage(err)
			}
			// 创建入库单记录
			if err := repository.NewWarehouseFormRepo(tx).CreateWarehouseFormItemRecords(warehouseForm.WarehouseFormItems); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 修改送厨商品
		productionRepo := repository.NewProductionRepo(tx)
		if keepNum > 0 { // 修改数量
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)},
				map[string]any{"num": keepNum}); err != nil {
				return errors.WithMessage(err)
			}
		} else { // 修改数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)},
				map[string]any{"num": 0}); err != nil {
				return errors.WithMessage(err)
			}
		}

		// 修改送厨子商品的数量
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				unitNum := decimal.NewFromFloat(subProduct.GetUnitNum())
				num := decimal.NewFromFloat(keepNum).Mul(unitNum).Round(3).InexactFloat64()
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(subProduct.Uuid)},
					map[string]any{"num": num}); err != nil {
					return errors.WithMessage(err)
				}
			}
		}

		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}

	// 发布"退菜"事件
	utils.Go(func() {
		var convertToEventOrderProduct func(saleOrderProduct *model.SaleOrderProduct, isSubProduct bool) event.CancelSaleOrderProductProduct
		convertToEventOrderProduct = func(saleOrderProduct *model.SaleOrderProduct, isSubProduct bool) event.CancelSaleOrderProductProduct {
			return event.CancelSaleOrderProductProduct{
				OrderProductId:  req.SaleOrderProductUuid,
				ProductId:       saleOrderProduct.ProductPackageUuid,
				ProductName:     saleOrderProduct.MultiLanguageName.GetNames(),
				ProductAttr:     saleOrderProduct.GetAttributeName(),
				ProductAttrList: saleOrderProduct.GetAttributeNameList(),
				TotalNum: func() float64 {
					if !isSubProduct {
						return req.Num
					}
					return decimal.NewFromFloat(req.Num).Mul(decimal.NewFromFloat(saleOrderProduct.GetUnitNum())).Round(3).InexactFloat64()
				}(),
				IsBuffet:     saleOrderProduct.IsBuffet == 1,
				Remark:       saleOrderProduct.Remark,
				Reason:       model.GetReturnFoodReasonNames(returnFoodReason),
				CustomReason: saleOrderProduct.CancelReason,
				Sign:         saleOrderProduct.Sign,
				SubProducts: func() []event.CancelSaleOrderProductProduct {
					if saleOrderProduct.IsPackageProduct() {
						subProducts := make([]event.CancelSaleOrderProductProduct, 0)
						for _, subProduct := range saleOrder.GetPackageSubProductList(returnSaleOrderProduct.Uuid) {
							subProducts = append(subProducts, convertToEventOrderProduct(subProduct, true))
						}
						return subProducts
					}
					return nil
				}(),
			}
		}
		s.bus.PublishCancelSaleOrderProductEvent(event.CancelSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 退菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			CancelSaleOrderProductProduct: convertToEventOrderProduct(returnSaleOrderProduct, false),
		})
	})

	// 送厨成功后，推送更新订单
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
			"update_time": time.Now().Unix(),
		})
	})

	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderCartProductCancelReturning 取消退菜购物车商品
func (s *orderSrv) InstantOrderCartProductCancelReturning(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品未取消")
	}
	// 更新销售订单商品
	errUpdate := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductReasons(
			saleOrderProduct.SaleOrderUuid,
			saleOrderProduct.Uuid,
			constant.ProductReasonTypeReturnFood,
		); err != nil {
			return errors.WithMessage(err)
		}
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			Status:       0,
			CancelTime:   0,
			CancelReason: "",
		}, map[string]bool{
			"Status":       true,
			"CancelTime":   true,
			"CancelReason": true,
		})
		// 更新套餐子商品
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					Status:       0,
					CancelTime:   0,
					CancelReason: "",
				}, map[string]bool{
					"Status":       true,
					"CancelTime":   true,
					"CancelReason": true,
				})
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.New("更新数据失败")
		}
		return nil
	})
	if errUpdate != nil {
		return nil, errors.WithMessage(errUpdate, "操作失败")
	}

	// 发布取消退菜事件
	utils.Go(func() {
		s.bus.PublishCancelReturnSaleOrderProductEvent(event.CancelReturnSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 取消退菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			Num:            saleOrderProduct.Num,
			ParentId:       saleOrder.SaleBillUuid,
			OrderName:      saleOrder.Uuid,
			Sign:           saleOrderProduct.Sign,
		})
	})

	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// InstantOrderCartProductChangeDesk 转菜购物车商品
func (s *orderSrv) InstantOrderCartProductChangeDesk(ctx context.Context, req req.OrderCartProductChangeDeskReq) (*resp.ShopCart, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 禁止并发操作：在方法开头就加锁，确保获取目标订单UUID时使用的是最新数据
	if ctx.NoLock() {
		systemLock := lock.NewSystemLock()

		// 获取目标订单 UUID（通过目标桌台 UUID 查询，在锁内获取确保数据一致性）
		targetSaleBillUuid, err := repository.NewDeskRepo(db).GetSaleBillUuidByDeskUuid(req.DeskUuid)
		if err != nil {
			return nil, errors.WithMessage(err, "获取目标桌台的销售账单UUID失败")
		}
		if targetSaleBillUuid == 0 {
			return nil, errors.WithMessage(errors.New("目标桌台没有关联订单"))
		}

		// 收集需要锁定的订单：源订单 + 目标订单
		orderUuids := []uint64{req.SaleBillUuid, targetSaleBillUuid}

		// 锁定源订单和目标订单（按 UUID 排序）
		// LockMultipleUuids 会自动去重和排序，返回排序后的 UUID 列表
		lockedUuids := lock.LockMultipleUuids(systemLock, orderUuids)

		// 按相反顺序释放锁（UnlockMultipleUuids 内部会使用相同的排序策略）
		defer func() {
			lock.UnlockMultipleUuids(systemLock, lockedUuids)
		}()

		ctx.AddLock()
	}

	// 获取验证销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderChangeTable, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}

	// 获取目标桌台的信息和销售账单（在锁内获取，确保数据一致性）
	targetDesk, err := repository.NewDeskRepo(db).GetDeskAndSaleBillByDeskUuid(req.DeskUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if err := targetDesk.CheckProductChangeDesk(); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取目标桌台的销售账单信息
	targetSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(targetDesk.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "目标桌台的销售账单不存在", fmt.Sprintf("targetDesk.SaleBillUuid: %d", targetDesk.SaleBillUuid))
	}
	// 不能转菜到自助餐桌台
	if targetSaleBill.IsBuffetSaleBill() {
		return nil, errors.WithMessage(errors.New("不能转菜到自助餐桌台"))
	}

	// 将销售订单商品的sale_bill_uuid和sale_order_uuid更新为新的桌台
	targetSaleOrder := targetSaleBill.GetFirstSaleOrder()

	// 判断订单状态
	if err := targetSaleOrder.ValidateOrderStatus(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 更新数据
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 处理订单商品
		var handleOrderProducts func(saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct) error
		handleOrderProducts = func(saleOrder *model.SaleOrder, saleOrderProduct *model.SaleOrderProduct) error {
			// 获取原生产单Uuid
			oldProductionOrderUuid := saleOrderProduct.ProductionOrderUuid

			var newProductionOrderUuid uint64
			if oldProductionOrderUuid != 0 {
				newProductionOrderUuid, _ = utils.GetID()
			}

			// 修改生产单Uuid
			saleOrderProduct.ProductionOrderUuid = newProductionOrderUuid
			// 设置已接单
			saleOrderProduct.SetAcceptOrderProduct()
			// h5订单只有一个商品时拒单
			if saleOrderProduct.H5OrderUuid != 0 {
				// 如果这个商品是h5订单的最后一个商品，则删除拒单该h5订单
				s.deleteOrRejectH5OrderProduct(ctx, db, saleOrderProduct)
				// 删除掉h5订单uuid
				saleOrderProduct.H5OrderUuid = 0
			}
			// 更新销售订单商品
			saleOrderProduct.SaleBillUuid = targetDesk.SaleBillUuid
			saleOrderProduct.SaleOrderUuid = targetSaleOrder.Uuid
			// 将商品的折扣改为使用目标桌台的折扣
			saleOrderProduct.MemberDiscountRate = targetSaleOrder.MemberDiscountRate
			saleOrderProduct.MemberCardDiscountRate = targetSaleOrder.MemberCardDiscountRate
			saleOrderProduct.CustomDiscountRate = targetSaleOrder.CustomDiscountRate
			// 判断商品是否在目标桌台是否是自助餐商品
			productPackageUuidMap := targetSaleBill.GetBuffetProductMap()
			if _, ok := productPackageUuidMap[saleOrderProduct.ProductPackageUuid]; ok {
				saleOrderProduct.SetIsBuffet()
			} else {
				saleOrderProduct.SetNotBuffet()
			}
			saleOrderProduct.SetUpdate() // 标记该商品的记录要更新，会在原桌台账单的CalcAndSaveSaleBill方法中更新

			// 将商品添加到目标桌台的销售订单中
			targetSaleOrder.SaleOrderProducts = append(targetSaleOrder.SaleOrderProducts, saleOrderProduct)

			// 重新计算原桌台的销售账单
			if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
				return errors.WithMessage(err)
			}

			// 重新计算目标桌台的销售账单
			if err := s.CalcAndSaveSaleBill(ctx, tx, targetSaleBill); err != nil {
				return errors.WithMessage(err)
			}

			// 更新生产单
			productionRepo := repository.NewProductionRepo(db)
			oldProductionOrder, _ := productionRepo.GetProductionOrder(productionRepo.WhereUuid(oldProductionOrderUuid))
			if oldProductionOrder != nil {
				if err := productionRepo.CreateProductionOrder(&model.ProductionOrder{
					BaseModel:     model.BaseModel{Uuid: newProductionOrderUuid},
					DeskUuid:      targetDesk.Uuid,
					SaleOrderUuid: targetSaleOrder.Uuid,
					SaleBillUuid:  targetSaleBill.Uuid,
					Source:        oldProductionOrder.Source,
				}); err != nil {
					return errors.WithMessage(err)
				}
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(saleOrderProduct.Uuid)}, map[string]any{
					"sale_bill_uuid":        targetSaleBill.Uuid,
					"production_order_uuid": newProductionOrderUuid,
				}); err != nil {
					return errors.WithMessage(err)
				}
			}

			// 如果是套餐商品，则更新套餐子商品
			if saleOrderProduct.IsPackageProduct() {
				subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
				for _, subProduct := range subProducts {
					handleOrderProducts(saleOrder, subProduct)
				}
			}

			return nil
		}
		return handleOrderProducts(saleOrder, saleOrderProduct)

	}); err != nil {
		return nil, errors.WithMessage(err, "更新数据失败")
	}

	// 发布转菜事件
	utils.Go(func() {
		s.bus.PublishChangeDeskSaleOrderProductEvent(event.ChangeDeskSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 转菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			TotalNum:       saleOrderProduct.Num,
			ToOrderId:      targetSaleOrder.Uuid,
			ToTableId:      targetDesk.Uuid,
			ToTableNo:      targetDesk.DeskNo,
		})
	})

	// 获取新的购物车信息
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo, "获取购物车信息失败")
	}

	return cartInfo, nil
}

// OrderCartProductWrap 打包购物车商品
func (s *orderSrv) OrderCartProductWrap(ctx context.Context, req req.OrderCartProductWrapReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}

	// 设置打包时间
	saleOrderProduct.SetWrap()

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetWrap()
		}
	}

	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			WrapTime: saleOrderProduct.WrapTime,
			Sign:     saleOrderProduct.Sign,
		})
		// 更新套餐子商品打包时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					WrapTime: subProduct.WrapTime,
					Sign:     subProduct.Sign,
				})
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	// 发布打包事件
	utils.Go(func() {
		s.bus.PublishWrapSaleOrderProductEvent(event.WrapSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 打包
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleOrderProductUuid: req.SaleOrderProductUuid,
			ProductPackageUuid:   saleOrderProduct.ProductPackageUuid,
			ProductName:          saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:          saleOrderProduct.GetAttributeName(),
			ProductPrice:         saleOrderProduct.Price,
			Num:                  saleOrderProduct.Num,
		})
	})
	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo)
	}
	return cartInfo, nil
}

// OrderCartProductUnwrap 取消打包购物车商品
func (s *orderSrv) OrderCartProductUnwrap(ctx context.Context, req req.OrderCartProductUnwrapReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}

	// 设置打包时间
	saleOrderProduct.SetUnwrap()
	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetUnwrap()
		}
	}
	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			WrapTime: saleOrderProduct.WrapTime,
		})
		// 更新套餐子商品打包时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					WrapTime: subProduct.WrapTime,
				})
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	// 发布打包事件
	utils.Go(func() {
		s.bus.PublishUnwrapSaleOrderProductEvent(event.UnwrapSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 打包
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			SaleOrderProductUuid: req.SaleOrderProductUuid,
			ProductPackageUuid:   saleOrderProduct.ProductPackageUuid,
			ProductName:          saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:          saleOrderProduct.GetAttributeName(),
			ProductPrice:         saleOrderProduct.Price,
			Num:                  saleOrderProduct.Num,
		})
	})
	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo)
	}
	return cartInfo, nil
}

// InstantOrderCartProductGiving 赠送购物车商品
func (s *orderSrv) InstantOrderCartProductGiving(ctx context.Context, req req.OrderCartProductGivingReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}

	if ctx.GetSource() == constant.SourceAssistant && saleBill.IsSplit() {
		return nil, errors.WithMessage(errors.New("当前订单已拆单，请前去收银机操作"))
	}

	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case saleOrderProduct.IsCancelProduct():
		return nil, errors.New("商品已取消")
	}
	//  验证赠菜标签
	reasons := [][2]uint64{}
	if len(req.GiftIds) > 0 {
		_reasons, notFound, err := base.NewGiftOrFreeOrderReasonRepo(db).ExistsByUuids(req.GiftIds)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if len(notFound) > 0 {
			return nil, fmt.Errorf("以下赠菜原因不存在: %v", notFound)
		}
		reasons = _reasons
	}
	// 设置赠菜时间
	saleOrderProduct.SetGiftProduct(req.Reason)
	// 如果是套餐商品，则更新套餐子商品赠菜时间
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
		for _, subProduct := range subProducts {
			subProduct.SetGiftProduct(req.Reason)
		}
	}

	// 执行
	if errUpdateDB := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			GiftTime:   saleOrderProduct.GiftTime,
			Sign:       saleOrderProduct.Sign,
			GiftReason: saleOrderProduct.GiftReason,
		})
		// 更新套餐子商品赠菜时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					GiftTime:   subProduct.GiftTime,
					Sign:       subProduct.Sign,
					GiftReason: subProduct.GiftReason,
				})
			}
		}
		// 添加赠菜原因
		if len(reasons) > 0 {
			if err := repository.NewSaleOrderProductRepo(tx).CreateSaleOrderProductReasons(
				saleOrderProduct.SaleOrderUuid,
				saleOrderProduct.Uuid,
				constant.ProductReasonTypeGift,
				reasons,
			); err != nil {
				return errors.WithMessage(err)
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); errUpdateDB != nil {
		return nil, errors.WithMessage(errUpdateDB, "更新数据失败")
	}
	// 发布赠菜事件
	utils.Go(func() {
		s.bus.PublishGiftSaleOrderProductEvent(event.GiftSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 赠菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			ProductPrice:   saleOrderProduct.Price,
			TotalNum:       saleOrderProduct.Num,
			TotalPrice:     decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromInt(int64(saleOrderProduct.Num))).Round(2).InexactFloat64(),
			FreeTagIds:     req.GiftIds,
			FreeRemark:     req.Reason,
		})
	})
	// 获取新的购物车信息
	var cartInfo *resp.ShopCart
	cartInfo, errGetCartInfo := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if errGetCartInfo != nil {
		return nil, errors.WithMessage(errGetCartInfo)
	}
	return cartInfo, nil
}

// InstantOrderCartProductCancelGiving 取消赠送购物车商品
func (s *orderSrv) InstantOrderCartProductCancelGiving(ctx context.Context, req req.OrderCartProduct) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	// 获取验证销售账单信息
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "销售账单不存在")
	}
	// 判断订单状态
	if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderCancelRefundProduct, req.SaleOrderUuid); err != nil {
		return nil, errors.WithMessage(err)
	}
	// 获取销售订单和商品
	saleOrder, saleOrderProduct := saleBill.GetSaleOrderAndProduct(req.SaleOrderUuid, req.SaleOrderProductUuid)
	switch {
	case saleOrder == nil:
		return nil, errors.New("销售订单不存在")
	case saleOrderProduct == nil:
		return nil, errors.New("销售订单商品不存在")
	case !saleOrderProduct.IsGiftProduct():
		return nil, errors.New("商品未赠送")
	}

	// 如果是套餐商品，则更新套餐子商品数量
	subProducts := make([]*model.SaleOrderProduct, 0)
	if saleOrderProduct.IsPackageProduct() {
		subProducts = saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
	}

	// 更新销售订单商品
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewSaleOrderProductRepo(db).DeleteSaleOrderProductReasons(
			saleOrderProduct.SaleOrderUuid,
			saleOrderProduct.Uuid,
			constant.ProductReasonTypeGift,
		); err != nil {
			return errors.WithMessage(err)
		}
		saleBill.SetProductFields(saleOrderProduct.Uuid, model.SaleOrderProduct{
			GiftTime:   0,
			GiftReason: "",
		}, map[string]bool{
			"GiftTime":   true,
			"GiftReason": true,
		})
		// 更新套餐子商品赠菜时间
		if len(subProducts) > 0 {
			for _, subProduct := range subProducts {
				saleBill.SetProductFields(subProduct.Uuid, model.SaleOrderProduct{
					GiftTime:   0,
					GiftReason: "",
				}, map[string]bool{
					"GiftTime":   true,
					"GiftReason": true,
				})
			}
		}
		// 计算订单商品、订单、账单金额并更新或创建
		if err := s.CalcAndSaveSaleBill(ctx, tx, saleBill); err != nil {
			return errors.New("更新数据失败")
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err, "操作失败")
	}

	// 发布取消赠菜事件
	utils.Go(func() {
		s.bus.PublishCancelGiftSaleOrderProductEvent(event.CancelGiftSaleOrderProductPayload{
			BasePayload: event.BasePayload{ // 取消赠菜
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  req.SaleBillUuid,
				SaleOrderUuid: req.SaleOrderUuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
			},
			OrderProductId: req.SaleOrderProductUuid,
			ProductId:      saleOrderProduct.ProductPackageUuid,
			ProductName:    saleOrderProduct.MultiLanguageName.GetNames(),
			ProductAttr:    saleOrderProduct.GetAttributeName(),
			ProductPrice:   saleOrderProduct.Price,
			TotalNum:       saleOrderProduct.Num,
			TotalPrice:     decimal.NewFromFloat(saleOrderProduct.Price).Mul(decimal.NewFromInt(int64(saleOrderProduct.Num))).Round(2).InexactFloat64(),
			ParentId:       saleOrder.SaleBillUuid,
			OrderName:      saleOrder.Uuid,
		})
	})

	// 获取新的购物车信息
	cartInfo, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取购物车信息失败")
	}
	return cartInfo, nil
}

// GetOrderCartInfo 获取点餐购物车信息
func (s *orderSrv) GetOrderCartInfo(ctx context.Context, saleBillUuid uint64, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 追加请求头参数，从http的header中获取h5_order_uuid
	h5OrderUuid := context.GetH5OrderUuid(ctx)
	if h5OrderUuid != 0 {
		opts = append(opts, repository.WithH5OrderUuid(h5OrderUuid))
	}

	option := &repository.OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	db := ctx.GetDB()
	orderRepo := repository.NewOrderRepo(db)

	// 通过销售订单ID得到订单商品列表、订单金额信息、账单的销售订单列表
	shopCart, err := orderRepo.GetOrderCartInfo(saleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("saleBillUuid: %d", saleBillUuid))
	}
	if !option.FilterEndStatus && shopCart.SaleBill.IsEndStatus() {
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeDeskOrderEnd, "桌台订单结束"))
	}
	// 重新计算金额
	if option.H5OrderUuid == 0 {
		shopCart.SaleBill.CalcAll()
	} else {
		// 如果从待接单进入桌台时，要把h5订单商品计算在内
		shopCart.SaleBill.CalcAll(model.WithAllAndOneH5Order(option.H5OrderUuid))
	}

	// 给订单列表添加订单
	saleOrderList := make([]resp.SaleOrder, 0)
	for _, saleOrder := range shopCart.SaleBill.SaleOrders {
		productList := make([]resp.Product, 0)
		// 给商品列表条件顾客类型
		// 如不是桌台订单、不是自助餐，这个Buffets列表是空的，故不会往productList里加入商品
		{
			customerList := saleOrder.GetCustomerList()
			productList = append(productList, customerList...)
		}

		// 添加加钟商品
		{
			delayProductList := saleOrder.GetDelayProductList()
			productList = append(productList, delayProductList...)
		}

		// 添加正常商品
		{
			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			products := saleOrder.GetProductList(option.UnorderedH5Product == repository.OrderedH5ProductWithReject, businessSetting.OpenIsBatch())
			productList = append(productList, products...)
		}

		// 商品计数
		productNum := 0.0
		for _, product := range productList {
			// 退菜的商品不计入
			if product.IsCancel {
				continue
			}
			productNum += product.Num
		}

		// 填写订单信息
		order := resp.SaleOrder{
			Uuid:                saleOrder.Uuid,
			OrderNo:             saleOrder.OrderNo,
			Status:              saleOrder.Status,
			ProductNum:          decimal.NewFromFloat(productNum).Truncate(3).InexactFloat64(),
			ProductList:         productList,
			IsDiscount:          saleOrder.IsManualDiscount(uint8(shopCart.SaleBill.SaleBillSetting.ZeroRule)),
			IsMemberDiscount:    saleOrder.IsMemberDiscount(),
			CustomDiscountRate:  saleOrder.CustomDiscountRate,
			ZeroRule:            saleOrder.ZeroRule,
			AutoDiscountMessage: saleOrder.GetAutoDiscountMessage(*shopCart.SaleBill.SaleBillSetting, ctx.GetLanguage()),
			// 订单金额信息
			AmountInfo: resp.AmountInfo{
				ProductOriginalAmount: saleOrder.ProductOriginalAmount,
				ProductAmount:         saleOrder.ProductAmount,
				ServiceAmount:         saleOrder.ServiceFee,
				TaxAmount:             saleOrder.TaxFee,
				DiscountAmount:        decimal.NewFromFloat(saleOrder.CustomDiscountFee).Round(2).InexactFloat64(),
				MemberDiscountAmount:  saleOrder.MemberDiscountFee,
				Amount:                saleOrder.GetAmount(),
			},
		}
		saleOrderList = append(saleOrderList, order)
	}

	takeout := shopCart.SaleBill.IsTakeout()
	orderRemark, _ := shopCart.SaleBill.GetOrderRemarkRes()
	shopCartInfo := &resp.ShopCart{
		SaleBillUuid:    saleBillUuid,
		IsDeskOrder:     shopCart.IsDeskShopCart(),
		IsLock:          shopCart.SaleBill.IsLockStatus(),
		OrderSourceUuid: shopCart.SaleBill.OrderSourceUuid,
		NationalityUuid: shopCart.SaleBill.NationalityUuid,
		Takeout:         &takeout,
		Desk:            nil,
		Buffet:          nil,
		DiningMethod:    shopCart.SaleBill.DiningMethod,
		SaleOrderList:   saleOrderList,
		UpdateTime:      shopCart.SaleBill.UpdateTime,
		OrderRemark:     orderRemark,
	}
	// 如果是桌台购物车
	if shopCart.IsDeskShopCart() {
		deskInfo := resp.DeskInfo{
			Uuid:      shopCart.SaleBill.Desk.Uuid,
			DeskNo:    shopCart.SaleBill.Desk.DeskNo,
			MealNum:   shopCart.SaleBill.MealNum,
			StartTime: shopCart.SaleBill.CreateTime,
			Duration:  time.Now().Unix() - shopCart.SaleBill.CreateTime,
		}
		shopCartInfo.Desk = &deskInfo
		// 如果是自助餐桌台
		if shopCart.SaleBill.IsBuffetSaleBill() {
			shopCartInfo.Buffet = &resp.BuffetInfo{
				RemainingSeconds:      shopCart.SaleBill.GetTotalRemainingSeconds(),
				IsTimeLimited:         shopCart.SaleBill.IsTimeLimited(),
				LocaleName:            shopCart.SaleBill.GetBuffetName(),
				RemainingOrderingTime: shopCart.SaleBill.GetRemainingOrderingSeconds(),
				ReminderOrderTime:     int64(shopCart.SaleBill.ReminderOrderTime) * 60,
				IsTabletH5TimeSet:     shopCart.SaleBill.IsBuffetTabletH5TimeSet(),
			}
		}
	}

	// 获取必点方案
	if !option.NoQueryMustPlan {
		saleOrder := shopCart.SaleBill.GetFirstSaleOrder()
		var productMustPlanList *resp.ProductMustPlanList
		// 如果销售账单需要显示必点方案
		if shopCart.SaleBill.IsShowMustPlan() {
			// 获取必点方案列表
			var mustPlan *resp.InstantProductMustPlanResp
			var isAutoAdd bool
			if shopCart.SaleBill.IsDeskSaleBill() {
				mustPlan, isAutoAdd, err = s.DeskOrderMustPlan(ctx, saleBillUuid, saleOrder.Uuid, shopCart.SaleBill.MealNum, option.H5AutoAdd, option.NoAutoAdd)
				if err != nil {
					ctx.Log().Info("获取桌台必点方案列表失败", zap.Error(errors.WithMessage(err)))
					if !shopCart.SaleBill.IsEndStatus() {
						return nil, errors.WithMessage(errors.New("获取桌台必点方案列表失败"), err.Error())
					}
				}
			} else {
				mustPlan, isAutoAdd, err = s.InstantOrderMustPlan(ctx, ctx.GetDeviceSn())
				if err != nil {
					ctx.Log().Info("获取点餐必点方案列表失败", zap.Error(errors.WithMessage(err)))
					if !shopCart.SaleBill.IsEndStatus() {
						return nil, errors.WithMessage(errors.New("获取点餐必点方案列表失败"), err.Error())
					}
				}
			}

			if mustPlan != nil {
				// 如果已经自动加购完成，则不在显示必点方案
				finish := true
				for _, mustPlan := range mustPlan.List {
					// 如果必点方案中还有需要加购的商品，则显示必点方案
					if mustPlan.NeedNum != 0 {
						finish = false
					}
				}
				if isAutoAdd {
					opts = append(opts, repository.WithCanCloseMustPlanView()) // 设置可以关闭必点弹窗
					return s.GetOrderCartInfo(ctx, saleBillUuid, opts...)
				} else {
					productMustPlanList = &resp.ProductMustPlanList{
						List: mustPlan.List,
					}

					// 如果在可以关闭必点弹窗的场景下，且已经自动加购完成
					if option.CanCloseMustPlanView && finish {
						// 如果已经自动加购完成，则不在显示必点方案.并更新sale_bill为已完成必点
						err := repository.NewSaleBillRepo(db).UpdateSaleBillShowMustPlan(saleBillUuid)
						if err != nil {
							return nil, errors.WithMessage(err)
						}
						// 清空必点方案, 不给前端返回必点数据
						productMustPlanList = nil

					}
				}
			}
		}
		// 如果要显示必点信息
		if productMustPlanList != nil {
			shopCartInfo.MustPlans = productMustPlanList
		}

		// 判断是否需要弹出分批送厨弹窗。只有收银机和助手端需要判断
		if ctx.GetSource() == constant.SourceAssistant || ctx.GetSource() == constant.SourceCashier {
			// 获取门店业务设置
			businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			if businessSetting.OpenIsBatch() {
				if shopCart.SaleBill.IsNeedBatchSendCooking() {
					shopCartInfo.SetCode(constant.CodeOrderCheckProductBatch)
				}
			}
		}
	}
	return shopCartInfo, nil
}

// OrderCartProductPackageAdd 往购物车添加套餐
func (s *orderSrv) OrderCartProductPackageAdd(ctx context.Context, request req.OrderCartProductPackageAddReq) (*resp.ShopCart, error) {

	// 当不填销售账单ID时，表示要新建一个销售账单
	if request.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			request.SaleBillUuid = billInfo.Uuid
			request.SaleOrderUuid = billInfo.SaleOrders[0].Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			request.SaleBillUuid = order.SaleBillUuid
			request.SaleOrderUuid = order.SaleOrderUuid
		}
	}

	// 上锁防止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 兼容2.10.0之前的版本. 通过product_bom_uuid查询套餐product_package_uuid
	productPackageUuid := func() uint64 {
		if ctx.Version(context.LT, constant.ClientVersionV2100) {
			if productPackageUuid, err := repository.NewProductBomRepo(db).GetProductPackageUuidByBomUuid(request.ProductPackageUuid); err != nil {
				ctx.Log().Error("通过商品bom uuid查询套餐product_package_uuid失败", zap.Uint64("company_uuid", ctx.GetCompanyUuid()), zap.Uint64("product_bom_uuid", request.ProductPackageUuid), zap.Error(err))
				return 0
			} else {
				return productPackageUuid
			}
		}
		return request.ProductPackageUuid
	}()

	// 查询套餐分组配置，用于验证分组选择
	productPackage, err := repository.NewProductPackageRepo(db).GetProductPackage(
		repository.CommonRepo.WhereByUuid(productPackageUuid),
		repository.CommonRepo.WhereBySoftDelete(),
		repository.NewProductPackageRepo(db).WithProductPackageGroups(
			repository.CommonRepo.WhereBySoftDelete(),
		),
		repository.NewProductPackageRepo(db).WithProductPackageGroupMultiLanguageName(),
		repository.NewProductPackageRepo(db).WithProductPackageGroupItems(
			repository.CommonRepo.WhereBySoftDelete(),
		),
		repository.NewProductPackageRepo(db).WithProductBoms(
			repository.CommonRepo.WhereBySoftDelete(),
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询套餐信息失败")
	}

	// 分组选择验证
	if err := s.validatePackageGroupSelection(ctx, request.Products, productPackage.ProductPackageGroups); err != nil {
		return nil, errors.WithMessage(err)
	}

	productPackageFlavorBomUuid := func() uint64 {
		if len(productPackage.ProductBoms) == 0 {
			return 0
		}
		return productPackage.ProductBoms[0].Uuid
	}()
	if productPackageFlavorBomUuid == 0 {
		return nil, errors.WithMessage(errors.New("套餐商品规格不存在"), "套餐商品规格不存在")
	}

	// 往销售账单里添加商品
	productParam := req.ProductParams{
		FlavorProductBomUuid: productPackageFlavorBomUuid,
		Num:                  request.Num,
		Operation:            "add",
	}
	// 记录相关的子商品。
	subProducts := make([]req.ProductParams, 0)
	for _, productReq := range request.Products {
		subProduct := req.ProductParams{
			FlavorProductBomUuid:            productReq.FlavorUuid,
			Num:                             productReq.Num,
			UnitNum:                         productReq.UnitNum,
			ProductPackageAttributeUuidList: productReq.AttributeUuidList,
			ProductPackageGroupUuid:         productReq.ProductPackageGroupUuid,
			AddPrice:                        productReq.AddPrice, // 传递加价金额
			Operation:                       "add",
		}
		subProducts = append(subProducts, subProduct)
	}
	productParam.SetIsPackageProduct(subProducts) // 设置为套餐商品

	shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products: []req.ProductParams{
			productParam,
		},
		IsH5Product: request.IsH5Product(),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return shopCart, nil
}

// validatePackageGroupSelection 验证套餐分组选择是否符合套餐设置
func (s *orderSrv) validatePackageGroupSelection(ctx context.Context, selectedProducts []req.ProductRequest, packageGroups []model.ProductPackageGroup) error {
	// 按分组UUID分组统计已选商品
	groupSelectedMap := make(map[uint64][]req.ProductRequest)
	for _, product := range selectedProducts {
		if product.ProductPackageGroupUuid > 0 {
			groupSelectedMap[product.ProductPackageGroupUuid] = append(
				groupSelectedMap[product.ProductPackageGroupUuid],
				product,
			)
		}
	}

	// 遍历每个分组进行验证
	for _, group := range packageGroups {
		if group.IsDelete() {
			continue
		}

		selectedProducts := groupSelectedMap[group.Uuid]

		if group.GroupType == 0 {
			// 固定分组：验证是否包含所有商品
			// 获取分组内所有商品（未删除的）
			validItems := make([]model.ProductPackageGroupItem, 0)
			for _, item := range group.ProductPackageGroupItems {
				if !item.IsDelete() {
					validItems = append(validItems, item)
				}
			}

			if len(selectedProducts) != len(validItems) {
				groupName := group.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				return errors.New(fmt.Sprintf("固定分组「%s」必须选择所有商品，请重新选择", groupName))
			}

			// 验证商品UUID是否匹配
			selectedUuidMap := make(map[uint64]bool)
			for _, p := range selectedProducts {
				selectedUuidMap[p.FlavorUuid] = true
			}

			for _, item := range validItems {
				if !selectedUuidMap[item.ProductBomUuid] {
					groupName := group.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					return errors.New(fmt.Sprintf("固定分组「%s」必须选择所有商品，请重新选择", groupName))
				}
			}
		} else {
			// 兼容2.10.0之前的版本. 提示请升级客户端版本。可选分组功能在2.10.0版本中引入
			if ctx.Version(context.LT, constant.ClientVersionV2100) {
				return errors.WithMessage(errors.New("请升级客户端版本到2.10.0及以上版本"), "请升级客户端版本到2.10.0及以上版本")
			}

			// 可选分组：验证已选数量是否等于 optional_count
			selectedCount := 0.0
			for _, p := range selectedProducts {
				selectedCount += p.Num // 按份数统计
			}

			if int(selectedCount) != group.OptionalCount {
				groupName := group.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				diff := group.OptionalCount - int(selectedCount)
				if diff > 0 {
					return errors.New(fmt.Sprintf("该分组「%s」需要选择 %d 个商品，当前已选 %d 个，还差 %d 个",
						groupName, group.OptionalCount, int(selectedCount), diff))
				} else {
					return errors.New(fmt.Sprintf("该分组「%s」最多选择 %d 个商品，当前已选 %d 个，请删除多余商品",
						groupName, group.OptionalCount, int(selectedCount)))
				}
			}
		}
	}

	return nil
}

func (s *orderSrv) OrderCartProductFlavorAndAttribute(ctx context.Context, request req.OrderCartProductFlavorAndAttributeReq) (*resp.ProductFlavorAndAttributeRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("销售订单不存在"), "销售订单不存在")
	}

	saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
	if errProduct != nil {
		return nil, errors.WithMessage(errProduct)
	}

	if saleOrderProduct.IsPackageProduct() {
		product, errProduct := s.GetProductDetail(ctx, saleOrderProduct.ProductPackageUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		productDetail := saleOrderProduct.GetPackageDetail()
		return &resp.ProductFlavorAndAttributeRes{
			ProductType: uint(saleOrderProduct.ProductType),
			ProductInfo: &resp.ProductInfo{
				Product: product,
				SelectedProductPackage: &resp.ProductSelectedInfoList{
					List: productDetail,
				},
			},
		}, nil
	} else {
		product, errProduct := s.GetProductDetail(ctx, saleOrderProduct.ProductPackageUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		productDetail := saleOrderProduct.GetProductPackageDetail()
		return &resp.ProductFlavorAndAttributeRes{
			ProductType: uint(saleOrderProduct.ProductType),
			ProductInfo: &resp.ProductInfo{
				Product:         product,
				SelectedProduct: &productDetail,
			},
		}, nil
	}
}
func (s *orderSrv) OrderCartProductFlavorAndAttributeChange(ctx context.Context, request req.OrderCartProductFlavorAndAttributeChangeReq) (*resp.ShopCart, error) {

	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.WithMessage(errors.New("销售订单不存在"), "销售订单不存在")
	}

	// 如果是修改普通商品
	if request.ProductType == 0 {
		// 普通商品
		saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		if !saleOrderProduct.IsCanEdit() {
			return nil, errors.WithMessage(errors.New("商品不可编辑"), "商品不可编辑")
		}

		saleOrderProduct, err := EditProduct(ctx, db, saleOrder, saleOrderProduct, request.EditProductReq)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		// 套餐商品
		saleOrderProduct, _, errProduct := saleOrder.GetSaleOrderProduct(request.SaleOrderProductUuid)
		if errProduct != nil {
			return nil, errors.WithMessage(errProduct)
		}
		if !saleOrderProduct.IsCanEdit() {
			return nil, errors.WithMessage(errors.New("商品不可编辑"), "商品不可编辑")
		}
		// 获取套餐子商品
		subProducts := saleOrder.GetPackageSubProductList(request.SaleOrderProductUuid)
		subProductParamMap := make(map[string]req.ProductRequest) // 套餐子商品参数. key为商品规格uuid+套餐分组uuid
		for i := range request.Products {
			params := request.Products[i]
			subProductParamMap[fmt.Sprintf("%d-%d", params.FlavorUuid, params.ProductPackageGroupUuid)] = params
		}
		if len(subProducts) != len(subProductParamMap) {
			return nil, errors.WithMessage(errors.New("修改前后套餐子商品数量不一致"), "修改前后套餐子商品数量不一致")
		}
		for _, subProduct := range subProducts {
			key := fmt.Sprintf("%d-%d", subProduct.GetFlavorSaleOrderProductBom().ProductBomUuid, subProduct.PackageGroupUuid)
			params, ok := subProductParamMap[key]
			if !ok {
				return nil, errors.WithMessage(errors.New("套餐子商品不存在"), "套餐子商品不存在")
			}
			_, err := EditProduct(ctx, db, saleOrder, subProduct, params.EditProductReq)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
		}
		// 重新记录套餐商品的子商品参数
		saleOrderProduct.PackageSubProductParams = func() string {
			subProductList := make([]req.SubProduct, 0)
			for i := range subProducts {
				params := request.Products[i]
				subProductList = append(subProductList, req.SubProduct{
					FlavorUuid:              params.FlavorUuid,
					AttributeUuid:           params.AttributeUuidList,
					ProductPackageGroupUuid: params.ProductPackageGroupUuid,
					Num:                     params.Num,
					UnitNum:                 params.UnitNum,
				})
			}
			return utils.ToJson(subProductList)
		}()
		// 重新计算套餐商品的签名
		saleOrderProduct.UpdateSign()
	}

	// 从新计算订单并保存
	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// InstantOrderCartProductAdd 点餐页面，往购物车添加商品。
func (s *orderSrv) InstantOrderCartProductAdd(ctx context.Context, request req.OrderCartProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	// 当不填销售账单ID时，表示要新建一个销售账单
	if request.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			request.SaleBillUuid = billInfo.Uuid
			request.SaleOrderUuid = billInfo.SaleOrders[0].Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			request.SaleBillUuid = order.SaleBillUuid
			request.SaleOrderUuid = order.SaleOrderUuid
		}
	}

	// 判断商品价格是否与后台设置的最新价格不一致
	// 查询商品规格的最新价格
	// 查询所选加料的最新价格
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	if request.Price != nil {
		productInfo, err := s.getInfo(ctx, req.ProductParams{
			FlavorProductBomUuid:    request.FlavorUuid,
			Price:                   request.Price,
			IsBuffet:                request.IsBuffet,
			SauceProductBomUuidList: request.SauceUuidList,
		}, db)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if productInfo != nil {
			return &resp.ShopCart{
				Product: productInfo,
			}, errors.ErrProductPriceChanged
		}
	}

	num := 1.0
	if request.Num != nil {
		num = *request.Num
	}

	// 往销售账单里添加商品
	shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products: []req.ProductParams{
			{
				FlavorProductBomUuid:            request.FlavorUuid,
				Num:                             num,
				SauceProductBomUuidList:         request.SauceUuidList,
				ProductPackageAttributeUuidList: request.AttributeUuidList,
				Operation:                       request.Operation,
				MustPlanUuid:                    request.MustPlanUuid,
				BatchTagUuid:                    request.BatchTagUuid,
			},
		},
		IsH5Product: request.IsH5Product(),
	}, opts...)

	// 往销售账单里添加商品
	// shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
	// 	SaleBillUuid:  request.SaleBillUuid,
	// 	SaleOrderUuid: request.SaleOrderUuid,
	// 	Products: []req.ProductParams{
	// 		{
	// 			FlavorProductBomUuid:            request.FlavorUuid,
	// 			Num:                             num,
	// 			SauceProductBomUuidList:         request.SauceUuidList,
	// 			ProductPackageAttributeUuidList: request.AttributeUuidList,
	// 			Operation:                       request.Operation,
	// 			MustPlanUuid:                    request.MustPlanUuid,
	// 		},
	// 	},
	// 	IsH5Product: request.IsH5Product(),
	// }, opts...)
	if err != nil {
		ctx.Log().Info("往点餐账单里添加商品失败", zap.Any("req", request), zap.Any("error", err))
		return nil, errors.WithMessage(err)
	}
	return shopCart, nil
}

// ChangeBatchTag 更换分批类型（前置模式）
func (s *orderSrv) ChangeBatchTag(ctx context.Context, req req.ChangeBatchTagReq) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}

	// 验证参数
	if err := req.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取业务设置，判断是否为前置模式
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if businessSetting.BatchCookingMode != constant.BatchCookingModePre {
		return nil, errors.WithMessage(errors.New("当前不是前置模式，不支持更换分批类型"))
	}

	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(req.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}

	// 验证新的 batch_tag_uuid 的有效性
	batchTagRepo := repository.NewBatchTagRepo(db)
	_, err = batchTagRepo.GetBatchTagInfo(req.BatchTagUuid)
	if err != nil {
		return nil, errors.WithMessage(fmt.Errorf("分批类型不存在"), err.Error())
	}

	// 获取要更换的商品列表
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, productUuid := range req.SaleOrderProductUuids {
		product := saleBill.GetSaleOrderProductByUuid(productUuid)
		if product == nil {
			continue
		}
		// 验证商品是否已送厨（已送厨则不允许修改）
		if !product.IsPreCooking() {
			return nil, errors.WithMessage(errors.New("商品已送厨，不能修改分批类型"))
		}
		// 验证商品是否为分批商品
		if !product.IsBatchBool() {
			return nil, errors.WithMessage(errors.New("商品不是分批商品，不能修改分批类型"))
		}
		saleOrderProducts = append(saleOrderProducts, product)
	}

	if len(saleOrderProducts) == 0 {
		return nil, errors.WithMessage(errors.New("未找到要更换的商品"))
	}

	// 更新商品的分批类型关联
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		for _, product := range saleOrderProducts {
			product.BatchTagUuid = req.BatchTagUuid
			// 更新签名（因为签名包含 batch_tag_uuid）
			product.UpdateSign()
		}
		// 更新数据库
		if err := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(saleOrderProducts); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, req.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

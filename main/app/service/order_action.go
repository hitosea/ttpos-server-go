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

// ActionCooking 送厨
func (s *orderSrv) ActionCooking(ctx context.Context, req req.OrderCartProductCookingReq, saleBill *model.SaleBill, unCookingSaleOrderProducts []*model.SaleOrderProduct) (*resp.OrderCheckServiceRes, error) {
	var productionOrder *model.ProductionOrder
	var warehouseOutForm *model.WarehouseOutForm

	if ctx.NoLock() {
		s.lock.LockUuid(req.SaleBillUuid)
		defer s.lock.UnlockUuid(req.SaleBillUuid)
		ctx.AddLock()
	}
	db := s.dbm.GetDB(ctx.GetDbId())

	// 送厨相关
	{
		// 获取所有商品,用于检查限购
		saleOrderProductAll := saleBill.GetSaleOrderProductAll()

		// 对商品进行送厨检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购、必点为选择
		checkServiceRes, errCheck := s.checkOrder(ctx, false, db, req.SaleBillUuid, saleBill.DeskUuid, unCookingSaleOrderProducts, saleOrderProductAll)
		if errCheck != nil {
			ctx.Log().Error("检查商品失败", zap.Error(errCheck))
			return nil, errors.New("检查商品失败")
		}
		if checkServiceRes != nil {
			if checkServiceRes.Code == constant.CodeOrderCheckProductMust && req.IgnoreMust {
				// 必点方案未选择，且忽略必点方案
			} else {
				return checkServiceRes, nil
			}
		}

		// 构建送厨单
		productionOrder := newProductionOrder(ctx, req.SaleOrderUuid, req.SaleBillUuid, unCookingSaleOrderProducts)

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
		warehouseOutForm = model.NewWarehouseOutForm(decreaseStockList, false, req.SaleBillUuid)
	}

	ctx.Log().Debug("准备开始更新")
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 送厨相关
		{
			// 修改订单商品状态为已送厨
			errUpdateSaleProductStatus := repository.NewSaleOrderProductRepo(tx).UpdateSaleOrderProductList(unCookingSaleOrderProducts)
			if errUpdateSaleProductStatus != nil {
				ctx.Log().Debug("商品状态更新失败", zap.Error(errUpdateSaleProductStatus))
				return errors.New(errUpdateSaleProductStatus.Error())
			}
			ctx.Log().Debug("商品状态成功")
			errCreateProduction := repository.NewProductionRepo(tx).CreateProductionOrder(productionOrder)
			if errCreateProduction != nil {
				ctx.Log().Debug("创建送厨单失败", zap.Error(errCreateProduction))
				return errors.New(errCreateProduction.Error())
			}

			// 如果账单有更新，则更新账单
			if saleBill.GetUpdate() {
				if err := repository.NewSaleBillRepo(tx).UpdateSaleBillRecord(*saleBill); err != nil {
					ctx.Log().Debug("更新账单失败", zap.Error(err))
					return errors.New(err.Error())
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
					SaleBillUuid:  req.SaleBillUuid,
					SaleOrderUuid: req.SaleOrderUuid,
					OperatorUuid:  int64(ctx.GetStaffUuid()),
				},
				Products: products,
			})
		}()
	}
	return nil, nil
}

package event

import (
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_pay_finish_member_dine_in_order_event_handler sync.Once

// PayFinishMemberDineInOrderEventHandler 会员端堂食订单支付完成事件处理器
func PayFinishMemberDineInOrderEventHandler() {
	once_pay_finish_member_dine_in_order_event_handler.Do(func() {
		// 生成 h5_order 记录，用于在收银机接单列表中显示
		event.NewSystemBus().SubscribePayFinishMemberDineInOrderEvent(createH5OrderForMemberDineIn)
	})
}

// createH5OrderForMemberDineIn 创建 h5_order 记录
func createH5OrderForMemberDineIn(payload event.PayFinishMemberDineInOrderPayload) {
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(payload.CompanyUuid)

	// 获取销售账单信息（包含商品列表）
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
	if err != nil {
		logger.Logger.Error("createH5OrderForMemberDineIn, GetSaleBillAllInfo failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Error(err))
		return
	}

	// 获取销售订单
	var saleOrder *model.SaleOrder
	for _, order := range saleBill.SaleOrders {
		if order.Uuid == payload.SaleOrderUuid {
			saleOrder = order
			break
		}
	}
	if saleOrder == nil {
		logger.Logger.Error("createH5OrderForMemberDineIn, SaleOrder not found",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleOrderUuid", payload.SaleOrderUuid))
		return
	}

	// 获取销售订单商品
	saleOrderProducts := saleOrder.SaleOrderProducts
	if len(saleOrderProducts) == 0 {
		logger.Logger.Info("createH5OrderForMemberDineIn, no products",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid))
		return
	}

	// 获取公司设置（优先从 ctx 获取）
	companySetting := payload.Ctx.GetCompanySetting()
	if companySetting.Uuid == 0 {
		settingSrv := setting.NewSrv(dbm, cache.Global)
		cs, err := settingSrv.GetCompanySetting(payload.Ctx)
		if err != nil {
			logger.Logger.Error("createH5OrderForMemberDineIn, GetCompanySetting failed",
				zap.Uint64("company_uuid", payload.CompanyUuid),
				zap.Uint64("saleBillUuid", payload.SaleBillUuid),
				zap.Error(err))
			return
		}
		companySetting = cs
		payload.Ctx.SetCompanySetting(companySetting)
	}

	// 创建 h5_order 记录
	h5OrderUuid, _ := utils.GetID()
	now := time.Now().Unix()

	// 构建 h5_order_product 列表（仅包含主商品，套餐子商品随主商品一起处理）
	lang := payload.Ctx.GetLanguage()
	h5OrderProductList := make([]*model.H5OrderProduct, 0, len(saleOrderProducts))
	for _, product := range saleOrderProducts {
		// 跳过套餐子商品
		if product.PackageUuid > 0 {
			continue
		}
		h5OrderProductList = append(h5OrderProductList, &model.H5OrderProduct{
			// 快照信息
			Name:          product.MultiLanguageName.GetNameByLang(lang),
			Price:         product.GetFinalSalePrice(),
			SalePrice:     product.SalePrice,
			Num:           product.Num,
			AttributeText: product.GetAttributeNamesByLang(lang),
			Remark:        product.Remark,
			// 关联uuid
			SaleOrderProductUuid: product.Uuid,
			H5OrderUuid:          h5OrderUuid,
			SaleBillUuid:         saleBill.Uuid,
		})
	}

	h5Order := &model.H5Order{
		BaseModel: model.BaseModel{
			Uuid:       h5OrderUuid,
			CreateTime: now,
			UpdateTime: now,
		},
		DeskUuid:        0,                                // 会员端堂食订单无桌台
		SaleOrderUuid:   saleOrder.Uuid,                   // 销售订单uuid
		SaleBillUuid:    saleBill.Uuid,                    // 销售账单uuid
		DeskNo:          saleBill.SerialNo,                // 取餐号（使用 sale_bill.serial_no）
		Status:          constant.H5OrderStatusOrder,      // 状态：待接单
		OrderType:       constant.H5OrderTypeMemberDineIn, // 订单类型：会员端堂食订单
		IsAutoAccept:    0,                                // 非自动接单
		IsNeedAudit:     companySetting.IsOpenH5Order,     // 关闭扫码点餐接单则不需要审核，直接送厨
		OrderTime:       now,                              // 下单时间
		H5OrderProducts: h5OrderProductList,
	}

	// 更新 sale_order_product 的 h5_order_uuid 和 is_accept_order. 更新内存对象，确保后续逻辑使用到的 sale_order_product 是最新的
	for index := range saleOrderProducts {
		product := saleOrderProducts[index]
		product.H5OrderUuid = h5OrderUuid
		product.IsAcceptOrder = constant.OrderProductIsAcceptOrderUnAccept
	}

	// 保存到数据库
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		h5OrderRepo := repository.NewH5OrderRepo(tx)
		saleOrderProductRepo := repository.NewSaleOrderProductRepo(tx)

		// 创建 h5_order
		if _, err := h5OrderRepo.CreateH5Order(*h5Order); err != nil {
			return err
		}

		// 批量创建 h5_order_product
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			if _, err := h5OrderRepo.CreateH5OrderProduct(*h5OrderProduct); err != nil {
				return err
			}
		}

		// 批量更新 sale_order_product 的 h5_order_uuid 和 is_accept_order
		if err := saleOrderProductRepo.Update(map[string]any{
			"h5_order_uuid":   h5OrderUuid,
			"is_accept_order": constant.OrderProductIsAcceptOrderUnAccept,
		},
			repository.CommonRepo.WhereBySaleOrderUuid(saleOrder.Uuid),
			repository.CommonRepo.WhereBySoftDelete(),
		); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logger.Logger.Error("createH5OrderForMemberDineIn, create h5_order failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Error(err))
		return
	}

	logger.Logger.Info("createH5OrderForMemberDineIn, h5_order created",
		zap.Uint64("company_uuid", payload.CompanyUuid),
		zap.Uint64("h5OrderUuid", h5OrderUuid),
		zap.Uint64("saleBillUuid", payload.SaleBillUuid),
		zap.String("deskNo", saleBill.SerialNo))

	// 支付完成后将订单标记为已完成
	markMemberDineInOrderComplete(payload, saleBill)

	// 支付完成后尝试自动接单（如果开启且符合条件）
	// 自动接单会同时完成接单和送厨
	// 注意：必须在 markMemberDineInOrderComplete 之后执行，否则统计不会执行
	autoAcceptMemberDineInOrder(payload)
}

// autoAcceptMemberDineInOrder 自动接单会员端堂食订单
// 使用接单设置（AcceptOrderSetting）判断是否开启自动接单及金额限额
// 注意：必须在 markMemberDineInOrderComplete 之后调用，内部会重新查询最新的 saleBill
func autoAcceptMemberDineInOrder(payload event.PayFinishMemberDineInOrderPayload) {
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(payload.CompanyUuid)

	// 重新从数据库获取最新的 saleBill（markMemberDineInOrderComplete 已更新数据库）
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
	if err != nil {
		logger.Logger.Error("autoAcceptMemberDineInOrder, GetSaleBillAllInfo failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
			zap.Error(err))
		return
	}

	settingSrv := setting.NewSrv(dbm, cache.Global)

	// 确保上下文中有 CompanySetting（支付回调路由无 JWT 中间件，ctx 中可能缺失）
	companySetting := payload.Ctx.GetCompanySetting()
	if companySetting.Uuid == 0 {
		cs, err := settingSrv.GetCompanySetting(payload.Ctx)
		if err != nil {
			logger.Logger.Error("autoAcceptMemberDineInOrder, GetCompanySetting failed",
				zap.Uint64("company_uuid", payload.CompanyUuid),
				zap.Uint64("saleBillUuid", payload.SaleBillUuid),
				zap.Error(err))
			return
		}
		companySetting = cs
		payload.Ctx.SetCompanySetting(companySetting)
	}

	// 关闭h5接单功能时，直接自动接单，无需检查接单设置
	if !companySetting.GetIsOpenH5Order() {
		logger.Logger.Info("autoAcceptMemberDineInOrder, h5 order disabled, auto accept directly",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid))
	} else {
		// 开启h5接单功能时，根据接单设置（金额限额）判断是否自动接单
		acceptOrderSetting, err := settingSrv.GetAcceptOrderSetting(payload.Ctx)
		if err != nil {
			logger.Logger.Error("autoAcceptMemberDineInOrder, GetAcceptOrderSetting failed",
				zap.Uint64("company_uuid", payload.CompanyUuid),
				zap.Uint64("saleBillUuid", payload.SaleBillUuid),
				zap.Error(err))
			return
		}

		// 获取未接单商品的总金额
		saleOrder := saleBill.GetFirstSaleOrder()
		if saleOrder == nil {
			return
		}
		totalPrice := saleBill.GetUnAcceptH5OrderProductTotalPrice(saleOrder.SaleOrderProducts)

		// 判断是否可以自动接单（是否开启 + 金额是否在限额内）
		if !acceptOrderSetting.CanAutoOrder(totalPrice) {
			logger.Logger.Info("autoAcceptMemberDineInOrder, auto accept not allowed",
				zap.Uint64("company_uuid", payload.CompanyUuid),
				zap.Uint64("saleBillUuid", payload.SaleBillUuid),
				zap.Float64("totalPrice", totalPrice))
			return
		}
	}

	// 获取 H5 订单 UUID（由 createH5OrderForMemberDineIn 创建）
	h5OrderRepo := repository.NewH5OrderRepo(db)
	h5Order, err := h5OrderRepo.GetH5Order(
		h5OrderRepo.WhereSaleBillUuid(payload.SaleBillUuid),
		h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
	)
	if err != nil {
		logger.Logger.Error("autoAcceptMemberDineInOrder, GetH5OrderInfo failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
			zap.Error(err))
		return
	}

	// 检查 H5 订单状态，只有待接单状态才能自动接单
	if h5Order.Status != constant.H5OrderStatusOrder {
		logger.Logger.Info("autoAcceptMemberDineInOrder, h5 order status is not pending",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("h5OrderUuid", h5Order.Uuid),
			zap.Uint("status", h5Order.Status))
		return
	}

	// 初始化订单服务
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache.Global)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 设置上下文来源为会员端
	payload.Ctx.SetSource(constant.SourceMember)

	// 确保上下文中有 Company（支付回调路由无 JWT 中间件，ctx 中可能缺失）
	if payload.Ctx.GetCompany().Uuid == 0 {
		company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(payload.CompanyUuid)
		if err != nil {
			logger.Logger.Error("autoAcceptMemberDineInOrder, GetCompanyInfoByUuid failed",
				zap.Uint64("company_uuid", payload.CompanyUuid),
				zap.Uint64("saleBillUuid", payload.SaleBillUuid),
				zap.Error(err))
			return
		}
		payload.Ctx.SetCompany(*company)
		if company.CompanySetting != nil {
			payload.Ctx.SetCompanySetting(*company.CompanySetting)
		}
	}

	// 执行自动接单（同时完成接单和送厨）
	result, err := orderSrv.AcceptH5Order(payload.Ctx, h5Order.Uuid, true)
	if err != nil {
		logger.Logger.Error("autoAcceptMemberDineInOrder, AcceptH5Order failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("h5OrderUuid", h5Order.Uuid),
			zap.Error(err))
		return
	}
	if result != nil {
		logger.Logger.Warn("autoAcceptMemberDineInOrder, AcceptH5Order check failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("h5OrderUuid", h5Order.Uuid),
			zap.Any("result", result))
		return
	}

	logger.Logger.Info("autoAcceptMemberDineInOrder, auto accept success",
		zap.Uint64("company_uuid", payload.CompanyUuid),
		zap.Uint64("saleBillUuid", payload.SaleBillUuid),
		zap.Uint64("h5OrderUuid", h5Order.Uuid))
}

// markMemberDineInOrderComplete 支付完成后将订单标记为已完成
// 会员端堂食订单支付完成后，直接将 SaleBill 和 SaleOrder 状态设置为已完成
func markMemberDineInOrderComplete(payload event.PayFinishMemberDineInOrderPayload, saleBill *model.SaleBill) {
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(payload.CompanyUuid)

	// 检查订单状态，避免重复处理
	if saleBill.Status == constant.SaleBillStatusComplete {
		logger.Logger.Info("markMemberDineInOrderComplete, order already completed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid))
		return
	}

	// 初始化设置服务，用于获取货币单位和积分设置
	settingSrv := setting.NewSrv(dbm, cache.Global)

	// 获取货币单位设置
	currencySetting, err := settingSrv.GetCurrencySetting(payload.Ctx)
	if err != nil {
		logger.Logger.Error("markMemberDineInOrderComplete, GetCurrencySetting failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
			zap.Error(err))
		return
	}

	// 获取积分设置
	pointsSetting, err := settingSrv.GetPointsSetting(payload.Ctx)
	if err != nil {
		logger.Logger.Error("markMemberDineInOrderComplete, GetPointsSetting failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
			zap.Error(err))
		return
	}

	// 更新 SaleOrder 状态为已完成，并记录结账完成后的字段
	for index := range saleBill.SaleOrders {
		saleOrder := saleBill.SaleOrders[index]
		// 计算手续费（各付款单手续费之和）
		commissionFee := saleOrder.CalcCommissionFee()

		// 计算总付款金额（各付款单实收金额之和）
		totalPay := float64(0)
		for _, paymentOrder := range saleOrder.PaymentOrders {
			if paymentOrder.IsDelete() {
				continue
			}
			totalPay = decimal.NewFromFloat(totalPay).Add(decimal.NewFromFloat(paymentOrder.Amount)).InexactFloat64()
		}

		// 计算抹零金额（只有没有手续费时才能抹零）
		if commissionFee == 0 {
			saleOrder.SetCheckOutZeroFee()
		}

		// 最终应收 = 应收金额 + 手续费 - 结账抹零金额
		finalAmount := decimal.NewFromFloat(saleOrder.GetAmountValue()).
			Add(decimal.NewFromFloat(commissionFee)).
			Sub(decimal.NewFromFloat(saleOrder.ZeroCheckoutFee)).InexactFloat64()

		// 计算找零金额
		changeAmount := float64(0)
		if totalPay > finalAmount {
			changeAmount = decimal.NewFromFloat(totalPay).Sub(decimal.NewFromFloat(finalAmount)).InexactFloat64()
		}

		// 实际支付金额 = 总付款金额 - 找零金额
		paymentAmount := totalPay
		if changeAmount > 0 {
			paymentAmount = decimal.NewFromFloat(totalPay).Sub(decimal.NewFromFloat(changeAmount)).InexactFloat64()
		}

		// 构建结算金额并设置订单完成状态
		final := model.FinalAmount{
			CouponAmount:         saleOrder.CalcCouponExchangeAmount(),
			ActivityAmount:       saleOrder.ActivityAmount,
			PaymentAmount:        paymentAmount,
			ChangeAmount:         changeAmount,
			ZeroCheckoutFee:      saleOrder.CalcCheckOutZeroFee(),
			FinalPrice:           finalAmount,
			PaymentCommissionFee: commissionFee,
			GiftAmount:           saleOrder.CalcGiftAmount(saleOrder.SaleOrderProducts),
			Unit:                 currencySetting.Unit,
		}
		saleOrder.SetFinishStatus(final)

		// 设置满减活动信息（FullReductionActivityMessage 已在下单时写入 saleOrder，此处无需重新计算）

		// 计算积分赠送（如果订单有会员）
		if saleOrder.ConsumerUuid != 0 && saleOrder.Member != nil {
			pointsRule := pointsSetting.GetPointsGiftRule(saleBill.IsBuffetSaleBill(), saleOrder.Member.MemberLevelUuid)
			saleOrder.SetGiftPointsRate(int(saleBill.MealNum), pointsRule)

			// 记录会员余额和会员等级名称
			saleOrder.SetMemberBalance()
		}
	}

	// 完成销售账单（参考收银机 FinishSaleBill 流程）
	if saleBill.CanFinishSaleBill() {
		if saleBill.DutyNo == "" {
			// 按优先级查找可用班次（自动接单在 markComplete 之后执行，此时 duty_no 尚未设置）
			// 1. 主收银机的班次 2. 最先登录的收银机班次 3. 留空等下次登录分配
			shiftLogRepo := repository.NewShiftLogRepo(db)
			dutyNo, cashierUuid, cashierName, err := shiftLogRepo.GetAvailableShiftInfo()
			if err != nil {
				logger.Logger.Warn("markMemberDineInOrderComplete, GetAvailableShiftInfo failed",
					zap.Uint64("company_uuid", payload.CompanyUuid),
					zap.Uint64("saleBillUuid", payload.SaleBillUuid),
					zap.Error(err))
			}
			saleBill.SetFinishSaleBill(dutyNo, cashierUuid, cashierName)
		} else {
			// duty_no 已由 AcceptH5Order 自动接单设置，仅标记完成状态
			saleBill.Status = constant.SaleBillStatusComplete
			saleBill.FinishTime = time.Now().Unix()
		}
		// 内存中将sale_order_product改为已接单,让会员端堂食订单的价格计算正确
		for _, saleOrder := range saleBill.SaleOrders {
			for index := range saleOrder.SaleOrderProducts {
				product := saleOrder.SaleOrderProducts[index]
				product.IsAcceptOrder = constant.OrderProductIsAcceptOrderAccepted
			}
		}
		saleBill.CalcAll()
	}

	// 保存到数据库
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		saleBillRepo := repository.NewSaleBillRepo(tx)
		saleOrderRepo := repository.NewSaleOrderRepo(tx)

		// 更新 SaleBill
		if err := saleBillRepo.UpdateSaleBillRecord(*saleBill); err != nil {
			return err
		}

		// 更新 SaleOrder
		for _, saleOrder := range saleBill.SaleOrders {
			if err := saleOrderRepo.UpdateSaleOrderRecord(*saleOrder); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		logger.Logger.Error("markMemberDineInOrderComplete, update order status failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid),
			zap.Error(err))
		return
	}

	logger.Logger.Info("markMemberDineInOrderComplete, order marked as completed",
		zap.Uint64("company_uuid", payload.CompanyUuid),
		zap.Uint64("saleBillUuid", payload.SaleBillUuid),
		zap.Uint64("saleOrderUuid", payload.SaleOrderUuid))
}

// cookingAndFinishMemberDineInOrder 执行送厨和结账逻辑（备用，当前未使用）
func cookingAndFinishMemberDineInOrder(payload event.PayFinishMemberDineInOrderPayload) {
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(payload.CompanyUuid)

	// 初始化订单服务
	settingSrv := setting.NewSrv(dbm, cache.Global)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache.Global)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(payload.SaleBillUuid)
	if err != nil {
		logger.Logger.Error("cookingAndFinishMemberDineInOrder, GetSaleBillAllInfo failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Error(err))
		return
	}

	// 检查是否有未送厨的商品
	unCookingProducts := saleBill.GetSaleOrderProductUnCooking()
	if len(unCookingProducts) == 0 {
		logger.Logger.Info("cookingAndFinishMemberDineInOrder, no uncooking product",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Uint64("saleBillUuid", payload.SaleBillUuid))
		return
	}

	// 设置上下文来源为 member
	payload.Ctx.SetSource(constant.SourceMember)

	// 执行送厨
	_, checkRes, err := orderSrv.InstantOrderCartProductCooking(payload.Ctx, req.OrderCartProductCookingReq{
		SaleBillUuid: payload.SaleBillUuid,
		IgnoreMust:   true, // 会员端堂食订单忽略必点方案检查
	})
	if err != nil {
		logger.Logger.Error("cookingAndFinishMemberDineInOrder, InstantOrderCartProductCooking failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Error(err))
		return
	}
	if checkRes != nil {
		logger.Logger.Warn("cookingAndFinishMemberDineInOrder, InstantOrderCartProductCooking check failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Any("checkRes", checkRes))
		return
	}

	logger.Logger.Info("cookingAndFinishMemberDineInOrder, InstantOrderCartProductCooking success",
		zap.Uint64("company_uuid", payload.CompanyUuid),
		zap.Uint64("saleBillUuid", payload.SaleBillUuid))

	// 完成订单（结账）
	_, err = orderSrv.InstantOrderPaymentFinish(payload.Ctx, req.InstantOrderPaymentFinishReq{
		SaleBillUuid:  payload.SaleBillUuid,
		SaleOrderUuid: payload.SaleOrderUuid,
	})
	if err != nil {
		logger.Logger.Error("cookingAndFinishMemberDineInOrder, InstantOrderPaymentFinish failed",
			zap.Uint64("company_uuid", payload.CompanyUuid),
			zap.Any("payload", utils.ToJson(payload)),
			zap.Error(err))
		return
	}

	logger.Logger.Info("cookingAndFinishMemberDineInOrder, order completed",
		zap.Uint64("company_uuid", payload.CompanyUuid),
		zap.Uint64("saleBillUuid", payload.SaleBillUuid))
}

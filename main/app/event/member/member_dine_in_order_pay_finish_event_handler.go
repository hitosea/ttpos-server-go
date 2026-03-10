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

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var once_pay_finish_member_dine_in_order_event_handler sync.Once

// PayFinishMemberDineInOrderEventHandler 会员端堂食订单支付完成事件处理器
func PayFinishMemberDineInOrderEventHandler() {
	once_pay_finish_member_dine_in_order_event_handler.Do(func() {
		// 生成 h5_order 记录，用于在收银机接单列表中显示
		event.NewSystemBus().SubscribePayFinishMemberDineInOrderEvent(createH5OrderForMemberDineIn)

		// 执行送厨和结账逻辑
		// event.NewSystemBus().SubscribePayFinishMemberDineInOrderEvent(cookingAndFinishMemberDineInOrder)
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

	// 创建 h5_order 记录
	h5OrderUuid, _ := utils.GetID()
	now := time.Now().Unix()

	// 构建 h5_order_product 列表（仅包含主商品，套餐子商品随主商品一起处理）
	h5OrderProductList := make([]*model.H5OrderProduct, 0, len(saleOrderProducts))
	for _, product := range saleOrderProducts {
		// 跳过套餐子商品
		if product.PackageUuid > 0 {
			continue
		}
		h5OrderProductList = append(h5OrderProductList, &model.H5OrderProduct{
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
		IsNeedAudit:     1,                                // 需要审核
		OrderTime:       now,                              // 下单时间
		H5OrderProducts: h5OrderProductList,
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

		// 更新 sale_order_product 的 h5_order_uuid
		for _, product := range saleOrderProducts {
			product.H5OrderUuid = h5OrderUuid
			if err := saleOrderProductRepo.UpdateSaleOrderProductRecord(*product); err != nil {
				return err
			}
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
}

// cookingAndFinishMemberDineInOrder 执行送厨和结账逻辑
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

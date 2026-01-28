package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/errors"
	coreModel "ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/domain/event"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/repository"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/lineman"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ITakeoutOrderSrv 外卖订单服务接口
type ITakeoutOrderSrv interface {
	// 订单查询
	GetList(ctx context.Context, req *request.TakeoutOrderListReq) (*response.TakeoutOrderListResp, error)
	GetByUuid(ctx context.Context, uuid uint64) (*response.TakeoutOrderResp, error)
	// 创建或更新订单（接受已转换的订单对象，商品数据从 order.RawData 中解析）- 由 Application 层调用
	CreateOrder(ctx context.Context, order *takeoutModel.TakeoutOrder) error
	// 更新订单状态 - 由 Application 层调用
	UpdateOrderStatus(ctx context.Context, orderUuid string, newStatus string, message string) error
	// 订单操作
	AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error
	RejectOrder(ctx context.Context, req *request.TakeoutOrderRejectReq) error
	CallRider(ctx context.Context, req *request.TakeoutOrderCallRiderReq) error
	CheckOrderCancelable(ctx context.Context, req *request.TakeoutOrderCheckCancelableReq) (*response.TakeoutOrderCancelCheckResp, error)
	CancelOrder(ctx context.Context, req *request.TakeoutOrderCancelReq) error
	// GetOrderForPrint 获取打印所需的订单数据
	GetOrderForPrint(ctx context.Context, orderUuid uint64) (*takeoutModel.TakeoutOrder, error)
	// IncrementalUpdateOrder 增量更新订单（基于变动结果进行增删改）
	// existingOrder: 数据库中的现有订单
	// updatedOrder: 平台推送的更新订单
	// changeResult: 变动检测结果
	IncrementalUpdateOrder(ctx context.Context, existingOrder, updatedOrder *takeoutModel.TakeoutOrder, changeResult *valueobject.OrderChangeResult) error
	// EnrichOrderItems 为订单商品填充 TTPOS 商品映射信息
	// 解析平台商品ID，查询TTPOS商品信息（名称、分类、ERP编码等），填充到订单商品中
	// 返回: abnormalItemUuids (商品映射异常的商品UUID列表)
	EnrichOrderItems(ctx context.Context, order *takeoutModel.TakeoutOrder) (abnormalItemUuids []uint64, err error)
	// BuildBomQuantityMap 构建订单商品的 BOM 数量映射（用于库存检查和出库）
	// 返回: bomQuantityMap (BOM UUID -> 数量), bomItemMap (BOM UUID -> 订单商品，包含 modifiers), error
	BuildBomQuantityMap(ctx context.Context, order *takeoutModel.TakeoutOrder) (map[uint64]int, map[uint64]*takeoutModel.TakeoutOrderItem, error)
	// CalculateTakeoutOrderSalesVolume 计算外卖订单销量
	// 返回: productBoms (BOM UUID -> 销量), productPackages (Package UUID -> 销量), error
	CalculateTakeoutOrderSalesVolume(order *takeoutModel.TakeoutOrder) (map[uint64]float64, map[uint64]float64, error)
	// BatchAssignShiftLogToPendingOrders 批量将待分配班次的订单分配给指定班次
	// 对于已接单的订单，会同步生成 ERP 发票
	BatchAssignShiftLogToPendingOrders(ctx context.Context, shiftLogUuid, staffUuid uint64) error
}

// takeoutOrderSrv 外卖订单服务实现
type takeoutOrderSrv struct {
	dbm            *database.DBManager
	menuRepo       menuRepo.IMenuDataRepository
	erpSyncService ITakeoutErpSyncService
}

// NewTakeoutOrderSrv 创建外卖订单服务
func NewTakeoutOrderSrv(dbm *database.DBManager) ITakeoutOrderSrv {
	return &takeoutOrderSrv{
		dbm:            dbm,
		menuRepo:       persistence.NewMenuDataRepository(dbm),
		erpSyncService: NewTakeoutErpSyncService(),
	}
}

// GetList 获取订单列表
func (s *takeoutOrderSrv) GetList(ctx context.Context, req *request.TakeoutOrderListReq) (*response.TakeoutOrderListResp, error) {
	db := ctx.GetDB()

	// 创建 Repository
	orderRepo := persistence.NewTakeoutOrderRepo(db)

	// 构建查询选项
	options := []persistence.DBOption{
		orderRepo.WherePlatform(req.Platform),
		orderRepo.WhereOrderState(req.Status),
		orderRepo.WhereTimeRange(req.StartTime, req.EndTime),
		orderRepo.WhereSearch(req.Search),
		orderRepo.WhereIsHistoryOrder(req.IsHistory),
		orderRepo.WithTakeoutOrderUpdateLogs(orderRepo.Select("uuid", "create_time", "takeout_order_uuid")),
		orderRepo.Limit(req.PageSize),
		orderRepo.Offset((req.PageNo - 1) * req.PageSize),
	}

	// 查询订单列表
	orders, total, err := orderRepo.GetList(options...)
	if err != nil {
		logger.Logger.Error("查询订单列表失败", zap.Error(err), zap.Any("options", options))
		return nil, errors.WithMessage(errors.New("查询订单列表失败"), err.Error())
	}

	// 构建响应
	list := make([]*response.TakeoutOrderListItemResp, 0, len(orders))
	for _, order := range orders {
		orderResp := &response.TakeoutOrderListItemResp{
			Uuid:             order.Uuid,
			Platform:         order.Platform,
			ShortOrderNumber: order.ShortOrderNumber,
			OrderState:       order.OrderState,
			IsAbnormal:       order.IsAbnormal,
			Subtotal:         order.PlatformTotal, // 列表返回 platform_total 字段
			TotalItems:       len(order.TakeoutOrderItems),
			UpdateLogs: func() []response.TakeoutOrderUpdateLogResp {
				updateLogs := make([]response.TakeoutOrderUpdateLogResp, 0, len(order.TakeoutOrderUpdateLogs))
				for _, updateLog := range order.TakeoutOrderUpdateLogs {
					updateLogs = append(updateLogs, response.TakeoutOrderUpdateLogResp{
						Uuid:       updateLog.Uuid,
						CreateTime: updateLog.CreateTime,
					})
				}
				return updateLogs
			}(),
		}
		list = append(list, orderResp)
	}

	return &response.TakeoutOrderListResp{
		List: list,
		Meta: response.Meta{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    int(total),
		},
	}, nil
}

// GetByUuid 根据UUID获取订单详情
func (s *takeoutOrderSrv) GetByUuid(ctx context.Context, uuid uint64) (*response.TakeoutOrderResp, error) {
	db := ctx.GetDB()

	// 创建 Repository
	orderRepo := persistence.NewTakeoutOrderRepo(db)

	// 查询订单
	order, err := orderRepo.GetByUuid(
		uuid,
		orderRepo.WithPreload(persistence.DBOption(func(db *gorm.DB) *gorm.DB {
			return db.Preload("TakeoutOrderItems", func(db *gorm.DB) *gorm.DB {
				return db.Where("delete_time = ?", 0)
			}).
				Preload("TakeoutOrderItems.TakeoutOrderItemModifiers", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("TakeoutOrderReceiver", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("TakeoutOrderCampaigns", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("TakeoutOrderPromos", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				})
		})),
		orderRepo.WithTakeoutOrderUpdateLogs(orderRepo.Select("uuid", "create_time", "takeout_order_uuid")),
	)
	if err != nil {
		logger.Logger.Error("查询订单失败",
			zap.Error(err),
			zap.Uint64("uuid", uuid))
		return nil, errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	// 构建商品列表
	itemList := make([]response.TakeoutOrderItemResp, 0, len(order.TakeoutOrderItems))
	for _, item := range order.TakeoutOrderItems {
		itemList = append(itemList, response.TakeoutOrderItemResp{
			Uuid:           item.Uuid,            // 商品UUID
			Specifications: item.Specifications,  // 规格说明
			Quantity:       item.Quantity,        // 数量
			Price:          item.GetTotalPrice(), // 价格
			Tax:            item.Tax,             // 税费
			ItemName:       *language.JsonToLocaleResponse(item.ItemName),
			IsPackage:      item.IsPackage(),
			SubItems: func() []response.TakeoutOrderItemResp {
				subItems := make([]response.TakeoutOrderItemResp, 0, len(item.TakeoutOrderItemModifiers))
				if !item.IsPackage() {
					return subItems
				}
				for _, modifier := range item.TakeoutOrderItemModifiers {
					if modifier.IsCommodity() {
						subItems = append(subItems, response.TakeoutOrderItemResp{
							Uuid:     modifier.Uuid,
							ItemName: *language.JsonToLocaleResponse(modifier.ModifierName),
							Quantity: modifier.Quantity,
							Price:    modifier.Price, // 价格
							Tax:      modifier.Tax,   // 税费
						})
					}
				}
				return subItems
			}(),
			Modifiers: func() string {
				if item.IsPackage() {
					return ""
				}
				modifierNames := make([]string, 0, len(item.TakeoutOrderItemModifiers))
				for _, modifier := range item.TakeoutOrderItemModifiers {
					modifierName := *language.JsonToLocaleResponse(modifier.ModifierName)
					modifierNames = append(modifierNames, modifierName.GetLocale(ctx.GetLanguage()))
				}
				return strings.Join(modifierNames, ",")
			}(),
		})
	}

	// 构建收货人信息
	var receiverResp response.TakeoutOrderReceiverResp
	if order.TakeoutOrderReceiver != nil {
		receiverResp = response.TakeoutOrderReceiverResp{
			ReceiverName:        order.TakeoutOrderReceiver.ReceiverName,
			ReceiverPhones:      order.TakeoutOrderReceiver.ReceiverPhones,
			UnitNumber:          order.TakeoutOrderReceiver.UnitNumber,
			DeliveryInstruction: order.TakeoutOrderReceiver.DeliveryInstruction,
			Address:             order.TakeoutOrderReceiver.Address,
		}
	}

	// 构建活动信息
	campaigns := make([]response.TakeoutOrderCampaignResp, 0)
	for _, campaign := range order.TakeoutOrderCampaigns {
		campaigns = append(campaigns, response.TakeoutOrderCampaignResp{
			Uuid:           campaign.Uuid,
			CampaignName:   campaign.GetCampaignName(),
			CampaignType:   campaign.CampaignType,
			DeductedAmount: campaign.DeductedAmount,
		})
	}

	// 构建促销信息
	promos := make([]response.TakeoutOrderPromoResp, 0, len(order.TakeoutOrderPromos))
	for _, promo := range order.TakeoutOrderPromos {
		promos = append(promos, response.TakeoutOrderPromoResp{
			Uuid:             promo.Uuid,
			PromoCode:        promo.PromoCode,
			PromoName:        promo.PromoName,
			PromoDescription: promo.PromoDescription,
			PromoAmount:      promo.PromoAmount,
			MexFundedRatio:   promo.MexFundedRatio,
		})
	}

	// 构建价格信息
	priceResp := response.TakeoutOrderPriceResp{
		Subtotal:          order.Subtotal,
		DeliveryFee:       order.DeliveryFee,
		EaterPayment:      order.EaterPayment,
		PlatformDiscount:  order.PlatformDiscount,
		MerchantDiscount:  order.MerchantDiscount,
		BasketPromo:       order.BasketPromo,
		Tax:               order.Tax,
		SmallOrderFee:     order.SmallOrderFee,
		MerchantChargeFee: order.MerchantChargeFee,
		PlatformTotal:     order.PlatformTotal,
	}

	// 版本兼容处理：	 之前的版本，eater_payment 返回 platform_total 的值
	if ctx.Version(context.LT, "2.16.0") {
		priceResp.EaterPayment = order.PlatformTotal
	}

	return &response.TakeoutOrderResp{
		Uuid:             order.Uuid,
		Platform:         order.Platform,
		ShortOrderNumber: order.ShortOrderNumber,
		OrderState:       order.OrderState,
		RiderStatus:      order.GetRiderStatus(),
		IsAbnormal:       order.IsAbnormal,
		AbnormalDetail:   order.AbnormalDetail,
		OrderTimes: response.TakeoutOrderTimesResp{
			OrderTime:          order.OrderTime,
			SubmitTime:         order.SubmitTime,
			AcceptedTime:       order.AcceptedTime,
			CompletedTime:      order.CompletedTime,
			EstimatedReadyTime: order.EstimatedReadyTime,
			MaxReadyTime:       order.MaxReadyTime,
		},
		Price: priceResp,
		Currencies: response.TakeoutOrderCurrencyResp{
			CurrencyCode:     order.CurrencyCode,
			CurrencySymbol:   order.CurrencySymbol,
			CurrencyExponent: order.CurrencyExponent,
		},
		PaymentType: order.PaymentType,
		// 订单优惠
		Discounts: response.TakeoutOrderDiscountsResp{
			PlatformDiscount: order.PlatformDiscount,
			MerchantDiscount: order.MerchantDiscount,
		},
		// 订单类型
		Cutlery:              order.Cutlery,
		AdditionalProperties: order.GetAdditionalProperties(ctx.GetLanguage()),
		OrderType:            order.OrderType,
		TotalItems:           len(itemList),
		Items:                itemList,
		Receiver:             receiverResp,
		Campaigns:            campaigns,
		Promos:               promos,
		UpdateLogs: func() []response.TakeoutOrderUpdateLogResp {
			updateLogs := make([]response.TakeoutOrderUpdateLogResp, 0, len(order.TakeoutOrderUpdateLogs))
			for _, updateLog := range order.TakeoutOrderUpdateLogs {
				updateLogs = append(updateLogs, response.TakeoutOrderUpdateLogResp{
					Uuid:       updateLog.Uuid,
					CreateTime: updateLog.CreateTime,
				})
			}
			return updateLogs
		}(),
	}, nil
}

// AcceptOrder 接单
func (s *takeoutOrderSrv) AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	acceptedTime := time.Now().Unix()
	userUuid := ctx.GetStaffUuid()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	// 检查订单状态
	if err := order.IsPendingOrder(); err != nil {
		return err
	}

	// 调用 BMP RPC 通知平台接受订单
	if !order.IsAutoAcceptOrder() {
		rpcClient, err := rpc.NewBMPTakeoutClient()
		if err != nil {
			logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
		}
		defer rpcClient.Close()
		// 调用 PrepareOrder 接口（接受订单）
		if err := rpcClient.PrepareOrder(ctx.GetContext(), order.TakeoutOrderUuid, "Accepted"); err != nil {
			// 检查是否是订单状态不允许更新的错误
			logger.Logger.Error("调用 BMP PrepareOrder 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			if !strings.Contains(err.Error(), "order status can't be updated, because order state isn't NEW") {
				return errors.WithMessage(errors.New("通知平台接受订单失败"), err.Error())
			}
		}
	}

	// 设置员工班次信息】
	if !order.IsExistShiftLog() {
		if err := orderRepo.SetStaffShiftLogUuid(order); err != nil {
			logger.Logger.Error("设置员工班次日志UUID失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		} else {
			order.AcceptedBy = userUuid
		}
	}

	// 更新订单状态
	if err := db.Transaction(func(tx *gorm.DB) error {
		if order.AcceptedTime > 0 {
			acceptedTime = order.AcceptedTime
		}
		updateData := map[string]interface{}{
			"order_state":          valueobject.TakeoutOrderStateAccepted,
			"accepted_time":        acceptedTime,
			"staff_shift_log_uuid": order.StaffShiftLogUuid,
			"accepted_by":          order.AcceptedBy,
			"update_time":          currentTime,
		}
		if err := persistence.NewTakeoutOrderRepo(tx).UpdateByMap(order.Uuid, updateData); err != nil {
			logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 同步到 ERP
	if order.IsExistShiftLog() {
		if err := s.erpSyncService.SyncOrderToERP(ctx, order.Uuid); err != nil {
			logger.Logger.Error("同步 Grab 订单到 ERP 失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("同步 Grab 订单到 ERP 失败"), err.Error())
		}
	}

	// 发布订单接受事件
	event.GetDispatcher().Publish(event.NewOrderAcceptedEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		userUuid,
		valueobject.TakeoutOrderAcceptedTypeManual,
		ctx.GetCompanyUuid(),
	))
	return nil
}

// RejectOrder 拒单
func (s *takeoutOrderSrv) RejectOrder(ctx context.Context, req *request.TakeoutOrderRejectReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	userUuid := ctx.GetStaffUuid()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单状态
	if order.OrderState != valueobject.TakeoutOrderStatePending {
		return errors.New("订单状态不正确")
	}

	// 调用 BMP RPC 通知平台拒绝订单
	rpcClient, err := rpc.NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
	}
	defer rpcClient.Close()

	// 调用 PrepareOrder 接口（拒绝订单）
	if err := rpcClient.PrepareOrder(ctx.GetContext(), order.TakeoutOrderUuid, "Rejected"); err != nil {
		logger.Logger.Error("调用 BMP PrepareOrder 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("通知平台拒绝订单失败"), err.Error())
	}

	// 更新订单状态
	updateData := map[string]interface{}{
		"order_state":        valueobject.TakeoutOrderStateRejected,
		"rejected_time":      currentTime,
		"rejected_by":        userUuid,
		"reject_reason_code": req.RejectReasonCode,
		"reject_reason":      req.RejectReasonCode, // 使用 code 作为原因
		"update_time":        currentTime,
	}

	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
	}

	// 发布订单拒绝事件
	rejectedEvent := event.NewOrderRejectedEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		userUuid,
		req.RejectReasonCode,
		req.RejectReasonCode, // 使用 code 作为原因
		ctx.GetCompanyUuid(),
	)
	event.GetDispatcher().Publish(rejectedEvent)

	return nil
}

// CallRider 呼叫骑手（标记订单准备完成）
func (s *takeoutOrderSrv) CallRider(ctx context.Context, req *request.TakeoutOrderCallRiderReq) error {
	db := ctx.GetDB()
	userUuid := ctx.GetStaffUuid()
	currentTime := time.Now().Unix()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	if order.IsLinemanOrder() {
		return errors.New("LINE MAN 订单不支持呼叫骑手")
	}

	// 检查订单状态 - 只有已接单配餐中的订单才能呼叫骑手
	if order.OrderState != valueobject.TakeoutOrderStateAccepted {
		return errors.New("订单状态不正确，只有已接单配餐中的订单才能呼叫骑手")
	}

	// 调用 BMP RPC 标记订单准备完成
	rpcClient, err := rpc.NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
	}
	defer rpcClient.Close()

	// 调用 MarkOrderReady 接口（标记订单准备完成，通知平台派送骑手）
	if err := rpcClient.MarkOrderReady(ctx.GetContext(), order.TakeoutOrderUuid); err != nil {
		logger.Logger.Error("调用 BMP MarkOrderReady 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("呼叫骑手失败"), err.Error())
	}

	// 更新订单状态为待骑手接单
	updateData := map[string]interface{}{
		"order_state": func() int {
			if order.Platform == valueobject.TakeoutPlatformGrab {
				if order.IsDineInOrder() {
					return valueobject.TakeoutOrderStateCompleted
				}
				return grab.ConvertPlatformStateToOrderState(order.PlatformOrderState, valueobject.TakeoutOrderStateRiderPending)
			}
			return valueobject.TakeoutOrderStateRiderProcessing
		}(),
		"update_time": currentTime,
	}

	// 如果是自动接单的订单，且已设置班次，则同时更新班次信息
	if order.IsAutoAcceptOrder() && order.StaffShiftLogUuid > 0 {
		updateData["staff_shift_log_uuid"] = order.StaffShiftLogUuid
		updateData["accepted_by"] = userUuid
	}

	// 更新订单状态并同步到 ERP
	if err := db.Transaction(func(tx *gorm.DB) error {
		ctxCopy := ctx.Copy()
		ctxCopy.SetDB(tx)
		// 更新订单状态
		if err := persistence.NewTakeoutOrderRepo(tx).UpdateByMap(order.Uuid, updateData); err != nil {
			logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
		}
		// 同步到 ERP
		if order.IsAutoAcceptOrder() {
			if err := s.erpSyncService.SyncOrderToERP(ctxCopy, order.Uuid); err != nil {
				logger.Logger.Error("同步 Grab 订单到 ERP 失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
				return errors.WithMessage(errors.New("同步 Grab 订单到 ERP 失败"), err.Error())
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 发布订单呼叫骑手事件
	event.GetDispatcher().Publish(event.NewOrderReadyEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		ctx.GetCompanyUuid(),
		userUuid,
	))

	return nil
}

// CheckOrderCancelable 检查订单是否可取消
func (s *takeoutOrderSrv) CheckOrderCancelable(ctx context.Context, req *request.TakeoutOrderCheckCancelableReq) (*response.TakeoutOrderCancelCheckResp, error) {
	db := ctx.GetDB()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
		return nil, errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
	}
	if order.IsLinemanOrder() {
		return nil, errors.New("LINE MAN 订单不支持取消")
	}

	// 检查订单状态
	if order.OrderState == valueobject.TakeoutOrderStateCompleted || order.IsDeletedOrCanceled() {
		return &response.TakeoutOrderCancelCheckResp{
			CanCancel:             false,
			NonCancellationReason: i18n.Translate(ctx.GetLanguage(), "已完成或已拒单或已取消的订单不能取消"),
			CancelReasons:         []response.TakeoutOrderCancelReason{},
		}, nil
	}

	// 调用 BMP RPC 检查订单是否可取消
	rpcClient, err := rpc.NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return nil, errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
	}
	defer rpcClient.Close()

	// 调用 CheckOrderCancelable 接口
	canCancel, reason, rawData, err := rpcClient.CheckOrderCancelable(ctx.GetContext(), order.TakeoutOrderUuid)
	if err != nil {
		logger.Logger.Error("调用 BMP CheckOrderCancelable 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return nil, errors.WithMessage(errors.New("检查订单可取消状态失败"), err.Error())
	}

	// 构建响应
	resp := &response.TakeoutOrderCancelCheckResp{
		CanCancel:             canCancel,
		NonCancellationReason: reason,
		CancelReasons:         []response.TakeoutOrderCancelReason{},
	}

	// 如果有 raw_data，解析它
	if rawData != "" {
		var rawDataMap map[string]interface{}
		if err := json.Unmarshal([]byte(rawData), &rawDataMap); err != nil {
			logger.Logger.Warn("解析 raw_data 失败", zap.Error(err), zap.String("rawData", rawData))
		} else {
			// 提取 cancelReasons 列表
			if cancelReasons, ok := rawDataMap["cancelReasons"].([]interface{}); ok {
				for _, reasonItem := range cancelReasons {
					if reasonMap, ok := reasonItem.(map[string]interface{}); ok {
						cancelReason := response.TakeoutOrderCancelReason{}
						if code, ok := reasonMap["code"].(float64); ok {
							cancelReason.Code = strconv.FormatFloat(code, 'f', -1, 64)
						}
						if reasonText, ok := reasonMap["reason"].(string); ok {
							cancelReason.Reason = reasonText
						}
						resp.CancelReasons = append(resp.CancelReasons, cancelReason)
					}
				}
			}
		}
	}

	return resp, nil
}

// CancelOrder 取消订单
func (s *takeoutOrderSrv) CancelOrder(ctx context.Context, req *request.TakeoutOrderCancelReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	userUuid := ctx.GetStaffUuid()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.IsLinemanOrder() {
		return errors.New("LINE MAN 订单不支持取消")
	}

	// 检查订单状态
	if order.OrderState == valueobject.TakeoutOrderStateCompleted || order.IsDeletedOrCanceled() {
		return errors.New("已完成或已拒单的订单不能取消")
	}

	// 先检查订单是否可取消
	checkResp, err := s.CheckOrderCancelable(ctx, &request.TakeoutOrderCheckCancelableReq{Uuid: req.Uuid})
	if err != nil {
		return err
	}
	if !checkResp.CanCancel {
		return errors.New(checkResp.NonCancellationReason)
	}

	// 调用 BMP RPC 取消订单
	if order.IsErpInvoiceSynced() {
		rpcClient, err := rpc.NewBMPTakeoutClient()
		if err != nil {
			logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return err
		}
		defer rpcClient.Close()
		// 调用 CancelOrder 接口（通知平台取消订单）
		if err := rpcClient.CancelOrder(ctx.GetContext(), order.TakeoutOrderUuid, req.ReasonCode); err != nil {
			logger.Logger.Error("调用 BMP CancelOrder 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return err
		}
	}

	// 更新订单状态为已取消
	reasonText := ""
	for _, reason := range checkResp.CancelReasons {
		if reason.Code == req.ReasonCode {
			reasonText = reason.Reason
			break
		}
	}
	updateData := map[string]interface{}{
		"order_state":        valueobject.TakeoutOrderStateCanceled,
		"rejected_by":        userUuid,
		"rejected_time":      currentTime,
		"reject_reason_code": req.ReasonCode,
		"reject_reason":      reasonText, // 取消原因描述
		"update_time":        currentTime,
	}

	// 如果订单的 staff_shift_log_uuid 不存在，尝试自动分配班次
	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
	}

	// 发布订单取消事件
	event.GetDispatcher().Publish(event.NewOrderCancelEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		ctx.GetCompanyUuid(),
		reasonText,
		valueobject.TakeoutOrderStateCanceled,
	))

	return nil
}

// findProductMapping 查找商品映射关系
// 返回: isMapped, ttposProductUuid, ttposProductType, error
func (s *takeoutOrderSrv) findProductMapping(platform, platformItemId string) (int, uint64, int, error) {
	// 解析商品ID，提取 TTPOS UUID
	result, err := valueobject.ParsePlatformID(platform, platformItemId)
	if err != nil {
		// ID格式错误，标记为未映射
		return 0, 0, 0, nil
	}

	if !result.IsMapped {
		return 0, 0, 0, nil
	}

	switch result.IDType {
	case valueobject.IDTypePackage:
		return 1, result.UUID, 1, nil
	default:
		return 1, result.UUID, 0, nil
	}
}

// findModifierMapping 查找修饰符映射关系
// 返回: isMapped, ttposModifierUuid, ttposModifierType, error
//
// ttposModifierType 可能的值：
//   - "flavor": 规格（如大杯、中杯、小杯）
//   - "sauce": 加料（如珍珠、椰果、布丁）
//   - "attr": 属性（如冰度、糖度）
func (s *takeoutOrderSrv) findModifierMapping(platform, platformModifierId string) (isMapped int, ttposModifierUuid uint64, ttposModifierType string, err error) {
	// 使用统一的 ID 解析器
	result, parseErr := valueobject.ParsePlatformID(platform, platformModifierId)
	if parseErr != nil {
		// ID格式错误，标记为未映射
		return 0, 0, "", nil
	}

	if !result.IsMapped {
		return 0, 0, "", nil
	}

	// 根据 ID 类型确定修饰符类型
	var modifierType string
	switch result.IDType {
	case valueobject.IDTypeFlavor:
		modifierType = string(valueobject.ModifierTypeFlavor) // 规格
	case valueobject.IDTypeSauce:
		modifierType = string(valueobject.ModifierTypeSauce) // 加料
	case valueobject.IDTypeAttr:
		modifierType = string(valueobject.ModifierTypeAttr) // 属性
	case valueobject.IDTypePackageItem:
		modifierType = string(valueobject.ModifierTypeCommodity) // 套餐商品
	default:
		// 不是修饰符类型，标记为未映射
		return 0, 0, "", nil
	}

	return 1, result.UUID, modifierType, nil
}

// CreateOrder 创建订单（接受已转换的订单对象）- 由 Application 层调用
func (s *takeoutOrderSrv) CreateOrder(ctx context.Context, order *takeoutModel.TakeoutOrder) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	orderRepo := persistence.NewTakeoutOrderRepo(db)

	// 检查订单是否已存在
	existingOrder, err := orderRepo.GetByPlatformOrderId(order.Platform, order.PlatformOrderId)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.String("platform", order.Platform), zap.String("platformOrderId", order.PlatformOrderId))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if existingOrder != nil && existingOrder.Uuid > 0 {
		// 订单已存在，直接返回
		return nil
	}

	// 验证商品数据
	if len(order.TakeoutOrderItems) == 0 {
		return errors.New("订单商品数据不能为空")
	}

	// 设置员工班次信息】
	if !order.IsExistShiftLog() {
		if err := orderRepo.SetStaffShiftLogUuid(order); err != nil {
			logger.Logger.Error("设置员工班次日志UUID失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		}
	}

	// 开启事务
	if err := db.Transaction(func(tx *gorm.DB) error {
		ctxTx := ctx.Copy()
		ctxTx.SetDB(tx)
		// 创建订单仓储
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)

		// 保存或者更新订单
		if err := orderRepoTx.Create(order); err != nil {
			logger.Logger.Error("创建订单失败", zap.Error(err), zap.Any("order", order))
			return errors.WithMessage(errors.New("创建订单失败"), err.Error())
		}

		// 2. 处理订单商品和修饰符
		hasUnmapped := false
		abnormalProductIds := make([]uint64, 0)
		orderItemRepo := persistence.NewTakeoutOrderItemRepo(tx)
		modifierRepo := persistence.NewTakeoutOrderItemModifierRepo(tx)

		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]

			// 补全商品信息（UUID、映射、名称等）
			isAbnormal, err := s.enrichItemInfo(ctx, order.Platform, order.Uuid, item)
			if err != nil {
				return err
			}
			if isAbnormal {
				abnormalProductIds = append(abnormalProductIds, item.TtposProductPackageUuid)
			}
			if item.IsMapped == 0 {
				hasUnmapped = true
			}

			// 创建订单商品
			if err := orderItemRepo.Create(item); err != nil {
				logger.Logger.Error("创建订单商品失败", zap.Error(err), zap.Any("item", item))
				return errors.WithMessage(errors.New("创建订单商品失败"), err.Error())
			}

			// 补全修饰符信息（UUID、映射、名称等）
			modifierAbnormalUuids, err := s.enrichModifiersInfo(ctx, order.Platform, item)
			if err != nil {
				return err
			}
			abnormalProductIds = append(abnormalProductIds, modifierAbnormalUuids...)

			// 创建修饰符
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]
				if err := modifierRepo.Create(modifier); err != nil {
					logger.Logger.Error("创建商品修饰符失败", zap.Error(err), zap.Any("modifier", modifier))
					return errors.WithMessage(errors.New("创建商品修饰符失败"), err.Error())
				}
			}
		}

		// 3. 保存收货人信息
		if order.TakeoutOrderReceiver != nil {
			receiverRepo := persistence.NewTakeoutOrderReceiverRepo(tx)
			receiverUuid, err := utils.GetID()
			if err != nil {
				logger.Logger.Error("生成收货人UUID失败", zap.Error(err))
				return errors.WithMessage(err, "生成收货人UUID失败")
			}
			order.TakeoutOrderReceiver.Uuid = receiverUuid
			order.TakeoutOrderReceiver.TakeoutOrderUuid = order.Uuid
			order.TakeoutOrderReceiver.CreateTime = currentTime
			order.TakeoutOrderReceiver.UpdateTime = currentTime
			if err := receiverRepo.Create(order.TakeoutOrderReceiver); err != nil {
				logger.Logger.Error("创建收货人信息失败", zap.Error(err))
				return errors.WithMessage(err, "创建收货人信息失败")
			}
		}

		// 4. 保存活动信息
		if len(order.TakeoutOrderCampaigns) > 0 {
			campaignRepo := persistence.NewTakeoutOrderCampaignRepo(tx)
			for i := range order.TakeoutOrderCampaigns {
				campaign := &order.TakeoutOrderCampaigns[i]
				campaignUuid, err := utils.GetID()
				if err != nil {
					logger.Logger.Error("生成活动UUID失败", zap.Error(err))
					return errors.WithMessage(err, "生成活动UUID失败")
				}
				campaign.Uuid = campaignUuid
				campaign.TakeoutOrderUuid = order.Uuid
				campaign.CreateTime = currentTime
				campaign.UpdateTime = currentTime
			}

			// 批量创建活动
			campaignsPtr := make([]*takeoutModel.TakeoutOrderCampaign, len(order.TakeoutOrderCampaigns))
			for i := range order.TakeoutOrderCampaigns {
				campaignsPtr[i] = &order.TakeoutOrderCampaigns[i]
			}
			if err := campaignRepo.BatchCreate(campaignsPtr); err != nil {
				logger.Logger.Error("批量创建活动信息失败", zap.Error(err))
				return errors.WithMessage(err, "批量创建活动信息失败")
			}
		}

		// 5. 检查商品映射状态
		if hasUnmapped || len(abnormalProductIds) > 0 {
			// 标记订单为异常状态
			order.IsAbnormal = 1
			if len(abnormalProductIds) > 0 {
				abnormalProductIdsStr := make([]string, 0, len(abnormalProductIds))
				for _, productId := range abnormalProductIds {
					abnormalProductIdsStr = append(abnormalProductIdsStr, strconv.FormatUint(productId, 10))
				}
				order.AbnormalDetail = "订单包含被删除的商品: " + strings.Join(abnormalProductIdsStr, ",")
			} else {
				order.AbnormalDetail = "订单包含未映射的商品"
			}
			if err := orderRepoTx.Update(order); err != nil {
				logger.Logger.Error("更新订单异常状态失败", zap.Error(err))
				return errors.WithMessage(err, "更新订单异常状态失败")
			}
		}

		// 6. 计算并保存订单原料（提前计算，不等到接单）
		if err := s.extractAndSaveTakeoutOrderMaterials(ctxTx, order); err != nil {
			logger.Logger.Warn("提取并保存订单原料失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(err, "提取并保存订单原料失败")
		}

		return nil
	}); err != nil {
		return err
	}

	// 发布订单创建事件
	event.GetDispatcher().Publish(event.NewOrderCreatedEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		order.EaterPayment,
		ctx.GetCompanyUuid(),
	))

	// 自动接单
	if order.IsAutoAcceptOrder() {
		if err := s.AcceptOrder(ctx, &request.TakeoutOrderAcceptReq{
			Uuid: order.Uuid,
		}); err != nil {
			logger.Logger.Error("接单失败", zap.Error(err))
		}
	}

	return nil
}

// IncrementalUpdateOrder 增量更新订单（基于变动结果进行增删改）
// 相比 CreateOrUpdateOrder 的全量删除重建，此方法只更新变动的部分
func (s *takeoutOrderSrv) IncrementalUpdateOrder(
	ctx context.Context,
	existingOrder, updatedOrder *takeoutModel.TakeoutOrder,
	changeResult *valueobject.OrderChangeResult,
) error {
	db := ctx.GetDB()
	orderUuid := existingOrder.Uuid
	platform := existingOrder.Platform

	// 构建现有商品的映射表（用于复用 UUID 和映射信息）
	existingItemMap := make(map[string]*takeoutModel.TakeoutOrderItem)
	for i := range existingOrder.TakeoutOrderItems {
		item := &existingOrder.TakeoutOrderItems[i]
		existingItemMap[item.PlatformItemId] = item
	}

	// 为 updatedOrder 的商品填充映射信息（从 existingOrder 复制）
	for i := range updatedOrder.TakeoutOrderItems {
		item := &updatedOrder.TakeoutOrderItems[i]
		if existingItem, exists := existingItemMap[item.PlatformItemId]; exists {
			// 复用现有商品的 UUID 和映射信息
			item.Uuid = existingItem.Uuid
			item.TakeoutOrderUuid = orderUuid
			item.IsMapped = existingItem.IsMapped
			item.TtposProductPackageUuid = existingItem.TtposProductPackageUuid
			item.TtposProductType = existingItem.TtposProductType
		}
	}

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)
		itemRepoTx := persistence.NewTakeoutOrderItemRepo(tx)
		modifierRepoTx := persistence.NewTakeoutOrderItemModifierRepo(tx)

		// 1. 更新订单主表字段
		updateFields := map[string]interface{}{
			"subtotal":              updatedOrder.Subtotal,
			"delivery_fee":          updatedOrder.DeliveryFee,
			"small_order_fee":       updatedOrder.SmallOrderFee,
			"eater_payment":         updatedOrder.EaterPayment,
			"platform_discount":     updatedOrder.PlatformDiscount,
			"merchant_discount":     updatedOrder.MerchantDiscount,
			"basket_promo":          updatedOrder.BasketPromo,
			"tax":                   updatedOrder.Tax,
			"merchant_charge_fee":   updatedOrder.MerchantChargeFee,
			"platform_total":        updatedOrder.PlatformTotal,
			"additional_properties": updatedOrder.AdditionalProperties,
			"raw_data":              updatedOrder.RawData,
		}
		if err := orderRepoTx.UpdateByMap(orderUuid, updateFields); err != nil {
			logger.Logger.Error("更新订单主表失败", zap.Uint64("orderUuid", orderUuid), zap.Error(err))
			return fmt.Errorf("更新订单主表失败: %w", err)
		}

		// 2. 处理退菜项（删除或减少数量）
		for _, change := range changeResult.ReturnItems {
			switch change.ChangeType {
			case valueobject.ChangeTypeRemoved:
				// 删除商品（软删除）
				if change.OldItem != nil && change.OldItem.Uuid > 0 {
					if err := itemRepoTx.Delete(change.OldItem.Uuid); err != nil {
						logger.Logger.Error("删除退菜商品失败",
							zap.Uint64("itemUuid", change.OldItem.Uuid),
							zap.String("productName", change.ProductName),
							zap.Error(err))
					}
				}
			case valueobject.ChangeTypeQuantity:
				// 更新数量（减少）
				if change.OldItem != nil && change.OldItem.Uuid > 0 {
					if err := itemRepoTx.UpdateQuantity(change.OldItem.Uuid, change.NewQuantity); err != nil {
						logger.Logger.Error("更新退菜商品数量失败",
							zap.Uint64("itemUuid", change.OldItem.Uuid),
							zap.Error(err))
					}
				}
			case valueobject.ChangeTypeAttribute:
				// 属性变动：更新商品和修饰符
				if change.OldItem != nil && change.OldItem.Uuid > 0 {
					// 删除旧修饰符
					if err := modifierRepoTx.DeleteByItemUuid(change.OldItem.Uuid); err != nil {
						logger.Logger.Error("删除旧修饰符失败", zap.Uint64("itemUuid", change.OldItem.Uuid), zap.Error(err))
					}
					// 更新数量
					if err := itemRepoTx.UpdateQuantity(change.OldItem.Uuid, change.NewQuantity); err != nil {
						logger.Logger.Error("更新商品数量失败", zap.Error(err))
					}
					// 插入新修饰符
					for j := range updatedOrder.TakeoutOrderItems {
						newItem := &updatedOrder.TakeoutOrderItems[j]
						if newItem.PlatformItemId == change.PlatformItemId {
							// 使用已存在商品的 UUID，确保修饰符关联正确
							newItem.Uuid = change.OldItem.Uuid
							// 补全修饰符信息（UUID、映射、名称等）
							if _, err := s.enrichModifiersInfo(ctx, platform, newItem); err != nil {
								logger.Logger.Error("补全修饰符信息失败", zap.Error(err))
							}
							if len(newItem.TakeoutOrderItemModifiers) > 0 {
								if err := modifierRepoTx.BatchCreate(s.toModifierPointers(newItem.TakeoutOrderItemModifiers)); err != nil {
									logger.Logger.Error("创建新修饰符失败", zap.Error(err))
								}
							}
							break
						}
					}
				}
			}
		}

		// 3. 处理送厨项（新增或增加数量）
		// 注意：使用索引迭代以便更新 changeResult 中的 NewItem（用于后续生产单同步）
		for i := range changeResult.KitchenItems {
			change := &changeResult.KitchenItems[i]
			switch change.ChangeType {
			case valueobject.ChangeTypeAdded:
				// 新增商品
				for j := range updatedOrder.TakeoutOrderItems {
					newItem := &updatedOrder.TakeoutOrderItems[j]
					if newItem.PlatformItemId == change.PlatformItemId {
						// 补全商品信息（UUID、映射、名称等）
						if _, err := s.enrichItemInfo(ctx, platform, orderUuid, newItem); err != nil {
							logger.Logger.Error("补全商品信息失败", zap.Error(err))
							continue
						}

						// 同步更新 changeResult 中的 NewItem 映射信息（用于后续生产单同步和打印）
						if change.NewItem != nil {
							change.NewItem.Uuid = newItem.Uuid
							change.NewItem.TtposProductPackageUuid = newItem.TtposProductPackageUuid
							change.NewItem.TtposProductType = newItem.TtposProductType
							change.NewItem.ItemName = newItem.ItemName
						}

						// 插入商品
						if err := itemRepoTx.Create(newItem); err != nil {
							logger.Logger.Error("创建新增商品失败", zap.Error(err))
							continue
						}

						// 补全修饰符信息（UUID、映射、名称等）
						if _, err := s.enrichModifiersInfo(ctx, platform, newItem); err != nil {
							logger.Logger.Error("补全修饰符信息失败", zap.Error(err))
						}
						if len(newItem.TakeoutOrderItemModifiers) > 0 {
							if err := modifierRepoTx.BatchCreate(s.toModifierPointers(newItem.TakeoutOrderItemModifiers)); err != nil {
								logger.Logger.Error("创建新增商品修饰符失败", zap.Error(err))
							}
						}
						break
					}
				}
			case valueobject.ChangeTypeQuantity:
				// 数量增加：更新数量
				if change.OldItem != nil && change.OldItem.Uuid > 0 {
					if err := itemRepoTx.UpdateQuantity(change.OldItem.Uuid, change.NewQuantity); err != nil {
						logger.Logger.Error("更新加菜商品数量失败", zap.Uint64("itemUuid", change.OldItem.Uuid), zap.Error(err))
					}
				}
			}
		}

		return nil
	})
}

// toModifierPointers 将修饰符切片转换为指针切片
func (s *takeoutOrderSrv) toModifierPointers(modifiers []takeoutModel.TakeoutOrderItemModifier) []*takeoutModel.TakeoutOrderItemModifier {
	result := make([]*takeoutModel.TakeoutOrderItemModifier, len(modifiers))
	for i := range modifiers {
		result[i] = &modifiers[i]
	}
	return result
}

// EnrichOrderItems 为订单商品填充 TTPOS 商品映射信息
// 遍历订单中的所有商品，调用 enrichItemInfo 和 enrichModifiersInfo 填充：
// - 商品映射信息（TtposProductPackageUuid、TtposProductType）
// - 商品详细信息（名称、价格、分类、ERP编码等）
// - 修饰符映射和详细信息
// 返回值：abnormalItemUuids 是异常商品的UUID列表（商品或修饰符信息缺失）
func (s *takeoutOrderSrv) EnrichOrderItems(ctx context.Context, order *takeoutModel.TakeoutOrder) (abnormalItemUuids []uint64, err error) {
	if order == nil || len(order.TakeoutOrderItems) == 0 {
		return nil, nil
	}

	abnormalUuidSet := make(map[uint64]struct{})

	for i := range order.TakeoutOrderItems {
		item := &order.TakeoutOrderItems[i]

		// 1. 填充商品信息
		isAbnormal, err := s.enrichItemInfo(ctx, order.Platform, order.Uuid, item)
		if err != nil {
			logger.Logger.Error("填充商品信息失败",
				zap.Error(err),
				zap.String("platformItemId", item.PlatformItemId),
				zap.Uint64("orderUuid", order.Uuid))
			return nil, errors.WithMessage(errors.New("填充商品信息失败"), err.Error())
		}
		if isAbnormal {
			abnormalUuidSet[item.Uuid] = struct{}{}
		}

		// 2. 填充修饰符信息
		modifierAbnormalUuids, err := s.enrichModifiersInfo(ctx, order.Platform, item)
		if err != nil {
			logger.Logger.Error("填充修饰符信息失败",
				zap.Error(err),
				zap.String("platformItemId", item.PlatformItemId),
				zap.Uint64("orderUuid", order.Uuid))
			return nil, errors.WithMessage(errors.New("填充修饰符信息失败"), err.Error())
		}
		for _, abnormalUuid := range modifierAbnormalUuids {
			abnormalUuidSet[abnormalUuid] = struct{}{}
		}
	}

	// 转换为切片返回
	abnormalItemUuids = make([]uint64, 0, len(abnormalUuidSet))
	for uuid := range abnormalUuidSet {
		abnormalItemUuids = append(abnormalItemUuids, uuid)
	}

	return abnormalItemUuids, nil
}

// enrichModifiersInfo 完整处理商品的修饰符信息
// 包含：生成 UUID、查询映射、补全名称/价格/分类等详细信息
// 返回值：abnormalUuids 是异常修饰符的商品UUID列表（名称为空），err 表示处理错误
func (s *takeoutOrderSrv) enrichModifiersInfo(ctx context.Context, platform string, item *takeoutModel.TakeoutOrderItem) (abnormalUuids []uint64, err error) {
	if len(item.TakeoutOrderItemModifiers) == 0 {
		return nil, nil
	}

	currentTime := time.Now().Unix()

	// 第一阶段：生成 UUID、设置基础信息、查询映射
	modifierUuids := make([]uint64, 0)
	modifierTypes := make(map[uint64]string)
	modifierPlatformIds := make([]string, 0)

	for j := range item.TakeoutOrderItemModifiers {
		modifier := &item.TakeoutOrderItemModifiers[j]

		// 生成修饰符 UUID（如果没有）
		if modifier.Uuid == 0 {
			modifierUuid, err := utils.GetID()
			if err != nil {
				logger.Logger.Error("生成修饰符UUID失败", zap.Error(err))
				return nil, errors.WithMessage(errors.New("生成修饰符UUID失败"), err.Error())
			}
			modifier.Uuid = modifierUuid
		}

		// 设置关联和时间
		modifier.TakeoutOrderItemUuid = item.Uuid
		modifier.CreateTime = currentTime
		modifier.UpdateTime = currentTime

		// 查询修饰符映射（如果没有）
		if modifier.TtposModifierUuid == 0 {
			isMapped, ttposUuid, ttposType, err := s.findModifierMapping(platform, modifier.PlatformModifierId)
			if err != nil {
				logger.Logger.Error("查询修饰符映射失败", zap.Error(err), zap.String("platformModifierId", modifier.PlatformModifierId))
				return nil, errors.WithMessage(errors.New("查询修饰符映射失败"), err.Error())
			}
			modifier.IsMapped = isMapped
			modifier.TtposModifierUuid = ttposUuid
			modifier.TtposModifierType = ttposType
		}

		// 收集需要查询详细信息的修饰符
		if modifier.IsMapped == 1 && modifier.TtposModifierUuid > 0 {
			modifierUuids = append(modifierUuids, modifier.TtposModifierUuid)
			modifierTypes[modifier.TtposModifierUuid] = modifier.TtposModifierType
		} else if modifier.IsMapped == 0 {
			modifierPlatformIds = append(modifierPlatformIds, modifier.PlatformModifierId)
		}
	}

	// 第二阶段：批量查询修饰符名称
	modifierInfos := s.menuRepo.GetModifierNamesByUuids(ctx, modifierUuids, modifierTypes)
	platformModifierNames := s.menuRepo.GetModifierNamesByPlatformIds(ctx, platform, modifierPlatformIds)

	// 第三阶段：设置修饰符详细信息
	abnormalUuids = make([]uint64, 0)
	for j := range item.TakeoutOrderItemModifiers {
		modifier := &item.TakeoutOrderItemModifiers[j]

		if modifier.IsMapped == 1 && modifier.TtposModifierUuid > 0 {
			// 已映射修饰符：使用修饰符名称和数量
			if info, ok := modifierInfos[modifier.TtposModifierUuid]; ok {
				// ModifierName: 用于显示（commodity: 外卖表优先, 其他: 核心表）
				modifier.ModifierName = info.Name
				// TtposModifierName: 用于标识（始终来自核心表）
				modifier.TtposModifierName = info.TtposName
				// TtposModifierErpCode: ERP编码
				modifier.TtposModifierErpCode = info.TtposErpCode
				// TtposPrice: TTPOS 店内价格
				modifier.TtposPrice = info.TtposPrice
				// 分类信息
				modifier.TtposCategoryUuid = info.TtposCategoryUuid
				modifier.TtposCategoryName = info.TtposCategoryName
				modifier.TtposParentCategoryUuid = info.TtposParentCategoryUuid
				modifier.TtposParentCategoryName = info.TtposParentCategoryName
				// 计算单价
				if modifier.Quantity > 0 {
					modifier.Price = decimal.NewFromInt(int64(modifier.Price)).Div(decimal.NewFromInt(int64(modifier.Quantity))).InexactFloat64()
				}
				// 如果是 commodity 类型，设置规格信息和数量
				if modifier.IsCommodity() {
					modifier.TtposProductPackageUuid = info.TtposProductPackageUuid
					modifier.TtposFlavorProductBomUuid = info.TtposFlavorProductBomUuid
					modifier.TtposFlavorName = info.TtposFlavorName
					// 使用TTPOS数量覆盖平台数量
					if info.Num > 0 {
						modifier.Price = decimal.NewFromInt(int64(modifier.Price)).Div(decimal.NewFromInt(int64(info.Num))).InexactFloat64()
						modifier.Quantity = int(decimal.NewFromInt(int64(info.Num)).Mul(decimal.NewFromInt(int64(item.Quantity))).InexactFloat64())
					}
				}
				// 检查是否异常（名称为空）
				if info.Name == "" || info.TtposName == "" {
					abnormalUuids = append(abnormalUuids, modifier.TtposProductPackageUuid)
				}
			}
		} else if modifier.IsMapped == 0 {
			if name, ok := platformModifierNames[modifier.PlatformModifierId]; ok && name != "" {
				modifier.ModifierName = name
			}
		}
	}
	return abnormalUuids, nil
}

// enrichItemInfo 完整处理商品信息
// 包含：生成 UUID、查询映射、补全名称/价格/分类等详细信息
// orderUuid: 订单UUID，用于设置商品的关联
// 返回值：isAbnormal 表示商品是否异常（名称为空），err 表示处理错误
func (s *takeoutOrderSrv) enrichItemInfo(ctx context.Context, platform string, orderUuid uint64, item *takeoutModel.TakeoutOrderItem) (isAbnormal bool, err error) {
	// 生成商品 UUID（如果没有）
	if item.Uuid == 0 {
		itemUuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成商品UUID失败", zap.Error(err))
			return false, errors.WithMessage(errors.New("生成商品UUID失败"), err.Error())
		}
		item.Uuid = itemUuid
	}

	// 设置订单关联
	item.TakeoutOrderUuid = orderUuid

	// 查询商品映射（如果没有）
	if item.TtposProductPackageUuid == 0 {
		isMapped, productUuid, productType, err := s.findProductMapping(platform, item.PlatformItemId)
		if err != nil {
			logger.Logger.Error("查询商品映射失败", zap.Error(err), zap.String("platformItemId", item.PlatformItemId))
			return false, errors.WithMessage(errors.New("查询商品映射失败"), err.Error())
		}
		item.IsMapped = isMapped
		item.TtposProductPackageUuid = productUuid
		item.TtposProductType = productType
	}

	// 补全商品详细信息
	if item.IsMapped == 1 && item.TtposProductPackageUuid > 0 {
		// 已映射商品：查询商品名称
		productUuids := []uint64{item.TtposProductPackageUuid}
		productTypes := map[uint64]int{item.TtposProductPackageUuid: item.TtposProductType}
		productInfos := s.menuRepo.GetProductNamesByUuids(ctx, productUuids, productTypes)

		if info, ok := productInfos[item.TtposProductPackageUuid]; ok {
			// ItemName: 显示用名称（外卖表优先）
			if info.Name != "" {
				item.ItemName = info.Name
			}
			// TtposItemName: TTPOS 核心表名称
			item.TtposItemName = info.TtposName
			// TtposItemErpCode: ERP编码
			item.TtposItemErpCode = info.TtposErpCode
			// TtposPrice: TTPOS 店内价格
			item.TtposPrice = info.TtposPrice
			// 分类信息
			item.TtposCategoryUuid = info.TtposCategoryUuid
			item.TtposCategoryName = info.TtposCategoryName
			item.TtposParentCategoryUuid = info.TtposParentCategoryUuid
			item.TtposParentCategoryName = info.TtposParentCategoryName
			// 检查是否异常（名称为空）
			if info.TtposName == "" || info.Name == "" {
				isAbnormal = true
			}
		}
	} else if item.IsMapped == 0 {
		// 未映射商品：使用平台菜单名称
		platformIds := []string{item.PlatformItemId}
		menuNames := s.menuRepo.GetMenuNamesByPlatformItemIds(ctx, platform, platformIds)
		if name, ok := menuNames[item.PlatformItemId]; ok && name != "" {
			item.ItemName = name
		}
	}
	return isAbnormal, nil
}

// extractAndSaveTakeoutOrderMaterials 提取并保存外卖订单原料（在创建订单时调用）
//
// 功能：根据订单商品的 BOM 配方，计算并保存原料消耗记录
// 流程：商品 → BOM配方 → 原料清单 → 计算消耗量 → 保存记录
func (s *takeoutOrderSrv) extractAndSaveTakeoutOrderMaterials(ctx context.Context, order *takeoutModel.TakeoutOrder) error {
	db := ctx.GetDB()

	// ==================== Step 1: 构建 BOM 数量映射 ====================
	// bomQuantityMap: BOM UUID → 出库数量（用于计算原料消耗）
	// bomItemMap: BOM UUID → 订单商品信息（包含来源追溯）
	bomQuantityMap, bomItemMap, err := s.BuildBomQuantityMap(ctx, order)
	if err != nil {
		return errors.WithMessage(err, "构建BOM数量映射失败")
	}
	if len(bomQuantityMap) == 0 {
		return nil // 订单没有配方商品，无需处理原料
	}

	// ==================== Step 2: 批量查询 BOM 配方（预加载原材料） ====================
	bomUuids := make([]uint64, 0, len(bomQuantityMap))
	for bomUuid := range bomQuantityMap {
		bomUuids = append(bomUuids, bomUuid)
	}
	productBoms, err := persistence.NewTakeoutBomMappingRepo(db).GetProductBomsWithMaterials(bomUuids)
	if err != nil {
		return errors.WithMessage(err, "查询BOM配方信息失败")
	}

	// 构建 BOM UUID → ProductBom 的快速查找映射
	bomConfigMap := make(map[uint64]*coreModel.ProductBom, len(productBoms))
	for _, bom := range productBoms {
		bomConfigMap[bom.Uuid] = bom
	}

	// ==================== Step 3: 收集原料并查询 ERP 编码 ====================
	// 收集所有涉及的原料UUID
	materialUuidSet := make(map[uint64]bool)

	// 定义辅助函数：获取配方的原材料清单（优先使用配方卡）
	getMaterials := func(bom *coreModel.ProductBom, isFlavor bool) []*coreModel.RelatedMaterial {
		if isFlavor {
			if bom.HasProductBomCard() && bom.ProductBomCard != nil {
				return bom.ProductBomCard.RelatedMaterials
			}
			return bom.FlavorMaterials
		} else { // sauce
			if bom.ProductSauce.HasProductBomCard() && bom.ProductSauce.ProductBomCard != nil {
				return bom.ProductSauce.ProductBomCard.RelatedMaterials
			}
			return bom.ProductSauce.SauceMaterials
		}
	}

	// 遍历所有BOM，收集原料UUID
	for bomUuid := range bomQuantityMap {
		bom, ok := bomConfigMap[bomUuid]
		if !ok {
			continue
		}

		// 收集规格的原材料
		if bom.IsFlavor() {
			for _, material := range getMaterials(bom, true) {
				if material.Material != nil {
					materialUuidSet[material.MaterialUuid] = true
				}
			}
		}

		// 收集加料的原材料
		if bom.IsSauce() && bom.ProductSauceUuid > 0 {
			for _, material := range getMaterials(bom, false) {
				if material.Material != nil {
					materialUuidSet[material.MaterialUuid] = true
				}
			}
		}
	}

	// 批量查询原料信息（获取ERP编码和名称）
	materialUuids := make([]uint64, 0, len(materialUuidSet))
	for materialUuid := range materialUuidSet {
		materialUuids = append(materialUuids, materialUuid)
	}
	materials, err := repository.NewMaterialRepo(db).GetMaterialContainsDeletedByUuids(materialUuids)
	if err != nil {
		return errors.WithMessage(err, "批量查询原料信息失败")
	}

	// 构建 原料UUID → (ERP编码, 名称) 的映射
	type materialInfo struct {
		erpCode string
		name    string
	}
	materialInfoMap := make(map[uint64]materialInfo, len(materials))
	for _, material := range materials {
		materialInfoMap[material.Uuid] = materialInfo{
			erpCode: material.Code,
			name:    material.Name,
		}
	}

	// ==================== Step 4: 计算原料消耗量（带来源追溯） ====================
	// 原料消耗来源记录
	type materialSource struct {
		materialUuid                 uint64  // 原料UUID
		warehouseUuid                uint64  // 仓库UUID
		consumptionNum               float64 // 消耗数量
		productBomUuid               uint64  // BOM UUID（用于出库清单聚合）
		takeoutOrderItemUuid         uint64  // 来源：订单商品项
		takeoutOrderItemModifierUuid uint64  // 来源：订单商品修饰符
		baseUnitUom                  string  // 基准单位UOM（来自RelatedMaterial.BaseUnitUom）
	}

	// 定义组合键：material + warehouse + modifier（确保每个 modifier 独立记录）
	type sourceKey struct {
		materialUuid  uint64
		warehouseUuid uint64
		modifierUuid  uint64 // 关键：按 modifier 区分
	}

	// 消耗映射：sourceKey → source
	consumptionMap := make(map[sourceKey]*materialSource)

	// 定义辅助函数：获取默认仓库
	getDefaultWarehouse := func(material *coreModel.Material) uint64 {
		if material.WarehouseUuid > 0 {
			return material.WarehouseUuid
		}
		if len(material.WarehouseItems) > 0 {
			return material.WarehouseItems[0].WarehouseUuid
		}
		return 0
	}

	// 遍历每个BOM，计算原料消耗
	for bomUuid := range bomQuantityMap {
		bom, ok := bomConfigMap[bomUuid]
		if !ok {
			continue
		}

		// 获取该 BOM 对应的商品项信息
		sourceItem := bomItemMap[bomUuid]
		if sourceItem == nil {
			continue
		}

		// 关键：遍历该 BOM 的每个 modifier，独立计算原料消耗
		for _, modifier := range sourceItem.TakeoutOrderItemModifiers {
			sourceItemUuid := sourceItem.Uuid
			sourceModifierUuid := modifier.Uuid

			// 每个 modifier 的消耗数量：modifier.Quantity * item.Quantity
			productNum := float64(modifier.Quantity * sourceItem.Quantity)

			// 处理规格的原料消耗
			if bom.IsFlavor() {
				for _, material := range getMaterials(bom, true) {
					if material.Material == nil {
						continue
					}
					warehouseUuid := getDefaultWarehouse(material.Material)
					consumptionNum := material.Num * productNum // 原料配比 × 商品数量

					// 构建唯一键（按 modifier 区分）
					key := sourceKey{
						materialUuid:  material.MaterialUuid,
						warehouseUuid: warehouseUuid,
						modifierUuid:  sourceModifierUuid,
					}

					// 如果已存在则累加，否则创建新记录
					if consumptionMap[key] == nil {
						consumptionMap[key] = &materialSource{
							materialUuid:                 material.MaterialUuid,
							warehouseUuid:                warehouseUuid,
							consumptionNum:               0,
							productBomUuid:               bomUuid,
							takeoutOrderItemUuid:         sourceItemUuid,
							takeoutOrderItemModifierUuid: sourceModifierUuid,
							baseUnitUom:                  material.BaseUnitUom,
						}
					}
					consumptionMap[key].consumptionNum += consumptionNum
				}
			}

			// 处理加料的原料消耗
			if bom.IsSauce() && bom.ProductSauceUuid > 0 {
				for _, material := range getMaterials(bom, false) {
					if material.Material == nil {
						continue
					}
					warehouseUuid := getDefaultWarehouse(material.Material)
					consumptionNum := material.Num * productNum // 原料配比 × 商品数量

					// 构建唯一键（按 modifier 区分）
					key := sourceKey{
						materialUuid:  material.MaterialUuid,
						warehouseUuid: warehouseUuid,
						modifierUuid:  sourceModifierUuid,
					}

					// 如果已存在则累加，否则创建新记录
					if consumptionMap[key] == nil {
						consumptionMap[key] = &materialSource{
							materialUuid:                 material.MaterialUuid,
							warehouseUuid:                warehouseUuid,
							consumptionNum:               0,
							productBomUuid:               bomUuid,
							takeoutOrderItemUuid:         sourceItemUuid,
							takeoutOrderItemModifierUuid: sourceModifierUuid,
							baseUnitUom:                  material.BaseUnitUom,
						}
					}
					consumptionMap[key].consumptionNum += consumptionNum
				}
			}
		}
	}

	// ==================== Step 5: 构建并保存原料消耗记录 ====================
	materialRecords := make([]*takeoutModel.TakeoutOrderMaterial, 0)
	for _, source := range consumptionMap {
		info := materialInfoMap[source.materialUuid]
		materialRecords = append(materialRecords, &takeoutModel.TakeoutOrderMaterial{
			TakeoutOrderUuid:             order.Uuid,
			TakeoutOrderItemUuid:         source.takeoutOrderItemUuid,
			TakeoutOrderItemModifierUuid: source.takeoutOrderItemModifierUuid,
			MaterialUuid:                 source.materialUuid,
			MaterialName:                 info.name,
			ErpCode:                      info.erpCode,
			BaseUnitUom:                  source.baseUnitUom, // 从 source 中获取
			WarehouseUuid:                source.warehouseUuid,
			Num:                          source.consumptionNum,
			IsSummarized:                 0, // 初始为未统计
			ProductBomUuid:               source.productBomUuid,
		})
	}

	// 批量保存
	if len(materialRecords) > 0 {
		materialSrv := NewTakeoutOrderMaterialSrv(s.dbm)
		if err := materialSrv.SaveOrderMaterials(ctx, materialRecords); err != nil {
			return errors.WithMessage(err, "保存订单原料失败")
		}
	}

	return nil
}

// UpdateOrderStatus 更新订单状态
func (s *takeoutOrderSrv) UpdateOrderStatus(ctx context.Context, orderUuid string, status string, message string) error {
	db := ctx.GetDB()

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)

		// 1. 查询订单（通过 takeout_order_uuid 字符串查询）
		order, err := orderRepoTx.GetByTakeoutOrderUuid(orderUuid)
		if err != nil {
			logger.Logger.Error("查询订单失败", zap.String("order_uuid", orderUuid), zap.Error(err))
			return errors.WithMessage(err, "查询订单失败")
		}

		if order == nil {
			logger.Logger.Error("订单不存在", zap.String("order_uuid", orderUuid))
			return errors.New("订单不存在")
		}

		// 3. 转换并更新订单状态
		// 使用平台转换器将平台状态转换为内部状态码
		oldOrderState := order.OrderState
		newOrderState := func() int {
			if order.IsLinemanOrder() {
				return lineman.ConvertPlatformStateToOrderState(status, order.OrderState)
			}
			return grab.ConvertPlatformStateToOrderState(status, order.OrderState)
		}()
		if newOrderState != -1 {
			order.OrderState = newOrderState // 更新内部状态码
		}
		order.PlatformOrderState = status // 更新平台原始状态

		// 更新取消时间
		if newOrderState == valueobject.TakeoutOrderStateCanceled || newOrderState == valueobject.TakeoutOrderStateRejected {
			order.RejectedTime = time.Now().Unix()
			order.RejectReason = message
		}

		// 4. 更新订单到数据库
		if err := orderRepoTx.Update(order); err != nil {
			logger.Logger.Error("更新订单数据失败", zap.String("order_uuid", orderUuid), zap.Error(err))
			return errors.WithMessage(err, "更新订单数据失败")
		}

		// 5. 发布订单状态更新事件（仅在状态发生变化时）
		if oldOrderState != newOrderState {
			switch newOrderState {
			case valueobject.TakeoutOrderStateRiderProcessing:
				// 骑手配送中事件
				event.GetDispatcher().Publish(event.NewOrderRiderProcessingEvent(
					order.Uuid,
					order.Platform,
					order.PlatformOrderId,
					order.ShortOrderNumber,
					order.TakeoutOrderUuid,
					ctx.GetCompanyUuid(),
				))
			case valueobject.TakeoutOrderStateCanceled, valueobject.TakeoutOrderStateRejected:
				// 订单取消事件
				event.GetDispatcher().Publish(event.NewOrderCancelEvent(
					order.Uuid,
					order.Platform,
					order.PlatformOrderId,
					order.ShortOrderNumber,
					order.TakeoutOrderUuid,
					ctx.GetCompanyUuid(),
					"订单已取消",
					newOrderState,
				))
			case valueobject.TakeoutOrderStateCompleted:
				// 订单完成事件
				event.GetDispatcher().Publish(event.NewOrderCompletedEvent(
					order.Uuid,
					order.Platform,
					order.PlatformOrderId,
					order.ShortOrderNumber,
					order.TakeoutOrderUuid,
					ctx.GetCompanyUuid(),
				))
			}
		}

		return nil
	})
}

// BuildBomQuantityMap 构建订单商品的 BOM 数量映射（用于库存检查和出库）
// 返回: bomQuantityMap (BOM UUID -> 出库数量), bomItemMap (BOM UUID -> 订单商品，包含 modifiers), bomUuids (BOM UUID 列表), error
//
// 出库数量计算规则：
//   - 主商品：商品数量（item.Quantity）
//   - 规格(flavor)/加料(sauce): modifier.Quantity * item.Quantity (modifier数量 × 主商品数量)
//   - 套餐商品(commodity): groupItem.Num * item.Quantity (套餐配置数量 × 主商品数量)
//
// 销售数量：主商品的 item.Quantity（用于统计销售）
func (s *takeoutOrderSrv) BuildBomQuantityMap(ctx context.Context, order *takeoutModel.TakeoutOrder) (map[uint64]int, map[uint64]*takeoutModel.TakeoutOrderItem, error) {
	db := ctx.GetDB()
	bomMappingRepo := persistence.NewTakeoutBomMappingRepo(db)

	// 收集 modifier 信息的辅助结构
	type modifierInfo struct {
		modifier         *takeoutModel.TakeoutOrderItemModifier
		item             *takeoutModel.TakeoutOrderItem // 主商品
		outboundQuantity int                            // 出库数量
	}

	// 按类型分组收集 modifier
	var (
		flavorModifiers    = make(map[uint64]*modifierInfo) // bomUuid -> modifierInfo
		sauceModifiers     = make(map[uint64]*modifierInfo) // bomUuid -> modifierInfo
		commodityModifiers = make(map[uint64]*modifierInfo) // groupItemUuid -> modifierInfo
	)

	// 1. 遍历订单商品，按类型收集所有 modifier
	for i := range order.TakeoutOrderItems {
		item := &order.TakeoutOrderItems[i]
		for j := range item.TakeoutOrderItemModifiers {
			modifier := &item.TakeoutOrderItemModifiers[j]
			if modifier.IsMapped == 0 {
				continue
			}
			info := &modifierInfo{
				modifier:         modifier,
				item:             item,
				outboundQuantity: modifier.Quantity * item.Quantity,
			}
			switch {
			case modifier.IsFlavor():
				flavorModifiers[modifier.TtposModifierUuid] = info
			case modifier.IsSauce():
				sauceModifiers[modifier.TtposModifierUuid] = info
			case modifier.IsCommodity():
				commodityModifiers[modifier.TtposModifierUuid] = info
			}
		}
	}

	// 2. 查询 commodity 类型的 BOM 映射和配置数量
	groupItemToBomMap := make(map[uint64]persistence.GroupItemBomMapping) // groupItemUuid -> {bomUuid, num}
	if len(commodityModifiers) > 0 {
		groupItemUuids := make([]uint64, 0, len(commodityModifiers))
		for groupItemUuid := range commodityModifiers {
			groupItemUuids = append(groupItemUuids, groupItemUuid)
		}

		groupItemBomMapping, err := bomMappingRepo.GetGroupItemBomMapping(groupItemUuids)
		if err != nil {
			logger.Logger.Error("查询套餐商品BOM映射失败", zap.Error(err))
			return nil, nil, errors.WithMessage(errors.New("查询套餐商品BOM映射失败"), err.Error())
		}
		groupItemToBomMap = groupItemBomMapping
	}

	// 3. 构建 bomUuid -> [modifierInfo] 映射，并累加出库数量
	type bomData struct {
		quantity  int                            // 出库数量
		modifiers []*modifierInfo                // 关联的 modifiers
		item      *takeoutModel.TakeoutOrderItem // 主商品（取第一个 modifier 的 item）
	}
	bomMap := make(map[uint64]*bomData)

	// 辅助函数：添加 modifier 到 bomMap
	addToBomMap := func(bomUuid uint64, info *modifierInfo) {
		if data, ok := bomMap[bomUuid]; ok {
			data.quantity += info.outboundQuantity
			data.modifiers = append(data.modifiers, info)
		} else {
			bomMap[bomUuid] = &bomData{
				quantity:  info.outboundQuantity,
				modifiers: []*modifierInfo{info},
				item:      info.item,
			}
		}
	}

	// 添加 flavor 和 sauce（BOM UUID 直接可用）
	for bomUuid, info := range flavorModifiers {
		addToBomMap(bomUuid, info)
	}
	for bomUuid, info := range sauceModifiers {
		addToBomMap(bomUuid, info)
	}

	// 添加 commodity（需要通过 groupItemUuid 查找 BOM UUID）
	for groupItemUuid, info := range commodityModifiers {
		if mapping, ok := groupItemToBomMap[groupItemUuid]; ok {
			addToBomMap(mapping.ProductBomUuid, info)
		}
	}

	// 4. 构建返回结果
	bomQuantityMap := make(map[uint64]int, len(bomMap))
	bomItemMap := make(map[uint64]*takeoutModel.TakeoutOrderItem, len(bomMap))

	for bomUuid, data := range bomMap {
		bomQuantityMap[bomUuid] = data.quantity

		// 创建 item 副本，包含该 BOM 对应的所有 modifiers
		itemCopy := *data.item
		itemCopy.TakeoutOrderItemModifiers = make([]takeoutModel.TakeoutOrderItemModifier, len(data.modifiers))
		for i, info := range data.modifiers {
			itemCopy.TakeoutOrderItemModifiers[i] = *info.modifier
		}
		bomItemMap[bomUuid] = &itemCopy
	}

	return bomQuantityMap, bomItemMap, nil
}

// CalculateTakeoutOrderSalesVolume 计算外卖订单销量
// 返回: productBoms (BOM UUID -> 销量), productPackages (Package UUID -> 销量)
func (s *takeoutOrderSrv) CalculateTakeoutOrderSalesVolume(order *takeoutModel.TakeoutOrder) (map[uint64]float64, map[uint64]float64, error) {
	productBoms := make(map[uint64]float64)     // 规格商品销量 map[BOM UUID]销量
	productPackages := make(map[uint64]float64) // 套餐商品销量 map[Package UUID]销量

	// 遍历订单商品
	for _, item := range order.TakeoutOrderItems {
		// 只处理已映射的商品
		if item.IsMapped != 1 || item.TtposProductPackageUuid == 0 {
			continue
		}

		itemQuantity := float64(item.Quantity)

		// 统计主商品的 Package 销量
		// 不管是套餐还是普通商品，TtposProductPackageUuid 都是 ProductPackage 的 UUID
		productPackages[item.TtposProductPackageUuid] += itemQuantity

		// 遍历修饰符，计算规格和加料的销量
		for _, modifier := range item.TakeoutOrderItemModifiers {
			if modifier.IsMapped != 1 || modifier.TtposModifierUuid == 0 {
				continue
			}

			// 规格(flavor)的销量：这是主商品的 BOM 销量
			// 规格数量 × 主商品数量
			if modifier.IsFlavor() {
				modifierQuantity := float64(modifier.Quantity) * itemQuantity
				productBoms[modifier.TtposModifierUuid] += modifierQuantity
			}

			// 加料(sauce)的销量：额外添加的小料 BOM 销量
			// 加料数量 × 主商品数量
			if modifier.IsSauce() {
				modifierQuantity := float64(modifier.Quantity) * itemQuantity
				productBoms[modifier.TtposModifierUuid] += modifierQuantity
			}

			// 套餐商品(commodity)的销量：套餐内的子商品 BOM 销量
			// 子商品数量 × 主商品数量
			if modifier.IsCommodity() {
				modifierQuantity := float64(modifier.Quantity) * itemQuantity
				productBoms[modifier.TtposModifierUuid] += modifierQuantity
			}
		}
	}

	return productBoms, productPackages, nil
}

// GetOrderForPrint 获取打印所需的订单数据
func (s *takeoutOrderSrv) GetOrderForPrint(ctx context.Context, orderUuid uint64) (*takeoutModel.TakeoutOrder, error) {
	db := ctx.GetDB()

	// 查询订单（包含商品、修饰符等信息）
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(
		orderUuid,
		orderRepo.WithTakeoutOrderItems(),
		orderRepo.WithTakeoutOrderItemModifiers(),
		orderRepo.WithTakeoutOrderUpdateLogs(orderRepo.Select("uuid", "takeout_order_uuid", "create_time")),
	)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.Uint64("uuid", orderUuid))
		return nil, errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	return order, nil
}

// BatchAssignShiftLogToPendingOrders 批量将待分配班次的订单分配给指定班次
// 对于已接单的订单，会同步生成 ERP 发票
func (s *takeoutOrderSrv) BatchAssignShiftLogToPendingOrders(ctx context.Context, shiftLogUuid, staffUuid uint64) error {
	db := ctx.GetDB()
	orderRepo := persistence.NewTakeoutOrderRepo(db)

	// 1. 通过 repository 查询所有待分配班次的订单
	pendingOrders, err := orderRepo.GetPendingShiftLogOrders()
	if err != nil {
		return errors.WithMessage(err, "查询待分配班次的订单失败")
	}

	if len(pendingOrders) == 0 {
		logger.Logger.Info("没有待分配班次的订单")
		return nil
	}

	// 2. 将订单分为两类：需要生成 ERP 发票的订单和不需要的订单
	var ordersNeedErpInvoice []*takeoutModel.TakeoutOrder
	var ordersNoErpInvoice []*takeoutModel.TakeoutOrder

	for _, order := range pendingOrders {
		// 需要生成 ERP 发票的条件：已接单且未同步 ERP 发票
		if order.OrderState != valueobject.TakeoutOrderStatePending && !order.IsErpInvoiceSynced() {
			ordersNeedErpInvoice = append(ordersNeedErpInvoice, order)
		} else {
			ordersNoErpInvoice = append(ordersNoErpInvoice, order)
		}
	}

	// 3. 对于需要生成 ERP 发票的订单，在事务中同时完成班次分配和 ERP 发票生成
	successCount := 0
	for _, order := range ordersNeedErpInvoice {
		err := db.Transaction(func(tx *gorm.DB) error {
			ctxTx := ctx.Copy()
			ctxTx.SetDB(tx)

			// 3.1 在事务中分配班次
			updatedCount, err := persistence.NewTakeoutOrderRepo(tx).BatchAssignShiftLog(
				shiftLogUuid,
				staffUuid,
				[]uint64{order.Uuid},
			)
			if err != nil {
				return errors.WithMessage(err, "分配班次失败")
			}
			if updatedCount == 0 {
				return errors.New("订单班次分配失败，可能已被其他进程处理")
			}

			// 3.2 在事务中同步到 ERP（生成发票）
			if err := s.erpSyncService.SyncOrderToERP(ctxTx, order.Uuid); err != nil {
				return errors.WithMessage(err, "同步订单到 ERP 失败")
			}

			return nil
		})

		// 继续处理其他订单，不返回错误
		if err != nil {
			logger.Logger.Warn("处理订单失败（事务回滚）", zap.Error(err), zap.Uint64("orderUuid", order.Uuid), zap.Uint64("shiftLogUuid", shiftLogUuid))
		} else {
			successCount++
			logger.Logger.Info("订单班次分配和 ERP 发票生成成功", zap.Uint64("orderUuid", order.Uuid), zap.Uint64("shiftLogUuid", shiftLogUuid))
		}
	}

	// 4. 对于不需要生成 ERP 发票的订单，批量分配班次
	if len(ordersNoErpInvoice) > 0 {
		orderUuids := make([]uint64, 0, len(ordersNoErpInvoice))
		for _, order := range ordersNoErpInvoice {
			orderUuids = append(orderUuids, order.Uuid)
		}
		_, err := orderRepo.BatchAssignShiftLog(shiftLogUuid, staffUuid, orderUuids)
		if err != nil {
			logger.Logger.Warn("批量分配订单班次失败", zap.Error(err), zap.Uint64("shiftLogUuid", shiftLogUuid), zap.Uint64("staffUuid", staffUuid))
		}
	}

	// 5. 批量记录高峰期（包括已接单和已取消的订单）
	allProcessedOrders := append(ordersNeedErpInvoice, ordersNoErpInvoice...)
	// 批量发布高峰期记录事件
	if len(allProcessedOrders) > 0 {
		companyUuid := ctx.GetCompanyUuid()
		// 收集订单UUID列表
		orderUuids := make([]uint64, 0, len(allProcessedOrders))
		for _, order := range allProcessedOrders {
			orderUuids = append(orderUuids, order.Uuid)
		}
		// 创建单个批量事件并发布
		peakTimeEvent := event.NewOrderPeakTimeRecordEvent(orderUuids, companyUuid)
		event.GetDispatcher().Publish(peakTimeEvent)
		logger.Logger.Debug("批量发布高峰期记录事件完成", zap.Int("count", len(orderUuids)))
	}

	logger.Logger.Debug("批量分配班次完成", zap.Int("erpInvoiceSuccessCount", successCount),
		zap.Int("erpInvoiceTotalCount", len(ordersNeedErpInvoice)),
		zap.Int("noErpInvoiceCount", len(ordersNoErpInvoice)),
		zap.Int("peakTimeRecordCount", len(allProcessedOrders)))

	return nil
}

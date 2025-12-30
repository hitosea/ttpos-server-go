package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/app/modules/takeout/domain/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/repository"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ITakeoutOrderSrv 外卖订单服务接口
type ITakeoutOrderSrv interface {
	// 订单查询
	GetList(ctx context.Context, req *request.TakeoutOrderListReq) (*response.TakeoutOrderListResp, error)
	GetByUuid(ctx context.Context, uuid uint64) (*response.TakeoutOrderResp, error)
	// 创建订单（接受已转换的订单对象，商品数据从 order.RawData 中解析）- 由 Application 层调用
	CreateOrder(ctx context.Context, order *model.TakeoutOrder) error
	// 更新订单状态 - 由 Application 层调用
	UpdateOrderStatus(ctx context.Context, orderUuid string, newStatus string) error
	// 订单操作
	AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error
	RejectOrder(ctx context.Context, req *request.TakeoutOrderRejectReq) error
	CallRider(ctx context.Context, req *request.TakeoutOrderCallRiderReq) error
	CheckOrderCancelable(ctx context.Context, req *request.TakeoutOrderCheckCancelableReq) (*response.TakeoutOrderCancelCheckResp, error)
	CancelOrder(ctx context.Context, req *request.TakeoutOrderCancelReq) error
	// GetOrderForPrint 获取打印所需的订单数据
	GetOrderForPrint(ctx context.Context, orderUuid uint64) (*model.TakeoutOrder, error)
	// BuildBomQuantityMap 构建订单商品的 BOM 数量映射（用于库存检查和出库）
	// 返回: bomQuantityMap (BOM UUID -> 数量), bomItemMap (BOM UUID -> 订单商品，包含 modifiers), error
	BuildBomQuantityMap(ctx context.Context, order *model.TakeoutOrder) (map[uint64]int, map[uint64]*model.TakeoutOrderItem, error)
	// CalculateTakeoutOrderSalesVolume 计算外卖订单销量
	// 返回: productBoms (BOM UUID -> 销量), productPackages (Package UUID -> 销量), error
	CalculateTakeoutOrderSalesVolume(order *model.TakeoutOrder) (map[uint64]float64, map[uint64]float64, error)
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
			Subtotal:         order.EaterPayment,
			TotalItems:       len(order.TakeoutOrderItems),
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
			return db.Preload("TakeoutOrderItems").
				Preload("TakeoutOrderItems.TakeoutOrderItemModifiers").
				Preload("TakeoutOrderReceiver").
				Preload("TakeoutOrderCampaigns").
				Preload("TakeoutOrderPromos")
		})),
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

	return &response.TakeoutOrderResp{
		Uuid:             order.Uuid,
		Platform:         order.Platform,
		ShortOrderNumber: order.ShortOrderNumber,
		OrderState:       order.OrderState,
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
		Price: response.TakeoutOrderPriceResp{
			Subtotal:          order.Subtotal,
			DeliveryFee:       order.DeliveryFee,
			EaterPayment:      order.EaterPayment,
			PlatformDiscount:  order.PlatformDiscount,
			MerchantDiscount:  order.MerchantDiscount,
			BasketPromo:       order.BasketPromo,
			Tax:               order.Tax,
			MerchantChargeFee: order.MerchantChargeFee,
		},
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
		Cutlery:    order.Cutlery,
		OrderType:  order.OrderType,
		TotalItems: len(itemList),
		Items:      itemList,
		Receiver:   receiverResp,
		Campaigns:  campaigns,
		Promos:     promos,
	}, nil
}

// AcceptOrder 接单
func (s *takeoutOrderSrv) AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
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
	// 检查订单状态
	if err := order.IsPendingOrder(); err != nil {
		return err
	}

	// 如果不是自动接单，则通知平台接受订单
	if !order.IsAutoAcceptOrder() {
		// 调用 BMP RPC 通知平台接受订单
		rpcClient, err := rpc.NewBMPTakeoutClient()
		if err != nil {
			logger.Logger.Error("创建 BMP RPC 客户端失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
		}
		defer rpcClient.Close()
		// 调用 PrepareOrder 接口（接受订单）
		if err := rpcClient.PrepareOrder(ctx.GetContext(), order.TakeoutOrderUuid, "Accepted"); err != nil {
			logger.Logger.Error("调用 BMP PrepareOrder 接口失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("通知平台接受订单失败"), err.Error())
		}
	}

	// 设置员工班次信息
	if err := orderRepo.SetStaffShiftLogUuid(order, userUuid); err != nil {
		logger.Logger.Error("设置员工班次日志UUID失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		ctxCopy := ctx.Copy()
		ctxCopy.SetDB(tx)
		// 更新订单状态
		updateData := map[string]interface{}{
			"order_state":          valueobject.TakeoutOrderStateAccepted,
			"accepted_time":        currentTime,
			"staff_shift_log_uuid": order.StaffShiftLogUuid,
			"accepted_by":          userUuid,
			"update_time":          currentTime,
		}
		if err := persistence.NewTakeoutOrderRepo(tx).UpdateByMap(order.Uuid, updateData); err != nil {
			logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
			return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
		}
		// 同步到 ERP
		if !order.IsAutoAcceptOrder() {
			if err := s.erpSyncService.SyncOrderToERP(ctxCopy, order.Uuid); err != nil {
				return errors.WithMessage(errors.New("同步 Grab 订单到 ERP 失败"), err.Error())
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
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

	// 检查订单状态 - 只有已接单配餐中的订单才能呼叫骑手
	if order.OrderState != valueobject.TakeoutOrderStateAccepted {
		return errors.New("订单状态不正确，只有已接单配餐中的订单才能呼叫骑手")
	}

	// 自动接单的订单在呼叫骑手时设置班次
	if order.IsAutoAcceptOrder() && order.StaffShiftLogUuid == 0 {
		if err := orderRepo.SetStaffShiftLogUuid(order, userUuid); err != nil {
			logger.Logger.Error("自动接单订单呼叫骑手时设置班次失败", zap.Error(err), zap.Uint64("order_uuid", order.Uuid))
		}
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
		return nil, errors.New("订单不存在")
	}
	// 检查订单状态
	if order.OrderState == valueobject.TakeoutOrderStateCompleted || order.OrderState == valueobject.TakeoutOrderStateRejected {
		return &response.TakeoutOrderCancelCheckResp{
			CanCancel:             false,
			NonCancellationReason: i18n.Translate(ctx.GetLanguage(), "已完成或已拒单的订单不能取消"),
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

	// 检查订单状态
	if order.OrderState == valueobject.TakeoutOrderStateCompleted || order.OrderState == valueobject.TakeoutOrderStateRejected {
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

	// 更新订单状态为已拒单（取消订单使用拒单状态）
	reasonText := ""
	for _, reason := range checkResp.CancelReasons {
		if reason.Code == req.ReasonCode {
			reasonText = reason.Reason
			break
		}
	}
	updateData := map[string]interface{}{
		"order_state":        valueobject.TakeoutOrderStateRejected,
		"rejected_by":        userUuid,
		"rejected_time":      currentTime,
		"reject_reason_code": req.ReasonCode,
		"reject_reason":      reasonText, // 取消原因描述
		"update_time":        currentTime,
	}

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
func (s *takeoutOrderSrv) CreateOrder(ctx context.Context, order *model.TakeoutOrder) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()

	// 检查订单是否已存在
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	existingOrder, err := orderRepo.GetByPlatformOrderId(order.Platform, order.PlatformOrderId)
	if err != nil {
		logger.Logger.Error("查询订单失败", zap.Error(err), zap.String("platform", order.Platform), zap.String("platformOrderId", order.PlatformOrderId))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if existingOrder != nil {
		return nil
	}

	// 验证商品数据
	if len(order.TakeoutOrderItems) == 0 {
		return errors.New("订单商品数据不能为空")
	}

	// 开启事务
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 创建订单仓储
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)
		// 1. 创建订单
		if err := orderRepoTx.Create(order); err != nil {
			logger.Logger.Error("创建订单失败", zap.Error(err), zap.Any("order", order))
			return errors.WithMessage(errors.New("创建订单失败"), err.Error())
		}

		// 2. 处理订单商品和修饰符
		hasUnmapped := false
		orderItemRepo := persistence.NewTakeoutOrderItemRepo(tx)
		modifierRepo := persistence.NewTakeoutOrderItemModifierRepo(tx)

		// Step 1: 批量处理商品信息（生成UUID、查询映射、收集查询参数）
		productIds := make([]string, 0, len(order.TakeoutOrderItems))      // 未映射商品ID
		productUuids := make([]uint64, 0, len(order.TakeoutOrderItems))    // 已映射商品UUID
		productTypes := make(map[uint64]int, len(order.TakeoutOrderItems)) // 商品类型映射

		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]

			// 生成商品 UUID
			if itemUuid, err := utils.GetID(); err != nil {
				logger.Logger.Error("生成商品UUID失败", zap.Error(err))
				return errors.WithMessage(err, "生成商品UUID失败")
			} else {
				item.Uuid = itemUuid
				item.TakeoutOrderUuid = order.Uuid
			}

			// 查询商品映射
			isMapped, productUuid, productType, err := s.findProductMapping(order.Platform, item.PlatformItemId)
			if err != nil {
				logger.Logger.Error("查询商品映射失败", zap.Error(err), zap.String("platform", order.Platform), zap.String("platformItemId", item.PlatformItemId))
				return errors.WithMessage(err, "查询商品映射失败")
			}

			item.IsMapped = isMapped
			item.TtposProductUuid = productUuid
			item.TtposProductType = productType

			// 收集查询参数
			if isMapped == 0 {
				hasUnmapped = true
				productIds = append(productIds, item.PlatformItemId)
			} else if productUuid > 0 {
				productUuids = append(productUuids, productUuid)
				productTypes[productUuid] = productType
			}
		}

		// Step 2: 批量查询商品名称（已映射商品）
		// productInfos 中包含两个名称：Name (显示用), TtposName (标识用)
		productInfos := s.menuRepo.GetProductNamesByUuids(ctx, productUuids, productTypes)
		// Step 3: 批量查询菜单名称（未映射商品）
		menuNames := s.menuRepo.GetMenuNamesByPlatformItemIds(ctx, order.Platform, productIds)

		// 处理订单商品和修饰符
		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]

			// 设置商品名称
			if item.IsMapped == 1 && item.TtposProductUuid > 0 {
				// 已映射商品：使用商品名称
				if info, ok := productInfos[item.TtposProductUuid]; ok {
					// ItemName: 显示用名称（外卖表优先）
					if info.Name != "" {
						item.ItemName = info.Name
					}
					// TtposItemName: TTPOS 核心表名称
					item.TtposItemName = info.TtposName
					// TtposItemErpCode: ERP编码
					item.TtposItemErpCode = info.TtposErpCode
				}
			} else if item.IsMapped == 0 {
				// 未映射商品：使用平台菜单名称
				if name, ok := menuNames[item.PlatformItemId]; ok && name != "" {
					item.ItemName = name
				}
			}
			// 创建订单商品
			if err := orderItemRepo.Create(item); err != nil {
				logger.Logger.Error("创建订单商品失败", zap.Error(err), zap.Any("item", item))
				return errors.WithMessage(errors.New("创建订单商品失败"), err.Error())
			}

			// 为每个修饰符生成 UUID 和设置基础信息
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]
				// 生成修饰符 UUID
				modifierUuid, err := utils.GetID()
				if err != nil {
					logger.Logger.Error("生成修饰符UUID失败", zap.Error(err))
					return errors.WithMessage(errors.New("生成修饰符UUID失败"), err.Error())
				}
				modifier.Uuid = modifierUuid
				modifier.TakeoutOrderItemUuid = item.Uuid
				modifier.CreateTime = currentTime
				modifier.UpdateTime = currentTime
				// 查询修饰符映射
				modifier.IsMapped, modifier.TtposModifierUuid, modifier.TtposModifierType, err = s.findModifierMapping(
					order.Platform,
					modifier.PlatformModifierId,
				)
				if err != nil {
					logger.Logger.Error("查询修饰符映射失败", zap.Error(err), zap.String("platform", order.Platform), zap.String("platformModifierId", modifier.PlatformModifierId))
					return errors.WithMessage(errors.New("查询修饰符映射失败"), err.Error())
				}
			}
		}

		// 批量查询修饰符名称
		modifierUuids := make([]uint64, 0)
		modifierTypes := make(map[uint64]string)
		modifierPlatformIds := make([]string, 0) // 未映射修饰符的平台ID
		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]
				if modifier.IsMapped == 1 && modifier.TtposModifierUuid > 0 {
					modifierUuids = append(modifierUuids, modifier.TtposModifierUuid)
					modifierTypes[modifier.TtposModifierUuid] = modifier.TtposModifierType
				} else if modifier.IsMapped == 0 {
					modifierPlatformIds = append(modifierPlatformIds, modifier.PlatformModifierId)
				}
			}
		}

		// 批量查询修饰符名称
		// modifierInfos 中包含两个名称：Name (显示用), TtposName (标识用)
		modifierInfos := s.menuRepo.GetModifierNamesByUuids(ctx, modifierUuids, modifierTypes)
		platformModifierNames := s.menuRepo.GetModifierNamesByPlatformIds(ctx, order.Platform, modifierPlatformIds)

		// 设置修饰符名称并创建
		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]

				// 设置修饰符名称
				if modifier.IsMapped == 1 && modifier.TtposModifierUuid > 0 {
					// 已映射修饰符：使用修饰符名称和数量
					if info, ok := modifierInfos[modifier.TtposModifierUuid]; ok && info.Name != "" {
						// ModifierName: 用于显示（commodity: 外卖表优先, 其他: 核心表）
						modifier.ModifierName = info.Name
						// TtposModifierName: 用于标识（始终来自核心表）
						modifier.TtposModifierName = info.TtposName
						// TtposModifierErpCode: ERP编码
						modifier.TtposModifierErpCode = info.TtposErpCode

						// 如果是 commodity 类型，设置规格信息和数量
						if modifier.TtposModifierType == string(valueobject.ModifierTypeCommodity) {
							modifier.TtposProductUuid = info.TtposProductUuid
							modifier.TtposFlavorUuid = info.TtposFlavorUuid
							modifier.TtposFlavorName = info.TtposFlavorName

							// 使用TTPOS数量覆盖平台数量
							if info.Num > 0 {
								modifier.Quantity = int(info.Num)
							}
						}
					}
				} else if modifier.IsMapped == 0 {
					// 未映射修饰符：使用平台菜单名称
					if name, ok := platformModifierNames[modifier.PlatformModifierId]; ok && name != "" {
						modifier.ModifierName = name
					}
				}
				// 创建修饰符
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
			campaignsPtr := make([]*model.TakeoutOrderCampaign, len(order.TakeoutOrderCampaigns))
			for i := range order.TakeoutOrderCampaigns {
				campaignsPtr[i] = &order.TakeoutOrderCampaigns[i]
			}
			if err := campaignRepo.BatchCreate(campaignsPtr); err != nil {
				logger.Logger.Error("批量创建活动信息失败", zap.Error(err))
				return errors.WithMessage(err, "批量创建活动信息失败")
			}
		}

		// 5. 检查商品映射状态
		if hasUnmapped {
			// 标记订单为异常状态
			order.IsAbnormal = 1
			order.AbnormalDetail = "订单包含未映射的商品"
			if err := orderRepoTx.Update(order); err != nil {
				logger.Logger.Error("更新订单异常状态失败", zap.Error(err))
				return errors.WithMessage(err, "更新订单异常状态失败")
			}
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

// UpdateOrderStatus 更新订单状态
func (s *takeoutOrderSrv) UpdateOrderStatus(ctx context.Context, orderUuid string, status string) error {
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
		newOrderState := grab.ConvertPlatformStateToOrderState(status, order.OrderState)
		if newOrderState != -1 {
			order.OrderState = newOrderState // 更新内部状态码
		}
		order.PlatformOrderState = status // 更新平台原始状态

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
			case valueobject.TakeoutOrderStateRejected:
				// 订单取消事件
				event.GetDispatcher().Publish(event.NewOrderCancelEvent(
					order.Uuid,
					order.Platform,
					order.PlatformOrderId,
					order.ShortOrderNumber,
					order.TakeoutOrderUuid,
					ctx.GetCompanyUuid(),
					"订单已取消",
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
// 返回: bomQuantityMap (BOM UUID -> 出库数量), bomItemMap (BOM UUID -> 订单商品，包含 modifiers), error
//
// 出库数量计算规则：
//   - 主商品：商品数量（item.Quantity）
//   - 规格(flavor)/加料(sauce): modifier.Quantity * item.Quantity (modifier数量 × 主商品数量)
//   - 套餐商品(commodity): groupItem.Num * item.Quantity (套餐配置数量 × 主商品数量)
//
// 销售数量：主商品的 item.Quantity（用于统计销售）
func (s *takeoutOrderSrv) BuildBomQuantityMap(ctx context.Context, order *model.TakeoutOrder) (map[uint64]int, map[uint64]*model.TakeoutOrderItem, error) {
	db := ctx.GetDB()
	bomMappingRepo := persistence.NewTakeoutBomMappingRepo(db)

	// 收集 modifier 信息的辅助结构
	type modifierInfo struct {
		modifier         *model.TakeoutOrderItemModifier
		item             *model.TakeoutOrderItem // 主商品
		outboundQuantity int                     // 出库数量
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
		quantity  int                     // 出库数量
		modifiers []*modifierInfo         // 关联的 modifiers
		item      *model.TakeoutOrderItem // 主商品（取第一个 modifier 的 item）
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
	bomItemMap := make(map[uint64]*model.TakeoutOrderItem, len(bomMap))

	for bomUuid, data := range bomMap {
		bomQuantityMap[bomUuid] = data.quantity

		// 创建 item 副本，包含该 BOM 对应的所有 modifiers
		itemCopy := *data.item
		itemCopy.TakeoutOrderItemModifiers = make([]model.TakeoutOrderItemModifier, len(data.modifiers))
		for i, info := range data.modifiers {
			itemCopy.TakeoutOrderItemModifiers[i] = *info.modifier
		}
		bomItemMap[bomUuid] = &itemCopy
	}

	return bomQuantityMap, bomItemMap, nil
}

// CalculateTakeoutOrderSalesVolume 计算外卖订单销量
// 返回: productBoms (BOM UUID -> 销量), productPackages (Package UUID -> 销量)
func (s *takeoutOrderSrv) CalculateTakeoutOrderSalesVolume(order *model.TakeoutOrder) (map[uint64]float64, map[uint64]float64, error) {
	productBoms := make(map[uint64]float64)     // 规格商品销量 map[BOM UUID]销量
	productPackages := make(map[uint64]float64) // 套餐商品销量 map[Package UUID]销量

	// 遍历订单商品
	for _, item := range order.TakeoutOrderItems {
		// 只处理已映射的商品
		if item.IsMapped != 1 || item.TtposProductUuid == 0 {
			continue
		}

		itemQuantity := float64(item.Quantity)

		// 统计主商品的 Package 销量
		// 不管是套餐还是普通商品，TtposProductUuid 都是 ProductPackage 的 UUID
		productPackages[item.TtposProductUuid] += itemQuantity

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
func (s *takeoutOrderSrv) GetOrderForPrint(ctx context.Context, orderUuid uint64) (*model.TakeoutOrder, error) {
	db := ctx.GetDB()

	// 查询订单（包含商品、修饰符等信息）
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(
		orderUuid,
		orderRepo.WithTakeoutOrderItems(),
		orderRepo.WithTakeoutOrderItemModifiers(),
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

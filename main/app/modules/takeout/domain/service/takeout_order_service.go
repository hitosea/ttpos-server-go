package service

import (
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
	"ttpos-server-go/app/repository"
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

	// 订单操作
	AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error
	RejectOrder(ctx context.Context, req *request.TakeoutOrderRejectReq) error
	CallRider(ctx context.Context, req *request.TakeoutOrderCallRiderReq) error

	// 创建订单（接受已转换的订单对象，商品数据从 order.RawData 中解析）- 由 Application 层调用
	CreateOrder(ctx context.Context, order *model.TakeoutOrder) error

	// 更新订单状态 - 由 Application 层调用
	UpdateOrderStatus(ctx context.Context, orderUuid string, newStatus string) error
}

// takeoutOrderSrv 外卖订单服务实现
type takeoutOrderSrv struct {
	dbm      *database.DBManager
	menuRepo menuRepo.IMenuDataRepository
}

// NewTakeoutOrderSrv 创建外卖订单服务
func NewTakeoutOrderSrv(dbm *database.DBManager) ITakeoutOrderSrv {
	return &takeoutOrderSrv{
		dbm:      dbm,
		menuRepo: persistence.NewMenuDataRepository(dbm),
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
		logger.Logger.Error("查询订单列表失败",
			zap.Error(err),
			zap.Any("options", options))
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
			Subtotal:         order.Subtotal / 100,
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
			Uuid:           item.Uuid,           // 商品UUID
			Specifications: item.Specifications, // 规格说明
			Quantity:       item.Quantity,       // 数量
			Price:          item.Price,          // 价格
			Tax:            item.Tax,            // 税费
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
							Price:    modifier.Price,
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
			DeductedAmount: campaign.DeductedAmount / 100,
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
			PromoAmount:      promo.PromoAmount / 100,
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
			SubmitTime:         order.SubmitTime,
			AcceptedTime:       order.AcceptedTime,
			CompletedTime:      order.CompletedTime,
			EstimatedReadyTime: order.EstimatedReadyTime,
			MaxReadyTime:       order.MaxReadyTime,
		},
		Price: response.TakeoutOrderPriceResp{
			Subtotal:          order.Subtotal,
			DeliveryFee:       order.DeliveryFee,
			SmallOrderFee:     order.SmallOrderFee,
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
		logger.Logger.Error("查询订单失败",
			zap.Error(err),
			zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单状态
	if order.OrderState != valueobject.TakeoutOrderStatePending {
		return errors.New("订单状态不正确")
	}

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

	// 更新订单状态
	updateData := map[string]interface{}{
		"order_state":         valueobject.TakeoutOrderStateAccepted,
		"accepted_time":       currentTime,
		"accepted_by":         userUuid,
		"order_accepted_type": valueobject.TakeoutOrderAcceptedTypeManual,
		"update_time":         currentTime,
	}

	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		logger.Logger.Error("更新订单状态失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
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
	currentTime := time.Now().Unix()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid)
	if err != nil {
		logger.Logger.Error("查询订单失败",
			zap.Error(err),
			zap.Uint64("uuid", req.Uuid))
		return errors.WithMessage(errors.New("查询订单失败"), err.Error())
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单状态 - 只有已接单配餐中的订单才能呼叫骑手
	if order.OrderState != valueobject.TakeoutOrderStateAccepted {
		return errors.New("订单状态不正确，只有已接单配餐中的订单才能呼叫骑手")
	}

	// 调用 BMP RPC 标记订单准备完成
	rpcClient, err := rpc.NewBMPTakeoutClient()
	if err != nil {
		logger.Logger.Error("创建 BMP RPC 客户端失败",
			zap.Error(err),
			zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("创建 BMP RPC 客户端失败"), err.Error())
	}
	defer rpcClient.Close()

	// 调用 MarkOrderReady 接口（标记订单准备完成，通知平台派送骑手）
	if err := rpcClient.MarkOrderReady(ctx.GetContext(), order.TakeoutOrderUuid); err != nil {
		logger.Logger.Error("调用 BMP MarkOrderReady 接口失败",
			zap.Error(err),
			zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("呼叫骑手失败"), err.Error())
	}

	// 更新订单状态为待骑手接单
	updateData := map[string]interface{}{
		"order_state": valueobject.TakeoutOrderStateRiderPending,
		"update_time": currentTime,
	}

	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		logger.Logger.Error("更新订单状态失败",
			zap.Error(err),
			zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("更新订单状态失败"), err.Error())
	}

	// 发布订单呼叫骑手事件
	event.GetDispatcher().Publish(event.NewOrderReadyEvent(
		order.Uuid,
		order.Platform,
		order.PlatformOrderId,
		order.ShortOrderNumber,
		order.TakeoutOrderUuid,
		ctx.GetCompanyUuid(),
	))

	logger.Logger.Info("呼叫骑手成功，已发布事件",
		zap.Uint64("orderUuid", order.Uuid),
		zap.String("platform", order.Platform),
		zap.String("shortOrderNumber", order.ShortOrderNumber))

	return nil
}

// findProductMapping 查找商品映射关系
// 返回: isMapped, ttposProductUuid, ttposSkuUuid, error
func (s *takeoutOrderSrv) findProductMapping(platform, platformItemId string) (int, uint64, uint64, int, error) {
	// 解析商品ID，提取 TTPOS UUID
	result, err := valueobject.ParsePlatformID(platform, platformItemId)
	if err != nil {
		// ID格式错误，标记为未映射
		return 0, 0, 0, 0, nil
	}

	if !result.IsMapped {
		return 0, 0, 0, 0, nil
	}

	switch result.IDType {
	case valueobject.IDTypePackage:
		return 1, result.UUID, 0, 1, nil
	default:
		return 1, result.UUID, 0, 0, nil
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
		modifierType = "flavor" // 规格
	case valueobject.IDTypeSauce:
		modifierType = "sauce" // 加料
	case valueobject.IDTypeAttr:
		modifierType = "attr" // 属性
	case valueobject.IDTypePackageItem:
		modifierType = "commodity" // 套餐商品
	default:
		// 不是修饰符类型，标记为未映射
		return 0, 0, "", nil
	}

	return 1, result.UUID, modifierType, nil
}

// checkStock 检查商品库存
func (s *takeoutOrderSrv) checkStock(ctx context.Context, items []*model.TakeoutOrderItem) (int, error) {
	// TODO: 实现库存检查逻辑
	// 1. 查询 TTPOS 商品表获取库存信息
	// 2. 对比订单数量和可用库存
	// 3. 返回库存状态

	// 当前简化实现：假设库存充足
	return valueobject.TakeoutStockStatusSufficient, nil
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
		return errors.New("订单已存在")
	}

	// 验证商品数据
	if len(order.TakeoutOrderItems) == 0 {
		return errors.New("订单商品数据不能为空")
	}

	// 开启事务
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)
		// 1. 创建订单
		if order.OrderAcceptedType == valueobject.TakeoutOrderAcceptedTypeAuto {
			order.OrderState = valueobject.TakeoutOrderStateAccepted
		}
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
			isMapped, productUuid, skuUuid, productType, err := s.findProductMapping(order.Platform, item.PlatformItemId)
			if err != nil {
				logger.Logger.Error("查询商品映射失败", zap.Error(err), zap.String("platform", order.Platform), zap.String("platformItemId", item.PlatformItemId))
				return errors.WithMessage(err, "查询商品映射失败")
			}

			item.IsMapped = isMapped
			item.TtposProductUuid = productUuid
			item.TtposSkuUuid = skuUuid
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
		productNames := s.menuRepo.GetProductNamesByUuids(ctx, productUuids, productTypes)
		// Step 3: 批量查询菜单名称（未映射商品）
		menuNames := s.menuRepo.GetMenuNamesByPlatformItemIds(ctx, order.Platform, productIds)

		// 处理订单商品和修饰符
		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]

			// 设置商品名称
			if item.IsMapped == 1 && item.TtposProductUuid > 0 {
				// 已映射商品：使用 TTPOS 商品名称
				if name, ok := productNames[item.TtposProductUuid]; ok && name != "" {
					item.ItemName = name
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

		// 批量查询未映射修饰符的菜单名称（从 ttpos_takeout 表）
		modifierNames := s.menuRepo.GetModifierNamesByUuids(ctx, modifierUuids, modifierTypes)
		platformModifierNames := s.menuRepo.GetModifierNamesByPlatformIds(ctx, order.Platform, modifierPlatformIds)

		// 设置修饰符名称并创建
		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]
				// 设置修饰符名称
				if modifier.IsMapped == 1 && modifier.TtposModifierUuid > 0 {
					// 已映射修饰符：使用 TTPOS 修饰符名称
					if name, ok := modifierNames[modifier.TtposModifierUuid]; ok && name != "" {
						modifier.ModifierName = name
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

	// TODO: 待完善
	// 如果订单是自动接单，则发送 WebSocket 通知到前端
	if order.OrderAcceptedType == valueobject.TakeoutOrderAcceptedTypeAuto {
		// websocket.SendMessage(websocket.MessageTypeOrderAccepted, order.Uuid)
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
		newOrderState := grab.ConvertPlatformStateToOrderState(status)
		if newOrderState == -1 {
			logger.Logger.Error("转换订单状态失败", zap.String("order_uuid", orderUuid), zap.String("status", status))
			return errors.New("转换订单状态失败")
		}
		order.PlatformOrderState = status // 更新平台原始状态
		order.OrderState = newOrderState  // 更新内部状态码

		// 4. 更新订单到数据库
		if err := orderRepoTx.Update(order); err != nil {
			logger.Logger.Error("更新订单数据失败", zap.String("order_uuid", orderUuid), zap.Error(err))
			return errors.WithMessage(err, "更新订单数据失败")
		}

		return nil
	})
}

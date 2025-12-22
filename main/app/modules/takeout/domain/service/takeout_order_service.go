package service

import (
	"encoding/json"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/app/modules/takeout/interfaces/response"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
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

	// 创建订单（接受已转换的订单对象，商品数据从 order.RawData 中解析）- 由 Application 层调用
	CreateOrder(ctx context.Context, order *model.TakeoutOrder) error

	// 更新订单状态 - 由 Application 层调用
	UpdateOrderStatus(ctx context.Context, orderUuid string, newStatus string, rawData map[string]interface{}) error
}

// takeoutOrderSrv 外卖订单服务实现
type takeoutOrderSrv struct {
	dbm *database.DBManager
}

// NewTakeoutOrderSrv 创建外卖订单服务
func NewTakeoutOrderSrv(dbm *database.DBManager) ITakeoutOrderSrv {
	return &takeoutOrderSrv{
		dbm: dbm,
	}
}

// GetList 获取订单列表
func (s *takeoutOrderSrv) GetList(ctx context.Context, req *request.TakeoutOrderListReq) (*response.TakeoutOrderListResp, error) {
	db := ctx.GetDB()

	// 创建 Repository
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	itemRepo := persistence.NewTakeoutOrderItemRepo(db)

	// 构建查询选项
	options := []persistence.DBOption{
		orderRepo.WherePlatform(req.Platform),
		orderRepo.WhereOrderState(req.Status),
		orderRepo.WhereTimeRange(req.StartTime, req.EndTime),
		orderRepo.WhereSearch(req.Search),
		orderRepo.Limit(req.PageSize),
		orderRepo.Offset((req.PageNo - 1) * req.PageSize),
	}

	// 查询订单列表
	orders, total, err := orderRepo.GetList(options...)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单列表失败")
	}

	// 构建响应
	list := make([]*response.TakeoutOrderResp, 0, len(orders))
	for _, order := range orders {
		// 查询订单商品
		items, err := itemRepo.GetByOrderUuid(order.Uuid)
		if err != nil {
			return nil, errors.WithMessage(err, "查询订单商品失败")
		}

		orderResp := s.buildOrderResp(order, items)
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
	itemRepo := persistence.NewTakeoutOrderItemRepo(db)

	// 查询订单
	order, err := orderRepo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}

	// 查询订单商品
	items, err := itemRepo.GetByOrderUuid(order.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "查询订单商品失败")
	}

	return s.buildOrderResp(order, items), nil
}

// AcceptOrder 接单
func (s *takeoutOrderSrv) AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	userUuid := ctx.GetStaffUuid()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.OrderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单状态
	if order.OrderState != valueobject.TakeoutOrderStatePending {
		return errors.New("订单状态不正确")
	}

	// TODO: 调用 RPC 通知平台
	// 这里需要调用 ttpos-bmp 的 gRPC 接口通知平台

	// 更新订单状态
	updateData := map[string]interface{}{
		"order_state":         valueobject.TakeoutOrderStateAccepted,
		"accepted_time":       currentTime,
		"accepted_by":         userUuid,
		"order_accepted_type": valueobject.TakeoutOrderAcceptedTypeManual,
		"update_time":         currentTime,
	}

	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		return errors.WithMessage(err, "更新订单状态失败")
	}

	// TODO: 发送通知到 KDS（厨显）
	// TODO: 发送 WebSocket 通知到前端

	return nil
}

// RejectOrder 拒单
func (s *takeoutOrderSrv) RejectOrder(ctx context.Context, req *request.TakeoutOrderRejectReq) error {
	db := ctx.GetDB()
	currentTime := time.Now().Unix()
	userUuid := ctx.GetStaffUuid()

	// 查询订单
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.OrderUuid)
	if err != nil {
		return errors.WithMessage(err, "查询订单失败")
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	// 检查订单状态
	if order.OrderState != valueobject.TakeoutOrderStatePending {
		return errors.New("订单状态不正确")
	}

	// TODO: 调用 RPC 通知平台
	// 这里需要调用 ttpos-bmp 的 gRPC 接口通知平台拒单

	// 更新订单状态
	updateData := map[string]interface{}{
		"order_state":        valueobject.TakeoutOrderStateRejected,
		"rejected_time":      currentTime,
		"rejected_by":        userUuid,
		"reject_reason_code": req.RejectReasonCode,
		"update_time":        currentTime,
	}

	if err := orderRepo.UpdateByMap(order.Uuid, updateData); err != nil {
		return errors.WithMessage(err, "更新订单状态失败")
	}

	// TODO: 发送 WebSocket 通知到前端

	return nil
}

// buildOrderResp 构建订单响应
func (s *takeoutOrderSrv) buildOrderResp(order *model.TakeoutOrder, items []*model.TakeoutOrderItem) *response.TakeoutOrderResp {
	// 构建商品列表
	itemList := make([]*response.TakeoutOrderItemResp, 0, len(items))
	for _, item := range items {
		itemResp := &response.TakeoutOrderItemResp{
			Uuid:             item.Uuid,
			PlatformItemId:   item.PlatformItemId,
			PlatformItemName: item.PlatformItemName,
			Quantity:         item.Quantity,
			Price:            item.Price,
			Tax:              item.Tax,
			Specifications:   item.Specifications,
			IsMapped:         item.IsMapped,
		}
		itemList = append(itemList, itemResp)
	}

	return &response.TakeoutOrderResp{
		Uuid:             order.Uuid,
		Platform:         order.Platform,
		PlatformOrderId:  order.PlatformOrderId,
		ShortOrderNumber: order.ShortOrderNumber,
		OrderState:       order.OrderState,
		IsAbnormal:       order.IsAbnormal,
		AbnormalDetail:   order.AbnormalDetail,
		StockStatus:      order.StockStatus,
		Subtotal:         order.Subtotal,
		DeliveryFee:      order.DeliveryFee,
		TotalAmount:      order.TotalAmount,
		CurrencyCode:     order.CurrencyCode,
		CurrencySymbol:   order.CurrencySymbol,
		PaymentType:      order.PaymentType,
		OrderTime:        order.OrderTime,
		AcceptedTime:     order.AcceptedTime,
		Cutlery:          order.Cutlery,
		OrderType:        order.OrderType,
		Items:            itemList,
	}
}

// findProductMapping 查找商品映射关系
// 返回: isMapped, ttposProductUuid, ttposSkuUuid, error
func (s *takeoutOrderSrv) findProductMapping(platform, platformItemId string) (int, uint64, uint64, error) {
	// 解析商品ID，提取 TTPOS UUID
	result, err := valueobject.ParsePlatformID(platform, platformItemId)
	if err != nil {
		// ID格式错误，标记为未映射
		return 0, 0, 0, nil
	}

	if !result.IsMapped {
		return 0, 0, 0, nil
	}

	// TODO: 如果需要，可以从数据库查询 SKU UUID
	// 目前简化实现，只返回 Product UUID
	return 1, result.UUID, 0, nil
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
		return errors.WithMessage(err, "查询订单失败")
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
		// 1. 创建订单
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)
		if err := orderRepoTx.Create(order); err != nil {
			return errors.WithMessage(err, "创建订单失败")
		}

		// 2. 处理订单商品和修饰符
		hasUnmapped := false
		orderItemRepo := persistence.NewTakeoutOrderItemRepo(tx)
		modifierRepo := persistence.NewTakeoutOrderItemModifierRepo(tx)

		for i := range order.TakeoutOrderItems {
			item := &order.TakeoutOrderItems[i]

			// 生成商品 UUID
			itemUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成商品UUID失败")
			}
			item.Uuid = itemUuid
			item.TakeoutOrderUuid = order.Uuid

			// 查询商品映射
			item.IsMapped, item.TtposProductUuid, item.TtposSkuUuid, err = s.findProductMapping(
				order.Platform,
				item.PlatformItemId,
			)
			if err != nil {
				return errors.WithMessage(err, "查询商品映射失败")
			}

			// 检查是否有未映射商品
			if item.IsMapped == 0 {
				hasUnmapped = true
			}

			// 创建订单商品
			if err := orderItemRepo.Create(item); err != nil {
				return errors.WithMessage(err, "创建订单商品失败")
			}

			// 处理修饰符
			for j := range item.TakeoutOrderItemModifiers {
				modifier := &item.TakeoutOrderItemModifiers[j]

				// 生成修饰符 UUID
				modifierUuid, err := utils.GetID()
				if err != nil {
					return errors.WithMessage(err, "生成修饰符UUID失败")
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
					return errors.WithMessage(err, "查询修饰符映射失败")
				}

				// 创建修饰符
				if err := modifierRepo.Create(modifier); err != nil {
					return errors.WithMessage(err, "创建商品修饰符失败")
				}
			}
		}

		// 3. 检查商品映射状态
		if hasUnmapped {
			// 标记订单为异常状态
			order.IsAbnormal = 1
			order.AbnormalDetail = "订单包含未映射的商品"
			if err := orderRepoTx.Update(order); err != nil {
				return errors.WithMessage(err, "更新订单异常状态失败")
			}
		}

		// 4. 检查库存（仅检查已映射的商品）
		if !hasUnmapped {
			// 提取已创建的商品列表用于库存检查
			createdItems := make([]*model.TakeoutOrderItem, len(order.TakeoutOrderItems))
			for i := range order.TakeoutOrderItems {
				createdItems[i] = &order.TakeoutOrderItems[i]
			}

			stockStatus, err := s.checkStock(ctx, createdItems)
			if err != nil {
				return errors.WithMessage(err, "检查库存失败")
			}

			if stockStatus != valueobject.TakeoutStockStatusSufficient {
				order.StockStatus = stockStatus
				if err := orderRepoTx.Update(order); err != nil {
					return errors.WithMessage(err, "更新库存状态失败")
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	// TODO: 检查是否自动接单
	// TODO: 发送 WebSocket 通知到前端

	return nil
}

// UpdateOrderStatus 更新订单状态
func (s *takeoutOrderSrv) UpdateOrderStatus(ctx context.Context, orderUuid string, newStatus string, rawData map[string]interface{}) error {
	db := ctx.GetDB()

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		orderRepoTx := persistence.NewTakeoutOrderRepo(tx)

		// 1. 查询订单（通过 takeout_order_uuid 字符串查询）
		order, err := orderRepoTx.GetByTakeoutOrderUuid(orderUuid)
		if err != nil {
			logger.Logger.Error("查询订单失败",
				zap.String("order_uuid", orderUuid),
				zap.Error(err))
			return errors.WithMessage(err, "查询订单失败")
		}

		if order == nil {
			logger.Logger.Error("订单不存在", zap.String("order_uuid", orderUuid))
			return errors.New("订单不存在")
		}

		// 2. 记录旧状态（保存原始的平台状态字符串用于日志）
		oldStatusStr := newStatus // 这里应该从 rawData 或其他地方获取旧状态，暂时占位

		// 3. 更新订单原始数据（保持 RawData 字段更新）
		if rawData != nil {
			rawDataJSON, err := json.Marshal(rawData)
			if err != nil {
				logger.Logger.Error("序列化订单数据失败", zap.Error(err))
				return errors.WithMessage(err, "序列化订单数据失败")
			}
			order.RawData = string(rawDataJSON)
		}

		// 4. 更新订单（这里只更新 RawData，状态映射应该由平台转换器处理）
		// 注意：OrderState 是 int 类型，需要由调用方传入映射后的状态码
		// 这里暂时只更新 RawData，状态转换留给后续优化
		if err := orderRepoTx.Update(order); err != nil {
			logger.Logger.Error("更新订单数据失败",
				zap.String("order_uuid", orderUuid),
				zap.Error(err))
			return errors.WithMessage(err, "更新订单数据失败")
		}

		logger.Logger.Info("订单状态已更新",
			zap.String("order_uuid", orderUuid),
			zap.String("platform_status", newStatus),
			zap.String("old_status", oldStatusStr))

		return nil
	})
}

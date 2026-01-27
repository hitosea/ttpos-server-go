package application

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	"ttpos-server-go/app/modules/takeout/domain/event"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/lineman"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	grabfood "github.com/grab/grabfood-api-sdk-go"
	"go.uber.org/zap"
)

// ITakeoutOrderAppService 外卖订单应用服务接口
type ITakeoutOrderAppService interface {
	// 处理订单状态变更
	HandlePushOrderState(ctx context.Context, takeoutOrderEvent request.TakeoutOrderEvent) error
	// 订单同步（从 RPC 接收新订单）
	SyncNewOrder(ctx context.Context, platform string, takeoutOrderUuid string, rawData map[string]interface{}, orderDataMap map[string]interface{}) error
	// 更新订单（从 RPC 接收订单更新）
	UpdateOrder(ctx context.Context, orderUuid uint64, platform string, rawData map[string]interface{}, orderDataMap map[string]interface{}) error
	// 接单
	AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error
	// 检查订单库存
	CheckOrderStock(ctx context.Context, order *model.TakeoutOrder) (error, []string)
}

// takeoutOrderAppService 外卖订单应用服务实现
type takeoutOrderAppService struct {
	// RPC 调用相关
	rpcService *rpcAdapter.TakeoutRPCService
	// 数据库管理器
	dbm *database.DBManager
	// 平台转换器映射（用于菜单和订单）
	converters map[string]service.IOrderConverter
	// 订单服务
	orderService service.ITakeoutOrderSrv
	// 系统锁
	systemLock lock.Lock
}

// NewTakeoutOrderAppService 创建外卖订单应用服务
func NewTakeoutOrderAppService(
	dbm *database.DBManager,
) ITakeoutOrderAppService {
	// 初始化 RPC 服务
	rpcService := rpcAdapter.NewTakeoutRPCService()

	// 初始化平台转换器
	converters := make(map[string]service.IOrderConverter)
	converters[value_object.TakeoutPlatformGrab] = grab.NewGrabConverter(dbm)
	converters[value_object.TakeoutPlatformLineman] = lineman.NewLineManConverter()

	return &takeoutOrderAppService{
		rpcService:   rpcService,
		dbm:          dbm,
		converters:   converters,
		orderService: service.NewTakeoutOrderSrv(dbm),
		systemLock:   lock.NewSystemLock(),
	}
}

// HandlePushOrderState 处理订单状态变更
func (s *takeoutOrderAppService) HandlePushOrderState(ctx context.Context, takeoutOrderEvent request.TakeoutOrderEvent) error {
	// 并发控制：使用分布式锁防止同一订单被并发处理
	lockKey := fmt.Sprintf("takeout_order:%s:%s", takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid)
	s.systemLock.LockUuidString(lockKey)
	defer s.systemLock.UnlockUuidString(lockKey)

	// 判断是否开发模式，如果是开发模式，则使用 6293997752320000
	if config.Server.Mode == "debug" && takeoutOrderEvent.ProviderName == value_object.TakeoutPlatformLineman {
		if config.Takeout.TakeoutLinemanStoreId != 0 {
			takeoutOrderEvent.ShopUuid = strconv.FormatUint(config.Takeout.TakeoutLinemanStoreId, 10)
		}
	}

	// 根据 Action 处理订单
	switch takeoutOrderEvent.Action {
	case "create":
		// 通过 RPC 查询订单信息
		orderInfo, orderDataMap, err := s.rpcService.GetOrderInfo(ctx, takeoutOrderEvent.ShopUuid, takeoutOrderEvent.OrderUuid)
		if err != nil {
			logger.Logger.Error("RPC查询订单失败", zap.Error(err))
			return fmt.Errorf("RPC查询订单失败: %w", err)
		}
		return s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo, orderDataMap)
	case "update":
		// 通过 RPC 查询订单信息
		orderInfo, orderDataMap, err := s.rpcService.GetOrderInfo(ctx, takeoutOrderEvent.ShopUuid, takeoutOrderEvent.OrderUuid)
		if err != nil {
			logger.Logger.Error("RPC查询订单失败", zap.Error(err))
			return fmt.Errorf("RPC查询订单失败: %w", err)
		}
		// 先查询订单是否存在
		orderRepo := persistence.NewTakeoutOrderRepo(ctx.GetDB())
		existingOrder, err := orderRepo.GetByTakeoutOrderUuid(takeoutOrderEvent.OrderUuid)
		if err != nil || existingOrder == nil {
			return s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo, orderDataMap)
		}
		// 订单已存在，更新订单信息
		return s.UpdateOrder(ctx, existingOrder.Uuid, takeoutOrderEvent.ProviderName, orderInfo, orderDataMap)
	// 订单状态变更
	case "status_update", "cancel":
		// 查询订单是否存在
		order, err := persistence.NewTakeoutOrderRepo(ctx.GetDB()).GetByTakeoutOrderUuid(takeoutOrderEvent.OrderUuid)
		if err != nil || order == nil {
			// 订单不存在，通过 RPC 查询订单信息
			orderInfo, orderDataMap, err := s.rpcService.GetOrderInfo(ctx, takeoutOrderEvent.ShopUuid, takeoutOrderEvent.OrderUuid)
			if err != nil {
				logger.Logger.Error("RPC查询订单失败", zap.Error(err))
				return fmt.Errorf("RPC查询订单失败: %w", err)
			}
			if err := s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo, orderDataMap); err != nil {
				logger.Logger.Error("同步订单失败", zap.Error(err))
				return fmt.Errorf("同步订单失败: %w", err)
			}
		}
		// 更新订单状态
		return s.orderService.UpdateOrderStatus(ctx, takeoutOrderEvent.OrderUuid, takeoutOrderEvent.Status, takeoutOrderEvent.Message)
	}
	return nil
}

// SyncNewOrder 同步新订单（从 RPC 接收）
func (s *takeoutOrderAppService) SyncNewOrder(ctx context.Context, platform string, takeoutOrderUuid string, rawData map[string]interface{}, orderDataMap map[string]interface{}) error {
	// 1. 解析订单 Webhook 数据
	rawDataJSON, webhookInterface, platformOrderId, err := s.parseOrderWebhookData(platform, rawData, orderDataMap)
	if err != nil {
		return err
	}

	// 2. 生成订单 UUID
	orderUuid, err := utils.GetID()
	if err != nil {
		logger.Logger.Error("生成订单UUID失败", zap.Error(err))
		return fmt.Errorf("生成订单UUID失败: %w", err)
	}

	// 3. 获取平台转换器
	converter, ok := s.converters[platform]
	if !ok {
		return fmt.Errorf("不支持的平台: %s", platform)
	}

	// 4. 使用转换器将平台订单转换为通用订单格式（包括商品数据）
	order, err := converter.ConvertOrderToTakeoutOrder(
		orderUuid,
		platform,
		platformOrderId,
		webhookInterface, // 传入 interface{}，由转换器内部处理
		rawDataJSON,
		time.Now().Unix(),
	)
	if err != nil {
		logger.Logger.Error("转换订单数据失败", zap.Error(err))
		return fmt.Errorf("转换订单数据失败: %w", err)
	}

	// 5. 验证商品数据
	if len(order.TakeoutOrderItems) == 0 {
		logger.Logger.Error("订单商品数据不存在或格式错误")
		return fmt.Errorf("订单商品数据不存在或格式错误")
	}

	// 设置订单UUID
	order.TakeoutOrderUuid = takeoutOrderUuid

	// 6. 调用 Domain Service 创建订单
	return s.orderService.CreateOrder(ctx, order)
}

// AcceptOrder 接单
func (s *takeoutOrderAppService) AcceptOrder(ctx context.Context, req *request.TakeoutOrderAcceptReq) error {
	// 1. 获取订单信息（包含商品列表）
	db := ctx.GetDB()
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	order, err := orderRepo.GetByUuid(req.Uuid, orderRepo.WithTakeoutOrderItems(), orderRepo.WithTakeoutOrderItemModifiers())
	if err != nil {
		return err
	}
	// 2. 检查订单状态
	if err := order.IsPendingOrder(); err != nil {
		return err
	}
	// 检查订单状态
	if order.IsAbnormalOrder() {
		return errors.New("订单异常，不能接单")
	}
	// 3. 验证库存
	if err, outOfStockNames := s.CheckOrderStock(ctx, order); err != nil {
		return errors.NewWithCodeAndData(constant.CodeOrderCheckProductStockZero, outOfStockNames, err.Error())
	}
	// 4. 调用 Domain Service 执行接单逻辑
	if err := s.orderService.AcceptOrder(ctx, req); err != nil {
		return err
	}
	return nil
}

// checkOrderStock 检查订单商品库存
func (s *takeoutOrderAppService) CheckOrderStock(ctx context.Context, order *model.TakeoutOrder) (error, []string) {
	// ==================== Step 1: 构建 BOM 数量映射 ====================
	bomQuantityMap, bomItemMap, err := s.orderService.BuildBomQuantityMap(ctx, order)
	if err != nil {
		return errors.NewWithCodeAndData(constant.CodeOrderCheckProductStockZero, []string{}, "构建BOM数量映射失败"), nil
	}
	if len(bomQuantityMap) == 0 {
		return nil, nil // 订单没有配方商品，无需处理原料
	}

	// ==================== Step 2: 调用库存模块检查库存 ====================
	inventoryAppSrv := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	insufficientBomUuids, err := inventoryAppSrv.CheckStock(ctx, bomQuantityMap)
	if err != nil {
		logger.Logger.Error("检查库存失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("检查库存失败"), err.Error()), nil
	}

	// ==================== Step 3: 如果有库存不足的商品，构建提示信息并返回错误 ====================
	if len(insufficientBomUuids) > 0 {
		outOfStockNamesMap := make(map[string]struct{})
		for _, bomUuid := range insufficientBomUuids {
			if item, ok := bomItemMap[bomUuid]; ok {
				for i := range item.TakeoutOrderItemModifiers {
					modifier := &item.TakeoutOrderItemModifiers[i]
					var name string

					if modifier.IsCommodity() {
						commodityName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						flavorName := language.JsonToLocaleResponse(modifier.TtposFlavorName).GetLocale(ctx.GetLanguage())
						if flavorName != "" {
							name = fmt.Sprintf("%s (%s)", commodityName, flavorName)
						} else {
							name = commodityName
						}
					} else if modifier.IsFlavor() {
						itemName := language.JsonToLocaleResponse(item.ItemName).GetLocale(ctx.GetLanguage())
						flavorName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						name = fmt.Sprintf("%s (%s)", itemName, flavorName)
					} else if modifier.IsSauce() {
						// sauce 类型：显示主商品名称(加料)
						itemName := language.JsonToLocaleResponse(item.ItemName).GetLocale(ctx.GetLanguage())
						sauceName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						name = fmt.Sprintf("%s (%s)", itemName, sauceName)
					}
					if name != "" {
						outOfStockNamesMap[name] = struct{}{}
					}
				}
			}
		}
		// 转为切片
		if len(outOfStockNamesMap) > 0 {
			outOfStockNames := make([]string, 0, len(outOfStockNamesMap))
			for name := range outOfStockNamesMap {
				outOfStockNames = append(outOfStockNames, name)
			}
			//
			outOfStockMsg := "以下商品库存不足: " + strings.Join(outOfStockNames, ", ")
			return errors.New(outOfStockMsg), outOfStockNames
		}
	}

	return nil, nil
}

// parseOrderWebhookData 解析订单 Webhook 数据（通用方法）
// 返回: rawDataJSON, webhookInterface, platformOrderId, error
func (s *takeoutOrderAppService) parseOrderWebhookData(
	platform string,
	rawData map[string]interface{},
	orderDataMap map[string]interface{},
) ([]byte, interface{}, string, error) {
	// 1. 获取平台转换器
	converter, ok := s.converters[platform]
	if !ok {
		logger.Logger.Error("不支持的平台", zap.String("platform", platform))
		return nil, nil, "", fmt.Errorf("不支持的平台: %s", platform)
	}

	// 2. 将原始数据转为 JSON
	var rawDataJSON []byte
	var err error

	switch platform {
	case value_object.TakeoutPlatformGrab:
		rawDataJSON, err = json.Marshal(rawData)
		if err != nil {
			logger.Logger.Error("序列化订单数据失败", zap.Error(err))
			return nil, nil, "", fmt.Errorf("序列化订单数据失败: %w", err)
		}
	case value_object.TakeoutPlatformLineman:
		// 将 orderData 转为 JSON
		rawDataJSON, err = json.Marshal(orderDataMap)
		if err != nil {
			logger.Logger.Error("序列化订单数据失败", zap.Error(err))
			return nil, nil, "", fmt.Errorf("序列化订单数据失败: %w", err)
		}
	default:
		logger.Logger.Error("不支持的平台", zap.String("platform", platform))
		return nil, nil, "", fmt.Errorf("不支持的平台: %s", platform)
	}

	// 3. 使用转换器解析 Webhook
	webhookInterface, err := converter.ParseOrderWebhook(rawDataJSON)
	if err != nil {
		logger.Logger.Error("解析订单 Webhook 失败", zap.Error(err))
		return nil, nil, "", fmt.Errorf("解析订单 Webhook 失败: %w", err)
	}

	// 4. 获取平台订单 ID
	var platformOrderId string
	switch platform {
	case value_object.TakeoutPlatformGrab:
		// Grab 平台：从 SubmitOrderRequest 获取
		if submitOrderReq, ok := webhookInterface.(*grabfood.SubmitOrderRequest); ok {
			platformOrderId = submitOrderReq.GetOrderID()
		}
	case value_object.TakeoutPlatformLineman:
		// Lineman 平台：从 rawData 获取 orderId
		if orderIdVal, ok := rawData["orderId"].(string); ok {
			platformOrderId = orderIdVal
		} else if orderIdVal, ok := rawData["orderID"].(string); ok {
			platformOrderId = orderIdVal
		}
	default:
		logger.Logger.Error("不支持的平台", zap.String("platform", platform))
		return nil, nil, "", fmt.Errorf("不支持的平台: %s", platform)
	}

	if platformOrderId == "" {
		logger.Logger.Error("无法获取平台订单ID",
			zap.String("platform", platform),
			zap.Any("rawData", rawData))
		return nil, nil, "", fmt.Errorf("无法获取平台订单ID")
	}

	return rawDataJSON, webhookInterface, platformOrderId, nil
}

// UpdateOrder 更新订单信息（从 RPC 接收）
func (s *takeoutOrderAppService) UpdateOrder(ctx context.Context, orderUuid uint64, platform string, rawData map[string]interface{}, orderDataMap map[string]interface{}) error {
	// 1. 解析订单 Webhook 数据
	rawDataJSON, webhookInterface, platformOrderId, err := s.parseOrderWebhookData(platform, rawData, orderDataMap)
	if err != nil {
		return err
	}

	// 2. 查询现有订单（加载所有关联数据用于日志记录）
	db := ctx.GetDB()
	orderRepo := persistence.NewTakeoutOrderRepo(db)
	existingOrder, err := orderRepo.GetByUuid(
		orderUuid,
		orderRepo.WithTakeoutOrderItems(),
		orderRepo.WithTakeoutOrderItemModifiers(),
		orderRepo.WithTakeoutOrderReceiver(),
		orderRepo.WithTakeoutOrderCampaigns(),
		orderRepo.WithTakeoutOrderPromos(),
		orderRepo.WithTakeoutOrderMaterials(),
	)
	if err != nil || existingOrder == nil {
		logger.Logger.Error("查询订单失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.Error(err))
		return fmt.Errorf("查询订单失败: %w", err)
	}

	// 3. 获取平台转换器
	converter, ok := s.converters[platform]
	if !ok {
		return fmt.Errorf("不支持的平台: %s", platform)
	}

	// 4. 使用转换器将平台订单转换为通用订单格式
	updatedOrder, err := converter.ConvertOrderToTakeoutOrder(
		orderUuid,
		platform,
		platformOrderId,
		webhookInterface,
		rawDataJSON,
		time.Now().Unix(),
	)
	if err != nil {
		logger.Logger.Error("转换订单数据失败", zap.Error(err))
		return fmt.Errorf("转换订单数据失败: %w", err)
	}

	// 5. 验证商品数据
	if len(updatedOrder.TakeoutOrderItems) == 0 {
		logger.Logger.Error("订单商品数据不存在或格式错误")
		return fmt.Errorf("订单商品数据不存在或格式错误")
	}

	// 6. 检测订单菜品变动
	changeResult := service.NewOrderChangeDetector().DetectChanges(
		existingOrder.TakeoutOrderItems,
		updatedOrder.TakeoutOrderItems,
	)

	// 记录变动日志
	if changeResult.HasChange {
		logger.Logger.Info("检测到订单菜品变动",
			zap.Uint64("orderUuid", orderUuid),
			zap.String("platform", platform),
			zap.Int("returnItemCount", changeResult.GetReturnItemCount()),
			zap.Int("kitchenItemCount", changeResult.GetKitchenItemCount()),
			zap.Int("totalChanges", len(changeResult.AllChanges)))
	}

	// 7. 调用 Domain Service 进行增量更新
	err = s.orderService.IncrementalUpdateOrder(ctx, existingOrder, updatedOrder, changeResult)
	if err != nil {
		logger.Logger.Error("更新订单事务失败",
			zap.Uint64("orderUuid", orderUuid),
			zap.Error(err))
		return fmt.Errorf("更新订单事务失败: %w", err)
	}

	// 发布订单更新事件（带变动信息）
	event.GetDispatcher().Publish(event.NewOrderUpdatedEventWithChange(
		orderUuid,
		updatedOrder.Platform,
		updatedOrder.PlatformOrderId,
		updatedOrder.ShortOrderNumber,
		updatedOrder.TakeoutOrderUuid,
		ctx.GetCompanyUuid(),
		changeResult,
	))

	return nil
}

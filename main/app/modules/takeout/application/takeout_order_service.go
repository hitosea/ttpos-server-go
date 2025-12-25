package application

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
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
	SyncNewOrder(ctx context.Context, platform string, takeoutOrderUuid string, rawData map[string]interface{}) error
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
	converters map[string]*grab.GrabConverter
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
	converters := make(map[string]*grab.GrabConverter)
	grabConverter := grab.NewGrabConverter(dbm, nil)
	converters["grab"] = grabConverter
	// 后续可添加其他平台：converters["lineman"] = lineman.NewLinemanConverter(dbm)

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
	// 根据 Action 处理订单
	switch takeoutOrderEvent.Action {
	case "create":
		// 通过 RPC 查询订单信息
		orderInfo, err := s.rpcService.GetOrderInfo(ctx, takeoutOrderEvent.ShopUuid, takeoutOrderEvent.OrderUuid)
		if err != nil {
			logger.Logger.Error("RPC查询订单失败", zap.Error(err))
			return fmt.Errorf("RPC查询订单失败: %w", err)
		}
		return s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo)
	case "status_update", "cancel":
		return s.orderService.UpdateOrderStatus(ctx, takeoutOrderEvent.OrderUuid, takeoutOrderEvent.Status)
	}
	return nil
}

// SyncNewOrder 同步新订单（从 RPC 接收）
func (s *takeoutOrderAppService) SyncNewOrder(ctx context.Context, platform string, takeoutOrderUuid string, rawData map[string]interface{}) error {
	// 1. 获取平台转换器
	converter, ok := s.converters[platform]
	if !ok {
		logger.Logger.Error("不支持的平台", zap.String("platform", platform))
		return fmt.Errorf("不支持的平台: %s", platform)
	}

	// 2. 将原始数据转为 JSON
	rawDataJSON, err := json.Marshal(rawData)
	if err != nil {
		logger.Logger.Error("序列化订单数据失败", zap.Error(err))
		return fmt.Errorf("序列化订单数据失败: %w", err)
	}

	// 3. 使用转换器解析 Webhook
	webhookInterface, err := converter.ParseOrderWebhook(rawDataJSON)
	if err != nil {
		logger.Logger.Error("解析订单 Webhook 失败", zap.Error(err))
		return fmt.Errorf("解析订单 Webhook 失败: %w", err)
	}

	// 4. 类型断言为 Grab SubmitOrderRequest（Grab 平台特定）
	submitOrderReq, ok := webhookInterface.(*grabfood.SubmitOrderRequest)
	if !ok {
		logger.Logger.Error("Webhook 类型断言失败", zap.Any("webhookInterface", webhookInterface))
		return fmt.Errorf("Webhook 类型断言失败，期望 *grabfood.SubmitOrderRequest，实际类型：%T", webhookInterface)
	}

	// 5. 生成订单 UUID
	orderUuid, err := utils.GetID()
	if err != nil {
		logger.Logger.Error("生成订单UUID失败", zap.Error(err))
		return fmt.Errorf("生成订单UUID失败: %w", err)
	}

	// 6. 获取平台订单 ID（从 submitOrderReq 中获取）
	platformOrderId := submitOrderReq.GetOrderID()
	if platformOrderId == "" {
		// 兜底：如果 submitOrderReq 中没有，尝试从 rawData 获取
		if orderIdVal, ok := rawData["orderID"].(string); ok {
			platformOrderId = orderIdVal
		}
	}

	// 6. 使用转换器将平台订单转换为通用订单格式（包括商品数据）
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

	// 7. 验证商品数据
	if len(order.TakeoutOrderItems) == 0 {
		logger.Logger.Error("订单商品数据不存在或格式错误")
		return fmt.Errorf("订单商品数据不存在或格式错误")
	}

	// 设置订单UUID
	order.TakeoutOrderUuid = takeoutOrderUuid

	// 8. 调用 Domain Service 创建订单
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
	return s.orderService.AcceptOrder(ctx, req)
}

// checkOrderStock 检查订单商品库存
func (s *takeoutOrderAppService) CheckOrderStock(ctx context.Context, order *model.TakeoutOrder) (error, []string) {
	// 1. 构建 BOM 数量映射（委托给 Domain Service）
	bomQuantityMap, bomItemMap, err := s.orderService.BuildBomQuantityMap(ctx, order)
	if err != nil {
		return err, nil
	}

	// 2. 如果没有需要检查的 BOM，直接返回
	if len(bomQuantityMap) == 0 {
		return nil, nil
	}

	// 3. 调用库存模块检查库存
	inventoryAppSrv := inventoryApp.NewProductInventoryAppServiceWithDependencies(s.dbm, cache.Global)
	insufficientBomUuids, err := inventoryAppSrv.CheckStock(ctx, bomQuantityMap)
	if err != nil {
		logger.Logger.Error("检查库存失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
		return errors.WithMessage(errors.New("检查库存失败"), err.Error()), nil
	}

	// 4. 如果有库存不足的商品，构建提示信息并返回错误
	if len(insufficientBomUuids) > 0 {
		outOfStockNames := make([]string, 0, len(insufficientBomUuids))
		for _, bomUuid := range insufficientBomUuids {
			if item, ok := bomItemMap[bomUuid]; ok {
				// 遍历该 item 的所有 modifiers，找到与当前 bomUuid 匹配的 modifier
				for i := range item.TakeoutOrderItemModifiers {
					modifier := &item.TakeoutOrderItemModifiers[i]

					if modifier.IsCommodity() {
						// commodity 类型：显示套餐商品名称(规格)
						commodityName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						flavorName := language.JsonToLocaleResponse(modifier.TtposSkuName).GetLocale(ctx.GetLanguage())
						if flavorName != "" {
							outOfStockNames = append(outOfStockNames, fmt.Sprintf("%s(%s)", commodityName, flavorName))
						} else {
							outOfStockNames = append(outOfStockNames, commodityName)
						}
					} else if modifier.IsFlavor() {
						// flavor 类型：显示主商品名称(规格)
						itemName := language.JsonToLocaleResponse(item.ItemName).GetLocale(ctx.GetLanguage())
						flavorName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						outOfStockNames = append(outOfStockNames, fmt.Sprintf("%s(%s)", itemName, flavorName))
					} else if modifier.IsSauce() {
						// sauce 类型：显示主商品名称(加料)
						itemName := language.JsonToLocaleResponse(item.ItemName).GetLocale(ctx.GetLanguage())
						sauceName := language.JsonToLocaleResponse(modifier.ModifierName).GetLocale(ctx.GetLanguage())
						outOfStockNames = append(outOfStockNames, fmt.Sprintf("%s(%s)", itemName, sauceName))
					}
				}
			}
		}
		outOfStockMsg := "以下商品库存不足: " + strings.Join(outOfStockNames, ", ")
		return errors.New(outOfStockMsg), outOfStockNames
	}

	return nil, nil
}

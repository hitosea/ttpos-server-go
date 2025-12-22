package application

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/app/modules/takeout/interfaces/request"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
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

	// 通过 RPC 查询订单信息
	orderInfo, err := s.rpcService.GetOrderInfo(ctx, takeoutOrderEvent.ShopUuid, takeoutOrderEvent.OrderUuid)
	if err != nil {
		logger.Logger.Error("RPC查询订单失败", zap.Error(err))
		return fmt.Errorf("RPC查询订单失败: %w", err)
	}

	// 根据 Action 处理订单
	switch takeoutOrderEvent.Action {
	case "create":
		return s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo)
	case "status_update":
		return s.UpdateOrderStatus(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo)
	case "cancel":
		return s.SyncNewOrder(ctx, takeoutOrderEvent.ProviderName, takeoutOrderEvent.OrderUuid, orderInfo)
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

// UpdateOrderStatus 更新订单状态（从 RPC 接收状态变更）
func (s *takeoutOrderAppService) UpdateOrderStatus(ctx context.Context, platform string, orderUuid string, rawData map[string]interface{}) error {
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

	// 3. 使用转换器解析状态更新 Webhook
	webhookInterface, err := converter.ParseOrderWebhook(rawDataJSON)
	if err != nil {
		logger.Logger.Error("解析订单状态 Webhook 失败", zap.Error(err))
		return fmt.Errorf("解析订单状态 Webhook 失败: %w", err)
	}

	// 4. 从 Webhook 提取新状态（支持多种类型）
	var newStatus string

	// 尝试作为 SubmitOrderRequest 处理
	if submitOrderReq, ok := webhookInterface.(*grabfood.SubmitOrderRequest); ok {
		if submitOrderReq.HasOrderState() {
			newStatus = submitOrderReq.GetOrderState()
		}
	} else if orderStateReq, ok := webhookInterface.(*grabfood.OrderStateRequest); ok {
		// 作为 OrderStateRequest 处理（状态更新）
		newStatus = orderStateReq.State
	} else {
		logger.Logger.Error("Webhook 类型断言失败", zap.Any("webhookInterface", webhookInterface))
		return fmt.Errorf("Webhook 类型断言失败，期望 *grabfood.SubmitOrderRequest 或 *grabfood.OrderStateRequest，实际类型：%T", webhookInterface)
	}

	if newStatus == "" {
		logger.Logger.Error("订单状态为空")
		return fmt.Errorf("订单状态为空")
	}

	logger.Logger.Info("更新订单状态",
		zap.String("order_uuid", orderUuid),
		zap.String("new_status", newStatus),
		zap.String("platform", platform))

	// 6. 调用 Domain Service 更新订单状态
	return s.orderService.UpdateOrderStatus(ctx, orderUuid, newStatus, rawData)
}

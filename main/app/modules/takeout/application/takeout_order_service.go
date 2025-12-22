package application

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/adapter/grab"
	rpcAdapter "ttpos-server-go/app/modules/takeout/infrastructure/adapter/rpc"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// ITakeoutOrderAppService 外卖订单应用服务接口
type ITakeoutOrderAppService interface {
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

	// 初始化订单服务（不需要注入转换器，Application 层负责转换）
	orderService := service.NewTakeoutOrderSrv(dbm)

	return &takeoutOrderAppService{
		rpcService:   rpcService,
		dbm:          dbm,
		converters:   converters,
		orderService: orderService,
	}
}

// ==================== 订单处理相关方法 ====================

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

	// 4. 类型断言为 Grab Webhook（Grab 平台特定）
	webhook, ok := webhookInterface.(*grab.GrabOrderWebhook)
	if !ok {
		logger.Logger.Error("Webhook 类型断言失败", zap.Any("webhookInterface", webhookInterface))
		return fmt.Errorf("Webhook 类型断言失败，期望 *grab.GrabOrderWebhook，实际类型：%T", webhookInterface)
	}

	// 5. 生成订单 UUID
	orderUuid, err := utils.GetID()
	if err != nil {
		logger.Logger.Error("生成订单UUID失败", zap.Error(err))
		return fmt.Errorf("生成订单UUID失败: %w", err)
	}

	// 5. 获取平台订单 ID（从 webhook 中获取，而不是从 rawData）
	platformOrderId := webhook.OrderID
	if platformOrderId == "" {
		// 兜底：如果 webhook 中没有，尝试从 rawData 获取
		if orderIdVal, ok := rawData["order_id"].(string); ok {
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

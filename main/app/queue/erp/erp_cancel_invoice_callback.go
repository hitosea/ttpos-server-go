package erp

import (
	"context"
	"encoding/json"
	"strconv"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	ttposContext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// ErpInvoiceCancelNotifyMsg ERP 发票取消通知消息
// BMP 发送的消息结构（对应 BMP 的 InvoiceCancelNotifyMsg）
type ErpInvoiceCancelNotifyMsg struct {
	OrderNo     string `json:"order_no"`     // 订单号
	InvoiceName string `json:"invoice_name"` // 发票名称
	Remark      string `json:"remark"`       // 附注信息（JSON 格式，包含 shop_uuid 等路由信息）
}

// ErpInvoiceCancelRemarkData Remark 字段中的 JSON 数据
// ttpos 在取消发票时，将路由信息编码为 JSON 存入 Remark 字段
type ErpInvoiceCancelRemarkData struct {
	ShopUuid     uint64 `json:"shop_uuid"`            // 商户UUID（必填，用于路由到正确的数据库）
	OrderType    string `json:"order_type,omitempty"` // 订单类型（可选，用于业务扩展）
	ShouldResync bool   `json:"should_resync"`        // 是否需要重新创建发票（订单变更=true，订单取消=false）
}

// ErpCancelInvoiceCallbackHandler 处理 ERP 发票取消成功回调
// @version v2.16.0
// @spec story-erp-invoice-cancel-notification
//
// 功能说明：
//
//	处理两种场景的 ERP 发票取消回调：
//	1. 订单取消场景（should_resync=false）：只取消发票，不重新创建
//	2. 订单变更场景（should_resync=true）：取消旧发票后，重新创建新发票
//
// 完整流程：
//  1. ttpos 调用 SyncOrderCancelledToERP 取消发票（在 Remark 中传递 shop_uuid 和 should_resync）
//  2. BMP 取消成功后，发送通知到 erp-invoice-cancel topic（Remark 原样返回）
//  3. 本消费者收到通知，解析 Remark 获取 shop_uuid 和 should_resync
//  4. 根据 should_resync 标志决定是否调用 SyncOrderToERP 创建新发票
func ErpCancelInvoiceCallbackHandler(ctx context.Context, msg *primitive.MessageExt) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("处理 ERP 发票取消回调消息失败（panic）", zap.Any("panic", err))
		}
	}()

	// 1. 解析消息
	var notifyMsg ErpInvoiceCancelNotifyMsg
	if err := json.Unmarshal(msg.Body, &notifyMsg); err != nil {
		logger.Logger.Error("解析 ERP 发票取消回调消息失败",
			zap.String("msg_id", msg.MsgId),
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		return err
	}

	logger.Logger.Info("收到 ERP 发票取消成功回调",
		zap.String("msg_id", msg.MsgId),
		zap.String("order_no", notifyMsg.OrderNo),
		zap.String("remark", notifyMsg.Remark))

	// 2. 解析 Remark 字段获取路由信息
	if notifyMsg.Remark == "" {
		return nil // 返回 nil 避免消息重试
	}

	var remarkData ErpInvoiceCancelRemarkData
	if err := json.Unmarshal([]byte(notifyMsg.Remark), &remarkData); err != nil {
		logger.Logger.Error("解析 Remark 字段失败",
			zap.String("order_no", notifyMsg.OrderNo),
			zap.String("remark", notifyMsg.Remark),
			zap.Error(err))
		return nil // 返回 nil 避免消息重试
	}

	if remarkData.ShopUuid == 0 {
		logger.Logger.Error("Remark 中 shop_uuid 为空",
			zap.String("order_no", notifyMsg.OrderNo),
			zap.String("remark", notifyMsg.Remark))
		return nil // 返回 nil 避免消息重试
	}

	// 3. 解析订单UUID
	orderUuid, err := strconv.ParseUint(notifyMsg.OrderNo, 10, 64)
	if err != nil {
		logger.Logger.Error("解析订单UUID失败",
			zap.String("order_no", notifyMsg.OrderNo),
			zap.Error(err))
		return nil // 返回 nil 避免消息重试
	}

	if remarkData.OrderType == "takeout" {
		// 4. 获取数据库连接
		dbm := database.GetDBManager(config.Database)
		db := dbm.GetDB(remarkData.ShopUuid)
		if db == nil {
			logger.Logger.Error("获取数据库连接失败",
				zap.Uint64("shop_uuid", remarkData.ShopUuid),
				zap.Uint64("order_uuid", orderUuid))
			return nil // 返回 nil 避免消息重试
		}

		// 5. 获取公司信息
		companyInfo, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(remarkData.ShopUuid)
		if err != nil {
			logger.Logger.Error("获取公司信息失败",
				zap.Uint64("shop_uuid", remarkData.ShopUuid),
				zap.Uint64("order_uuid", orderUuid),
				zap.Error(err))
			return nil // 返回 nil 避免消息重试
		}

		// 6. 创建上下文
		ctxHelper := ttposContext.NewDefaultContext()
		ctxHelper.SetBasicInfo(db, companyInfo)

		// 7. 根据 should_resync 标志决定是否重新创建发票
		if !remarkData.ShouldResync {
			return nil
		}
		// 8. 订单变更场景：调用 SyncOrderToERP 创建新发票
		erpSyncService := service.NewTakeoutErpSyncService()
		if err := erpSyncService.SyncOrderToERP(ctxHelper, orderUuid); err != nil {
			logger.Logger.Error("重新同步订单到 ERP 失败",
				zap.Uint64("order_uuid", orderUuid),
				zap.Uint64("company_uuid", remarkData.ShopUuid),
				zap.Bool("should_resync", remarkData.ShouldResync),
				zap.Error(err))
			return err // 返回错误，触发消息重试
		}

		logger.Logger.Info("ERP 发票取消回调处理成功，已重新创建发票",
			zap.Uint64("order_uuid", orderUuid),
			zap.Uint64("company_uuid", remarkData.ShopUuid),
			zap.String("order_type", remarkData.OrderType),
			zap.Bool("should_resync", remarkData.ShouldResync))
	}

	return nil
}

package erp

import (
	"context"
	"encoding/json"
	"strconv"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

// ErpSalesInvoiceCallbackMsg BMP → Main 的 SI 异步回调消息
type ErpSalesInvoiceCallbackMsg struct {
	CompanyUuid       string `json:"company_uuid"`
	SaleOrderUuid     string `json:"sale_order_uuid"`
	SyncStatus        int    `json:"sync_status"`         // 0=未同步 1=已入队 2=进行中 3=成功 4=失败 5=外部取消
	SalesInvoiceName  string `json:"sales_invoice_name"`
	PaymentEntryNames string `json:"payment_entry_names"`
	ErrorMsg          string `json:"error_msg,omitempty"`
	OrderType         string `json:"order_type"`          // 订单类型: sale_order/recharge/takeout
}

// ErpSalesInvoiceCallbackHandler 处理 BMP 的 Sales Invoice 异步回调
// BMP consumer 创建 SI+PE 完成后，通过 MQ 通知 Main 更新 sale_order
func ErpSalesInvoiceCallbackHandler(ctx context.Context, msg *primitive.MessageExt) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("处理 SI 回调消息失败（panic）", zap.Any("panic", err))
		}
	}()

	var callbackMsg ErpSalesInvoiceCallbackMsg
	if err := json.Unmarshal(msg.Body, &callbackMsg); err != nil {
		logger.Logger.Error("解析 SI 回调消息失败",
			zap.String("msg_id", msg.MsgId),
			zap.Error(err),
			zap.String("body", string(msg.Body)))
		return nil // 不重试
	}

	logger.Logger.Info("收到 SI 异步回调",
		zap.String("msg_id", msg.MsgId),
		zap.String("company_uuid", callbackMsg.CompanyUuid),
		zap.String("sale_order_uuid", callbackMsg.SaleOrderUuid),
		zap.Int("sync_status", callbackMsg.SyncStatus),
		zap.String("si_name", callbackMsg.SalesInvoiceName))

	if callbackMsg.CompanyUuid == "" || callbackMsg.SaleOrderUuid == "" {
		logger.Logger.Error("SI 回调消息缺少必要字段",
			zap.String("company_uuid", callbackMsg.CompanyUuid),
			zap.String("sale_order_uuid", callbackMsg.SaleOrderUuid))
		return nil
	}

	companyUuid, err := strconv.ParseUint(callbackMsg.CompanyUuid, 10, 64)
	if err != nil {
		logger.Logger.Error("解析 company_uuid 失败",
			zap.String("company_uuid", callbackMsg.CompanyUuid),
			zap.Error(err))
		return nil
	}

	saleOrderUuid, err := strconv.ParseUint(callbackMsg.SaleOrderUuid, 10, 64)
	if err != nil {
		logger.Logger.Error("解析 sale_order_uuid 失败",
			zap.String("sale_order_uuid", callbackMsg.SaleOrderUuid),
			zap.Error(err))
		return nil
	}

	// 获取商户数据库连接
	dbm := database.GetDBManager(config.Database)
	db := dbm.GetDB(companyUuid)
	if db == nil {
		logger.Logger.Error("获取商户数据库连接失败",
			zap.Uint64("company_uuid", companyUuid))
		return nil
	}

	// ERPNext 端外部取消，需要清空 SI/PE 名称
	isExternalCancel := callbackMsg.SyncStatus == constant.ErpSyncStatusExternalCancel

	// 根据订单类型路由更新
	switch callbackMsg.OrderType {
	case "recharge":
		// 充值订单：更新 member_recharge_order
		rechargeRepo := repository.NewMemberRechargeOrderRepo(db)
		if isExternalCancel {
			// 外部取消：清空发票信息
			if err := rechargeRepo.UpdateErpSalesInvoiceInfo(saleOrderUuid, "", ""); err != nil {
				logger.Logger.Error("清空充值订单 ERP 发票信息失败",
					zap.Uint64("company_uuid", companyUuid),
					zap.Uint64("order_uuid", saleOrderUuid),
					zap.Error(err))
				return err
			}
		} else if callbackMsg.SalesInvoiceName != "" {
			if err := rechargeRepo.UpdateErpSalesInvoiceInfo(
				saleOrderUuid,
				callbackMsg.SalesInvoiceName,
				callbackMsg.PaymentEntryNames,
			); err != nil {
				logger.Logger.Error("更新充值订单 ERP 发票名称失败",
					zap.Uint64("company_uuid", companyUuid),
					zap.Uint64("order_uuid", saleOrderUuid),
					zap.Error(err))
				return err
			}
		}
		// 回写同步状态
		if err := rechargeRepo.Update(saleOrderUuid, map[string]any{
			"erp_sync_status": callbackMsg.SyncStatus,
		}); err != nil {
			logger.Logger.Error("更新充值订单 ERP 同步状态失败",
				zap.Uint64("company_uuid", companyUuid),
				zap.Uint64("order_uuid", saleOrderUuid),
				zap.Error(err))
			return err
		}

	case "takeout":
		// 外卖订单：通过 Repository 更新 takeout_order
		takeoutRepo := persistence.NewTakeoutOrderRepo(db)
		updateMap := map[string]any{
			"erp_sync_status": callbackMsg.SyncStatus,
		}
		if isExternalCancel {
			// 外部取消：清空发票信息
			updateMap["erp_pos_invoice_resp"] = ""
		} else if callbackMsg.SalesInvoiceName != "" {
			siResp := map[string]any{
				"sales_invoice_name":  callbackMsg.SalesInvoiceName,
				"payment_entry_names": callbackMsg.PaymentEntryNames,
			}
			if respJSON, err := json.Marshal(siResp); err == nil {
				updateMap["erp_pos_invoice_resp"] = string(respJSON)
			}
		}
		if err := takeoutRepo.UpdateByMap(saleOrderUuid, updateMap); err != nil {
			logger.Logger.Error("更新外卖订单 ERP 同步状态失败",
				zap.Uint64("company_uuid", companyUuid),
				zap.Uint64("order_uuid", saleOrderUuid),
				zap.Error(err))
			return err
		}

	default:
		// POS 订单（sale_order）
		saleOrderRepo := repository.NewSaleOrderRepo(db)
		if isExternalCancel {
			// 外部取消：清空 SI/PE 名称
			err = saleOrderRepo.ClearErpSalesInvoiceInfo(saleOrderUuid, callbackMsg.SyncStatus)
		} else {
			err = saleOrderRepo.UpdateSaleOrderErpSyncStatus(
				saleOrderUuid,
				callbackMsg.SyncStatus,
				callbackMsg.SalesInvoiceName,
				callbackMsg.PaymentEntryNames,
			)
		}
		if err != nil {
			logger.Logger.Error("更新 sale_order ERP 同步状态失败",
				zap.Uint64("company_uuid", companyUuid),
				zap.Uint64("sale_order_uuid", saleOrderUuid),
				zap.Error(err))
			return err
		}
	}

	logger.Logger.Info("SI 回调处理成功",
		zap.Uint64("company_uuid", companyUuid),
		zap.Uint64("sale_order_uuid", saleOrderUuid),
		zap.Int("sync_status", callbackMsg.SyncStatus),
		zap.String("si_name", callbackMsg.SalesInvoiceName),
		zap.String("order_type", callbackMsg.OrderType))

	return nil
}

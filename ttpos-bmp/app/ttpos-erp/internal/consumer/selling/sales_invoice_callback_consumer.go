package selling

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ========== SalesInvoiceCallbackConsumer ==========
// 消费 SI 回调消息，将 SI/PE 结果回写到 Main 的 shop 数据库

type SalesInvoiceCallbackConsumer struct{}

func (*SalesInvoiceCallbackConsumer) GetTopic() string {
	return string(consts.TopicErpSalesInvoiceCallback)
}

func (*SalesInvoiceCallbackConsumer) GetConcurrency() int {
	return 5
}

func (*SalesInvoiceCallbackConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	g.Log().Info(ctx, "收到SI回调消息：", string(mqMsg.Body))

	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析SI回调JSON失败")
	}
	msg := &mq.SalesInvoiceCallbackMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描SI回调JSON失败")
	}

	if msg.CompanyUuid == "" || msg.SaleOrderUuid == "" {
		g.Log().Warningf(ctx, "SI回调缺少必要字段: company_uuid=%s, sale_order_uuid=%s", msg.CompanyUuid, msg.SaleOrderUuid)
		return nil
	}

	shopDb := "shop" + msg.CompanyUuid
	db := g.DB()

	// 根据订单类型路由到不同的表
	switch msg.OrderType {
	case "recharge":
		// 充值订单：更新 member_recharge_order
		if msg.SalesInvoiceName != "" {
			sql := fmt.Sprintf("UPDATE `%s`.`ttpos_member_recharge_order` SET `erp_products_invoice_name`=? WHERE `uuid`=?", shopDb)
			if _, err = db.Exec(ctx, sql, msg.SalesInvoiceName, msg.SaleOrderUuid); err != nil {
				g.Log().Errorf(ctx, "回写充值订单失败: db=%s, uuid=%s, err=%v", shopDb, msg.SaleOrderUuid, err)
				return gerror.Wrapf(err, "回写充值订单失败")
			}
		} else if msg.SyncStatus == 4 {
			g.Log().Errorf(ctx, "充值订单SI同步失败: db=%s, uuid=%s, err=%s", shopDb, msg.SaleOrderUuid, msg.ErrorMsg)
		}
	case "takeout":
		// 外卖订单：更新 takeout_order
		if msg.SalesInvoiceName != "" {
			respData := g.Map{
				"sales_invoice_name":  msg.SalesInvoiceName,
				"payment_entry_names": msg.PaymentEntryNames,
			}
			respJson, encErr := gjson.Encode(respData)
			if encErr != nil {
				g.Log().Errorf(ctx, "序列化外卖SI响应失败: %v", encErr)
				return gerror.Wrapf(encErr, "序列化外卖SI响应失败")
			}
			sql := fmt.Sprintf("UPDATE `%s`.`ttpos_takeout_order` SET `erp_pos_invoice_resp`=? WHERE `uuid`=?", shopDb)
			if _, err = db.Exec(ctx, sql, string(respJson), msg.SaleOrderUuid); err != nil {
				g.Log().Errorf(ctx, "回写外卖订单失败: db=%s, uuid=%s, err=%v", shopDb, msg.SaleOrderUuid, err)
				return gerror.Wrapf(err, "回写外卖订单失败")
			}
		} else if msg.SyncStatus == 4 {
			g.Log().Errorf(ctx, "外卖订单SI同步失败: db=%s, uuid=%s, err=%s", shopDb, msg.SaleOrderUuid, msg.ErrorMsg)
		}
	default:
		// POS 订单（sale_order）
		var sql string
		if msg.SyncStatus == 3 {
			sql = fmt.Sprintf("UPDATE `%s`.`ttpos_sale_order` SET `erp_sales_invoice_name`=?, `erp_payment_entry_names`=?, `erp_sync_status`=? WHERE `uuid`=?", shopDb)
			_, err = db.Exec(ctx, sql, msg.SalesInvoiceName, msg.PaymentEntryNames, msg.SyncStatus, msg.SaleOrderUuid)
		} else {
			sql = fmt.Sprintf("UPDATE `%s`.`ttpos_sale_order` SET `erp_sync_status`=? WHERE `uuid`=?", shopDb)
			_, err = db.Exec(ctx, sql, msg.SyncStatus, msg.SaleOrderUuid)
		}
		if err != nil {
			g.Log().Errorf(ctx, "回写shop数据库失败: db=%s, uuid=%s, err=%v", shopDb, msg.SaleOrderUuid, err)
			return gerror.Wrapf(err, "回写shop数据库失败")
		}
	}

	g.Log().Infof(ctx, "回写shop数据库成功: db=%s, uuid=%s, status=%d, si=%s, order_type=%s", shopDb, msg.SaleOrderUuid, msg.SyncStatus, msg.SalesInvoiceName, msg.OrderType)
	return nil
}

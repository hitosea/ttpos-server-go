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

	var sql string
	if msg.SyncStatus == 2 {
		// 成功：回写 SI 名称、PE 名称、同步状态
		sql = fmt.Sprintf("UPDATE `%s`.`ttpos_sale_order` SET `erp_sales_invoice_name`=?, `erp_payment_entry_names`=?, `erp_sync_status`=? WHERE `uuid`=?", shopDb)
		_, err = db.Exec(ctx, sql, msg.SalesInvoiceName, msg.PaymentEntryNames, msg.SyncStatus, msg.SaleOrderUuid)
	} else {
		// 失败：只更新同步状态
		sql = fmt.Sprintf("UPDATE `%s`.`ttpos_sale_order` SET `erp_sync_status`=? WHERE `uuid`=?", shopDb)
		_, err = db.Exec(ctx, sql, msg.SyncStatus, msg.SaleOrderUuid)
	}

	if err != nil {
		g.Log().Errorf(ctx, "回写shop数据库失败: db=%s, uuid=%s, err=%v", shopDb, msg.SaleOrderUuid, err)
		return gerror.Wrapf(err, "回写shop数据库失败")
	}

	g.Log().Infof(ctx, "回写shop数据库成功: db=%s, uuid=%s, status=%d, si=%s", shopDb, msg.SaleOrderUuid, msg.SyncStatus, msg.SalesInvoiceName)
	return nil
}

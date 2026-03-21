package consumer

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/consumer/selling"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// DocChangeConsumer 消费 erp-doc-change topic
// 处理 ERPNext 系统通过 Webhook 发来的文档变更事件
type DocChangeConsumer struct{}

func (*DocChangeConsumer) GetTopic() string {
	return string(consts.TopicDocChange)
}

func (*DocChangeConsumer) GetConcurrency() int {
	return 5
}

// docChangeMsg 文档变更消息结构（与 DocChangeReq 对应）
type docChangeMsg struct {
	Event    string `json:"event"`
	DocType  string `json:"docType"`
	DocName  string `json:"docName"`
	SiteCode string `json:"siteCode"`
	Data     *docChangeData `json:"data,omitempty"`
}

type docChangeData struct {
	SaleOrderUuid string `json:"sale_order_uuid,omitempty"`
}

func (*DocChangeConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) error {
	g.Log().Info(ctx, "收到DocChange消息：", string(mqMsg.Body))

	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		g.Log().Errorf(ctx, "解析DocChange消息失败: %v", err)
		return nil // 不重试
	}

	msg := &docChangeMsg{}
	if err = j.Scan(msg); err != nil {
		g.Log().Errorf(ctx, "扫描DocChange消息失败: %v", err)
		return nil
	}

	// 仅处理 Sales Invoice 的 on_cancel 事件
	if msg.Event != erp.EventOnCancel || msg.DocType != erp.DocTypeSalesInvoice {
		g.Log().Debugf(ctx, "DocChange跳过: event=%s, docType=%s", msg.Event, msg.DocType)
		return nil
	}

	if msg.DocName == "" || msg.SiteCode == "" {
		g.Log().Warningf(ctx, "DocChange缺少必要字段: docName=%s, siteCode=%s", msg.DocName, msg.SiteCode)
		return nil
	}

	g.Log().Infof(ctx, "处理SI外部取消: siteCode=%s, siName=%s", msg.SiteCode, msg.DocName)

	// 查找对应的 SI 记录
	cols := dao.ReceiveSalesInvoice.Columns()
	var record *entity.ReceiveSalesInvoice
	err = dao.ReceiveSalesInvoice.Ctx(ctx).
		Where(cols.SalesInvoiceName, msg.DocName).
		Where(cols.SiteCode, msg.SiteCode).
		Scan(&record)
	if err != nil {
		g.Log().Errorf(ctx, "查询SI记录失败: siName=%s, err=%v", msg.DocName, err)
		return nil
	}
	if record == nil || record.Id == 0 {
		g.Log().Infof(ctx, "SI记录不存在，跳过: siName=%s", msg.DocName)
		return nil
	}

	// 已经是 Cancelled 状态则跳过（幂等）
	if record.Docstatus == erp.DocstatusCancelled {
		g.Log().Infof(ctx, "SI已是Cancelled状态，跳过: id=%d, siName=%s", record.Id, msg.DocName)
		return nil
	}

	// 只处理 Submitted 状态的 SI（Draft 状态说明 BMP 侧未成功创建）
	if record.Docstatus != erp.DocstatusSubmitted {
		g.Log().Warningf(ctx, "SI非Submitted状态，跳过: id=%d, docstatus=%s", record.Id, record.Docstatus)
		return nil
	}

	// 更新本地 SI 状态为 Cancelled
	_, err = dao.ReceiveSalesInvoice.Ctx(ctx).WherePri(record.Id).Data(do.ReceiveSalesInvoice{
		Docstatus: erp.DocstatusCancelled,
		RespBody:  "ERPNext外部取消(webhook on_cancel)",
		UpdatedAt: int(time.Now().Unix()),
	}).Update()
	if err != nil {
		g.Log().Errorf(ctx, "更新SI为Cancelled失败: id=%d, err=%v", record.Id, err)
		return nil
	}

	// 提取原始请求中的 companyUuid 和 orderType
	companyUuid, orderType := selling.ExtractReqMetadata(record)

	// 发送 syncStatus=5 回调通知 Main
	selling.SendSalesInvoiceCallback(ctx, companyUuid, record.SaleOrderUuid, erp.SyncStatusExternalCancel, "", "", "ERPNext外部取消", orderType)

	g.Log().Infof(ctx, "SI外部取消处理完成: id=%d, siName=%s, saleOrderUuid=%s", record.Id, msg.DocName, record.SaleOrderUuid)
	return nil
}

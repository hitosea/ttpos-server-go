package callback

import (
	"context"
	"time"

	v1 "ttpos-bmp/app/ttpos-erp/api/callback/v1"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	erp "ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/frame/g"
)

// RetrySalesInvoice 重试失败的 Sales Invoice 记录
// 查询 retry_count >= MaxRetryCount 且 docstatus=Draft 的记录，重置后重新推入 MQ
func (c *ControllerV1) RetrySalesInvoice(ctx context.Context, req *v1.RetrySalesInvoiceReq) (res *v1.RetrySalesInvoiceRes, err error) {
	cols := dao.ReceiveSalesInvoice.Columns()
	query := dao.ReceiveSalesInvoice.Ctx(ctx).
		Where(cols.Docstatus, erp.DocstatusDraft).
		WhereGTE(cols.RetryCount, erp.MaxRetryCount)

	if req.SiteCode != "" {
		query = query.Where(cols.SiteCode, req.SiteCode)
	}
	if req.RecordId > 0 {
		query = query.WherePri(req.RecordId)
	}
	if req.SaleOrderUuid != "" {
		query = query.Where(cols.SaleOrderUuid, req.SaleOrderUuid)
	}

	var records []entity.ReceiveSalesInvoice
	if err = query.Scan(&records); err != nil {
		return nil, err
	}

	count := 0
	for _, r := range records {
		// 重置重试计数和错误信息
		_, err = dao.ReceiveSalesInvoice.Ctx(ctx).WherePri(r.Id).Data(do.ReceiveSalesInvoice{
			RetryCount: 0,
			RespBody:   "",
			UpdatedAt:  int(time.Now().Unix()),
		}).Update()
		if err != nil {
			g.Log().Errorf(ctx, "重置SI记录失败: id=%d, err=%v", r.Id, err)
			continue
		}

		// 重新推入 MQ
		if pushErr := queue.PushWithContext(ctx, string(consts.TopicSaveSalesInvoice), &mq.AsyncSalesInvoiceMsg{
			RecordId: r.Id,
			MsgType:  mq.MsgTypeSaveSalesInvoice,
			SiteCode: r.SiteCode,
		}); pushErr != nil {
			g.Log().Errorf(ctx, "重新推送SI消息失败: id=%d, err=%v", r.Id, pushErr)
			continue
		}

		g.Log().Infof(ctx, "SI记录已重新入队: id=%d, order=%s", r.Id, r.SaleOrderUuid)
		count++
	}

	return &v1.RetrySalesInvoiceRes{
		Code:       "0",
		Msg:        "success",
		RetryCount: count,
	}, nil
}

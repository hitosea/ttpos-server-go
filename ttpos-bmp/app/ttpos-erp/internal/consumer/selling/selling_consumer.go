package selling

import (
	"context"
	"fmt"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/protobuf/proto"
)

type SavePosInvoiceConsumer struct {
}

func (*SavePosInvoiceConsumer) GetTopic() string {
	return string(consts.TopicSavePosInvoice)
}
func (*SavePosInvoiceConsumer) GetConcurrency() int {
	return 10
}

func (*SavePosInvoiceConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	// 这里编写商品同步的具体处理逻辑
	// 可以根据 mqMsg.Body 解析消息内容并进行相应的业务处理
	// 示例：打印消息内容
	// 实际业务中请替换为同步商品的具体实现
	// 中文注释：处理商品同步队列消息
	g.Log().Info(ctx, "收到保存发票消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSellingMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}
	cachedRecord := &entity.ReceivePosInvoice{}
	posInvoiceDao := dao.ReceivePosInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = posInvoiceDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询保存发票异步记录失败")
	}
	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.SavePosInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}
	//设置siteCode
	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})
	resp, err := service.Selling().SavePosInvoice(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，异步保存发票失败: %v", err)
		if _, err := posInvoiceDao.Data(do.ReceivePosInvoice{
			RespBody: fmt.Sprintf("保存发票失败，异步保存发票失败: %v", err),
		}).Update(); err != nil {
			return gerror.Wrapf(err, "保存发票失败，更新日志记录失败")
		}
		return gerror.Wrapf(err, "保存发票失败，异步保存发票失败")
	}
	if resp != nil {
		respBuf, err := proto.Marshal(resp)
		if err != nil {
			return gerror.Wrapf(err, "保存发票失败，序列化响应参数失败")
		}
		if _, err := posInvoiceDao.Data(do.ReceivePosInvoice{
			Docstatus:           erp.DocstatusSubmitted,
			ProductsInvoiceName: resp.ProductsInvoiceName,
			MaterialInvoiceName: resp.MaterialInvoiceName,
			RespMessage:         gbase64.EncodeToString(respBuf),
			RespBody:            resp.String(),
		}).Update(); err != nil {
			return gerror.Wrapf(err, "保存发票失败，更新日志记录失败")
		}
	}

	return nil
}

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

type ReturnPosInvoiceConsumer struct {
}

func (*ReturnPosInvoiceConsumer) GetTopic() string {
	return string(consts.TopicReturnPosInvoice)
}
func (*ReturnPosInvoiceConsumer) GetConcurrency() int {
	return 10
}

func (*ReturnPosInvoiceConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {

	g.Log().Info(ctx, "收到退款发票消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSellingMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}
	cachedRecord := &entity.ReceiveReturnPosInvoice{}
	returnInvoiceDao := dao.ReceiveReturnPosInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = returnInvoiceDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询退款发票异步记录失败")
	}

	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.ReturnPosInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}
	//设置siteCode
	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})
	returnDao := dao.ReceiveReturnPosInvoice.Ctx(ctx).WherePri(msg.RecordId)
	receivePosInvoice, err := service.AsyncSelling().GetLatestReceivePosInvoice(ctx, &do.ReceivePosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
	})
	if err != nil {
		respMessage := fmt.Sprintf("退款发票失败，查询原POS记录失败: %v", err)
		if _, updateErr := returnDao.Data(do.ReceiveReturnPosInvoice{
			RespBody: respMessage,
		}).Update(); updateErr != nil {
			return gerror.Wrapf(updateErr, "退款发票失败，更新日志记录失败")
		}
		return gerror.Wrapf(err, "退款发票失败，请求参数: %v", req)
	}
	//开始退款
	if receivePosInvoice == nil {
		respMessage := fmt.Sprintf("退款发票失败，未查询到原POS记录，请求参数: %v", req)
		if _, updateErr := returnDao.Data(do.ReceiveReturnPosInvoice{
			RespBody: respMessage,
		}).Update(); updateErr != nil {
			return gerror.Wrapf(updateErr, "退款发票失败，更新日志记录失败")
		}
		return gerror.Wrap(gerror.New(respMessage), respMessage)
	} else {
		//判断发票是什么类型
		switch req.InvoiceType {
		case 1:
			req.InvoiceName = receivePosInvoice.ProductsInvoiceName
		case 2:
			req.InvoiceName = receivePosInvoice.MaterialInvoiceName
		}
		resp, err := service.Selling().ReturnPosInvoice(ctx, req)
		if err != nil {
			respMessage := fmt.Sprintf("退款发票失败，异步退款发票失败: %v", err)
			if _, updateErr := returnDao.Data(do.ReceiveReturnPosInvoice{
				RespBody: respMessage,
			}).Update(); updateErr != nil {
				return gerror.Wrapf(updateErr, "退款发票失败，更新日志记录失败")
			}
			return gerror.Wrapf(err, "退款发票失败，请求参数: %v", req)
		}
		if resp != nil {
			respBuf, err := proto.Marshal(resp)
			if err != nil {
				return gerror.Wrapf(err, "退款发票失败，序列化响应参数失败: %v", resp)
			}
			if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
				Docstatus:   erp.DocstatusSubmitted,
				RespMessage: gbase64.EncodeToString(respBuf),
				RespBody:    resp.String(),
			}).Update(); err != nil {
				return gerror.Wrapf(err, "退款发票失败，更新日志记录失败")
			}
		}
	}

	return nil
}

type CancelPosInvoice struct {
}

func (*CancelPosInvoice) GetTopic() string {
	return string(consts.TopicCancelPosInvoice)
}
func (*CancelPosInvoice) GetConcurrency() int {
	return 10
}

func (*CancelPosInvoice) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {

	g.Log().Info(ctx, "收到取消发票消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSellingMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}
	cachedRecord := &entity.ReceiveCancelPosInvoice{}
	cancelPosInvoiceDao := dao.ReceiveCancelPosInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = cancelPosInvoiceDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询取消发票异步记录失败")
	}

	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.CancelPosInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}

	//设置siteCode
	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})

	cancelDao := dao.ReceiveCancelPosInvoice.Ctx(ctx).WherePri(msg.RecordId)
	receivePosInvoice, err := service.AsyncSelling().GetLatestReceivePosInvoice(ctx, &do.ReceivePosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
	})
	if err != nil {
		respMessage := fmt.Sprintf("取消发票失败，查询原POS记录失败: %v", err)
		if _, updateErr := cancelDao.Data(do.ReceiveCancelPosInvoice{
			RespBody: respMessage,
		}).Update(); updateErr != nil {
			return gerror.Wrapf(updateErr, "取消发票失败，更新日志记录失败")
		}
		return gerror.Wrapf(err, "取消发票失败，请求参数: %v", req)
	}
	if receivePosInvoice != nil {
		// 调用服务层取消商品发票
		if len(receivePosInvoice.ProductsInvoiceName) > 0 {
			err := service.Selling().CancelPosInvoice(ctx, receivePosInvoice.ProductsInvoiceName)
			if err != nil {
				respMessage := fmt.Sprintf("取消商品发票[%s]失败: %v", receivePosInvoice.ProductsInvoiceName, err)
				if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
					RespBody: respMessage,
				}).Update(); err != nil {
					return gerror.Wrapf(err, "取消发票失败，更新日志记录失败")
				}
			}
		}
		// 调用服务层取消材料发票
		if len(receivePosInvoice.MaterialInvoiceName) > 0 {
			err := service.Selling().CancelPosInvoice(ctx, receivePosInvoice.MaterialInvoiceName)
			if err != nil {
				respMessage := fmt.Sprintf("取消材料发票[%s]失败: %v", receivePosInvoice.MaterialInvoiceName, err)
				if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
					RespBody: respMessage,
				}).Update(); err != nil {
					return gerror.Wrapf(err, "取消材料发票失败，更新日志记录失败")
				}
			}
		}
		//更新原下单发票记录
		_, err = dao.ReceivePosInvoice.Ctx(ctx).WherePri(receivePosInvoice.Id).Update(do.ReceivePosInvoice{
			Docstatus: erp.DocstatusCancelled,
		})
		if err != nil {
			return gerror.Wrapf(err, "取消发票失败，更新原下单发票记录失败")
		}
		if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
			Docstatus: erp.DocstatusSubmitted,
		}).Update(); err != nil {
			return gerror.Wrapf(err, "取消发票失败，更新日志记录失败")
		}
	}

	return nil
}

type ClosePosEntryConsumer struct {
}

func (*ClosePosEntryConsumer) GetTopic() string {
	return string(consts.TopicClosePosEntry)
}
func (*ClosePosEntryConsumer) GetConcurrency() int {
	return 10
}

func (*ClosePosEntryConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {

	g.Log().Info(ctx, "收到关账消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSellingMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}
	cachedRecord := &entity.ReceiveClosePos{}
	closePosEntryDao := dao.ReceiveClosePos.Ctx(ctx).WherePri(msg.RecordId)
	if err = closePosEntryDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询关账异步记录失败")
	}

	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.ClosePosEntryReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}

	//设置siteCode
	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})
	closePosDao := dao.ReceiveClosePos.Ctx(ctx).WherePri(msg.RecordId)
	resp, err := service.Selling().ClosePosEntry(ctx, req)
	if err != nil {
		if _, err := closePosDao.Data(do.ReceiveClosePos{
			RespBody: fmt.Sprintf("保存关账失败，异步保存失败: %v", err),
		}).Update(); err != nil {
			return gerror.Wrapf(err, "保存关账失败，更新日志记录失败")
		}
		return gerror.Wrapf(err, "关帐失败，异步关帐失败: %v", req)
	}
	if resp != nil {
		respBuf, err := proto.Marshal(resp)
		if err != nil {
			return gerror.Wrapf(err, "保存关账失败，序列化响应参数失败: %v", resp)
		}
		if _, err := closePosDao.Data(do.ReceiveClosePos{
			Docstatus:   erp.DocstatusSubmitted,
			RespMessage: gbase64.EncodeToString(respBuf),
			RespBody:    resp.String(),
		}).Update(); err != nil {
			return gerror.Wrapf(err, "保存关账失败，更新日志记录失败")
		}
	}
	return nil
}

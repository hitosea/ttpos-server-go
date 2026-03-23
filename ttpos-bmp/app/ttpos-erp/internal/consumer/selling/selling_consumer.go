package selling

import (
	"context"
	"fmt"
	"time"
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
		//TODO 可能存在新的错误信息更新不上来的情况
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

	// 检查发票是否存在或状态是否为 Draft（创建中）
	// 注意：GetLatestReceivePosInvoice 查不到记录时返回空结构体（Id=0），不是 nil
	invoiceNotExistOrDraft := receivePosInvoice == nil || receivePosInvoice.Id == 0 || receivePosInvoice.Docstatus == erp.DocstatusDraft
	if invoiceNotExistOrDraft {
		// 检查是否超时（超过2分钟不再重试）
		elapsed := time.Now().Unix() - int64(cachedRecord.CreatedAt)
		if elapsed > consts.MaxWaitForSISeconds {
			respMessage := fmt.Sprintf("退款发票超时，发票记录不存在或创建未完成: OrderNo=%s, elapsed=%ds", req.OrderNo, elapsed)
			g.Log().Errorf(ctx, respMessage)
			if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
				RespBody: respMessage,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "更新退款发票记录失败: %v", err)
			}
			return nil // 超时不再重试
		}
		g.Log().Infof(ctx, "发票记录不存在或正在创建中，5秒后重试: OrderNo=%s, OpenPosEntryName=%s, elapsed=%ds",
			req.OrderNo, req.OpenPosEntryName, elapsed)
		// 发送延时消息，5秒后重试
		if err := queue.DelayPush(string(consts.TopicReturnPosInvoice), &mq.AsyncSellingMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeReturnPosInvoice,
		}, 5*time.Second); err != nil {
			g.Log().Errorf(ctx, "发送延时重试消息失败: %v", err)
			return gerror.Wrapf(err, "发送延时重试消息失败")
		}
		return nil // 当前消息处理完成，等待延时消息重试
	}

	//开始退款（此时发票记录已存在且状态为 Submitted）
	if receivePosInvoice == nil || receivePosInvoice.Id == 0 {
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

// sendInvoiceCancelNotify 发送发票取消通知
// 异步发送，失败仅记录日志，不阻塞主流程
func sendInvoiceCancelNotify(ctx context.Context, req *selling.CancelPosInvoiceReq) {
	notifyMsg := &mq.InvoiceCancelNotifyMsg{
		OrderNo:     req.OrderNo,
		InvoiceName: req.ProductsInvoiceName,
		Remark:      req.Remark,
	}

	if err := queue.PushWithContext(ctx, string(consts.TopicErpInvoiceCancel), notifyMsg); err != nil {
		g.Log().Errorf(ctx, "发送发票取消通知失败: order_no=%s, invoice_name=%s, err=%v",
			req.OrderNo, req.ProductsInvoiceName, err)
		return
	}

	g.Log().Infof(ctx, "发票取消通知已发送: order_no=%s, invoice_name=%s, remark=%s",
		req.OrderNo, req.ProductsInvoiceName, req.Remark)
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

	// 检查发票是否存在或状态是否为 Draft（创建中）
	// 注意：GetLatestReceivePosInvoice 查不到记录时返回空结构体（Id=0），不是 nil
	invoiceNotExistOrDraft := receivePosInvoice == nil || receivePosInvoice.Id == 0 || receivePosInvoice.Docstatus == erp.DocstatusDraft
	if invoiceNotExistOrDraft {
		// 检查是否超时（超过2分钟不再重试）
		elapsed := time.Now().Unix() - int64(cachedRecord.CreatedAt)
		if elapsed > consts.MaxWaitForSISeconds {
			respMessage := fmt.Sprintf("取消发票超时，发票记录不存在或创建未完成: OrderNo=%s, elapsed=%ds", req.OrderNo, elapsed)
			g.Log().Errorf(ctx, respMessage)
			if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
				RespBody: respMessage,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "更新取消发票记录失败: %v", err)
			}
			return nil // 超时不再重试
		}
		g.Log().Infof(ctx, "发票记录不存在或正在创建中，5秒后重试: OrderNo=%s, OpenPosEntryName=%s, elapsed=%ds",
			req.OrderNo, req.OpenPosEntryName, elapsed)
		// 发送延时消息，5秒后重试
		if err := queue.DelayPush(string(consts.TopicCancelPosInvoice), &mq.AsyncSellingMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeCancelPosInvoice,
		}, 5*time.Second); err != nil {
			g.Log().Errorf(ctx, "发送延时重试消息失败: %v", err)
			return gerror.Wrapf(err, "发送延时重试消息失败")
		}
		return nil // 当前消息处理完成，等待延时消息重试
	}

	if receivePosInvoice != nil && receivePosInvoice.Id > 0 {
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
		_, err = dao.ReceivePosInvoice.Ctx(ctx).WherePri(receivePosInvoice.Id).Data(do.ReceivePosInvoice{
			Docstatus: erp.DocstatusCancelled,
		}).Update()
		if err != nil {
			return gerror.Wrapf(err, "取消发票失败，更新原下单发票记录失败")
		}
		if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
			Docstatus: erp.DocstatusSubmitted,
		}).Update(); err != nil {
			return gerror.Wrapf(err, "取消发票失败，更新日志记录失败")
		}

		// 发送发票取消通知到 erp-invoice-cancel topic
		sendInvoiceCancelNotify(ctx, req)
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

	count, err := dao.ReceivePosInvoice.Ctx(ctx).Where(g.Map{
		dao.ReceivePosInvoice.Columns().Docstatus:        0,
		dao.ReceivePosInvoice.Columns().OpenPosEntryName: cachedRecord.PosOpenEntryName,
		dao.ReceivePosInvoice.Columns().SiteCode:         cachedRecord.SiteCode,
	}).Count()
	if err != nil {
		return gerror.Wrapf(err, "查询POS发票未处理完的处理失败")
	}
	if count > 0 {
		return gerror.New("POS发票还有未处理完的处理，不允许提交")
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

type RedoPosConsumer struct {
}

func (*RedoPosConsumer) GetTopic() string {
	return string(consts.TopicRedoPos)
}
func (*RedoPosConsumer) GetConcurrency() int {
	return 10
}

// Handle 重做未处理的订单
// 消息参考: {"msg_type":"save-pos-invoice","pos_open_entry_name":"POS-OPE-2025-00238","site_code":"SITE001"}
func (*RedoPosConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	g.Log().Info(ctx, "收到重做消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSellingMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}
	if msg.PosOpenEntryName == "" {
		g.Log().Infof(ctx, "重做开单名称不能为空：%s", msg.PosOpenEntryName)
		return nil
	}
	// 验证 SiteCode（向后兼容）
	if msg.SiteCode == "" {
		g.Log().Warningf(ctx, "重做消息缺少 SiteCode，可能存在跨站点风险：%s", string(mqMsg.Body))
	}
	//重发所有未处理的商品发票
	switch msg.MsgType {
	case mq.MsgTypeSavePosInvoice:
		posInvoiceDao := dao.ReceivePosInvoice.Ctx(ctx)
		//查询所有未处理的商品发票
		posInvoiceList := make([]*entity.ReceivePosInvoice, 0)
		whereCondition := do.ReceivePosInvoice{
			OpenPosEntryName: msg.PosOpenEntryName,
			Docstatus:        erp.DocstatusDraft,
		}
		// 添加 SiteCode 过滤（当 SiteCode 不为空时）
		if msg.SiteCode != "" {
			whereCondition.SiteCode = msg.SiteCode
		}
		err = posInvoiceDao.Where(whereCondition).Scan(&posInvoiceList)
		if err != nil {
			return gerror.Wrapf(err, "查询商品发票失败")
		}
		//重发所有未处理的商品发票
		for _, posInvoice := range posInvoiceList {
			//发送消息
			if err = queue.Push(string(consts.TopicSavePosInvoice), &mq.AsyncSellingMsg{
				RecordId: posInvoice.Id,
				MsgType:  mq.MsgTypeSavePosInvoice,
			}); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，发送异步消息失败: %v", err)
				return gerror.Wrapf(err, "保存发票失败，发送异步消息失败: %v", posInvoice.Id)
			}
		}
	case mq.MsgTypeCancelPosInvoice:
		cancelDao := dao.ReceiveCancelPosInvoice.Ctx(ctx)
		//查询所有未处理的取消发票
		cancelList := make([]*entity.ReceiveCancelPosInvoice, 0)
		whereCondition := do.ReceiveCancelPosInvoice{
			OpenPosEntryName: msg.PosOpenEntryName,
			Docstatus:        erp.DocstatusDraft,
		}
		// 添加 SiteCode 过滤（当 SiteCode 不为空时）
		if msg.SiteCode != "" {
			whereCondition.SiteCode = msg.SiteCode
		}
		err = cancelDao.Where(whereCondition).Scan(&cancelList)
		if err != nil {
			return gerror.Wrapf(err, "查询取消发票失败")
		}
		//重发所有未处理的取消发票
		for _, cancel := range cancelList {
			//发送消息
			if err = queue.Push(string(consts.TopicCancelPosInvoice), &mq.AsyncSellingMsg{
				RecordId: cancel.Id,
				MsgType:  mq.MsgTypeCancelPosInvoice,
			}); err != nil {
				g.Log().Errorf(ctx, "取消发票失败，发送异步消息失败: %v", err)
				return gerror.Wrapf(err, "取消发票失败，发送异步消息失败: %v", cancel.Id)
			}
		}
	case mq.MsgTypeReturnPosInvoice:
		returnDao := dao.ReceiveReturnPosInvoice.Ctx(ctx)
		//查询所有未处理的退货发票
		returnList := make([]*entity.ReceiveReturnPosInvoice, 0)
		whereCondition := do.ReceiveReturnPosInvoice{
			OpenPosEntryName: msg.PosOpenEntryName,
			Docstatus:        erp.DocstatusDraft,
		}
		// 添加 SiteCode 过滤（当 SiteCode 不为空时）
		if msg.SiteCode != "" {
			whereCondition.SiteCode = msg.SiteCode
		}
		err = returnDao.Where(whereCondition).Scan(&returnList)
		if err != nil {
			return gerror.Wrapf(err, "查询退货发票失败")
		}
		//重发所有未处理的退货发票
		for _, returnInvoice := range returnList {
			//发送消息
			if err = queue.Push(string(consts.TopicReturnPosInvoice), &mq.AsyncSellingMsg{
				RecordId: returnInvoice.Id,
				MsgType:  mq.MsgTypeReturnPosInvoice,
			}); err != nil {
				g.Log().Errorf(ctx, "退货发票失败，发送异步消息失败: %v", err)
				return gerror.Wrapf(err, "退货发票失败，发送异步消息失败: %v", returnInvoice.Id)
			}
		}
	case mq.MsgTypeClosePosEntry:
		closePosDao := dao.ReceiveClosePos.Ctx(ctx)
		//查询所有未处理的关账记录
		closePosList := make([]*entity.ReceiveClosePos, 0)
		whereCondition := do.ReceiveClosePos{
			PosOpenEntryName: msg.PosOpenEntryName,
			Docstatus:        erp.DocstatusDraft,
		}
		// 添加 SiteCode 过滤（当 SiteCode 不为空时）
		if msg.SiteCode != "" {
			whereCondition.SiteCode = msg.SiteCode
		}
		err = closePosDao.Where(whereCondition).Scan(&closePosList)
		if err != nil {
			return gerror.Wrapf(err, "查询关账记录失败")
		}
		//重发所有未处理的关账记录
		for _, closePos := range closePosList {
			//发送消息
			if err = queue.Push(string(consts.TopicClosePosEntry), &mq.AsyncSellingMsg{
				RecordId: closePos.Id,
				MsgType:  mq.MsgTypeClosePosEntry,
			}); err != nil {
				g.Log().Errorf(ctx, "关帐失败，发送异步消息失败: %v", err)
				return gerror.Wrapf(err, "关帐失败，发送异步消息失败: %v", closePos.Id)
			}
		}
	default:
		g.Log().Infof(ctx, "重做消息类型[%s]未处理", msg.MsgType)
		return nil
	}

	return nil
}

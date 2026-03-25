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

const maxRetryCount = erp.MaxRetryCount

// retryDelay 根据重试次数计算延迟时间（指数退避：5s, 10s, 20s, 40s, 80s）
func retryDelay(retryCount int) time.Duration {
	return time.Duration(5*(1<<retryCount)) * time.Second
}

// ========== SaveSalesInvoiceConsumer ==========

type SaveSalesInvoiceConsumer struct{}

func (*SaveSalesInvoiceConsumer) GetTopic() string {
	return string(consts.TopicSaveSalesInvoice)
}

func (*SaveSalesInvoiceConsumer) GetConcurrency() int {
	return 10
}

func (*SaveSalesInvoiceConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	g.Log().Info(ctx, "收到SaveSalesInvoice消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSalesInvoiceMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}

	cachedRecord := &entity.ReceiveSalesInvoice{}
	siDao := dao.ReceiveSalesInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = siDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询SI异步记录失败")
	}

	// 幂等性检查1：已成功提交则跳过
	if cachedRecord.Docstatus == erp.DocstatusSubmitted {
		g.Log().Infof(ctx, "SI已提交，跳过: record_id=%d", msg.RecordId)
		return nil
	}

	// 幂等性检查2：基于 mq_msg_id 避免重复处理（MQ 重试场景）
	// 如果当前消息 ID 与数据库记录的 mq_msg_id 相同，且状态为 Draft（处理中），说明正在处理中或已处理
	if cachedRecord.MqMsgId == mqMsg.MsgId && cachedRecord.Docstatus == erp.DocstatusDraft {
		g.Log().Infof(ctx, "检测到重复消息（MQ重试），跳过: record_id=%d, mq_msg_id=%s", msg.RecordId, mqMsg.MsgId)
		return nil
	}

	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.SaveSalesInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}

	// 设置 siteCode
	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})

	resp, err := service.Selling().SaveSalesInvoice(ctx, req)
	if err != nil {
		// 应用层重试：使用数据库 retry_count 字段追踪重试次数
		retryCount := cachedRecord.RetryCount + 1
		g.Log().Errorf(ctx, "SaveSalesInvoice失败 (retry=%d/%d): %v", retryCount, maxRetryCount, err)

		// 更新数据库重试计数和错误信息
		siDao.Data(do.ReceiveSalesInvoice{
			RetryCount: retryCount,
			MqMsgId:    mqMsg.MsgId,
			RespBody:   fmt.Sprintf(consts.ErrMsgSaveSIRetry, retryCount, err),
			UpdatedAt:  int(time.Now().Unix()),
		}).Update()

		if retryCount < maxRetryCount {
			// 延迟重试：使用指数退避策略重新推入队列
			queue.DelayPush(string(consts.TopicSaveSalesInvoice), &mq.AsyncSalesInvoiceMsg{
				RecordId: msg.RecordId,
				MsgType:  mq.MsgTypeSaveSalesInvoice,
				SiteCode: msg.SiteCode,
			}, retryDelay(retryCount))
			return nil // 不返回 error，避免 RocketMQ 自身重试
		}

		// 重试耗尽：1. 发送失败回调到 Main，2. 推送到 DLQ 作为兜底
		SendSalesInvoiceCallback(ctx, req.CompanyUuid, req.SaleOrderUuid, erp.SyncStatusFailed, "", "", err.Error(), req.OrderType)
		PushToDLQ(ctx, consts.TopicSaveSalesInvoiceDLQ, &mq.AsyncSalesInvoiceMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeSaveSalesInvoice,
			SiteCode: msg.SiteCode,
		}, retryCount, err.Error(), mqMsg.MsgId, req.CompanyUuid, req.OrderType, req.SaleOrderUuid)
		return nil // 已处理失败，不需要 MQ 重试
	}

	if resp != nil {
		respBuf, err := proto.Marshal(resp)
		if err != nil {
			return gerror.Wrapf(err, "序列化响应参数失败")
		}
		peNamesJson, _ := gjson.Encode(resp.PaymentEntryNames)
		if _, err := siDao.Data(do.ReceiveSalesInvoice{
			Docstatus:         erp.DocstatusSubmitted,
			SalesInvoiceName:  resp.SalesInvoiceName,
			PaymentEntryNames: string(peNamesJson),
			RespMessage:       gbase64.EncodeToString(respBuf),
			RespBody:          resp.String(),
			MqMsgId:           mqMsg.MsgId,
			UpdatedAt:         int(time.Now().Unix()),
		}).Update(); err != nil {
			return gerror.Wrapf(err, "更新SI日志记录失败")
		}

		// 发送成功回调到 Main
		SendSalesInvoiceCallback(ctx, req.CompanyUuid, req.SaleOrderUuid, erp.SyncStatusSuccess, resp.SalesInvoiceName, string(peNamesJson), "", req.OrderType)
	}

	return nil
}

// SendSalesInvoiceCallback 通过 MQ 发送 SI 回调消息到 Main
// syncStatus: 0=未同步 1=已入队 2=进行中 3=成功 4=失败 5=外部取消
func SendSalesInvoiceCallback(ctx context.Context, companyUuid, saleOrderUuid string, syncStatus int, siName, peNamesJson, errMsg, orderType string) {
	if companyUuid == "" {
		g.Log().Warningf(ctx, "SI回调跳过: company_uuid为空, sale_order_uuid=%s", saleOrderUuid)
		return
	}
	callbackMsg := &mq.SalesInvoiceCallbackMsg{
		CompanyUuid:       companyUuid,
		SaleOrderUuid:     saleOrderUuid,
		SyncStatus:        syncStatus,
		SalesInvoiceName:  siName,
		PaymentEntryNames: peNamesJson,
		ErrorMsg:          errMsg,
		OrderType:         orderType,
	}
	if err := queue.PushWithContext(ctx, string(consts.TopicErpSalesInvoiceCallback), callbackMsg); err != nil {
		g.Log().Errorf(ctx, "发送SI回调消息失败: uuid=%s, err=%v", saleOrderUuid, err)
	} else {
		g.Log().Infof(ctx, "发送SI回调消息成功: uuid=%s, status=%d, si=%s", saleOrderUuid, syncStatus, siName)
	}
}

// PushToDLQ 推送失败消息到自定义 DLQ Topic
// 应用层重试耗尽后调用，作为最后的兜底保障
func PushToDLQ(ctx context.Context, dlqTopic consts.Topic, originalMsg *mq.AsyncSalesInvoiceMsg, retryCount int, lastError string, lastMqMsgId string, companyUuid, orderType, orderUuid string) {
	dlqMsg := &mq.SalesInvoiceDLQMsg{
		OriginalMsg: *originalMsg,
		RetryCount:  retryCount,
		LastError:   lastError,
		LastMqMsgId: lastMqMsgId,
		FailedAt:    time.Now().Unix(),
		CompanyUuid: companyUuid,
		OrderType:   orderType,
		OrderUuid:   orderUuid,
	}

	if err := queue.PushWithContext(ctx, string(dlqTopic), dlqMsg); err != nil {
		g.Log().Errorf(ctx, "推送到DLQ失败 [%s]: record_id=%d, err=%v", dlqTopic, originalMsg.RecordId, err)
	} else {
		g.Log().Warningf(ctx, "已推送到DLQ [%s]: record_id=%d, retry_count=%d, company=%s, order=%s",
			dlqTopic, originalMsg.RecordId, retryCount, companyUuid, orderUuid)
	}
}

// ExtractReqMetadata 从 ReceiveSalesInvoice 的 ReqMessage 中提取 CompanyUuid 和 OrderType
func ExtractReqMetadata(record *entity.ReceiveSalesInvoice) (companyUuid, orderType string) {
	if record == nil || record.ReqMessage == "" {
		return "", ""
	}
	reqBuf, err := gbase64.DecodeString(record.ReqMessage)
	if err != nil {
		return "", ""
	}
	req := &selling.SaveSalesInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return "", ""
	}
	return req.CompanyUuid, req.OrderType
}

// ========== CancelSalesInvoiceConsumer ==========

type CancelSalesInvoiceConsumer struct{}

func (*CancelSalesInvoiceConsumer) GetTopic() string {
	return string(consts.TopicCancelSalesInvoice)
}

func (*CancelSalesInvoiceConsumer) GetConcurrency() int {
	return 10
}

func (*CancelSalesInvoiceConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	g.Log().Info(ctx, "收到CancelSalesInvoice消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSalesInvoiceMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}

	// 直接读取原始 SI 记录（Cancel 不再插入新记录，RecordId 就是原 SI 的 ID）
	originalRecord := &entity.ReceiveSalesInvoice{}
	siDao := dao.ReceiveSalesInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = siDao.Scan(&originalRecord); err != nil {
		return gerror.Wrap(err, "查询原SI记录失败")
	}
	if originalRecord == nil || originalRecord.Id == 0 {
		g.Log().Errorf(ctx, "原SI记录不存在: record_id=%d", msg.RecordId)
		return nil
	}

	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: originalRecord.SiteCode,
	})

	// 幂等性检查1：已取消则跳过
	if originalRecord.Docstatus == erp.DocstatusCancelled {
		g.Log().Infof(ctx, "原SI已取消，跳过: sale_order_uuid=%s", originalRecord.SaleOrderUuid)
		return nil
	}

	// 幂等性检查2：基于 mq_msg_id 避免重复处理（MQ 重试场景）
	// 如果当前消息 ID 与数据库记录的 mq_msg_id 相同，且状态为 Cancelled，说明已处理
	if originalRecord.MqMsgId == mqMsg.MsgId && originalRecord.Docstatus == erp.DocstatusCancelled {
		g.Log().Infof(ctx, "检测到重复消息（MQ重试），跳过: record_id=%d, mq_msg_id=%s", msg.RecordId, mqMsg.MsgId)
		return nil
	}
	if originalRecord.Docstatus == erp.DocstatusDraft {
		// SI 尚未完成，检查是否已失败
		if originalRecord.RespBody != "" {
			g.Log().Infof(ctx, "原SI已失败，无需取消: sale_order_uuid=%s", originalRecord.SaleOrderUuid)
			return nil
		}
		// SI 还在处理中，延迟重试（业务等待，非失败重试）
		elapsed := time.Now().Unix() - int64(originalRecord.CreatedAt)
		if elapsed > consts.MaxWaitForSISeconds {
			g.Log().Errorf(ctx, "取消SI超时，原SI未完成: sale_order_uuid=%s, elapsed=%ds", originalRecord.SaleOrderUuid, elapsed)
			return nil
		}
		queue.DelayPush(string(consts.TopicCancelSalesInvoice), &mq.AsyncSalesInvoiceMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeCancelSalesInvoice,
			SiteCode: msg.SiteCode,
		}, 5*time.Second)
		return nil
	}

	// Docstatus == Submitted，执行取消
	// 从原记录构建取消请求
	var peNames []string
	if originalRecord.PaymentEntryNames != "" {
		_ = gjson.DecodeTo([]byte(originalRecord.PaymentEntryNames), &peNames)
	}
	cancelReq := &selling.CancelSalesInvoiceReq{
		OrderNo:           originalRecord.OrderNo,
		SaleOrderUuid:     originalRecord.SaleOrderUuid,
		SalesInvoiceName:  originalRecord.SalesInvoiceName,
		PaymentEntryNames: peNames,
	}

	if err := service.Selling().CancelSalesInvoice(ctx, cancelReq); err != nil {
		// 应用层重试：使用数据库 retry_count 字段追踪重试次数
		retryCount := originalRecord.RetryCount + 1
		g.Log().Errorf(ctx, "取消SI失败 (retry=%d/%d): %v", retryCount, maxRetryCount, err)

		// 更新数据库重试计数和错误信息
		siDao.Data(do.ReceiveSalesInvoice{
			RetryCount: retryCount,
			MqMsgId:    mqMsg.MsgId,
			RespBody:   fmt.Sprintf(consts.ErrMsgCancelSIRetry, retryCount, err),
			UpdatedAt:  int(time.Now().Unix()),
		}).Update()

		if retryCount < maxRetryCount {
			// 延迟重试：使用指数退避策略重新推入队列
			queue.DelayPush(string(consts.TopicCancelSalesInvoice), &mq.AsyncSalesInvoiceMsg{
				RecordId: msg.RecordId,
				MsgType:  mq.MsgTypeCancelSalesInvoice,
				SiteCode: msg.SiteCode,
			}, retryDelay(retryCount))
			return nil // 不返回 error，避免 RocketMQ 自身重试
		}

		// 重试耗尽：1. 发送失败回调到 Main，2. 推送到 DLQ 作为兜底
		companyUuid, orderType := ExtractReqMetadata(originalRecord)
		SendSalesInvoiceCallback(ctx, companyUuid, originalRecord.SaleOrderUuid, erp.SyncStatusFailed, "", "", err.Error(), orderType)
		PushToDLQ(ctx, consts.TopicCancelSalesInvoiceDLQ, &mq.AsyncSalesInvoiceMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeCancelSalesInvoice,
			SiteCode: msg.SiteCode,
		}, retryCount, err.Error(), mqMsg.MsgId, companyUuid, orderType, originalRecord.SaleOrderUuid)
		return nil // 已处理失败，不需要 MQ 重试
	}

	// 更新原记录状态为 Cancelled
	if _, updateErr := siDao.Data(do.ReceiveSalesInvoice{
		Docstatus: erp.DocstatusCancelled,
		MqMsgId:   mqMsg.MsgId,
		UpdatedAt: int(time.Now().Unix()),
	}).Update(); updateErr != nil {
		g.Log().Errorf(ctx, "更新取消SI状态失败: record_id=%d, err=%v", msg.RecordId, updateErr)
	}

	return nil
}

// ========== ReturnSalesInvoiceConsumer ==========

type ReturnSalesInvoiceConsumer struct{}

func (*ReturnSalesInvoiceConsumer) GetTopic() string {
	return string(consts.TopicReturnSalesInvoice)
}

func (*ReturnSalesInvoiceConsumer) GetConcurrency() int {
	return 10
}

func (*ReturnSalesInvoiceConsumer) Handle(ctx context.Context, mqMsg queue.MqMsg) (err error) {
	g.Log().Info(ctx, "收到ReturnSalesInvoice消息：", string(mqMsg.Body))
	j, err := gjson.DecodeToJson(mqMsg.Body)
	if err != nil {
		return gerror.Wrap(err, "解析JSON数据失败")
	}
	msg := &mq.AsyncSalesInvoiceMsg{}
	if err = j.Scan(msg); err != nil {
		return gerror.Wrap(err, "扫描JSON数据失败")
	}

	// Return 使用独立的退款记录表
	cachedRecord := &entity.ReceiveReturnSalesInvoice{}
	returnDao := dao.ReceiveReturnSalesInvoice.Ctx(ctx).WherePri(msg.RecordId)
	if err = returnDao.Scan(&cachedRecord); err != nil {
		return gerror.Wrap(err, "查询退款SI异步记录失败")
	}

	// 幂等性检查：基于 mq_msg_id 避免重复处理（MQ 重试场景）
	if cachedRecord.MqMsgId == mqMsg.MsgId && cachedRecord.Docstatus == erp.DocstatusSubmitted {
		g.Log().Infof(ctx, "检测到重复消息（MQ重试），跳过: record_id=%d, mq_msg_id=%s", msg.RecordId, mqMsg.MsgId)
		return nil
	}

	reqBuf, err := gbase64.DecodeString(cachedRecord.ReqMessage)
	if err != nil {
		return gerror.Wrap(err, "解码请求参数失败")
	}
	req := &selling.ReturnSalesInvoiceReq{}
	if err = proto.Unmarshal(reqBuf, req); err != nil {
		return gerror.Wrap(err, "反序列化请求参数失败")
	}

	ctx = grpcx.Ctx.SetIncoming(ctx, g.Map{
		consts.ContextSiteCode: cachedRecord.SiteCode,
	})

	// 确保原 SI 已完成
	originalRecord, err := service.AsyncSelling().GetLatestReceiveSalesInvoice(ctx, req.SaleOrderUuid)
	if err != nil {
		return gerror.Wrapf(err, "查询原SI记录失败")
	}
	if originalRecord == nil || originalRecord.Id == 0 || originalRecord.Docstatus != erp.DocstatusSubmitted {
		// 检查原 SI 是否已失败（Draft 但有 RespBody）
		if originalRecord != nil && originalRecord.Docstatus == erp.DocstatusDraft && originalRecord.RespBody != "" {
			g.Log().Infof(ctx, "原SI已失败，无法退款: sale_order_uuid=%s", req.SaleOrderUuid)
			returnDao.Data(do.ReceiveReturnSalesInvoice{
				RespBody:  "原SI已失败，跳过退款",
				UpdatedAt: int(time.Now().Unix()),
			}).Update()
			return nil
		}
		// SI 尚未完成，延迟重试（业务等待，非失败重试）
		elapsed := time.Now().Unix() - int64(cachedRecord.CreatedAt)
		if elapsed > consts.MaxWaitForSISeconds {
			g.Log().Errorf(ctx, "退款SI超时: sale_order_uuid=%s", req.SaleOrderUuid)
			returnDao.Data(do.ReceiveReturnSalesInvoice{
				RespBody:  fmt.Sprintf("退款SI超时，原SI未完成: elapsed=%ds", elapsed),
				UpdatedAt: int(time.Now().Unix()),
			}).Update()
			return nil
		}
		queue.DelayPush(string(consts.TopicReturnSalesInvoice), &mq.AsyncSalesInvoiceMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeReturnSalesInvoice,
			SiteCode: msg.SiteCode,
		}, 5*time.Second)
		return nil
	}

	// 已成功则跳过（幂等）
	if cachedRecord.Docstatus == erp.DocstatusSubmitted {
		g.Log().Infof(ctx, "退款SI已提交，跳过: record_id=%d", msg.RecordId)
		return nil
	}

	resp, err := service.Selling().ReturnSalesInvoice(ctx, req)
	if err != nil {
		// 应用层重试：使用数据库 retry_count 字段追踪重试次数
		retryCount := cachedRecord.RetryCount + 1
		g.Log().Errorf(ctx, "退款SI失败 (retry=%d/%d): %v", retryCount, maxRetryCount, err)

		// 更新数据库重试计数和错误信息
		returnDao.Data(do.ReceiveReturnSalesInvoice{
			RetryCount: retryCount,
			MqMsgId:    mqMsg.MsgId,
			RespBody:   fmt.Sprintf(consts.ErrMsgReturnSIRetry, retryCount, err),
			UpdatedAt:  int(time.Now().Unix()),
		}).Update()

		if retryCount < maxRetryCount {
			// 延迟重试：使用指数退避策略重新推入队列
			queue.DelayPush(string(consts.TopicReturnSalesInvoice), &mq.AsyncSalesInvoiceMsg{
				RecordId: msg.RecordId,
				MsgType:  mq.MsgTypeReturnSalesInvoice,
				SiteCode: msg.SiteCode,
			}, retryDelay(retryCount))
			return nil // 不返回 error，避免 RocketMQ 自身重试
		}

		// 重试耗尽：1. 发送失败回调到 Main，2. 推送到 DLQ 作为兜底
		companyUuid, orderType := ExtractReqMetadata(originalRecord)
		SendSalesInvoiceCallback(ctx, companyUuid, req.SaleOrderUuid, erp.SyncStatusFailed, "", "", err.Error(), orderType)
		PushToDLQ(ctx, consts.TopicReturnSalesInvoiceDLQ, &mq.AsyncSalesInvoiceMsg{
			RecordId: msg.RecordId,
			MsgType:  mq.MsgTypeReturnSalesInvoice,
			SiteCode: msg.SiteCode,
		}, retryCount, err.Error(), mqMsg.MsgId, companyUuid, orderType, req.SaleOrderUuid)
		return nil // 已处理失败，不需要 MQ 重试
	}

	if resp != nil {
		respBuf, _ := proto.Marshal(resp)
		peNamesJson, _ := gjson.Encode(resp.PaymentEntryNames)
		if _, updateErr := returnDao.Data(do.ReceiveReturnSalesInvoice{
			Docstatus:         erp.DocstatusSubmitted,
			SalesInvoiceName:  resp.CreditNoteName,
			PaymentEntryNames: string(peNamesJson),
			RespMessage:       gbase64.EncodeToString(respBuf),
			RespBody:          resp.String(),
			MqMsgId:           mqMsg.MsgId,
			UpdatedAt:         int(time.Now().Unix()),
		}).Update(); updateErr != nil {
			return gerror.Wrapf(updateErr, "更新退款SI日志记录失败")
		}
	}

	return nil
}

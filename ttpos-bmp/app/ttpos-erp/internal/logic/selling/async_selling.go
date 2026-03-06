package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/dao"
	"ttpos-bmp/app/ttpos-erp/internal/model/do"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	"ttpos-bmp/app/ttpos-erp/internal/model/entity"
	"ttpos-bmp/app/ttpos-erp/internal/model/mq"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/internal/pkg/queue"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"google.golang.org/protobuf/proto"
)

var (
	AsyncSelling = &sAsyncSelling{}
)

type sAsyncSelling struct {
}

func init() {
	service.RegisterAsyncSelling(AsyncSelling)
}

func (s *sAsyncSelling) CancelPosInvoice(ctx context.Context, req *selling.CancelPosInvoiceReq) (asyncRecordId string, err error) {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "取消发票失败，序列化请求参数失败: %v", req)
		return "", gerror.Wrapf(err, "取消发票失败，序列化请求参数失败: %v", req)
	}
	reqMsg := gbase64.EncodeToString(buf)

	//设置siteCode
	siteCode := service.Rpc().GetSiteCode(ctx)

	//暂存请求信息
	recordId, insertErr := dao.ReceiveCancelPosInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceiveCancelPosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		ReqBody:          req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})
	if insertErr != nil {
		g.Log().Errorf(ctx, "取消发票失败，插入记录失败: %v", insertErr)
		return "", gerror.Wrapf(insertErr, "取消发票失败，插入记录失败: %v", req)
	}

	//发送消息
	if err = queue.Push(string(consts.TopicCancelPosInvoice), &mq.AsyncSellingMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeCancelPosInvoice,
	}); err != nil {
		g.Log().Errorf(ctx, "取消发票失败，发送异步消息失败: %v", err)
		return "", gerror.Wrapf(err, "取消发票失败，发送异步消息失败: %v", req)
	}

	return gconv.String(recordId), nil
}

func (*sAsyncSelling) SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，序列化请求参数失败: %v", req)
		return nil, gerror.Wrapf(err, "保存发票失败，序列化请求参数失败: %v", req)
	}
	reqMsg := gbase64.EncodeToString(buf)

	//设置siteCode
	siteCode := service.Rpc().GetSiteCode(ctx)

	//检查是否已存在
	count, err := dao.ReceivePosInvoice.Ctx(ctx).Count(&do.ReceivePosInvoice{
		OrderNo:   req.OrderNo,
		Docstatus: erp.DocstatusSubmitted,
	})
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，查询记录失败: %v", err)
		return nil, gerror.Wrapf(err, "保存发票失败，查询记录失败: %v", err)
	}
	if req.Remark != consts.RemarkBatchRedo && count > 0 {
		return nil, gerror.New("订单发票已存在,请勿重复下单")
	}

	//暂存请求信息
	recordId, err := dao.ReceivePosInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceivePosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		Branch:           req.Branch,
		PostingDatetime:  req.PostingDatetime,
		ReqBody:          req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，插入记录失败: %v", err)
		return nil, gerror.Wrapf(err, "保存发票失败，插入记录失败: %v", err)
	}

	//发送消息
	if err = queue.Push(string(consts.TopicSavePosInvoice), &mq.AsyncSellingMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeSavePosInvoice,
	}); err != nil {
		g.Log().Errorf(ctx, "保存发票失败，发送异步消息失败: %v", err)
		return nil, gerror.Wrapf(err, "保存发票失败，发送异步消息失败: %v", req)
	}

	return &selling.SavePosInvoiceResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

func (s *sAsyncSelling) ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error) {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "退款发票失败，序列化请求参数失败: %v", req)
		return nil, gerror.Wrapf(err, "退款发票失败，序列化请求参数失败: %v", req)
	}
	reqMsg := gbase64.EncodeToString(buf)

	//设置siteCode
	siteCode := service.Rpc().GetSiteCode(ctx)

	//暂存请求信息
	recordId, err := dao.ReceiveReturnPosInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceiveReturnPosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		ReqBody:          req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})
	if err != nil {
		g.Log().Errorf(ctx, "保存退款发票失败，插入记录失败: %v", err)
		return nil, gerror.Wrapf(err, "保存退款发票失败，插入记录失败: %v", err)
	}

	//发送消息
	if err = queue.Push(string(consts.TopicReturnPosInvoice), &mq.AsyncSellingMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeReturnPosInvoice,
	}); err != nil {
		g.Log().Errorf(ctx, "退款发票失败，发送异步消息失败: %v", err)
		return nil, gerror.Wrapf(err, "退款发票失败，发送异步消息失败: %v", req)
	}

	return &selling.ReturnPosInvoiceResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

func (*sAsyncSelling) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error) {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "关帐失败，序列化请求参数失败: %v", req)
		return nil, gerror.Wrapf(err, "关帐失败，序列化请求参数失败: %v", req)
	}
	reqMsg := gbase64.EncodeToString(buf)
	//设置siteCode
	siteCode := service.Rpc().GetSiteCode(ctx)

	//暂存请求信息
	recordId, err := dao.ReceiveClosePos.Ctx(ctx).InsertAndGetId(&entity.ReceiveClosePos{
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		PosOpenEntryName: req.PosOpenEntryName,
		PeriodEndDate:    req.PeriodEndDate,
		ReqBody:          req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})
	if err != nil {
		g.Log().Errorf(ctx, "关帐失败，插入记录失败: %v", err)
		return nil, gerror.Wrapf(err, "关帐失败，插入记录失败: %v", err)
	}
	//发送消息
	if err = queue.Push(string(consts.TopicClosePosEntry), &mq.AsyncSellingMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeClosePosEntry,
	}); err != nil {
		g.Log().Errorf(ctx, "关帐失败，发送异步消息失败: %v", err)
		return nil, gerror.Wrapf(err, "关帐失败，发送异步消息失败: %v", req)
	}

	return &selling.ClosePosEntryResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

func (*sAsyncSelling) GetLatestReceivePosInvoice(ctx context.Context, req *do.ReceivePosInvoice) (*entity.ReceivePosInvoice, error) {
	receivePosInvoice := &entity.ReceivePosInvoice{}
	err := dao.ReceivePosInvoice.Ctx(ctx).Where(req).
		OrderDesc("id").Limit(1).Scan(&receivePosInvoice)
	if err != nil {
		g.Log().Errorf(ctx, "查询发票失败，查询记录失败: %v", err)
		return nil, gerror.Wrapf(err, "查询发票失败，查询记录失败: %v", err)
	}
	return receivePosInvoice, nil
}

// ========== Sales Invoice 异步方法 ==========

func (*sAsyncSelling) SaveSalesInvoice(ctx context.Context, req *selling.SaveSalesInvoiceReq) (*selling.SaveSalesInvoiceResp, error) {
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "保存SI失败，序列化请求参数失败: %v", req)
		return nil, gerror.Wrapf(err, "保存SI失败，序列化请求参数失败")
	}
	reqMsg := gbase64.EncodeToString(buf)
	siteCode := service.Rpc().GetSiteCode(ctx)

	// 幂等检查: order_no 已提交则跳过
	count, err := dao.ReceiveSalesInvoice.Ctx(ctx).Where(do.ReceiveSalesInvoice{
		OrderNo:   req.OrderNo,
		Docstatus: erp.DocstatusSubmitted,
	}).Count()
	if err != nil {
		return nil, gerror.Wrapf(err, "保存SI失败，查询记录失败")
	}
	if count > 0 {
		return nil, gerror.New("Sales Invoice已存在，请勿重复提交")
	}

	recordId, err := dao.ReceiveSalesInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceiveSalesInvoice{
		OrderNo:         req.OrderNo,
		SaleOrderUuid:   req.SaleOrderUuid,
		PosProfile:      req.PosProfile,
		PostingDatetime: req.PostingDatetime,
		Docstatus:       erp.DocstatusDraft,
		SiteCode:        siteCode,
		ReqMessage:      reqMsg,
		ReqBody:         req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "保存SI失败，插入记录失败")
	}

	if err = queue.Push(string(consts.TopicSaveSalesInvoice), &mq.AsyncSalesInvoiceMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeSaveSalesInvoice,
		SiteCode: siteCode,
	}); err != nil {
		return nil, gerror.Wrapf(err, "保存SI失败，发送异步消息失败")
	}

	return &selling.SaveSalesInvoiceResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

func (*sAsyncSelling) CancelSalesInvoice(ctx context.Context, req *selling.CancelSalesInvoiceReq) (*selling.CancelSalesInvoiceResp, error) {
	siteCode := service.Rpc().GetSiteCode(ctx)

	// 查找原始 SI 记录，直接复用，不插入新记录
	originalRecord, err := service.AsyncSelling().GetLatestReceiveSalesInvoice(ctx, req.SaleOrderUuid)
	if err != nil {
		return nil, gerror.Wrapf(err, "取消SI失败，查询原SI记录失败")
	}
	if originalRecord == nil || originalRecord.Id == 0 {
		return nil, gerror.Newf("取消SI失败，原SI记录不存在: sale_order_uuid=%s", req.SaleOrderUuid)
	}

	if err = queue.Push(string(consts.TopicCancelSalesInvoice), &mq.AsyncSalesInvoiceMsg{
		RecordId: originalRecord.Id,
		MsgType:  mq.MsgTypeCancelSalesInvoice,
		SiteCode: siteCode,
	}); err != nil {
		return nil, gerror.Wrapf(err, "取消SI失败，发送异步消息失败")
	}

	return &selling.CancelSalesInvoiceResp{
		AsyncRecordId: gconv.String(originalRecord.Id),
	}, nil
}

func (*sAsyncSelling) ReturnSalesInvoice(ctx context.Context, req *selling.ReturnSalesInvoiceReq) (*selling.ReturnSalesInvoiceResp, error) {
	// Return 需要存储退款请求明细（退哪些商品、金额），因此仍需插入记录
	buf, err := proto.Marshal(req)
	if err != nil {
		return nil, gerror.Wrapf(err, "退款SI失败，序列化请求参数失败")
	}
	reqMsg := gbase64.EncodeToString(buf)
	siteCode := service.Rpc().GetSiteCode(ctx)

	// 插入独立的退款记录表（与 POS Invoice 模式一致）
	recordId, err := dao.ReceiveReturnSalesInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceiveReturnSalesInvoice{
		OrderNo:       req.OrderNo,
		SaleOrderUuid: req.SaleOrderUuid,
		Docstatus:     erp.DocstatusDraft,
		SiteCode:      siteCode,
		ReqMessage:    reqMsg,
		ReqBody:       req.String(),
		CreatedAt:     int(gtime.Now().Timestamp()),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "退款SI失败，插入记录失败")
	}

	if err = queue.Push(string(consts.TopicReturnSalesInvoice), &mq.AsyncSalesInvoiceMsg{
		RecordId: recordId,
		MsgType:  mq.MsgTypeReturnSalesInvoice,
		SiteCode: siteCode,
	}); err != nil {
		return nil, gerror.Wrapf(err, "退款SI失败，发送异步消息失败")
	}

	return &selling.ReturnSalesInvoiceResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

func (*sAsyncSelling) GetLatestReceiveSalesInvoice(ctx context.Context, saleOrderUuid string) (*entity.ReceiveSalesInvoice, error) {
	record := &entity.ReceiveSalesInvoice{}
	err := dao.ReceiveSalesInvoice.Ctx(ctx).Where(do.ReceiveSalesInvoice{
		SaleOrderUuid: saleOrderUuid,
	}).OrderDesc("id").Limit(1).Scan(&record)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询SI记录失败")
	}
	return record, nil
}


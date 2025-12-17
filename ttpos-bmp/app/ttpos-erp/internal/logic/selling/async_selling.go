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

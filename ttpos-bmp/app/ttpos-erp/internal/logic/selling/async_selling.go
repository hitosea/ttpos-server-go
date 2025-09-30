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
	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/contrib/rpc/grpcx/v2"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
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
	recordId, _ := dao.ReceiveCancelPosInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceiveCancelPosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		ReqBody:          req.String(),
		CreatedAt:        int(gtime.Now().Timestamp()),
	})

	go func() {

		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})

		cancelDao := dao.ReceiveCancelPosInvoice.Ctx(ctx).WherePri(recordId)
		receivePosInvoice, err := s.GetLatestReceivePosInvoice(ctx, &do.ReceivePosInvoice{
			OrderNo:          req.OrderNo,
			OpenPosEntryName: req.OpenPosEntryName,
			Docstatus:        erp.DocstatusSubmitted, // 取已提交的记录来取消
		})
		if err != nil {
			respMessage := fmt.Sprintf("取消发票失败，查询原POS记录失败: %v", req)
			g.Log().Errorf(ctx, respMessage, err)
			if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
				RespBody: respMessage,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "取消发票失败，更新日志记录失败: %v", err)
				return
			}
			return
		}
		if receivePosInvoice != nil {
			// 调用服务层取消商品发票
			if len(receivePosInvoice.ProductsInvoiceName) > 0 {
				err := service.Selling().CancelPosInvoice(ctx, receivePosInvoice.ProductsInvoiceName)
				if err != nil {
					respMessage := fmt.Sprintf("取消商品发票[%s]失败", receivePosInvoice.ProductsInvoiceName)
					g.Log().Errorf(ctx, respMessage, err)
					if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
						RespBody: respMessage,
					}).Update(); err != nil {
						g.Log().Errorf(ctx, "取消发票失败，更新日志记录失败: %v", err)
						return
					}
				}
			}
			// 调用服务层取消材料发票
			if len(receivePosInvoice.MaterialInvoiceName) > 0 {
				err := service.Selling().CancelPosInvoice(ctx, receivePosInvoice.MaterialInvoiceName)
				if err != nil {
					respMessage := fmt.Sprintf("取消材料发票[%s]失败", receivePosInvoice.ProductsInvoiceName)
					g.Log().Errorf(ctx, respMessage, err)
					if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
						RespBody: respMessage,
					}).Update(); err != nil {
						g.Log().Errorf(ctx, "取消材料发票失败，更新日志记录失败: %v", err)
						return
					}
				}
			}
			if _, err := dao.ReceivePosInvoice.Ctx(ctx).WherePri(receivePosInvoice.Id).Data(do.ReceivePosInvoice{
				Docstatus: erp.DocstatusCancelled,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "取消发票失败，更新原下单发票日志记录失败: %v", err)
				return
			}
			if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
				Docstatus: erp.DocstatusSubmitted,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "取消发票失败，更新日志记录失败: %v", err)
				return
			}
		}
	}()

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
	if count > 0 {
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

	//异步保存发票
	go func() {
		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})
		posInvoiceDao := dao.ReceivePosInvoice.Ctx(ctx).WherePri(recordId)
		resp, err := service.Selling().SavePosInvoice(ctx, req)
		if err != nil {
			g.Log().Errorf(ctx, "保存发票失败，异步保存发票失败: %v", err)
			if _, err := posInvoiceDao.Data(do.ReceivePosInvoice{
				RespBody: fmt.Sprintf("保存发票失败，异步保存发票失败: %v", err),
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
				return
			}
			return
		}
		if resp != nil {
			respBuf, err := proto.Marshal(resp)
			if err != nil {
				g.Log().Errorf(ctx, "保存发票失败，序列化响应参数失败: %v", resp)
				return
			}
			if _, err := posInvoiceDao.Data(do.ReceivePosInvoice{
				Docstatus:           erp.DocstatusSubmitted,
				ProductsInvoiceName: resp.ProductsInvoiceName,
				MaterialInvoiceName: resp.MaterialInvoiceName,
				RespMessage:         gbase64.EncodeToString(respBuf),
				RespBody:            resp.String(),
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
				return
			}
		}
	}()

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

	go func() {
		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})
		returnDao := dao.ReceiveReturnPosInvoice.Ctx(ctx).WherePri(recordId)
		receivePosInvoice, err := s.GetLatestReceivePosInvoice(ctx, &do.ReceivePosInvoice{
			OrderNo:          req.OrderNo,
			OpenPosEntryName: req.OpenPosEntryName,
		})
		if err != nil {
			respMessage := fmt.Sprintf("退款发票失败，查询原POS记录失败: %v", req)
			g.Log().Errorf(ctx, respMessage, err)
			if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
				RespBody: respMessage,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "退款发票失败，更新日志记录失败: %v", err)
				return
			}
			return
		}
		//开始退款
		if receivePosInvoice == nil {
			respMessage := fmt.Sprintf("退款发票失败，查询原POS记录失败: %v", req)
			g.Log().Errorf(ctx, respMessage, err)
			if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
				RespBody: respMessage,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "退款发票失败，更新日志记录失败: %v", err)
				return
			}
			return
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
				g.Log().Errorf(ctx, "保存发票失败，异步保存发票失败: %v", err)
				if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
					RespBody: fmt.Sprintf("保存发票失败，异步保存发票失败: %v", err),
				}).Update(); err != nil {
					g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
					return
				}
				return
			}
			if resp != nil {
				respBuf, err := proto.Marshal(resp)
				if err != nil {
					g.Log().Errorf(ctx, "保存发票失败，序列化响应参数失败: %v", resp)
					return
				}
				if _, err := returnDao.Data(do.ReceiveReturnPosInvoice{
					Docstatus:   erp.DocstatusSubmitted,
					RespMessage: gbase64.EncodeToString(respBuf),
					RespBody:    resp.String(),
				}).Update(); err != nil {
					g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
					return
				}
			}
		}
	}()

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

	go func() {
		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})
		closePosDao := dao.ReceiveClosePos.Ctx(ctx).WherePri(recordId)
		resp, err := service.Selling().ClosePosEntry(ctx, req)
		if err != nil {
			g.Log().Errorf(ctx, "关帐失败，异步关帐失败: %v", err)
			if _, err := closePosDao.Data(do.ReceiveClosePos{
				RespBody: fmt.Sprintf("保存发票失败，异步保存发票失败: %v", err),
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
				return
			}
			return
		}
		if resp != nil {
			respBuf, err := proto.Marshal(resp)
			if err != nil {
				g.Log().Errorf(ctx, "保存发票失败，序列化响应参数失败: %v", resp)
				return
			}
			if _, err := closePosDao.Data(do.ReceiveClosePos{
				Docstatus:   erp.DocstatusSubmitted,
				RespMessage: gbase64.EncodeToString(respBuf),
				RespBody:    resp.String(),
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
				return
			}
		}
	}()

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

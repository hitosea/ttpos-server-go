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

func (*sAsyncSelling) AsyncCancelPosInvoice(ctx context.Context, req *selling.CancelPosInvoiceReq) error {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "取消发票失败，序列化请求参数失败: %v", req)
		return gerror.Wrapf(err, "取消发票失败，序列化请求参数失败: %v", req)
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
	})

	go func() {

		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})

		cancelDao := dao.ReceiveCancelPosInvoice.Ctx(ctx).WherePri(recordId)

		receivePosInvoice := &entity.ReceivePosInvoice{}
		err := dao.ReceivePosInvoice.Ctx(ctx).Where(do.ReceivePosInvoice{
			OrderNo:          req.OrderNo,
			OpenPosEntryName: req.OpenPosEntryName,
		}).Scan(&receivePosInvoice)

		if err != nil {
			respMessage := fmt.Sprintf("取消发票失败，查询原POS记录失败: %v", req)
			g.Log().Errorf(ctx, respMessage, err)
			if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
				RespMessage: respMessage,
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
						RespMessage: respMessage,
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
						RespMessage: respMessage,
					}).Update(); err != nil {
						g.Log().Errorf(ctx, "取消材料发票失败，更新日志记录失败: %v", err)
						return
					}
				}
			}
			if _, err := cancelDao.Data(do.ReceiveCancelPosInvoice{
				Docstatus: erp.DocstatusSubmitted,
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "取消发票失败，更新日志记录失败: %v", err)
				return
			}
		}
	}()

	return nil
}

func (*sAsyncSelling) AsyncSavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
	// 异步模式
	buf, err := proto.Marshal(req)
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，序列化请求参数失败: %v", req)
		return nil, gerror.Wrapf(err, "保存发票失败，序列化请求参数失败: %v", req)
	}
	reqMsg := gbase64.EncodeToString(buf)

	//设置siteCode
	siteCode := service.Rpc().GetSiteCode(ctx)

	//暂存请求信息
	recordId, err := dao.ReceivePosInvoice.Ctx(ctx).InsertAndGetId(&entity.ReceivePosInvoice{
		OrderNo:          req.OrderNo,
		OpenPosEntryName: req.OpenPosEntryName,
		Docstatus:        erp.DocstatusDraft,
		ReqMessage:       reqMsg,
		SiteCode:         siteCode,
		Branch:           req.Branch,
		PostingDatetime:  req.PostingDatetime,
	})
	if err != nil {
		g.Log().Errorf(ctx, "保存发票失败，插入记录失败: %v", err)
		return nil, gerror.Wrapf(err, "保存发票失败，插入记录失败: %v", err)
	}

	//异步保存发票
	go func(mainCtx context.Context) {
		//设置siteCode
		ctx := grpcx.Ctx.SetIncoming(gctx.New(), g.Map{
			consts.ContextSiteCode: siteCode,
		})
		posInvoiceDao := dao.ReceivePosInvoice.Ctx(ctx).WherePri(recordId)
		resp, err := service.Selling().SavePosInvoice(ctx, req)
		if err != nil {
			g.Log().Errorf(ctx, "保存发票失败，异步保存发票失败: %v", err)
			if _, err := posInvoiceDao.Data(do.ReceivePosInvoice{
				RespMessage: fmt.Sprintf("保存发票失败，异步保存发票失败: %v", err),
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
			}).Update(); err != nil {
				g.Log().Errorf(ctx, "保存发票失败，更新日志记录失败: %v", err)
				return
			}
		}
	}(ctx)

	return &selling.SavePosInvoiceResp{
		AsyncRecordId: gconv.String(recordId),
	}, nil
}

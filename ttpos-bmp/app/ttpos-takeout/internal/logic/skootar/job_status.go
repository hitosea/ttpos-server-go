package skootar

import (
	"context"
	v1 "ttpos-bmp/app/ttpos-takeout/api/callback/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
)

func (s *sSkootar) JobStatusChange(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error) {
	var (
		order             *entity.Order
		orderStatusBefore string
		jobDetail         *skootar.JobDetail
		orderModel        *gdb.Model
	)

	// 使用事务确保数据一致性
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 记录回调消息
		if _, err = dao.CallbackMsg.Ctx(ctx).TX(tx).Data(&do.CallbackMsg{
			Uuid:           guid.S(),
			TakeoutRefNo:   req.JobId,
			StatusDatetime: gtime.NewFromTime(req.StatusDatetime),
			Content:        gjson.MustEncodeString(req),
		}).Insert(); err != nil {
			return err
		}

		// 2. 查询主表订单
		orderModel = dao.Order.Ctx(ctx).TX(tx).Where(do.Order{
			PartnerOrderId: req.JobId,
			ProviderName:   consts.ProviderSkootar,
		})
		if err = orderModel.Scan(&order); err != nil {
			return err
		}
		if order == nil {
			return gerror.Newf("外送任务不存在")
		}
		orderStatusBefore = order.OrderStatus

		// 3. 记录状态变更日志
		if _, err = dao.OrderStatusLog.Ctx(ctx).TX(tx).Data(&do.OrderStatusLog{
			Uuid:         guid.S(),
			OrderUuid:    order.Uuid,
			ProviderName: consts.ProviderSkootar,
			StatusBefore: orderStatusBefore,
			StatusAfter:  gconv.String(req.StatusAfter),
			ChangeSource: "WEBHOOK",
		}).Insert(); err != nil {
			return err
		}

		// 4. 更新主表订单状态
		if _, err = orderModel.Data(do.Order{
			OrderStatus: gconv.String(req.StatusAfter),
			UpdatedAt:   gtime.Now(),
		}).Update(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, gerror.Wrap(err, "更新任务状态失败")
	}

	// 如果状态变更为 5 Assigned，则查询订单详情获取骑手信息并更新扩展表
	if req.StatusAfter == int(consts.JobStatusAssigned) {
		jobDetail, err = s.JobDetail4Food(ctx, &skootar.JobDetailInp{
			JobId: req.JobId,
		})
		if err != nil {
			g.Log().Errorf(ctx, "获取订单详情失败:%v", err)
			return nil, gerror.Wrap(err, "获取订单详情失败")
		}

		// 更新扩展表中的骑手信息
		if _, err = dao.OrderSkootar.Ctx(ctx).Where(do.OrderSkootar{
			OrderUuid: order.Uuid,
		}).Data(do.OrderSkootar{
			SkootarId:       jobDetail.SkootarId,
			SkootarName:     jobDetail.SkootarName,
			SkootarPhone:    jobDetail.SkootarPhone,
			SkootarImageUrl: jobDetail.SkootarImageUrl,
			SkootarRating:   jobDetail.SkootarRating,
			UpdatedAt:       gtime.Now(),
		}).Update(); err != nil {
			return nil, gerror.Wrap(err, "更新骑手信息失败")
		}
	}

	// 回调 ttpos (保持原有逻辑，需要从 order 中获取 callback_url)
	// 注意：新表结构中没有 callback_url 字段，需要从其他地方获取或移除此逻辑
	// 暂时注释掉，等待确认业务需求
	/*
		if len(order.CallbackUrl) > 0 {
			go func() {
				callbackRes := &gclient.Response{}
				if callbackRes, err = g.Client().SetHeader(consts.TTPOS_HEADER_CALLBACK_AUTH, s.getCallBackAuth(order.PartnerOrderId)).ContentJson().Post(gctx.New(), order.CallbackUrl, g.Map{
					"shopRefNo":       order.PartnerOrderId,
					"takeoutRefNo":    order.PartnerOrderId,
					"providerName":    order.ProviderName,
					"jobStatusBefore": orderStatusBefore,
					"jobStatusAfter":  gconv.String(req.StatusAfter),
				}); err != nil {
					g.Log().Infof(ctx, "发起回调ttpos异常: %v", err)
				}
				defer callbackRes.Close()
				callbackRes.RawDump()
			}()
		}
	*/

	return &v1.SkootarStatusRes{
		Code:    "200",
		Message: "订单状态已更新",
	}, nil
}

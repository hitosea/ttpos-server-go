package skootar

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/guid"
	v1 "ttpos-bmp/app/ttpos-takeout/api/callback/v1"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"
)

func (s *sSkootar) JobStatusChange(ctx context.Context, req *v1.SkootarStatusReq) (res *v1.SkootarStatusRes, err error) {
	var (
		job             *entity.Job
		jobStatusBefore string
		jobDetail       *skootar.JobDetail
		jobModel        *gdb.Model
	)

	//先做入库，后面再说
	err = dao.Job.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err = dao.CallbackMsg.Ctx(ctx).Data(&do.CallbackMsg{
			Uuid:           guid.S(),
			TakeoutRefNo:   req.JobId,
			StatusDatetime: gtime.NewFromTime(req.StatusDatetime),
			Content:        gjson.MustEncodeString(req),
		}).Insert(); err != nil {
			return err
		}

		//状态变更
		jobModel = dao.Job.Ctx(ctx).Where(do.Job{
			TakeoutRefNo: req.JobId,
			ProviderName: consts.ProviderSkootar,
		})
		if err = jobModel.Scan(&job); err != nil {
			return err
		}
		if job == nil {
			return gerror.Newf("外送任务不存在")
		}
		jobStatusBefore = job.JobStatus
		//状态变更, 后续需考虑兼容其它供应商
		if _, err = dao.JobStatusLog.Ctx(ctx).Data(&do.JobStatusLog{
			Uuid:         guid.S(),
			JobUuid:      job.Uuid,
			StatusBefore: jobStatusBefore,
			StatusAfter:  gconv.String(req.StatusAfter),
		}).Insert(); err != nil {
			return err
		}

		if _, err = jobModel.Data(do.Job{
			JobStatus: req.StatusAfter,
		}).Update(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, gerror.Wrap(err, "更新任务状态失败")
	}
	//如果状态变更为 5 Assigned, 则查询订单详情获取skootar骑手信息
	if req.StatusAfter == int(consts.JobStatusAssigned) {
		jobDetail, err = s.JobDetail4Food(ctx, &skootar.JobDetailInp{
			JobId: req.JobId,
		})
		if err != nil {
			g.Log().Errorf(ctx, "获取订单详情失败:%v", err)
			return nil, gerror.Wrap(err, "获取订单详情失败")
		}
		//更新骑手信息
		if _, err = jobModel.Data(do.Job{
			SkootarId:       jobDetail.SkootarId,
			SkootarName:     jobDetail.SkootarName,
			SkootarPhone:    jobDetail.SkootarPhone,
			SkootarImageUrl: jobDetail.SkootarImageUrl,
			SkootarRating:   jobDetail.SkootarRating,
		}).Update(); err != nil {
			return nil, gerror.Wrap(err, "更新骑手信息失败")
		}
	}
	//回调ttpos
	if len(job.CallbackUrl) > 0 {
		//后面改成异步mq
		//报文参考 `{"jobStatusAfter":"5","jobStatusBefore":"5","providerName":"skootar","shopRefNo":"8888777","takeoutRefNo":"J25070856667"}`
		go func() {
			callbackRes := &gclient.Response{}
			if callbackRes, err = g.Client().SetHeader(consts.TTPOS_HEADER_CALLBACK_AUTH, s.getCallBackAuth(job.ShopRefNo)).ContentJson().Post(gctx.New(), job.CallbackUrl, g.Map{
				"shopRefNo":       job.ShopRefNo,
				"takeoutRefNo":    job.TakeoutRefNo,
				"providerName":    job.ProviderName,
				"jobStatusBefore": jobStatusBefore,
				"jobStatusAfter":  gconv.String(req.StatusAfter),
			}); err != nil {
				g.Log().Infof(ctx, "发起回调ttpos异常: %v", err)
			}
			defer callbackRes.Close()
			callbackRes.RawDump()
		}()
	}
	return &v1.SkootarStatusRes{
		Code:    "200",
		Message: "订单状态已更新",
	}, nil
}

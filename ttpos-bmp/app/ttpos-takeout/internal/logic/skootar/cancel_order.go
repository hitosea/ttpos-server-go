package skootar

import (
	"context"
	"encoding/json"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var cancelOrderApiPath = "/api/cancel_created_job"

// CancelOrder 取消订单
func (s *sSkootar) CancelOrder(ctx context.Context, req *dto.CancelOrderInp) (res *api.CancelOrderResp, err error) {
	reqInp := &skootar.CancelOrderInp{
		ReqBase:      s.ReqBase(),
		JobId:        req.JobId,
		CancelReason: req.Reason,
	}
	resp := &skootar.CancelOrderOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(cancelOrderApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("取消订单失败:%+v", reqInp)
	}
	if err = json.Unmarshal(rr.Bytes(), resp); err != nil {
		return nil, gerror.Newf("取消订单失败:%+v", err)
	}

	if _, err = dao.Job.Ctx(ctx).Where(do.Job{TakeoutRefNo: req.JobId, ProviderName: consts.ProviderSkootar}).
		Data(do.Job{JobStatus: consts.JobStatusCanceled}).Update(); err != nil {
		return nil, gerror.Newf("取消订单失败:%+v", err)
	}
	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("取消订单异常:%v", resp.ResponseDesc)
	}
	// 解析接口返回结果·
	res = &api.CancelOrderResp{}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "取消订单")
	}
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

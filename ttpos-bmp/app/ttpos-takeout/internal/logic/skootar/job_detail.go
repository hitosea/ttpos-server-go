package skootar

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
)

var (
	jobDetailFoodApi = "/partner/api/v1/job"
	jobDetailApi     = "/api/get_job_detail"
)

func (s *sSkootar) JobDetail4Food(ctx context.Context, req *skootar.JobDetailInp) (jobDetail *skootar.JobDetail, err error) {
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(jobDetailFoodApi), &skootar.JobDetailReq{
		ReqBase: s.ReqBase(),
		JobId:   req.JobId,
	})
	resp := &skootar.JobDetailResp{}
	if rr == nil {
		return nil, gerror.Newf("获取订单详情失败:%+v", req)
	}
	json.Unmarshal(rr.Bytes(), resp)
	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("获取订单详情异常:%v", resp.ResponseDesc)
	}

	return &resp.JobDetail, err
}

package skootar

import (
	"context"
	"encoding/json"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var merchantConfirmApiPath = "/partner/api/v1/job"

// ConfirmOrder 商家确认订单
func (s *sSkootar) ConfirmOrder(ctx context.Context, req *dto.ConfirmOrderInp) (res *api.ConfirmOrderResp, err error) {
	reqInp := &skootar.ConfirmOrderInp{
		ReqBase: s.ReqBase(),
		JobId:   req.JobId,
	}
	resp := &skootar.ConfirmOrderOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(merchantConfirmApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("商家确认订单失败:%+v", reqInp)
	}
	if err := json.Unmarshal(rr.Bytes(), resp); err != nil {
		return nil, gerror.Newf("商家确认订单失败:%+v", err)
	}
	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("商家确认订单异常:%v", resp.ResponseDesc)
	}
	// 解析接口返回结果·
	res = &api.ConfirmOrderResp{}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "商家确认订单")
	}
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

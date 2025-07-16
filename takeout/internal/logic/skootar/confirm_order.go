package skootar

import (
	"context"
	"encoding/json"
	"takeout/api"
	"takeout/internal/model/input/skootar"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var merchantConfirmApiPath = "/partner/api/v1/job"

// CreateOrder 创建订单
func (s *sSkootar) ConfirmOrder(ctx context.Context, req *api.ConfirmOrderReq) (res *api.ConfirmOrderResp, err error) {
	reqInp := &skootar.ConfirmOrderInp{
		ReqBase: s.ReqBase(),
		JobId:   req.OrderId,
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
	// ToDo 更新订单状态
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

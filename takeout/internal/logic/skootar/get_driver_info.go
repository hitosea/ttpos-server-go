package skootar

import (
	"context"
	"encoding/json"
	"takeout/api"
	"takeout/internal/model/input"
	"takeout/internal/model/input/skootar"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var trackingDriverApiPath = "/api/tracking_driver"

// CreateOrder 创建订单
func (s *sSkootar) GetDriverInfo(ctx context.Context, req *input.GetDriverInfoInp) (res *api.GetDriverInfoResp, err error) {
	reqInp := &skootar.GetDriverLocationInp{
		ReqBase:   s.ReqBase(),
		SKootarId: req.SKootarId,
	}
	resp := &skootar.GetDriverLocationOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(trackingDriverApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("获取司机位置失败:%+v", reqInp)
	}

	if err := json.Unmarshal(rr.Bytes(), resp); err != nil {
		return nil, gerror.Newf("获取司机位置失败:%+v", err)
	}
	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("获取司机位置异常:%v", resp.ResponseDesc)
	}
	res = &api.GetDriverInfoResp{
		Lat:    float32(resp.Lat),
		Lng:    float32(resp.Lng),
		Name:   req.Name,
		Phone:  req.Phone,
		Avatar: req.Avatar,
		Rating: float32(req.Rating),
	}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "获取司机位置")
	}
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

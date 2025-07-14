package skootar

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"takeout/api"
	"takeout/internal/consts"
	"takeout/internal/model/input/skootar"
)

var apiPath = "/api/get_estimate_price"

/**
预估配送费
Estimate price based on pick up and delivery locations
*/

func (s *sSkootar) EstimatePrice(ctx context.Context, req *api.EstimatePriceReq) (res *api.EstimatePriceResp, err error) {
	var locationList []skootar.Location
	for i, address := range req.Address {
		locationList = append(locationList, skootar.Location{
			Lat: address.Lat,
			Lng: address.Lng,
			Seq: i + 1,
		})
	}
	//TODO 调整Opting 配置
	reqInp := &skootar.EstimatePriceInp{
		ReqBase:      s.ReqBase(),
		LocationList: locationList,
		JobType:      consts.JobTypeFood,
	}
	resp := &skootar.EstimatePriceOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(apiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("获取预估价格失败:%+v", reqInp)
	}
	json.Unmarshal(rr.Bytes(), resp)

	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("获取预估价格异常:%v", resp.ResponseDesc)
	}
	res = &api.EstimatePriceResp{}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "获取预估价格失败")
	}
	//TODO 自定返回值
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

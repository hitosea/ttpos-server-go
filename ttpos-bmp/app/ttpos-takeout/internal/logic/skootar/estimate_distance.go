package skootar

import (
	"context"
	"encoding/json"
	"fmt"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var estimateApiPath = "/api/get_estimate_price"

// EstimateDistance 获取预估距离
func (s *sSkootar) EstimateDistance(ctx context.Context, req *api.EstimateDistanceReq) (res *api.EstimateDistanceResp, err error) {
	var locationList []skootar.Location
	for i, address := range req.Address {
		locationList = append(locationList, skootar.Location{
			Lat: address.Lat,
			Lng: address.Lng,
			Seq: i + 1,
		})
	}
	reqInp := &skootar.EstimatePriceInp{
		ReqBase:      s.ReqBase(),
		LocationList: locationList,
		JobType:      consts.JobTypeFood,
	}
	resp := &skootar.EstimatePriceOut{}
	fmt.Println("11111 url:", s.GetUrl(estimateApiPath))
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(estimateApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("获取预估价格失败:%+v", reqInp)
	}
	fmt.Println("22222 rr:", rr.String())
	json.Unmarshal(rr.Bytes(), resp)

	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("获取预估价格异常:%v", resp.ResponseDesc)
	}
	res = &api.EstimateDistanceResp{}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "获取预估价格失败")
	}
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

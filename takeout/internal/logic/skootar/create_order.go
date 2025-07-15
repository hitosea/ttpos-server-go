package skootar

import (
	"context"
	"encoding/json"
	"takeout/api"
	"takeout/internal/consts"
	"takeout/internal/model/input/skootar"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var createNewJobApiPath = "/api/create_new_job"

// CreateOrder 创建订单
func (s *sSkootar) CreateOrder(ctx context.Context, req *api.CreateOrderReq) (res *api.CreateOrderResp, err error) {
	locationList := []skootar.Location{
		{
			AddressName:  req.MerchantLocation.AddressName,
			Address:      req.MerchantLocation.Address,
			Lat:          req.MerchantLocation.Lat,
			Lng:          req.MerchantLocation.Lng,
			CashFee:      consts.No, // TODO 预付费这个为NO
			ContactName:  req.MerchantLocation.ContactName,
			ContactPhone: req.MerchantLocation.ContactPhone,
			Seq:          1,
		}, {
			AddressName:  req.CustomerLocation.AddressName,
			Address:      req.CustomerLocation.Address,
			Lat:          req.CustomerLocation.Lat,
			Lng:          req.CustomerLocation.Lng,
			CashFee:      consts.Yes, // TODO 预付费这个为NO
			ContactName:  req.CustomerLocation.ContactName,
			ContactPhone: req.CustomerLocation.ContactPhone,
			Seq:          2,
		},
	}

	// TODO 保存订单

	reqInp := &skootar.CreateOrderInp{
		ReqBase:         s.ReqBase(),
		LocationList:    locationList,
		Vehicle:         "Motorcycle",
		JobType:         "3",
		JobDate:         time.Now().Format(time.DateOnly),
		StartTime:       time.Now().Format("15:04"),
		PaymentType:     "cash",
		MerchantConfirm: 1,
		CallbackUrl:     req.CallbackUrl, // TODO 构造回调地址
		Option:          "10",
	}

	resp := &skootar.CreateOrderOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(createNewJobApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("创建订单失败:%+v", reqInp)
	}

	if err := json.Unmarshal(rr.Bytes(), resp); err != nil {
		return nil, gerror.Newf("创建订单失败:%+v", err)
	}

	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("创建订单异常:%v", resp.ResponseDesc)
	}
	res = &api.CreateOrderResp{
		OrderId: resp.JobDetail.JobID,
		Status:  resp.JobDetail.JobStatus,
	}
	if err = gconv.Struct(resp, res); err != nil {
		return nil, gerror.Wrap(err, "创建订单失败")
	}
	res.ResponseInfo = &api.ResponseInfo{
		Code:    resp.ResponseCode,
		Message: resp.ResponseDesc,
	}
	return
}

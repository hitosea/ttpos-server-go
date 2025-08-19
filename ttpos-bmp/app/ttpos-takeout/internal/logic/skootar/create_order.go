package skootar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	v1 "ttpos-bmp/app/ttpos-takeout/api/callback/v1"
	api "ttpos-bmp/app/ttpos-takeout/api/takeout"
	"ttpos-bmp/app/ttpos-takeout/internal/consts"
	"ttpos-bmp/app/ttpos-takeout/internal/dao"
	"ttpos-bmp/app/ttpos-takeout/internal/model/do"
	"ttpos-bmp/app/ttpos-takeout/internal/model/dto/skootar"
	"ttpos-bmp/app/ttpos-takeout/internal/model/entity"

	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gmeta"
	"github.com/gogf/gf/v2/util/guid"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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

	reqInp := &skootar.CreateOrderInp{
		ReqBase:         s.ReqBase(),
		LocationList:    locationList,
		Vehicle:         gconv.String(consts.VehicleMotorcycle),
		JobType:         gconv.String(consts.JobTypeFood),
		JobDate:         time.Now().Format(time.DateOnly),
		StartTime:       consts.DEFAULT_START_TIME,
		PaymentType:     gconv.String(consts.PaymentTypeCash),
		MerchantConfirm: 1,                // 送餐默认需要商家确认
		CallbackUrl:     getCallbackUrl(), // 构造外送模块回调地址
		Option:          gconv.String(consts.EstimateOptionFood),
		RefNo:           req.ShopOrderUuid,
		Remark:          req.Remark,
	}

	resp := &skootar.CreateOrderOut{}
	rr := g.Client().ContentJson().PostVar(ctx, s.GetUrl(createNewJobApiPath), reqInp)
	if rr == nil {
		return nil, gerror.Newf("创建订单失败:%+v", reqInp)
	}

	if err = json.Unmarshal(rr.Bytes(), resp); err != nil {
		return nil, gerror.Newf("创建订单失败:%+v", err)
	}

	// 处理返回结果
	if resp.ResponseCode != "200" {
		return nil, gerror.Newf("创建订单异常:%v", resp.ResponseDesc)
	}

	//保存订单数据，后面移到通用的服务里面
	job := &entity.Job{
		Uuid:                 guid.S(),
		ShopLocationUuid:     guid.S(),
		ConsumerLocationUuid: guid.S(),
		TakeoutRefNo:         resp.JobDetail.JobId,
		ShopRefNo:            req.ShopOrderUuid,
		PaymentType:          reqInp.PaymentType,
		CallbackUrl:          req.CallbackUrl, // 来源订单的回调
		ProviderName:         gconv.String(consts.ProviderSkootar),
		JobDate:              reqInp.JobDate,
		StartTime:            reqInp.StartTime,
		FinishTime:           resp.JobDetail.FinishTime,
		JobStatus:            gconv.String(resp.JobDetail.JobStatus),
		Remark:               reqInp.Remark,
	}
	locationModel := dao.JobLocation.Ctx(ctx)
	if _, err = locationModel.Data(do.JobLocation{
		Uuid:         job.ShopLocationUuid,
		LocationType: 0, //餐馆
		AddressName:  req.MerchantLocation.AddressName,
		Address:      req.MerchantLocation.Address,
		Lat:          req.MerchantLocation.Lat,
		Lng:          req.MerchantLocation.Lng,
		ContactName:  req.MerchantLocation.ContactName,
		ContactPhone: req.MerchantLocation.ContactPhone,
		Seq:          1,
	}).Insert(); err != nil {
		return nil, gerror.Wrap(err, "保存商家位置失败")
	}
	if _, err = locationModel.Data(do.JobLocation{
		Uuid:         job.ConsumerLocationUuid,
		LocationType: 1, //客户
		AddressName:  req.CustomerLocation.AddressName,
		Address:      req.CustomerLocation.Address,
		Lat:          req.CustomerLocation.Lat,
		Lng:          req.CustomerLocation.Lng,
		ContactName:  req.CustomerLocation.ContactName,
		ContactPhone: req.CustomerLocation.ContactPhone,
		Seq:          2,
	}).Insert(); err != nil {
		return nil, gerror.Wrap(err, "保存商家位置失败")
	}

	// 保存订单
	if _, err = dao.Job.Ctx(ctx).Data(job).Insert(); err != nil {
		return nil, gerror.Wrap(err, "创建订单失败")
	}

	res = &api.CreateOrderResp{
		ResponseInfo: &api.ResponseInfo{
			Code:    resp.ResponseCode,
			Message: resp.ResponseDesc,
		},
		TakeoutJobUuid: job.Uuid,
		TakeoutRefNo:   job.TakeoutRefNo,
		ShopOrderUuid:  job.ShopRefNo,
		Status:         job.JobStatus,
		FinishTime:     job.FinishTime,
	}
	return
}

func getCallbackUrl() string {
	return fmt.Sprintf("%s%s", g.Cfg().MustGet(gctx.GetInitCtx(), "app.serviceUrl").String(), gmeta.Get(v1.SkootarStatusReq{}, "path"))
}

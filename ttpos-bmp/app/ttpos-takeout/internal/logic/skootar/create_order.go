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

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
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

	// 保存订单数据到新的双表结构
	orderUuid := guid.S()
	shopLocationUuid := guid.S()
	consumerLocationUuid := guid.S()

	// 使用事务确保数据一致性
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 1. 写入主表 takeout_order (通用字段)
		order := &entity.Order{
			Uuid:               orderUuid,
			ShopUuid:           "", // TODO: 从配置或上下文获取 shop_uuid
			ProviderMerchantId: "", // Skootar 没有 merchant_id
			ProviderOrderId:    resp.JobDetail.JobId,
			ShopRefNo:          req.ShopOrderUuid, // TTPOS 内部订单号
			ProviderName:       gconv.String(consts.ProviderSkootar),
			OrderStatus:        gconv.String(resp.JobDetail.JobStatus),
			OrderTime:          gtime.Now(),
			PaymentType:        reqInp.PaymentType,
			CustomerPhone:      req.CustomerLocation.ContactPhone,
			Note:               reqInp.Remark,
			CreatedAt:          gtime.Now(),
			UpdatedAt:          gtime.Now(),
		}
		if _, err := dao.Order.Ctx(ctx).TX(tx).Data(order).Insert(); err != nil {
			return gerror.Wrap(err, "创建订单主表失败")
		}

		// 2. 写入扩展表 takeout_order_skootar (Skootar 特有字段)
		// 注意：此时司机信息可能为空，等待回调更新
		orderSkootar := &entity.OrderSkootar{
			Uuid:      guid.S(),
			OrderUuid: orderUuid,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}
		if _, err := dao.OrderSkootar.Ctx(ctx).TX(tx).Data(orderSkootar).Insert(); err != nil {
			return gerror.Wrap(err, "创建订单扩展表失败")
		}

		// 3. 写入位置信息 (保持原有逻辑)
		locationModel := dao.JobLocation.Ctx(ctx).TX(tx)
		if _, err := locationModel.Data(do.JobLocation{
			Uuid:         shopLocationUuid,
			LocationType: 0, //餐馆
			AddressName:  req.MerchantLocation.AddressName,
			Address:      req.MerchantLocation.Address,
			Lat:          req.MerchantLocation.Lat,
			Lng:          req.MerchantLocation.Lng,
			ContactName:  req.MerchantLocation.ContactName,
			ContactPhone: req.MerchantLocation.ContactPhone,
			Seq:          1,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "保存商家位置失败")
		}
		if _, err := locationModel.Data(do.JobLocation{
			Uuid:         consumerLocationUuid,
			LocationType: 1, //客户
			AddressName:  req.CustomerLocation.AddressName,
			Address:      req.CustomerLocation.Address,
			Lat:          req.CustomerLocation.Lat,
			Lng:          req.CustomerLocation.Lng,
			ContactName:  req.CustomerLocation.ContactName,
			ContactPhone: req.CustomerLocation.ContactPhone,
			Seq:          2,
		}).Insert(); err != nil {
			return gerror.Wrap(err, "保存客户位置失败")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res = &api.CreateOrderResp{
		ResponseInfo: &api.ResponseInfo{
			Code:    resp.ResponseCode,
			Message: resp.ResponseDesc,
		},
		TakeoutJobUuid: orderUuid,
		TakeoutRefNo:   resp.JobDetail.JobId,
		ShopOrderUuid:  req.ShopOrderUuid,
		Status:         gconv.String(resp.JobDetail.JobStatus),
		FinishTime:     resp.JobDetail.FinishTime,
	}
	return
}

func getCallbackUrl() string {
	return fmt.Sprintf("%s%s", g.Cfg().MustGet(gctx.GetInitCtx(), "app.serviceUrl").String(), gmeta.Get(v1.SkootarStatusReq{}, "path"))
}

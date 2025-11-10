package service

import (
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// IMemberCallbackSrv 定义会员回调服务接口
type IMemberCallbackSrv interface {
	ParseMemberCallbackData(ctx context.Context, data member_req.CallbackReq) error
}

type memberCallbackSrv struct {
	dbm        *database.DBManager   // 数据库管理器
	bus        *event.SystemEventBus // 事件总线
	localeSrv  ILocaleSrv            // 多语言名称服务
	settingSrv setting.ISrv
	orderSrv   IOrderSrv
}

func NewMemberCallbackSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, orderSrv IOrderSrv) IMemberCallbackSrv {
	return NewMemberCallbackSrvImpl(dbm, localeSrv, settingSrv, orderSrv)
}

func NewMemberCallbackSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, orderSrv IOrderSrv) IMemberCallbackSrv {
	return &memberCallbackSrv{
		dbm:        dbm,
		bus:        event.NewSystemBus(),
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
		orderSrv:   orderSrv,
	}
}

// ParseMemberCallbackData 解析会员回调数据
func (s *memberCallbackSrv) ParseMemberCallbackData(ctx context.Context, data member_req.CallbackReq) error {
	companyUuid := ctx.GetCompanyUuid()

	// 检查数据是否合法
	providerName := strings.ToLower(data.ProviderName)
	if providerName != constant.ProviderNameSkootar &&
		providerName != constant.ProviderNameGrab {
		return errors.New("外送渠道名称不合法")
	}
	if data.JobStatusAfter != constant.SkootarCancel &&
		data.JobStatusAfter != constant.SkootarNew &&
		data.JobStatusAfter != constant.SkootarReviewing &&
		data.JobStatusAfter != constant.SkootarApproved &&
		data.JobStatusAfter != constant.SkootarMatchingDriver &&
		data.JobStatusAfter != constant.SkootarDriverMatched &&
		data.JobStatusAfter != constant.SkootarFetching &&
		data.JobStatusAfter != constant.SkootarDelivering &&
		data.JobStatusAfter != constant.SkootarPendingModification &&
		data.JobStatusAfter != constant.SkootarCompleted &&
		data.JobStatusAfter != constant.SkootarInvoiceGenerated &&
		data.JobStatusAfter != constant.SkootarServiceFeePaid {
		return errors.New("外送订单状态不合法")
	}

	// skootar的回调时
	if providerName == constant.ProviderNameSkootar {
		// 当状态“正在前往取货 OTW to pickup”时，触发“骑手接单”事件
		if data.JobStatusAfter == constant.SkootarFetching {
			memberSaleOrderUuid, err := utils.StringToUint64(data.ShopRefNo)
			if err != nil {
				ctx.Log().Error("骑手接单回调时，订单号不合法", zap.Error(err))
				return errors.WithMessage(err, "骑手接单回调时，订单号不合法")
			}
			utils.Go(func() {
				s.bus.PublishRiderAcceptMemberSaleOrderEvent(event.RiderAcceptMemberSaleOrderPayload{
					BasePayload: event.BasePayload{
						Ctx:         ctx,
						CompanyUuid: companyUuid,
					},
					MemberSaleOrderUuid: memberSaleOrderUuid,
				})
			})
			return nil
		}
		// 当状态“正在前往送货 OTW to delivery”时，触发“骑手完成取餐”事件
		if data.JobStatusAfter == constant.SkootarDelivering {
			memberSaleOrderUuid, err := utils.StringToUint64(data.ShopRefNo)
			if err != nil {
				ctx.Log().Error("骑手完成取餐回调时，订单号不合法", zap.Error(err))
				return errors.WithMessage(err, "骑手完成取餐回调时，订单号不合法")
			}
			utils.Go(func() {
				s.bus.PublishRiderDeliveryMemberSaleOrderEvent(event.RiderDeliveryMemberSaleOrderPayload{
					BasePayload: event.BasePayload{
						Ctx:         ctx,
						CompanyUuid: companyUuid,
					},
					MemberSaleOrderUuid: memberSaleOrderUuid,
				})
			})
			return nil
		}
		// 当状态“已完成 Completed”时，触发“骑手送达”事件
		if data.JobStatusAfter == constant.SkootarCompleted {
			memberSaleOrderUuid, err := utils.StringToUint64(data.ShopRefNo)
			if err != nil {
				ctx.Log().Error("骑手送达回调时，订单号不合法", zap.Error(err))
				return errors.WithMessage(err, "骑手送达回调时，订单号不合法")
			}
			if err := s.orderSrv.CompleteMemberSaleOrder(ctx, memberSaleOrderUuid); err != nil {
				ctx.Log().Error("完成外送订单失败", zap.Error(err))
				return errors.WithMessage(err, "完成外送订单失败")
			}
			return nil
		}
	}

	return errors.WithMessage(errors.New("未找到对应的回调事件"), "未找到对应的回调事件")
}

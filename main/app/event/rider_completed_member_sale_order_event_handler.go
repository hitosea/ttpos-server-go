package event

import (
	"context"
	"fmt"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

var once_rider_completed_member_sale_order_event_handler sync.Once

// init 自动注册事件处理器
func init() {
	// 只初始化一次
	riderCompletedMemberSaleOrderEventHandler()
}

// riderCompletedMemberSaleOrderEventHandler "骑手配送完成"事件处理器
func riderCompletedMemberSaleOrderEventHandler() {
	once_rider_completed_member_sale_order_event_handler.Do(func() {
		// 骑手配送完成后，更新订单状态
		event.NewSystemBus().SubscribeRiderCompletedMemberSaleOrderEvent(func(payload event.RiderCompletedMemberSaleOrderPayload) {
			db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)

			// 更新会员端销售订单状态
			updateMemberSaleOrder := model.MemberSaleOrder{
				BaseModel: model.BaseModel{
					Uuid: payload.MemberSaleOrderUuid,
				},
				Sort:       constant.MemberSaleOrderSortDefault,
				Status:     constant.MemberSaleOrderStatusCompleted, // 骑手配送完成
				FinishTime: time.Now().Unix(),
			}

			// 获取骑手信息
			driverInfoResp, err := takeout.NewTakeoutSrv().GetDriverInfo(context.Background(), &req.GetDriverInfoReq{
				ShopOrderUuid: fmt.Sprintf("%d", payload.MemberSaleOrderUuid),
			})
			if err == nil && driverInfoResp != nil {
				// 设置payload
				payload.RiderName = driverInfoResp.Name
				payload.RiderPhone = driverInfoResp.Phone
				// 更新会员端销售订单状态
				updateMemberSaleOrder.RiderCompleted(
					payload.RiderName,
					payload.RiderPhone,
					fmt.Sprintf("%f,%f", driverInfoResp.Lat, driverInfoResp.Lng),
				)
			}

			// 更新会员端销售订单状态
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderCompleted(updateMemberSaleOrder); err != nil {
				logger.Logger.Error("更新会员端销售订单-骑手配送完成失败", zap.Error(err))
				return
			}

			// 设置sort排序。 ！！！ 注意是修改sort字段为0，gorm默认不修改值为0的字段
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderSort(updateMemberSaleOrder.Uuid, constant.MemberSaleOrderSortRiderDelivering); err != nil {
				payload.Ctx.Log().Error("更新会员端销售订单-骑手配送中排序失败", zap.Error(err))
			}

			// 获取会员端销售订单记录
			memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecordOnly(updateMemberSaleOrder.Uuid)
			if err != nil {
				logger.Logger.Error("获取会员端销售订单操作记录失败", zap.Error(err))
				return
			}

			// 设置payload
			payload.SaleBillUuid = memberSaleOrder.SaleBillUuid
			payload.SaleOrderUuid = memberSaleOrder.SaleOrderUuid
			if payload.RiderName == "" {
				payload.RiderName = "-" // 骑手默认名
			}

			// 创建“骑手接单”操作记录
			go func() {
				record := model.SaleOrderOperationRecord{
					Source:        constant.SourceRider, // 骑手端
					Action:        constant.OrderFinishMemberSaleOrder,
					Remark:        "配送完成，订单完成",
					SaleBillUuid:  payload.SaleBillUuid,
					SaleOrderUuid: payload.SaleOrderUuid,
					OperatorUuid:  payload.GetOperatorUuid(),
				}
				record.Data = payload.ToJsonString()
				record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
				uuid, err := repository.NewOrderOperationRecordRepo(db).CreateSaleOrderOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeRiderCompletedMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:配送完成，订单完成 %+v", payload), zap.Uint64("record", uuid))
			}()

			// 外送订单完结时, 发布"统计"事件
			go func() {
				event.NewSystemBus().PublishStatisticsSaleEvent(event.StatisticsSalePayload{
					BasePayload: event.BasePayload{ // 统计
						Ctx: payload.Ctx,
					},
					SaleBillUuid: memberSaleOrder.SaleBillUuid,
				})
			}()

			// 发送奖励
			go func() {
				// 当前销售账单数据
				saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(0, repository.WithMemberSaleOrderUuid(memberSaleOrder.Uuid))
				if errSaleBill != nil {
					logger.Logger.Error("发送奖励-获取当前销售账单数据失败", zap.Error(errSaleBill))
					return
				}
				// 处理邀请有礼活动-统计获奖
				HandleActivityConsumption(event.CheckoutSaleOrderPayload{
					BasePayload: event.BasePayload{
						Ctx:                 payload.Ctx,
						CompanyUuid:         payload.CompanyUuid,
						Source:              payload.Source,
						SaleBillUuid:        memberSaleOrder.SaleBillUuid,
						SaleOrderUuid:       memberSaleOrder.SaleOrderUuid,
						MemberSaleOrderUuid: memberSaleOrder.Uuid,
						MemberUuid:          memberSaleOrder.MemberUuid,
					},
					SaleBill: saleBill,
				})
			}()

			// 更新会员消费金额和消费次数
			go func() {
				if err := repository.NewMemberRepo(db).IncConsumptionAmount(memberSaleOrder.MemberUuid, memberSaleOrder.GetActualConsumptionAmount()); err != nil {
					logger.Logger.Error("更新会员消费金额和消费次数-更新会员消费金额失败", zap.Error(err))
					return
				}
				if err := repository.NewMemberRepo(db).IncConsumptionCount(memberSaleOrder.MemberUuid); err != nil {
					logger.Logger.Error("更新会员消费金额和消费次数-更新会员消费次数失败", zap.Error(err))
					return
				}
			}()
		})
	})

}

package event

import (
	"context"
	"fmt"
	"sync"
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

			memberSaleOrder := model.MemberSaleOrder{
				BaseModel: model.BaseModel{
					Uuid: payload.MemberSaleOrderUuid,
				},
			}

			takeoutSrv := takeout.NewTakeoutSrv()
			driverInfoResp, err := takeoutSrv.GetDriverInfo(context.Background(), &req.GetDriverInfoReq{
				ShopOrderUuid: fmt.Sprintf("%d", payload.MemberSaleOrderUuid),
			})
			if err != nil {
				payload.Ctx.Log().Error("获取骑手信息失败", zap.Error(err))
			}

			var riderName string     // 骑手名称
			var riderPhone string    // 骑手手机号
			var riderLocation string // 骑手经纬度
			if driverInfoResp != nil {
				riderName = driverInfoResp.Name
				riderPhone = driverInfoResp.Phone
				riderLocation = fmt.Sprintf("%f,%f", driverInfoResp.Lat, driverInfoResp.Lng)
			}
			memberSaleOrder.RiderCompleted(riderName, riderPhone, riderLocation)
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderCompleted(memberSaleOrder); err != nil {
				payload.Ctx.Log().Error("更新会员端销售订单-骑手配送完成失败", zap.Error(err))
				return
			}

			// 创建“骑手接单”操作记录
			go func() {
				payload.RiderName = riderName
				payload.RiderPhone = riderPhone
				if payload.RiderName == "" {
					payload.RiderName = "-" // 骑手默认名
				}
				memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecordOnly(memberSaleOrder.Uuid)
				if err != nil {
					payload.Ctx.Log().Error("获取会员端销售订单操作记录失败", zap.Error(err))
					return
				}
				payload.SaleBillUuid = memberSaleOrder.SaleBillUuid
				payload.SaleOrderUuid = memberSaleOrder.SaleOrderUuid
				db := database.GetDBManager(config.DatabaseConf{}).GetDB(payload.CompanyUuid)
				orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
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
				uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeRiderCompletedMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:配送完成，订单完成 %+v", payload), zap.Uint64("record", uuid))
			}()

			// 外送订单完结时, 发布"统计"事件
			order, err := repository.NewMemberSaleOrderRepoImpl(db).GetMemberSaleOrder(
				repository.NewCommonRepoImpl().WhereByUuid(memberSaleOrder.Uuid),
			)
			if err != nil {
				payload.Ctx.Log().Error("获取会员端销售订单记录失败", zap.Error(err))
				return
			}
			go func() {
				event.NewSystemBus().PublishStatisticsSaleEvent(event.StatisticsSalePayload{
					BasePayload: event.BasePayload{ // 统计
						Ctx: payload.Ctx,
					},
					SaleBillUuid: order.SaleBillUuid,
				})
			}()
		})
	})
}

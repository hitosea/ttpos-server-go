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

var once_rider_delivery_member_sale_order_event_handler sync.Once

// RiderDeliveryMemberSaleOrderEventHandler "骑手配送中"事件处理器
func RiderDeliveryMemberSaleOrderEventHandler() {
	once_rider_delivery_member_sale_order_event_handler.Do(func() {
		// 骑手配送中后，更新订单状态
		event.NewSystemBus().SubscribeRiderDeliveryMemberSaleOrderEvent(func(payload event.RiderDeliveryMemberSaleOrderPayload) {
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
			memberSaleOrder.RiderDelivery(riderName, riderPhone, riderLocation)
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderRiderDelivery(memberSaleOrder); err != nil {
				payload.Ctx.Log().Error("更新会员端销售订单-骑手配送中失败", zap.Error(err))
				return
			}

			// 创建“骑手接单”操作记录
			utils.Go(func() {
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
				orderRecordRepo := repository.NewOrderOperationRecordRepo(db)
				record := model.SaleOrderOperationRecord{
					Source:        constant.SourceRider, // 骑手端
					Action:        constant.OrderDeliveryMemberSaleOrder,
					Remark:        "骑手取货完成",
					SaleBillUuid:  payload.SaleBillUuid,
					SaleOrderUuid: payload.SaleOrderUuid,
					OperatorUuid:  payload.GetOperatorUuid(),
				}
				record.Data = payload.ToJsonString()
				record.SetDutyNo(payload.Ctx.GetStaff().DutyNo)
				uuid, err := orderRecordRepo.CreateSaleOrderOperationRecord(record)
				if err != nil {
					logger.Logger.Error("SubscribeRiderDeliveryMemberSaleOrderEvent process, CreateSaleOrderOperationRecord failed", zap.Any("record", utils.ToJson(record)), zap.Error(err))
					return
				}
				logger.Logger.Info(fmt.Sprintf("操作记录:骑手取货完成 %+v", payload), zap.Uint64("record", uuid))
			})

			// 设置sort排序
			if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderSort(memberSaleOrder.Uuid, constant.MemberSaleOrderSortRiderDelivering); err != nil {
				payload.Ctx.Log().Error("更新会员端销售订单-骑手配送中排序失败", zap.Error(err))
			}
		})
	})
}

package consumer

import (
	"encoding/json"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// MemberOrderCancelParams 会员订单取消队列参数
type MemberOrderCancelParams struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员订单UUID
	CompanyUuid         uint64 `json:"company_uuid"`           // 公司UUID
	Reason              string `json:"reason"`                 // 取消原因
}

// ProcessMemberOrderCancel 处理会员订单自动取消
func ProcessMemberOrderCancel(paramsJson string) bool {
	var params MemberOrderCancelParams
	if err := json.Unmarshal([]byte(paramsJson), &params); err != nil {
		logger.Logger.Error("ProcessMemberOrderCancel 处理会员订单自动取消失败", zap.String("paramsJson", paramsJson), zap.Error(err))
		return true
	}

	logger.Logger.Info("开始处理会员订单自动取消",
		zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid),
		zap.Uint64("companyUuid", params.CompanyUuid),
		zap.String("reason", params.Reason))

	// 获取DB
	dbm := database.GetDBManager(config.DatabaseConf{})
	db := dbm.GetDB(params.CompanyUuid)

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(params.MemberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 2. 检查订单状态是否可以取消 - 只有待支付状态的订单才能自动取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusPendingPayment {
		logger.Logger.Info("订单状态不是待支付，跳过自动取消",
			zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid),
			zap.Uint("currentStatus", memberSaleOrder.Status))
		return true
	}

	// 3. 执行取消操作
	memberSaleOrder.SetCancel(params.Reason)
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		logger.Logger.Error("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid), zap.Error(err))
		return false
	}

	// 4. 取消订单（如果有关联的销售账单）
	if memberSaleOrder.SaleBillUuid > 0 {
		ctx := context.NewContext(
			context.WithCompanyUuid(params.CompanyUuid),
			context.WithLogger(logger.Logger),
		)
		ctx.SetDB(db)

		if err := repository.NewOrderRepo(db).CancelOrder(ctx, memberSaleOrder.SaleBillUuid, 0, params.Reason); err != nil {
			logger.Logger.Error("取消销售账单失败",
				zap.Uint64("memberSaleOrderUuid", params.MemberSaleOrderUuid),
				zap.Uint64("saleBillUuid", memberSaleOrder.SaleBillUuid),
				zap.Error(err))
			// 这里不返回false，因为会员订单状态已经更新成功了
		}
	}

	return true
}

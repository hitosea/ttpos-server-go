package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	errors2 "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/cryptor"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                                                                                // 获取等级列表
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                                                            // 模糊搜索
	AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error                                                               // 添加会员
	GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember                                                      // 获取充值会员信息
	GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder                                                                    // 获取进行中的会员充值订单
	CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error)                                 // 创建充值订单
	AddPaymentMethod(ctx context.Context, addPaymentMethod req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error)          // 充值订单添加支付方式
	CancelPaymentMethod(ctx context.Context, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error) // 充值订单撤销支付方式
	ConfirmRechargeOrder(ctx context.Context, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error)    // 确认充值订单
}

// memberSrv 会员服务结构体
type memberSrv struct {
	dbm              *database.DBManager // 数据库管理器
	paymentMethodSrv IPaymentMethodSrv
	settingSrv       setting.ISrv
	cashBoxSrv       ICashBoxSrv
}

// NewMemberSrv 创建新的会员服务
func NewMemberSrv(dbm *database.DBManager, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv) IMemberSrv {
	return NewMemberSrvImpl(dbm, paymentMethodSrv, settingSrv, cashBoxSrv)
}

// NewMemberSrvImpl 创建新的会员服务实现
func NewMemberSrvImpl(dbm *database.DBManager, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv, cashBoxSrv ICashBoxSrv) IMemberSrv {
	return &memberSrv{
		dbm:              dbm,
		paymentMethodSrv: paymentMethodSrv,
		settingSrv:       settingSrv,
		cashBoxSrv:       cashBoxSrv,
	}
}

// GetLevels 获取等级列表
func (s *memberSrv) GetLevels(companyUuid uint64) resp.MemberLevelList {
	memberLevels := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).GetMemberLevels()
	respMemberLevels := make([]resp.MemberLevel, 0)
	for _, memberLevel := range memberLevels {
		var respMemberLevel resp.MemberLevel
		copier.Copy(&respMemberLevel, memberLevel)
		respMemberLevels = append(respMemberLevels, respMemberLevel)
	}
	return resp.MemberLevelList{
		List: respMemberLevels,
	}
}

// SearchMember 模糊搜索会员
func (s *memberSrv) SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList {
	searchMembers := make([]resp.SearchMember, 0)
	if keyword == "" {
		return resp.SearchMemberList{List: searchMembers}
	}
	members := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).SearchMember(keyword)
	for _, member := range members {
		var searchMember resp.SearchMember
		copier.Copy(&searchMember, member)
		searchMembers = append(searchMembers, searchMember)
	}
	return resp.SearchMemberList{
		List: searchMembers,
	}
}

// AddMember 添加会员
func (s *memberSrv) AddMember(ctx context.Context, addMemberReq req.AddMemberReq) error {
	companyUuid := ctx.GetCompanyUuid()
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))

	// 判断等级是否存在
	if !memberRepo.CheckLevelExists(addMemberReq.LevelUuid) {
		return errors.New("会员等级不存在")
	}

	// 判断是否存在
	if memberRepo.CheckMemberExists(addMemberReq.Phone) {
		return errors.New("会员已存在")
	}
	if addMemberReq.Password != "" {
		addMemberReq.Password = cryptor.Md5String(addMemberReq.Password)
	}
	if err := memberRepo.CreateMember(model.Member{
		MemberNo:        utils.RandomNumber(5), // 5位数字
		Nickname:        addMemberReq.Nickname,
		Phone:           addMemberReq.Phone,
		Password:        "",
		MemberLevelUuid: addMemberReq.LevelUuid,
	}); err != nil {
		ctx.Log().Error("添加会员失败", zap.Error(err))
		return errors.New("添加会员失败")
	}
	return nil
}

// GetRechargeMember 获取充值会员信息
func (s *memberSrv) GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember {
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetByUuid(memberUuid, memberRepo.WithMemberCard(), memberRepo.WithMemberCardType(), memberRepo.WithMemberLevel())

	var cardName string
	if member.MemberCard != nil && member.MemberCard.MemberCardType != nil {
		cardName = member.MemberCard.MemberCardType.Name
	}
	var level string
	if member.MemberLevel != nil {
		level = member.MemberLevel.Name
	}
	return resp.RechargeMember{
		Uuid:      member.Uuid,
		Nickname:  member.Nickname,
		CardName:  cardName,
		LevelName: level,
		Balance:   member.Balance + member.GiftBalance,
		Points:    member.Point,
	}
}

// GetPendingRechargeOrder 进行中的会员充值订单
func (s *memberSrv) GetPendingRechargeOrder(companyUuid uint64) resp.RechargeOrder {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	// 进行中的充值订单
	rechargeOrder := rechargeOrderRepo.GetPendingRechargeOrder(rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	if rechargeOrder.Uuid == 0 {
		return resp.RechargeOrder{PaymentOrders: make([]resp.PaymentOrder, 0)}
	}

	paymentOrderCount := len(rechargeOrder.PaymentOrders)
	respPaymentOrders := make([]resp.PaymentOrder, 0, paymentOrderCount)
	if paymentOrderCount > 0 {
		var respPaymentOrder resp.PaymentOrder
		for _, paymentOrder := range rechargeOrder.PaymentOrders {
			copier.Copy(&respPaymentOrder, paymentOrder)
			respPaymentOrder.PaymentMethodCode = paymentOrder.PaymentMethod.Code
			respPaymentOrders = append(respPaymentOrders, respPaymentOrder)
		}
	}

	var respRechargeOrder resp.RechargeOrder
	copier.Copy(&respRechargeOrder, rechargeOrder)

	respRechargeOrder.PaymentOrders = respPaymentOrders
	return respRechargeOrder
}

// 计算支付订单初始金额
func (s *memberSrv) sumPaymentAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentAmount
	}
	return sum
}

// 计算应收金额
func (s *memberSrv) getRechargeOrderAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentAmount + paymentOrder.PaymentCommissionFee
	}
	return sum
}

// 计算找零
func (s *memberSrv) getChargeDue(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.Amount - paymentOrder.PaymentCommissionFee - paymentOrder.PaymentAmount
	}
	return sum
}

// 计算支付订单实收金额
func (s *memberSrv) getActualAmount(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.Amount
	}
	return sum
}

// 计算支付订单手续费
func (s *memberSrv) getPayFee(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		sum = sum + paymentOrder.PaymentCommissionFee
	}
	return sum
}

// CreateRechargeOrder 创建充值订单
func (s *memberSrv) CreateRechargeOrder(ctx context.Context, rechargeReq req.RechargeReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	// 判断会员是否存在
	member := repository.NewMemberRepo(s.dbm.GetDB(companyUuid)).GetByUuid(rechargeReq.MemberUuid)
	if member.ID == 0 {
		return orderResp, errors.New("会员不存在")
	}

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))

	if rechargeReq.RechargeOrderUuid != 0 {
		// 如果已经存在已支付的payment_order，则充值金额不能小于现有的 "充值金额不能小于已充值金额"
		rechargeOrder := rechargeOrderRepo.GetPendingRechargeOrder(rechargeOrderRepo.WithPaymentOrders())
		if rechargeOrder.Uuid != 0 {
			oldRechargeAmount := rechargeOrder.Amount

			sumPaymentAmount := s.sumPaymentAmount(rechargeOrder.PaymentOrders)
			if rechargeReq.RechargeAmount < sumPaymentAmount {
				return orderResp, errors.New("充值金额不能小于已充值金额")
			}
			if err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
				// 更新充值订单信息
				err := repository.NewMemberRechargeOrderRepo(tx).Update(rechargeOrder.Uuid, map[string]any{
					"recharge_amount": rechargeReq.RechargeAmount,
					"gift_amount":     rechargeReq.GiftAmount,
					"gift_point":      rechargeReq.GiftPoint,
					"member_uuid":     rechargeReq.MemberUuid,
					"staff_uuid":      ctx.GetStaffUuid(),
				})
				// 会员充值操作日志
				operationData, _ := json.Marshal(map[string]any{
					"recharge_money":     rechargeReq.RechargeAmount,
					"old_recharge_money": oldRechargeAmount,
				})
				err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
					OperatorName:      ctx.GetStaff().RealName,
					OperatorEmail:     ctx.GetStaff().Username,
					Client:            ctx.GetSource(),
					Message:           "变更充值金额",
					Action:            constant.RechargeOrderActionChangeAmount,
					Data:              string(operationData),
					RechargeOrderUuid: rechargeOrder.Uuid,
				})
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				ctx.Log().Error("修改充值订单失败", zap.Error(err))
				return orderResp, errors.New("修改充值订单失败")
			}

			// 返回充值订单信息
			return s.GetPendingRechargeOrder(companyUuid), nil
		}
	}

	// 如果已存在进行中的充值订单，且未传递进行中的充值订单Uuid，则直接返回
	pendingRechargeOrder := rechargeOrderRepo.GetPendingRechargeOrder()
	if pendingRechargeOrder.Uuid != 0 {
		return s.GetPendingRechargeOrder(companyUuid), nil
	}

	err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
		// 创建充值订单
		order, err := repository.NewMemberRechargeOrderRepo(tx).Create(model.MemberRechargeOrder{
			RechargeAmount: rechargeReq.RechargeAmount,
			GiftAmount:     rechargeReq.GiftAmount,
			GiftPoint:      rechargeReq.GiftPoint,
			MemberUuid:     rechargeReq.MemberUuid,
			StaffUuid:      ctx.GetStaffUuid(),
		})
		if err != nil {
			return err
		}
		// 会员充值操作日志
		err = repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "生成订单",
			Action:            constant.RechargeOrderActionGenerateOrder,
			RechargeOrderUuid: order.Uuid,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return orderResp, errors.New("创建充值订单失败")
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// AddPaymentMethod 充值订单添加支付方式
func (s *memberSrv) AddPaymentMethod(ctx context.Context, addReq req.RechargeOrderAddPaymentMethodReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetByUuid(addReq.RechargeOrderUuid, rechargeOrderRepo.WithPaymentOrders())
	if rechargeOrder.Uuid == 0 || rechargeOrder.Status != constant.RechargeOrderStatusPending {
		return orderResp, errors.New("充值订单不存在")
	}

	// todo 在线支付订单暂不处理

	// 根据Uuid 获取支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(companyUuid))
	paymentMethod := paymentMethodRepo.GetPaymentMethod(paymentMethodRepo.WhereUuid(addReq.PaymentMethodUuid))
	if paymentMethod.Uuid == 0 {
		return orderResp, errors.New("支付方式不存在")
	}
	// 支付方式是否可用
	if paymentMethod.IsShowMemberRecharge == 0 || !s.paymentMethodSrv.IsEnabled(ctx, paymentMethod, addReq.CompanySetting) {
		return orderResp, errors.New("支付方式未开启")
	}
	if paymentMethod.Code == constant.PaymentMethodCodeBalance {
		return orderResp, errors.New("不能使用余额支付充值")
	}

	// 计算支付手续费
	paymentCommissionFee := s.paymentMethodSrv.CalculatePaymentCommissionFee(paymentMethod, addReq.PaymentAmount)

	sumPaymentAmount := s.sumPaymentAmount(rechargeOrder.PaymentOrders)
	// 支付订单金额大于充值金额
	if sumPaymentAmount >= rechargeOrder.RechargeAmount {
		return orderResp, errors.New("当前已足额")
	}
	sumPaymentAmountAddCash := sumPaymentAmount + addReq.PaymentAmount
	if paymentMethod.Code != constant.PaymentMethodCodeCash && sumPaymentAmountAddCash > rechargeOrder.RechargeAmount {
		return orderResp, errors.New("非现金支付不能大于应收")
	}

	// 支付订单总金额 = 支付金额 + 支付手续费
	amount := addReq.PaymentAmount + paymentCommissionFee

	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	// 获取已存在的该支付方式的支付订单
	paymentOrder, _ := paymentOrderRepo.GetPaymentOrder(
		paymentOrderRepo.WhereRelatedUuid(rechargeOrder.Uuid), paymentOrderRepo.WherePaymentTypeUuid(paymentMethod.Uuid))

	var rechargeAmountLeft float64
	var cashPaidPaymentAmount float64
	if paymentMethod.Code == constant.PaymentMethodCodeCash {
		// 如果现金支付订单存在
		if paymentOrder.Uuid != 0 {
			cashPaidPaymentAmount = paymentOrder.PaymentAmount
		}
		rechargeAmountLeft = rechargeOrder.RechargeAmount - (sumPaymentAmount - cashPaidPaymentAmount)
		if rechargeAmountLeft > 0 && addReq.PaymentAmount > rechargeAmountLeft {
			addReq.PaymentAmount = rechargeAmountLeft
		}
	}

	if paymentOrder.Uuid != 0 { // 存在则更新
		err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
			"payment_amount":         addReq.PaymentAmount,
			"amount":                 amount,
			"payment_commission_fee": paymentCommissionFee,
		})
		if err != nil {
			logger.Logger.Error("添加支付订单-更新支付方式", zap.Error(err))
			return orderResp, errors.New("添加支付方式失败")
		}
	} else { // 不存在则创建
		currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
		if err != nil {
			logger.Logger.Error("添加支付订单-获取货币设置失败", zap.Error(err))
			return orderResp, errors.New("添加支付方式失败")
		}
		_, err = paymentOrderRepo.Create(model.PaymentOrder{
			PaymentMethodName:    paymentMethod.PaymentName,
			PaymentMethodUuid:    paymentMethod.Uuid,
			PaymentFeePercent:    paymentMethod.FeePercent,
			RelatedUuid:          rechargeOrder.Uuid,
			CurrencyUnit:         currencySetting.Unit, // 留档使用
			PaymentAmount:        addReq.PaymentAmount,
			PaymentCommissionFee: paymentCommissionFee,
			Amount:               amount,
			Status:               constant.PaymentOrderStatusPaid, // ToDo 手动添加在线支付标为0，处理lianlianpay
		})
		if err != nil {
			logger.Logger.Error("添加支付订单-创建支付订单", zap.Error(err))
			return orderResp, errors.New("添加支付方式失败")
		}
	}

	return s.GetPendingRechargeOrder(companyUuid), nil
}

// CancelPaymentMethod 充值订单撤销支付方式
func (s *memberSrv) CancelPaymentMethod(ctx context.Context, cancelReq req.RechargeOrderCancelPaymentMethodReq) (resp.RechargeOrder, error) {
	companyUuid := ctx.GetCompanyUuid()
	var orderResp resp.RechargeOrder

	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	// 判断支付订单
	paymentOrder, err := paymentOrderRepo.GetPaymentOrder(paymentOrderRepo.WhereUuid(cancelReq.PaymentOrderUuid), paymentOrderRepo.WithPaymentMethod(), paymentOrderRepo.WithMemberRechargeOrder())
	if err != nil || paymentOrder.Uuid == 0 {
		return orderResp, errors.New("支付订单不存在")
	}
	if paymentOrder.MemberRechargeOrder == nil || paymentOrder.RelatedUuid != cancelReq.RechargeOrderUuid {
		return orderResp, errors.New("充值订单不存在")
	}
	// lianlianpay不支持撤销
	if paymentOrder.PaymentMethod == nil || paymentOrder.PaymentMethod.Source == constant.PaymentMethodSourceLianLianPay {
		return orderResp, errors.New("支付方式不可撤销")
	}
	// 标记支付订单删除
	if err := paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
		"delete_time": time.Now().Unix(),
	}); err != nil {
		logger.Logger.Error("标记支付订单失败", zap.Error(err))
		return orderResp, errors2.ErrInternal
	}

	// 更新现金支付订单
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetByUuid(paymentOrder.RelatedUuid, rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	var cashPaymentOrder model.PaymentOrder
	var sumPaymentAmount float64
	for _, order := range rechargeOrder.PaymentOrders {
		if order.PaymentMethod.Code == constant.PaymentMethodCodeCash {
			cashPaymentOrder = order
		}
		sumPaymentAmount = sumPaymentAmount + order.PaymentAmount
	}
	if cashPaymentOrder.Uuid != 0 {
		paymentOrderRepo.Update(cashPaymentOrder.Uuid, map[string]any{
			"payment_amount": rechargeOrder.RechargeAmount - sumPaymentAmount,
		})
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// 计算非现金支付订单金额(payment_amount)
func (s *memberSrv) sumPaymentAmountExcludeCash(paymentOrders []model.PaymentOrder) float64 {
	var sum float64
	for _, paymentOrder := range paymentOrders {
		if paymentOrder.PaymentMethod.Code != constant.PaymentMethodCodeCash {
			sum = sum + paymentOrder.PaymentAmount
		}
	}
	return sum
}

// ConfirmRechargeOrder 确认充值订单
func (s *memberSrv) ConfirmRechargeOrder(ctx context.Context, confirmReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error) {
	var confirmResp resp.ConfirmRechargeOrder

	companyUuid := ctx.GetCompanyUuid()
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetByUuid(confirmReq.MemberUuid)
	if member.Uuid == 0 {
		return confirmResp, errors.New("会员不存在")
	}

	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetByUuid(confirmReq.RechargeOrderUuid, rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())
	if rechargeOrder.Uuid == 0 || rechargeOrder.Status != constant.RechargeOrderStatusPending {
		return confirmResp, errors.New("充值订单不存在")
	}

	// 非现金超额
	sumPaymentAmountExcludeCash := s.sumPaymentAmountExcludeCash(rechargeOrder.PaymentOrders)
	if sumPaymentAmountExcludeCash > rechargeOrder.RechargeAmount {
		return confirmResp, errors.New("收款金额大于充值金额，请先修改收款金额")
	}
	// 所有支付订单金额总数小于充值订单充值金额
	sumPaymentAmount := s.sumPaymentAmount(rechargeOrder.PaymentOrders)
	if sumPaymentAmount < rechargeOrder.RechargeAmount {
		return confirmResp, errors.New("未足额支付")
	}

	// 更新充值订单
	updates := map[string]any{
		"amount":       s.getRechargeOrderAmount(rechargeOrder.PaymentOrders), // 应收金额
		"status":       constant.RechargeOrderStatusPaid,                      // 状态,0-pending待支付 1-paid已支付 2-canceled已取消 3-exp已过期
		"payment_time": time.Now().Unix(),
		"charge_due":   s.getChargeDue(rechargeOrder.PaymentOrders), // 找零
	}

	var updateMemberPoints bool

	err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {

		err := repository.NewMemberRechargeOrderRepo(tx).Update(rechargeOrder.Uuid, updates)
		if err != nil {
			return errors.New("更新充值订单失败")
		}

		// 处理积分
		if rechargeOrder.GiftPoint > 0 {
			if _, err := repository.NewMemberPointLogRepo(tx).Create(model.MemberPointLog{
				MemberUuid: rechargeOrder.MemberUuid,
				Scene:      constant.MemberPointLogSceneRechargeGive,
				Value:      rechargeOrder.GiftPoint,
				Describe:   fmt.Sprintf("收银机管理员充值赠送操作 [%s]", ctx.GetStaff().RealName),
			}); err != nil {
				return errors.New("处理会员积分失败")
			}
			//
			if err := repository.NewMemberRepo(tx).Update(rechargeOrder.MemberUuid, map[string]any{
				"point": gorm.Expr("point + ?", rechargeOrder.GiftPoint),
			}); err != nil {
				return errors.New("处理会员积分失败")
			}
			updateMemberPoints = true
		}

		// 处理会员余额，充值金额+赠送金额 > 0
		money := rechargeOrder.RechargeAmount + rechargeOrder.GiftAmount
		if money > 0 {
			if err := repository.NewMemberRepo(tx).Update(confirmReq.MemberUuid, map[string]any{
				"balance":                     gorm.Expr("balance + ?", rechargeOrder.RechargeAmount),
				"gift_balance":                gorm.Expr("gift_balance + ?", rechargeOrder.GiftAmount),
				"accumulated_recharge_amount": gorm.Expr("accumulated_recharge_amount + ?", rechargeOrder.RechargeAmount),
			}); err != nil {
				return errors.New("更新会员余额失败")
			}
			if _, err := repository.NewMemberBalanceLogRepo(tx).Create(model.MemberBalanceLog{
				MemberUuid: rechargeOrder.MemberUuid,
				Scene:      constant.MemberBalanceLogRecharge,
				Money:      money,
				GiftMoney:  rechargeOrder.GiftAmount,
				Describe:   fmt.Sprintf("收银机管理员操作 [%s]", ctx.GetStaff().RealName),
			}); err != nil {
				return errors.New("处理会员余额失败")
			}
		}

		if sumPaymentAmount > sumPaymentAmountExcludeCash {
			if err := s.cashBoxSrv.UpdateBalance(ctx, tx, constant.CashBoxLogTypeIn, sumPaymentAmount-sumPaymentAmountExcludeCash, rechargeOrder.Uuid); err != nil {
				return err
			}
		}

		rechargeOperation := map[string]any{
			"order_price":    rechargeOrder.RechargeAmount,             //  订单金额
			"recharge_money": rechargeOrder.RechargeAmount,             //  充值金额
			"pay_price":      rechargeOrder.Amount,                     //  订单应付
			"gift_money":     rechargeOrder.GiftAmount,                 //  赠送金额
			"gift_point":     rechargeOrder.GiftPoint,                  //  赠送积分
			"pay_fee":        s.getPayFee(rechargeOrder.PaymentOrders), //  支付手续费
			"change_due":     rechargeOrder.ChargeDue,                  //  找零
			"pay_type":       rechargeOrder.PaymentOrders,              //  支付方式
		}

		operationData, _ := json.Marshal(rechargeOperation)

		if err := repository.NewMemberRechargeOperationRepo(tx).AddLog(model.MemberRechargeOrderOperationLog{
			OperatorName:      ctx.GetStaff().RealName,
			OperatorEmail:     ctx.GetStaff().Username,
			Client:            ctx.GetSource(),
			Message:           "充值",
			Action:            constant.RechargeOrderActionRecharge,
			Data:              string(operationData),
			RechargeOrderUuid: rechargeOrder.Uuid,
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return confirmResp, err
	}

	if updateMemberPoints {
		go s.handleMemberUpgrade(companyUuid, member.Uuid)
	}
	// ToDO 打印充值单

	return s.confirmRechargeOrderResp(companyUuid, rechargeOrder.Uuid), nil
}

// 处理会员升级
func (s *memberSrv) handleMemberUpgrade(companyUuid uint64, memberUuid uint64) {
	db := s.dbm.GetDB(companyUuid)
	memberRepo := repository.NewMemberRepo(db)
	member := memberRepo.GetByUuid(memberUuid, memberRepo.WithMemberLevel())
	if member.Uuid == 0 || member.MemberLevel == nil {
		return
	}
	memberLevels := repository.NewMemberRepo(db).GetMemberLevels()
	if len(memberLevels) == 0 {
		return
	}
	var upgradeLevel model.MemberLevel
	for _, level := range memberLevels {
		if level.IsDefault == 1 {
			continue
		}
		if s.checkCanUpgrade(member, level) {
			upgradeLevel = level
			break
		}
	}
	if upgradeLevel.Uuid == 0 ||
		member.MemberLevel.Priority > upgradeLevel.Priority ||
		member.MemberLevel.Uuid == upgradeLevel.Uuid {
		return
	}
	// 更新会员等级ID
	if err := memberRepo.Update(member.Uuid, map[string]any{"member_level_uuid": upgradeLevel.Uuid}); err != nil {
		return
	}
	// 添加等级变动日志
	if _, err := repository.NewMemberLevelLogRepo(db).Create(model.MemberLevelLog{
		MemberUuid: memberUuid,
		OldLevelId: member.MemberLevelUuid,
		NewLevelId: upgradeLevel.Uuid,
		ChangeType: constant.MemberLevelLogTypeAutoUpgrade,
	}); err != nil {
		return
	}
}

// checkCanUpgrade 检查会员是否可以升级
func (s *memberSrv) checkCanUpgrade(member model.Member, level model.MemberLevel) bool {
	if (level.OpenMoney == 1 && level.OpenPoint == 1) &&
		(member.AccumulatedConsumptionAmount >= level.UpgradeMoney && member.Point >= level.UpgradePoint) {
		return true
	}
	if level.OpenMoney == 1 && member.AccumulatedConsumptionAmount >= level.UpgradeMoney {
		return true
	}
	if level.OpenPoint == 1 && member.Point >= level.UpgradePoint {
		return true
	}
	return false
}

func (s *memberSrv) confirmRechargeOrderResp(companyUuid uint64, rechargeOrderUuid uint64) resp.ConfirmRechargeOrder {
	rechargeOrderRepo := repository.NewMemberRechargeOrderRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := rechargeOrderRepo.GetByUuid(rechargeOrderUuid, rechargeOrderRepo.WithPaymentOrders(), rechargeOrderRepo.WithPaymentOrderPaymentMethod())

	paymentMethods := make([]string, 0, len(rechargeOrder.PaymentOrders))
	for _, order := range rechargeOrder.PaymentOrders {
		paymentMethods = append(paymentMethods, order.PaymentMethodName)
	}
	return resp.ConfirmRechargeOrder{
		Amount:         rechargeOrder.Amount,
		ActualAmount:   s.getActualAmount(rechargeOrder.PaymentOrders),
		ChargeDue:      rechargeOrder.ChargeDue,
		PaymentMethods: paymentMethods,
	}
}

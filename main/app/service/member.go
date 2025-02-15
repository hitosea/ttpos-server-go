package service

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	errors2 "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/cryptor"
)

// IMemberSrv 定义会员服务接口
type IMemberSrv interface {
	GetLevels(companyUuid uint64) resp.MemberLevelList                                                                                      // 获取等级列表
	SearchMember(companyUuid uint64, keyword string) resp.SearchMemberList                                                                  // 模糊搜索
	AddMember(companyUuid uint64, addMemberReq req.AddMemberReq) error                                                                      // 添加会员
	GetRechargeMember(companyUuid uint64, memberUuid uint64) resp.RechargeMember                                                            // 获取充值会员信息
	GetPendingRechargeOrder(companyUuid uint64) resp.PendingRechargeOrder                                                                   // 获取进行中的会员充值订单
	CreateRechargeOrder(companyUuid uint64, createRechargeOrderReq req.CreateRechargeOrderReq) (resp.PendingRechargeOrder, error)           // 创建充值订单
	AddPaymentMethod(companyUuid uint64, addPaymentMethod req.RechargeOrderAddPaymentMethodReq) (resp.PendingRechargeOrder, error)          // 充值订单添加支付方式
	CancelPaymentMethod(companyUuid uint64, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.PendingRechargeOrder, error) // 充值订单撤销支付方式
	ConfirmRechargeOrder(companyUuid uint64, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error)           // 确认充值订单
}

// memberSrv 会员服务结构体
type memberSrv struct {
	dbm              *database.DBManager // 数据库管理器
	paymentMethodSrv IPaymentMethodSrv
	settingSrv       setting.ISrv
}

// NewMemberSrv 创建新的会员服务
func NewMemberSrv(dbm *database.DBManager, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv) IMemberSrv {
	return NewMemberSrvImpl(dbm, paymentMethodSrv, settingSrv)
}

// NewMemberSrvImpl 创建新的会员服务实现
func NewMemberSrvImpl(dbm *database.DBManager, paymentMethodSrv IPaymentMethodSrv, settingSrv setting.ISrv) IMemberSrv {
	return &memberSrv{
		dbm:              dbm,
		paymentMethodSrv: paymentMethodSrv,
		settingSrv:       settingSrv,
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
func (s *memberSrv) AddMember(companyUuid uint64, addMemberReq req.AddMemberReq) error {
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
		logger.Logger.Error("添加会员失败", zap.Error(err))
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

// GetPendingRechargeOrder 获取充值会员信息
func (s *memberSrv) GetPendingRechargeOrder(companyUuid uint64) resp.PendingRechargeOrder {
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	pendingRechargeOrder := memberRepo.GetPendingRechargeOrder(0, memberRepo.WithPaidPaymentOrder())
	respPaymentOrders := make([]resp.PaymentOrder, 0)
	if len(pendingRechargeOrder.PaymentOrders) > 0 {
		var respPaymentOrder resp.PaymentOrder
		for _, paymentOrder := range pendingRechargeOrder.PaymentOrders {
			respPaymentOrder.Uuid = paymentOrder.Uuid
			respPaymentOrder.Amount = paymentOrder.Amount
			respPaymentOrder.PaymentMethodUuid = paymentOrder.PaymentTypeUuid
		}
	}
	return resp.PendingRechargeOrder{
		MemberUuid:    pendingRechargeOrder.MemberUuid,
		Uuid:          pendingRechargeOrder.Uuid,
		RechargeMoney: pendingRechargeOrder.RechargeAmount,
		GiftMoney:     pendingRechargeOrder.GiftAmount,
		GiftPoint:     pendingRechargeOrder.GiftPoint,
		PaymentOrders: respPaymentOrders,
	}
}

// CreateRechargeOrder 创建充值订单
func (s *memberSrv) CreateRechargeOrder(companyUuid uint64, rechargeOrderReq req.CreateRechargeOrderReq) (resp.PendingRechargeOrder, error) {
	var orderResp resp.PendingRechargeOrder
	// 判断会员是否存在
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	member := memberRepo.GetByUuid(rechargeOrderReq.RechargeOrderUuid)
	if member.ID == 0 {
		return orderResp, errors.New("会员不存在")
	}
	if rechargeOrderReq.RechargeOrderUuid != 0 {
		// 如果已经存在已支付的payment_order，则充值金额不能小于现有的 "充值金额不能小于已充值金额"
		rechargeOrder := memberRepo.GetPendingRechargeOrder(rechargeOrderReq.RechargeOrderUuid, memberRepo.WithPaidPaymentOrder())
		if rechargeOrder.Uuid != 0 {
			oldRechargeAmount := rechargeOrder.Amount
			var paidAmount float64
			for _, paymentOrder := range rechargeOrder.PaymentOrders {
				paidAmount = paidAmount + paymentOrder.PaymentAmount
			}
			if rechargeOrderReq.RechargeAmount < paidAmount {
				return orderResp, errors.New("充值金额不能小于已充值金额")
			}
			if err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
				// 更新充值订单信息
				err := repository.NewMemberRepo(tx).UpdateRechargeOrder(rechargeOrder.Uuid, map[string]any{
					"recharge_amount": rechargeOrderReq.RechargeAmount,
					"gift_amount":     rechargeOrderReq.GiftAmount,
					"gift_point":      rechargeOrderReq.GiftPoint,
					"member_uuid":     rechargeOrderReq.MemberUuid,
					"staff_uuid":      rechargeOrderReq.StaffUuid,
				})
				// 会员充值操作日志
				operationData, _ := json.Marshal(map[string]any{
					"recharge_money":     rechargeOrderReq.RechargeAmount,
					"old_recharge_money": oldRechargeAmount,
				})
				err = repository.NewMemberRechargeOperationRepo(tx).Add(model.MemberRechargeOrderOperationLog{
					OperatorName:      rechargeOrderReq.StaffName,
					OperatorEmail:     rechargeOrderReq.StaffEmail,
					Client:            rechargeOrderReq.Source,
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
				logger.Logger.Error("修改充值订单失败", zap.Error(err))
				return orderResp, errors.New("修改充值订单失败")
			}

			// 返回充值订单信息
			return s.GetPendingRechargeOrder(companyUuid), nil
		}
	}
	err := s.dbm.GetDB(companyUuid).Transaction(func(tx *gorm.DB) error {
		// 创建充值订单
		order, err := repository.NewMemberRepo(tx).CreateRechargeOrder(model.MemberRechargeOrder{
			RechargeAmount: rechargeOrderReq.RechargeAmount,
			GiftAmount:     rechargeOrderReq.GiftAmount,
			GiftPoint:      rechargeOrderReq.GiftPoint,
			MemberUuid:     rechargeOrderReq.MemberUuid,
			StaffUuid:      rechargeOrderReq.StaffUuid,
		})
		if err != nil {
			return err
		}
		// 会员充值操作日志
		err = repository.NewMemberRechargeOperationRepo(tx).Add(model.MemberRechargeOrderOperationLog{
			OperatorName:      rechargeOrderReq.StaffName,
			OperatorEmail:     rechargeOrderReq.StaffEmail,
			Client:            rechargeOrderReq.Source,
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
func (s *memberSrv) AddPaymentMethod(companyUuid uint64, addPaymentMethodReq req.RechargeOrderAddPaymentMethodReq) (resp.PendingRechargeOrder, error) {
	var orderResp resp.PendingRechargeOrder
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := memberRepo.GetPendingRechargeOrder(addPaymentMethodReq.RechargeOrderUuid)
	if rechargeOrder.Uuid == 0 {
		return orderResp, errors.New("充值订单不存在")
	}

	// todo 在线支付订单暂不处理

	// 根据Uuid 获取支付方式
	paymentMethod := repository.NewPaymentMethodRepo(s.dbm.GetDB(companyUuid)).GetByUuid(addPaymentMethodReq.PaymentMethodUuid)
	// 支付方式是否可用
	if !s.paymentMethodSrv.IsAvailable(paymentMethod, addPaymentMethodReq.CompanySetting) {
		return orderResp, errors.New("支付方式未开启")
	}
	if paymentMethod.Code == constant.PaymentMethodCodeBalance {
		return orderResp, errors.New("不能使用余额支付充值")
	}
	// 计算支付手续费
	paymentCommissionFee := s.paymentMethodSrv.CalculatePaymentCommissionFee(paymentMethod, addPaymentMethodReq.PaymentAmount)

	// 获取所有已支付的支付订单，计算已支付金额
	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	paidPaymentOrders, err := paymentOrderRepo.GetPaymentOrderList(paymentOrderRepo.WhereStatus(constant.PaymentOrderStatusPaid))
	if err != nil {
		return orderResp, errors2.ErrInternal
	}
	var paidPaymentOrderSum float64
	for _, paidPaymentOrder := range paidPaymentOrders {
		paidPaymentOrderSum = paidPaymentOrderSum + paidPaymentOrder.PaymentAmount
	}
	// 所有已支付的订单金额 >= 充值订单的 pay_Price
	if paidPaymentOrderSum >= rechargeOrder.RechargeAmount {
		return orderResp, errors.New("当前已足额")
	}
	paidPaymentOrderSum = paidPaymentOrderSum + addPaymentMethodReq.PaymentAmount
	if paymentMethod.Code != constant.PaymentMethodCodeCash && paidPaymentOrderSum > rechargeOrder.RechargeAmount {
		return orderResp, errors.New("非现金支付不能大于应收")
	}

	// 支付订单总金额 = 支付金额 + 支付手续费
	amount := addPaymentMethodReq.PaymentAmount + paymentCommissionFee

	paymentOrder, err := paymentOrderRepo.GetPaymentOrder(
		paymentOrderRepo.WhereRelatedUuid(addPaymentMethodReq.RechargeOrderUuid),
		paymentOrderRepo.WherePaymentTypeUuid(addPaymentMethodReq.PaymentMethodUuid))

	if err != nil {
		return orderResp, errors2.ErrInternal
	}

	if paymentOrder.Uuid != 0 { // 更新
		err = paymentOrderRepo.Update(paymentOrder.Uuid, map[string]any{
			"payment_amount":         addPaymentMethodReq.PaymentAmount,
			"amount":                 amount,
			"payment_commission_fee": paymentCommissionFee,
		})
	} else {
		currencySetting, err := s.settingSrv.GetCurrencySetting(companyUuid)
		if err != nil {
			return orderResp, errors2.ErrInternal
		}
		_, err = paymentOrderRepo.Create(model.PaymentOrder{
			PaymentTypeName:      paymentMethod.PaymentName,
			PaymentTypeUuid:      paymentMethod.Uuid,
			PaymentFeePercent:    paymentMethod.FeePercent,
			RelatedUuid:          rechargeOrder.Uuid,
			CurrencyUnit:         currencySetting.Unit, // 留档使用
			PaymentAmount:        addPaymentMethodReq.PaymentAmount,
			PaymentCommissionFee: paymentCommissionFee,
			Amount:               amount,
			Status:               constant.PaymentOrderStatusPaid, // ToDo 手动添加在线支付标为0，处理lianlianpay
		})
	}

	return s.GetPendingRechargeOrder(companyUuid), nil
}

// CancelPaymentMethod 充值订单撤销支付方式
func (s *memberSrv) CancelPaymentMethod(companyUuid uint64, cancelPaymentMethod req.RechargeOrderCancelPaymentMethodReq) (resp.PendingRechargeOrder, error) {
	var orderResp resp.PendingRechargeOrder
	memberRepo := repository.NewMemberRepo(s.dbm.GetDB(companyUuid))
	rechargeOrder := memberRepo.GetPendingRechargeOrder(cancelPaymentMethod.RechargeOrderUuid)
	if rechargeOrder.Uuid == 0 {
		return orderResp, errors.New("充值订单不存在")
	}
	paymentOrderRepo := repository.NewPaymentOrderRepo(s.dbm.GetDB(companyUuid))
	// 判断支付订单
	paymentOrder, err := paymentOrderRepo.GetPaymentOrder(paymentOrderRepo.WhereUuid(cancelPaymentMethod.RechargeOrderUuid))
	if err != nil || paymentOrder.Uuid == 0 {
		return orderResp, errors.New("支付订单不存在")
	}
	if err := memberRepo.CancelPaymentOrder(paymentOrder.Uuid); err != nil {
		return orderResp, errors2.ErrInternal
	}
	return s.GetPendingRechargeOrder(companyUuid), nil
}

// ConfirmRechargeOrder 确认充值订单
func (s *memberSrv) ConfirmRechargeOrder(companyUuid uint64, confirmRechargeOrderReq req.ConfirmRechargeOrder) (resp.ConfirmRechargeOrder, error) {
	return resp.ConfirmRechargeOrder{}, nil
}

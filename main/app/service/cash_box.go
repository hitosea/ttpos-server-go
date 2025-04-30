package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"

	"github.com/shopspring/decimal"
)

type UpdateCashBalanceParam struct {
	Amount         float64 // 钱箱变动的金额
	CashWithdrawal float64 // 中途取出金额
	CashDeposit    float64 // 中途存入金额
	Scene          int     // 场景, 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零 8-交班
	OrderUuid      uint64  // 订单ID,场景为1时填销售订单uuid、场景为6时填充值订单uuid
	Remark         string  // 备注
}

type ICashBoxSrv interface {
	UpdateBalance(ctx context.Context, param UpdateCashBalanceParam) error
	GetBalance(ctx context.Context) (float64, error)
	GetRecord(ctx context.Context) (model.CashBox, error)
}

func NewCashBoxSrv(dbm *database.DBManager) ICashBoxSrv {
	return NewCashBoxSrvImpl(dbm)
}

type cashBoxSrv struct {
	dbm      *database.DBManager
	uuidLock lock.Lock
}

func NewCashBoxSrvImpl(dbm *database.DBManager) ICashBoxSrv {
	return &cashBoxSrv{
		dbm:      dbm,
		uuidLock: lock.NewSystemLock(),
	}
}

// UpdateBalance 更新钱箱余额
func (s *cashBoxSrv) UpdateBalance(ctx context.Context, param UpdateCashBalanceParam) error {
	companyUuid := ctx.GetCompanyUuid()

	if ctx.NoLock() {
		s.uuidLock.LockUuid(companyUuid)
		defer s.uuidLock.UnlockUuid(companyUuid)
		ctx.AddLock()
	}

	tx := ctx.GetDB()
	var err error
	cashBoxRepo := repository.NewCashBoxRepo(tx)
	cashBox := cashBoxRepo.Get()

	if cashBox.Uuid != 0 { // 已存在钱箱
		if err = cashBoxRepo.Update(cashBox.Uuid, map[string]any{
			"frozen_balance":  decimal.NewFromFloat(param.Amount).Add(decimal.NewFromFloat(cashBox.FrozenBalance)).Truncate(2).InexactFloat64(),
			"cash_withdrawal": param.CashWithdrawal + cashBox.CashWithdrawal,
			"cash_deposit":    param.CashDeposit + cashBox.CashDeposit,
		}); err != nil {
			return errors.WithMessage(err, "更新钱箱失败")
		}
	} else { // 不存在钱箱
		if cashBox, err = cashBoxRepo.Create(model.CashBox{
			FrozenBalance: param.Amount,
		}); err != nil {
			return errors.WithMessage(err, "更新钱箱失败")
		}
	}

	log := model.CashBoxLog{
		Scene:  param.Scene,
		Amount: param.Amount,
		Remark: param.Remark,
	}

	switch param.Scene {
	case constant.CashBoxLogSceneRecharge:
		log.RelatedUuid = param.OrderUuid
	case constant.CashBoxLogSceneRefund:
		log.RefundOrderAmountUuid = param.OrderUuid
	}
	_, err = repository.NewCashBoxLogRepo(tx).Create(log)

	return errors.WithMessage(err)
}

// GetBalance 获取钱箱余额
func (s *cashBoxSrv) GetBalance(ctx context.Context) (float64, error) {
	cashBox, err := s.GetRecord(ctx)
	if err != nil {
		return 0, errors.WithMessage(err, "获取钱箱余额失败")
	}
	return cashBox.GetBalance(), nil
}

// GetRecord 获取钱箱记录
func (s *cashBoxSrv) GetRecord(ctx context.Context) (model.CashBox, error) {
	return repository.NewCashBoxRepo(ctx.GetDB()).Get(), nil
}

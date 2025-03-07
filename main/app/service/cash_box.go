package service

import (
	"gorm.io/gorm"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
)

type UpdateCashBalanceParam struct {
	CashBoxLogType int
	Amount         float64
	Scene          int
	OrderUuid      uint64
}

type ICashBoxSrv interface {
	UpdateBalance(ctx context.Context, param UpdateCashBalanceParam) error
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
	cashBoxLogType := param.CashBoxLogType
	if !slices.Contains([]int{constant.CashBoxLogTypeIn, constant.CashBoxLogTypeOut}, cashBoxLogType) {
		return errors.New("钱箱操作类型错误")
	}
	companyUuid := ctx.GetCompanyUuid()

	if ctx.NoLock() {
		s.uuidLock.LockUuid(companyUuid)
		defer s.uuidLock.UnlockUuid(companyUuid)
		ctx.AddLock()
	}

	amount := param.Amount
	fn := func(tx *gorm.DB) error {

		var err error
		cashBoxRepo := repository.NewCashBoxRepo(tx)
		cashBox := cashBoxRepo.Get()

		if cashBox.Uuid != 0 { // 已存在钱箱
			if err = cashBoxRepo.Update(cashBox.Uuid, map[string]any{
				"balance": amount + cashBox.Balance,
			}); err != nil {
				return errors.WithMessage(err, "更新钱箱失败")
			}
		} else { // 不存在钱箱
			if cashBox, err = cashBoxRepo.Create(model.CashBox{
				Balance: amount,
			}); err != nil {
				return errors.WithMessage(err, "更新钱箱失败")
			}
		}

		log := model.CashBoxLog{
			Type:   cashBoxLogType,
			Scene:  param.Scene,
			Amount: amount,
		}

		// TODO 钱箱日志，如果是充值订单付现金，订单如何保存；如果是退款充值订单，又该如何保存
		switch param.Scene {
		//case constant.CashBoxLogSceneRecharge:
		//	log.PaymentBillUuid = param.OrderUuid
		case constant.CashBoxLogSceneRefund:
			log.RefundOrderAmountUuid = param.OrderUuid
		}
		_, err = repository.NewCashBoxLogRepo(tx).Create(log)

		return errors.WithMessage(err)
	}

	tx := ctx.GetDB()
	if tx == nil {
		return s.dbm.GetDB(ctx.GetCompanyUuid()).Transaction(func(tx *gorm.DB) error {
			return fn(tx)
		})
	} else {
		return fn(tx)
	}
}

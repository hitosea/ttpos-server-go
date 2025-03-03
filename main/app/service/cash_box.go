package service

import (
	"errors"
	"gorm.io/gorm"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/lock"
)

type ICashBoxSrv interface {
	UpdateBalance(ctx context.Context, cashLogType int, amount float64, orderUuid uint64) error
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
func (s *cashBoxSrv) UpdateBalance(ctx context.Context, cashBoxLogType int, amount float64, orderUuid uint64) error {
	if !slices.Contains([]int{constant.CashBoxLogTypeIn, constant.CashBoxLogTypeOut}, cashBoxLogType) {
		return errors.New("钱箱操作类型错误")
	}
	companyUuid := ctx.GetCompanyUuid()
	s.uuidLock.LockUuid(companyUuid)
	defer s.uuidLock.UnlockUuid(companyUuid)

	fn := func(tx *gorm.DB) error {

		var err error
		cashBoxRepo := repository.NewCashBoxRepo(tx)
		cashBox := cashBoxRepo.Get()

		if cashBox.Uuid != 0 { // 已存在钱箱
			if err = cashBoxRepo.Update(cashBox.Uuid, map[string]any{
				"balance": amount + cashBox.Balance,
			}); err != nil {
				return errors.New("更新钱箱失败")
			}
		} else { // 不存在钱箱
			if cashBox, err = cashBoxRepo.Create(model.CashBox{
				Balance: amount,
			}); err != nil {
				return errors.New("更新钱箱失败")
			}
		}

		_, err = repository.NewCashBoxLogRepo(tx).Create(model.CashBoxLog{
			Type:            cashBoxLogType,
			Scene:           constant.CashBoxLogSceneRecharge,
			Amount:          amount,
			PaymentBillUuid: orderUuid,
		})

		return err
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

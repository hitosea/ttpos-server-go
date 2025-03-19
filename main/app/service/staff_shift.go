package service

import (
	"fmt"
	"math/rand"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	printerService "ttpos-server-go/app/printer/service"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/duke-git/lancet/convertor"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type IStaffShiftSrv interface {
	GetCashierReport(ctx context.Context) (*resp.CashierReportResp, error)
	SubmitCashierReport(ctx context.Context, req req.CashierReportReq) error
	CreateWorkingLog(staff model.Staff) (model.StaffShiftLog, error)
	GetShiftInfo(ctx context.Context) (*resp.ShiftInfo, error)                          // 获取交班信息
	SubmitShift(ctx context.Context, req req.SubmitShiftReq) (*resp.ShiftSubmit, error) // 提交交班
}

func NewStaffShiftSrv(cache cache.Cache, dbm *database.DBManager) IStaffShiftSrv {
	return NewShiftSrvImpl(cache, dbm)
}

type staffShiftSrv struct {
	dbm            *database.DBManager
	cache          cache.Cache
	cacheKeyPrefix string
}

func NewShiftSrvImpl(cache cache.Cache, dbm *database.DBManager) IStaffShiftSrv {
	return &staffShiftSrv{
		dbm:            dbm,
		cache:          cache,
		cacheKeyPrefix: "__USERSHIFTLOG_GENERATENUMBER__",
	}
}

// CreateWorkingLog 创建当班记录
func (s *staffShiftSrv) CreateWorkingLog(staff model.Staff) (model.StaffShiftLog, error) {
	shiftLogRepo := repository.NewShiftLogRepo(s.dbm.GetDB(staff.CompanyUuid))
	previousShiftCash, _ := shiftLogRepo.GetPreviousShiftCash()
	startTime := staff.CashierLoginTime
	if startTime == 0 {
		startTime = time.Now().Unix()
	}

	shiftLog, _ := shiftLogRepo.Create(model.StaffShiftLog{
		StaffUuid:         staff.Uuid,
		ShiftNo:           s.generateNumber(),
		PreviousShiftCash: previousShiftCash,
		CurrentCashTotal:  previousShiftCash,
		CashLeft:          previousShiftCash,
		ShiftStartTime:    startTime,
		ShiftEndTime:      0,
	})
	return shiftLog, nil
}

func (s *staffShiftSrv) generateNumber() string {
	// 日期部分：年月日
	datePart := time.Now().Format("20060102")
	// 固定部分
	fixedPart := "01"
	// 随机部分：8位数字
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	no := datePart + fixedPart + randomPart
	cacheKey := s.cacheKeyPrefix + no
	if _, ok := s.cache.Get(cacheKey); ok {
		return s.generateNumber()
	}
	s.cache.Set(cacheKey, no, 86400*time.Second)
	// 组合订单号
	return no
}

// GetShiftInfo 获取交班信息
func (s *staffShiftSrv) GetShiftInfo(ctx context.Context) (*resp.ShiftInfo, error) {
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	// 查询交班记录
	shiftLogResp := repository.NewShiftLogRepo(db)
	log, err := shiftLogResp.GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.New("当前班次不存在")
	}
	if log.IsHandedOver() {
		return nil, errors.New("当前班次已交班")
	}
	// 统计当班用餐订/充值单退款金额
	saleRefundAmount := repository.NewStatisticsRepo(db).CountShiftSaleRefundAmount(log.ShiftNo)
	rechargeRefundAmount := repository.NewStatisticsRepo(db).CountShiftRechargeRefundAmount(log.ShiftNo)
	totalRefundAmount := decimal.NewFromFloat(saleRefundAmount.RefundAmount.Float64).
		Add(decimal.NewFromFloat(rechargeRefundAmount.RefundAmount.Float64)).
		Round(2).InexactFloat64()
	// 统计当班支付方式收入
	paymentMethodIncomeList, _ := s.CountShiftPaymentMethodIncome(db, log.ShiftNo, ctx.GetLanguage())

	return &resp.ShiftInfo{
		PreviousShiftCash: log.PreviousShiftCash,
		WithdrawCash:      log.WithdrawCash,
		DepositCash:       log.DepositCash,
		CurrentCashTotal:  log.CurrentCashTotal,
		RefundAmount:      totalRefundAmount,
		PaymentMethodIncome: resp.PaymentMethodIncomeList{
			List: paymentMethodIncomeList,
		},
	}, nil
}

// SubmitShift 提交交班
func (s *staffShiftSrv) SubmitShift(ctx context.Context, req req.SubmitShiftReq) (*resp.ShiftSubmit, error) {
	// 验证参数
	if req.WithdrawCash < 0 {
		return nil, errors.New("取出金额不能小于0")
	}
	if req.LeaveCash < 0 {
		return nil, errors.New("遗留现金不能小于0")
	}
	// 获取当前班次
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	shiftLogRepo := repository.NewShiftLogRepo(db)
	shiftLog, err := shiftLogRepo.GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.New("当前班次不存在")
	}
	if shiftLog.IsHandedOver() {
		return nil, errors.New("当前班次已交班")
	}
	withdrawCash := decimal.NewFromFloat(req.WithdrawCash)
	leaveCash := decimal.NewFromFloat(req.LeaveCash)
	// 当前班次取出金额 + 遗留现金 = 当前钱箱现金总计
	if !withdrawCash.Add(leaveCash).
		Equal(decimal.NewFromFloat(shiftLog.CurrentCashTotal)) {
		return nil, errors.New("输入的本班取出現金和本班遗留备用金总额与当前钱箱现金总计不符")
	}
	// 当前班次支付方式收入
	paymentMethodIncomeList, cashAmount := s.CountShiftPaymentMethodIncome(db, shiftLog.ShiftNo, ctx.GetLanguage())
	incomes, _ := convertor.ToJson(paymentMethodIncomeList)
	// 当前班次营业额
	totalBusiness := 0
	// 更新当班记录
	shiftLogRepo.Update(shiftLog, map[string]interface{}{
		"status":             constant.StaffHandedOver,      // 交班状态
		"current_cash_total": leaveCash.InexactFloat64(),    // 当前钱箱现金总计
		"cash_taken_out":     withdrawCash.InexactFloat64(), // 取出现金
		"cash_left":          leaveCash.InexactFloat64(),    // 遗留现金
		"shift_end_time":     time.Now().Unix(),             // 交班时间
		"cash_income":        cashAmount,                    // 现金收入
		"incomes":            incomes,                       // 支付方式收入
		"total_business":     totalBusiness,                 // 营业额
	})

	return &resp.ShiftSubmit{
		CashIncome:   cashAmount,
		CashTakenOut: withdrawCash.InexactFloat64(),
		CashLeft:     leaveCash.InexactFloat64(),
	}, nil
}

// CountShiftPaymentMethodIncome 统计当班支付方式收入
func (s *staffShiftSrv) CountShiftPaymentMethodIncome(db *gorm.DB, shiftNo string, language string) ([]resp.PaymentMethodIncome, float64) {
	paymentMethodAmount := repository.NewStatisticsRepo(db).CountShiftPaymentMethodAmount(shiftNo)
	var (
		paymentMethodIncomeList = make([]resp.PaymentMethodIncome, 0, len(paymentMethodAmount)) // 支付方式收入列表
		cashAmount              decimal.Decimal                                                 // 现金收入
	)
	for _, v := range paymentMethodAmount {
		payAmount := decimal.NewFromFloat(v.PayAmount.Float64)
		refundAmount := decimal.NewFromFloat(v.RefundAmount.Float64)
		amount := payAmount.Sub(refundAmount)
		paymentMethodIncomeList = append(paymentMethodIncomeList, resp.PaymentMethodIncome{
			Name:   v.PaymentName,
			Amount: amount.InexactFloat64(),
		})
		if v.PaymentCode == constant.PaymentMethodCodeCash {
			cashAmount = cashAmount.Add(amount)
		}
	}
	// 统计当班用餐订单免单金额
	freeAmount := repository.NewStatisticsRepo(db).CountShiftSaleFreeAmount(shiftNo)
	if freeAmount.FreeAmount.Float64 > 0 {
		paymentMethodIncomeList = append(paymentMethodIncomeList, resp.PaymentMethodIncome{
			Name:   i18n.Translate(language, "免单金额"),
			Amount: freeAmount.FreeAmount.Float64,
		})
	}

	return paymentMethodIncomeList, cashAmount.InexactFloat64()
}

// GetCashierReport 获取报备信息
func (s *staffShiftSrv) GetCashierReport(ctx context.Context) (*resp.CashierReportResp, error) {
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	// 查询交班记录
	log, err := repository.NewShiftLogRepo(db).GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.New("当前班次不存在")
	}
	if log.IsHandedOver() {
		return nil, errors.New("当前班次已交班")
	}

	// 获取打印机配置
	settingSrv := setting.NewSrv(s.dbm, s.cache)
	printerData, err := printerService.NewPrinterLogSrv(s.dbm, settingSrv).GetStaticOpenCashBoxPrinterConfig(ctx)
	if err != nil {
		return nil, err
	}

	//
	return &resp.CashierReportResp{
		PreviousShiftCash: log.PreviousShiftCash,
		PrinterData:       *printerData,
	}, nil
}

// SubmitCashierReport 提交报备信息
func (s *staffShiftSrv) SubmitCashierReport(ctx context.Context, req req.CashierReportReq) error {
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	shiftLogRepo := repository.NewShiftLogRepo(db)

	// 查询交班记录
	log, err := shiftLogRepo.GetShiftLog(
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return errors.New("当前班次错误，请退出重新登录")
	}
	if log.IsHandedOver() {
		return errors.New("当前班次已交班")
	}
	// todo 是否已报备 暂时不拦截
	// if log.IsReported() {
	// 	return errors.New("当前班次已经报备")
	// }

	// 更新交班记录
	_, err = shiftLogRepo.Update(log, map[string]interface{}{
		"exception_remark": req.ExceptionRemark,
	})
	if err != nil {
		return errors.New("更新交班记录失败")
	}

	return nil
}

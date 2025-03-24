package service

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/business_data_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/printer"
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
	ShiftWithdraw(ctx context.Context, req req.ShiftWithdrawReq) error
	ShiftDeposit(ctx context.Context, req req.ShiftDepositReq) error
	ShiftPrinter(ctx context.Context, req req.ShiftPrinterReq) (*resp.PrinterData, error)
}

func NewStaffShiftSrv(cache cache.Cache, dbm *database.DBManager, cashBoxSrv ICashBoxSrv) IStaffShiftSrv {
	return NewShiftSrvImpl(cache, dbm, cashBoxSrv)
}

type staffShiftSrv struct {
	dbm            *database.DBManager
	cache          cache.Cache
	cacheKeyPrefix string
	cashBoxSrv     ICashBoxSrv
}

func NewShiftSrvImpl(cache cache.Cache, dbm *database.DBManager, cashBoxSrv ICashBoxSrv) IStaffShiftSrv {
	return &staffShiftSrv{
		dbm:            dbm,
		cache:          cache,
		cacheKeyPrefix: "__USERSHIFTLOG_GENERATENUMBER__",
		cashBoxSrv:     cashBoxSrv,
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
	var (
		withdrawCash decimal.Decimal // 取出现金
		leaveCash    decimal.Decimal // 遗留现金
		cashAmount   float64         // 现金收入
	)
	// 获取当前员工
	var (
		staff model.Staff
		db    *gorm.DB = s.dbm.GetDB(ctx.GetCompanyUuid())
	)
	if req.IsBackground {
		staff = repository.NewStaffRepo(db).GetStaff(repository.CommonRepo.WhereByUuid(req.StaffUuid))
	} else {
		staff = ctx.GetStaff()
	}
	if staff.Uuid == 0 {
		return nil, errors.New("员工不存在")
	}
	err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 获取当前班次
		shiftLogRepo := repository.NewShiftLogRepo(db)
		shiftLog, err := shiftLogRepo.GetShiftLog(
			repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
			repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
		)
		if err != nil {
			return errors.New("当前班次不存在")
		}
		if shiftLog.IsHandedOver() {
			return errors.New("当前班次已交班")
		}
		withdrawCash = decimal.NewFromFloat(req.WithdrawCash)
		leaveCash = decimal.NewFromFloat(req.LeaveCash)
		// 后台交班时，取出金额为当前钱箱现金总计，遗留现金为0
		if req.IsBackground {
			withdrawCash = decimal.NewFromFloat(shiftLog.CurrentCashTotal)
			leaveCash = decimal.Zero
		}
		// 当前班次取出金额 + 遗留现金 = 当前钱箱现金总计
		if !withdrawCash.Add(leaveCash).
			Equal(decimal.NewFromFloat(shiftLog.CurrentCashTotal)) {
			return errors.New("输入的本班取出現金和本班遗留备用金总额与当前钱箱现金总计不符")
		}
		// 当前班次支付方式收入
		var paymentMethodIncomeList = make([]resp.PaymentMethodIncome, 0)
		paymentMethodIncomeList, cashAmount = s.CountShiftPaymentMethodIncome(db, shiftLog.ShiftNo, ctx.GetLanguage())
		incomes, _ := convertor.ToJson(paymentMethodIncomeList)
		// 当前班次营业额
		totalBusiness := 0
		// 更新当班记录
		_, err = shiftLogRepo.Update(shiftLog, map[string]interface{}{
			"status":              constant.StaffHandedOver,      // 交班状态
			"previous_shift_cash": shiftLog.CurrentCashTotal,     // 上交班现金总计
			"current_cash_total":  leaveCash.InexactFloat64(),    // 当前钱箱现金总计
			"cash_taken_out":      withdrawCash.InexactFloat64(), // 取出现金
			"cash_left":           leaveCash.InexactFloat64(),    // 遗留现金
			"shift_end_time":      time.Now().Unix(),             // 交班时间
			"cash_income":         cashAmount,                    // 现金收入
			"incomes":             incomes,                       // 支付方式收入
			"total_business":      totalBusiness,                 // 营业额
		})
		if err != nil {
			return errors.New("交班失败")
		}
		// 更新钱箱记录
		if withdrawCash.GreaterThan(decimal.Zero) {
			err := s.cashBoxSrv.UpdateBalance(
				ctx,
				UpdateCashBalanceParam{
					CashBoxLogType: constant.CashBoxLogTypeOut,
					Amount:         withdrawCash.InexactFloat64(),
					Scene:          constant.CashBoxLogSceneShift,
				},
			)
			if err != nil {
				return errors.New("交班失败")
			}
		}
		// 员工下线
		err = repository.NewStaffRepo(db).Update(staff.Uuid, map[string]any{
			"cashier_online":     0,
			"cashier_login_time": 0,
			"duty_no":            "",
		})
		if err != nil {
			return errors.New("交班失败")
		}
		deviceRepo := repository.NewDeviceRepo(db)
		device, err := deviceRepo.GetDevice(
			deviceRepo.WhereSource(constant.SourceCashier),
			deviceRepo.WhereSn(staff.BindKey),
		)
		if err != nil {
			return errors.New("交班失败")
		}
		repository.NewDeviceRepo(db).UpdateDevice(device.Uuid, map[string]any{
			"finally_login_uuid": 0,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

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

	// 返回报备信息
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
	// 更新交班记录
	_, err = shiftLogRepo.Update(log, map[string]interface{}{
		"exception_remark": req.ExceptionRemark,
	})
	if err != nil {
		return errors.New("更新交班记录失败")
	}

	return nil
}

// ShiftWithdraw 交班取钱
func (s *staffShiftSrv) ShiftWithdraw(ctx context.Context, req req.ShiftWithdrawReq) error {
	if req.WithdrawCash <= 0 {
		return errors.New("请输入正确金额")
	}
	// 验证参数, 大于0, 最多小数点后两位
	reg := regexp.MustCompile(`^([1-9]\d*|0)(\.\d{1,2})?$`)
	if !reg.MatchString(convertor.ToString(req.WithdrawCash)) {
		return errors.New("请输入正确金额")
	}
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	shiftLogRepo := repository.NewShiftLogRepo(db)
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

	err = repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 更新交班记录
		_, err = shiftLogRepo.Update(log, map[string]interface{}{
			"withdraw_cash":      gorm.Expr("withdraw_cash + ?", req.WithdrawCash),
			"current_cash_total": gorm.Expr("current_cash_total - ?", req.WithdrawCash),
			"cash_left":          gorm.Expr("cash_left - ?", req.WithdrawCash),
		})
		if err != nil {
			return errors.New("取钱失败")
		}
		// 更新钱箱记录
		err = s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
			CashBoxLogType: constant.CashBoxLogTypeOut,
			Amount:         req.WithdrawCash,
			CashWithdrawal: req.WithdrawCash,
			Scene:          constant.CashBoxLogSceneOut,
		})
		if err != nil {
			return errors.New("取钱失败")
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// ShiftDeposit 交班存钱
func (s *staffShiftSrv) ShiftDeposit(ctx context.Context, req req.ShiftDepositReq) error {
	// 验证参数, 大于0, 最多小数点后两位
	if req.DepositCash <= 0 {
		return errors.New("请输入正确金额")
	}
	reg := regexp.MustCompile(`^([1-9]\d*|0)(\.\d{1,2})?$`)
	if !reg.MatchString(convertor.ToString(req.DepositCash)) {
		return errors.New("请输入正确金额")
	}
	staff := ctx.GetStaff()
	db := s.dbm.GetDB(staff.CompanyUuid)
	shiftLogRepo := repository.NewShiftLogRepo(db)
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

	err = repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 更新交班记录
		_, err = shiftLogRepo.Update(log, map[string]interface{}{
			"deposit_cash":       gorm.Expr("deposit_cash + ?", req.DepositCash),
			"current_cash_total": gorm.Expr("current_cash_total + ?", req.DepositCash),
			"cash_left":          gorm.Expr("cash_left + ?", req.DepositCash),
		})
		if err != nil {
			return errors.New("存钱失败")
		}
		// 更新钱箱记录
		err = s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
			CashBoxLogType: constant.CashBoxLogTypeIn,
			Amount:         req.DepositCash,
			CashDeposit:    req.DepositCash,
			Scene:          constant.CashBoxLogSceneIn,
		})
		if err != nil {
			return errors.New("存钱失败")
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// ShiftPrinter 交班打印
func (s *staffShiftSrv) ShiftPrinter(ctx context.Context, req req.ShiftPrinterReq) (*resp.PrinterData, error) {
	staff := ctx.GetStaff()
	shiftLogRepo := repository.NewShiftLogRepo(ctx.GetDB())

	// 查询交班记录
	log, err := shiftLogRepo.GetShiftLog(
		shiftLogRepo.WithStaff(),
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
	)
	if err != nil {
		return nil, errors.New("当前班次错误，请退出重新登录")
	}
	if log.IsHandedOver() {
		return nil, errors.New("当前班次已交班")
	}

	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:             log.TotalBusiness,
		TotalReceivedPrice:     log.CurrentCashTotal,
		TotalPayPrice:          21230,
		TotalPayFeeMoney:       2110,
		TotalServiceMoney:      120,
		TotalTaxMoney:          10124,
		TotalUserDiscountMoney: 120,
		TotalDiscountMoney:     120,
		TotalFreeOrderPrice:    120,
		TotalRefundMoney:       10,
		TotalOrderNum:          1230,
		TotalPeopleNum:         120,
		TotalProductNum:        320,
		TotalTableNum:          120,
		AvgOrderPrice:          620,
		MinOrderPrice:          120,
		MaxOrderPrice:          1200,
		AllTableOrderNum:       1230,
		AllTablePeopleNum:      120,
		AllTableAvgOrderPrice:  620,
		AllTableMinOrderPrice:  120,
		AllTableMaxOrderPrice:  1200,
		AllTablePeopleAvg:      10,
		PaymentMethodIncomes: []business_data_resp.PaymentMethodIncome{
			{
				Name:     "现金",
				OrderNum: 1,
				Amount:   123213,
				Code:     40,
			},
			{
				Name:     "支付宝",
				OrderNum: 1,
				Amount:   24121,
				Code:     41,
			},
			{
				Name:     "微信支付",
				OrderNum: 1,
				Amount:   123213,
				Code:     42,
			},
		},
		AbnormalData: business_data_resp.AbnormalData{},
		MemberData: business_data_resp.MemberData{
			RechargeAmount: 120,
			GiftMoney:      120,
			GiftPoints:     120,
		},
		PeakHourList: func() []business_data_resp.PeakHour {
			peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
				uint(log.ShiftStartTime),
				uint(log.ShiftEndTime),
				log.StaffUuid,
			)
			if err != nil {
				return []business_data_resp.PeakHour{}
			}
			return peakHours
		}(),
		CategoryList: []business_data_resp.Category{
			{
				Name:     "12",
				SalesNum: 1,
				Prices:   323,
			},
			{
				Name:     "121232",
				SalesNum: 2,
				Prices:   23,
			},
		},
		PercentageList: []business_data_resp.Percentage{
			{
				TaxRate:        120,
				ConsumptionTax: 120,
			},
			{
				TaxRate:        110,
				ConsumptionTax: 2120,
			},
		},
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx).PrintingHandoverOrder(
		&log,
		&businessData,
		1,
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

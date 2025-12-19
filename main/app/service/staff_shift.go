package service

import (
	"encoding/json"
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
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/duke-git/lancet/convertor"
	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IStaffShiftSrv interface {
	GetCashierReport(ctx context.Context) (*resp.CashierReportResp, error)
	SubmitCashierReport(ctx context.Context, req req.CashierReportReq) error
	CreateWorkingLog(ctx context.Context, staff model.Staff) (model.StaffShiftLog, error)
	GetShiftInfo(ctx context.Context) (*resp.ShiftInfo, error)                          // 获取交班信息
	SubmitShift(ctx context.Context, req req.SubmitShiftReq) (*resp.ShiftSubmit, error) // 提交交班
	GetCachedSubmitShift(ctx context.Context) (*resp.ShiftSubmit, error)                // 提交交班
	ShiftWithdraw(ctx context.Context, req req.ShiftWithdrawReq) error
	ShiftDeposit(ctx context.Context, req req.ShiftDepositReq) error
	ShiftPrinter(ctx context.Context, req req.ShiftPrinterReq, firstExecution int, openMoneybox bool, staffs ...model.Staff) (*resp.PrinterData, error)
	CreateShiftSnapshot(ctx context.Context, shiftLog model.StaffShiftLog) error
}

func NewStaffShiftSrv(cache cache.Cache, dbm *database.DBManager, cashBoxSrv ICashBoxSrv, statisticsSrv IStatisticsSrv) IStaffShiftSrv {
	return NewShiftSrvImpl(cache, dbm, cashBoxSrv, statisticsSrv)
}

type staffShiftSrv struct {
	dbm            *database.DBManager
	cache          cache.Cache
	cacheKeyPrefix string
	cashBoxSrv     ICashBoxSrv
	statisticsSrv  IStatisticsSrv
	MaxAmount      string
}

func NewShiftSrvImpl(cache cache.Cache, dbm *database.DBManager, cashBoxSrv ICashBoxSrv, statisticsSrv IStatisticsSrv) IStaffShiftSrv {
	return &staffShiftSrv{
		dbm:            dbm,
		cache:          cache,
		cacheKeyPrefix: "__USERSHIFTLOG_GENERATENUMBER__",
		cashBoxSrv:     cashBoxSrv,
		statisticsSrv:  statisticsSrv,
		MaxAmount:      "1000000000000",
	}
}

// CreateWorkingLog 创建当班记录
func (s *staffShiftSrv) CreateWorkingLog(ctx context.Context, staff model.Staff) (model.StaffShiftLog, error) {
	shiftLogRepo := repository.NewShiftLogRepo(s.dbm.GetDB(staff.CompanyUuid))
	cashBox := repository.NewCashBoxRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
	previousShiftCash := cashBox.GetBalance()
	startTime := staff.CashierLoginTime
	if startTime == 0 {
		startTime = time.Now().Unix()
	}

	var erpnextOpenPosEntryName string
	// 调用rpc接口生成PosEntry
	companySetting := ctx.GetCompanySetting()
	company := ctx.GetCompany()
	if company.IsOpenErp() && companySetting.ErpnextSiteCode != "" {
		erpSrv := erp.NewIErpSrv(s.dbm)
		openPosEntryName, err := erpSrv.OpenPosEntry(ctx.GetContext(), req.OpenPosEntryReq{
			SiteCode:        companySetting.ErpnextSiteCode,
			PosProfileName:  companySetting.ErpnextPosProfileName,
			CashierEmail:    companySetting.ErpnextAdminEmail,
			CompanyAbbr:     companySetting.ErpnextCompanyAbbr,
			PeriodStartDate: time.Now().Unix(),
			OpenPosEntryDetail: []req.OpenPosEntryDetail{{
				ModeOfPayment: "Cash",
				OpeningAmount: previousShiftCash,
			}},
			Branch: companySetting.ErpnextBranchName,
		})
		if err != nil {
			return model.StaffShiftLog{}, errors.WithMessage(err, "开账失败")
		}
		erpnextOpenPosEntryName = openPosEntryName
	}

	shiftLog, _ := shiftLogRepo.Create(model.StaffShiftLog{
		StaffUuid:         staff.Uuid,
		ShiftNo:           s.generateNumber(),
		PreviousShiftCash: previousShiftCash,
		CurrentCashTotal:  previousShiftCash,
		CashLeft:          previousShiftCash,
		ShiftStartTime:    startTime,
		ShiftEndTime:      0,

		ErpnextOpenPosEntryName: erpnextOpenPosEntryName,
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

	// 获取公司设置
	companySetting := ctx.GetCompanySetting()
	settingSrv := setting.NewSrvImpl(s.dbm, s.cache)
	setting := settingSrv.GetDataManageSetting(ctx)

	// 查询钱箱
	cashBox := repository.NewCashBoxRepo(db).Get()

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

	var excludeDataManage, onlyDataManage bool
	if companySetting.IsOpenDataManagement() && setting.IsEnableDataManage {
		excludeDataManage = true
		onlyDataManage = true
	}

	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		DutyNo:            log.ShiftNo,
		ExcludeDataManage: excludeDataManage,
	})
	paymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
		DutyNo:            log.ShiftNo,
		ExcludeDataManage: excludeDataManage,
	})
	refundAmount := s.statisticsSrv.CountShiftRefundAmount(ctx, CountReq{
		DutyNo:            log.ShiftNo,
		ExcludeDataManage: excludeDataManage,
	})

	var paymentMethodIncomeList = make([]resp.PaymentMethodIncome, 0)
	for _, payment := range paymentData.PaymentList {
		paymentMethodIncomeList = append(paymentMethodIncomeList, resp.PaymentMethodIncome{
			Name:   payment.PaymentName,
			Amount: payment.TotalPaymentAmount,
		})
	}
	if saleData.TotalFreeAmount > 0 {
		paymentMethodIncomeList = append(paymentMethodIncomeList, resp.PaymentMethodIncome{
			Name:   i18n.Translate(ctx.GetLanguage(), "免单金额"),
			Amount: saleData.TotalFreeAmount,
		})
	}

	// 数据管理现金收入
	manageCash := decimal.Zero
	if excludeDataManage && onlyDataManage {
		managePaymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
			DutyNo:         log.ShiftNo,
			OnlyDataManage: onlyDataManage,
		})
		for _, payment := range managePaymentData.PaymentList {
			if payment.PaymentCode == constant.PaymentMethodCodeCash {
				manageCash = manageCash.Add(decimal.NewFromFloat(payment.TotalPaymentAmount))
			}
		}
	}

	// 当前钱箱现金总计 = 钱箱余额 - 数据管理现金收入
	boxAmount := decimal.NewFromFloat(cashBox.GetBalance()).Sub(manageCash)

	return &resp.ShiftInfo{
		PreviousShiftCash: log.PreviousShiftCash,
		WithdrawCash:      log.WithdrawCash,
		DepositCash:       log.DepositCash,
		CurrentCashTotal:  boxAmount.InexactFloat64(), // 当前钱箱现金总计（钱箱余额）。 未交班时从钱箱表中获取数据，交班后将当前钱箱现金总计（钱箱余额）记录到交班表中
		RefundAmount:      refundAmount,
		PaymentMethodIncome: resp.PaymentMethodIncomeList{
			List: paymentMethodIncomeList,
		},
	}, nil
}

// GetShiftInfo 获取缓存交班信息
func (s *staffShiftSrv) GetCachedSubmitShift(ctx context.Context) (*resp.ShiftSubmit, error) {
	cacheKey := cryptor.Md5String(ctx.GetGin().GetHeader("Authorization"))
	if ss, exists := s.cache.Get(cacheKey); exists {
		if val, ok := ss.(string); ok {
			var shiftSubmitInfo resp.ShiftSubmit
			if err := json.Unmarshal([]byte(val), &shiftSubmitInfo); err == nil {
				return &shiftSubmitInfo, nil
			}
		}
	}
	return nil, nil
}

// SubmitShift 提交交班
func (s *staffShiftSrv) SubmitShift(ctx context.Context, reqs req.SubmitShiftReq) (*resp.ShiftSubmit, error) {
	// 验证参数
	if reqs.WithdrawCash < 0 {
		return nil, errors.New("取出金额不能小于0")
	}
	if reqs.LeaveCash < 0 {
		return nil, errors.New("遗留现金不能小于0")
	}
	var (
		withdrawCash     decimal.Decimal // 取出现金
		leaveCash        decimal.Decimal // 遗留现金
		cashAmount       float64         // 现金收入
		showWithdrawCash decimal.Decimal // 显示取出现金
	)
	// 获取当前员工
	var (
		staff model.Staff
		db    *gorm.DB = s.dbm.GetDB(ctx.GetCompanyUuid())
	)
	if reqs.IsBackground {
		staff, _ = repository.NewStaffRepo(db).GetStaff(repository.CommonRepo.WhereByUuid(reqs.StaffUuid))
	} else {
		staff = ctx.GetStaff()
	}
	if staff.Uuid == 0 {
		return nil, errors.New("员工不存在")
	}
	// 获取所有支付方式
	paymentRepo := repository.NewPaymentMethodRepo(db)
	paymentMethodList := paymentRepo.GetPaymentMethodList(paymentRepo.WhereExistsErpnextPayment())
	paymentMethodMap := make(map[string]string)
	for _, paymentMethod := range paymentMethodList {
		if paymentMethod.ErpnextPayment != "" {
			paymentMethodMap[paymentMethod.ErpnextPayment] = paymentMethod.PaymentName
		}
	}
	companySetting := ctx.GetCompanySetting()
	settingSrv := setting.NewSrvImpl(s.dbm, s.cache)
	setting := settingSrv.GetDataManageSetting(ctx)
	var excludeDataManage, onlyDataManage bool
	if companySetting.IsOpenDataManagement() && setting.IsEnableDataManage {
		excludeDataManage = true
		onlyDataManage = true
	}
	var currentCashTotal decimal.Decimal
	err := repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 获取当前班次
		shiftLogRepo := repository.NewShiftLogRepo(db)
		shiftLog, err := shiftLogRepo.GetShiftLog(
			repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
			repository.CommonRepo.WhereByShiftNo(staff.DutyNo),
		)
		if err != nil {
			logger.Logger.Error("SubmitShift",
				zap.Any("tips", "当前班次不存在"),
				zap.Any("staff", staff),
				zap.Error(err),
			)
			return errors.New("当前班次不存在")
		}
		if shiftLog.IsHandedOver() {
			return errors.New("当前班次已交班")
		}

		manageCash := decimal.Zero
		if excludeDataManage && onlyDataManage {
			managePaymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
				DutyNo:         shiftLog.ShiftNo,
				OnlyDataManage: onlyDataManage,
			})
			for _, payment := range managePaymentData.PaymentList {
				if payment.PaymentCode == constant.PaymentMethodCodeCash {
					manageCash = manageCash.Add(decimal.NewFromFloat(payment.TotalPaymentAmount))
				}
			}
		}

		cashBox := repository.NewCashBoxRepo(db).Get()
		boxAmount := decimal.NewFromFloat(cashBox.GetBalance())
		showWithdrawCash = decimal.NewFromFloat(reqs.WithdrawCash)
		withdrawCash = decimal.NewFromFloat(reqs.WithdrawCash).Add(manageCash)
		leaveCash = decimal.NewFromFloat(reqs.LeaveCash)
		// 后台交班时，取出金额为当前钱箱现金总计，遗留现金为0
		if reqs.IsBackground {
			withdrawCash = boxAmount
			leaveCash = decimal.Zero
		}
		// 当前班次取出金额 + 遗留现金 = 当前钱箱现金总计
		if !withdrawCash.Add(leaveCash).
			Equal(boxAmount) {
			return errors.New("输入的本班取出現金和本班遗留备用金总额与当前钱箱现金总计不符")
		}
		// 当前班次支付方式收入
		var (
			paymentMethodIncomeList = make([]resp.PaymentMethodIncome, 0)
		)
		saleData := s.statisticsSrv.CountSale(ctx, CountReq{
			DutyNo:            shiftLog.ShiftNo,
			ExcludeDataManage: excludeDataManage,
		})
		paymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
			DutyNo:            shiftLog.ShiftNo,
			ExcludeDataManage: excludeDataManage,
		})
		company := ctx.GetCompany()
		needErpClosePos := company.IsOpenErp() && companySetting.ErpnextSiteCode != "" && shiftLog.ErpnextOpenPosEntryName != ""
		closePosEntryDetail := make([]req.ClosePosEntryDetail, 0)
		for _, payment := range paymentData.PaymentList {
			paymentMethodIncomeList = append(paymentMethodIncomeList, resp.PaymentMethodIncome{
				Name:   payment.PaymentName,
				Amount: payment.TotalPaymentAmount,
			})
			if payment.PaymentCode == constant.PaymentMethodCodeCash {
				cashAmount = decimal.NewFromFloat(cashAmount).Add(decimal.NewFromFloat(payment.TotalPaymentAmount)).InexactFloat64()
			} else {
				if needErpClosePos && decimal.NewFromFloat(payment.TotalPaymentAmount).GreaterThan(decimal.NewFromFloat(999999999999.99)) {
					return errors.WithMessage(errors.New(paymentMethodMap[payment.ErpnextPayment]+i18n.Translate(ctx.GetLanguage(), "最大为999999999999.99")), "交班失败")
				}
				closePosEntryDetail = append(closePosEntryDetail, req.ClosePosEntryDetail{
					ModeOfPayment: payment.ErpnextPayment,
					OpeningAmount: 0,
					ClosingAmount: payment.TotalPaymentAmount,
				})
			}
		}
		cashAmountDecimal := decimal.NewFromFloat(cashAmount).Add(decimal.NewFromFloat(shiftLog.PreviousShiftCash))
		if needErpClosePos && cashAmountDecimal.GreaterThan(decimal.NewFromFloat(999999999999.99)) {
			return errors.WithMessage(errors.New(paymentMethodMap["Cash"]+i18n.Translate(ctx.GetLanguage(), "最大为999999999999.99")), "交班失败")
		}
		closePosEntryDetail = append(closePosEntryDetail, req.ClosePosEntryDetail{
			ModeOfPayment: "Cash",
			OpeningAmount: shiftLog.PreviousShiftCash,
			ClosingAmount: cashAmountDecimal.InexactFloat64(),
		})
		if saleData.TotalFreeAmount > 0 {
			closePosEntryDetail = append(closePosEntryDetail, req.ClosePosEntryDetail{
				ModeOfPayment: "Free Meal",
				OpeningAmount: 0,
				ClosingAmount: saleData.TotalFreeAmount,
			})
		}
		incomes, _ := convertor.ToJson(paymentMethodIncomeList)
		// 更新当班记录
		currentCashTotal = leaveCash
		var erpnextAsyncRecordId string
		if needErpClosePos {
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpnextAsyncRecordId, err = erpSrv.ClosePosEntry(ctx.GetContext(), req.ClosePosEntryReq{
				SiteCode:            companySetting.ErpnextSiteCode,
				PosProfileName:      companySetting.ErpnextPosProfileName,
				PosOpenEntryName:    shiftLog.ErpnextOpenPosEntryName,
				PeriodEndDate:       time.Now().Unix(),
				ClosePosEntryDetail: closePosEntryDetail,
				InvoiceCount:        s.CountInvoiceCount(ctx, shiftLog.Uuid),
			})
			if err != nil {
				ctx.Log().Info("ClosePosEntry", zap.String("msg", err.Error()))
				// 1、交班时物品被禁用
				msg := utils.ShortenErpnextError(err.Error())
				if isItemDisabled, itemCode := utils.IsItemDisabledError(msg); isItemDisabled {
					material, err := repository.NewMaterialRepo(db).GetMaterialByErpCode(itemCode)
					if err != nil {
						ctx.Log().Info("GetMaterialByErpCode", zap.String("msg", err.Error()), zap.String("erp_msg", msg))
						return errors.WithMessage(err)
					}
					// 物品被禁用，无法交班
					tips := i18n.Translate(ctx.GetLanguage(), "物品被禁用，无法交班")
					name := material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					return errors.WithMessage(errors.New(fmt.Sprintf("%s: %s", tips, name)), fmt.Sprintf("物品%s已禁用", name))
				}
				return errors.WithMessage(errors.New("交班失败，请重试"), "交班失败")
			}
		}
		shiftEndTime := time.Now().Unix()
		_, err = shiftLogRepo.Update(shiftLog, map[string]any{
			"status":                  constant.StaffHandedOver,          // 交班状态
			"current_cash_total":      currentCashTotal.InexactFloat64(), // 当前钱箱现金总计
			"cash_taken_out":          showWithdrawCash.InexactFloat64(), // 取出现金
			"cash_left":               currentCashTotal.InexactFloat64(), // 遗留现金
			"shift_end_time":          shiftEndTime,                      // 交班时间
			"cash_income":             cashAmount,                        // 现金收入
			"incomes":                 incomes,                           // 支付方式收入
			"total_business":          saleData.TotalSaleAmount,          // 营业额
			"total_income":            saleData.TotalReceivedAmount,      // 总收入
			"erpnext_async_record_id": erpnextAsyncRecordId,              // erpnext异步记录ID
		})
		if err != nil {
			return errors.New("交班失败")
		}
		// 更新钱箱记录
		if withdrawCash.GreaterThan(decimal.Zero) {
			err := s.cashBoxSrv.UpdateBalance(
				ctx,
				UpdateCashBalanceParam{
					Amount: -withdrawCash.InexactFloat64(),
					Scene:  constant.CashBoxLogSceneShift,
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

		utils.Go(func() {
			// 创建交班快照
			err = s.CreateShiftSnapshot(ctx.Copy(), shiftLog)
			if err != nil {
				ctx.Log().Error("交班-创建交班快照失败", zap.Error(err))
			}
		})

		//
		utils.Go(func() {
			websocket.PushClient(staff.CompanyUuid, websocket.SourceAssistant, websocket.SourceAll, websocket.UPDATE_CONFIG, map[string]any{
				"update_time": time.Now().Unix(),
			})
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	shiftSubmitInfo := resp.ShiftSubmit{
		CashIncome:   cashAmount,
		CashTakenOut: showWithdrawCash.InexactFloat64(),
		CashLeft:     currentCashTotal.InexactFloat64(),
		DutyNo:       staff.DutyNo,
		PrinterData: func() resp.PrinterData {
			printerData, err := s.ShiftPrinter(ctx, req.ShiftPrinterReq{
				WithdrawCash: showWithdrawCash.InexactFloat64(),
				LeaveCash:    currentCashTotal.InexactFloat64(),
				DutyNo:       staff.DutyNo,
			}, utils.IfInt(reqs.IsBackground, 0, 1), true, staff)
			if err != nil {
				fmt.Println(err)
				return resp.PrinterData{}
			}
			return *printerData
		}(),
	}
	cacheKey := cryptor.Md5String(ctx.GetGin().GetHeader("Authorization"))
	// 缓存交班数据12小时
	if err = s.cache.Set(cacheKey, shiftSubmitInfo, 12*time.Hour); err != nil {
		logger.Logger.Error("cache shiftSubmitInfo error", zap.Error(err))
	}
	return &shiftSubmitInfo, nil
}

// 统计当班期间发票数量。包括退款单数，不包含取消的单数
func (s *staffShiftSrv) CountInvoiceCount(ctx context.Context, shiftLogUuid uint64) int64 {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 查询当前班次中，所有完成结账的sale_order数量
	saleOrderCount := int64(0)
	db.Model(&model.SaleOrder{}).Where("staff_shift_log_uuid = ? and delete_time = 0 and status = ?", shiftLogUuid, constant.SaleOrderStatusFinish).Count(&saleOrderCount)
	// 查询当前班次中，所有return_order数量
	returnOrderCount := int64(0)
	db.Model(&model.ReturnOrder{}).Where("staff_shift_log_uuid = ? and delete_time = 0", shiftLogUuid).Count(&returnOrderCount)
	return returnOrderCount + saleOrderCount
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
	_, err = shiftLogRepo.Update(log, map[string]any{
		"exception_remark": req.ExceptionRemark,
	})
	if err != nil {
		return errors.New("更新交班记录失败")
	}

	return nil
}

// ShiftWithdraw 交班取钱
func (s *staffShiftSrv) ShiftWithdraw(ctx context.Context, req req.ShiftWithdrawReq) error {
	withdrawCash, err := decimal.NewFromString(req.WithdrawCash)
	if err != nil {
		return errors.New("请输入正确金额")
	}
	if withdrawCash.LessThanOrEqual(decimal.Zero) {
		return errors.New("请输入正确金额")
	}
	if withdrawCash.LessThanOrEqual(decimal.Zero) {
		return errors.New("请输入正确金额")
	}
	// 验证参数, 大于0, 最多小数点后两位
	reg := regexp.MustCompile(`^([1-9]\d*|0)(\.\d{1,2})?$`)
	if !reg.MatchString(withdrawCash.String()) {
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

	maxAmount, err := decimal.NewFromString(s.MaxAmount)
	if err != nil {
		return errors.New("最大金额格式错误")
	}
	if decimal.NewFromFloat(log.WithdrawCash).Add(withdrawCash).GreaterThanOrEqual(maxAmount) {
		return errors.NewWithReplace("当前班次取钱累计金额不能大于最大金额%s", []string{s.MaxAmount})
	}

	err = repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 更新钱箱记录
		err = s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
			Amount:         -withdrawCash.InexactFloat64(),
			CashWithdrawal: withdrawCash.InexactFloat64(),
			Scene:          constant.CashBoxLogSceneOut,
		})
		if err != nil {
			return errors.New("取钱失败")
		}

		// 更新交班记录
		balance, err := s.cashBoxSrv.GetBalance(ctx)
		if err != nil {
			return errors.New("获取钱箱余额失败")
		}
		_, err = shiftLogRepo.Update(log, map[string]any{
			"withdraw_cash": gorm.Expr("withdraw_cash + ?", withdrawCash.InexactFloat64()),
			"cash_left":     balance,
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
	depositCash, err := decimal.NewFromString(req.DepositCash)
	if err != nil {
		return errors.New("请输入正确金额")
	}
	if depositCash.LessThanOrEqual(decimal.Zero) {
		return errors.New("请输入正确金额")
	}
	reg := regexp.MustCompile(`^([1-9]\d*|0)(\.\d{1,2})?$`)
	if !reg.MatchString(depositCash.String()) {
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

	maxAmount, err := decimal.NewFromString(s.MaxAmount)
	if err != nil {
		return errors.New("最大金额格式错误")
	}

	// 当前班次存钱累计金额不能大于最大金额
	if decimal.NewFromFloat(log.DepositCash).Add(depositCash).GreaterThanOrEqual(maxAmount) {
		return errors.NewWithReplace("当前班次存钱累计金额不能大于最大金额%s", []string{s.MaxAmount})
	}

	// 钱箱余额不能大于最大金额
	balance, err := s.cashBoxSrv.GetBalance(ctx)
	if err != nil {
		return errors.New("获取钱箱余额失败")
	}
	if decimal.NewFromFloat(balance).Add(depositCash).GreaterThanOrEqual(maxAmount) {
		return errors.NewWithReplace("钱箱余额不能大于最大金额%s", []string{s.MaxAmount})
	}

	err = repository.NewCommonRepo().Transaction(db, func(tx *gorm.DB) error {
		// 更新交班记录
		_, err = shiftLogRepo.Update(log, map[string]any{
			"deposit_cash": gorm.Expr("deposit_cash + ?", depositCash.InexactFloat64()),
			"cash_left":    gorm.Expr("cash_left + ?", depositCash.InexactFloat64()),
		})
		if err != nil {
			return errors.New("存钱失败")
		}
		// 更新钱箱记录
		err = s.cashBoxSrv.UpdateBalance(ctx, UpdateCashBalanceParam{
			Amount:      depositCash.InexactFloat64(),
			CashDeposit: depositCash.InexactFloat64(),
			Scene:       constant.CashBoxLogSceneIn,
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
func (s *staffShiftSrv) ShiftPrinter(ctx context.Context, req req.ShiftPrinterReq, firstExecution int, openMoneybox bool, staffs ...model.Staff) (*resp.PrinterData, error) {
	staff := ctx.GetStaff()
	if len(staffs) > 0 {
		staff = staffs[0]
	}

	setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	storeSetting, err := setting.GetStoreSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店设置失败", zap.Error(err))
		return nil, errors.WithMessage(err)
	}

	// 获取打印机设置
	printerSetting, err := setting.GetPrinterSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取打印机设置失败", zap.Error(err))
		fmt.Println("获取打印机设置失败", zap.Error(err))
	}
	// 设置打印语言
	ctx.SetLanguage(printerSetting.DefaultLanguage)

	//
	shiftLogRepo := repository.NewShiftLogRepo(ctx.GetDB())

	// 查询交班记录
	log, err := shiftLogRepo.GetShiftLog(
		shiftLogRepo.WithStaff(),
		repository.CommonRepo.WhereByStaffUuid(staff.Uuid),
		repository.CommonRepo.WhereByShiftNo(utils.IfString(req.DutyNo != "", req.DutyNo, staff.DutyNo)),
	)
	if err != nil {
		return nil, errors.New("当前班次错误，请退出重新登录")
	}

	// 查询钱箱
	cashBox, err := s.cashBoxSrv.GetRecord(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取钱箱记录失败")
	}

	log.CashLeft = cashBox.GetBalance()

	companySetting := ctx.GetCompanySetting()
	dataSetting := setting.GetDataManageSetting(ctx)
	excludeDataManage := companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage

	// 销售数据
	saleData := s.statisticsSrv.CountSale(ctx, CountReq{
		DutyNo:            log.ShiftNo,
		ExcludeDataManage: excludeDataManage,
	})

	// 交班退款
	refundAmount := s.statisticsSrv.CountShiftRefundAmount(ctx, CountReq{
		DutyNo:            log.ShiftNo,
		ExcludeDataManage: excludeDataManage,
	})

	// 会员数量, 根据当班记录的开始和结束时间
	queryEndTime := time.Now().Unix()
	if log.ShiftEndTime > 0 {
		queryEndTime = log.ShiftEndTime
	}
	memberNum := s.statisticsSrv.CountMemberNum(ctx, CountReq{
		QueryStartTime: log.ShiftStartTime,
		QueryEndTime:   queryEndTime,
	})

	// 营业数据
	var businessData = business_data_resp.BusinessDataAll{
		TotalSales:             saleData.TotalSaleAmount,
		TotalReceivedPrice:     saleData.TotalReceivedAmount,
		TotalProductPrice:      saleData.TotalProductOriginPrice,
		TotalPayPrice:          saleData.TotalSaleAmount,
		TotalPayFeeMoney:       saleData.TotalPaymentFee,
		TotalServiceMoney:      saleData.TotalServiceFee,
		TotalTaxMoney:          saleData.TotalTax,
		TotalUserDiscountMoney: saleData.TotalDiscountMember,
		TotalDiscountMoney:     saleData.TotalDiscount,
		TotalFreeOrderPrice:    saleData.TotalFreeAmount,
		TotalRefundMoney:       refundAmount,
		TotalGiveProductPrice:  saleData.TotalGiftAmount,
		TotalGiveProductNum:    saleData.TotalGiftNum,
		TotalOrderNum:          int(saleData.TotalOrderNum),
		TotalPeopleNum:         int(saleData.TotalMealNum),
		TotalProductNum:        saleData.TotalProductNum,
		TotalTableNum:          int(saleData.TotalDeskNum),
		TotalCancelOrderNum:    int(saleData.TotalCancelOrderNum),
		TotalCancelOrderAmount: saleData.TotalCancelOrderAmount,
		AvgOrderPrice:          saleData.AvgOrderAmount,
		MinOrderPrice:          saleData.MinOrderAmount,
		MaxOrderPrice:          saleData.MaxOrderAmount,
		// 桌台方式
		AllTableOrderNum:      int(saleData.TotalDeskNum),
		AllTablePeopleNum:     int(saleData.TotalMealNum),
		AllTableAvgOrderPrice: saleData.AvgDeskOrderAmount,
		AllTableMinOrderPrice: saleData.MinDeskOrderAmount,
		AllTableMaxOrderPrice: saleData.MaxDeskOrderAmount,
		AllTablePeopleAvg:     saleData.AvgDeskPeopleOrderAmount,
		// 收银方式-店内
		AllCashierOrderNum:      int(saleData.TotalInstantOrderNum),
		AllCashierMinOrderPrice: saleData.MinInstantOrderAmount,
		AllCashierMaxOrderPrice: saleData.MaxInstantOrderAmount,
		AllCashierAvgOrderPrice: saleData.AvgInstantOrderAmount,
		// 收银方式-外卖
		AllTakeawayOrderNum:      int(saleData.TotalInstantOrderTakeawayNum),
		AllTakeawayMinOrderPrice: saleData.MinInstantOrderTakeawayAmount,
		AllTakeawayMaxOrderPrice: saleData.MaxInstantOrderTakeawayAmount,
		AllTakeawayAvgOrderPrice: saleData.AvgInstantOrderTakeawayAmount,
		// 支付方式
		PaymentMethodIncomes: func() []business_data_resp.PaymentMethodIncome {
			// 支付数据
			paymentData := s.statisticsSrv.CountPayment(ctx, CountReq{
				DutyNo:            log.ShiftNo,
				ExcludeDataManage: excludeDataManage,
			})
			paymentMethodIncomes := make([]business_data_resp.PaymentMethodIncome, 0, len(paymentData.PaymentList))
			for _, v := range paymentData.PaymentList {
				paymentMethodIncomes = append(paymentMethodIncomes, business_data_resp.PaymentMethodIncome{
					Name:     v.PaymentName,
					Amount:   v.TotalPaymentAmount,
					OrderNum: int(v.TotalOrderNum),
					Code:     v.PaymentCode,
				})
			}
			return paymentMethodIncomes
		}(),
		AbnormalData: func() business_data_resp.AbnormalData {
			AbnormalData, err := repository.NewOrderAbnormalRecordRepo(ctx.GetDB()).GetRecordInfo(
				0,
				ctx.GetStaff().DutyNo,
			)
			if err != nil {
				return business_data_resp.AbnormalData{}
			}
			return *AbnormalData
		}(),
		MemberData: func() business_data_resp.MemberData {
			memberData := s.statisticsSrv.CountMember(ctx, CountReq{
				DutyNo: log.ShiftNo,
			})
			return business_data_resp.MemberData{
				RechargeAmount: memberData.TotalRechargeAmount,
				GiftMoney:      memberData.TotalGiveAmount,
				GiftPoints:     memberData.TotalGivePoint,
				UserCount:      int(memberNum),
			}
		}(),
		PeakHourList: func() []business_data_resp.PeakHour {
			peakHours, err := repository.NewSaleOrderPeakTimeRepo(ctx.GetDB()).GetMaxRecord(
				storeSetting.TimeZone,
				uint(log.ShiftStartTime),
				uint(queryEndTime),
				log.StaffUuid,
				excludeDataManage,
			)
			if err != nil {
				return []business_data_resp.PeakHour{}
			}
			return peakHours
		}(),
		CategoryList: func() []business_data_resp.Category {
			// 统计分类
			categoryData := s.statisticsSrv.CountCategory(ctx, CountReq{
				DutyNo:            log.ShiftNo,
				ExcludeDataManage: excludeDataManage,
			})

			categoryList := make([]business_data_resp.Category, 0, len(categoryData.CategoryList))
			for _, v := range categoryData.CategoryList {
				categoryList = append(categoryList, business_data_resp.Category{
					Name:     v.CategoryName,
					SalesNum: v.SaleNum,
					Prices:   v.SaleAmount,
				})
			}
			return categoryList
		}(),
		PercentageList: func() []business_data_resp.Percentage {
			taxData := s.statisticsSrv.CountTax(ctx, CountReq{
				DutyNo:            log.ShiftNo,
				ExcludeDataManage: excludeDataManage,
			})

			percentageList := make([]business_data_resp.Percentage, 0, len(taxData))
			for _, v := range taxData {
				percentageList = append(percentageList, business_data_resp.Percentage{
					TaxRate:        v.TaxRate,
					ConsumptionTax: v.TotalTaxFee,
					TotalPrice:     v.TotalProductAmount,
				})
			}
			return percentageList
		}(),
	}

	// 打印
	printerData, err := printer.NewPrinterRepo(ctx).PrintingHandoverOrder(
		&log,
		&businessData,
		firstExecution,
		openMoneybox,
		utils.IfString(firstExecution == 0, staff.BindKey, ctx.GetDeviceSn()),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerData, nil
}

// CreateShiftSnapshot 创建交班快照
func (s *staffShiftSrv) CreateShiftSnapshot(ctx context.Context, shiftLog model.StaffShiftLog) error {
	// 获取最新交班数据
	shiftLogRepo := repository.NewShiftLogRepo(ctx.GetDB())
	log, err := shiftLogRepo.GetShiftLog(
		shiftLogRepo.WithStaff(),
		repository.NewCommonRepo().WhereByUuid(shiftLog.Uuid),
	)
	if err != nil {
		return errors.New("获取交班数据失败")
	}

	// 获取交班快照
	snapshot, _ := shiftLogRepo.GetSnapshot(
		repository.NewCommonRepo().WhereByShiftLogUuid(shiftLog.Uuid),
	)
	if snapshot.Uuid > 0 {
		return nil
	}

	companySetting := ctx.GetCompanySetting()
	settingSrv := setting.NewSrvImpl(s.dbm, s.cache)
	dataSetting := settingSrv.GetDataManageSetting(ctx)
	excludeDataManage := companySetting.IsOpenDataManagement() && dataSetting.IsEnableDataManage

	// 获取交班详情
	// uploadFileSrv := service.NewUploadFileSrv(s.dbm)
	businessSrv := NewBusinessSrv(s.statisticsSrv, nil)
	businessData, err := businessSrv.CountBusiness(ctx, req.BusinessDataCountReq{
		QueryStartTime:    log.ShiftStartTime,
		QueryEndTime:      log.ShiftEndTime,
		ExcludeDataManage: excludeDataManage,
	})
	if err != nil {
		return errors.New("获取交班数据失败")
	}

	time := utils.SetTimezone(ctx.GetCompany().CompanySetting.Timezone)

	// 高峰期列表
	peakHourList := make([]model.StaffShiftSnapshotPeakHour, 0, len(businessData.PeakHourList))
	for _, v := range businessData.PeakHourList {
		peakHourList = append(peakHourList, model.StaffShiftSnapshotPeakHour{
			TimePeriod: v.TimePeriod,
			Num:        v.OrderNum,
			Amount:     v.Amount,
		})
	}

	// 税率列表
	percentageList := make([]model.StaffShiftSnapshotPercentage, 0, len(businessData.PercentageList))
	for _, v := range businessData.PercentageList {
		percentageList = append(percentageList, model.StaffShiftSnapshotPercentage{
			TotalPrice:     v.TotalPrice,
			TaxRate:        v.TaxRate,
			ConsumptionTax: v.ConsumptionTax,
		})
	}

	// 支付统计
	incomes := make([]model.StaffShiftSnapshotIncome, 0, len(businessData.PaymentMethodIncomes))
	for _, v := range businessData.PaymentMethodIncomes {
		incomes = append(incomes, model.StaffShiftSnapshotIncome{
			PayType:     v.Code,
			PayTypeName: v.Name,
			Price:       v.Amount,
			OrderNum:    v.OrderNum,
		})
	}

	// 销售统计
	salesInfo := make([]model.StaffShiftSnapshotSalesInfo, 0, len(businessData.CategoryList))
	for _, v := range businessData.CategoryList {
		salesInfo = append(salesInfo, model.StaffShiftSnapshotSalesInfo{
			Name:     v.Name,
			Sales:    convertor.ToString(v.SalesNum),
			Prices:   fmt.Sprintf("%.2f", v.Prices),
			NameText: v.Name,
		})
	}

	// 统计退款金额
	refundAmount := s.statisticsSrv.CountShiftRefundAmount(ctx, CountReq{DutyNo: log.ShiftNo, ExcludeDataManage: excludeDataManage})

	userIsDelete := 0
	if log.Staff.IsDelete() {
		userIsDelete = 1
	}
	userIsStatus := 1
	if log.Staff.IsDisable == 1 {
		userIsStatus = 0
	}

	content := model.StaffShiftSnapshotContent{
		ID:                log.Uuid,
		ShiftUserID:       log.StaffUuid,
		ShiftNo:           log.ShiftNo,
		Status:            log.Status,
		PreviousShiftCash: fmt.Sprintf("%.2f", log.PreviousShiftCash),
		CurrentCashTotal:  fmt.Sprintf("%.2f", log.CurrentCashTotal),
		Incomes:           incomes,
		TotalIncome:       fmt.Sprintf("%.2f", log.TotalIncome),
		CashTakenOut:      fmt.Sprintf("%.2f", log.CashTakenOut),
		CashLeft:          fmt.Sprintf("%.2f", log.CashLeft),
		CashIncome:        fmt.Sprintf("%.2f", log.CashIncome),
		TotalBusiness:     fmt.Sprintf("%.2f", log.TotalBusiness),
		IsPrinted:         log.IsPrinted,
		Remark:            log.Remark,
		WithdrawCash:      fmt.Sprintf("%.2f", log.WithdrawCash),
		DepositCash:       fmt.Sprintf("%.2f", log.DepositCash),
		ExceptionRemark:   log.ExceptionRemark,
		AppID:             log.Staff.CompanyUuid,
		ShopSupplierID:    log.Staff.CompanyUuid,
		ShiftStartTime:    time.FormatUnixTimeDefault(log.ShiftStartTime),
		ShiftEndTime:      time.FormatUnixTimeDefault(log.ShiftEndTime),
		CreateTime:        time.FormatUnixTimeDefault(log.CreateTime),
		UpdateTime:        time.FormatUnixTimeDefault(log.UpdateTime),
		Abnormal: model.StaffShiftSnapshotAbnormal{
			RefundProductTimes:    businessData.AbnormalData.RefundProductTimes,
			ProductFreeTimes:      businessData.AbnormalData.ProductFreeTimes,
			RefundTime:            businessData.AbnormalData.RefundTimes,
			ProductMoveTimes:      businessData.AbnormalData.ProductMoveTimes,
			ChangePriceTimes:      businessData.AbnormalData.ChangePriceTimes,
			ChangeOrderPriceTimes: businessData.AbnormalData.ChangeOrderPriceTimes,
			DiscountOrderTimes:    businessData.AbnormalData.DiscountOrderTimes,
			RoundOrderTimes:       businessData.AbnormalData.RoundOrderTimes,
			FreeOrderTimes:        businessData.AbnormalData.FreeOrderTimes,
			ReverseSettleTimes:    businessData.AbnormalData.ReverseSettleTimes,
		},
		Order: model.StaffShiftSnapshotOrder{
			ServiceMoney:           businessData.TotalServiceMoney,
			DiscountMoney:          businessData.TotalDiscountMoney,
			ConsumptionTaxMoney:    businessData.TotalTaxMoney,
			PayFeeMoney:            businessData.TotalPayFeeMoney,
			ReceivedPrice:          businessData.TotalReceivedPrice,
			TotalGiveProductPrice:  businessData.TotalGiveProductPrice,
			ProductNum:             businessData.TotalProductNum,
			UserDiscountMoney:      businessData.TotalUserDiscountMoney,
			RefundMoney:            refundAmount,
			TotalOrderNum:          businessData.TotalOrderNum,
			TotalTableNum:          businessData.TotalTableNum,
			TotalCancelOrderNum:    businessData.TotalCancelOrderNum,
			TotalCancelOrderAmount: businessData.TotalCancelOrderAmount,
			TotalPeopleNum:         businessData.TotalPeopleNum,
			MinOrderPrice:          businessData.MinOrderPrice,
			MaxOrderPrice:          businessData.MaxOrderPrice,
			AvgOrderPrice:          businessData.AvgOrderPrice,
			// 桌台方式
			TableOrderNum:      businessData.AllTableOrderNum,
			TablePeopleNum:     businessData.AllTablePeopleNum,
			TableMinOrderPrice: businessData.AllTableMinOrderPrice,
			TableMaxOrderPrice: businessData.AllTableMaxOrderPrice,
			TableAvgOrderPrice: businessData.AllTableAvgOrderPrice,
			TablePeopleAvg:     businessData.AllTablePeopleAvg,
			// 收银方式-店内
			CashierOrderNum:      businessData.AllCashierOrderNum,
			CashierMinOrderPrice: businessData.AllCashierMinOrderPrice,
			CashierMaxOrderPrice: businessData.AllCashierMaxOrderPrice,
			CashierAvgOrderPrice: businessData.AllCashierAvgOrderPrice,
			// 收银方式-外卖
			TakeawayOrderNum:      businessData.AllTakeawayOrderNum,
			TakeawayMinOrderPrice: businessData.AllTakeawayMinOrderPrice,
			TakeawayMaxOrderPrice: businessData.AllTakeawayMaxOrderPrice,
			TakeawayAvgOrderPrice: businessData.AllTakeawayAvgOrderPrice,
			// 会员数据
			GiftPoints:     businessData.MemberData.GiftPoints,
			GiftMoney:      businessData.MemberData.GiftMoney,
			RechargeAmount: businessData.MemberData.RechargeAmount,
			PeakHourList:   peakHourList,
			PercentageList: percentageList,
			Incomes:        incomes,
		},
		SalesInfo: salesInfo,
		User: model.StaffShiftSnapshotUser{
			ShopUserID:       log.Staff.Uuid,
			UserName:         log.Staff.Username,
			Password:         log.Staff.Password,
			Phone:            log.Staff.Phone,
			PasswordChange:   log.Staff.PasswordChangeCount,
			RealName:         log.Staff.RealName,
			IsSuper:          log.Staff.IsSuper,
			ShopSupplierID:   log.Staff.CompanyUuid,
			IsDelete:         userIsDelete,
			UserType:         log.Staff.UserType,
			IsStatus:         userIsStatus,
			AppID:            log.Staff.CompanyUuid,
			BindKey:          log.Staff.BindKey,
			CashierOnline:    log.Staff.CashierOnline,
			CashierLoginTime: log.Staff.CashierLoginTime,
			DutyNo:           log.Staff.DutyNo,
			CreateTime:       time.FormatUnixTimeDefault(log.Staff.CreateTime),
			UpdateTime:       time.FormatUnixTimeDefault(log.Staff.UpdateTime),
		},
	}

	jsonData, err := json.Marshal(content)
	if err != nil {
		return errors.New("序列化交班快照数据失败")
	}

	_, err = shiftLogRepo.CreateSnapshot(model.StaffShiftSnapshot{
		ShiftLogUuid: log.Uuid,
		Content:      string(jsonData),
	})
	if err != nil {
		return errors.New("创建交班快照失败")
	}

	return nil
}

package selling

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/consts"
	"ttpos-bmp/app/ttpos-erp/internal/logic/erpnext"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"
	"ttpos-bmp/app/ttpos-erp/internal/service"
	"ttpos-bmp/utility/uuid"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// 常量定义
const (
	// 默认值
	DefaultPaymentAllowInReturns = 1
	DefaultCashPaymentDefault    = 1
	DefaultTaxChargeType         = "Actual"
	DefaultTaxAccountHeadPrefix  = "Cash - "

	// 时间格式
	DateFormat = "Y-m-d"
	TimeFormat = "H:i:s"
)

var (
	Selling = new(sSelling)
)

// sSelling 销售服务结构体
type sSelling struct{}

func init() {
	service.RegisterSelling(Selling)
}

// GetPosProfileList 查询POS配置文件列表
// 参数：
//   - ctx: 上下文对象
//   - req: 查询请求参数
//
// 返回：
//   - *selling.PosProfileListResp: POS配置文件列表响应
//   - error: 错误信息
func (s *sSelling) GetPosProfileList(ctx context.Context, req *selling.PosProfileReq) (res *selling.PosProfileListResp, err error) {
	// 构建过滤条件
	filters := s.buildPosProfileFilters(ctx, req)

	// 查询POS配置文件列表
	list, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosProfile,
	}, &erp.RequestParams{
		Fields:  []string{"name", "company", "warehouse", "branch"},
		Filters: filters,
	})
	if err != nil {
		g.Log().Error(ctx, "查询POS配置文件失败", "filters", filters, "error", err)
		return nil, gerror.Wrapf(err, "查询POS配置文件失败，过滤条件: %v", filters)
	}

	// 解析响应数据
	res, err = s.parsePosProfileListResponse(list)
	if err != nil {
		g.Log().Error(ctx, "解析POS配置文件列表响应失败", "error", err)
		return nil, gerror.Wrapf(err, "解析POS配置文件列表响应失败")
	}

	return res, err
}

// buildPosProfileFilters 构建POS配置文件查询过滤条件
// 参数：
//   - ctx: 上下文对象
//   - req: 查询请求参数
//
// 返回：
//   - [][]string: 过滤条件列表
func (s *sSelling) buildPosProfileFilters(ctx context.Context, req *selling.PosProfileReq) [][]string {
	var filters = make([][]string, 0)

	if len(req.Name) > 0 {
		filters = append(filters, []string{"name", "like", req.Name})
	}
	if len(req.Company) > 0 {
		filters = append(filters, []string{"company", "=", req.Company})
	}
	if len(req.CompanyAbbr) > 0 {
		company, err := service.Company().GetCompanyWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			g.Log().Error(ctx, "根据公司缩写查询公司失败", "company_abbr", req.CompanyAbbr, "error", err)
			return filters
		}
		if company.CompanyName != "" {
			filters = append(filters, []string{"company", "=", company.CompanyName})
		}
	}

	return filters
}

// parsePosProfileListResponse 解析POS配置文件列表响应
// 参数：
//   - list: 响应数据
//
// 返回：
//   - *selling.PosProfileListResp: 解析后的响应数据
//   - error: 错误信息
func (s *sSelling) parsePosProfileListResponse(list *gjson.Json) (*selling.PosProfileListResp, error) {
	j := list

	// 遍历响应数据，构建结果列表
	dataList := make([]*selling.PosProfile, 0)
	dataArray := j.GetJsons("data")
	for _, item := range dataArray {
		dataInfo := &selling.PosProfile{
			Name:      item.Get("name").String(),
			Company:   item.Get("company").String(),
			Branch:    item.Get("branch").String(),
			Warehouse: item.Get("warehouse").String(),
		}
		dataList = append(dataList, dataInfo)
	}

	return &selling.PosProfileListResp{
		ProfileList: dataList,
	}, nil
}

// CreateModePaymentAccount 创建支付方式账户
// 参数：
//   - ctx: 上下文对象
//   - req: 创建支付方式账户请求参数
//
// 返回：
//   - error: 错误信息
func (s *sSelling) CreateModePaymentAccount(ctx context.Context, req *setup.CreateModePaymentAccountInp) error {

	var modePayment *erp.ModeOfPayment
	// 检查支付方式是否已存在
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
	}, &erp.RequestParams{
		Filters: [][]string{
			{"name", "=", req.PaymentType},
		},
	})
	if err != nil {
		return gerror.Wrapf(err, "查询支付方式失败")
	}
	if count == 0 {
		modePayment = &erp.ModeOfPayment{
			ModeOfPayment: req.PaymentType,
			Accounts:      make([]erp.ModeOfPaymentAccount, 0),
			Type:          "Cash",
		}
	} else {
		// 获取已有支付方式信息
		modePayment, err = s.getModeOfPayment(ctx, req.PaymentType)
		if err != nil {
			return err
		}
	}

	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	// 创建默认支付账户
	accountExists := false
	for _, account := range modePayment.Accounts {
		if account.Company == companyName {
			accountExists = true
			break
		}
	}
	if !accountExists {
		err = s.createDefaultPaymentAccount(ctx, modePayment, companyName, req.CompanyAbbr)
		if err != nil {
			return gerror.Wrapf(err, "创建默认支付账户失败")
		}
	}

	return nil
}

// getModeOfPayment 获取支付方式信息
// 参数：
//   - ctx: 上下文对象
//   - paymentType: 支付类型
//
// 返回：
//   - *erp.ModeOfPayment: 支付方式信息
//   - error: 错误信息
func (s *sSelling) getModeOfPayment(ctx context.Context, paymentType string) (*erp.ModeOfPayment, error) {
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
		Name:    paymentType,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取支付方式失败")
	}

	j, err := gjson.DecodeToJson(resp)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析支付方式失败")
	}

	modePayment := &erp.ModeOfPayment{}
	j.GetJson("data").Scan(&modePayment)
	return modePayment, nil
}

// createDefaultPaymentAccount 创建默认支付账户
// 参数：
//   - ctx: 上下文对象
//   - modePayment: 支付方式信息
//   - companyName: 公司名称
//   - companyAbbr: 公司缩写
//
// 返回：
//   - error: 错误信息
func (s *sSelling) createDefaultPaymentAccount(ctx context.Context, modePayment *erp.ModeOfPayment, companyName, companyAbbr string) error {
	payAccounts := make([]erp.ModeOfPaymentAccount, 0)
	payAccounts = append(payAccounts, modePayment.Accounts...)
	payAccounts = append(payAccounts, erp.ModeOfPaymentAccount{
		Company:        companyName,
		DefaultAccount: "Cash - " + companyAbbr,
	})
	modePayment.Accounts = payAccounts

	if len(modePayment.Creation) > 0 {
		_, err := service.Document().Update(ctx, &erp.ErpReq{
			DocType: erp.DocTypeModeOfPayment,
			Name:    modePayment.Name,
		}, modePayment)
		if err != nil {
			return gerror.Wrapf(err, "更新支付方式失败")
		}
	} else {
		_, err := service.Document().Create(ctx, erp.DocTypeModeOfPayment, modePayment)
		if err != nil {
			return gerror.Wrapf(err, "创建支付方式失败")
		}
	}

	return nil
}

// CreatePosProfile 创建默认的POS配置文件
// 参数：
//   - ctx: 上下文对象
//   - req: 创建POS配置文件请求参数
//
// 返回：
//   - *erp.POSProfile: POS配置文件信息
//   - error: 错误信息
func (s *sSelling) CreatePosProfile(ctx context.Context, req *setup.CreatePosProfileInp) (*erp.POSProfile, error) {
	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	// 构建POS配置文件
	profile := s.buildPosProfile(req, companyName)

	// 创建POS配置文件
	resp, err := service.Document().Create(ctx, erp.DocTypePosProfile, profile)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建POS配置文件失败")
	}

	// 解析响应数据
	posProfile, err := s.parsePosProfileResponse(resp)
	if err != nil {
		return nil, err
	}

	return posProfile, nil
}

// buildPosProfile 构建POS配置文件
// 参数：
//   - req: 创建POS配置文件请求参数
//   - companyName: 公司名称
//
// 返回：
//   - *erp.POSProfile: POS配置文件
func (s *sSelling) buildPosProfile(req *setup.CreatePosProfileInp, companyName string) *erp.POSProfile {
	profile := &erp.POSProfile{
		Name:                req.PosProfileName,
		Company:             companyName,
		Warehouse:           req.Warehouse,
		Branch:              req.Branch,
		Currency:            req.Currency,
		WriteOffAccount:     req.WriteOffAccount,
		WriteOffLimit:       req.WriteOffLimit,
		WriteOffCostCenter:  req.WriteOffCostCenter,
		AllowPartialPayment: 1, // 允许部分支付
		DisableRoundedTotal: 1, // 禁用四舍五入总金额
	}
	if len(req.ApplicableForUsers) > 0 {
		profile.ApplicableForUsers = make([]erp.POSProfileUser, 0)
		for _, user := range req.ApplicableForUsers {
			profile.ApplicableForUsers = append(profile.ApplicableForUsers, erp.POSProfileUser{
				User: user.User,
			})
		}
	}
	// 处理支付方式
	profile.Payments = s.buildPosProfilePaymentMethods(req.Payments)

	return profile
}

// buildPosProfilePaymentMethods 构建支付方式列表
// 参数：
//   - payments: 支付方式列表
//
// 返回：
//   - []erp.POSPaymentMethod: 支付方式方法列表
func (s *sSelling) buildPosProfilePaymentMethods(payments []string) []erp.POSPaymentMethod {
	paymentMethods := make([]erp.POSPaymentMethod, 0, len(payments))

	for _, payment := range payments {
		paymentInfo := erp.POSPaymentMethod{
			ModeOfPayment:  payment,
			AllowInReturns: DefaultPaymentAllowInReturns,
		}

		// 默认现金
		if payment == "Cash" {
			paymentInfo.Default = DefaultCashPaymentDefault
		}

		paymentMethods = append(paymentMethods, paymentInfo)
	}

	return paymentMethods
}

// parsePosProfileResponse 解析POS配置文件响应
// 参数：
//   - resp: 响应数据
//
// 返回：
//   - *erp.POSProfile: 解析后的POS配置文件
//   - error: 错误信息
func (s *sSelling) parsePosProfileResponse(resp *gjson.Json) (*erp.POSProfile, error) {
	posProfile := &erp.POSProfile{}
	j := resp

	j.GetJson("data").Scan(posProfile)
	return posProfile, nil
}

// OpenPosEntry 开帐
// 参数：
//   - ctx: 上下文对象
//   - req: 开帐请求参数
//
// 返回：
//   - *selling.OpenPosEntryResp: 开帐响应信息
//   - error: 错误信息
func (s *sSelling) OpenPosEntry(ctx context.Context, req *selling.OpenPosEntryReq) (*selling.OpenPosEntryResp, error) {
	// 检查POS配置文件是否已开帐
	// opening, err := s.IsProfileOpening(ctx, req.PosProfileName, req.CashierEmail)
	//if err != nil {
	//	return nil, gerror.Wrapf(err, "查询POS配置文件是否开帐失败")
	//}
	//新的异步模式下，不再约束已开账检查
	// if opening {
	// return nil, gerror.New("POS配置文件已开帐")
	// }

	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	// 构建开账明细（支持 payment_id 自动查询 mode_of_payment）
	openDetails, err := s.buildOpeningEntryDetails(ctx, req.OpenPosEntryDetail)
	if err != nil {
		return nil, err
	}
	reqInfo := s.buildOpeningEntryRequest(req, companyName, openDetails)

	// 创建开帐记录
	resp, err := service.Document().Create(ctx, erp.DocTypePosOpeningEntry, reqInfo)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建开帐记录失败")
	}

	// 解析响应数据
	openEntry, err := s.parseOpeningEntryResponse(resp)
	if err != nil {
		return nil, err
	}

	// 提交开帐记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosOpeningEntry, openEntry.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交开帐记录失败")
	}

	return &selling.OpenPosEntryResp{
		OpenPosEntryInfo: &selling.OpenPosEntryInfo{
			OpenPosEntryName:   openEntry.Name,
			PosProfileName:     openEntry.PosProfile,
			CashierEmail:       openEntry.User,
			CompanyAbbr:        req.CompanyAbbr,
			OpenPosEntryDetail: req.OpenPosEntryDetail,
		},
	}, nil
}

// buildOpeningEntryDetails 构建开账明细
// 参数：
//   - ctx: 上下文
//   - details: 开账明细列表
//
// 返回：
//   - []erp.POSOpeningEntryDetail: 开账明细列表
//   - error: 错误信息
func (s *sSelling) buildOpeningEntryDetails(ctx context.Context, details []*selling.OpenPosEntryDetail) ([]erp.POSOpeningEntryDetail, error) {
	var openingDetails []erp.POSOpeningEntryDetail

	for i, detail := range details {
		var modeOfPayment string

		// 如果提供了 payment_id，自动查询 mode_of_payment
		if detail.PaymentId != nil && *detail.PaymentId != "" {
			getModeResp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
				PaymentId: detail.PaymentId,
			})
			if err != nil {
				g.Log().Error(ctx, "查询支付方式失败",
					g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
				return nil, gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", *detail.PaymentId)
			}

			if getModeResp == nil || getModeResp.Name == "" {
				return nil, gerror.Newf("支付方式不存在或未启用，payment_id: %s", *detail.PaymentId)
			}

			modeOfPayment = getModeResp.Name

			g.Log().Info(ctx, "开账详情: 通过 payment_id 查询到 mode_of_payment",
				g.Map{"index": i, "payment_id": *detail.PaymentId, "mode_of_payment": modeOfPayment})
		} else if detail.ModeOfPayment != nil {
			// 直接使用 mode_of_payment（向后兼容）
			modeOfPayment = *detail.ModeOfPayment
		}

		openingDetails = append(openingDetails, erp.POSOpeningEntryDetail{
			ModeOfPayment: modeOfPayment,
			OpeningAmount: detail.OpeningAmount,
		})
	}

	return openingDetails, nil
}

// buildOpeningEntryRequest 构建开帐请求
// 参数：
//   - req: 开帐请求参数
//   - companyName: 公司名称
//   - openDetails: 开帐明细列表
//
// 返回：
//   - *erp.POSOpeningEntry: 开帐请求信息
func (s *sSelling) buildOpeningEntryRequest(req *selling.OpenPosEntryReq, companyName string, openDetails []erp.POSOpeningEntryDetail) *erp.POSOpeningEntry {
	return &erp.POSOpeningEntry{
		PosProfile:      req.PosProfileName,
		Company:         companyName,
		PeriodStartDate: service.Setup().MustGetLocalDateTime(gctx.GetInitCtx(), gtime.New(req.PeriodStartDate)).Format("Y-m-d H:i:s"),
		User:            req.CashierEmail,
		BalanceDetails:  openDetails,
	}
}

// parseOpeningEntryResponse 解析开帐响应
// 参数：
//   - resp: 响应数据
//
// 返回：
//   - *erp.POSOpeningEntry: 解析后的开帐信息
//   - error: 错误信息
func (s *sSelling) parseOpeningEntryResponse(resp *gjson.Json) (*erp.POSOpeningEntry, error) {
	res := &erp.POSOpeningEntry{}
	j := resp

	gconv.Scan(j.GetJson("data"), res)
	return res, nil
}

// ClosePosEntry 关帐
// 参数：
//   - ctx: 上下文对象
//   - req: 关帐请求参数
//
// 返回：
//   - *selling.ClosePosEntryResp: 关帐响应信息
//   - error: 错误信息
func (s *sSelling) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error) {
	// 构建关帐明细（支持 payment_id 自动查询 mode_of_payment）
	closeDetails, err := s.buildClosingEntryDetails(ctx, req.ClosePosEntryDetail)
	if err != nil {
		return nil, err
	}

	// 构建关帐请求
	reqInfo := s.buildClosingEntryRequest(req, closeDetails)

	// 获取开帐信息
	openEntry, err := s.GetPosOpeningEntry(ctx, req.PosOpenEntryName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取开帐信息失败")
	}
	reqInfo.PosProfile = openEntry.PosProfile
	reqInfo.Company = openEntry.Company

	// 获取期间发票
	invoices, err := s.GetPosInvoiceList(ctx, &dtoSelling.GetPosInvoiceListReq{
		PosProfile: openEntry.PosProfile,
		//StartDate:  openEntry.PeriodStartDate,
		//EndDate:    service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.PeriodEndDate)).Format("Y-m-d H:i:s"),
		User:                  openEntry.User,
		Docstatus:             erp.DocstatusSubmitted,
		CustomPosOpeningEntry: req.PosOpenEntryName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取期间发票失败")
	}

	//如果请求订单数大于0，且请求订单数小于发票数，则返回错误。 ttpos订单如果包含物品和商品，在erpnext上会有2个pos invoice
	// if req.InvoiceCount > 0 && req.InvoiceCount < int64(len(invoices)) {
	// 	return nil, gerror.Newf("关账订单数(%d)与请求订单数(%d)不一致", len(invoices), req.InvoiceCount)
	// }

	// 将发票记录到关帐信息
	reqInfo.PosTransactions = s.buildPosTransactions(invoices)

	// 创建关帐记录
	resp, err := service.Document().Create(ctx, erp.DocTypePosClosingEntry, reqInfo)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建关帐记录失败")
	}

	// 解析响应数据
	closeEntry, err := s.parseClosingEntryResponse(resp)
	if err != nil {
		return nil, err
	}

	// 提交关帐记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosClosingEntry, closeEntry.Name, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交关帐记录失败")
	}

	return &selling.ClosePosEntryResp{
		ClosePosEntryInfo: &selling.ClosePosEntryInfo{
			ClosePosEntryName:   closeEntry.Name,
			PosProfileName:      closeEntry.PosProfile,
			ClosePosEntryDetail: req.ClosePosEntryDetail,
		},
	}, nil
}

// buildClosingEntryDetails 构建关帐明细
// 参数：
//   - ctx: 上下文
//   - details: 关帐明细列表
//
// 返回：
//   - []erp.POSPaymentReconciliation: 关帐明细列表
//   - error: 错误信息
func (s *sSelling) buildClosingEntryDetails(ctx context.Context, details []*selling.ClosePosEntryDetail) ([]erp.POSPaymentReconciliation, error) {
	closeDetails := make([]erp.POSPaymentReconciliation, 0)
	for i, detail := range details {
		var modeOfPayment string

		// 如果提供了 payment_id，自动查询 mode_of_payment
		if detail.PaymentId != nil && *detail.PaymentId != "" {
			getModeResp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
				PaymentId: detail.PaymentId,
			})
			if err != nil {
				g.Log().Error(ctx, "查询支付方式失败",
					g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
				return nil, gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", *detail.PaymentId)
			}

			if getModeResp == nil || getModeResp.Name == "" {
				return nil, gerror.Newf("支付方式不存在或未启用，payment_id: %s", *detail.PaymentId)
			}

			modeOfPayment = getModeResp.Name

			g.Log().Info(ctx, "关账详情: 通过 payment_id 查询到 mode_of_payment",
				g.Map{"index": i, "payment_id": *detail.PaymentId, "mode_of_payment": modeOfPayment})
		} else if detail.ModeOfPayment != nil {
			// 直接使用 mode_of_payment（向后兼容）
			modeOfPayment = *detail.ModeOfPayment
		}

		closeDetails = append(closeDetails, erp.POSPaymentReconciliation{
			ModeOfPayment: modeOfPayment,
			ClosingAmount: detail.ClosingAmount,
			OpeningAmount: detail.OpeningAmount,
		})
	}
	return closeDetails, nil
}

// buildClosingEntryRequest 构建关帐请求
// 参数：
//   - req: 关帐请求参数
//   - closeDetails: 关帐明细列表
//
// 返回：
//   - *erp.POSCloseEntry: 关帐请求信息
func (s *sSelling) buildClosingEntryRequest(req *selling.ClosePosEntryReq, closeDetails []erp.POSPaymentReconciliation) *erp.POSCloseEntry {
	return &erp.POSCloseEntry{
		PosOpeningEntry:       req.PosOpenEntryName,
		PeriodEndDate:         service.Setup().MustGetLocalDateTime(gctx.GetInitCtx(), gtime.New(req.PeriodEndDate)).Format("Y-m-d H:i:s"),
		PaymentReconciliation: closeDetails,
	}
}

// buildPosTransactions 构建POS交易记录
// 参数：
//   - invoices: 发票列表
//
// 返回：
//   - []erp.POSTransaction: POS交易记录列表
func (s *sSelling) buildPosTransactions(invoices []dtoSelling.SimplePosInvoice) []erp.POSTransaction {
	transactions := make([]erp.POSTransaction, 0)
	for _, invoice := range invoices {
		transactions = append(transactions, erp.POSTransaction{
			PosInvoice:  invoice.Name,
			PostingDate: invoice.PostingDate,
			GrandTotal:  invoice.GrandTotal,
		})
	}
	return transactions
}

// parseClosingEntryResponse 解析关帐响应
// 参数：
//   - resp: 响应数据
//
// 返回：
//   - *erp.POSCloseEntry: 解析后的关帐信息
//   - error: 错误信息
func (s *sSelling) parseClosingEntryResponse(resp *gjson.Json) (*erp.POSCloseEntry, error) {
	res := &erp.POSCloseEntry{}
	j := resp

	gconv.Scan(j.GetJson("data"), res)
	return res, nil
}

// IsProfileOpening 查询POS配置文件是否开帐
// 参数：
//   - ctx: 上下文对象
//   - posProfile: POS配置文件名称
//
// 返回：
//   - bool: 是否已开帐
//   - error: 错误信息
func (s *sSelling) IsProfileOpening(ctx context.Context, posProfile, user string) (bool, error) {
	count, err := service.Doctype().Count(ctx, &erp.ErpReq{
		DocType: "POS Opening Entry",
	}, &erp.RequestParams{
		Filters: [][]string{{"pos_profile", "=", posProfile}, {"status", "=", "Open"}, {"user", "=", user}},
	})
	if err != nil {
		return false, gerror.Wrapf(err, "查询POS配置文件是否开帐失败")
	}
	return count > 0, nil
}

// GetPosInvoiceList 获取POS发票列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取POS发票列表请求参数
//
// 返回：
//   - []dtoSelling.SimplePosInvoice: POS发票列表
//   - error: 错误信息
func (s *sSelling) GetPosInvoiceList(ctx context.Context, req *dtoSelling.GetPosInvoiceListReq) ([]dtoSelling.SimplePosInvoice, error) {
	filters := make([][]string, 0)
	filters = append(filters, []string{"pos_profile", "=", req.PosProfile})
	filters = append(filters, []string{"owner", "=", req.User})
	filters = append(filters, []string{"docstatus", "=", req.Docstatus})
	if req.CustomPosOpeningEntry != "" {
		filters = append(filters, []string{"custom_pos_opening_entry", "=", req.CustomPosOpeningEntry})
	}
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosInvoice,
	}, &erp.RequestParams{
		Fields:  g.ArrayStr{"name", "posting_date", "customer", "grand_total", "is_return", "return_against"},
		Filters: filters,
		Limit:   consts.Limit9999,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询POS发票列表失败")
	}

	// 解析响应数据
	res := make([]dtoSelling.SimplePosInvoice, 0)
	j := resp

	err = gconv.Scan(j.Get("data"), &res)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS发票数据列表响应失败")
	}

	return res, nil
}

// resolvePaymentIDs 解析支付列表中的 payment_id
// 参数：
//   - ctx: 上下文对象
//   - payments: 支付列表
//
// 返回：
//   - error: 错误信息
//
// 功能：
//   - 遍历支付列表，如果 payment_id 不为空，调用 GetModeOfPayment 查询 mode_of_payment
//   - 验证支付方式是否启用（enabled = true）
//   - 使用缓存优化：相同 payment_id 只查询一次
func (s *sSelling) resolvePaymentIDs(ctx context.Context, payments []*selling.PosInvoicePayment) error {
	// 1. 收集所有需要解析的 payment_id（去重）
	cache := make(map[string]string) // payment_id -> mode_of_payment

	for i, payment := range payments {
		// 2. 如果 payment_id 不为空，进行解析
		if payment.PaymentId != nil && *payment.PaymentId != "" {
			paymentID := *payment.PaymentId

			// 检查缓存
			if modeOfPayment, cached := cache[paymentID]; cached {
				g.Log().Infof(ctx, "[resolvePaymentIDs] 使用缓存: payment_id=%s -> mode_of_payment=%s",
					paymentID, modeOfPayment)
				payment.ModeOfPayment = modeOfPayment
				continue
			}

			// 3. 调用 GetModeOfPayment 查询
			g.Log().Infof(ctx, "[resolvePaymentIDs] 解析支付项 %d: payment_id=%s", i+1, paymentID)
			mopResp, err := s.GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
				PaymentId: &paymentID,
			})
			if err != nil {
				g.Log().Errorf(ctx, "[resolvePaymentIDs] 解析支付项 %d 失败: payment_id=%s, err=%v",
					i+1, paymentID, err)
				return gerror.Wrapf(err, "解析支付项 %d 的 payment_id 失败", i+1)
			}

			// 4. 验证支付方式是否启用
			if !mopResp.Enabled {
				g.Log().Warningf(ctx, "[resolvePaymentIDs] 支付方式已禁用: name=%s, payment_id=%s",
					mopResp.Name, paymentID)
				return gerror.Newf("支付项 %d: 支付方式 %s 已禁用", i+1, mopResp.Name)
			}

			// 5. 赋值并缓存
			payment.ModeOfPayment = mopResp.Name
			cache[paymentID] = mopResp.Name

			g.Log().Infof(ctx, "[resolvePaymentIDs] 解析成功: payment_id=%s -> mode_of_payment=%s",
				paymentID, mopResp.Name)
		}

		// 6. 验证 mode_of_payment 是否存在
		if payment.ModeOfPayment == "" {
			g.Log().Errorf(ctx, "[resolvePaymentIDs] 支付项 %d: mode_of_payment 和 payment_id 都未提供", i+1)
			return gerror.Newf("支付项 %d: mode_of_payment 和 payment_id 至少提供一个", i+1)
		}
	}

	g.Log().Infof(ctx, "[resolvePaymentIDs] 支付项解析完成，共处理 %d 项，缓存 %d 个",
		len(payments), len(cache))
	return nil
}

// SavePosInvoice 保存POS发票
// 参数：
//   - ctx: 上下文对象
//   - req: 保存POS发票请求参数
//
// 返回：
//   - *selling.SavePosInvoiceResp: 保存POS发票响应信息
//   - error: 错误信息
func (s *sSelling) SavePosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq) (*selling.SavePosInvoiceResp, error) {
	// 新增：支付项预处理 - 解析 payment_id
	if err := s.resolvePaymentIDs(ctx, req.Payments); err != nil {
		return nil, err
	}

	// 获取开帐记录
	openingEntry, err := s.GetPosOpeningEntry(ctx, req.OpenPosEntryName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取POS开帐记录失败")
	}

	res := &selling.SavePosInvoiceResp{}

	if len(req.MaterialItems) > 0 {
		invoiceName, err := s.SavePosInvoiceStep(ctx, req, openingEntry, true)
		if err != nil {
			//取消已保存草稿的物品POS发票
			if len(invoiceName) > 0 {
				if err := s.CancelPosInvoice(ctx, invoiceName); err != nil {
					g.Log().Errorf(ctx, "取消物品POS发票失败，发票名称：%s，错误信息：%v", invoiceName, err)
					//取消失败返回前台的是保存失败的原因
				}
			}
			return nil, gerror.Wrapf(err, "保存物品POS发票失败")
		}
		res.MaterialInvoiceName = invoiceName
	}
	invoiceName, err := s.SavePosInvoiceStep(ctx, req, openingEntry, false)
	if err != nil {
		if res.MaterialInvoiceName != "" {
			if err := s.CancelPosInvoice(ctx, res.MaterialInvoiceName); err != nil {
				g.Log().Errorf(ctx, "取消物品POS发票失败，发票名称：%s，错误信息：%v", res.MaterialInvoiceName, err)
				//取消失败返回前台的是保存失败的原因
			}
		}
		//取消已保存草稿的商品POS发票
		if len(invoiceName) > 0 {
			if err := s.CancelPosInvoice(ctx, invoiceName); err != nil {
				g.Log().Errorf(ctx, "取消商品POS发票失败，发票名称：%s，错误信息：%v", invoiceName, err)
				//取消失败返回前台的是保存失败的原因
			}
		}
		return nil, gerror.Wrapf(err, "保存商品POS发票失败")
	}
	res.ProductsInvoiceName = invoiceName

	return res, nil
}

// SavePosInvoiceStep 保存POS发票步骤
// 参数：
//   - ctx: 上下文对象
//   - req: 保存POS发票请求参数
//   - openingEntry: POS开帐记录
//   - isMaterialItem: 是否为物品发票
//
// 返回：
//   - string: POS发票名称
//   - error: 错误信息
func (s *sSelling) SavePosInvoiceStep(ctx context.Context, req *selling.SavePosInvoiceReq, openingEntry *erp.POSOpeningEntry, isMaterialItem bool) (string, error) {
	var InvoiceName = ""
	// 构建POS发票
	posInvoice := s.buildPosInvoice(ctx, req, openingEntry)
	if isMaterialItem {
		posInvoice.Items = s.buildInvoiceItems(req.MaterialItems)
		//物品无税费
		posInvoice.Taxes = nil
		posInvoice.UpdateStock = req.UpdateStock // 更新库存
	}
	//反结账后重新结账
	if len(req.AmendedProductsInvoiceName) > 0 {
		posInvoice.AmendedFrom = req.AmendedProductsInvoiceName
	}
	// 创建POS发票
	// 特殊处理，使用开帐收银员创建发票
	//创建商品销售记录
	resp, err := service.Document().Create(erpnext.SetFakeUser(ctx, openingEntry.User), erp.DocTypePosInvoice, posInvoice)
	if err != nil {
		return InvoiceName, gerror.Wrapf(err, "创建POS发票失败")
	}

	// 解析响应数据
	j := resp
	InvoiceName = j.Get("data.name").String()

	// 提交发票记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosInvoice, InvoiceName, erp.DocstatusSubmitted)
	if err != nil {
		return InvoiceName, gerror.Wrapf(err, "提交发票记录失败")
	}
	return InvoiceName, nil
}

// buildPosInvoice 构建POS发票
// 参数：
//   - req: 保存POS发票请求参数
//   - companyName: 公司名称
//   - posProfile: POS配置文件名称
//
// 返回：
//   - *erp.POSInvoice: POS发票信息
func (s *sSelling) buildPosInvoice(ctx context.Context, req *selling.SavePosInvoiceReq, openingEntry *erp.POSOpeningEntry) *erp.POSInvoice {
	postingDatetime, _ := gtime.New(req.PostingDatetime).ToZone(service.User().MustGetUserTimeZone(ctx, openingEntry.User))
	//postingDatetime := service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.PostingDatetime))

	posInvoice := &erp.POSInvoice{
		PosProfile:            openingEntry.PosProfile,
		Company:               openingEntry.Company,
		PostingDate:           postingDatetime.Format(DateFormat),
		PostingTime:           postingDatetime.Format(TimeFormat),
		Currency:              req.Currency,
		PriceListCurrency:     req.PriceListCurrency,
		UpdateStock:           req.UpdateStock, // 更新库存, 如果有 1 个 pos invoice 不为 1，sales invoice 的 update_stock 也不会为 1.
		CustomerOrder:         req.OrderNo,
		SetPostingTime:        1,
		CustomPosOpeningEntry: req.OpenPosEntryName,
	}

	// 设置客户信息
	if len(req.CustomerUuid) > 0 {
		posInvoice.Customer = "Member"
		posInvoice.CustomerUUID = req.CustomerUuid
	} else {
		posInvoice.Customer = "Default"
	}

	// 构建发票项目
	posInvoice.Items = s.buildInvoiceItems(req.Items)

	// 构建发票税费
	posInvoice.Taxes = s.buildInvoiceTaxes(req.Taxes, req.CompanyAbbr)

	// 构建支付信息
	posInvoice.Payments = s.buildInvoicePayments(req.Payments)

	// 设置备注
	if len(req.Remark) > 0 {
		posInvoice.Remarks = req.Remark
	}

	// 设置外卖订单字段
	if len(req.GetTakeoutOrderNo()) > 0 {
		posInvoice.CustomTakeoutOrderNo = req.GetTakeoutOrderNo()
	}
	if len(req.GetTakeoutProvider()) > 0 {
		posInvoice.CustomTakeoutProvider = req.GetTakeoutProvider()
	}

	return posInvoice
}

// buildInvoiceItems 构建发票项目
// 参数：
//   - items: 发票项目列表
//
// 返回：
//   - []erp.POSInvoiceItem: 发票项目列表
func (s *sSelling) buildInvoiceItems(items []*selling.PosInvoiceItem) []erp.POSInvoiceItem {
	invoiceItems := make([]erp.POSInvoiceItem, 0, len(items))
	for _, item := range items {
		// 构建发票项目
		invoiceItem := erp.POSInvoiceItem{
			ItemCode:    item.ItemCode,
			Qty:         item.Qty,
			Rate:        item.Rate,
			Amount:      item.Amount,
			Description: item.Description,
			Uom:         item.Uom,
			IsFreeItem:  item.IsFreeItem,
		}
		//如果是免费物品，金额为0
		if item.IsFreeItem {
			invoiceItem.Rate = 0
			invoiceItem.BaseRate = 0
			invoiceItem.BaseAmount = 0
			invoiceItem.Amount = 0
			invoiceItem.DiscountAmount = item.Amount
			invoiceItem.DiscountPercentage = 100
		}
		invoiceItems = append(invoiceItems, invoiceItem)
	}
	return invoiceItems
}

// buildInvoiceTaxes 构建发票税费
// 参数：
//   - taxes: 税费列表
//   - companyAbbr: 公司缩写
//
// 返回：
//   - []erp.POSInvoiceTax: 发票税费列表
func (s *sSelling) buildInvoiceTaxes(taxes []*selling.PosInvoiceTax, companyAbbr string) []erp.POSInvoiceTax {
	invoiceTaxes := make([]erp.POSInvoiceTax, 0, len(taxes))
	for _, tax := range taxes {
		invoiceTaxes = append(invoiceTaxes, erp.POSInvoiceTax{
			ChargeType:  DefaultTaxChargeType,                      // 默认实际
			AccountHead: DefaultTaxAccountHeadPrefix + companyAbbr, // 默认现金
			Rate:        tax.Rate,
			TaxAmount:   tax.TaxAmount,
			Description: tax.Description,
		})
	}
	return invoiceTaxes
}

// GetPosOpeningEntry 获取POS开帐记录
// 参数：
//   - ctx: 上下文对象
//   - name: 开帐记录名称
//
// 返回：
//   - *erp.POSOpeningEntry: POS开帐记录信息
//   - error: 错误信息
func (s *sSelling) GetPosOpeningEntry(ctx context.Context, name string) (*erp.POSOpeningEntry, error) {
	resp, err := service.Document().Get(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosOpeningEntry,
		Name:    name,
	}, nil)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询POS开帐记录失败")
	}

	// 解析响应数据
	res := &erp.POSOpeningEntry{}
	j := resp

	gconv.Scan(j.Get("data"), res)
	return res, nil
}

// buildInvoicePayments 构建发票支付信息
// 参数：
//   - payments: 支付信息列表
//
// 返回：
//   - []erp.POSInvoicePayment: 发票支付信息列表
func (s *sSelling) buildInvoicePayments(payments []*selling.PosInvoicePayment) []erp.POSInvoicePayment {
	invoicePayments := make([]erp.POSInvoicePayment, 0, len(payments))
	for _, payment := range payments {
		invoicePayments = append(invoicePayments, erp.POSInvoicePayment{
			ModeOfPayment: payment.ModeOfPayment,
			Amount:        payment.Amount,
		})
	}
	return invoicePayments
}

// ReturnPosInvoice 退货POS发票
// 参数：
//   - ctx: 上下文对象
//   - req: 退货POS发票请求参数
//
// 返回：
//   - *selling.ReturnPosInvoiceResp: 退货POS发票响应参数
//   - error: 错误信息
//
// 功能：
//   - 退货指定名称的POS发票
func (s *sSelling) ReturnPosInvoice(ctx context.Context, req *selling.ReturnPosInvoiceReq) (*selling.ReturnPosInvoiceResp, error) {

	// 获取开帐记录
	openingEntry, err := s.GetPosOpeningEntry(ctx, req.OpenPosEntryName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取POS开帐记录失败")
	}

	//
	resp, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiMethodMakeMappedDoc,
	}, g.MapStrStr{
		"method":      "erpnext.accounts.doctype.pos_invoice.pos_invoice.make_sales_return",
		"source_name": req.InvoiceName,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售退款订单失败")
	}
	// 解析响应数据
	j := resp

	saleInvoice := &erp.POSInvoice{}
	// 解析响应数据
	err = j.GetJson("data").Scan(saleInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析原销售订单响应失败")
	}

	//postingDatetime := service.Setup().MustGetLocalDateTime(ctx, gtime.New(req.PostingDatetime))

	grandTotal := 0.0
	for _, payment := range req.Payments {
		grandTotal += payment.Amount
	}

	returnInvoice := &erp.POSInvoice{
		CustomerOrder:     req.OrderNo,
		Items:             s.buildInvoiceItems(req.Items),
		Payments:          s.buildInvoicePayments(req.Payments),
		Taxes:             s.buildInvoiceTaxes(req.Taxes, req.CompanyAbbr),
		PosProfile:        saleInvoice.PosProfile,
		Company:           saleInvoice.Company,
		Currency:          saleInvoice.Currency,
		PriceListCurrency: saleInvoice.PriceListCurrency,
		//PostingDate:       postingDatetime.Format(DateFormat),
		//PostingTime:       postingDatetime.Format(TimeFormat),
		UpdateStock:   0,
		ReturnAgainst: req.InvoiceName,
		IsReturn:      1,
		IsPos:         1,
		GrandTotal:    grandTotal,
		PaidAmount:    grandTotal,
		//SetPostingTime:    1,
		Customer:           saleInvoice.Customer,
		WriteOffAccount:    saleInvoice.WriteOffAccount,
		WriteOffCostCenter: saleInvoice.WriteOffCostCenter,

		//UpdateBilledAmountInSalesOrder: 1,
		CustomPosOpeningEntry: req.OpenPosEntryName,
	}

	//totalTaxAndCharges := 0.0
	//for _, tax := range returnInvoice.Taxes {
	//	totalTaxAndCharges += tax.TaxAmount
	//}
	//returnInvoice.TotalTaxesAndCharges = totalTaxAndCharges
	//netTotal := 0.0
	//for _, item := range returnInvoice.Items {
	//	netTotal += item.Amount
	//}
	//returnInvoice.NetTotal = netTotal

	//创建物品销售记录
	resp, err = service.Document().Create(erpnext.SetFakeUser(ctx, openingEntry.User), erp.DocTypePosInvoice, returnInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售退款订单发票失败")
	}
	// 解析响应数据
	j = resp

	res := &selling.ReturnPosInvoiceResp{
		InvoiceName: j.Get("data.name").String(),
	}

	// 提交发票记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosInvoice, res.InvoiceName, erp.DocstatusSubmitted)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交销售退款订单发票记录失败")
	}
	return res, nil
}

// CancelPosInvoice 取消POS发票
// 参数：
//   - ctx: 上下文对象
//   - invoiceName: 发票名称
//
// 返回：
//   - error: 错误信息
//
// 功能：
//   - 取消指定名称的POS发票
func (*sSelling) CancelPosInvoice(ctx context.Context, invoiceName string) error {
	_, err := service.Rpc().Execute(ctx, &erp.ErpReq{
		Method: erp.ApiSaveCancel,
	}, g.MapStrStr{
		"doctype": erp.DocTypePosInvoice,
		"name":    invoiceName,
	})
	if err != nil {
		return gerror.Wrapf(err, "取消发票[%s]失败", invoiceName)
	}
	return nil
}

//func (*sSelling) AmendPosInvoice(ctx context.Context, invoiceName string) error {
//	_, err := service.Rpc().Execute(ctx, &erp.ErpReq{
//		Method: erp.ApiIsDocumentAmend,
//	}, g.MapStrStr{
//		"doctype": erp.DocTypePosInvoice,
//		"name":    invoiceName,
//	})
//	if err != nil {
//		return gerror.Wrapf(err, "取消发票[%s]失败", invoiceName)
//	}
//	return nil
//}

// queryModeOfPaymentList 查询支付方式列表（内部方法）
// 参数：
//   - ctx: 上下文对象
//   - filters: 查询过滤条件
//   - limit: 返回数量限制，0 表示不限制
//
// 返回：
//   - []*selling.ModeOfPayment: 支付方式列表
//   - error: 错误信息
func (s *sSelling) queryModeOfPaymentList(ctx context.Context, filters [][]string, limit int) ([]*selling.ModeOfPayment, error) {
	params := &erp.RequestParams{
		Fields:  []string{"name", "enabled", "custom_payment_id"},
		Filters: filters,
	}

	if limit > 0 {
		params.Limit = limit
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
	}, params)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询支付方式失败")
	}

	// 手动解析数据，确保字段映射正确
	dataArray := resp.GetJsons("data")
	result := make([]*selling.ModeOfPayment, 0, len(dataArray))

	for _, item := range dataArray {
		modeOfPayment := &selling.ModeOfPayment{
			Name:      item.Get("name").String(),
			Enabled:   item.Get("enabled").Int() == 1,
			PaymentId: item.Get("custom_payment_id").String(),
		}
		result = append(result, modeOfPayment)
	}

	return result, nil
}

// GetModeOfPaymentList 获取支付方式列表
// 参数：
//   - ctx: 上下文对象
//   - req: 获取支付方式列表请求参数
//
// 返回：
//   - *selling.GetModeOfPaymentListResp: 获取支付方式列表响应参数
//   - error: 错误信息
//
// 功能：
//   - 获取支付方式列表（包含所有支付方式，不管启用状态）
//   - 从ERP读取enabled字段并填充到响应中
func (s *sSelling) GetModeOfPaymentList(ctx context.Context, req *selling.GetModeOfPaymentListReq) (*selling.GetModeOfPaymentListResp, error) {
	filters := make([][]string, 0)
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取公司[%s]失败", req.CompanyAbbr)
		}
		filters = append(filters, []string{"custom_company", "=", companyName})
	}
	if len(req.Branch) > 0 {
		filters = append(filters, []string{"custom_branch", "=", req.Branch})
	}

	// 调用统一查询方法，直接获取解析后的数据
	modeOfPaymentList, err := s.queryModeOfPaymentList(ctx, filters, 0)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取支付方式列表失败")
	}

	return &selling.GetModeOfPaymentListResp{
		ModeOfPaymentList: modeOfPaymentList,
	}, nil
}

// GetModeOfPayment 查询单个支付方式
// 参数：
//   - ctx: 上下文对象
//   - req: 查询单个支付方式请求参数（name 或 payment_id 至少提供一个）
//
// 返回：
//   - *selling.ModeOfPayment: 支付方式信息
//   - error: 错误信息
//
// 功能：
//   - 支持通过 name 查询（主键查询，性能最优）
//   - 支持通过 payment_id 查询（Filter 查询）
func (s *sSelling) GetModeOfPayment(ctx context.Context, req *selling.GetModeOfPaymentReq) (*selling.ModeOfPayment, error) {
	// 1. 参数校验
	if (req.Name == nil || *req.Name == "") && (req.PaymentId == nil || *req.PaymentId == "") {
		return nil, gerror.New("name 或 payment_id 至少提供一个")
	}

	var modeOfPayment *selling.ModeOfPayment

	// 2. 通过 name 查询（主键查询，性能最优）
	if req.Name != nil && *req.Name != "" {
		g.Log().Infof(ctx, "[GetModeOfPayment] 通过 name 查询支付方式: name=%s", *req.Name)

		resp, err := service.Document().Get(ctx, &erp.ErpReq{
			DocType: erp.DocTypeModeOfPayment,
			Name:    *req.Name,
		}, &erp.RequestParams{
			Fields: []string{"name", "enabled", "custom_payment_id"},
		})
		if err != nil {
			g.Log().Errorf(ctx, "[GetModeOfPayment] 查询支付方式失败: name=%s, err=%v", *req.Name, err)
			return nil, gerror.Wrapf(err, "查询支付方式失败: name=%s", *req.Name)
		}

		// 数据映射
		modeOfPayment = &selling.ModeOfPayment{
			Name:      resp.Get("data.name").String(),
			Enabled:   resp.Get("data.enabled").Int() == 1,
			PaymentId: resp.Get("data.custom_payment_id").String(),
		}

		g.Log().Infof(ctx, "[GetModeOfPayment] 查询成功: name=%s, enabled=%v, payment_id=%s",
			modeOfPayment.Name, modeOfPayment.Enabled, modeOfPayment.PaymentId)
		return modeOfPayment, nil
	}

	// 3. 通过 payment_id 查询（Filter 查询）
	g.Log().Infof(ctx, "[GetModeOfPayment] 通过 payment_id 查询支付方式: payment_id=%s", *req.PaymentId)

	filters := [][]string{{"custom_payment_id", "=", *req.PaymentId}}
	modeOfPaymentList, err := s.queryModeOfPaymentList(ctx, filters, 1)
	if err != nil {
		g.Log().Errorf(ctx, "[GetModeOfPayment] 查询支付方式失败: payment_id=%s, err=%v", *req.PaymentId, err)
		return nil, gerror.Wrapf(err, "查询支付方式失败: payment_id=%s", *req.PaymentId)
	}

	// 4. 检查结果
	if len(modeOfPaymentList) == 0 {
		g.Log().Warningf(ctx, "[GetModeOfPayment] 支付方式不存在: payment_id=%s", *req.PaymentId)
		return nil, gerror.Newf("支付方式不存在: payment_id=%s", *req.PaymentId)
	}

	// 5. 返回第一条记录
	modeOfPayment = modeOfPaymentList[0]

	g.Log().Infof(ctx, "[GetModeOfPayment] 查询成功: name=%s, enabled=%v, payment_id=%s",
		modeOfPayment.Name, modeOfPayment.Enabled, modeOfPayment.PaymentId)
	return modeOfPayment, nil
}

// SaveModeOfPayment 保存/同步支付方式
func (s *sSelling) SaveModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
	// 判断是更新操作还是创建操作
	// 如果传入了 name 或 payment_id，则执行更新操作
	if (req.Name != nil && strings.TrimSpace(*req.Name) != "") ||
		(req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "") {
		return s.updateModeOfPayment(ctx, req)
	}

	// 否则执行创建操作（保持现有逻辑）
	return s.createModeOfPayment(ctx, req)
}

// createModeOfPayment 创建支付方式
func (s *sSelling) createModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司失败", req.CompanyAbbr)
	}

	// 允许 channel 为空，前缀仅拼接非空段，末尾统一加 -
	parts := make([]string, 0, 2)
	if strings.TrimSpace(req.Channel) != "" {
		parts = append(parts, req.Channel)
	}
	if strings.TrimSpace(req.PayType) != "" {
		parts = append(parts, req.PayType)
	}
	prefix := strings.Join(parts, "-") + "-"

	// 根据 added_by 字段决定序号生成策略
	var nextSeq int
	if req.AddedBy != nil && strings.TrimSpace(*req.AddedBy) == consts.PaymentAddedBySystem {
		// 系统创建：使用固定序号
		nextSeq = consts.PaymentSeqSystem
		g.Log().Infof(ctx, "[createModeOfPayment] 系统创建支付方式，使用固定序号 %04d, company=%s, prefix=%s",
			nextSeq, req.CompanyAbbr, prefix)
	} else {
		// 用户创建：自动递增序号
		nextSeq, err = s.nextModeOfPaymentSeq(ctx, prefix, companyName)
		if err != nil {
			return nil, err
		}
		g.Log().Debugf(ctx, "[createModeOfPayment] 用户创建支付方式，使用递增序号 %04d, company=%s, prefix=%s",
			nextSeq, req.CompanyAbbr, prefix)
	}

	name := fmt.Sprintf("%s%04d - %s", prefix, nextSeq, req.CompanyAbbr)

	// 生成或使用提供的 PaymentID
	paymentID := req.PaymentId
	if paymentID == "" {
		// 自动生成：PID + 16位数字
		paymentID = fmt.Sprintf("PID%d", uuid.MustGetID())
	}
	paymentType := "General"
	//名字是Cash 才设置成 Cash
	if name == "Cash" {
		paymentType = "Cash"
	}
	payload := g.Map{
		"mode_of_payment":   name,
		"name":              name,
		"type":              paymentType, // 需求默认通用类型，后续可按渠道扩展
		"custom_branch":     req.Branch,
		"custom_company":    companyName,
		"custom_payment_id": paymentID,
	}

	// 如果请求明确携带enabled字段，则更新ERP的启用状态
	// 如果未携带，则不设置enabled字段（保持ERP原值，兼容旧客户端）
	if req.Enabled != nil {
		// BoolValue.Value 为实际的bool值
		if req.GetEnabled() {
			payload["enabled"] = 1
		} else {
			payload["enabled"] = 0
		}
	}

	resp, err := service.Document().Create(ctx, erp.DocTypeModeOfPayment, payload)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建支付方式失败")
	}

	createdName := name
	if resp != nil {
		if v := resp.Get("data.name").String(); v != "" {
			createdName = v
		}
	}

	// 创建对应公司的支付账户关联，确保新支付方式可用
	if err := service.Selling().CreateModePaymentAccount(ctx, &setup.CreateModePaymentAccountInp{
		CompanyAbbr: req.CompanyAbbr,
		PaymentType: createdName,
	}); err != nil {
		return nil, gerror.Wrapf(err, "创建支付方式账号关联失败")
	}

	return &selling.SaveModeOfPaymentResp{
		Name:      createdName,
		PaymentId: paymentID, // 返回生成或指定的 PaymentID
	}, nil

}

// nextModeOfPaymentSeq 计算下一个支付方式序号
func (s *sSelling) nextModeOfPaymentSeq(ctx context.Context, prefix, companyName string) (int, error) {
	filters := [][]string{
		{"name", "like", prefix + "%"},
	}
	if companyName != "" {
		filters = append(filters, []string{"custom_company", "=", companyName})
	}

	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
	}, &erp.RequestParams{
		Fields:  []string{"name"},
		Filters: filters,
	})
	if err != nil {
		return 0, gerror.Wrapf(err, "获取支付方式列表失败")
	}

	maxSeq := -1
	for _, item := range resp.GetJsons("data") {
		name := item.Get("name").String()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		rest := strings.TrimPrefix(name, prefix)
		if len(rest) < 4 {
			continue
		}
		if idx := strings.Index(rest, " "); idx >= 0 {
			rest = rest[:idx]
		}
		seqVal, err := strconv.Atoi(rest)
		if err != nil {
			continue
		}
		if seqVal > maxSeq {
			maxSeq = seqVal
		}
	}

	next := maxSeq + 1
	if maxSeq < 0 {
		next = 1 //默认从1开始
	}
	if next > consts.Limit9999 {
		return 0, gerror.New("支付方式序号已超出上限")
	}
	return next, nil
}

// updateModeOfPayment 更新支付方式
func (s *sSelling) updateModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
	var resp *gjson.Json
	var err error
	var name string
	var queryKey string

	// 构建查询过滤器
	var filters [][]string

	// 1. 优先使用 PaymentId 查询（业务主键）
	if req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "" {
		paymentId := strings.TrimSpace(req.PaymentId)
		queryKey = fmt.Sprintf("payment_id=%s", paymentId)
		filters = [][]string{{"custom_payment_id", "=", paymentId}}
		g.Log().Infof(ctx, "[updateModeOfPayment] 通过 payment_id 查询支付方式: %s", queryKey)
	} else if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		// 2. 使用 Name 查询（ERP 主键）
		name = strings.TrimSpace(*req.Name)
		queryKey = fmt.Sprintf("name=%s", name)
		filters = [][]string{{"name", "=", name}}
		g.Log().Infof(ctx, "[updateModeOfPayment] 通过 name 查询支付方式: %s", queryKey)
	} else {
		return nil, gerror.New("name 或 payment_id 至少提供一个")
	}

	// 3. 统一使用 List 接口查询
	resp, err = service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
	}, &erp.RequestParams{
		Fields:  []string{"name", "custom_company", "custom_branch", "enabled", "custom_payment_id"},
		Filters: filters,
		Limit:   1,
	})
	if err != nil {
		g.Log().Errorf(ctx, "[updateModeOfPayment] 查询支付方式失败: %s, err=%v", queryKey, err)
		return nil, gerror.Wrapf(err, "查询支付方式失败")
	}

	// 4. 检查查询结果
	dataArray := resp.GetJsons("data")
	if len(dataArray) == 0 {
		g.Log().Warningf(ctx, "[updateModeOfPayment] 支付方式不存在: %s", queryKey)
		return nil, gerror.Newf("支付方式不存在: %s", queryKey)
	}

	// 5. 获取查询到的支付方式信息
	data := dataArray[0]
	name = data.Get("name").String()
	erpCompany := data.Get("custom_company").String()

	// 6. 权限校验：确认支付方式属于当前公司
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写[%s]查询公司失败", req.CompanyAbbr)
	}

	if erpCompany != companyName {
		g.Log().Warningf(ctx, "[updateModeOfPayment] 尝试越权修改支付方式: name=%s, 请求公司=%s, ERP公司=%s",
			name, companyName, erpCompany)
		return nil, gerror.New("无权限修改此支付方式")
	}

	// 7. 构建更新数据
	updateData := g.Map{}

	// 仅在明确传入 enabled 时才更新
	if req.Enabled != nil {
		if req.GetEnabled() {
			updateData["enabled"] = 1
		} else {
			updateData["enabled"] = 0
		}
	}

	// 8. 如果有字段需要更新，则调用 ERP 更新接口
	if len(updateData) > 0 {
		_, err = service.Document().Update(ctx, &erp.ErpReq{
			DocType: erp.DocTypeModeOfPayment,
			Name:    name,
		}, updateData)
		if err != nil {
			return nil, gerror.Wrapf(err, "更新支付方式失败")
		}

		// 9. 记录审计日志
		g.Log().Infof(ctx, "[updateModeOfPayment] 更新成功: name=%s, company=%s, branch=%s, updateData=%v",
			name, req.CompanyAbbr, req.Branch, updateData)
	} else {
		g.Log().Infof(ctx, "[updateModeOfPayment] 未传入任何可更新字段，跳过更新: name=%s", name)
	}

	// 10. 读取更新后的 payment_id（优先使用更新值，否则使用原值）
	finalPaymentID := data.Get("custom_payment_id").String()
	if req.PaymentId != "" {
		finalPaymentID = req.PaymentId
	}

	return &selling.SaveModeOfPaymentResp{
		Name:      name,
		PaymentId: finalPaymentID,
	}, nil
}

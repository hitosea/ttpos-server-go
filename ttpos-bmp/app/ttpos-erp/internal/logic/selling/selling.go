package selling

import (
	"context"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-bmp/app/ttpos-erp/internal/logic/erpnext"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/erp"
	dtoSelling "ttpos-bmp/app/ttpos-erp/internal/model/dto/selling"
	"ttpos-bmp/app/ttpos-erp/internal/model/dto/setup"

	"ttpos-bmp/app/ttpos-erp/internal/service"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
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
func (s *sSelling) parsePosProfileListResponse(list *g.Var) (*selling.PosProfileListResp, error) {
	j, err := gjson.DecodeToJson(list.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS配置文件列表响应失败")
	}

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
	j.Get("data").Scan(&modePayment)
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
		Name:               req.PosProfileName,
		Company:            companyName,
		Warehouse:          req.Warehouse,
		Branch:             req.Branch,
		Currency:           req.Currency,
		WriteOffAccount:    req.WriteOffAccount,
		WriteOffLimit:      req.WriteOffLimit,
		WriteOffCostCenter: req.WriteOffCostCenter,
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
func (s *sSelling) parsePosProfileResponse(resp *g.Var) (*erp.POSProfile, error) {
	posProfile := &erp.POSProfile{}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS配置文件失败")
	}

	j.Get("data").Scan(posProfile)
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
	opening, err := s.IsProfileOpening(ctx, req.PosProfileName, req.CashierEmail)
	if err != nil {
		return nil, gerror.Wrapf(err, "查询POS配置文件是否开帐失败")
	}
	if opening {
		return nil, gerror.New("POS配置文件已开帐")
	}

	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	// 构建开帐信息
	openDetails := s.buildOpeningEntryDetails(req.OpenPosEntryDetail)
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
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosOpeningEntry, openEntry.Name, 1)
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

// buildOpeningEntryDetails 构建开帐明细
// 参数：
//   - details: 开帐明细列表
//
// 返回：
//   - []erp.POSOpeningEntryDetail: 开帐明细列表
func (s *sSelling) buildOpeningEntryDetails(details []*selling.OpenPosEntryDetail) []erp.POSOpeningEntryDetail {
	openDetails := make([]erp.POSOpeningEntryDetail, 0)
	for _, detail := range details {
		openDetails = append(openDetails, erp.POSOpeningEntryDetail{
			ModeOfPayment: detail.ModeOfPayment,
			OpeningAmount: detail.OpeningAmount,
		})
	}
	return openDetails
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
		PeriodStartDate: gtime.New(req.PeriodStartDate).Format("Y-m-d H:i:s"),
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
func (s *sSelling) parseOpeningEntryResponse(resp *g.Var) (*erp.POSOpeningEntry, error) {
	res := &erp.POSOpeningEntry{}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析开帐响应失败")
	}

	gconv.Scan(j.Get("data"), res)
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
	// 构建关帐明细
	closeDetails := s.buildClosingEntryDetails(req.ClosePosEntryDetail)

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
		StartDate:  gtime.New(openEntry.PeriodStartDate),
		EndDate:    gtime.New(req.PeriodEndDate),
		User:       openEntry.User,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取期间发票失败")
	}

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
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosClosingEntry, closeEntry.Name, 1)
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
//   - details: 关帐明细列表
//
// 返回：
//   - []erp.POSPaymentReconciliation: 关帐明细列表
func (s *sSelling) buildClosingEntryDetails(details []*selling.ClosePosEntryDetail) []erp.POSPaymentReconciliation {
	closeDetails := make([]erp.POSPaymentReconciliation, 0)
	for _, detail := range details {
		closeDetails = append(closeDetails, erp.POSPaymentReconciliation{
			ModeOfPayment: detail.ModeOfPayment,
			ClosingAmount: detail.ClosingAmount,
			OpeningAmount: detail.OpeningAmount,
		})
	}
	return closeDetails
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
		PeriodEndDate:         gtime.New(req.PeriodEndDate).Format("Y-m-d H:i:s"),
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
func (s *sSelling) parseClosingEntryResponse(resp *g.Var) (*erp.POSCloseEntry, error) {
	res := &erp.POSCloseEntry{}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析关帐响应失败")
	}

	gconv.Scan(j.Get("data"), res)
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
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypePosInvoice,
	}, &erp.RequestParams{
		Fields: g.ArrayStr{"name", "posting_date", "customer", "grand_total", "is_return", "return_against"},
		Filters: [][]string{{"pos_profile", "=", req.PosProfile},
			{"owner", "=", req.User},
			{"creation", ">=", req.StartDate.Format("Y-m-d H:i:s")}, {"creation", "<=", req.EndDate.Format("Y-m-d H:i:s")}},
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "查询POS发票列表失败")
	}

	// 解析响应数据
	res := make([]dtoSelling.SimplePosInvoice, 0)
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS发票列表响应失败")
	}

	err = gconv.Scan(j.Get("data"), &res)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS发票数据列表响应失败")
	}

	return res, nil
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
	// 获取公司名称
	companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
	if err != nil {
		return nil, gerror.Wrapf(err, "根据公司缩写查询公司名称失败")
	}

	// 获取开帐记录
	openingEntry, err := s.GetPosOpeningEntry(ctx, req.OpenPosEntryName)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取POS开帐记录失败")
	}

	// 构建POS发票
	posInvoice := s.buildPosInvoice(req, companyName, openingEntry.PosProfile)
	//反结账后重新结账
	if len(req.AmendedProductsInvoiceName) > 0 {
		posInvoice.AmendedFrom = req.AmendedProductsInvoiceName
	}
	// 创建POS发票
	// 特殊处理，使用开帐收银员创建发票
	//创建商品销售记录
	resp, err := service.Document().Create(erpnext.SetFakeUser(ctx, openingEntry.User), erp.DocTypePosInvoice, posInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建POS发票失败")
	}

	// 解析响应数据
	res := &selling.SavePosInvoiceResp{}
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS发票响应失败")
	}

	res.ProductsInvoiceName = j.Get("data.name").String()

	// 提交发票记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosInvoice, res.ProductsInvoiceName, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交发票记录失败")
	}

	//创建物品销售记录
	posRmInvoice := s.buildPosInvoice(req, companyName, openingEntry.PosProfile)
	posRmInvoice.Items = s.buildInvoiceItems(req.MaterialItems)
	//移除支付信息
	posRmInvoicePayments := make([]erp.POSInvoicePayment, 0)
	for _, payment := range posRmInvoice.Payments {
		posRmInvoicePayments = append(posRmInvoicePayments, erp.POSInvoicePayment{
			ModeOfPayment: payment.ModeOfPayment,
			Amount:        0,
		})
	}
	posRmInvoice.Payments = posRmInvoicePayments
	posRmInvoice.Taxes = nil
	//反结账后重新结账
	if len(req.AmendedMaterialInvoiceName) > 0 {
		posRmInvoice.AmendedFrom = req.AmendedMaterialInvoiceName
	}

	//创建物品销售记录
	resp, err = service.Document().Create(erpnext.SetFakeUser(ctx, openingEntry.User), erp.DocTypePosInvoice, posRmInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建物品POS发票失败")
	}

	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS发票响应失败")
	}

	res.MaterialInvoiceName = j.Get("data.name").String()

	// 提交发票记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosInvoice, res.MaterialInvoiceName, 1)
	if err != nil {
		return nil, gerror.Wrapf(err, "提交发票记录失败")
	}
	return res, nil
}

// buildPosInvoice 构建POS发票
// 参数：
//   - req: 保存POS发票请求参数
//   - companyName: 公司名称
//   - posProfile: POS配置文件名称
//
// 返回：
//   - *erp.POSInvoice: POS发票信息
func (s *sSelling) buildPosInvoice(req *selling.SavePosInvoiceReq, companyName string, posProfile string) *erp.POSInvoice {
	postingDatetime := gtime.New(req.PostingDatetime)
	posInvoice := &erp.POSInvoice{
		PosProfile:        posProfile,
		Company:           companyName,
		PostingDate:       postingDatetime.Format(DateFormat),
		PostingTime:       postingDatetime.Format(TimeFormat),
		Currency:          req.Currency,
		PriceListCurrency: req.PriceListCurrency,
		UpdateStock:       req.UpdateStock, // 更新库存
		CustomerOrder:     req.OrderNo,
	}

	// 设置客户信息
	if len(req.CustomerUuid) > 0 {
		posInvoice.Customer = "Member"
		posInvoice.CustomerUUID = req.CustomerUuid
	}

	// 构建发票项目
	posInvoice.Items = s.buildInvoiceItems(req.Items)

	// 构建发票税费
	posInvoice.Taxes = s.buildInvoiceTaxes(req.Taxes, req.CompanyAbbr)

	// 构建支付信息
	posInvoice.Payments = s.buildInvoicePayments(req.Payments)

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
		invoiceItems = append(invoiceItems, erp.POSInvoiceItem{
			ItemCode:    item.ItemCode,
			Qty:         item.Qty,
			Rate:        item.Rate,
			Amount:      item.Amount,
			Description: item.Description,
		})
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
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析POS开帐记录响应失败")
	}

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
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析销售退款订单响应失败")
	}
	saleInvoice := &erp.POSInvoice{}
	// 解析响应数据
	err = j.Get("data").Scan(saleInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "解析原销售订单响应失败")
	}

	postingDatetime := gtime.New(req.PostingDatetime) //这里挖坑，没处理时区

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
		PostingDate:       postingDatetime.Format(DateFormat),
		PostingTime:       postingDatetime.Format(TimeFormat),
		UpdateStock:       0,
		ReturnAgainst:     req.InvoiceName,
		IsReturn:          1,
		IsPos:             1,
		GrandTotal:        grandTotal,
		PaidAmount:        grandTotal,
	}

	//创建物品销售记录
	resp, err = service.Document().Create(erpnext.SetFakeUser(ctx, openingEntry.User), erp.DocTypePosInvoice, returnInvoice)
	if err != nil {
		return nil, gerror.Wrapf(err, "创建销售退款订单发票失败")
	}
	// 解析响应数据
	j, err = gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析销售退款订单发票响应失败")
	}

	res := &selling.ReturnPosInvoiceResp{
		InvoiceName: j.Get("data.name").String(),
	}

	// 提交发票记录
	_, err = service.Document().ChangeDocStatus(ctx, erp.DocTypePosInvoice, res.InvoiceName, 1)
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
//   - 获取支付方式列表
func (*sSelling) GetModeOfPaymentList(ctx context.Context, req *selling.GetModeOfPaymentListReq) (*selling.GetModeOfPaymentListResp, error) {
	filters := make([][]string, 0)
	if len(req.CompanyAbbr) > 0 {
		companyName, err := service.Company().GetCompanyNameWithAbbr(ctx, req.CompanyAbbr)
		if err != nil {
			return nil, gerror.Wrapf(err, "获取公司[%s]失败", req.CompanyAbbr)
		}
		filters = append(filters, []string{"company", "=", companyName})
	}
	if len(req.Branch) > 0 {
		filters = append(filters, []string{"branch", "=", req.Branch})
	}
	//只返回启用的
	filters = append(filters, []string{"enabled", "=", "1"})
	resp, err := service.Document().List(ctx, &erp.ErpReq{
		DocType: erp.DocTypeModeOfPayment,
	}, &erp.RequestParams{
		Fields:  []string{"name"},
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "获取支付方式列表失败")
	}
	// 解析响应数据
	j, err := gjson.DecodeToJson(resp.Bytes())
	if err != nil {
		return nil, gerror.Wrapf(err, "解析支付方式列表响应失败")
	}
	var modeOfPaymentList []*selling.ModeOfPayment
	j.Get("data").Scan(&modeOfPaymentList)
	return &selling.GetModeOfPaymentListResp{
		ModeOfPaymentList: modeOfPaymentList,
	}, nil
}

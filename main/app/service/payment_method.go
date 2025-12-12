package service

import (
	"errors"
	"slices"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	appErrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/saas"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IPaymentMethodSrv 定义支付方式服务接口
type IPaymentMethodSrv interface {
	IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool // 支付方式是否已启用
	GetList(ctx context.Context, typ string) resp.PaymentMethodList                                             // 获取支付方式列表

	// 新增管理方法
	GetManagementList(ctx context.Context, listReq *req.PaymentMethodManagementListReq) (*resp.PaymentMethodListResp, error)
	GetDetail(ctx context.Context, getReq *req.PaymentMethodGetReq) (*resp.PaymentMethodDetailResp, error)
	GetDefaultPayList(ctx context.Context) ([]*resp.DefaultPaymentMethodResp, error) // 获取默认支付方式列表
	Create(ctx context.Context, createReq *req.PaymentMethodCreateReq) error         // 批量创建支付方式
	Update(ctx context.Context, updateReq *req.PaymentMethodUpdateReq) error
	Delete(ctx context.Context, deleteReq *req.PaymentMethodDeleteReq) error
	UpdateSort(ctx context.Context, sortReq *req.PaymentMethodSortUpdateReq) error
	GetLianlianPayConfig(ctx context.Context) (*resp.LianlianPayConfigResp, error)
	UpdateLianlianPayConfig(ctx context.Context, configReq *req.LianlianPayConfigUpdateReq) error
}

// paymentMethodSrv  支付方式服务结构体
type paymentMethodSrv struct {
	dbm        *database.DBManager // 数据库管理器
	settingSrv setting.ISrv
}

// NewPaymentMethodSrv 创建新的收银产品类别服务
func NewPaymentMethodSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
	return NewPaymentMethodSrvImpl(dbm, settingSrv)
}

// NewPaymentMethodSrvImpl 创建新的收银服务实现
func NewPaymentMethodSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
	return &paymentMethodSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// IsEnabled 支付方式是否已启用
func (s *paymentMethodSrv) IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool {
	// 获取支付设置
	paymentSetting, err := s.settingSrv.GetPaymentSetting(ctx, companySetting)
	if err != nil {
		ctx.Log().Error("获取支付设置失败", zap.Error(err))
		return false
	}
	var availableCodes []int
	if paymentSetting.IsBalance == "1" {
		availableCodes = append(availableCodes, constant.PaymentMethodCodeBalance)
	}
	if paymentSetting.IsCash == "1" {
		availableCodes = append(availableCodes, constant.PaymentMethodCodeCash)
	}
	if paymentSetting.IsOther == "1" && paymentMethod.Status == 1 {
		availableCodes = append(availableCodes, paymentMethod.Code)
	}
	return slices.Contains(availableCodes, paymentMethod.Code)
}

// GetList 获取支付方式列表
func (s *paymentMethodSrv) GetList(ctx context.Context, typ string) resp.PaymentMethodList {
	if !slices.Contains([]string{constant.PaymentMethodShowAll, constant.PaymentMethodShowRecharge, constant.PaymentMethodShowCheckout}, typ) {
		return resp.PaymentMethodList{}
	}
	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	companySetting := ctx.GetCompanySetting()
	opts := []repository.DBOption{
		paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable),
	}
	if ctx.GetSource() == constant.SourceCashier {
		if typ != constant.PaymentMethodShowAll {
			switch typ {
			case constant.PaymentMethodShowRecharge:
				opts = append(opts, paymentMethodRepo.WhereCashierMemberRecharge())
			case constant.PaymentMethodShowCheckout:
				opts = append(opts, paymentMethodRepo.WhereCashier())
			}
		}
	} else if ctx.GetSource() == constant.SourceAssistant {
		if typ != constant.PaymentMethodShowAll {
			switch typ {
			case constant.PaymentMethodShowRecharge:
				return resp.PaymentMethodList{}
			case constant.PaymentMethodShowCheckout:
				opts = append(opts, paymentMethodRepo.WhereAssistant())
			}
		}
	}
	opts = append(opts, paymentMethodRepo.WithLogoFile(), paymentMethodRepo.WithQrcodeFile())
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(opts...)

	paymentMethodItems := make([]resp.PaymentMethodItem, 0, len(paymentMethods))
	for _, method := range paymentMethods {
		// 不显示免单
		if method.Code == constant.PaymentMethodCodeFreePay {
			continue
		}
		// 充值不显示余额
		if method.Code == constant.PaymentMethodCodeBalance &&
			(companySetting.IsOpenMember != 1 || typ == constant.PaymentMethodShowRecharge) {
			continue
		}
		var logo, qrcode string
		baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
		if method.LogoFile != nil {
			logo = method.LogoFile.GetUrl(baseUrl)
		}
		if logo == "" && method.DefaultImg != "" {
			logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
		}
		if method.QrcodeFile != nil {
			qrcode = method.QrcodeFile.GetUrl(baseUrl)
		}
		paymentMethodItems = append(paymentMethodItems, resp.PaymentMethodItem{
			SourceText:    i18n.Translate(i18n.GetAcceptLanguage(ctx.GetGin()), constant.PaymentMethodSourceTextMap[method.Source]),
			Uuid:          method.Uuid,
			PaymentName:   method.GetPaymentName(),
			PaymentMethod: method.GetName(),
			FeePercent:    method.FeePercent,
			Logo:          logo,
			Qrcode:        qrcode,
			Code:          method.Code,
			Source:        method.Source,
		})
	}
	return resp.PaymentMethodList{List: paymentMethodItems}
}

// GetManagementList 获取支付方式管理列表（分页）
func (s *paymentMethodSrv) GetManagementList(ctx context.Context, listReq *req.PaymentMethodManagementListReq) (*resp.PaymentMethodListResp, error) {
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	opts := []repository.DBOption{
		paymentMethodRepo.WithLogoFile(),
	}
	if companySetting.IsOpenMember == 0 {
		opts = append(opts, paymentMethodRepo.WhereNotCode([]int{constant.PaymentMethodCodeBalance}))
	}

	// 查询支付方式列表
	list, total, err := paymentMethodRepo.GetPaymentMethodListWithPagination(
		listReq.PageNo,
		listReq.PageSize,
		opts...,
	)
	if err != nil {
		return nil, appErrors.WithMessage(err, "查询支付方式列表失败")
	}

	// 转换为响应格式
	items := make([]*resp.PaymentMethodListItemResp, 0, len(list))
	for _, method := range list {
		var logo string
		baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
		if method.LogoFile != nil {
			logo = method.LogoFile.GetUrl(baseUrl)
		}
		if logo == "" && method.DefaultImg != "" {
			logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
		}

		items = append(items, &resp.PaymentMethodListItemResp{
			Uuid:     method.Uuid,
			Name:     method.Name,
			Source:   method.Source,
			Status:   method.Status,
			Sort:     method.Sort,
			LogoFile: logo,
		})
	}

	return &resp.PaymentMethodListResp{
		List: items,
		Meta: &dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetDetail 获取支付方式详情
func (s *paymentMethodSrv) GetDetail(ctx context.Context, getReq *req.PaymentMethodGetReq) (*resp.PaymentMethodDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 查询支付方式详情，关联文件
	paymentMethod, err := paymentMethodRepo.GetPaymentMethodError(
		paymentMethodRepo.WhereUuid(getReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
		paymentMethodRepo.WithLogoFile(),
		paymentMethodRepo.WithQrcodeFile(),
	)
	if err != nil {
		return nil, appErrors.WithMessage(err, "查询支付方式详情失败")
	}

	// 获取文件 URL
	var logo, qrcode string
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	if paymentMethod.LogoFile != nil {
		logo = paymentMethod.LogoFile.GetUrl(baseUrl)
	}
	if logo == "" && paymentMethod.DefaultImg != "" {
		logo = strings.TrimRight(baseUrl, "/") + paymentMethod.DefaultImg
	}
	if paymentMethod.QrcodeFile != nil {
		qrcode = paymentMethod.QrcodeFile.GetUrl(baseUrl)
	}

	// fee_percent 从 0-1 转换为 0-100
	feePercent := paymentMethod.FeePercent * 100

	return &resp.PaymentMethodDetailResp{
		Uuid:                 paymentMethod.Uuid,
		Name:                 paymentMethod.Name,
		PaymentName:          paymentMethod.PaymentName,
		Source:               paymentMethod.Source,
		LogoFileUuid:         paymentMethod.LogoFileUuid,
		LogoFile:             logo,
		QrcodeFileUuid:       paymentMethod.QrcodeFileUuid,
		QrcodeFile:           qrcode,
		FeePercent:           feePercent,
		IsShowCashier:        paymentMethod.IsShowCashier,
		IsShowAssistant:      paymentMethod.IsShowAssistant,
		IsShowMemberRecharge: paymentMethod.IsShowMemberRecharge,
		Status:               paymentMethod.Status,
		Sort:                 paymentMethod.Sort,
	}, nil
}

// GetDefaultPayList 获取默认支付方式列表（参考PHP defaultPay）
// 返回系统默认的支付方式列表，过滤掉余额、现金、微信、支付宝、POS、免费支付
func (s *paymentMethodSrv) GetDefaultPayList(ctx context.Context) ([]*resp.DefaultPaymentMethodResp, error) {
	// 定义系统默认支付方式列表（参考PHP OrderPayTypeEnum::data()）
	defaultPayments := []struct {
		Value int
		Name  string
		Img   string
		Sort  int
	}{
		{constant.PaymentMethodCodeJACreditCard, "クレジットカード", "/image/pay/ja_pay.png", 1},
		{constant.PaymentMethodCodeJAICTrafficCard, "IC交通卡", "/image/pay/ja_pay.png", 2},
		{constant.PaymentMethodCodeJAQRCode, "QRコード", "/image/pay/ja_pay.png", 3},
		{constant.PaymentMethodCodeQRPromptPay, "QR PromptPay", "/image/pay/qr_prompt_pay.png", 4},
		{constant.PaymentMethodCodeQRCode, "QR Code", "/image/pay/qr_code.png", 5},
		{constant.PaymentMethodCodeSCBEasy, "SCB EASY", "/image/pay/scb_easy.png", 6},
		{constant.PaymentMethodCodeKrungthaiNext, "Krungthai NEXT", "/image/pay/krungthai_next.png", 7},
		{constant.PaymentMethodCodeKrungsriMobile, "Krungsri Mobile", "/image/pay/krungsri_mobile.png", 8},
		{constant.PaymentMethodCodeCrossBorderQR, "Cross-Border QR", "/image/pay/cross_border_qr.png", 9},
		{constant.PaymentMethodCodeTrueMoneyWallet, "TrueMoney", "/image/pay/truemoney_wallet.png", 10},
		{constant.PaymentMethodCodeLINEPay, "LINE Pay", "/image/pay/line_pay.png", 11},
		{constant.PaymentMethodCodeOAliPay, "Alipay", "/image/pay/alipay.png", 12},
		{constant.PaymentMethodCodeOWechat, "WeChat Pay", "/image/pay/wechat_pay.png", 13},
		{constant.PaymentMethodCodeJAQCreditDebit, "Credit/Debit", "/image/pay/ja_pay.png", 14},
	}

	// 过滤掉余额、现金、微信、支付宝、POS、免费支付（参考PHP defaultPay逻辑）
	excludedCodes := []int{
		constant.PaymentMethodCodeBalance,
		constant.PaymentMethodCodeCash,
		constant.PaymentMethodCodeWechat,
		constant.PaymentMethodCodeAliPay,
		constant.PaymentMethodCodePOS,
		constant.PaymentMethodCodeFreePay,
	}

	result := make([]*resp.DefaultPaymentMethodResp, 0)
	for _, payment := range defaultPayments {
		// 排除指定的支付方式
		if slices.Contains(excludedCodes, payment.Value) {
			continue
		}
		baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
		if strings.HasSuffix(baseUrl, "/") {
			baseUrl = baseUrl[:len(baseUrl)-1]
		}
		result = append(result, &resp.DefaultPaymentMethodResp{
			Code: payment.Value,
			Name: payment.Name,
			Url:  baseUrl + payment.Img,
			Img:  payment.Img,
			Sort: payment.Sort,
		})
	}

	return result, nil
}

// generatePaymentCode 生成支付方式 code（手动添加 source=1）
func (s *paymentMethodSrv) generatePaymentCode(db *gorm.DB) int {
	var maxCode int
	db.Model(&model.PaymentMethod{}).Unscoped(). // 包含已删除的
							Where("source = ? AND code >= 20000", constant.PaymentMethodSourceDefault).
							Select("COALESCE(MAX(code), 19900)"). // 如果找不到，返回19900，+100后为20000
							Scan(&maxCode)

	return maxCode + 100 // 每次递增100
}

// Create 批量创建支付方式
func (s *paymentMethodSrv) Create(ctx context.Context, createReq *req.PaymentMethodCreateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 获取最大排序值
	maxSort, err := paymentMethodRepo.GetMaxSort()
	if err != nil {
		return appErrors.WithMessage(err, "获取最大排序值失败")
	}

	// 获取起始code（用于批量创建时递增）
	baseCode := s.generatePaymentCode(db)

	// 批量创建支付方式
	paymentMethods := make([]model.PaymentMethod, 0, len(createReq.Items))
	for i, item := range createReq.Items {
		// 如果指定了code（系统默认支付方式），使用指定的code；否则自动生成
		var code int
		if item.Code > 0 {
			// 使用指定的code，如果多个code一致，递增（参考PHP逻辑：code + key * 100）
			code = item.Code + i*100
		} else {
			// 自动生成code
			code = baseCode + i*100
		}

		source := constant.PaymentMethodSourceDefault // 手动添加为 1

		// fee_percent 从 0-100 转换为 0-1
		feePercent := item.FeePercent / 100

		// 如果logo_file_uuid为0，使用default_img（参考PHP逻辑）
		defaultImg := item.DefaultImg
		if item.LogoFileUuid == 0 && defaultImg == "" {
			// 如果都没有，保持为空
		}

		paymentMethod := model.PaymentMethod{
			BaseModel: model.BaseModel{
				// Uuid 会在 BeforeCreate hook 中自动生成
			},
			Name:                 item.Name,
			Code:                 code,
			PaymentName:          item.PaymentName,
			Source:               source,
			LogoFileUuid:         item.LogoFileUuid,
			QrcodeFileUuid:       item.QrcodeFileUuid,
			DefaultImg:           defaultImg,
			FeePercent:           feePercent,
			IsShowCashier:        item.IsShowCashier,
			IsShowAssistant:      item.IsShowAssistant,
			IsShowMemberRecharge: item.IsShowMemberRecharge,
			Status:               item.Status,
			Sort:                 maxSort + i + 1,
			// CreateTime 和 UpdateTime 由 GORM 自动管理
		}

		paymentMethods = append(paymentMethods, paymentMethod)
	}

	// 批量创建
	for _, paymentMethod := range paymentMethods {
		if err := paymentMethodRepo.CreatePaymentMethod(paymentMethod); err != nil {
			return appErrors.WithMessage(err, "创建支付方式失败")
		}
	}

	return nil
}

// Update 更新支付方式
func (s *paymentMethodSrv) Update(ctx context.Context, updateReq *req.PaymentMethodUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 查询现有支付方式（验证是否存在）
	paymentMethod, err := paymentMethodRepo.GetPaymentMethodError(
		paymentMethodRepo.WhereUuid(updateReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return appErrors.WithMessage(err, "支付方式不存在")
	}

	// 如果是LianLianPay支付（source=2），跳过名称、支付方式、图片logo修改
	isLianLianPay := paymentMethod.Source == constant.PaymentMethodSourceLianLianPay

	// 构建更新数据
	updateData := map[string]any{}
	// LianLianPay支付跳过名称、支付方式、图片logo修改
	if !isLianLianPay {
		updateData["name"] = updateReq.Name
		updateData["payment_name"] = updateReq.PaymentName
		updateData["logo_file_uuid"] = updateReq.LogoFileUuid
	}
	updateData["qrcode_file_uuid"] = updateReq.QrcodeFileUuid
	updateData["fee_percent"] = updateReq.FeePercent / 100
	updateData["is_show_cashier"] = updateReq.IsShowCashier
	updateData["is_show_assistant"] = updateReq.IsShowAssistant
	updateData["is_show_member_recharge"] = updateReq.IsShowMemberRecharge
	updateData["status"] = updateReq.Status

	// 更新支付方式
	if err := paymentMethodRepo.UpdatePaymentMethod(updateData, paymentMethodRepo.WhereUuid(updateReq.Uuid)); err != nil {
		return appErrors.WithMessage(err, "更新支付方式失败")
	}

	return nil
}

// Delete 删除支付方式
func (s *paymentMethodSrv) Delete(ctx context.Context, deleteReq *req.PaymentMethodDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 查询现有支付方式
	paymentMethod, err := paymentMethodRepo.GetPaymentMethodError(
		paymentMethodRepo.WhereUuid(deleteReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return appErrors.WithMessage(err, "支付方式不存在")
	}

	// 仅允许删除自行添加的支付方式（source=1）
	if paymentMethod.Source != constant.PaymentMethodSourceDefault {
		return appErrors.New("仅可删除自行添加的支付方式")
	}

	// 软删除（无需检查关联订单）
	if err := paymentMethodRepo.DeletePaymentMethod(deleteReq.Uuid); err != nil {
		return appErrors.WithMessage(err, "删除支付方式失败")
	}

	// 重新排序，确保排序值连续
	allMethods, _, err := paymentMethodRepo.GetPaymentMethodListWithPagination(1, 10000)
	if err == nil {
		// 重新分配排序值
		items := make([]model.PaymentMethod, 0, len(allMethods))
		for i, method := range allMethods {
			items = append(items, model.PaymentMethod{
				BaseModel: model.BaseModel{
					Uuid: method.Uuid,
				},
				Sort: i + 1,
			})
		}
		if len(items) > 0 {
			_ = paymentMethodRepo.BatchUpdateSort(items)
		}
	}

	return nil
}

// UpdateSort 批量更新支付方式排序
func (s *paymentMethodSrv) UpdateSort(ctx context.Context, sortReq *req.PaymentMethodSortUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 转换为模型列表
	items := make([]model.PaymentMethod, 0, len(sortReq.Items))
	for _, item := range sortReq.Items {
		items = append(items, model.PaymentMethod{
			BaseModel: model.BaseModel{
				Uuid: item.Uuid,
			},
			Sort: item.Sort,
		})
	}

	// 批量更新排序
	if err := paymentMethodRepo.BatchUpdateSort(items); err != nil {
		return appErrors.WithMessage(err, "批量更新排序失败")
	}

	return nil
}

// GetLianlianPayConfig 获取 LianlianPay 配置
func (s *paymentMethodSrv) GetLianlianPayConfig(ctx context.Context) (*resp.LianlianPayConfigResp, error) {
	companyUuid := ctx.GetCompanyUuid()

	// 使用 Repository 方法获取配置
	paymentAppRepo := saas.NewPaymentAppRepo(s.dbm.GetDB(constant.DefaultDB))
	paymentApp, err := paymentAppRepo.GetPaymentAppCompanyUuid(companyUuid)

	if err != nil {
		// 检查是否是记录不存在的错误（Repository 使用 appErrors.WithMessage 包装了错误）
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 配置不存在，返回空配置
			return &resp.LianlianPayConfigResp{}, nil
		}
		return nil, appErrors.WithMessage(err, "查询 LianlianPay 配置失败")
	}

	return &resp.LianlianPayConfigResp{
		LlWhiteIp:            paymentApp.LlWhiteIp,
		LlMerchantId:         paymentApp.LlMerchantId,
		LlStoreId:            paymentApp.LlStoreId,
		LlPublicKey:          paymentApp.LlPublicKey,
		LlMerchantPrivateKey: paymentApp.LlMerchantPrivateKey, // 暂不支持加密存储，直接返回
		LlToken:              paymentApp.LlToken,              // 暂不支持加密存储，直接返回
	}, nil
}

// UpdateLianlianPayConfig 更新 LianlianPay 配置
func (s *paymentMethodSrv) UpdateLianlianPayConfig(ctx context.Context, configReq *req.LianlianPayConfigUpdateReq) error {
	db := s.dbm.GetDB(constant.DefaultDB)
	companyUuid := ctx.GetCompanyUuid()

	// 暂不支持加密存储，直接存储敏感字段

	var paymentApp model.PaymentApp
	err := db.Model(&model.PaymentApp{}).
		Where("company_uuid = ? AND delete_time = 0", companyUuid).
		First(&paymentApp).Error

	if err == gorm.ErrRecordNotFound {
		// 配置不存在，创建新配置
		paymentApp = model.PaymentApp{
			CompanyUuid:          companyUuid,
			LlWhiteIp:            configReq.LlWhiteIp,
			LlMerchantId:         configReq.LlMerchantId,
			LlStoreId:            configReq.LlStoreId,
			LlPublicKey:          configReq.LlPublicKey,
			LlMerchantPrivateKey: configReq.LlMerchantPrivateKey, // 暂不支持加密存储，直接存储
			LlToken:              configReq.LlToken,              // 暂不支持加密存储，直接存储
		}
		if err := db.Create(&paymentApp).Error; err != nil {
			return appErrors.WithMessage(err, "创建 LianlianPay 配置失败")
		}
	} else if err != nil {
		return appErrors.WithMessage(err, "查询 LianlianPay 配置失败")
	} else {
		// 更新现有配置
		updateData := map[string]interface{}{
			"ll_white_ip":             configReq.LlWhiteIp,
			"ll_merchant_id":          configReq.LlMerchantId,
			"ll_store_id":             configReq.LlStoreId,
			"ll_public_key":           configReq.LlPublicKey,
			"ll_merchant_private_key": configReq.LlMerchantPrivateKey, // 暂不支持加密存储，直接存储
			"ll_token":                configReq.LlToken,              // 暂不支持加密存储，直接存储
			// update_time 由 GORM 自动管理
		}
		if err := db.Model(&model.PaymentApp{}).
			Where("company_uuid = ? AND delete_time = 0", companyUuid).
			Updates(updateData).Error; err != nil {
			return appErrors.WithMessage(err, "更新 LianlianPay 配置失败")
		}
	}

	return nil
}

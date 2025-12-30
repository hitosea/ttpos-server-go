package service

import (
	"fmt"
	"slices"
	"strings"
	"ttpos-bmp/app/ttpos-erp/api/selling"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/saas"
	erpService "ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/spf13/viper"
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
	GetDefaultPayList(ctx context.Context) []*resp.DefaultPaymentMethodResp  // 获取默认支付方式列表
	Create(ctx context.Context, createReq *req.PaymentMethodCreateReq) error // 批量创建支付方式
	Update(ctx context.Context, updateReq *req.PaymentMethodUpdateReq) error
	UpdateSort(ctx context.Context, sortReq *req.PaymentMethodSortUpdateReq) error
	GetLianlianPayConfig(ctx context.Context) resp.LianlianPayConfigResp
	UpdateLianlianPayConfig(ctx context.Context, configReq *req.LianlianPayConfigUpdateReq) error
	SyncPaymentMethod(ctx context.Context, syncHeadquarterData bool) error // 同步支付方式
	GetLogo(ctx context.Context, method *model.PaymentMethod) string
	SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error
	SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error
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

	// 连连支付是否可用
	lianLianPayAvailable := true
	if err := NewPaymentRepo(ctx, s.dbm).ValidateConfigError(ctx.GetCompanyUuid()); err != nil {
		lianLianPayAvailable = false
	}

	paymentMethodItems := make([]resp.PaymentMethodItem, 0, len(paymentMethods))
	for _, method := range paymentMethods {
		isAvailable := true
		// 不显示免单
		if method.Code == constant.PaymentMethodCodeFreePay {
			continue
		}
		if method.Code == constant.PaymentMethodCodeFreeMealForErp {
			continue
		}
		// 不显示 Grab 和 LINE MAN 支付方式
		if method.Code == constant.PaymentMethodCodeGrab || method.Code == constant.PaymentMethodCodeLineMan {
			continue
		}
		// 充值不显示余额
		if method.Code == constant.PaymentMethodCodeBalance &&
			(companySetting.IsOpenMember != 1 || typ == constant.PaymentMethodShowRecharge) {
			continue
		}
		// LianLianPay 没有配置支付信息 不显示
		if !lianLianPayAvailable && method.IsLianLianPay() {
			continue
		}
		var logo, qrcode string
		baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
		if method.LogoFile != nil {
			logo = method.LogoFile.GetUrl(baseUrl)
		}
		if logo == "" && method.DefaultImg != "" {
			if strings.HasPrefix(method.DefaultImg, "http") || strings.HasPrefix(method.DefaultImg, "https") {
				logo = method.DefaultImg
			} else {
				logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
			}
		}
		if method.QrcodeFile != nil {
			qrcode = method.QrcodeFile.GetUrl(baseUrl)
		}
		// 总部支付方式
		if method.IsHeadquarterPayment() && method.ErpnextPayment == "" {
			isAvailable = false
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
			IsAvailable:   isAvailable,
		})
	}
	return resp.PaymentMethodList{List: paymentMethodItems}
}

// GetManagementList 获取支付方式管理列表（分页）
func (s *paymentMethodSrv) GetManagementList(ctx context.Context, listReq *req.PaymentMethodManagementListReq) (*resp.PaymentMethodListResp, error) {
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	commonRepo := repository.NewCommonRepo()
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentRepo := NewPaymentRepo(ctx, s.dbm)

	excludeCodes := []int{
		constant.PaymentMethodCodeGrab,
		constant.PaymentMethodCodeLineMan,
		constant.PaymentMethodCodeFreeMealForErp, // 过滤 Free Meal for ERP
	}
	if companySetting.IsOpenMember == 0 {
		excludeCodes = append(excludeCodes, constant.PaymentMethodCodeBalance)
	}
	opts := []repository.DBOption{
		commonRepo.SortWithRaw("CAST(sort AS UNSIGNED), create_time desc"),
		commonRepo.WhereBySoftDelete(),
		paymentMethodRepo.WithLogoFile(),
		paymentMethodRepo.WhereNotCode(excludeCodes),
	}

	// 如果LianLianPay未配置，则过滤LianLianPay支付方式
	if err := paymentRepo.ValidateConfigError(ctx.GetCompanyUuid()); err != nil {
		opts = append(opts, paymentMethodRepo.WhereNotSource(constant.PaymentMethodSourceLianLianPay))
	}

	// 查询支付方式列表
	list, total, err := paymentMethodRepo.GetPaymentMethodListWithPagination(
		listReq.PageNo,
		listReq.PageSize,
		opts...,
	)
	if err != nil {
		return &resp.PaymentMethodListResp{
			List: make([]*resp.PaymentMethodListItemResp, 0),
			Meta: &dto.PageResponse{
				PageNo:   listReq.PageNo,
				PageSize: listReq.PageSize,
				Total:    0,
			},
		}, nil
	}

	// 转换为响应格式
	items := make([]*resp.PaymentMethodListItemResp, 0, len(list))
	for _, method := range list {
		logo := s.GetLogo(ctx, method)

		var status = method.Status
		if method.IsDraft() {
			status = constant.PaymentMethodStatusDraft
		}

		items = append(items, &resp.PaymentMethodListItemResp{
			Uuid:     method.Uuid,
			Name:     method.PaymentName,
			Source:   method.Source,
			Status:   status,
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
		return nil, errors.WithMessage(err, "查询支付方式详情失败")
	}

	// 获取文件 URL
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	logo := s.GetLogo(ctx, paymentMethod)
	var qrcode string
	if paymentMethod.QrcodeFile != nil {
		qrcode = paymentMethod.QrcodeFile.GetUrl(baseUrl)
	}

	// fee_percent 从 0-1 转换为 0-100
	feePercent := paymentMethod.FeePercent * 100

	return &resp.PaymentMethodDetailResp{
		Uuid:                 paymentMethod.Uuid,
		Name:                 paymentMethod.PaymentName,
		PaymentName:          paymentMethod.Name,
		Source:               paymentMethod.Source,
		LogoFileUuid:         paymentMethod.LogoFileUuid,
		LogoFile:             logo,
		QrcodeFileUuid:       paymentMethod.QrcodeFileUuid,
		QrcodeFile:           qrcode,
		FeePercent:           feePercent,
		IsShowCashier:        paymentMethod.IsShowCashier,
		IsShowAssistant:      paymentMethod.IsShowAssistant,
		IsShowKiosk:          paymentMethod.IsShowKiosk,
		IsShowMemberRecharge: paymentMethod.IsShowMemberRecharge,
		Status:               paymentMethod.Status,
		Sort:                 paymentMethod.Sort,
	}, nil
}

// GetDefaultPayList 获取默认支付方式列表（参考PHP defaultPay）
// 返回系统默认的支付方式列表，过滤掉余额、现金、微信、支付宝、POS、免费支付
func (s *paymentMethodSrv) GetDefaultPayList(ctx context.Context) []*resp.DefaultPaymentMethodResp {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	if strings.HasSuffix(baseUrl, "/") {
		baseUrl = baseUrl[:len(baseUrl)-1]
	}

	result := make([]*resp.DefaultPaymentMethodResp, 0)

	// 1. 定义Kbank支付方式列表（在最前面）
	kbankPayments := []struct {
		Code        int
		Name        string
		PaymentName string
		Img         string
		Sort        int
	}{
		{constant.PaymentMethodCodeKbankAlipay, "Alipay", "Alipay", "/image/pay/alipay.png", 0},
		{constant.PaymentMethodCodeKbankWechat, "WeChatPay", "WeChatPay", "/image/pay/wechat_pay.png", 1},
		{constant.PaymentMethodCodeKbankThaiQR, "Thai QR", "Thai QR", "/image/pay/thai_qr.png", 3},
		{constant.PaymentMethodCodeKbankCreditCard, "Credit Card", "Credit Card", "/image/pay/credit_card.png", 4},
	}

	// 2. 查询已添加的Kbank支付方式
	existingKbankPayments := paymentMethodRepo.GetPaymentMethodList(
		paymentMethodRepo.WhereSource(constant.PaymentMethodSourceKbank),
		repository.CommonRepo.WhereBySoftDelete(),
	)

	// 构建已添加的payment_name集合（用于快速查找）
	existingPaymentNames := make(map[int]bool)
	for _, pm := range existingKbankPayments {
		existingPaymentNames[pm.Code] = true
	}

	// 3. 构建Kbank支付方式响应列表
	for _, kp := range kbankPayments {
		canAdd := !existingPaymentNames[kp.Code]
		result = append(result, &resp.DefaultPaymentMethodResp{
			Code:        kp.Code,
			Name:        kp.Name,
			PaymentName: kp.PaymentName,
			Url:         baseUrl + kp.Img,
			Img:         kp.Img,
			Sort:        kp.Sort,
			CanAdd:      canAdd,
			Source:      constant.PaymentMethodSourceKbank,
		})
	}

	// 4. 定义系统默认支付方式列表（参考PHP OrderPayTypeEnum::data()）
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

	// 5. 过滤掉余额、现金、微信、支付宝、POS、免费支付（参考PHP defaultPay逻辑）
	excludedCodes := []int{
		constant.PaymentMethodCodeBalance,
		constant.PaymentMethodCodeCash,
		constant.PaymentMethodCodeWechat,
		constant.PaymentMethodCodeAliPay,
		constant.PaymentMethodCodePOS,
		constant.PaymentMethodCodeFreePay,
	}

	// 6. 添加其他默认支付方式
	for _, payment := range defaultPayments {
		// 排除指定的支付方式
		if slices.Contains(excludedCodes, payment.Value) {
			continue
		}
		result = append(result, &resp.DefaultPaymentMethodResp{
			Code:        payment.Value,
			Name:        payment.Name,
			PaymentName: payment.Name,
			Url:         baseUrl + payment.Img,
			Img:         payment.Img,
			Sort:        payment.Sort,
			CanAdd:      true,                                // 默认支付方式默认可添加
			Source:      constant.PaymentMethodSourceDefault, // 默认source为1
		})
	}

	return result
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

	if len(createReq.Items) == 0 {
		return errors.New("参数错误")
	}

	// 获取最大排序值
	maxSort, err := paymentMethodRepo.GetMaxSort()
	if err != nil {
		return errors.WithMessage(err, "获取最大排序值失败")
	}

	// 获取起始code（用于批量创建时递增）
	baseCode := s.generatePaymentCode(db)

	kbankCodes := []int{
		constant.PaymentMethodCodeKbankAlipay,
		constant.PaymentMethodCodeKbankWechat,
		constant.PaymentMethodCodeKbankCreditQR,
		constant.PaymentMethodCodeKbankThaiQR,
		constant.PaymentMethodCodeKbankCreditCard,
	}

	// 重复检测：检查是否已存在相同的payment_name和source组合
	for _, item := range createReq.Items {
		if item.Name == "" {
			return errors.New("名称不能为空")
		}
		if item.PaymentName == "" {
			return errors.New("支付方式不能为空")
		}
		// 处理source字段：如果传入source=0则使用默认值1，否则使用传入的值
		source := constant.PaymentMethodSourceDefault // 默认值
		if item.Source > 0 {
			source = item.Source
		}
		if source == constant.PaymentMethodSourceKbank {
			if !slices.Contains(kbankCodes, item.Code) {
				return errors.New(fmt.Sprintf("支付方式 %d (%s) 不存在", item.Code, constant.PaymentMethodSourceTextMap[source]))
			}
			existing := paymentMethodRepo.GetPaymentMethod(
				paymentMethodRepo.WherePaymentName(item.PaymentName),
				paymentMethodRepo.WhereCode(item.Code),
				paymentMethodRepo.WhereSource(source),
				repository.CommonRepo.WhereBySoftDelete(),
			)
			if existing.Uuid > 0 {
				return errors.New(fmt.Sprintf("支付方式 %s (%s) 已存在", item.PaymentName, constant.PaymentMethodSourceTextMap[source]))
			}
		}
	}

	// 批量创建支付方式
	paymentMethods := make([]*model.PaymentMethod, 0, len(createReq.Items))
	for i, item := range createReq.Items {
		// 如果指定了code（系统默认支付方式），使用指定的code；否则自动生成
		code := baseCode + i*100
		if item.Code > 0 {
			code = item.Code
		}

		// 处理source字段：如果传入source=0则使用默认值1，否则使用传入的值
		source := constant.PaymentMethodSourceDefault // 默认值
		if item.Source > 0 {
			source = item.Source
		}

		// fee_percent 从 0-100 转换为 0-1
		feePercent := item.FeePercent / 100

		// 如果logo_file_uuid为0，使用default_img（参考PHP逻辑）
		defaultImg := item.DefaultImg
		if item.LogoFileUuid == 0 && defaultImg == "" {
			return errors.New("请上传图片")
		}

		paymentMethod := &model.PaymentMethod{
			Name:                 item.PaymentName,
			Code:                 code,
			PaymentName:          item.Name,
			Source:               source,
			LogoFileUuid:         item.LogoFileUuid,
			QrcodeFileUuid:       item.QrcodeFileUuid,
			DefaultImg:           defaultImg,
			FeePercent:           feePercent,
			IsShowCashier:        item.IsShowCashier,
			IsShowAssistant:      item.IsShowAssistant,
			IsShowKiosk:          item.IsShowKiosk,
			IsShowMemberRecharge: item.IsShowMemberRecharge,
			Status:               item.Status,
			Sort:                 maxSort + i + 1,
		}

		paymentMethods = append(paymentMethods, paymentMethod)
	}

	// 批量创建
	for _, paymentMethod := range paymentMethods {
		err = db.Transaction(func(tx *gorm.DB) error {
			paymentMethodRepo := repository.NewPaymentMethodRepo(tx)
			if err := paymentMethodRepo.CreatePaymentMethodReturnRow(paymentMethod); err != nil {
				return err
			}

			// 如果开启了 ERP，同步支付方式到 ERP
			if ctx.GetCompany().IsOpenErp() && paymentMethod.ErpnextPayment == "" {
				erpSrv := erpService.NewIErpSrv(s.dbm)

				// 根据 source 确定 channel
				channel := erpService.GetChannelBySource(paymentMethod.Source)

				enabled := paymentMethod.Status == 1

				saveModeOfPaymentResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
					CompanyUuid: ctx.GetCompanyUuid(),
					Channel:     channel,
					PayType:     paymentMethod.PaymentName,
					Enabled:     &enabled,
				})
				if err != nil || saveModeOfPaymentResp == nil {
					return err
				}
				// 保存 Name 到 erpnext_payment，PaymentId 到 erpnext_payment_id
				updateData := make(map[string]any)
				if saveModeOfPaymentResp.Name != "" {
					updateData["erpnext_payment"] = saveModeOfPaymentResp.Name
				}
				if saveModeOfPaymentResp.PaymentId != "" {
					updateData["erpnext_payment_id"] = saveModeOfPaymentResp.PaymentId
				}
				if len(updateData) > 0 {
					if err := paymentMethodRepo.UpdatePaymentMethod(
						updateData,
						repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
					); err != nil {
						return err
					}
				}
			}
			return nil
		})

		if err != nil {
			logger.Logger.Error("批量创建支付方式-UpdatePaymentMethod",
				zap.Error(err),
				zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
				zap.Any("paymentMethod", paymentMethod))
			return errors.WithMessage(err, "创建支付方式失败")
		}
	}

	return nil
}

// Update 更新支付方式
func (s *paymentMethodSrv) Update(ctx context.Context, updateReq *req.PaymentMethodUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	if updateReq.Uuid == 0 {
		return errors.New("支付方式不存在")
	}

	// 查询现有支付方式（验证是否存在）
	paymentMethod, err := paymentMethodRepo.GetPaymentMethodError(
		paymentMethodRepo.WhereUuid(updateReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(err, "支付方式不存在")
	}

	if updateReq.Name == "" {
		return errors.New("名称不能为空")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		paymentMethodRepo := repository.NewPaymentMethodRepo(tx)
		isLianLianPay := paymentMethod.Source == constant.PaymentMethodSourceLianLianPay
		isKbank := paymentMethod.Source == constant.PaymentMethodSourceKbank

		name := strings.TrimSpace(updateReq.Name)

		// 构建更新数据
		updateData := map[string]any{}
		// LianLianPay支付、Kbank支付跳过名称、图片logo修改
		if !isLianLianPay && !isKbank {
			updateData["payment_name"] = name
			updateData["logo_file_uuid"] = updateReq.LogoFileUuid
		}
		updateData["qrcode_file_uuid"] = updateReq.QrcodeFileUuid
		updateData["fee_percent"] = updateReq.FeePercent / 100
		updateData["is_show_cashier"] = updateReq.IsShowCashier
		updateData["is_show_assistant"] = updateReq.IsShowAssistant
		updateData["is_show_kiosk"] = updateReq.IsShowKiosk
		updateData["is_show_member_recharge"] = updateReq.IsShowMemberRecharge
		updateData["status"] = updateReq.Status

		// 更新支付方式
		if err := paymentMethodRepo.UpdatePaymentMethod(updateData, paymentMethodRepo.WhereUuid(updateReq.Uuid)); err != nil {
			return errors.WithMessage(err, "更新支付方式失败")
		}

		if ctx.GetCompany().IsOpenErp() {
			// 创建
			channel := erpService.GetChannelBySource(paymentMethod.Source)
			erpSrv := erpService.NewIErpSrv(s.dbm)
			if paymentMethod.ErpnextPayment == "" && paymentMethod.ErpnextPaymentId == "" {
				var erpErr error
				enabled := updateReq.Status == 1
				saveModeOfPaymentResp, erpErr := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
					CompanyUuid: ctx.GetCompanyUuid(),
					Channel:     channel,
					PayType:     paymentMethod.PaymentName,
					Enabled:     &enabled,
				})
				if erpErr != nil {
					return erpErr
				} else if saveModeOfPaymentResp != nil {
					// 保存 Name 到 erpnext_payment，PaymentId 到 erpnext_payment_id
					updateData := make(map[string]any)
					if saveModeOfPaymentResp.Name != "" {
						updateData["erpnext_payment"] = saveModeOfPaymentResp.Name
					}
					if saveModeOfPaymentResp.PaymentId != "" {
						updateData["erpnext_payment_id"] = saveModeOfPaymentResp.PaymentId
					}
					if len(updateData) > 0 {
						if err := paymentMethodRepo.UpdatePaymentMethod(
							updateData,
							repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
						); err != nil {
							return err
						}
					}
				}
			} else {
				// 更新ERPNext支付方式状态
				isUpdateStatus := updateReq.Status != paymentMethod.Status
				if isUpdateStatus {
					enabled := updateReq.Status == 1
					var erpErr error
					// 更新时，如果PaymentId不为空，则传PaymentId，否则传Name
					saveReq := req.SaveModeOfPaymentReq{
						CompanyUuid: ctx.GetCompanyUuid(),
						Channel:     channel,
						PayType:     paymentMethod.PaymentName,
						Enabled:     &enabled,
					}
					if paymentMethod.ErpnextPaymentId != "" {
						// 优先使用 PaymentId
						saveReq.PaymentId = &paymentMethod.ErpnextPaymentId
					} else if paymentMethod.ErpnextPayment != "" {
						// 否则使用 Name
						saveReq.Name = &paymentMethod.ErpnextPayment
					}
					_, erpErr = erpSrv.SaveModeOfPayment(ctx, saveReq)
					if erpErr != nil {
						return erpErr
					}
				}
			}

		}
		return nil
	})
	if err != nil {
		logger.Logger.Error("Update-Transaction", zap.Any("payment_method", paymentMethod), zap.Error(err))
		return errors.WithMessage(err, "更新支付方式失败")
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
		return errors.WithMessage(err, "批量更新排序失败")
	}

	return nil
}

// GetLianlianPayConfig 获取 LianlianPay 配置
func (s *paymentMethodSrv) GetLianlianPayConfig(ctx context.Context) resp.LianlianPayConfigResp {
	companyUuid := ctx.GetCompanyUuid()

	llWhiteIp := viper.GetString("PAY_SERVICE_IP")

	// 使用 Repository 方法获取配置
	paymentAppRepo := saas.NewPaymentAppRepo(s.dbm.GetDB(constant.DefaultDB))
	paymentApp, err := paymentAppRepo.GetPaymentAppCompanyUuid(companyUuid)

	if err != nil || paymentApp == nil {
		return resp.LianlianPayConfigResp{
			LlWhiteIp: llWhiteIp,
		}
	}

	return resp.LianlianPayConfigResp{
		LlWhiteIp:            llWhiteIp,
		LlMerchantId:         paymentApp.LlMerchantId,
		LlStoreId:            paymentApp.LlStoreId,
		LlPublicKey:          paymentApp.LlPublicKey,
		LlMerchantPrivateKey: paymentApp.LlMerchantPrivateKey,
		LlToken:              paymentApp.LlToken,
	}
}

// UpdateLianlianPayConfig 更新 LianlianPay 配置
func (s *paymentMethodSrv) UpdateLianlianPayConfig(ctx context.Context, configReq *req.LianlianPayConfigUpdateReq) error {
	db := s.dbm.GetDB(constant.DefaultDB)
	paymentAppRepo := saas.NewPaymentAppRepo(db)
	companyUuid := ctx.GetCompanyUuid()

	paymentSrv := NewPaymentRepo(ctx, s.dbm)
	signSalt, err := paymentSrv.GetSignSalt(GetSignSaltParams{
		LlWhiteIp:            configReq.LlWhiteIp,
		LlMerchantId:         configReq.LlMerchantId,
		LlPublicKey:          configReq.LlPublicKey,
		LlMerchantPrivateKey: configReq.LlMerchantPrivateKey,
		LlToken:              configReq.LlToken,
		LlStoreId:            configReq.LlStoreId,
		ShopSupplierId:       companyUuid,
	})

	if err != nil {
		return errors.WithMessage(err, "更新支付配置失败")
	}

	_, err = paymentAppRepo.GetPaymentAppCompanyUuid(companyUuid)

	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// 配置不存在，创建新配置
			if err := db.Create(&model.PaymentApp{
				CompanyUuid:          companyUuid,
				LlWhiteIp:            configReq.LlWhiteIp,
				LlMerchantId:         configReq.LlMerchantId,
				LlStoreId:            configReq.LlStoreId,
				LlPublicKey:          configReq.LlPublicKey,
				LlMerchantPrivateKey: configReq.LlMerchantPrivateKey,
				LlToken:              configReq.LlToken,
				LlSignSalt:           signSalt,
			}).Error; err != nil {
				return errors.WithMessage(err, "更新支付配置失败")
			}
		} else {
			return errors.WithMessage(err, "更新支付配置失败")
		}
	} else {
		// 更新现有配置
		updateData := map[string]any{
			"ll_white_ip":             configReq.LlWhiteIp,
			"ll_merchant_id":          configReq.LlMerchantId,
			"ll_store_id":             configReq.LlStoreId,
			"ll_public_key":           configReq.LlPublicKey,
			"ll_merchant_private_key": configReq.LlMerchantPrivateKey,
			"ll_token":                configReq.LlToken,
			"ll_sign_salt":            signSalt,
		}
		if err := paymentAppRepo.UpdatePaymentApp(updateData, companyUuid); err != nil {
			return errors.WithMessage(err, "更新支付配置失败")
		}
	}

	// 配置保存成功后，同步支付方式到 ERP
	company := ctx.GetCompany()
	if company.IsOpenErp() && company.CompanySetting != nil && company.CompanySetting.ErpnextSiteCode != "" {
		erpSrv := erpService.NewIErpSrv(s.dbm)

		paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

		// 批量查询 LIANLIANPAY 支付方式（source=2）且未同步到 ERP 的（erpnext_payment 为空）
		paymentMethods := paymentMethodRepo.GetPaymentMethodList(
			repository.CommonRepo.WhereBySource(constant.PaymentMethodSourceLianLianPay),
			paymentMethodRepo.WhereNotExistsErpnextPayment(),
			repository.CommonRepo.WhereBySoftDelete(),
		)

		for _, paymentMethod := range paymentMethods {
			// 根据支付方式的 source 获取 channel
			channel := erpService.GetChannelBySource(paymentMethod.Source)

			// 同步到 ERP
			saveResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
				CompanyUuid: ctx.GetCompanyUuid(),
				Channel:     channel,
				PayType:     paymentMethod.PaymentName,
			})

			if err != nil || saveResp == nil {
				logger.Logger.Error("同步 LIANLIANPAY 支付方式到 ERP 失败",
					zap.Error(err),
					zap.Any("payment_method", paymentMethod),
				)
				continue
			}

			// 保存 Name 到 erpnext_payment，PaymentId 到 erpnext_payment_id
			updateData := make(map[string]any)
			if saveResp.Name != "" {
				updateData["erpnext_payment"] = saveResp.Name
			}
			if saveResp.PaymentId != "" {
				updateData["erpnext_payment_id"] = saveResp.PaymentId
			}
			if len(updateData) > 0 {
				if err := paymentMethodRepo.UpdatePaymentMethod(
					updateData,
					repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
				); err != nil {
					logger.Logger.Error("更新 LIANLIANPAY 支付方式 ERP 信息失败",
						zap.Error(err),
						zap.Any("payment_method", paymentMethod),
					)
				}
			}
		}
	}

	return nil
}

// GetLogo 获取支付方式logo
func (s *paymentMethodSrv) GetLogo(ctx context.Context, method *model.PaymentMethod) string {
	if method == nil {
		return ""
	}
	var logo string
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)

	if method.LogoFile != nil {
		logo = method.LogoFile.GetUrl(baseUrl)
	}
	if logo == "" && method.DefaultImg != "" {
		if strings.HasPrefix(method.DefaultImg, "http") || strings.HasPrefix(method.DefaultImg, "https") {
			logo = method.DefaultImg
		} else {
			logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
		}
	}
	if logo == "" && method.DefaultImg == "" {
		logo = baseUrl + "/image/pay/ja_pay.png"
	}

	return logo
}

// SyncPaymentMethod 同步支付方式
// 1. 所有商户类型（散户、总店、子店）都先从 ERP 同步支付方式
// 2. 如果是子店且 syncHeadquarterData=true，额外再从总店同步支付方式
func (s *paymentMethodSrv) SyncPaymentMethod(ctx context.Context, syncHeadquarterData bool) error {
	companySetting := ctx.GetCompanySetting()

	// 第一步：所有商户都从 ERP 同步支付方式
	if err := s.syncFromERP(ctx); err != nil {
		return errors.WithMessage(err, "从 ERP 同步支付方式失败")
	}

	// 第二步：如果是子店且需要同步总部数据，额外从总店同步支付方式
	if companySetting.IsSubShop() && syncHeadquarterData {
		if err := s.syncFromHeadquarter(ctx); err != nil {
			return errors.WithMessage(err, "子店从总店同步支付方式失败")
		}
	}

	return nil
}

// syncFromERP 从 ERP 同步支付方式（所有商户类型都可用）
func (s *paymentMethodSrv) syncFromERP(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(companySetting.CompanyUuid)

	// 1. 调用 ERP RPC 获取支付方式列表
	erpPayments, err := s.getModeOfPaymentListFromERP(ctx, &companySetting)
	if err != nil {
		return errors.WithMessage(err, "获取 ERP 支付方式列表失败")
	}

	logger.Logger.Info("开始从 ERP 同步支付方式",
		zap.Int("total", len(erpPayments)),
		zap.Uint64("company_uuid", companySetting.CompanyUuid),
		zap.String("company_abbr", companySetting.ErpnextCompanyAbbr))

	// 2. 开启事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Logger.Error("从 ERP 同步支付方式 panic", zap.Any("panic", r))
		}
	}()

	// 3. 遍历 ERP 支付方式
	var createdCount, updatedCount int
	for _, erpPayment := range erpPayments {
		// 4. 检查是否已存在（优先使用 PaymentID 匹配，降级使用 Name 匹配）
		var existPayment model.PaymentMethod
		var err error

		if erpPayment.PaymentId != "" {
			// 优先使用 PaymentID 匹配
			err = tx.Where("erpnext_payment_id = ? AND delete_time = 0", erpPayment.PaymentId).
				First(&existPayment).Error
			logger.Logger.Debug("使用 PaymentID 匹配",
				zap.String("payment_id", erpPayment.PaymentId),
				zap.String("name", erpPayment.Name))
		} else {
			// 降级使用 Name 匹配
			err = tx.Where("erpnext_payment = ? AND delete_time = 0", erpPayment.Name).
				First(&existPayment).Error
			logger.Logger.Info("ERP 未返回 PaymentID，使用 Name 匹配",
				zap.String("name", erpPayment.Name))
		}

		if err == gorm.ErrRecordNotFound {
			// 首次同步：创建新记录
			if err := s.createPaymentFromERP(tx, erpPayment, companySetting.CompanyUuid); err != nil {
				tx.Rollback()
				return errors.WithMessage(err, "创建支付方式失败")
			}
			createdCount++
		} else if err != nil {
			tx.Rollback()
			return errors.WithMessage(err, "查询支付方式失败")
		} else {
			// 后续同步：仅更新状态
			updates := map[string]any{
				"status": 0,
			}
			if erpPayment.Enabled {
				updates["status"] = 1
			}
			if erpPayment.PaymentId != "" {
				updates["erpnext_payment_id"] = erpPayment.PaymentId
			}
			if err := tx.Model(&model.PaymentMethod{}).
				Where("uuid = ?", existPayment.Uuid).
				Updates(updates).Error; err != nil {
				tx.Rollback()
				return errors.WithMessage(err, "更新支付方式状态失败")
			}
			logger.Logger.Info("更新支付方式状态",
				zap.String("name", erpPayment.Name),
				zap.String("payment_id", erpPayment.PaymentId),
				zap.Any("updates", updates),
				zap.Uint64("uuid", existPayment.Uuid))
			updatedCount++
		}
	}

	// 5. 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithMessage(err, "提交事务失败")
	}

	logger.Logger.Info("从 ERP 同步支付方式完成",
		zap.Int("total", len(erpPayments)),
		zap.Int("created", createdCount),
		zap.Int("updated", updatedCount),
		zap.Uint64("company_uuid", companySetting.CompanyUuid))

	return nil
}

// syncFromHeadquarter 子店从总店同步支付方式
func (s *paymentMethodSrv) syncFromHeadquarter(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	headquarterDB := s.dbm.GetDB(companySetting.HeadquarterUuid)
	subShopDB := s.dbm.GetDB(companySetting.CompanyUuid)
	headquarterUuid := companySetting.HeadquarterUuid

	logger.Logger.Info("开始从总店同步支付方式",
		zap.Uint64("sub_shop_uuid", companySetting.CompanyUuid),
		zap.Uint64("headquarter_uuid", headquarterUuid))

	// 查询总部支付方式（排除 code=40 和 code=10）
	var hqPayments []model.PaymentMethod
	err := headquarterDB.Where("delete_time = 0 AND headquarter_uuid = 0").
		Where("code NOT IN (?)", []int{
			model.PaymentMethodCash,
			model.PaymentMethodBalance,
			model.PaymentCodeLianlianWechat,
			model.PaymentCodeLianlianAli,
			model.PaymentCodeLianlianQrPromptPay}).
		Find(&hqPayments).Error
	if err != nil {
		return errors.WithMessage(err, "查询总部支付方式失败")
	}

	var createdCount, updatedCount, skippedCount int
	for _, hqPayment := range hqPayments {
		// 检查分店是否已有同名支付方式（payment_name）
		var existPayment model.PaymentMethod
		err := subShopDB.Where("payment_name = ? and source = ? AND delete_time = 0", hqPayment.PaymentName, model.PaymentSourceDefault).
			First(&existPayment).Error

		if err == nil {
			logger.Logger.Info("支付方式已存在，跳过同步",
				zap.String("name", hqPayment.PaymentName),
				zap.Int("code", existPayment.Code))
			skippedCount++
			continue
		}

		// 分店不存在，创建新支付方式
		newCode := s.generatePaymentCode(subShopDB)

		newPayment := model.PaymentMethod{
			HeadquarterUuid: headquarterUuid,
			PaymentName:     hqPayment.PaymentName,
			Name:            hqPayment.Name,
			Code:            newCode,                    // 生成新code
			Source:          model.PaymentSourceDefault, // 1-手动添加
			LogoFileUuid:    0,                          // 固定为0
			Sort:            hqPayment.Sort,             // 排序
			Status:          hqPayment.Status,           // 同步状态
			// 其他字段使用数据库默认值
		}

		err = subShopDB.Create(&newPayment).Error
		if err != nil {
			logger.Logger.Error("创建支付方式失败",
				zap.String("name", hqPayment.PaymentName),
				zap.Error(err))
			continue
		}

		logger.Logger.Info("从总店创建新支付方式",
			zap.String("name", hqPayment.PaymentName),
			zap.Int("code", newCode))
		createdCount++
	}

	logger.Logger.Info("从总店同步支付方式完成",
		zap.Int("total", len(hqPayments)),
		zap.Int("created", createdCount),
		zap.Int("updated", updatedCount),
		zap.Int("skipped", skippedCount))

	return nil
}

// getModeOfPaymentListFromERP 从 ERP 获取支付方式列表
func (s *paymentMethodSrv) getModeOfPaymentListFromERP(ctx context.Context, companySetting *model.CompanySetting) ([]*selling.ModeOfPayment, error) {
	client, conn, err := erpService.NewErpSellingClient()
	if err != nil {
		return nil, errors.WithMessage(err, "创建 ERP Selling 客户端失败")
	}
	defer conn.Close()

	req := &selling.GetModeOfPaymentListReq{
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	}

	result, err := client.GetModeOfPaymentList(erpService.WithSiteCode(ctx.GetContext(), companySetting.ErpnextSiteCode), req)
	if err != nil {
		return nil, errors.WithMessage(err, "调用 ERP GetModeOfPaymentList 失败")
	}

	if result.GetCode() != "0" || result.Data == nil {
		logger.Logger.Error("GetModeOfPaymentList 返回错误",
			zap.String("code", result.GetCode()),
			zap.String("message", result.GetMessage()))
		return nil, errors.New("获取 ERP 支付方式失败: " + result.GetMessage())
	}

	response := &selling.GetModeOfPaymentListResp{}
	if err := result.Data.UnmarshalTo(response); err != nil {
		return nil, errors.WithMessage(err, "解析 ERP 响应失败")
	}

	return response.ModeOfPaymentList, nil
}

// createPaymentFromERP 从 ERP 数据创建支付方式（首次同步）
func (s *paymentMethodSrv) createPaymentFromERP(
	tx *gorm.DB,
	erpPayment *selling.ModeOfPayment,
	_ uint64,
) error {
	// 1. 生成 code（使用 generatePaymentCode，从 20000 开始，每次递增 100）
	code := s.generatePaymentCode(tx)

	// 2. 确定状态
	status := 0
	if erpPayment.Enabled {
		status = 1
	}

	// 3. 创建支付方式
	newPayment := model.PaymentMethod{
		Name:                 erpPayment.Name,
		Code:                 code,
		PaymentName:          erpPayment.Name,
		Source:               model.PaymentSourceDefault, // 1=手动添加
		LogoFileUuid:         0,                          // 使用默认图标
		QrcodeFileUuid:       0,
		FeePercent:           0.0000,
		IsShowCashier:        1, // 收银机显示
		IsShowAssistant:      1, // 点餐助手显示
		IsShowKiosk:          0, // 自助机不显示
		IsShowMemberRecharge: 1, // 充值显示
		Status:               status,
		Sort:                 0,
		DefaultImg:           "/image/pay/ja_pay.png", // 使用默认图标
		ErpnextPayment:       erpPayment.Name,         // 关联 ERP 名称（向后兼容）
		ErpnextPaymentId:     erpPayment.PaymentId,    // 保存 ERP PaymentID（优先关联字段）
		HeadquarterUuid:      0,                       // 散户/总店/子店=0
	}
	// Uuid 会在 BeforeCreate Hook 中自动生成

	if err := tx.Create(&newPayment).Error; err != nil {
		return errors.WithMessage(err, "插入支付方式失败")
	}

	logger.Logger.Info("创建支付方式成功",
		zap.String("name", erpPayment.Name),
		zap.String("payment_id", erpPayment.PaymentId),
		zap.Int("code", code),
		zap.Int("status", status),
		zap.Uint64("uuid", newPayment.Uuid))

	return nil
}

// SaveGrabPaymentMethod 保存 Grab 支付方式，如果已存在则跳过（幂等性）
// 供外部服务调用，支持事务
func (s *paymentMethodSrv) SaveGrabPaymentMethod(ctx context.Context, tx *gorm.DB) error {
	paymentMethodRepo := repository.NewPaymentMethodRepo(tx)

	// 检查支付方式是否已存在（通过 payment_name 或 code）
	existPayment := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WherePaymentName(constant.PaymentMethodNameGrab),
	)
	if existPayment.Uuid > 0 {
		// 已存在，跳过
		return nil
	}

	// 如果通过 payment_name 没找到，再通过 code 查询
	existPayment = paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeGrab),
	)
	if existPayment.Uuid > 0 {
		// 已存在，跳过
		return nil
	}

	// 获取最大排序值
	maxSort, err := paymentMethodRepo.GetMaxSort()
	if err != nil {
		logger.Logger.Error("获取最大排序值失败", zap.Error(err))
		return errors.WithMessage(err, "获取最大排序值失败")
	}

	// 创建 Grab 支付方式
	paymentMethod := &model.PaymentMethod{
		Name:                 constant.PaymentMethodNameGrab,
		Code:                 constant.PaymentMethodCodeGrab,
		PaymentName:          constant.PaymentMethodNameGrab,
		Source:               constant.PaymentMethodSourceSystem, // 0=系统默认
		LogoFileUuid:         0,
		QrcodeFileUuid:       0,
		FeePercent:           0.0000,
		IsShowCashier:        0, // 不在收银端显示
		IsShowAssistant:      0,
		IsShowKiosk:          0,
		IsShowMemberRecharge: 0,
		Status:               constant.PaymentMethodStatusEnable,
		Sort:                 maxSort + 1,
		DefaultImg:           "", // 空字符串
		ErpnextPayment:       "",
		HeadquarterUuid:      0,
	}

	if err := paymentMethodRepo.CreatePaymentMethodReturnRow(paymentMethod); err != nil {
		logger.Logger.Error("创建 Grab 支付方式失败", zap.Error(err))
		return errors.WithMessage(err, "创建 Grab 支付方式失败")
	}

	// 如果开启了 ERP，同步支付方式到 ERP
	if ctx.GetCompany().IsOpenErp() {
		erpSrv := erpService.NewIErpSrv(s.dbm)

		// 根据 source 确定 channel
		channel := erpService.GetChannelBySource(paymentMethod.Source)

		saveModeOfPaymentResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
			CompanyUuid: ctx.GetCompanyUuid(),
			Channel:     channel,
			PayType:     paymentMethod.PaymentName,
		})
		if err != nil || saveModeOfPaymentResp == nil {
			logger.Logger.Error("同步 Grab 支付方式到 ERP 失败", zap.Error(err))
			return errors.WithMessage(err, "创建 Grab 支付方式失败")
		}
		if saveModeOfPaymentResp.Name != "" {
			if err := paymentMethodRepo.UpdatePaymentMethod(
				map[string]any{"erpnext_payment": saveModeOfPaymentResp.Name, "erpnext_payment_id": saveModeOfPaymentResp.PaymentId},
				repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
			); err != nil {
				logger.Logger.Error("更新 Grab 支付方式 ERP 名称失败", zap.Error(err))
				return errors.WithMessage(err, "创建 Grab 支付方式失败")
			}
		}
	}

	return nil
}

// SaveLineManPaymentMethod 保存 LINE MAN 支付方式，如果已存在则跳过（幂等性）
// 供外部服务调用，支持事务
func (s *paymentMethodSrv) SaveLineManPaymentMethod(ctx context.Context, tx *gorm.DB) error {
	paymentMethodRepo := repository.NewPaymentMethodRepo(tx)

	// 检查支付方式是否已存在（通过 payment_name 或 code）
	existPayment := paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WherePaymentName(constant.PaymentMethodNameLineMan),
	)
	if existPayment.Uuid > 0 {
		// 已存在，跳过
		return nil
	}

	// 如果通过 payment_name 没找到，再通过 code 查询
	existPayment = paymentMethodRepo.GetPaymentMethod(
		paymentMethodRepo.WhereCode(constant.PaymentMethodCodeLineMan),
	)
	if existPayment.Uuid > 0 {
		// 已存在，跳过
		return nil
	}

	// 获取最大排序值
	maxSort, err := paymentMethodRepo.GetMaxSort()
	if err != nil {
		logger.Logger.Error("获取最大排序值失败", zap.Error(err))
		return errors.WithMessage(err, "获取最大排序值失败")
	}

	// 创建 LINE MAN 支付方式
	paymentMethod := &model.PaymentMethod{
		Name:                 constant.PaymentMethodNameLineMan,
		Code:                 constant.PaymentMethodCodeLineMan,
		PaymentName:          constant.PaymentMethodNameLineMan,
		Source:               constant.PaymentMethodSourceSystem, // 0=系统默认
		LogoFileUuid:         0,
		QrcodeFileUuid:       0,
		FeePercent:           0.0000,
		IsShowCashier:        0, // 不在收银端显示
		IsShowAssistant:      0,
		IsShowKiosk:          0,
		IsShowMemberRecharge: 0,
		Status:               constant.PaymentMethodStatusEnable,
		Sort:                 maxSort + 1,
		DefaultImg:           "", // 空字符串
		ErpnextPayment:       "",
		HeadquarterUuid:      0,
	}

	if err := paymentMethodRepo.CreatePaymentMethodReturnRow(paymentMethod); err != nil {
		logger.Logger.Error("创建 LINE MAN 支付方式失败", zap.Error(err))
		return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
	}

	// 如果开启了 ERP，同步支付方式到 ERP
	if ctx.GetCompany().IsOpenErp() {
		erpSrv := erpService.NewIErpSrv(s.dbm)

		// 根据 source 确定 channel
		channel := erpService.GetChannelBySource(paymentMethod.Source)

		saveModeOfPaymentResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
			CompanyUuid: ctx.GetCompanyUuid(),
			Channel:     channel,
			PayType:     paymentMethod.PaymentName,
		})
		if err != nil || saveModeOfPaymentResp == nil {
			logger.Logger.Error("同步 LINE MAN 支付方式到 ERP 失败", zap.Error(err))
			return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
		}
		if saveModeOfPaymentResp.Name != "" {
			if err := paymentMethodRepo.UpdatePaymentMethod(
				map[string]any{"erpnext_payment": saveModeOfPaymentResp.Name, "erpnext_payment_id": saveModeOfPaymentResp.PaymentId},
				repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
			); err != nil {
				logger.Logger.Error("更新 LINE MAN 支付方式 ERP 名称失败", zap.Error(err))
				return errors.WithMessage(err, "创建 LINE MAN 支付方式失败")
			}
		}
	}

	return nil
}

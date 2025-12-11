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
	Create(ctx context.Context, createReq *req.PaymentMethodCreateReq) error
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
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	list, total, err := paymentMethodRepo.GetPaymentMethodListWithPagination(listReq.PageNo, listReq.PageSize)
	if err != nil {
		return nil, appErrors.WithMessage(err, "查询支付方式列表失败")
	}

	// 转换为响应格式
	items := make([]*resp.PaymentMethodListItemResp, 0, len(list))
	for _, method := range list {
		items = append(items, &resp.PaymentMethodListItemResp{
			Uuid:   method.Uuid,
			Name:   method.Name,
			Source: method.Source,
			Status: method.Status,
			Sort:   method.Sort,
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
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	var logoFile, qrcodeFile string
	if paymentMethod.LogoFile != nil {
		logoFile = paymentMethod.LogoFile.GetUrl(baseUrl)
	}
	if paymentMethod.QrcodeFile != nil {
		qrcodeFile = paymentMethod.QrcodeFile.GetUrl(baseUrl)
	}

	// fee_percent 从 0-1 转换为 0-100
	feePercent := paymentMethod.FeePercent * 100

	return &resp.PaymentMethodDetailResp{
		Uuid:                 paymentMethod.Uuid,
		Name:                 paymentMethod.Name,
		PaymentName:          paymentMethod.PaymentName,
		Source:               paymentMethod.Source,
		LogoFileUuid:         paymentMethod.LogoFileUuid,
		LogoFile:             logoFile,
		QrcodeFileUuid:       paymentMethod.QrcodeFileUuid,
		QrcodeFile:           qrcodeFile,
		FeePercent:           feePercent,
		IsShowCashier:        paymentMethod.IsShowCashier,
		IsShowAssistant:      paymentMethod.IsShowAssistant,
		IsShowMemberRecharge: paymentMethod.IsShowMemberRecharge,
		Status:               paymentMethod.Status,
		Sort:                 paymentMethod.Sort,
	}, nil
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

// Create 创建支付方式
func (s *paymentMethodSrv) Create(ctx context.Context, createReq *req.PaymentMethodCreateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 生成 code 和设置 source
	code := s.generatePaymentCode(db)
	source := constant.PaymentMethodSourceDefault // 手动添加为 1

	// 获取最大排序值
	maxSort, err := paymentMethodRepo.GetMaxSort()
	if err != nil {
		return appErrors.WithMessage(err, "获取最大排序值失败")
	}

	// fee_percent 从 0-100 转换为 0-1
	feePercent := createReq.FeePercent / 100

	// 创建支付方式
	paymentMethod := model.PaymentMethod{
		BaseModel: model.BaseModel{
			// Uuid 会在 BeforeCreate hook 中自动生成
		},
		Name:                 createReq.Name,
		Code:                 code,
		PaymentName:          createReq.PaymentName,
		Source:               source,
		LogoFileUuid:         createReq.LogoFileUuid,
		QrcodeFileUuid:       createReq.QrcodeFileUuid,
		FeePercent:           feePercent,
		IsShowCashier:        createReq.IsShowCashier,
		IsShowAssistant:      createReq.IsShowAssistant,
		IsShowMemberRecharge: createReq.IsShowMemberRecharge,
		Status:               createReq.Status,
		Sort:                 maxSort + 1,
		// CreateTime 和 UpdateTime 由 GORM 自动管理
	}

	if err := paymentMethodRepo.CreatePaymentMethod(paymentMethod); err != nil {
		return appErrors.WithMessage(err, "创建支付方式失败")
	}

	return nil
}

// Update 更新支付方式
func (s *paymentMethodSrv) Update(ctx context.Context, updateReq *req.PaymentMethodUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)

	// 查询现有支付方式（验证是否存在）
	_, err := paymentMethodRepo.GetPaymentMethodError(
		paymentMethodRepo.WhereUuid(updateReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return appErrors.WithMessage(err, "支付方式不存在")
	}

	// 任何来源的支付方式都可以编辑

	// 构建更新数据
	updateData := model.PaymentMethod{
		BaseModel: model.BaseModel{
			Uuid: updateReq.Uuid,
		},
	}
	if updateReq.Name != "" {
		updateData.Name = updateReq.Name
	}
	if updateReq.PaymentName != "" {
		updateData.PaymentName = updateReq.PaymentName
	}
	if updateReq.LogoFileUuid > 0 {
		updateData.LogoFileUuid = updateReq.LogoFileUuid
	}
	if updateReq.QrcodeFileUuid > 0 {
		updateData.QrcodeFileUuid = updateReq.QrcodeFileUuid
	}
	if updateReq.FeePercent > 0 {
		// fee_percent 从 0-100 转换为 0-1
		updateData.FeePercent = updateReq.FeePercent / 100
	}
	if updateReq.IsShowCashier >= 0 {
		updateData.IsShowCashier = updateReq.IsShowCashier
	}
	if updateReq.IsShowAssistant >= 0 {
		updateData.IsShowAssistant = updateReq.IsShowAssistant
	}
	if updateReq.IsShowMemberRecharge >= 0 {
		updateData.IsShowMemberRecharge = updateReq.IsShowMemberRecharge
	}
	if updateReq.Status >= 0 {
		updateData.Status = updateReq.Status
	}
	// UpdateTime 由 GORM 自动管理（autoUpdateTime）

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

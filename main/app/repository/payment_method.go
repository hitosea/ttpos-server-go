package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// IPaymentMethodRepo 定义仓库接口
type IPaymentMethodRepo interface {
	IPaymentMethodQueryRepo
	WhereUuid(uuid uint64) DBOption       // uuid条件
	WhereCashier() DBOption               // 在收银端结账显示
	WhereCashierMemberRecharge() DBOption // 收银端充值时显示
	WhereAssistant() DBOption             // 在助手端结账时显示
	WhereStatus(status int) DBOption
	WhereExistsErpnextPayment() DBOption // 排除ERPNext支付方式

	WithLogoFile() DBOption   // 关联logo文件
	WithQrcodeFile() DBOption // 关联二维码文件

	CreatePaymentMethod(paymentMethod model.PaymentMethod) error // 创建支付方式
}

// IPaymentMethodQueryRepo 定义仓库查询接口
type IPaymentMethodQueryRepo interface {
	GetPaymentMethod(opts ...DBOption) model.PaymentMethod
	GetPaymentMethodError(opts ...DBOption) (*model.PaymentMethod, error)
	GetPaymentMethodByUuid(uuid uint64) (*model.PaymentMethod, error)
	GetPaymentMethodList(opts ...DBOption) []*model.PaymentMethod
	GetPaymentMethodsByCtx(ctx context.Context) []*model.PaymentMethod // 获取收银机支付页面的支付方式列表
	GetLianLianPayPaymentMethodList() ([]*model.PaymentMethod, error)  // 查询连连支付的支付方式列表

	InitErpnextPayment(payments map[int]string) error
}

// paymentMethodRepo 仓库
type paymentMethodRepo struct {
	db *gorm.DB
}

// NewPaymentMethodRepo 创建新仓库
func NewPaymentMethodRepo(db *gorm.DB) IPaymentMethodRepo {
	return NewPaymentMethodRepoImpl(db)
}

// NewPaymentMethodRepoImpl 创建新仓库实现
func NewPaymentMethodRepoImpl(db *gorm.DB) IPaymentMethodRepo {
	return &paymentMethodRepo{db: db}
}

func (r *paymentMethodRepo) CreatePaymentMethod(paymentMethod model.PaymentMethod) error {
	paymentMethod.SetNil()
	if err := r.db.Create(&paymentMethod).Error; err != nil {
		return err
	}
	return nil
}

// GetPaymentMethod 获取支付方式
func (r *paymentMethodRepo) GetPaymentMethod(opts ...DBOption) model.PaymentMethod {
	var paymentMethod model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{})
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&paymentMethod)
	return paymentMethod
}

// GetPaymentMethodList  获取支付方式
func (r *paymentMethodRepo) GetPaymentMethodList(opts ...DBOption) []*model.PaymentMethod {
	var paymentMethods []*model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Order("CAST(sort AS UNSIGNED), create_time desc").Find(&paymentMethods)
	return paymentMethods
}

func (r *paymentMethodRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *paymentMethodRepo) WhereCashier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_cashier = 1")
	}
}

func (r *paymentMethodRepo) WhereCashierMemberRecharge() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_member_recharge = 1")
	}
}

func (r *paymentMethodRepo) WhereAssistant() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_assistant = 1")
	}
}

func (r *paymentMethodRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *paymentMethodRepo) WithLogoFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("LogoFile")
	}
}

func (r *paymentMethodRepo) WithQrcodeFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("QrcodeFile")
	}
}

func (r *paymentMethodRepo) GetPaymentMethodError(opts ...DBOption) (*model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&paymentMethod).Error
	return &paymentMethod, errors.WithMessage(err)
}

func (r *paymentMethodRepo) GetPaymentMethodByUuid(uuid uint64) (*model.PaymentMethod, error) {
	paymentMethod, err := r.GetPaymentMethodError(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereByStatus(constant.PaymentMethodStatusEnable),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "LogoFile",
			},
			WithPreload{
				Query: "QrcodeFile",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return paymentMethod, nil
}

func (r *paymentMethodRepo) GetPaymentMethodsByCtx(ctx context.Context) []*model.PaymentMethod {
	opts := []DBOption{
		r.WhereStatus(constant.PaymentMethodStatusEnable),
	}
	if ctx.GetSource() == constant.SourceCashier {
		opts = append(opts, r.WhereCashier())
	} else if ctx.GetSource() == constant.SourceAssistant {
		opts = append(opts, r.WhereAssistant())
	}
	opts = append(opts, r.WithLogoFile(), r.WithQrcodeFile())
	paymentMethods := r.GetPaymentMethodList(opts...)
	return paymentMethods
}

func (r *paymentMethodRepo) GetLianLianPayPaymentMethodList() ([]*model.PaymentMethod, error) {
	opts := []DBOption{
		// CommonRepo.WhereByStatus(constant.PaymentMethodStatusEnable),
		CommonRepo.WhereBySource(constant.PaymentMethodSourceLianLianPay),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "LogoFile",
			},
			WithPreload{
				Query: "QrcodeFile",
			},
		),
	}
	paymentMethods := r.GetPaymentMethodList(opts...)
	// 暂时不返回QR支付方式
	result := make([]*model.PaymentMethod, 0)
	if len(paymentMethods) > 0 {
		for _, paymentMethod := range paymentMethods {
			if paymentMethod.Code != constant.PaymentMethodCodeLianLianQRPromptPay {
				result = append(result, paymentMethod)
			}
		}
	}
	return result, nil
}

func (r *paymentMethodRepo) InitErpnextPayment(payments map[int]string) error {
	// 检查是否有数据需要更新
	if len(payments) == 0 {
		return nil
	}
	var codes []int
	// 构建 CASE WHEN 语句
	caseWhenSQL := "CASE code"
	var args []any
	for code, erpPaymentName := range payments {
		caseWhenSQL += " WHEN ? THEN ?"
		args = append(args, code, erpPaymentName)
		codes = append(codes, code)
	}
	caseWhenSQL += " END"
	// 一条SQL语句批量更新排序
	err := r.db.Model(&model.PaymentMethod{}).
		Where("code IN ?", codes).
		Update("erpnext_payment", gorm.Expr(caseWhenSQL, args...)).Error

	if err != nil {
		return errors.WithMessage(errors.New("更新支付方式失败"), err.Error())
	}
	return nil
}

func (r *paymentMethodRepo) WhereExistsErpnextPayment() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_payment != ''")
	}
}

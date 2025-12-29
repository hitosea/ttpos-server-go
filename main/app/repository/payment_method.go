package repository

import (
	"time"
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
	WhereKiosk() DBOption                 // 在自助点餐机结账时显示
	WhereStatus(status int) DBOption
	WhereExistsErpnextPayment() DBOption          // 存在ERPNext支付方式
	WhereNotExistsErpnextPayment() DBOption       // 不存在ERPNext支付方式
	WhereNotCode(codes []int) DBOption            // 排除支付方式代号
	WherePaymentName(paymentName string) DBOption // 按支付方式名称查询
	WhereCode(code int) DBOption                  // 按支付方式code查询
	WhereSource(source int) DBOption              // 按来源查询

	WithLogoFile() DBOption   // 关联logo文件
	WithQrcodeFile() DBOption // 关联二维码文件

	CreatePaymentMethod(paymentMethod model.PaymentMethod) error                                                      // 创建支付方式
	CreatePaymentMethodReturnRow(paymentMethod *model.PaymentMethod) error                                            // 批量创建支付方式
	UpdatePaymentMethod(date map[string]any, options ...DBOption) error                                               // 更新支付方式
	DeletePaymentMethod(uuid uint64) error                                                                            // 删除支付方式（软删除）
	CheckHasOrders(uuid uint64) (bool, error)                                                                         // 检查是否有关联订单
	GetMaxSort() (int, error)                                                                                         // 获取最大排序值
	BatchUpdateSort(items []model.PaymentMethod) error                                                                // 批量更新排序
	GetPaymentMethodListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.PaymentMethod, int64, error) // 分页查询支付方式列表
}

// IPaymentMethodQueryRepo 定义仓库查询接口
type IPaymentMethodQueryRepo interface {
	GetPaymentMethod(opts ...DBOption) model.PaymentMethod
	GetPaymentMethodError(opts ...DBOption) (*model.PaymentMethod, error)
	GetPaymentMethodByUuid(uuid uint64) (*model.PaymentMethod, error)
	GetPaymentMethodList(opts ...DBOption) []*model.PaymentMethod
	GetAllPaymentMethodList(opts ...DBOption) []*model.PaymentMethod
	GetPaymentMethodsByCtx(ctx context.Context) []*model.PaymentMethod // 获取收银机支付页面的支付方式列表
	GetLianLianPayPaymentMethodList() ([]*model.PaymentMethod, error)  // 查询连连支付的支付方式列表

	InitErpnextPayment(payments map[int]string) error
	InitErpnextPaymentWithId(payments map[int]ErpnextPaymentInfo) error
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

func (r *paymentMethodRepo) CreatePaymentMethodReturnRow(paymentMethod *model.PaymentMethod) error {
	if err := r.db.Create(paymentMethod).Error; err != nil {
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

// GetPaymentMethodList  获取支付方式
func (r *paymentMethodRepo) GetAllPaymentMethodList(opts ...DBOption) []*model.PaymentMethod {
	var paymentMethods []*model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{})
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

func (r *paymentMethodRepo) WhereKiosk() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_kiosk = 1")
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

// ErpnextPaymentInfo ERP支付方式信息
type ErpnextPaymentInfo struct {
	Name      string // ERP支付方式名称
	PaymentId string // ERP支付方式ID
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

// InitErpnextPaymentWithId 批量初始化ERP支付方式（同时保存Name和PaymentId）
func (r *paymentMethodRepo) InitErpnextPaymentWithId(payments map[int]ErpnextPaymentInfo) error {
	// 检查是否有数据需要更新
	if len(payments) == 0 {
		return nil
	}
	var codes []int
	// 构建 erpnext_payment 的 CASE WHEN 语句
	caseWhenSQLName := "CASE code"
	var argsName []any
	// 构建 erpnext_payment_id 的 CASE WHEN 语句
	caseWhenSQLId := "CASE code"
	var argsId []any

	for code, info := range payments {
		caseWhenSQLName += " WHEN ? THEN ?"
		argsName = append(argsName, code, info.Name)
		caseWhenSQLId += " WHEN ? THEN ?"
		argsId = append(argsId, code, info.PaymentId)
		codes = append(codes, code)
	}
	caseWhenSQLName += " END"
	caseWhenSQLId += " END"

	// 一条SQL语句批量更新两个字段
	err := r.db.Model(&model.PaymentMethod{}).
		Where("code IN ?", codes).
		Updates(map[string]any{
			"erpnext_payment":    gorm.Expr(caseWhenSQLName, argsName...),
			"erpnext_payment_id": gorm.Expr(caseWhenSQLId, argsId...),
		}).Error

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

func (r *paymentMethodRepo) WhereNotExistsErpnextPayment() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// 未同步到ERP：erpnext_payment 为空 且 erpnext_payment_id 为空
		return db.Where("erpnext_payment = '' AND erpnext_payment_id = ''")
	}
}

func (r *paymentMethodRepo) WhereNotCode(codes []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code NOT IN ?", codes)
	}
}

func (r *paymentMethodRepo) WherePaymentName(paymentName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("payment_name = ?", paymentName)
	}
}

func (r *paymentMethodRepo) WhereCode(code int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}
}

func (r *paymentMethodRepo) WhereSource(source int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

// UpdatePaymentMethod 更新支付方式
func (r *paymentMethodRepo) UpdatePaymentMethod(data map[string]any, options ...DBOption) error {
	db := r.db.Model(&model.PaymentMethod{})
	for _, opt := range options {
		db = opt(db)
	}
	if err := db.Updates(data).Error; err != nil {
		return errors.WithMessage(err, "更新支付方式失败")
	}
	return nil
}

// DeletePaymentMethod 删除支付方式（软删除）
func (r *paymentMethodRepo) DeletePaymentMethod(uuid uint64) error {
	now := time.Now().Unix()
	if err := r.db.Model(&model.PaymentMethod{}).
		Where("uuid = ? AND delete_time = 0", uuid).
		Update("delete_time", now).Error; err != nil {
		return errors.WithMessage(err, "删除支付方式失败")
	}
	return nil
}

// CheckHasOrders 检查支付方式是否有关联订单
func (r *paymentMethodRepo) CheckHasOrders(uuid uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.PaymentOrder{}).
		Where("payment_method_uuid = ? AND delete_time = 0", uuid).
		Count(&count).Error
	if err != nil {
		return false, errors.WithMessage(err, "检查关联订单失败")
	}
	return count > 0, nil
}

// GetMaxSort 获取最大排序值
func (r *paymentMethodRepo) GetMaxSort() (int, error) {
	var maxSort int
	err := r.db.Model(&model.PaymentMethod{}).
		Where("delete_time = 0").
		Select("IFNULL(MAX(sort), 0)").
		Scan(&maxSort).Error
	if err != nil {
		return 0, errors.WithMessage(err, "获取最大排序值失败")
	}
	return maxSort, nil
}

// BatchUpdateSort 批量更新排序
func (r *paymentMethodRepo) BatchUpdateSort(items []model.PaymentMethod) error {
	if len(items) == 0 {
		return nil
	}

	// 构建 CASE WHEN 语句
	caseWhenSQL := "CASE uuid"
	var args []any
	var uuids []uint64

	for _, item := range items {
		caseWhenSQL += " WHEN ? THEN ?"
		args = append(args, item.Uuid, item.Sort)
		uuids = append(uuids, item.Uuid)
	}
	caseWhenSQL += " END"

	// 使用事务批量更新排序
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.PaymentMethod{}).
			Where("uuid IN ? AND delete_time = 0", uuids).
			Update("sort", gorm.Expr(caseWhenSQL, args...)).Error
		if err != nil {
			return errors.WithMessage(err, "批量更新排序失败")
		}
		return nil
	})
}

// GetPaymentMethodListWithPagination 分页查询支付方式列表
func (r *paymentMethodRepo) GetPaymentMethodListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]*model.PaymentMethod, int64, error) {
	var list []*model.PaymentMethod
	var total int64

	db := r.db.Model(&model.PaymentMethod{})

	// 应用选项
	for _, opt := range opts {
		db = opt(db)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询支付方式总数失败")
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := db.Offset(offset).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询支付方式列表失败")
	}

	return list, total, nil
}

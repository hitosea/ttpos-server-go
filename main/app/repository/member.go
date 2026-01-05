package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type IMemberRepo interface {
	IMemberQueryRepo
	WithMemberLevel() DBOption      // Member 预加载会员等级
	WithMemberCard() DBOption       // Member 预加载会员卡
	WithMemberCardType() DBOption   // Member 预加载会员卡.卡类型
	WithMemberBalanceLog() DBOption // Member 预加载会员余额变动记录

	WhereUuid(uuid uint64) DBOption         // Uuid 条件
	WhereCardNo(cardNo string) DBOption     // 卡号条件
	WhereDeviceId(deviceId string) DBOption // 设备ID条件
	WhereIsVisitor(isVisitor bool) DBOption // 是否游客条件

	CreateMember(member *model.Member) error             // 添加会员
	CreateMemberCard(memberCard *model.MemberCard) error // 给会员发卡
	CreateMemberAndMemberCard(member model.Member) error // 添加会员并发放会员卡。仅用于数据迁移
	Update(uuid uint64, vars map[string]any) error       // 更新会员信息

	GetMemberInfoForSaleOrder(ctx context.Context, memberUuid uint64) (*model.Member, error) // 获取会员信息，用于销售订单结账时
	GetMemberWithAssociations(memberUuid uint64) (*model.Member, error)                      // 根据会员UUID查询会员信息（包含MemberLevel、MemberCard.MemberCardType、MemberBalanceLog）
	GetMembersWithAssociations(memberUuids []uint64) ([]*model.Member, error)                // 批量根据会员UUID列表查询会员信息（包含MemberLevel、MemberCard.MemberCardType、MemberBalanceLog）
	UpdateProcessed(uuids []uint64) error                                                    // 更新会员积分日志为已处理

	IncConsumptionAmount(uuid uint64, amount float64) error // 增加会员消费金额
	IncConsumptionCount(uuid uint64) error                  // 增加会员消费次数
	DecConsumptionAmount(uuid uint64, amount float64) error // 减少会员消费金额
	DecConsumptionCount(uuid uint64) error                  // 减少会员消费次数

	GetVisitorByDeviceId(deviceId string) (*model.Member, error) // 根据设备ID获取游客会员
	CheckVisitorExists(deviceId string) bool                     // 检查游客是否存在
	CheckNicknameExists(nickname string) bool                    // 根据昵称检查是否存在
}

// IMemberQueryRepo 会员查询
type IMemberQueryRepo interface {
	GetMember(opts ...DBOption) model.Member // 获取会员
	GetMemberList(opts ...DBOption) ([]*model.Member, error)
	GetMemberRecord(opts ...DBOption) (*model.Member, error)            // 获取会员
	GetMemberByUuid(uuid uint64) (*model.Member, error)                 // 根据uuid获取会员
	GetMemberByReferrerUuid(referrerUuid uint64) (*model.Member, error) // 根据uuid获取会员
	GetMemberByPhone(phone string) (*model.Member, error)               // 根据手机号获取会员
	GetMemberByPhoneContainDeleted(phone string) (*model.Member, error) // 根据手机号获取会员 (包含已删除的会员)
	GetMembersByUuids(uuids []uint64) ([]*model.Member, error)          // 根据uuid列表获取会员列表
	GetMemberLevels() []model.MemberLevel                               // 获取会员等级
	GetMemberLevelMinPriorityUuid() uint64                              // 获取会员等级权重最小的会员等级Uuid
	GetCardTypes() []model.MemberCardType                               // 获取会员卡类型
	GetCheckCardType(uuid uint64) model.MemberCardType                  // 获取会员卡类型
	GetMemberLevelsAllColumns() []model.MemberLevel                     // 获取会员等级所有列
	SearchMember(keyword string) []model.Member                         // 关键字搜索会员
	CheckMemberExists(phone string) bool                                // 根据手机号检查是否存在
	CheckNicknameExists(nickname string) bool                           // 根据昵称检查是否存在
	CheckLevelExists(uuid uint64) bool                                  // 根据Uuid检查等级是否存在
}

func NewMemberRepo(db *gorm.DB) IMemberRepo {
	return NewMemberRepoImpl(db)
}

type memberRepo struct {
	db *gorm.DB
}

func NewMemberRepoImpl(db *gorm.DB) IMemberRepo {
	return &memberRepo{db: db}
}

// GetMemberLevels 获取会员等级
func (r *memberRepo) GetMemberLevels() []model.MemberLevel {
	var levels []model.MemberLevel
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Select("uuid, name, priority, create_time, points_rate, points_quantity").Order("priority asc, create_time asc").Find(&levels)
	return levels
}

// GetMemberLevelMinPriorityUuid 获取会员等级权重最小的会员等级Uuid
func (r *memberRepo) GetMemberLevelMinPriorityUuid() uint64 {
	var uuid uint64
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Order("priority asc, create_time asc").Select("uuid").First(&uuid)
	return uuid
}

// GetCardTypes 获取会员卡类型
func (r *memberRepo) GetCardTypes() []model.MemberCardType {
	var cardTypes []model.MemberCardType
	r.db.Model(&model.MemberCardType{}).Scopes(NotDeleted).Where("status = ?", constant.CardTypeEnable).Order("sort asc").Find(&cardTypes)
	return cardTypes
}

// GetMemberLevelsAllColumns 获取会员等级
func (r *memberRepo) GetMemberLevelsAllColumns() []model.MemberLevel {
	var levels []model.MemberLevel
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Order("priority asc, create_time asc").Find(&levels)
	return levels
}

// SearchMember 关键字搜索会员
func (r *memberRepo) SearchMember(keyword string) []model.Member {
	var members []model.Member
	keyword = Like(keyword)
	r.db.Model(&model.Member{}).Scopes(NotDeleted).Where("is_visitor = ?", 0).Select("uuid, nickname, phone, member_card_no").Where("phone LIKE ? OR id LIKE ? OR nickname LIKE ? OR member_card_no LIKE ?", keyword, keyword, keyword, keyword).Limit(200).Find(&members)
	return members
}

// CreateMember 添加会员
func (r *memberRepo) CreateMember(member *model.Member) error {
	err := r.db.Create(member).Error
	return err
}

// CreateMemberCard 给会员发卡
func (r *memberRepo) CreateMemberCard(memberCard *model.MemberCard) error {
	err := r.db.Create(memberCard).Error
	return err
}

// CreateMemberAndMemberCard 添加会员并发放会员卡。仅用于数据迁移
func (r *memberRepo) CreateMemberAndMemberCard(obj model.Member) error {
	member := obj
	member.SetNil()
	if err := r.db.Create(&member).Error; err != nil {
		return errors.WithMessage(err, "添加会员失败")
	}
	memberCard := obj.MemberCard
	if err := r.db.Create(&memberCard).Error; err != nil {
		return errors.WithMessage(err, "添加会员卡失败")
	}
	return nil
}

// CheckMemberExists 根据手机号检查是否存在
func (r *memberRepo) CheckMemberExists(phone string) bool {
	var memberUuid uint64
	r.db.Model(&model.Member{}).Scopes(NotDeleted).Where("phone = ?", phone).Select("uuid").Scan(&memberUuid)
	return memberUuid > 0
}

// CheckLevelExists 根据uuid检查等级是否存在
func (r *memberRepo) CheckLevelExists(uuid uint64) bool {
	var exists uint64
	r.db.Model(&model.MemberLevel{}).Scopes(NotDeleted).Where("uuid = ?", uuid).Select("uuid").Scan(&exists)
	return exists > 0
}

// GetCheckCardType 获取会员卡类型
func (r *memberRepo) GetCheckCardType(uuid uint64) model.MemberCardType {
	var cardType model.MemberCardType
	r.db.Model(&model.MemberCardType{}).Scopes(NotDeleted).Where("uuid = ? AND status = ?", uuid, constant.CardTypeEnable).Find(&cardType)
	return cardType
}

// GetMember 查询会员
func (r *memberRepo) GetMember(opts ...DBOption) model.Member {
	var member model.Member
	db := r.db.Model(&model.Member{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&member)
	return member
}

// GetMemberList 查询会员列表
func (r *memberRepo) GetMemberList(opts ...DBOption) ([]*model.Member, error) {
	var members []*model.Member
	db := r.db.Model(&model.Member{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&members).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员列表失败")
	}
	return members, nil
}

// GetMembersByUuids 根据uuid列表查询会员列表
func (r *memberRepo) GetMembersByUuids(uuids []uint64) ([]*model.Member, error) {
	members, err := r.GetMemberList(
		CommonRepo.WhereInUuids(uuids),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员列表失败")
	}
	return members, nil
}

// GetMemberRecord 查询会员
func (r *memberRepo) GetMemberRecord(opts ...DBOption) (*model.Member, error) {
	var member model.Member
	db := r.db.Model(&model.Member{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&member).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员失败")
	}
	return &member, nil
}

// GetMemberByUuid 根据uuid查询会员
func (r *memberRepo) GetMemberByUuid(uuid uint64) (*model.Member, error) {
	member, err := r.GetMemberRecord(
		r.WhereUuid(uuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MemberLevel",
			},
			WithPreload{
				Query: "MemberCard",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员失败", utils.NumToStr(uuid))
	}
	return member, nil
}

// GetMemberByReferrerUuid 根据推荐人uuid查询会员
func (r *memberRepo) GetMemberByReferrerUuid(referrerUuid uint64) (*model.Member, error) {
	member, err := r.GetMemberRecord(r.WhereUuid(referrerUuid))
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员失败", utils.NumToStr(referrerUuid))
	}
	return member, nil
}

// GetMemberByPhone 根据手机号查询会员
func (r *memberRepo) GetMemberByPhone(phone string) (*model.Member, error) {
	var member model.Member
	err := r.db.Model(&model.Member{}).Scopes(NotDeleted).Where("phone = ?", phone).First(&member).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员失败", phone)
	}
	return &member, nil
}

// GetMemberByPhoneAndDeleted 根据手机号查询会员 (包含已删除的会员)
func (r *memberRepo) GetMemberByPhoneContainDeleted(phone string) (*model.Member, error) {
	var member model.Member
	err := r.db.Model(&model.Member{}).Where("phone = ?", phone).First(&member).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询会员失败", phone)
	}
	return &member, nil
}

// Update 更新会员信息
func (r *memberRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// WithMemberLevel Member 预加载会员等级
func (r *memberRepo) WithMemberLevel() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberLevel")
	}
}

// WithMemberCard Member 预加载会员卡
func (r *memberRepo) WithMemberCard() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard")
	}
}

// WithMemberCardType Member 预加载会员卡.卡类型
func (r *memberRepo) WithMemberCardType() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard.MemberCardType")
	}
}

// WithMemberBalanceLog Member 预加载会员余额变动记录
func (r *memberRepo) WithMemberBalanceLog() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberBalanceLog")
	}
}

// WhereUuid Uuid 条件
func (r *memberRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereCardNo 卡号条件
func (r *memberRepo) WhereCardNo(cardNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("member_card_no = ?", cardNo)
	}
}

// WhereDeviceId 设备ID条件
func (r *memberRepo) WhereDeviceId(deviceId string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceId)
	}
}

// WhereIsVisitor 是否游客条件
func (r *memberRepo) WhereIsVisitor(isVisitor bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_visitor = ?", isVisitor)
	}
}

func (r *memberRepo) GetMemberInfoForSaleOrder(ctx context.Context, memberUuid uint64) (*model.Member, error) {
	member := r.GetMember(r.WhereUuid(memberUuid), r.WithMemberCardType(), r.WithMemberLevel(), r.WithMemberBalanceLog())
	if member.Uuid == 0 {
		return nil, errors.New("会员不存在")
	}
	return &member, nil
}

// GetMemberWithAssociations 根据会员UUID查询会员信息（包含MemberLevel、MemberCard.MemberCardType、MemberBalanceLog）
func (r *memberRepo) GetMemberWithAssociations(memberUuid uint64) (*model.Member, error) {
	if memberUuid == 0 {
		return nil, errors.New("会员UUID不能为0")
	}
	member := r.GetMember(r.WhereUuid(memberUuid), r.WithMemberLevel(), r.WithMemberCardType(), r.WithMemberBalanceLog())
	if member.Uuid == 0 {
		return nil, errors.New("会员不存在")
	}
	return &member, nil
}

// GetMembersWithAssociations 批量根据会员UUID列表查询会员信息（包含MemberLevel、MemberCard.MemberCardType、MemberBalanceLog）
func (r *memberRepo) GetMembersWithAssociations(memberUuids []uint64) ([]*model.Member, error) {
	if len(memberUuids) == 0 {
		return []*model.Member{}, nil
	}
	// 过滤掉 0 值
	validUuids := make([]uint64, 0, len(memberUuids))
	for _, uuid := range memberUuids {
		if uuid > 0 {
			validUuids = append(validUuids, uuid)
		}
	}
	if len(validUuids) == 0 {
		return []*model.Member{}, nil
	}
	// 批量查询会员列表，并预加载关联数据
	members, err := r.GetMemberList(
		CommonRepo.WhereInUuids(validUuids),
		r.WithMemberLevel(),
		r.WithMemberCardType(),
		r.WithMemberBalanceLog(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询会员信息失败")
	}
	return members, nil
}

// UpdateProcessed 更新会员积分日志为已处理
func (r *memberRepo) UpdateProcessed(uuids []uint64) error {
	if err := r.db.Model(&model.MemberBalanceLog{}).Where("uuid IN (?)", uuids).Update("processed", constant.MemberPointLogOrBalanceProcessedSuccess).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// IncConsumptionAmount 增加会员消费金额
func (r *memberRepo) IncConsumptionAmount(uuid uint64, amount float64) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Update("accumulated_consumption_amount", gorm.Expr("accumulated_consumption_amount + ?", amount)).Error
}

// IncConsumptionCount 增加会员消费次数
func (r *memberRepo) IncConsumptionCount(uuid uint64) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Update("consumption_count", gorm.Expr("consumption_count + 1")).Error
}

// DecConsumptionAmount 减少会员消费金额
func (r *memberRepo) DecConsumptionAmount(uuid uint64, amount float64) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Update("accumulated_consumption_amount", gorm.Expr("accumulated_consumption_amount - ?", amount)).Error
}

// DecConsumptionCount 减少会员消费次数
func (r *memberRepo) DecConsumptionCount(uuid uint64) error {
	return r.db.Model(&model.Member{}).Where("uuid = ?", uuid).Update("consumption_count", gorm.Expr("consumption_count - 1")).Error
}

// GetVisitorByDeviceId 根据设备ID获取游客会员
func (r *memberRepo) GetVisitorByDeviceId(deviceId string) (*model.Member, error) {
	var member model.Member
	err := r.db.Model(&model.Member{}).Where("device_id = ? AND is_visitor = ?", deviceId, 1).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &member, nil
		}
		return nil, err
	}
	return &member, nil
}

// CheckVisitorExists 检查游客是否存在
func (r *memberRepo) CheckVisitorExists(deviceId string) bool {
	var exists uint64
	r.db.Model(&model.Member{}).Where("device_id = ? AND is_visitor = ?", deviceId, 1).Select("uuid").Scan(&exists)
	return exists > 0
}

// CheckNicknameExists 根据昵称检查是否存在
func (r *memberRepo) CheckNicknameExists(nickname string) bool {
	var exists uint64
	r.db.Model(&model.Member{}).Where("nickname = ?", nickname).Select("uuid").Scan(&exists)
	return exists > 0
}

package repository

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/encrypt"

	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// QrCodeParams 二维码参数
type QrCodeParams struct {
	Type         uint64 `json:"t"`       // 活动类型
	CompanyUuid  uint64 `json:"c_uuid"`  // 公司uuid
	MemberUuid   uint64 `json:"m_uuid"`  // 会员uuid
	ActivityUuid uint64 `json:"mc_uuid"` // 营销活动uuid
}

type IMarketingActivityRepo interface {
	GetActivity(uuid uint64) (*model.MarketingActivity, error)                            // 获取营销活动
	GetActivityAndPrizes(uuid uint64) (*model.MarketingActivity, error)                   // 获取营销活动与奖励
	GetMemberClientActivityList(dbOption ...DBOption) ([]*model.MarketingActivity, error) // 获取会员端营销活动列表
	GetValidActivityListByNow(dbOption ...DBOption) ([]*model.MarketingActivity, error)   // 获取正在进行中的营销活动列表
	GetValidActivity() (*model.MarketingActivity, error)                                  // 获取正在进行中的营销活动
	GetValidActivityByUuid(uuid uint64) (*model.MarketingActivity, error)                 // 根据uuid获取正在进行中的营销活动
	GenerateQrCode(params *QrCodeParams) (string, error)                                  // 生成二维码

	SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error // 软删除总部活动（排除指定uuid）
	FindHeadquarterByUuidsWithPrizes(uuids []uint64) ([]model.MarketingActivity, error)      // 根据uuid列表查询总部活动（含奖品）
	UnscopedDeleteByUuid(uuid uint64) error                                                   // 根据uuid硬删除活动
	CreateActivity(activity *model.MarketingActivity) error                                    // 创建活动
	UnscopedDeleteByHeadquarter(headquarterUuid uint64) error                                  // 根据总部uuid硬删除活动
	FindByHeadquarter(headquarterUuid uint64) ([]model.MarketingActivity, error)               // 根据总部uuid查询活动
	FindAllHeadquarterWithPrizes() ([]model.MarketingActivity, error)                          // 查询所有总部活动（含奖品）

	WithLanguage() DBOption
	WithPrizes() DBOption
}

// MarketingActivityRepo 营销活动仓库
type MarketingActivityRepo struct {
	db *gorm.DB
}

// NewMarketingActivityRepo 创建营销活动仓库
func NewMarketingActivityRepo(db *gorm.DB) IMarketingActivityRepo {
	return &MarketingActivityRepo{
		db: db,
	}
}

// WithLanguage 根据语言获取营销活动
func (r *MarketingActivityRepo) WithLanguage() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName").Preload("MultiLanguageDesc")
	}
}

// WithPrizes 根据语言获取营销活动
func (r *MarketingActivityRepo) WithPrizes() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Prizes", NotDeleted).Preload("Prizes.Coupon", NotDeleted)
	}
}

// GetActivity 获取营销活动
func (r *MarketingActivityRepo) GetActivity(uuid uint64) (*model.MarketingActivity, error) {
	var activity model.MarketingActivity
	err := r.db.Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetActivityList 获取营销活动列表
func (r *MarketingActivityRepo) GetMemberClientActivityList(opts ...DBOption) ([]*model.MarketingActivity, error) {
	activities := make([]*model.MarketingActivity, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix() // 7天前的时间戳
	nowUnix := time.Now().Unix()
	result := db.Scopes(NotDeleted)
	result = result.Where("start_time <= ? AND end_time >= ?", nowUnix, sevenDaysAgo)
	// 按活动状态排序：进行中(1)排最前，未开始(0)其次，已结束/无效(2)最后
	orderSQL := fmt.Sprintf(`
		CASE 
			WHEN is_invalid = 1 THEN 2
			WHEN start_time > %d THEN 0  
			WHEN end_time < %d THEN 2
			ELSE 1 
		END ASC, start_time DESC
	`, nowUnix, nowUnix)
	result = result.Order(orderSQL).Find(&activities)
	if result.Error != nil {
		return nil, result.Error
	}

	return activities, nil
}

// GetActivityAndPrizes 获取营销活动与奖励
func (r *MarketingActivityRepo) GetActivityAndPrizes(uuid uint64) (*model.MarketingActivity, error) {
	var activity model.MarketingActivity
	err := r.db.Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
		Preload("Prizes", NotDeleted).
		Preload("Prizes.Coupon", NotDeleted).
		First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetActivityListByNow 获取正在进行中的营销活动列表
func (r *MarketingActivityRepo) GetValidActivityListByNow(opts ...DBOption) ([]*model.MarketingActivity, error) {
	var activities []*model.MarketingActivity
	db := r.db.Model(&model.MarketingActivity{}).
		Preload("Prizes", NotDeleted).
		Preload("MultiLanguageName").
		Preload("MultiLanguageDesc").
		Where("delete_time = ? AND start_time <= ? AND end_time >= ? AND is_invalid = ?", constant.NotDeleted, time.Now().Unix(), time.Now().Unix(), 0).
		Order("start_time DESC").
		Order("end_time DESC")
	// 应用选项
	for _, opt := range opts {
		db = opt(db)
	}
	// 查询
	err := db.Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// GetValidActivity 获取正在进行中的营销活动
func (r *MarketingActivityRepo) GetValidActivity() (*model.MarketingActivity, error) {
	return r.GetValidActivityByUuid(0)
}

// GetValidActivityByUuid 根据uuid获取营销活动
func (r *MarketingActivityRepo) GetValidActivityByUuid(uuid uint64) (*model.MarketingActivity, error) {
	var activity model.MarketingActivity
	db := r.db.Model(&model.MarketingActivity{}).
		Preload("Prizes", NotDeleted).
		Preload("MultiLanguageName").
		Preload("MultiLanguageDesc").
		Where("delete_time = ? AND start_time <= ? AND end_time >= ? AND is_invalid = ?", constant.NotDeleted, time.Now().Unix(), time.Now().Unix(), 0).
		Order("start_time DESC").
		Order("end_time DESC")
	//
	if uuid != 0 {
		db = db.Where("uuid = ?", uuid)
	}
	//
	err := db.First(&activity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &activity, nil
}

// GenerateQrCode 生成二维码
func (r *MarketingActivityRepo) GenerateQrCode(params *QrCodeParams) (string, error) {
	// 将参数转换为JSON
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	// 加密JSON数据
	encryptedData, err := encrypt.EncryptAesString(string(jsonData))
	if err != nil {
		return "", err
	}
	// 生成二维码
	qr, err := qrcode.Encode(encryptedData, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	// 转换为base64
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(qr), nil
}

// SoftDeleteByHeadquarterExcludeUuids 软删除总部活动（排除指定uuid）
func (r *MarketingActivityRepo) SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error {
	return r.db.Model(&model.MarketingActivity{}).
		Where("headquarter_uuid = ? AND uuid NOT IN (?)", headquarterUuid, excludeUuids).
		Update("delete_time", time.Now().Unix()).Error
}

// FindHeadquarterByUuidsWithPrizes 根据uuid列表查询总部活动（含奖品）
func (r *MarketingActivityRepo) FindHeadquarterByUuidsWithPrizes(uuids []uint64) ([]model.MarketingActivity, error) {
	var activities []model.MarketingActivity
	err := r.db.Where("delete_time = 0 AND headquarter_uuid = 0 AND uuid IN (?)", uuids).
		Preload("Prizes").
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// UnscopedDeleteByUuid 根据uuid硬删除活动
func (r *MarketingActivityRepo) UnscopedDeleteByUuid(uuid uint64) error {
	return r.db.Unscoped().Where("uuid = ?", uuid).Delete(&model.MarketingActivity{}).Error
}

// CreateActivity 创建活动
func (r *MarketingActivityRepo) CreateActivity(activity *model.MarketingActivity) error {
	return r.db.Create(activity).Error
}

// UnscopedDeleteByHeadquarter 根据总部uuid硬删除活动
func (r *MarketingActivityRepo) UnscopedDeleteByHeadquarter(headquarterUuid uint64) error {
	return r.db.Unscoped().Where("headquarter_uuid = ?", headquarterUuid).Delete(&model.MarketingActivity{}).Error
}

// FindByHeadquarter 根据总部uuid查询活动
func (r *MarketingActivityRepo) FindByHeadquarter(headquarterUuid uint64) ([]model.MarketingActivity, error) {
	var activities []model.MarketingActivity
	err := r.db.Where("headquarter_uuid = ?", headquarterUuid).Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// FindAllHeadquarterWithPrizes 查询所有总部活动（含奖品）
func (r *MarketingActivityRepo) FindAllHeadquarterWithPrizes() ([]model.MarketingActivity, error) {
	var activities []model.MarketingActivity
	err := r.db.Where("delete_time = 0 AND headquarter_uuid = 0").
		Preload("Prizes").
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

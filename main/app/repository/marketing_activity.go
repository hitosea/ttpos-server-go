package repository

import (
	"encoding/base64"
	"encoding/json"
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
	GetActivityListByNow() ([]*model.MarketingActivity, error)                            // 获取正在进行中的营销活动列表
	GenerateQrCode(params *QrCodeParams) (string, error)                                  // 生成二维码

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
	result := db.Scopes(NotDeleted)
	result = result.Where("start_time <= ? AND end_time >= ?", time.Now().Unix(), sevenDaysAgo)
	result = result.Find(&activities)
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
func (r *MarketingActivityRepo) GetActivityListByNow() ([]*model.MarketingActivity, error) {
	var activities []*model.MarketingActivity
	db := r.db.Model(&model.MarketingActivity{}).
		Preload("Prizes", NotDeleted).
		Preload("MultiLanguageName").
		Preload("MultiLanguageDesc").
		Where("delete_time = ? AND start_time <= ? AND end_time >= ? AND is_invalid = ?", constant.NotDeleted, time.Now().Unix(), time.Now().Unix(), 0).
		Order("start_time DESC").
		Order("end_time DESC")
	err := db.Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
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

package repository

import (
	"encoding/base64"
	"encoding/json"
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
	// GetActivity 获取营销活动
	GetActivity(uuid uint64) (*model.MarketingActivity, error)
	// GetActivityAndPrizes 获取营销活动与奖励
	GetActivityAndPrizes(uuid uint64) (*model.MarketingActivity, error)
	// GetActivityListByNow 获取正在进行中的营销活动列表
	GetActivityListByNow() ([]*model.MarketingActivity, error)
	// GenerateQrCode 生成二维码
	GenerateQrCode(params *QrCodeParams) (string, error)
	// SendReward 发放奖励
	SendReward(activityUuid, memberUuid uint64) error
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

// GetActivity 获取营销活动
func (r *MarketingActivityRepo) GetActivity(uuid uint64) (*model.MarketingActivity, error) {
	var activity model.MarketingActivity
	err := r.db.Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
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

// 计算奖励次数
func (r *MarketingActivityRepo) calculateRewardCount(consumptionAmount, rewardConditionAmount float64) int {
	if rewardConditionAmount <= 0 {
		return 0
	}
	// 向下取整，确保不会超过实际消费金额对应的奖励次数
	return int(consumptionAmount / rewardConditionAmount)
}

// SendReward 发放奖励
func (r *MarketingActivityRepo) SendReward(activityUuid, memberUuid uint64) error {
	// 获取活动信息
	activity, err := r.GetActivityAndPrizes(activityUuid)
	if err != nil {
		return err
	}
	if activity == nil {
		return fmt.Errorf("activity not found")
	}

	// 检查活动是否有效
	if !activity.IsValid() {
		return fmt.Errorf("activity is invalid")
	}

	// 检查是否开启奖励次数限制, 如果已经达到奖励次数限制，则不再发放奖励
	rewardCount, err := NewMarketingActivityRecordRepo(r.db).GetRewardCount(activityUuid, memberUuid)
	if err != nil {
		return err
	}
	if activity.IsOpenRewardLimit == 1 && rewardCount >= activity.RewardLimit {
		return fmt.Errorf("reward limit reached")
	}

	// 获取推荐人的总消费金额
	consumptionAmount, err := NewMarketingActivityConsumptionRepo(r.db).GetByActivityAndReferrerConsumptionAmount(activityUuid, memberUuid)
	if err != nil {
		return err
	}

	// 计算应该发放的奖励次数
	rewardCountToGive := r.calculateRewardCount(consumptionAmount, activity.RewardConditionAmount)
	if rewardCountToGive <= 0 {
		return fmt.Errorf("no reward to give")
	}
	// 最多等于奖励次数限制
	if activity.IsOpenRewardLimit == 1 {
		if rewardCountToGive > int(activity.RewardLimit) {
			rewardCountToGive = int(activity.RewardLimit)
		}
	}
	// 计算应该发放的奖励次数，减去已经发放的奖励次数
	rewardCountToGive = rewardCountToGive - int(rewardCount)
	if rewardCountToGive < 0 {
		rewardCountToGive = 0
	}

	// 发放优惠券
	marketingCoupon := &model.MarketingCoupon{}
	if len(activity.Prizes) > 0 && activity.Prizes[0] != nil && activity.Prizes[0].Coupon != nil {
		marketingCoupon = activity.Prizes[0].Coupon
	}

	// 循环发放多次奖励
	for i := 0; i < rewardCountToGive; i++ {
		err = r.db.Transaction(func(tx *gorm.DB) error {
			// 只有当优惠券存在时才创建会员优惠券
			if marketingCoupon != nil && marketingCoupon.Uuid != 0 {
				// 创建优惠券
				err = NewMarketingCouponRepo(tx).DecreaseCouponQuantity(marketingCoupon.Uuid, memberUuid, activityUuid)
				if err != nil {
					return err
				}
				// 创建奖励记录
				err = NewMarketingActivityRecordRepo(tx).CreateRecord(&model.MarketingActivityRecord{
					ActivityUuid:   activityUuid,
					MemberUuid:     memberUuid,
					RewardCount:    1,
					LastRewardTime: time.Now().Unix(),
					PrizeUuid: func() uint64 {
						if len(activity.Prizes) > 0 && activity.Prizes[0] != nil {
							return activity.Prizes[0].PrizeUuid
						}
						return 0
					}(),
				})
				if err != nil {
					return err
				}
				// 创建会员优惠券
				err = NewMemberCouponRepo(tx).CreateMemberCoupon(&model.MemberCoupon{
					MemberUuid:     memberUuid,
					Type:           "deduction",
					Status:         0,
					CouponUuid:     marketingCoupon.Uuid,
					Name:           marketingCoupon.Name,
					DeductionType:  marketingCoupon.DeductionType,
					DayStartTime:   marketingCoupon.DayStartTime,
					DayEndTime:     marketingCoupon.DayEndTime,
					ValidStartTime: marketingCoupon.ValidStartTime,
					ValidEndTime:   marketingCoupon.ValidEndTime,
					Amount:         marketingCoupon.Amount,
					StartTime:      time.Now().Unix(),
					EndTime:        time.Now().AddDate(0, 0, marketingCoupon.ValidDays).Unix(),
				})
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

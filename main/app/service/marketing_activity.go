package service

import (
	"encoding/json"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/encrypt"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type IMarketingActivitySrv interface {
	MarketingActivity(ctx context.Context, req member_req.MarketingActivityReq) (*member_resp.MemberMarketingActivityListResp, error)               // 获取营销活动
	MarketingActivityDetail(ctx context.Context, req member_req.MarketingActivityDetailReq) (*member_resp.MemberMarketingActivityDetailResp, error) // 获取营销活动详情
	MarketingActivityList(ctx context.Context) (*member_resp.MemberMarketingActivityListsResp, error)                                               // 获取营销活动列表
	DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error)                                                   // 解密活动二维码
}

type marketingActivitySrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewMarketingActivitySrv(
	dbm *database.DBManager,
	cache cache.Cache,
) IMarketingActivitySrv {
	return NewMarketingActivitySrvImpl(dbm, cache)
}

func NewMarketingActivitySrvImpl(
	dbm *database.DBManager,
	cache cache.Cache,
) IMarketingActivitySrv {
	return &marketingActivitySrv{
		dbm:   dbm,
		cache: cache,
	}
}

// MarketingActivity 获取营销活动
func (s *marketingActivitySrv) MarketingActivity(ctx context.Context, req member_req.MarketingActivityReq) (*member_resp.MemberMarketingActivityListResp, error) {
	member := ctx.GetMember()
	company := ctx.GetCompany()

	// 获取正在进行中的营销活动列表
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivityList, err := marketingActivityRepo.GetValidActivityListByNow(
		func(db *gorm.DB) *gorm.DB {
			if req.Uuid != 0 {
				return db.Where("uuid = ?", req.Uuid)
			}
			return db
		},
	)
	if err != nil {
		return nil, err
	}
	// 生成二维码
	memberMarketingActivityResp := make([]member_resp.MemberMarketingActivityResp, 0)
	for _, marketingActivity := range marketingActivityList {
		qrCode, err := marketingActivityRepo.GenerateQrCode(&repository.QrCodeParams{
			Type:         uint64(marketingActivity.Type),
			CompanyUuid:  company.Uuid,
			MemberUuid:   member.Uuid,
			ActivityUuid: marketingActivity.Uuid,
		})
		if err != nil {
			return nil, err
		}
		memberMarketingActivityResp = append(memberMarketingActivityResp, member_resp.MemberMarketingActivityResp{
			Name:       marketingActivity.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			LocaleName: marketingActivity.MultiLanguageName.GetNames(),
			Desc:       marketingActivity.MultiLanguageDesc.GetNameByLang(ctx.GetLanguage()),
			LocaleDesc: marketingActivity.MultiLanguageDesc.GetNames(),
			QrCode:     qrCode,
			StartTime:  int64(marketingActivity.StartTime),
			EndTime:    int64(marketingActivity.EndTime),
			IsInvalid:  utils.IfInt(company.CompanySetting.IsOpenMarketing != 1, 1, 0),
		})
	}

	result := &member_resp.MemberMarketingActivityListResp{
		List: memberMarketingActivityResp,
		MemberInfo: member_resp.MemberInfoResp{
			Uuid:     member.Uuid,
			Nickname: member.Nickname,
			Phone:    member.Phone,
		},
		Company: member_resp.CompanyInfoResp{
			Uuid: company.Uuid,
			Name: company.Name,
			Logo: company.GetLogo(utils.GetBaseURL(ctx.Copy().GetGin().Request)),
		},
	}

	// 无营销活动
	if len(result.List) == 0 {
		qrCode, err := marketingActivityRepo.GenerateQrCode(&repository.QrCodeParams{
			Type:         0,
			CompanyUuid:  company.Uuid,
			MemberUuid:   member.Uuid,
			ActivityUuid: 0,
		})
		if err != nil {
			return nil, err
		}
		result.List = append(result.List, member_resp.MemberMarketingActivityResp{
			Name: "暂无营销活动",
			LocaleName: dto.LocaleResponse{
				ZH:   "暂无营销活动",
				TH:   "ไม่มีกิจกรรมการตลาด",
				EN:   "No Marketing Activities",
				ZHTW: "暫無行銷活動",
				JA:   "マーケティング活動なし",
				KO:   "마케팅 활동 없음",
				MY:   "Tiada Aktiviti Pemasaran",
			},
			Desc: "暂无营销活动",
			LocaleDesc: dto.LocaleResponse{
				ZH:   "暂无营销活动",
				TH:   "ไม่มีกิจกรรมการตลาด",
				EN:   "No Marketing Activities",
				ZHTW: "暫無行銷活動",
				JA:   "マーケティング活動なし",
				KO:   "마케팅 활동 없음",
				MY:   "Tiada Aktiviti Pemasaran",
			},
			QrCode:    qrCode,
			StartTime: 0,
			EndTime:   0,
			IsInvalid: 1,
		})
		return result, errors.NewWithCode(constant.CodeMarketingActivityInvalid, "营销活动已失效")
	}

	// 产品： 仍可正常登录进入商家会员服务，但是进入后toast提示：商家营销活动已结束。
	if company.CompanySetting.IsOpenMarketing != 1 {
		return result, errors.NewWithCode(constant.CodeMarketingActivityInvalid, "营销活动已失效")
	}

	// 返回
	return result, nil
}

// MarketingActivityDetail 获取营销活动详情
func (s *marketingActivitySrv) MarketingActivityDetail(ctx context.Context, req member_req.MarketingActivityDetailReq) (*member_resp.MemberMarketingActivityDetailResp, error) {
	member := ctx.GetMember()
	company := ctx.GetCompany()

	// 获取营销活动
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivity, err := marketingActivityRepo.GetValidActivity()
	if err != nil {
		return nil, err
	}

	// 结果初始化
	result := &member_resp.MemberMarketingActivityDetailResp{
		Activity: member_resp.MemberMarketingActivityResp{},
		MemberInfo: member_resp.MemberInfoResp{
			Uuid:     member.Uuid,
			Nickname: member.Nickname,
			Phone:    member.Phone,
		},
		Company: member_resp.CompanyInfoResp{
			Uuid: company.Uuid,
			Name: company.Name,
			Logo: func() string {
				baseURL := utils.GetBaseURL(ctx.Copy().GetGin().Request)
				logoBase64, err := utils.AddImageDomainAndConvertToBase64(company.Logo, baseURL, true)
				if err != nil {
					return utils.AddImageDomain(company.Logo, baseURL, true)
				}
				return logoBase64
			}(),
		},
	}

	// 无营销活动
	if marketingActivity != nil {
		result.Activity = member_resp.MemberMarketingActivityResp{
			Name:       marketingActivity.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			LocaleName: marketingActivity.MultiLanguageName.GetNames(),
			Desc:       marketingActivity.MultiLanguageDesc.GetNameByLang(ctx.GetLanguage()),
			LocaleDesc: marketingActivity.MultiLanguageDesc.GetNames(),
			QrCode: func() string {
				qrCode, err := marketingActivityRepo.GenerateQrCode(&repository.QrCodeParams{
					Type:         uint64(marketingActivity.Type),
					CompanyUuid:  company.Uuid,
					MemberUuid:   member.Uuid,
					ActivityUuid: marketingActivity.Uuid,
				})
				if err != nil {
					return ""
				}
				return qrCode
			}(),
			StartTime: int64(marketingActivity.StartTime),
			EndTime:   int64(marketingActivity.EndTime),
			IsInvalid: utils.IfInt(company.CompanySetting.IsOpenMarketing != 1, 1, 0),
		}
	} else {
		result.Activity = member_resp.MemberMarketingActivityResp{
			Name: "暂无营销活动",
			LocaleName: dto.LocaleResponse{
				ZH:   "暂无营销活动",
				TH:   "ไม่มีกิจกรรมการตลาด",
				EN:   "No Marketing Activities",
				ZHTW: "暫無行銷活動",
				JA:   "マーケティング活動なし",
				KO:   "마케팅 활동 없음",
				MY:   "Tiada Aktiviti Pemasaran",
			},
			Desc: "暂无营销活动",
			LocaleDesc: dto.LocaleResponse{
				ZH:   "暂无营销活动",
				TH:   "ไม่มีกิจกรรมการตลาด",
				EN:   "No Marketing Activities",
				ZHTW: "暫無行銷活動",
				JA:   "マーケティング活動なし",
				KO:   "마케팅 활동 없음",
				MY:   "Tiada Aktiviti Pemasaran",
			},
			QrCode: func() string {
				qrCode, err := marketingActivityRepo.GenerateQrCode(&repository.QrCodeParams{
					Type:         0,
					CompanyUuid:  company.Uuid,
					MemberUuid:   member.Uuid,
					ActivityUuid: 0,
				})
				if err != nil {
					return ""
				}
				return qrCode
			}(),
			StartTime: 0,
			EndTime:   0,
			IsInvalid: 1,
		}
		return result, errors.NewWithCode(constant.CodeMarketingActivityInvalid, "营销活动已失效")
	}

	// 产品： 仍可正常登录进入商家会员服务，但是进入后toast提示：商家营销活动已结束。
	if company.CompanySetting.IsOpenMarketing != 1 {
		return result, errors.NewWithCode(constant.CodeMarketingActivityInvalid, "营销活动已失效")
	}

	// 返回
	return result, nil
}

// DecryptQrCode 解密活动二维码
func (s *marketingActivitySrv) DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error) {
	decryptQrCodeResp, err := encrypt.DecryptAesString(req.Sign)
	if err != nil {
		return nil, errors.New("活动二维码无效")
	}
	// 转json
	qrCodeParams := &repository.QrCodeParams{}
	err = json.Unmarshal([]byte(decryptQrCodeResp), qrCodeParams)
	if err != nil {
		return nil, errors.New("活动二维码无效")
	}
	// 获取活动
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivity, err := marketingActivityRepo.GetActivity(qrCodeParams.ActivityUuid)
	if err != nil || marketingActivity == nil {
		return nil, errors.New("活动二维码无效")
	}
	// 获取会员
	memberRepo := repository.NewMemberRepo(ctx.GetDB())
	member, err := memberRepo.GetMemberByUuid(qrCodeParams.MemberUuid)
	if err != nil {
		return nil, errors.New("活动二维码无效")
	}
	// 返回
	return &resp.DecryptQrCodeResp{
		Uuid:         member.Uuid,
		Nickname:     member.Nickname,
		Phone:        member.Phone,
		ActivityUuid: marketingActivity.Uuid,
	}, nil
}

// MarketingActivityList 获取营销活动列表
func (s *marketingActivitySrv) MarketingActivityList(ctx context.Context) (*member_resp.MemberMarketingActivityListsResp, error) {
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivityList, err := marketingActivityRepo.GetMemberClientActivityList(
		marketingActivityRepo.WithLanguage(),
		marketingActivityRepo.WithPrizes(),
	)
	if err != nil {
		return nil, err
	}

	memberMarketingActivityInfoResp := make([]member_resp.MemberMarketingActivityInfoResp, 0)
	for _, marketingActivity := range marketingActivityList {
		memberMarketingActivityInfoResp = append(memberMarketingActivityInfoResp, member_resp.MemberMarketingActivityInfoResp{
			Uuid:       marketingActivity.Uuid,
			Type:       marketingActivity.Type,
			LocaleName: marketingActivity.MultiLanguageName.GetNames(),
			LocaleDesc: marketingActivity.MultiLanguageDesc.GetNames(),
			StartTime:  int64(marketingActivity.StartTime),
			EndTime:    int64(marketingActivity.EndTime),
			Status:     marketingActivity.GetStatus(),
			Prizes: func() []member_resp.MemberMarketingActivityPrizeResp {
				prizes := make([]member_resp.MemberMarketingActivityPrizeResp, 0)
				if marketingActivity.RewardType == 1 {
					prizes = append(prizes, member_resp.MemberMarketingActivityPrizeResp{
						Uuid: marketingActivity.Uuid,
						LocalePrizeName: dto.LocaleResponse{
							ZH:   "积分",
							TH:   "คะแนน",
							EN:   "Points",
							ZHTW: "積分",
							JA:   "ポイント",
							KO:   "포인트",
							MY:   "คะแนน",
							TR:   "Puan",
							SV:   "Poäng",
						},
						RewardType:  marketingActivity.RewardType,
						RewardValue: marketingActivity.RewardValue,
					})
				} else {
					for _, prize := range marketingActivity.Prizes {
						// 检查 Coupon 是否为 nil，避免空指针异常
						if prize.Coupon == nil {
							continue
						}
						prizes = append(prizes, member_resp.MemberMarketingActivityPrizeResp{
							LocalePrizeName: dto.LocaleResponse{
								ZH:   prize.Coupon.Name,
								TH:   prize.Coupon.Name,
								EN:   prize.Coupon.Name,
								ZHTW: prize.Coupon.Name,
								JA:   prize.Coupon.Name,
								KO:   prize.Coupon.Name,
								MY:   prize.Coupon.Name,
								TR:   prize.Coupon.Name,
								SV:   prize.Coupon.Name,
							},
							RewardValue: prize.Coupon.Amount,
							RewardType:  0,
						})
					}
				}
				return prizes
			}(),
		})
	}

	return &member_resp.MemberMarketingActivityListsResp{
		List: memberMarketingActivityInfoResp,
	}, nil
}

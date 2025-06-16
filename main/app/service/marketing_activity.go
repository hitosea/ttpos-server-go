package service

import (
	"encoding/json"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/encrypt"
	"ttpos-server-go/pkg/utils"
)

type IMarketingActivitySrv interface {
	MarketingActivity(ctx context.Context) (*member_resp.MemberMarketingActivityListResp, error)  // 获取营销活动
	DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error) // 解密活动二维码
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
func (s *marketingActivitySrv) MarketingActivity(ctx context.Context) (*member_resp.MemberMarketingActivityListResp, error) {
	member := ctx.GetMember()
	company := ctx.GetCompany()
	// 获取营销活动列表
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivityList, err := marketingActivityRepo.GetActivityListByNow()
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
			QrCodeCode: qrCode,
			Desc:       marketingActivity.MultiLanguageDesc.GetNameByLang(ctx.GetLanguage()),
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
			Name:       "暂无营销活动",
			QrCodeCode: qrCode,
			Desc:       "暂无营销活动",
			StartTime:  0,
			EndTime:    0,
			IsInvalid:  1,
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

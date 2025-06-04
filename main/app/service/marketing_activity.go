package service

import (
	"encoding/json"
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
)

type IMarketingActivitySrv interface {
	MarketingActivity(ctx context.Context, req member_req.MemberMarketingActivityReq) (*member_resp.MemberMarketingActivityListResp, error) // 获取营销活动
	DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error)                                           // 解密活动二维码
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
func (s *marketingActivitySrv) MarketingActivity(ctx context.Context, req member_req.MemberMarketingActivityReq) (*member_resp.MemberMarketingActivityListResp, error) {
	member := ctx.GetMember()
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
			CompanyUuid:  req.CompanyUuid,
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
		})
	}
	// 返回
	return &member_resp.MemberMarketingActivityListResp{
		List: memberMarketingActivityResp,
		MemberInfo: member_resp.MemberInfoResp{
			Uuid:     member.Uuid,
			Nickname: member.Nickname,
			Phone:    member.Phone,
		},
	}, nil
}

// DecryptQrCode 解密活动二维码
func (s *marketingActivitySrv) DecryptQrCode(ctx context.Context, req req.DecryptQrCodeReq) (*resp.DecryptQrCodeResp, error) {
	decryptQrCodeResp, err := encrypt.DecryptAesString(req.Sign)
	if err != nil {
		return nil, errors.WithMessage(err, "活动二维码无效")
	}
	// 转json
	qrCodeParams := &repository.QrCodeParams{}
	err = json.Unmarshal([]byte(decryptQrCodeResp), qrCodeParams)
	if err != nil {
		return nil, errors.WithMessage(err, "活动二维码无效")
	}
	// 获取活动
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivity, err := marketingActivityRepo.GetActivity(qrCodeParams.ActivityUuid)
	if err != nil || marketingActivity == nil {
		return nil, errors.WithMessage(err, "活动二维码无效")
	}
	// 获取会员
	memberRepo := repository.NewMemberRepo(ctx.GetDB())
	member, err := memberRepo.GetMemberByUuid(qrCodeParams.MemberUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "活动二维码无效")
	}
	// 返回
	return &resp.DecryptQrCodeResp{
		Uuid:         member.Uuid,
		Nickname:     member.Nickname,
		Phone:        member.Phone,
		ActivityUuid: marketingActivity.Uuid,
	}, nil
}

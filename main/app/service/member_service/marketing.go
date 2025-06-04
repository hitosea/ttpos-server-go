package member_service

import (
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

type IMarketingSrv interface {
	MarketingActivity(ctx context.Context, req member_req.MemberMarketingActivityReq) (member_resp.MemberMarketingActivityListResp, error) // 获取营销活动
}

type marketingSrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewMarketingSrv(
	dbm *database.DBManager,
	cache cache.Cache,
) IMarketingSrv {
	return NewMarketingSrvImpl(dbm, cache)
}

func NewMarketingSrvImpl(
	dbm *database.DBManager,
	cache cache.Cache,
) IMarketingSrv {
	return &marketingSrv{
		dbm:   dbm,
		cache: cache,
	}
}

// MarketingActivity 获取营销活动
func (s *marketingSrv) MarketingActivity(ctx context.Context, req member_req.MemberMarketingActivityReq) (member_resp.MemberMarketingActivityListResp, error) {
	member := ctx.GetMember()
	// 获取营销活动列表
	marketingActivityRepo := repository.NewMarketingActivityRepo(ctx.GetDB())
	marketingActivityList, err := marketingActivityRepo.GetActivityListByNow()
	if err != nil {
		return member_resp.MemberMarketingActivityListResp{}, err
	}
	// 生成二维码
	memberMarketingActivityResp := make([]member_resp.MemberMarketingActivityResp, 0)
	for _, marketingActivity := range marketingActivityList {
		qrCode, err := marketingActivityRepo.GenerateQrCode(&repository.QrCodeParams{
			Type:         uint64(marketingActivity.Type),
			MemberUuid:   member.Uuid,
			ActivityUuid: marketingActivity.Uuid,
		})
		if err != nil {
			return member_resp.MemberMarketingActivityListResp{}, err
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
	return member_resp.MemberMarketingActivityListResp{
		List: memberMarketingActivityResp,
		MemberInfo: member_resp.MemberInfoResp{
			Uuid:     member.Uuid,
			Nickname: member.Nickname,
			Phone:    member.Phone,
		},
	}, nil
}

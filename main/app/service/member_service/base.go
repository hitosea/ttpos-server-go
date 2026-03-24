package member_service

import (
	"errors"
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// IBaseSrv 会员基础信息相关服务接口
// 获取基础信息
type IBaseSrv interface {
	GetBaseInfo(ctx context.Context) (member_resp.MemberBaseInfoResp, error)                                            // 获取基础信息
	UpdateNickname(ctx context.Context, req member_req.MemberNicknameUpdateReq) (member_resp.MemberBaseInfoResp, error) // 修改会员昵称
}

// baseSrv 会员基础信息服务实现
type baseSrv struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewBaseSrv 创建会员基础信息服务实例
func NewBaseSrv(
	dbm *database.DBManager,
	cache cache.Cache,
) IBaseSrv {
	return NewBaseSrvImpl(dbm, cache)
}

// NewBaseSrvImpl 创建会员基础信息服务实现
func NewBaseSrvImpl(
	dbm *database.DBManager,
	cache cache.Cache,
) IBaseSrv {
	return &baseSrv{
		dbm:   dbm,
		cache: cache,
	}
}

// GetBaseInfo 获取基础信息
// 参数：ctx 上下文，req 基础信息请求
// 返回：基础信息响应，错误信息
func (s *baseSrv) GetBaseInfo(ctx context.Context) (member_resp.MemberBaseInfoResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if db == nil {
		return member_resp.MemberBaseInfoResp{}, errors.New("商家不存在")
	}

	// 获取国家代码并转换为大写
	countryCode := strings.ToUpper(ctx.GetCfIPCountry())

	// 定义国家代码到区号的映射
	countryToAreaCode := map[string]string{
		"TH": constant.ThailandPrefix, // 泰国
		"CN": constant.ChinaPrefix,    // 中国
		"HK": constant.ChinaPrefix,    // 中国
	}

	// 默认区号列表
	areaCodes := []string{constant.ThailandPrefix, constant.ChinaPrefix}

	// 如果找到匹配的国家区号，将其移到第一位
	if areaCode, exists := countryToAreaCode[countryCode]; exists {
		// 从原数组中移除该区号
		for i, code := range areaCodes {
			if code == areaCode {
				areaCodes = append(areaCodes[:i], areaCodes[i+1:]...)
				break
			}
		}
		// 将匹配的区号放在第一位
		areaCodes = append([]string{areaCode}, areaCodes...)
	}

	// 获取商家信息
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(ctx.GetCompanyUuid())
	if err != nil {
		return member_resp.MemberBaseInfoResp{}, errors.New("商家不存在")
	}

	settingSrv := setting.NewSrv(s.dbm, s.cache)

	// 获取语言列表
	ctx.SetCompanyUuid(ctx.GetCompanyUuid())

	// 获取门店业务设置
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店业务设置失败", zap.Error(err))
		fmt.Println("获取门店业务设置失败", zap.Error(err))
	}

	// 获取收银机设置
	cashierSetting, err := settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取收银机设置失败", zap.Error(err))
		fmt.Println("获取收银机设置失败", zap.Error(err))
	}

	// 获取货币设置
	currencySetting, err := settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		logger.Logger.Error("获取货币设置失败", zap.Error(err))
		fmt.Println("获取货币设置失败", zap.Error(err))
	}

	// 获取扫码H5设置
	h5Setting, err := settingSrv.GetH5Setting(ctx, nil)
	if err != nil {
		logger.Logger.Error("获取H5设置失败", zap.Error(err))
		fmt.Println("获取H5设置失败", zap.Error(err))
	}

	// 获取门店点餐设置
	storeScanOrderSetting, err := settingSrv.GetStoreScanOrderSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店点餐设置失败", zap.Error(err))
	}

	// 获取模板样式设置
	templateStyleSetting, err := settingSrv.GetTemplateStyleSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取模板样式设置失败", zap.Error(err))
	}

	// 获取会员端语言设置
	// 云平台开启H5扫码时，优先跟随H5扫码语言设置；关闭时，跟随商家语言设置
	var (
		memberLanguageList []dto.LanguageItem
		memberLanguage     []string
		memberDefaultLang  string
	)
	if company.CompanySetting.IsOpenH5 == 1 {
		memberLanguageList = h5Setting.LanguageList
		memberLanguage = h5Setting.Language
		memberDefaultLang = h5Setting.DefaultLanguage
	} else {
		storeLanguageList, _ := settingSrv.GetStoreLanguageList(ctx)
		if len(storeLanguageList) == 0 {
			storeLanguageList = make([]dto.LanguageItem, 0)
		}
		memberLanguageList = storeLanguageList
		memberLanguage = make([]string, 0, len(storeLanguageList))
		for _, item := range storeLanguageList {
			memberLanguage = append(memberLanguage, item.Name)
		}
		if len(memberLanguage) > 0 {
			memberDefaultLang = memberLanguage[0]
		}
	}

	// 返回
	member := ctx.GetMember()
	return member_resp.MemberBaseInfoResp{
		User: member_resp.UserResp{
			Id:        member.ID,
			Uuid:      member.Uuid,
			Nickname:  member.Nickname,
			Phone:     member.Phone,
			Point:     member.GetPoints(),
			Balance:   member.GetBalanceAll(),
			IsVisitor: member.IsVisitor,
		},
		Member: member_resp.MemberResp{
			IsMemberShowSoldOut:  cashierSetting.MemberShowSoldOut == "1",
			LanguageList:         memberLanguageList,
			Language:             memberLanguage,
			DefaultLanguage:      memberDefaultLang,
			IsOpenRider:          storeScanOrderSetting.EnableDelivery == 1,
			IsOpenStoreScanOrder: storeScanOrderSetting.EnableSelfPickup == 1,
			DeliveryAvailable:    company.CompanySetting.IsOpenRider(),
			SelfPickupAvailable:  company.CompanySetting.IsOpenMemberInstant == 1,
			IsStoreResting:       storeScanOrderSetting.IsStoreResting(company.CompanySetting.Timezone, businessSetting.OpeningHours),
			IsOrderFirstPayLater: storeScanOrderSetting.IsOrderFirstPayLater == 1,
			AreaCode:             areaCodes,
		},
		Company: member_resp.CompanyResp{
			Uuid:         company.Uuid,
			Name:         company.Name,
			Logo:         company.GetLogo(utils.GetBaseURL(ctx.Copy().GetGin().Request)),
			Address:      company.CompanySetting.Address,
			LinkPhone:    company.CompanySetting.LinkPhone, // 公司联系电话
			OpeningHours: businessSetting.OpeningHours,     // 公司营业时间
		},
		Currency:      currencySetting,
		TemplateStyle: templateStyleSetting.TemplateStyle,
	}, nil

}

// UpdateNickname 修改会员昵称
// 参数：ctx 上下文，req 修改会员昵称请求
// 返回：错误信息
func (s *baseSrv) UpdateNickname(ctx context.Context, req member_req.MemberNicknameUpdateReq) (member_resp.MemberBaseInfoResp, error) {
	if err := req.Validate(); err != nil {
		return member_resp.MemberBaseInfoResp{}, err
	}
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	if db == nil {
		return member_resp.MemberBaseInfoResp{}, errors.New("商家不存在")
	}
	member := ctx.GetMember()
	err := repository.NewMemberRepo(db).Update(member.Uuid, map[string]any{
		"nickname": req.Nickname,
	})
	if err != nil {
		return member_resp.MemberBaseInfoResp{}, err
	}
	// 更新缓存
	member.Nickname = req.Nickname
	ctx.SetMember(member)
	// 更新缓存
	return s.GetBaseInfo(ctx)
}

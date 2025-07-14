package member_service

import (
	"errors"
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// IBaseSrv 会员基础信息相关服务接口
// 获取基础信息
type IBaseSrv interface {
	GetBaseInfo(ctx context.Context) (member_resp.MemberBaseInfoResp, error) // 获取基础信息
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

	// 获取语言列表
	ctx.SetCompanyUuid(ctx.GetCompanyUuid())
	languageList, _ := setting.NewSrv(s.dbm, s.cache).GetStoreLanguageList(ctx)

	// 获取门店业务设置
	businessSetting, err := setting.NewSrv(s.dbm, s.cache).GetBusinessSetting(ctx)
	if err != nil {
		logger.Logger.Error("获取门店业务设置失败", zap.Error(err))
		fmt.Println("获取门店业务设置失败", zap.Error(err))
	}

	member := ctx.GetMember()

	// 返回
	return member_resp.MemberBaseInfoResp{
		Member: member_resp.MemberResp{
			Id:        uint64(member.ID),
			Uuid:      member.Uuid,
			Nickname:  member.Nickname,
			Phone:     member.Phone,
			Point:     member.Point,
			Balance:   member.Balance,
			IsVisitor: member.IsVisitor,
		},
		Company: member_resp.CompanyResp{
			Uuid:         company.Uuid,
			Name:         company.Name,
			Logo:         company.Logo,
			Address:      company.CompanySetting.Address,
			LinkPhone:    company.CompanySetting.LinkPhone, // 公司联系电话
			OpeningHours: businessSetting.OpeningHours,     // 公司营业时间
		},
		AreaCode:     areaCodes,
		LanguageList: languageList,
	}, nil

}

package member_service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
)

const (
	CodeCacheKey = "member:login:code:%s:%d"
	CodeCacheTTL = 5 * time.Minute
)

// ILoginSrv 会员登录相关服务接口
// 获取登录信息、发送验证码、登录
type ILoginSrv interface {
	GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) // 获取登录信息
	SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                         // 发送验证码
	Login(ctx context.Context, req member_req.MemberLoginReq) (resp.LoginResp, error)                             // 登录
}

// loginSrv 会员登录服务实现
type loginSrv struct {
	dbm        *database.DBManager
	cache      cache.Cache
	smsSrv     service.ISmsSrv
	settingSrv setting.ISrv
}

// NewLoginSrv 创建会员登录服务实例
func NewLoginSrv(
	dbm *database.DBManager,
	cache cache.Cache,
	smsSrv service.ISmsSrv,
	settingSrv setting.ISrv,
) ILoginSrv {
	return NewLoginSrvImpl(dbm, cache, smsSrv, settingSrv)
}

// NewLoginSrvImpl 创建会员登录服务实现
func NewLoginSrvImpl(
	dbm *database.DBManager,
	cache cache.Cache,
	smsSrv service.ISmsSrv,
	settingSrv setting.ISrv,
) ILoginSrv {
	return &loginSrv{
		dbm:        dbm,
		cache:      cache,
		smsSrv:     smsSrv,
		settingSrv: settingSrv,
	}
}

// GetLoginInfo 获取登录信息，用于获取区号列表
// 参数：ctx 上下文，req 登录信息请求
// 返回：登录信息响应，错误信息
func (s *loginSrv) GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) {
	db := s.dbm.GetDB(req.CompanyUuid)
	if db == nil {
		return member_resp.MemberLoginInfoResp{}, errors.New("商家不存在")
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
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(req.CompanyUuid)
	if err != nil {
		return member_resp.MemberLoginInfoResp{}, errors.New("商家不存在")
	}

	// 获取商家设置
	// companySetting := repository.NewCompanySettingRepo(db).Get()
	// if companySetting.IsOpenMember != 1 {
	// 	return member_resp.MemberLoginInfoResp{}, errors.New("商家未开通会员功能")
	// }
	// if companySetting.IsOpenMarketing != 1 {
	// 	return member_resp.MemberLoginInfoResp{}, errors.New("二维码已失效")
	// }

	//
	ctx.SetCompanyUuid(req.CompanyUuid)
	languageList, _ := s.settingSrv.GetStoreLanguageList(ctx)

	// 返回
	return member_resp.MemberLoginInfoResp{
		CompanyUuid:  req.CompanyUuid,
		CompanyName:  company.Name,
		Logo:         utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.Copy().GetGin().Request), true),
		AreaCode:     areaCodes,
		LanguageList: languageList,
	}, nil
}

// SendCode 发送验证码
// 参数：ctx 上下文，req 发送验证码请求
// 返回：错误信息
func (s *loginSrv) SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	db := s.dbm.GetDB(req.CompanyUuid)
	if db == nil {
		return errors.New("商家不存在")
	}
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(req.CompanyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return errors.New("无法使用该功能，请联系商家")
	}
	companySetting := repository.NewCompanySettingRepo(db).Get()
	// if companySetting.IsOpenMember != 1 {
	// 	return errors.New("商家未开通会员功能")
	// }
	// if companySetting.IsOpenMarketing != 1 {
	// 	return errors.New("二维码已失效")
	// }
	// 验证手机号是否存在
	member, err := repository.NewMemberRepo(db).GetMemberByPhone(req.Phone)
	if err != nil || member.IsDelete() {
		return errors.New("该手机号未在该商家进行注册")
	}
	// 生成验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000)) // 生成6位随机数字验证码，范围：000000-999999
	// 如果是否debug模式，则打印验证码
	if config.Server.Mode == "debug" {
		fmt.Println("code", code)
	}
	// 设置缓存key
	cacheKey := fmt.Sprintf(CodeCacheKey, req.Phone, req.CompanyUuid)
	// 将验证码存储到缓存中，设置5分钟过期
	if err := s.cache.Set(cacheKey, code, CodeCacheTTL); err != nil {
		return fmt.Errorf("存储验证码失败: %v", err)
	}
	// 发送验证码短信
	ctx.SetCompanyUuid(req.CompanyUuid)
	ctx.SetCompany(*company)
	ctx.SetCompanySetting(companySetting)
	if err := s.smsSrv.SendMemberCodeSMS(ctx, req.Phone, &sms.MemberSendCodeRequest{
		Code: code,
	}); err != nil {
		return err
	}
	return nil
}

// Login 登录
// 参数：ctx 上下文，req 登录请求
// 返回：登录响应，错误信息
func (s *loginSrv) Login(ctx context.Context, req member_req.MemberLoginReq) (resp.LoginResp, error) {
	if err := req.Validate(); err != nil {
		return resp.LoginResp{}, err
	}
	db := s.dbm.GetDB(req.CompanyUuid)
	if db == nil {
		return resp.LoginResp{}, errors.New("商家不存在")
	}

	// 验证验证码
	cacheKey := fmt.Sprintf(CodeCacheKey, req.Phone, req.CompanyUuid)
	code, ok := s.cache.Get(cacheKey)
	if !ok || code == "" || code.(string) != req.Code {
		if config.Server.Mode != "debug" || req.Code != "123456" {
			return resp.LoginResp{}, errors.New("验证码不正确")
		}
	}
	s.cache.Del(cacheKey)

	// 验证手机号是否存在
	member, err := repository.NewMemberRepo(db).GetMemberByPhone(req.Phone)
	if err != nil || member.IsDelete() {
		return resp.LoginResp{}, errors.New("该手机号未在该商家进行注册")
	}
	// 生成token
	claims := auth.Claims{
		Source:      constant.SourceMember,
		CompanyUuid: req.CompanyUuid,
		MemberUuid:  member.Uuid,
	}
	token, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return resp.LoginResp{}, errors.New("生成token失败")
	}
	refreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return resp.LoginResp{}, errors.New("生成refresh_token失败")
	}
	return resp.LoginResp{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

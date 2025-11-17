package member_service

import (
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp/member_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	saas_model "ttpos-server-go/app/model/saas"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/saas"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"
)

// ILoginSrv 会员登录相关服务接口
// 获取登录信息、发送验证码、登录
type ILoginSrv interface {
	GetLoginInfo(ctx context.Context, req member_req.MemberLoginInfoReq) (member_resp.MemberLoginInfoResp, error) // 获取登录信息
	SendCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                         // 发送验证码
	SendRegisterCode(ctx context.Context, req member_req.MemberSendCodeReq) error                                 // 发送注册验证码
	Login(ctx context.Context, req member_req.MemberLoginReq) (member_resp.LoginResp, error)                      // 登录
	VisitorLogin(ctx context.Context, loginReq req.VisitorLoginReq) (*member_resp.LoginResp, error)               // 游客登录
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

	//
	ctx.SetCompanyUuid(req.CompanyUuid)
	languageList, _ := s.settingSrv.GetStoreLanguageList(ctx)

	// 返回
	return member_resp.MemberLoginInfoResp{
		CompanyUuid: req.CompanyUuid,
		CompanyName: company.Name,
		Logo: func() string {
			baseURL := utils.GetBaseURL(ctx.Copy().GetGin().Request)
			logoBase64, err := utils.AddImageDomainAndConvertToBase64(company.Logo, baseURL, true)
			if err != nil {
				return utils.AddImageDomain(company.Logo, baseURL, true)
			}
			return logoBase64
		}(),
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
	if company.CompanySetting == nil || company.CompanySetting.IsOpenMember != 1 {
		return errors.New("商家会员服务已关闭")
	}
	// 验证手机号是否存在
	member, err := repository.NewMemberRepo(db).GetMemberByPhoneContainDeleted(req.Phone)
	if err != nil {
		return errors.New("该手机号未在该商家进行注册")
	}
	if member.IsDelete() {
		return errors.New("该会员已被注销，可联系商家处理")
	}
	// 获取验证码
	code, err := validator.GetCode(s.cache, req.CompanyUuid, req.Phone)
	if err != nil {
		return err
	}
	// 发送验证码短信
	ctx.SetCompanyUuid(req.CompanyUuid)
	ctx.SetCompany(*company)
	ctx.SetCompanySetting(*company.CompanySetting)
	if err := s.smsSrv.SendMemberCodeSMS(ctx, req.Phone, &sms.MemberSendCodeRequest{
		Code: code,
	}); err != nil {
		return err
	}
	return nil
}

// SendRegisterCode 发送注册验证码
// 参数：ctx 上下文，req 发送验证码请求
// 返回：错误信息
func (s *loginSrv) SendRegisterCode(ctx context.Context, req member_req.MemberSendCodeReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	//
	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)
	// 获取商家信息
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return errors.New("无法使用该功能，请联系商家")
	}
	if company.CompanySetting == nil || company.CompanySetting.IsOpenMember != 1 {
		return errors.New("商家会员服务已关闭")
	}
	// 获取验证码
	code, err := validator.GetCode(s.cache, companyUuid, req.Phone)
	if err != nil {
		return err
	}
	// 发送验证码短信
	ctx.SetCompanyUuid(companyUuid)
	ctx.SetCompany(*company)
	ctx.SetCompanySetting(*company.CompanySetting)
	if err := s.smsSrv.SendMemberRegisterCodeSMS(ctx, req.Phone, &sms.MemberSendCodeRequest{
		Code: code,
	}); err != nil {
		return err
	}
	return nil
}

// Login 登录
// 参数：ctx 上下文，req 登录请求
// 返回：登录响应，错误信息
func (s *loginSrv) Login(ctx context.Context, req member_req.MemberLoginReq) (member_resp.LoginResp, error) {
	if err := req.Validate(); err != nil {
		return member_resp.LoginResp{}, err
	}
	db := s.dbm.GetDB(req.CompanyUuid)
	if db == nil {
		return member_resp.LoginResp{}, errors.New("商家不存在")
	}

	// 验证验证码
	if err := validator.VerifyCode(s.cache, req.CompanyUuid, req.Phone, req.Code); err != nil {
		return member_resp.LoginResp{}, err
	}

	// 获取商家信息
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(req.CompanyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return member_resp.LoginResp{}, errors.New("无法使用该功能，请联系商家")
	}
	if company.CompanySetting == nil || company.CompanySetting.IsOpenMember != 1 {
		return member_resp.LoginResp{}, errors.New("商家会员服务已关闭")
	}

	// 验证手机号是否存在
	member, err := repository.NewMemberRepo(db).GetMemberByPhone(req.Phone)
	if err != nil || member.IsDelete() {
		return member_resp.LoginResp{}, errors.New("该手机号未在该商家进行注册")
	}
	// 生成token
	claims := auth.Claims{
		Source:      constant.SourceMember,
		CompanyUuid: req.CompanyUuid,
		MemberUuid:  member.Uuid,
	}
	token, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return member_resp.LoginResp{}, errors.New("生成token失败")
	}
	refreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return member_resp.LoginResp{}, errors.New("生成refresh_token失败")
	}
	return member_resp.LoginResp{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// VisitorLogin 游客登录
// 参数：ctx 上下文，loginReq 游客登录请求
// 返回：游客信息响应，错误信息
func (s *loginSrv) VisitorLogin(ctx context.Context, loginReq req.VisitorLoginReq) (*member_resp.LoginResp, error) {
	companyUuid := loginReq.CompanyUuid
	db := s.dbm.GetDB(companyUuid)
	if db == nil {
		return nil, errors.New("商家不存在")
	}

	ctx.SetDB(db)
	ctx.SetCompanyUuid(companyUuid)
	deviceId := ctx.GetDeviceSn()

	// 检查是否已存在相同设备ID的游客
	member, err := repository.NewMemberRepo(db).GetVisitorByDeviceId(deviceId)
	if err != nil {
		return nil, errors.WithMessage(err, "查询游客失败")
	}

	// 不存在，创建游客
	if member.ID == 0 {

		// 生成随机昵称
		nickname := service.NewMemberSrv(s.dbm, s.cache).GenerateRandomNickname()

		// 事务开始
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 获取默认会员等级
		defaultLevelUuid := repository.NewMemberRepo(tx).GetMemberLevelMinPriorityUuid()
		if defaultLevelUuid == 0 {
			return nil, errors.New("未找到默认会员等级")
		}

		// 创建游客会员
		member = &model.Member{
			MemberNo:        utils.RandomNumber(5), // 5位数字
			Nickname:        nickname,
			Gender:          constant.MemberGenderUnknown,
			Phone:           "",
			IsVisitor:       true,
			DeviceId:        deviceId,
			Password:        "",
			MemberLevelUuid: defaultLevelUuid,
		}

		if err := repository.NewMemberRepo(tx).CreateMember(member); err != nil {
			return nil, errors.WithMessage(err, "创建游客失败")
		}

		if err := saas.NewMemberRepo(s.dbm.GetDB(constant.DefaultDB)).CreateMember(&saas_model.Member{
			CompanyUuid: companyUuid,
			Nickname:    nickname,
			IsVisitor:   true,
			DeviceId:    deviceId,
		}); err != nil {
			return nil, errors.WithMessage(err, "创建游客失败")
		}

		if err := tx.Commit().Error; err != nil {
			return nil, errors.WithMessage(err, "创建游客失败")
		}
	}

	// 生成token
	claims := auth.Claims{
		Source:      constant.SourceMember,
		CompanyUuid: companyUuid,
		MemberUuid:  member.Uuid,
	}
	token, err := auth.GenerateToken(claims, config.JWT.Secret, 3155673600, false)
	if err != nil {
		return nil, errors.New("生成token失败")
	}
	refreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, 3155673600, true)
	if err != nil {
		return nil, errors.New("生成refresh_token失败")
	}
	return &member_resp.LoginResp{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

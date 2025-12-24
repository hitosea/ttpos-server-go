package service

import (
	"regexp"
	"slices"
	"time"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type IAuthSrv interface {
	Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error)                                              // 登录
	SaasLogin(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error)                                          // 统一认证登录
	StoreSwitch(ctx context.Context, switchReq req.StoreSwitchReq) (resp.LoginResp, error)                                 // 门店切换
	Logout(ctx context.Context) error                                                                                      // 退出登录
	CashierBase(ctx context.Context) (resp.CashierBase, error)                                                             // 收银端基本信息
	AssistantBase(ctx context.Context) (resp.AssistantBase, error)                                                         // 点餐助手端基本信息
	TabletBase(ctx context.Context) (resp.TabletBase, error)                                                               // 平板端基本信息
	KitchenBase(ctx context.Context) (resp.KitchenBase, error)                                                             // 厨显端基本信息
	KioskBase(ctx context.Context) (resp.KioskBase, error)                                                                 // 自助点餐机基本信息
	Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, model.Desk, error) // 鉴权
	AuthDesk(ctx context.Context, qrcodeToken string) (*model.Company, error)                                              // 鉴权桌台
	AuthMenu(ctx context.Context, qrcodeToken string) (*model.Company, error)                                              // 鉴权点子菜单
	BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, string, error)                                   // 点餐助手绑定收银机
	GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList                                                           // 获取在线收银机
	RefreshToken(ctx context.Context) (resp.LoginResp, error)                                                              // 刷新token
	ShopBase(ctx context.Context) (resp.ShopBase, error)                                                                   // 移动管理端基本信息
	ChangePassword(ctx context.Context, changePasswordReq req.ChangePasswordReq) error                                     // 移动管理端-修改密码
}

func NewAuthSrv(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv settingSrv.ISrv,
) IAuthSrv {
	return NewAuthSrvImpl(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
}

type authSrv struct {
	dbm           *database.DBManager
	captchaSrv    ICaptchaSrv
	roleAccessSrv IRoleAccessSrv
	deviceSrv     IDeviceSrv
	shiftSrv      IStaffShiftSrv
	settingSrv    settingSrv.ISrv

	assistantRoutes             []string
	tabletRoutes                []string
	memberFunctionRoutes        []string
	h5AcceptOrderFunctionRoutes []string
}

func NewAuthSrvImpl(
	dbm *database.DBManager,
	captchaSrv ICaptchaSrv,
	roleAccessSrv IRoleAccessSrv,
	deviceSrv IDeviceSrv,
	staffShiftSrv IStaffShiftSrv,
	settingSrv settingSrv.ISrv,
) IAuthSrv {
	return &authSrv{
		dbm:           dbm,
		captchaSrv:    captchaSrv,
		roleAccessSrv: roleAccessSrv,
		deviceSrv:     deviceSrv,
		shiftSrv:      staffShiftSrv,
		settingSrv:    settingSrv,

		assistantRoutes: []string{ // 已登录点餐助手，但未绑定收银机可以不判断收银机状态的接口
			"/api/v1/assistant/online_cashiers",
			"/api/v1/assistant/bind_cashier",
			"/api/v1/assistant/verify_lock_password",
		},
		tabletRoutes: []string{
			"/api/v1/tablet/desk/list",
			"/api/v1/tablet/desk/bind",
			"/api/v1/tablet/base",
			"/api/v1/tablet/logout",
			"/api/v1/tablet/check_update",
			"/api/v1/tablet/verify_advanced_password",
			"/api/v1/tablet/product/category/list",
			"/api/v1/tablet/product/list",
			"/api/v1/tablet/buffet/list",
		},
		memberFunctionRoutes: []string{
			"/api/v1/cashier/instant/order/member/confirm",  // 收银端使用会员优惠
			"/api/v1/assistant/member/add",                  // 助手端添加会员
			"/api/v1/cashier/member/add",                    // 收银端添加会员
			"/api/v1/cashier/member/create_recharge_order",  // 收银端创建会员充值单
			"/api/v1/cashier/member/confirm_recharge_order", // 收银端确认会员充值订单
			"/api/v1/assistant/desk/order/member/confirm",   // 助手端使用会员优惠
			"/api/v1/cashier/desk/order/member/confirm",     // 收银端使用会员优惠
			"/api/v1/cashier/recharge_order/list",           // 收银端订单列表
		},
		h5AcceptOrderFunctionRoutes: []string{
			"/api/v1/cashier/h5_order/list",   // h5订单列表
			"/api/v1/cashier/h5_order/detail", // h5订单详情
			"/api/v1/cashier/h5_order/reject", // h5拒单
			"/api/v1/cashier/h5_order/accept", // h5接单
		},
	}
}

// Login 登录
func (s *authSrv) Login(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error) {
	var loginResp resp.LoginResp
	var token, refreshToken string
	// 验证验证码
	if !s.captchaSrv.Verify(ctx.GetGin().GetHeader("X-SIGN"), loginReq.Code) && viper.GetString("GENERAL_VERIFY_CODE") != loginReq.Code {
		return loginResp, errors.New("验证码错误")
	}
	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		companyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsername(loginReq.Username))
		if companyStaff.CompanyUuid == 0 {
			return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
		}
		if companyStaff.Uuid == 0 {
			return loginResp, errors.New("账号或密码错误")
		}
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyStaff.CompanyUuid))
		staff, _ = staffRepo.GetStaff(staffRepo.WhereUuid(companyStaff.Uuid), staffRepo.WithCompany())
	} else { // 离线版本
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(constant.DefaultDB))
		staff, _ = staffRepo.GetStaff(staffRepo.WhereUsername(loginReq.Username), staffRepo.WithCompany())
	}
	// 商家状态
	if staff.Company == nil {
		return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
	}

	// 验证密码（支持 MD5 和 bcrypt）
	if staff.Uuid == 0 {
		return loginResp, errors.New("账号或密码错误")
	}
	isValid, needUpgrade := utils.VerifyPassword(loginReq.Password, staff.Password)
	if !isValid {
		return loginResp, errors.New("账号或密码错误")
	}

	// 如果需要升级密码，异步升级为 bcrypt
	if needUpgrade {
		utils.UpgradePasswordAsync(
			s.dbm.GetDB(staff.CompanyUuid),
			"ttpos_staff",
			"password",
			"uuid",
			staff.Uuid,
			loginReq.Password,
		)
	}
	// 检查员工状态
	if staff.DeleteTime != 0 {
		return loginResp, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return loginResp, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company.IsExpired() {
		return loginResp, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if staff.Company.IsException() {
		return loginResp, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	isFirstLogin := staff.CashierOnline == 0

	switch loginReq.Source {
	case constant.SourceCashier: // 收银端登录
		// 判断权限
		permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
		if err != nil {
			return loginResp, errors.WithMessage(err)
		}
		if len(permissions) == 0 {
			return loginResp, errors.New("当前无权限，请联系管理员")
		}

		// 获取员工仓库
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))

		// 检查是否有未交班的收银员
		if viper.GetString("CHECK_SHIFT_HANDOVER") != "false" {
			currentStaff, _ := staffRepo.GetStaff(staffRepo.WhereDeviceId(loginReq.DeviceId), staffRepo.WhereCashierOnline())
			if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
				return loginResp, errors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.GetUserName()})
			}
			// 是否已在其他收银机登录
			if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
				return loginResp, errors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{staff.GetUserName()})
			}
		}

		// 更新员工信息
		updates := map[string]any{
			"cashier_online": 1,
			"bind_key":       loginReq.DeviceId,
		}
		// 创建当班日志
		if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
			companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
			ctx.SetCompany(*staff.Company)
			ctx.SetCompanySetting(companySetting)
			shiftLog, err := s.shiftSrv.CreateWorkingLog(ctx, staff)
			if err != nil {
				return loginResp, errors.WithMessage(err, "创建当班日志失败")
			}
			updates["cashier_login_time"] = shiftLog.ShiftStartTime
			updates["duty_no"] = shiftLog.ShiftNo
		}
		err = staffRepo.Update(staff.Uuid, updates)
		if err != nil {
			return loginResp, errors.WithMessage(err, "更新信息失败")
		}
	case constant.SourceAssistant: // 点餐助手登录
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenAssistant != 1 {
			return loginResp, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
		}
	case constant.SourceKitchen: // 厨显端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
		if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
			return loginResp, errors.New("当前尚未开启厨显功能，如有需要，请联系销售代表")
		}
	case constant.SourceTablet: // 平板端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenTablet != 1 {
			return loginResp, errors.New("当前尚未开启平板点餐功能，如有需要，请联系销售代表")
		}
	case constant.SourceKiosk: // 自助点餐机
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if !companySetting.IsOpenKiosk() {
			return loginResp, errors.New("当前尚未开启自助点餐机功能，如有需要，请联系销售代表")
		}
	case constant.SourceShop: // 移动管理端
		// 权限获取在 shop_auth.go 中完成，这里不需要处理
	default:
		return loginResp, errors.New("登录来源错误")
	}

	// 登录时没有商家ID，补上
	ctx.SetCompanyUuid(staff.CompanyUuid)
	// 添加绑定记录
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         loginReq.DeviceId,
		Brand:            loginReq.Brand,
		Source:           loginReq.Source,
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: time.Now().Unix(),
		CompanyUuid:      staff.CompanyUuid,
	})
	if err != nil {
		return loginResp, errors.WithMessage(err)
	}

	claims := auth.Claims{
		Source:      loginReq.Source,
		CompanyUuid: staff.CompanyUuid,
		StaffUuid:   staff.Uuid,
		DeviceUuid:  deviceUuid,
		DeviceId:    loginReq.DeviceId,
		Assistant:   auth.Assistant{},
	}
	// 生成 JWT token，refresh_token
	token, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	refreshToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return loginResp, errors.New("生成refresh_token失败")
	}
	return resp.LoginResp{
		Token:               token,
		RefreshToken:        refreshToken,
		CashierIsFirstLogin: isFirstLogin,
		NeedChangePassword:  staff.PasswordChangeCount == 0,
	}, nil
}

// SaasLogin 统一认证登录
func (s *authSrv) SaasLogin(ctx context.Context, loginReq req.LoginReq) (resp.LoginResp, error) {
	var loginResp resp.LoginResp

	// 1. 验证验证码
	if !s.captchaSrv.Verify(ctx.GetGin().GetHeader("X-SIGN"), loginReq.Code) &&
		viper.GetString("GENERAL_VERIFY_CODE") != loginReq.Code {
		return loginResp, errors.New("验证码错误")
	}

	// 2. 查询统一账号表 saas.ttpos_staff
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)

	var saasStaff *model.SaasStaff
	var err error

	// 尝试通过邮箱查询
	saasStaff, err = saasStaffRepo.GetByEmail(loginReq.Username)
	if err != nil || saasStaff == nil {
		// 尝试通过手机号查询
		saasStaff, err = saasStaffRepo.GetByPhone(loginReq.Username)
		if err != nil || saasStaff == nil {
			return loginResp, errors.New("账号或密码错误")
		}
	}

	// 验证密码（支持 MD5 和 bcrypt）
	isValid, needUpgrade := utils.VerifyPassword(loginReq.Password, saasStaff.Password)
	if !isValid {
		return loginResp, errors.New("账号或密码错误")
	}

	// 如果需要升级密码，异步升级为 bcrypt
	if needUpgrade {
		utils.UpgradePasswordAsync(
			saasDB,
			"ttpos_staff",
			"password",
			"uuid",
			saasStaff.Uuid,
			loginReq.Password,
		)
	}

	// 检查账号是否被禁用
	if saasStaff.IsDisable == 1 {
		return loginResp, errors.New("账号被禁用，请联系管理员")
	}

	// 3. 查询关联的商家列表（关联 saas.ttpos_company 表获取商家信息）
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 查询员工关联的商家列表（关联查询商家信息）
	companyStaffList, err := companyStaffRepo.GetByStaffUuid(saasStaff.Uuid, companyStaffRepo.WithCompany())
	if err != nil {
		return loginResp, errors.WithMessage(err, "获取门店列表失败")
	}

	// 4. 过滤商家：遍历每个商家，过滤掉已过期、异常的商家
	validCompanyList := make([]*model.CompanyStaff, 0)
	for _, cs := range companyStaffList {
		// 过滤条件：员工在该商家未被禁用且未删除
		if cs.IsDisable != 0 || cs.DeleteTime != 0 {
			continue
		}

		// 商家不存在，跳过
		if cs.Company == nil {
			continue
		}

		// 过滤已过期、异常的商家
		if cs.Company.IsExpired() || cs.Company.IsException() {
			continue
		}

		validCompanyList = append(validCompanyList, cs)
	}

	// 5. 判断商家数量（基于过滤后的商家列表）
	if len(validCompanyList) == 0 {
		return loginResp, errors.New("登录失败：你暂无该门店的操作权限，请联系门店管理员开通权限。")
	}

	// 6. 只有一个商家时，走原有逻辑
	if len(validCompanyList) == 1 {
		companyUuid := validCompanyList[0].CompanyUuid
		return s.loginWithCompany(ctx, loginReq, saasStaff.Uuid, companyUuid, saasStaff.PasswordChangeCount == 0)
	}

	// 7. 多个商家时的处理（根据登录来源进行不同处理）
	if loginReq.Source != constant.SourceShop {
		// 非新管理端登录：生成 company_uuid 为 0 的 token
		return s.generateTokenWithCompanyUuidZero(ctx, loginReq, saasStaff.Uuid, saasStaff.PasswordChangeCount == 0)
	}

	// 新管理端登录：检查 last_company_uuid
	if saasStaff.LastCompanyUuid > 0 {
		// 检查 last_company_uuid 是否在过滤后的关联商家列表中
		for _, cs := range validCompanyList {
			if cs.CompanyUuid == saasStaff.LastCompanyUuid {
				// 从该商家数据库查询员工信息，走原有逻辑
				return s.loginWithCompany(ctx, loginReq, saasStaff.Uuid, saasStaff.LastCompanyUuid, saasStaff.PasswordChangeCount == 0)
			}
		}
	}

	// last_company_uuid 无效或不在关联列表中，生成 company_uuid 为 0 的 token
	return s.generateTokenWithCompanyUuidZero(ctx, loginReq, saasStaff.Uuid, saasStaff.PasswordChangeCount == 0)
}

// loginWithCompany 从指定商家数据库查询员工信息，走原有 Login 逻辑
func (s *authSrv) loginWithCompany(ctx context.Context, loginReq req.LoginReq, staffUuid, companyUuid uint64, needChangePassword bool) (resp.LoginResp, error) {
	// 从商家数据库查询员工信息
	companyDB := s.dbm.GetDB(companyUuid)
	if companyDB == nil {
		return resp.LoginResp{}, errors.New("未找到绑定的商家，请确认登录信息")
	}

	staffRepo := repository.NewStaffRepo(companyDB)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(staffUuid), staffRepo.WithCompany())
	if err != nil || staff.Uuid == 0 {
		return resp.LoginResp{}, errors.New("账号或密码错误")
	}

	// 验证商家状态
	if staff.Company == nil {
		return resp.LoginResp{}, errors.New("未找到绑定的商家，请确认登录信息")
	}
	if staff.Company.IsExpired() {
		return resp.LoginResp{}, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if staff.Company.IsException() {
		return resp.LoginResp{}, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}

	// 检查员工状态
	if staff.DeleteTime != 0 {
		return resp.LoginResp{}, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return resp.LoginResp{}, errors.New("账号被禁用，请联系管理员")
	}

	isFirstLogin := staff.CashierOnline == 0

	// 根据 source 进行不同的权限验证和处理（参考原有 Login 方法）
	switch loginReq.Source {
	case constant.SourceCashier: // 收银端登录
		// 判断权限
		permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
		if err != nil {
			return resp.LoginResp{}, errors.WithMessage(err)
		}
		if len(permissions) == 0 {
			return resp.LoginResp{}, errors.New("当前无权限，请联系管理员")
		}

		// 检查是否有未交班的收银员
		if viper.GetString("CHECK_SHIFT_HANDOVER") != "false" {
			currentStaff, _ := staffRepo.GetStaff(staffRepo.WhereDeviceId(loginReq.DeviceId), staffRepo.WhereCashierOnline())
			if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
				return resp.LoginResp{}, errors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.GetUserName()})
			}
			// 是否已在其他收银机登录
			if staff.CashierOnline == 1 && loginReq.DeviceId != staff.BindKey {
				return resp.LoginResp{}, errors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{staff.GetUserName()})
			}
		}

		// 更新员工信息
		updates := map[string]any{
			"cashier_online": 1,
			"bind_key":       loginReq.DeviceId,
		}
		// 创建当班日志
		if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
			companySetting := repository.NewCompanySettingRepo(companyDB).Get()
			ctx.SetCompany(*staff.Company)
			ctx.SetCompanySetting(companySetting)
			shiftLog, err := s.shiftSrv.CreateWorkingLog(ctx, staff)
			if err != nil {
				return resp.LoginResp{}, errors.WithMessage(err, "创建当班日志失败")
			}
			updates["cashier_login_time"] = shiftLog.ShiftStartTime
			updates["duty_no"] = shiftLog.ShiftNo
		}
		err = staffRepo.Update(staff.Uuid, updates)
		if err != nil {
			return resp.LoginResp{}, errors.WithMessage(err, "更新信息失败")
		}
	case constant.SourceAssistant: // 点餐助手登录
		companySetting := repository.NewCompanySettingRepo(companyDB).Get()
		if companySetting.IsOpenAssistant != 1 {
			return resp.LoginResp{}, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
		}
	case constant.SourceKitchen: // 厨显端
		companySetting := repository.NewCompanySettingRepo(companyDB).Get()
		kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
		if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
			return resp.LoginResp{}, errors.New("当前尚未开启厨显功能，如有需要，请联系销售代表")
		}
	case constant.SourceTablet: // 平板端
		companySetting := repository.NewCompanySettingRepo(companyDB).Get()
		if companySetting.IsOpenTablet != 1 {
			return resp.LoginResp{}, errors.New("当前尚未开启平板点餐功能，如有需要，请联系销售代表")
		}
	case constant.SourceKiosk: // 自助点餐机
		companySetting := repository.NewCompanySettingRepo(companyDB).Get()
		if !companySetting.IsOpenKiosk() {
			return resp.LoginResp{}, errors.New("当前尚未开启自助点餐机功能，如有需要，请联系销售代表")
		}
	case constant.SourceShop: // 移动管理端
		// 权限获取在 shop_auth.go 中完成，这里不需要处理
	default:
		return resp.LoginResp{}, errors.New("登录来源错误")
	}

	// 登录时没有商家ID，补上
	ctx.SetCompanyUuid(staff.CompanyUuid)
	// 添加绑定记录
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         loginReq.DeviceId,
		Brand:            loginReq.Brand,
		Source:           loginReq.Source,
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: time.Now().Unix(),
		CompanyUuid:      staff.CompanyUuid,
	})
	if err != nil {
		return resp.LoginResp{}, errors.WithMessage(err)
	}

	claims := auth.Claims{
		Source:      loginReq.Source,
		CompanyUuid: companyUuid,
		StaffUuid:   staffUuid,
		DeviceUuid:  deviceUuid,
		DeviceId:    loginReq.DeviceId,
		Assistant:   auth.Assistant{},
	}

	// 生成 JWT token，refresh_token
	token, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return resp.LoginResp{}, errors.New("生成token失败")
	}
	refreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return resp.LoginResp{}, errors.New("生成refresh_token失败")
	}

	return resp.LoginResp{
		Token:               token,
		RefreshToken:        refreshToken,
		CashierIsFirstLogin: isFirstLogin,
		NeedChangePassword:  needChangePassword,
	}, nil
}

// generateTokenWithCompanyUuidZero 生成 company_uuid 为 0 的 token
func (s *authSrv) generateTokenWithCompanyUuidZero(_ context.Context, loginReq req.LoginReq, staffUuid uint64, needChangePassword bool) (resp.LoginResp, error) {
	claims := auth.Claims{
		Source:      loginReq.Source,
		CompanyUuid: 0, // company_uuid 为 0，需要切换门店后设置
		StaffUuid:   staffUuid,
		DeviceUuid:  0,
		DeviceId:    loginReq.DeviceId,
		Assistant:   auth.Assistant{},
		Brand:       loginReq.Brand,
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
		Token:              token,
		RefreshToken:       refreshToken,
		NeedChangePassword: needChangePassword,
	}, nil
}

// Logout 退出登录
func (s *authSrv) Logout(ctx context.Context) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 解绑设备
	err := repository.NewDeviceRepo(db).UpdateDevice(ctx.GetDeviceUuid(), map[string]any{
		"finally_login_uuid": 0,
	})
	if err != nil {
		return errors.ErrInternal
	}
	if ctx.GetSource() == constant.SourceTablet {
		// 解绑桌台
		err = repository.NewDeskRepo(db).UnbindDesk(ctx.GetDeskUuid(), ctx.GetDeviceUuid())
		if err != nil {
			return errors.ErrInternal
		}
	}
	return nil
}

// CashierBase 获取收银端基本信息
func (s *authSrv) CashierBase(ctx context.Context) (resp.CashierBase, error) {
	var cashierBase resp.CashierBase

	// 如果 company_uuid 为 0，只返回可用门店列表
	if ctx.GetCompanyUuid() == 0 {
		cashierBase.CompanyList = s.getCompanyList(ctx)
		return cashierBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	staff := ctx.GetStaff()
	var (
		source             = ctx.GetSource()
		deviceId           = ctx.GetGin().GetString(jwt.DeviceId)
		cashierSettingResp setting.CashierResp
	)
	deviceRemark := s.deviceSrv.GetRemark(company.Uuid, source, deviceId)
	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	if len(permissions) == 0 {
		return cashierBase, errors.New("当前无权限，请联系管理员")
	}
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	copier.Copy(&cashierSettingResp, cashierSetting)
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return cashierBase, errors.WithMessage(err)
	}
	return resp.CashierBase{
		Username:     staff.Username,
		CashierUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Cashier:      cashierSettingResp,
		Business:     businessSetting,
		Buffet:       buffetSetting,
		Currency:     currencySetting,
		Permissions:  permissions,
		Company: resp.Company{ // cashier
			Uuid:                 company.Uuid,
			Name:                 company.Name,
			Logo:                 utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:             companySetting.Timezone,
			ExpireTime:           company.ExpireTime,
			IsOpenMember:         companySetting.IsOpenMember,
			IsOpenBuffet:         companySetting.IsOpenBuffet,
			IsOpenH5Order:        companySetting.IsOpenH5Order,
			IsOpenRider:          companySetting.IsOpenRider(),
			IsOpenOldOrder:       utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsEnableErp:          company.IsOpenErp(),
			IsOpenMap:            companySetting.IsOpenTableMap(),
			IsOpenDataManagement: companySetting.IsOpenDataManagement(),
			IsOpenGrabDelivery:   companySetting.IsOpenGrabDelivery(),
		},
		CloudBasic: cloudBasicSetting,
		Printer:    printerSetting,
		UpdateTime: time.Now().Unix(),

		CompanyList: s.getCompanyList(ctx),
	}, nil
}

// AssistantBase 获取助手端基本信息
func (s *authSrv) AssistantBase(ctx context.Context) (resp.AssistantBase, error) {
	assistantBase := resp.AssistantBase{}

	// 如果 company_uuid 为 0，只返回可用门店列表
	if ctx.GetCompanyUuid() == 0 {
		assistantBase.CompanyList = s.getCompanyList(ctx)
		return assistantBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	staff := helper.GetStaff(ctx.GetGin())
	staffRepo := repository.NewStaffRepo(db)
	assistantStaffUuid := ctx.GetGin().GetUint64(jwt.AssistantStaffUuid)
	assistantStaff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(assistantStaffUuid))
	perms, err := s.roleAccessSrv.GetPermission(constant.AssistantRouteName, assistantStaff.Uuid, assistantStaff.CompanyUuid)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	permissions := make([]string, 0, len(perms))
	for _, perm := range perms {
		permissions = append(permissions, perm.Path)
	}

	deviceRepo := repository.NewDeviceRepo(db)
	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()))
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	assistantSetting, err := s.settingSrv.GetAssistantSetting(ctx, nil)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	var assistantSettingResp setting.AssistantResp
	copier.CopyWithOption(&assistantSettingResp, assistantSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	printerSetting, err := s.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	var kitchenSettingResp setting.KitchenResp
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
	if err != nil {
		return assistantBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.AssistantBase{
		Permissions: permissions,
		CashierStaff: resp.CashierStaff{
			RealName:     staff.RealName,
			Username:     staff.Username,
			DeviceId:     staff.BindKey,
			DeviceRemark: device.Remark,
		},
		AssistantStaff: resp.AssistantStaff{
			Uuid:     assistantStaff.Uuid,
			RealName: assistantStaff.RealName,
			Phone:    assistantStaff.Phone,
			DeviceId: assistantStaff.BindKey,
		},
		Buffet:     buffetSetting,
		CloudBasic: cloudBasicSetting,
		Company: resp.Company{ // assistant
			Uuid:                 company.Uuid,
			Name:                 company.Name,
			Logo:                 utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:             companySetting.Timezone,
			ExpireTime:           company.ExpireTime,
			IsOpenMember:         companySetting.IsOpenMember,
			IsOpenBuffet:         companySetting.IsOpenBuffet,
			IsOpenH5Order:        companySetting.IsOpenH5Order,
			IsOpenOldOrder:       utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:          companySetting.IsOpenRider(),
			IsEnableErp:          company.IsOpenErp(),
			IsOpenMap:            companySetting.IsOpenTableMap(),
			IsOpenDataManagement: companySetting.IsOpenDataManagement(),
		},
		Currency:      currencySetting,
		Business:      businessSetting,
		Assistant:     assistantSettingResp,
		Printer:       printerSetting,
		Kitchen:       kitchenSettingResp,
		ClientVersion: clientVersion,
		ServerVersion: utils.GetVersion(),
		CompanyList:   s.getCompanyList(ctx),
	}, nil
}

// TabletBase 平板端基本信息
func (s *authSrv) TabletBase(ctx context.Context) (resp.TabletBase, error) {
	var tabletBase resp.TabletBase

	// 如果 company_uuid 为 0，只返回可用门店列表
	if ctx.GetCompanyUuid() == 0 {
		tabletBase.CompanyList = s.getCompanyList(ctx)
		return tabletBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	var (
		tabletSettingResp  setting.TabletResp
		kitchenSettingResp setting.KitchenResp
	)
	tabletSetting, err := s.settingSrv.GetTabletSetting(ctx, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	copier.Copy(&tabletSettingResp, tabletSetting)
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return tabletBase, errors.WithMessage(err)
	}
	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.TabletBase{
		RealName:      ctx.GetStaff().RealName,
		ServerVersion: utils.GetVersion(),
		ClientVersion: clientVersion,
		Buffet:        buffetSetting,
		CloudBasic:    cloudBasicSetting,
		Company: resp.Company{ // tablet
			Uuid:               company.Uuid,
			Name:               company.Name,
			Logo:               utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:           companySetting.Timezone,
			ExpireTime:         company.ExpireTime,
			IsOpenMember:       companySetting.IsOpenMember,
			IsOpenBuffet:       companySetting.IsOpenBuffet,
			IsOpenH5Order:      companySetting.IsOpenH5Order,
			IsOpenOldOrder:     utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:        companySetting.IsOpenRider(),
			IsEnableErp:        company.IsOpenErp(),
			IsOpenGrabDelivery: companySetting.IsOpenGrabDelivery(),
		},
		Currency: currencySetting,
		Business: businessSetting,
		Tablet:   tabletSettingResp,
		Kitchen:  kitchenSettingResp,

		CompanyList: s.getCompanyList(ctx),
	}, nil
}

// KitchenBase 获取厨显端基本信息
func (s *authSrv) KitchenBase(ctx context.Context) (resp.KitchenBase, error) {
	var (
		kitchenBase        resp.KitchenBase
		kitchenSettingResp setting.KitchenResp
	)

	// 如果 company_uuid 为 0，返回可用门店列表
	if ctx.GetCompanyUuid() == 0 {
		kitchenBase.CompanyList = s.getCompanyList(ctx)
		return kitchenBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	company := helper.GetCompany(ctx.GetGin())
	companySetting := helper.GetCompanySetting(ctx.GetGin())
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, companySetting, nil)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	copier.CopyWithOption(&kitchenSettingResp, kitchenSetting, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return kitchenBase, errors.WithMessage(err)
	}

	clientVersion := ctx.GetGin().GetHeader("Version-Name")
	if clientVersion == "" {
		clientVersion = "0.0.0"
	}
	return resp.KitchenBase{
		RealName:   ctx.GetStaff().RealName,
		Buffet:     buffetSetting,
		CloudBasic: cloudBasicSetting,
		Company: resp.Company{ // kitchen
			Uuid:           company.Uuid,
			Name:           company.Name,
			Logo:           utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:       companySetting.Timezone,
			ExpireTime:     company.ExpireTime,
			IsOpenMember:   companySetting.IsOpenMember,
			IsOpenBuffet:   companySetting.IsOpenBuffet,
			IsOpenH5Order:  companySetting.IsOpenH5Order,
			IsOpenOldOrder: utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:    companySetting.IsOpenRider(),
			IsEnableErp:    company.IsOpenErp(),
		},
		Currency:      currencySetting,
		Business:      businessSetting,
		Kitchen:       kitchenSettingResp,
		ServerVersion: utils.GetVersion(),
		ClientVersion: clientVersion,
		CompanyList:   s.getCompanyList(ctx),
	}, nil
}

// KioskBase 获取自助点餐机基本信息
func (s *authSrv) KioskBase(ctx context.Context) (resp.KioskBase, error) {
	var kioskBase resp.KioskBase

	// 如果 company_uuid 为 0，只返回可用门店列表
	// 设计原因：
	// 1. 支持统一账号认证的多门店切换场景：用户登录后可能未选择门店，需要返回可用门店列表供用户选择
	// 2. 保持与其他终端（Cashier、Tablet、Assistant等）Base 接口的一致性
	// 3. 前端可以根据 company_list 判断是否需要显示门店选择界面
	// 参考：story-auth-unified-account 统一账号认证功能设计
	if ctx.GetCompanyUuid() == 0 {
		kioskBase.CompanyList = s.getCompanyList(ctx)
		return kioskBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	staff := ctx.GetStaff()
	var (
		source   = ctx.GetSource()
		deviceId = ctx.GetGin().GetString(jwt.DeviceId)
	)
	deviceRemark := s.deviceSrv.GetRemark(company.Uuid, source, deviceId)

	// 获取自助点餐机设置（包含语言列表、轮播广告）
	kioskSetting, err := s.settingSrv.GetKioskSetting(ctx)
	if err != nil {
		return kioskBase, errors.WithMessage(err)
	}

	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return kioskBase, errors.WithMessage(err)
	}

	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return kioskBase, errors.WithMessage(err)
	}

	return resp.KioskBase{
		Username:     staff.Username,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Company: resp.Company{
			Uuid:       company.Uuid,
			Name:       company.Name,
			Logo:       utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:   companySetting.Timezone,
			ExpireTime: company.ExpireTime,
		},
		Currency:    currencySetting,
		Business:    businessSetting,
		Kiosk:       kioskSetting.KioskResp,
		UpdateTime:  time.Now().Unix(),
		CompanyList: s.getCompanyList(ctx),
	}, nil
}

// Auth 鉴权
func (s *authSrv) Auth(ctx context.Context, auth req.Authenticate) (model.Company, model.CompanySetting, model.Staff, model.Desk, error) {
	var (
		company        model.Company
		companySetting model.CompanySetting
		staff          model.Staff
		desk           model.Desk
		db             = s.dbm.GetDB(auth.CompanyUuid)
	)
	// 如果使用旧的token，但是商家已被删除，优化提示
	if db == nil {
		return company, companySetting, staff, desk, errors.New("未找到绑定的商家，请确认登录信息")
	}
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(auth.StaffUuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	if err != nil {
		logger.Logger.Error("获取员工信息失败", zap.Error(err))
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "没有找到用户信息")
	}
	if staff.Uuid == 0 {
		return company, companySetting, staff, desk, errors.New("用户不存在")
	}
	// 修改密码后，token失效
	if staff.PasswordChangeTime > auth.TokenIssuedAt {
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "登录失效，请重新登录")
	}
	if staff.DeleteTime != 0 {
		return company, companySetting, staff, desk, errors.New("用户被删除")
	}
	if staff.Company != nil {
		company = *staff.Company
		if staff.Company.CompanySetting != nil {
			companySetting = *staff.Company.CompanySetting
		}
	}
	if company.Uuid == 0 || companySetting.Uuid == 0 {
		return company, companySetting, staff, desk, errors.New("未找到绑定的商家，请确认登录信息")
	}
	if company.IsExpired() {
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if company.IsException() {
		return company, companySetting, staff, desk, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	// 验证设备是否绑定
	deviceId := auth.DeviceId
	if auth.Source == constant.SourceAssistant && auth.Assistant.DeviceId != "" { // 登录了点餐助手，且绑定了收银机
		deviceId = auth.Assistant.DeviceId
	}
	if auth.Source != constant.SourceShop && !s.deviceSrv.IsDeviceBind(ctx, auth.CompanyUuid, auth.Source, deviceId) {
		if printerData := repository.NewPrinterLogRepo(db).GetShiftPrinterData(auth.CompanyUuid, deviceId); printerData != nil {
			return company, companySetting, staff, desk, errors.NewWithCodeAndData(constant.CodeTokenInvalid, map[string]any{
				"printer_data": printerData,
			}, "设备已解绑，请重新绑定")
		}
		return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenInvalid, "设备已解绑，请重新绑定")
	}
	if slices.Contains(s.memberFunctionRoutes, auth.UrlPath) && companySetting.IsOpenMember != 1 {
		return company, companySetting, staff, desk, errors.New("当前尚未开启会员功能，如有需要，请联系销售代表")
	}
	if slices.Contains(s.h5AcceptOrderFunctionRoutes, auth.UrlPath) && companySetting.IsOpenH5Order != 1 {
		return company, companySetting, staff, desk, errors.New("当前尚未开启扫码点餐接单功能，如有需要，请联系销售代表")
	}
	switch auth.Source {
	case constant.SourceCashier: // 收银端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			// 检查收银机设置-收银用餐是否开启
			if !s.isCashierOpen(cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "收银用餐已关闭，请选择其他用餐方式")
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
			// 判断权限
			permissions, err := s.roleAccessSrv.GetApiPermission(staff.Uuid, auth.CompanyUuid)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			}
			permission := constant.CashierPermissions[auth.UrlPath]
			if permission != "" && !slices.Contains(permissions, permission) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeUnauthorized, "当前无权限，请联系管理员")
			}
			// 已交班，只能打印交班单、退出登录、获取基本信息、轮询打印数据，轮询未处理消息、获取广告
			if staff.DutyNo == "" && !slices.Contains([]string{"/api/v1/cashier/shift/printer", "/api/v1/cashier/logout"}, auth.UrlPath) {
				// 判断客户端版本，低于 2.3 的就返回 -101，直接退出
				// 高于等于的就返-108，能弹出来交班弹窗
				if ctx.Version(context.GTE, "2.3.0") { // 高于等于 2.3.0
					// 获取缓存
					if cachedSubmitShift, err := s.shiftSrv.GetCachedSubmitShift(ctx); err != nil || cachedSubmitShift == nil {
						return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenExpired, "当前班次不存在")
					} else {
						return company, companySetting, staff, desk, errors.NewWithCodeAndData(constant.CodeCashierHandedOver, *cachedSubmitShift, "已交班")
					}
				} else { // 低于 2.3.0
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTokenExpired, "当前班次不存在")
				}
			}
		}
	case constant.SourceAssistant: // 点餐助手端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			if companySetting.IsOpenAssistant != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启点餐助手功能，如有需要，请联系销售代表")
			}
			if !slices.Contains(s.assistantRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断收银机状态
				deviceRepo := repository.NewDeviceRepo(db)
				cashierDevice, _ := deviceRepo.GetDevice(deviceRepo.WhereSource(constant.SourceCashier), deviceRepo.WhereSn(auth.DeviceId))
				if cashierDevice.Uuid == 0 {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierNotLogin, "收银员设备已解绑，请重新绑定")
				}
				if staff.DutyNo == "" || auth.Assistant.StaffUuid == 0 {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierNotLogin, "收银员登录信息错误，请重新登录")
				}
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	case constant.SourceTablet: // 平板端
		{
			cashierSetting, err := s.GetCashierSetting(ctx)
			if err != nil {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeSystemError, "系统错误，请稍候再试")
			}
			if companySetting.IsOpenTablet != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启平板点餐功能，如有需要，请联系销售代表")
			}
			if !slices.Contains(s.tabletRoutes, auth.UrlPath) { // 除了这些接口外，其他都需要判断是否绑定了桌台
				deskRepo := repository.NewDeskRepo(db)
				var err error
				desk, err = deskRepo.GetDesk(deskRepo.WhereDeviceUuid(ctx.GetDeviceUuid()))
				if err != nil {
					return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeTabletNotBindDesk, "桌台未绑定")
				}
			}
			// 检查收银机设置-桌台用餐是否开启
			if !s.isTableOpen(ctx, cashierSetting, auth.UrlPath) {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
			}
		}
	case constant.SourceKitchen: // 厨显端
		{
			kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
			if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
				return company, companySetting, staff, desk, errors.NewWithCode(constant.CodeFunctionDisabled, "当前尚未开启厨显功能，如有需要，请联系销售代表")
			}
		}
	}
	return company, companySetting, staff, desk, nil
}

func (s *authSrv) AuthDesk(ctx context.Context, qrcodeToken string) (*model.Company, error) {
	companyUuid := ctx.GetCompanyUuid()
	deskUuid := ctx.GetDeskUuid()
	db := s.dbm.GetDB(companyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	companySetting := repository.NewCompanySettingRepo(db).Get()
	if companySetting.IsOpenH5 != 1 {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	deskInfo, err := repository.NewDeskRepo(db).GetDeskInfo(deskUuid)
	if err != nil || deskInfo.IsDisableDesk() || deskInfo.IsDelete() || deskInfo.QrcodeToken != qrcodeToken {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	cashierSetting, _ := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if cashierSetting.OrderMethod.IsTableOrder != "1" {
		return nil, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
	}

	return company, nil
}

func (s *authSrv) AuthMenu(ctx context.Context, qrcodeToken string) (*model.Company, error) {
	companyUuid := ctx.GetCompanyUuid()
	db := s.dbm.GetDB(companyUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil || businessSetting.QrCode != qrcodeToken {
		return nil, errors.NewWithCode(constant.CodeTokenInvalid, "二维码已失效，请联系商家")
	}

	cashierSetting, _ := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if cashierSetting.OrderMethod.IsTableOrder != "1" {
		return nil, errors.NewWithCode(constant.CodeCashierOrderMethodNotOpen, "桌台用餐已关闭，请选择其他用餐方式")
	}
	return company, nil
}

func (s *authSrv) GetCashierSetting(ctx context.Context) (*setting.Cashier, error) {
	cashierSetting, err := s.settingSrv.GetCashierSetting(ctx, []dto.LanguageItem{})
	if err != nil {
		return nil, err
	}
	return &cashierSetting, nil
}

// 检查收银机设置-收银用餐是否开启
func (s *authSrv) isCashierOpen(cashierSetting *setting.Cashier, pathUrl string) bool {
	// 检查收银用餐是否开启
	if cashierSetting.OrderMethod.IsCashierOrder != "1" && regexp.MustCompile(`^/api/v\d+/cashier/instant/`).Match([]byte(pathUrl)) {
		return false
	}
	return true
}

// 检查收银机设置-桌台用餐是否开启
func (s *authSrv) isTableOpen(ctx context.Context, cashierSetting *setting.Cashier, pathUrl string) bool {
	if cashierSetting.OrderMethod.IsTableOrder != "1" &&
		(slices.Contains([]string{constant.SourceAssistant, constant.SourceTablet}, ctx.GetSource()) ||
			(ctx.GetSource() == constant.SourceCashier && regexp.MustCompile(`^/api/v\d+/cashier/desk/`).Match([]byte(pathUrl)))) {
		return false
	}
	return true
}

// BindCashier 绑定收银机
func (s *authSrv) BindCashier(ctx context.Context, bindReq req.BindCashierReq) (string, string, error) {
	var newToken, refreshToken string
	companyUuid := helper.GetCompanyUuid(ctx.GetGin())
	if helper.GetSource(ctx.GetGin()) != constant.SourceAssistant {
		return newToken, refreshToken, errors.New("用户信息错误")
	}
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(bindReq.CashierUuid), staffRepo.WhereDeviceId(bindReq.DeviceId), staffRepo.WithCompany(), staffRepo.WithCompanySetting())
	// 检查传递的收银设备
	if staff.Uuid == 0 {
		return newToken, refreshToken, errors.New("用户不存在")
	}
	if staff.CashierOnline != 1 {
		return newToken, refreshToken, errors.New("收银设备不在线")
	}
	// 判断收银员状态
	if staff.DeleteTime != 0 {
		return newToken, refreshToken, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return newToken, refreshToken, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company == nil || staff.Company.Uuid == 0 || staff.Company.DeleteTime != 0 {
		return newToken, refreshToken, errors.New("未找到绑定的商家，请确认登录信息")
	}
	// 是否已开启点餐助手功能
	if staff.Company.CompanySetting == nil || staff.Company.CompanySetting.IsOpenAssistant != 1 {
		return newToken, refreshToken, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
	}

	claims := auth.Claims{
		Source:      constant.SourceAssistant,
		CompanyUuid: companyUuid,
		StaffUuid:   bindReq.CashierUuid,
		DeviceUuid:  ctx.GetDeviceUuid(),
		DeviceId:    bindReq.DeviceId,
		Assistant: auth.Assistant{
			DeviceId:  ctx.GetGin().GetString(jwt.DeviceId),
			StaffUuid: ctx.GetGin().GetUint64(jwt.StaffUuid),
		},
	}
	// 生成新的 JWT token，将点餐助手设备ID和员工Uuid单独存放
	newToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return newToken, refreshToken, errors.New("生成token失败")
	}
	newRefreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return newToken, refreshToken, errors.New("生成refresh_token失败")
	}
	return newToken, newRefreshToken, nil
}

// GetOnlineCashiers 获取在线收银机
func (s *authSrv) GetOnlineCashiers(companyUuid uint64) resp.OnlineCashierList {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staffs := staffRepo.GetStaffs(staffRepo.WhereCashierOnline(), staffRepo.WithDevice(constant.SettingCashier))

	cashiers := make([]resp.OnlineCashier, 0, len(staffs))
	for _, staff := range staffs {
		var remark string
		if staff.Device != nil {
			remark = staff.Device.Remark
		}
		cashiers = append(cashiers, resp.OnlineCashier{
			CashierUuid: staff.Uuid,
			Username:    staff.Username,
			DeviceId:    staff.BindKey,
			Remark:      remark,
		})
	}
	return resp.OnlineCashierList{
		List: cashiers,
	}
}

// RefreshToken 刷新token
func (s *authSrv) RefreshToken(ctx context.Context) (resp.LoginResp, error) {
	var (
		newToken, newRefreshToken string
		loginResp                 resp.LoginResp
		err                       error
	)
	claims := auth.Claims{
		Source:      ctx.GetSource(),
		CompanyUuid: ctx.GetCompanyUuid(),
		StaffUuid:   ctx.GetStaffUuid(),
		DeviceUuid:  ctx.GetDeviceUuid(),
		DeviceId:    ctx.GetDeviceSn(),
		Assistant: auth.Assistant{
			DeviceId:  ctx.GetGin().GetString(jwt.AssistantDeviceId),
			StaffUuid: ctx.GetGin().GetUint64(jwt.AssistantStaffUuid),
		},
	}
	// 生成 JWT new token，new refresh token
	newToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	newRefreshToken, err = auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return loginResp, errors.New("生成refresh_token失败")
	}
	return resp.LoginResp{
		Token:               newToken,
		RefreshToken:        newRefreshToken,
		CashierIsFirstLogin: false,
	}, nil
}

func (s *authSrv) ShopBase(ctx context.Context) (resp.ShopBase, error) {
	var shopBase resp.ShopBase

	// 如果 company_uuid 为 0，只返回可用门店列表
	if ctx.GetCompanyUuid() == 0 {
		shopBase.CompanyList = s.getCompanyList(ctx)
		return shopBase, nil
	}

	// company_uuid 不为 0，走原有逻辑
	company := ctx.GetCompany()
	companySetting := ctx.GetCompanySetting()
	staff := ctx.GetStaff()
	var (
		source   = ctx.GetSource()
		deviceId = ctx.GetGin().GetString(jwt.DeviceId)
	)
	deviceRemark := s.deviceSrv.GetRemark(company.Uuid, source, deviceId)

	// 获取员工权限（使用管理APP路由名称）
	permissions, err := s.roleAccessSrv.GetPermission(constant.ShopAppRouteName, staff.Uuid, staff.CompanyUuid)
	if err != nil {
		// 权限获取失败不影响基础信息获取，返回空权限数组
		logger.Logger.Error("获取权限失败", zap.Error(err))
		permissions = []*resp.Permission{}
	}

	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	buffetSetting, err := s.settingSrv.GetBuffetSetting(ctx, companySetting)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	cloudBasicSetting, err := s.settingSrv.GetCloudBasicSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	taxSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return shopBase, errors.WithMessage(err)
	}
	// 是否有数据管理权限
	hasDataPermission := false
	if companySetting.EnableDataManagement == 1 && (staff.HasDataPermission == 1 || staff.IsSuper == 1) {
		hasDataPermission = true
	}

	return resp.ShopBase{
		Username:     staff.Username,
		RealName:     staff.RealName,
		ProfileUuid:  staff.Uuid,
		DeviceId:     deviceId,
		DeviceRemark: deviceRemark,
		Permissions:  permissions,
		Phone:        staff.Phone,

		Business: businessSetting,
		Buffet:   buffetSetting,
		Currency: currencySetting,
		Company: resp.Company{ // shop
			Uuid:                 company.Uuid,
			Name:                 company.Name,
			Logo:                 utils.AddImageDomain(company.Logo, utils.GetBaseURL(ctx.GetGin().Request), true),
			TimeZone:             companySetting.Timezone,
			ExpireTime:           company.ExpireTime,
			IsOpenMember:         companySetting.IsOpenMember,
			IsOpenBuffet:         companySetting.IsOpenBuffet,
			IsOpenH5Order:        companySetting.IsOpenH5Order,
			IsOpenOldOrder:       utils.IfInt(company.OldCompanyId > 0, 1, 0),
			IsOpenRider:          companySetting.IsOpenRider(),
			IsEnableErp:          company.IsOpenErp(),
			IsOpenMap:            companySetting.IsOpenTableMap(),
			IsOpenDataManagement: companySetting.IsOpenDataManagement(),
			IsOpenKiosk:          companySetting.IsOpenKiosk(),
			IsOpenGrabDelivery:   companySetting.IsOpenGrabDelivery(),
		},
		CloudBasic: cloudBasicSetting,
		Profile: resp.ShopProfile{
			Address:         storeSetting.Address,
			Coordinates:     storeSetting.Coordinates,
			IpWhiteList:     storeSetting.IPWhiteList,
			Phone:           storeSetting.Phone,
			TaxNumber:       storeSetting.TaxNumber,
			StoreCode:       storeSetting.StoreCode,
			TimeZoneList:    storeSetting.TimeZoneList,
			DefaultLanguage: storeSetting.Language[0].Name,
			LanguageList:    storeSetting.Language,
			Language:        companySetting.GetLanguages(),
			CompanyName:     storeSetting.Company,
		},
		IsTtposSite:       companySetting.IsTtposSite(),
		IsHeadquarter:     companySetting.IsHeadquarter(),
		UpdateTime:        time.Now().Unix(),
		ServerVersion:     utils.GetVersion(),
		IsOpenTax:         taxSetting.IsOpen == "1",
		IsSyncing:         slices.Contains(syncTaskManager.getRunningCompanyUuids(), company.Uuid),
		LastSyncTime:      company.LastSyncTime,
		HasChildren:       companySetting.HasChildren == 1,
		HasDataPermission: hasDataPermission,
		CompanyList:       s.getCompanyList(ctx),
	}, nil
}

// 修改密码
func (s *authSrv) ChangePassword(ctx context.Context, changePasswordReq req.ChangePasswordReq) error {
	if ctx.GetSource() != constant.SourceShop {
		return errors.New("当前无权限，请联系管理员")
	}

	saasDB := s.dbm.GetDB(constant.DefaultDB)
	staffUuid := ctx.GetStaffUuid()

	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	staff, err := saasStaffRepo.GetByUuid(staffUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 验证旧密码（支持 MD5 和 bcrypt）
	isValid, _ := utils.VerifyPassword(changePasswordReq.OldPassword, staff.Password)
	if !isValid {
		return errors.New("旧密码错误")
	}

	// 使用 bcrypt 加密新密码
	newPasswordHash, err := utils.HashPasswordBcrypt(changePasswordReq.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新统一账号表
	update := map[string]any{
		"password":              newPasswordHash,
		"password_change_count": staff.PasswordChangeCount + 1,
		"password_change_time":  time.Now().Unix(),
	}
	saasStaffRepo.Update(staffUuid, update)

	// 同步更新到关联的每个商家数据库
	companyStaffRepo := repository.NewCompanyStaffRepo(s.dbm.GetDB(constant.DefaultDB))
	companyStaffList, err := companyStaffRepo.GetByStaffUuid(staffUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	for _, companyStaff := range companyStaffList {
		db := s.dbm.GetDB(companyStaff.CompanyUuid)
		if db == nil {
			continue
		}
		staffRepo := repository.NewStaffRepo(db)
		staffRepo.Update(staffUuid, update)
	}
	return nil
}

// StoreSwitch 门店切换（返回新 token）
func (s *authSrv) StoreSwitch(ctx context.Context, switchReq req.StoreSwitchReq) (resp.LoginResp, error) {
	var loginResp resp.LoginResp

	staffUuid := ctx.GetStaffUuid()
	source := ctx.GetSource()
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 验证员工是否有该门店权限（关联查询商家信息）
	companyStaff, err := companyStaffRepo.GetByStaffAndCompany(staffUuid, switchReq.CompanyUuid, companyStaffRepo.WithCompany())
	if err != nil || companyStaff == nil {
		return loginResp, errors.New("无权限访问该门店")
	}

	// 过滤条件：员工在该商家未被禁用
	if companyStaff.IsDisable != 0 {
		return loginResp, errors.New("无权限访问该门店")
	}

	// 从门店数据库查询员工信息，验证员工状态
	companyDB := s.dbm.GetDB(switchReq.CompanyUuid)
	if companyDB == nil {
		return loginResp, errors.New("门店不存在")
	}

	staffRepo := repository.NewStaffRepo(companyDB)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(staffUuid), staffRepo.WithCompany())

	// 商家状态
	if staff.Company == nil {
		return loginResp, errors.New("未找到绑定的商家，请确认登录信息")
	}
	if err != nil || staff.Uuid == 0 {
		return loginResp, errors.New("员工不存在")
	}
	// 检查员工状态
	if staff.DeleteTime != 0 {
		return loginResp, errors.New("账号被删除，请联系管理员")
	}
	if staff.IsDisable == 1 {
		return loginResp, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company.IsExpired() {
		return loginResp, errors.NewWithCode(constant.CodeCompanyLicenceExpired, "店铺状态已到期，如需继续使用，请联系销售代表")
	}
	if staff.Company.IsException() {
		return loginResp, errors.New("店铺状态异常，如需继续使用，请联系销售代表")
	}
	isFirstLogin := staff.CashierOnline == 0

	// 根据 source 进行权限验证（参考 loginWithCompany 的逻辑）
	switch ctx.GetSource() {
	case constant.SourceCashier: // 收银端登录
		// 判断权限
		permissions, err := s.roleAccessSrv.GetPermission(constant.CashierRouteName, staff.Uuid, staff.CompanyUuid)
		if err != nil {
			return loginResp, errors.WithMessage(err)
		}
		if len(permissions) == 0 {
			return loginResp, errors.New("当前无权限，请联系管理员")
		}

		// 获取员工仓库
		staffRepo := repository.NewStaffRepo(s.dbm.GetDB(staff.CompanyUuid))

		// 检查是否有未交班的收银员
		if viper.GetString("CHECK_SHIFT_HANDOVER") != "false" {
			currentStaff, _ := staffRepo.GetStaff(staffRepo.WhereDeviceId(ctx.GetDeviceSn()), staffRepo.WhereCashierOnline())
			if currentStaff.Uuid != 0 && currentStaff.Uuid != staff.Uuid {
				return loginResp, errors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.GetUserName()})
			}
			// 是否已在其他收银机登录
			if staff.CashierOnline == 1 && ctx.GetDeviceSn() != staff.BindKey {
				return loginResp, errors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{staff.GetUserName()})
			}
		}

		// 更新员工信息
		updates := map[string]any{
			"cashier_online": 1,
			"bind_key":       ctx.GetDeviceSn(),
		}
		// 创建当班日志
		if staff.CashierLoginTime == 0 || staff.CashierOnline == 0 {
			companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
			ctx.SetCompany(*staff.Company)
			ctx.SetCompanySetting(companySetting)
			shiftLog, err := s.shiftSrv.CreateWorkingLog(ctx, staff)
			if err != nil {
				return loginResp, errors.WithMessage(err, "创建当班日志失败")
			}
			updates["cashier_login_time"] = shiftLog.ShiftStartTime
			updates["duty_no"] = shiftLog.ShiftNo
		}
		err = staffRepo.Update(staff.Uuid, updates)
		if err != nil {
			return loginResp, errors.WithMessage(err, "更新信息失败")
		}
	case constant.SourceAssistant: // 点餐助手登录
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenAssistant != 1 {
			return loginResp, errors.New("当前尚未开启点餐助手功能，如有需要，请联系销售代表")
		}
	case constant.SourceKitchen: // 厨显端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		kitchenSetting, _ := s.settingSrv.GetKitchenSetting(ctx, companySetting, []dto.LanguageItem{})
		if kitchenSetting.IsOpen != "1" || companySetting.IsOpenKitchenKds != 1 {
			return loginResp, errors.New("当前尚未开启厨显功能，如有需要，请联系销售代表")
		}
	case constant.SourceTablet: // 平板端
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if companySetting.IsOpenTablet != 1 {
			return loginResp, errors.New("当前尚未开启平板点餐功能，如有需要，请联系销售代表")
		}
	case constant.SourceKiosk: // 自助点餐机
		companySetting := repository.NewCompanySettingRepo(s.dbm.GetDB(staff.CompanyUuid)).Get()
		if !companySetting.IsOpenKiosk() {
			return loginResp, errors.New("当前尚未开启自助点餐机功能，如有需要，请联系销售代表")
		}
	case constant.SourceShop: // 移动管理端
		// 权限获取在 shop_auth.go 中完成，这里不需要处理
	default:
		return loginResp, errors.New("登录来源错误")
	}

	// 登录时没有商家ID，补上
	ctx.SetCompanyUuid(staff.CompanyUuid)
	// 添加绑定记录
	deviceUuid, err := s.deviceSrv.AddDevice(ctx, req.AddDeviceReq{
		DeviceId:         ctx.GetDeviceSn(),
		Brand:            ctx.GetBrand(),
		Source:           ctx.GetSource(),
		FinallyLoginUuid: staff.Uuid,
		FinallyLoginTime: time.Now().Unix(),
		CompanyUuid:      staff.CompanyUuid,
	})
	if err != nil {
		return loginResp, errors.WithMessage(err)
	}

	claims := auth.Claims{
		Source:      ctx.GetSource(),
		CompanyUuid: staff.CompanyUuid,
		StaffUuid:   staff.Uuid,
		DeviceUuid:  deviceUuid,
		DeviceId:    ctx.GetDeviceSn(),
		Assistant:   auth.Assistant{},
	}
	// 生成 JWT token，refresh_token
	token, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.Expire, false)
	if err != nil {
		return loginResp, errors.New("生成token失败")
	}
	refreshToken, err := auth.GenerateToken(claims, config.JWT.Secret, config.JWT.RefreshExpire, true)
	if err != nil {
		return loginResp, errors.New("生成refresh_token失败")
	}

	// 更新 last_company_uuid（仅新管理端）
	if source == constant.SourceShop {
		saasStaffSrv := NewSaasStaffSrv(s.dbm)
		if err := saasStaffSrv.UpdateLastCompany(ctx, staffUuid, switchReq.CompanyUuid); err != nil {
			// 记录日志但不影响切换流程
			logger.Logger.Warn("更新上次登录门店失败", zap.Error(err))
		}
	}

	return resp.LoginResp{
		Token:               token,
		RefreshToken:        refreshToken,
		CashierIsFirstLogin: isFirstLogin,
		NeedChangePassword:  staff.PasswordChangeCount == 0,
	}, nil
}

// getAvailableCompanyList 获取员工可用的商家列表（过滤已过期、异常的商家）
func (s *authSrv) getCompanyList(ctx context.Context) []*resp.CompanyStaffResp {
	// 过滤可用门店（过滤已过期、异常的商家）
	availableCompanyList := make([]*resp.CompanyStaffResp, 0)

	staffUuid := ctx.GetStaffUuid()
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 获取员工关联的门店列表
	companyList, _ := companyStaffRepo.GetByStaffUuid(staffUuid, companyStaffRepo.WithCompany())

	for _, cs := range companyList {
		if cs.IsDisable == 1 {
			continue
		}
		shopDb := s.dbm.GetDB(cs.CompanyUuid)
		// 根据员工UUID查询角色列表
		staffRoleRepo := repository.NewStaffRoleRepo(shopDb)
		roleUuids, _ := staffRoleRepo.GetRoleUuidsByStaffUuid(cs.Uuid)
		roleRepo := repository.NewRoleRepo(shopDb)
		roles, _ := roleRepo.GetRoleList(roleRepo.WhereUuids(roleUuids))
		roleNames := make([]string, 0, len(roles))
		for _, role := range roles {
			roleNames = append(roleNames, role.Name)
		}
		company := cs.Company
		if company == nil || company.IsExpired() || company.IsException() {
			continue
		}
		availableCompanyList = append(availableCompanyList, &resp.CompanyStaffResp{
			CompanyUuid: cs.CompanyUuid,
			CompanyName: company.Name,
			Roles:       roleNames,
			IsSuper:     cs.IsSuper,
		})
	}

	return availableCompanyList
}

package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"ttpos-server-go/app/dto/req"

	"ttpos-server-go/app/constant"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/utils"
)

type CashierAuthService struct {
	companyStaffRepo *repository.CompanyStaffRepository
	staffRepo        *repository.StaffRepository
	loginLogRepo     *repository.LoginLogRepository
	captchaSrv       *CaptchaService
	roleAccessSrv    *RoleAccessService
	bindRecordSrv    *BindRecordService
}

func NewCashierAuthService(
	staffRepo *repository.StaffRepository,
	companyStaffRepo *repository.CompanyStaffRepository,
	captchaSrv *CaptchaService,
	roleAccessSrv *RoleAccessService,
	bindRecordSrv *BindRecordService,
) *CashierAuthService {
	return &CashierAuthService{
		companyStaffRepo: companyStaffRepo,
		staffRepo:        staffRepo,
		captchaSrv:       captchaSrv,
		roleAccessSrv:    roleAccessSrv,
		bindRecordSrv:    bindRecordSrv,
	}
}

// Login 登录
func (s *CashierAuthService) Login(loginReq req.CashierLoginRequest, captchaId, captchaCode string) (string, error) {
	var token string
	// 验证验证码
	if !s.captchaSrv.Verify(captchaId, captchaCode) {
		return token, apperrors.New("验证码错误")
	}

	var staff model.Staff
	if config.Server.DeployMode == "cloud" { // 云上版本
		companyStaff := s.companyStaffRepo.GetByUsername(loginReq.UserName, s.companyStaffRepo.WithCompany())
		if companyStaff.StaffId == 0 {
			return token, errors.New("账号不存在")
		}
		if companyStaff.CompanyId == 0 {
			return token, errors.New("未找到绑定的商家，请确认登录信息")
		}
		staff = s.staffRepo.GetById(companyStaff.StaffId, companyStaff.CompanyId, s.staffRepo.WithCompany())
	} else { // 离线版本
		staff = s.staffRepo.GetByUsername(loginReq.UserName, s.staffRepo.WithCompany())
	}

	if staff.ID == 0 {
		return token, errors.New("用户不存在")
	}

	if utils.EncryptPassword(loginReq.Password) != staff.Password {
		return token, errors.New("密码错误")
	}
	if staff.IsDelete == 1 {
		return token, apperrors.NewWithReplace("账号 %s 被删除，请联系管理员", []string{staff.Username})
	}
	if staff.IsDisable == 1 {
		return token, errors.New("账号被禁用，请联系管理员")
	}
	if staff.Company.IsRecycle != 0 {
		return token, errors.New("商家账号异常，请联系管理员")
	}

	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(true, constant.CASHIER_ROUTE_NAME, staff.ID, staff.CompanyId)
	if len(permissions) == 0 {
		return token, errors.New("当前无权限，请联系管理员")
	}

	//// 检查是否有未交班的收银员
	//currentStaff := s.staffRepo.GetCurrentCashier(staff.BindKey)
	//if currentStaff.ID != 0 && currentStaff.ID != staff.ID {
	//	return token, apperrors.NewWithReplace("当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentStaff.RealName})
	//}
	//
	//// 是否是首次接班
	////isFirstLogin := companyStaff.CashierOnline == 0
	//
	//// 是否已在其他收银机登录
	//if staff.CashierOnline == 1 && bindKey != staff.BindKey {
	//	cashierName := staff.RealName
	//	if cashierName == "" {
	//		cashierName = staff.Username
	//	}
	//	return token, apperrors.NewWithReplace("收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{cashierName})
	//}

	// // 绑定设备 先检测是否能进行绑定，避免先更新用户表信息再弹出绑定错误
	//        $data = [
	//            'key' => $device_id,
	//            'brand' => $brand,
	//            'source' => BindRecordModel::SOURCE_CASHIER,
	//            'finally_login_id' => $user['shop_user_id'],
	//            'finally_login_time' => time(),
	//            'app_id' => $user['app_id'],
	//            'shop_supplier_id' => $user['shop_supplier_id'],
	//        ];
	//        $model = new BindRecordModel;
	//        if (!$model->add($data, $user->license)) {
	//            $this->error = $model->getError() ?: __('绑定失败');
	//            return false;
	//        }

	// 生成 JWT token
	token, err = auth.GenerateToken(constant.SOURCE_CASHIER, staff.ID, staff.CompanyId, config.JWT.Secret, config.JWT.Expire)
	if err != nil {
		return token, errors.New("生成token失败")
	}
	return token, nil
}

// Logout 退出登录
func (s *CashierAuthService) Logout(cc *gin.Context) error {
	staff := s.staffRepo.GetById(cc.GetUint("staff_id"), cc.GetUint("company_id"))
	return s.bindRecordSrv.Unbind(cc.GetUint("appId"), constant.SOURCE_CASHIER, staff.BindKey, staff.ID)
}

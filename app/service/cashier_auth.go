package service

import (
	"errors"
	"github.com/gin-gonic/gin"

	"ttpos-server-go/app/constant"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/auth"
	"ttpos-server-go/pkg/utils"
)

type CashierAuthService struct {
	userRepo      *repository.UserRepository
	loginLogRepo  *repository.LoginLogRepository
	captchaSrv    *CaptchaService
	roleAccessSrv *RoleAccessService
	bindRecordSrv *BindRecordService
}

func NewCashierAuthService(
	userRepo *repository.UserRepository,
	captchaSrv *CaptchaService,
	roleAccessSrv *RoleAccessService,
	bindRecordSrv *BindRecordService,
) *CashierAuthService {
	return &CashierAuthService{
		userRepo:      userRepo,
		captchaSrv:    captchaSrv,
		roleAccessSrv: roleAccessSrv,
		bindRecordSrv: bindRecordSrv,
	}
}

// Login 登录
func (s *CashierAuthService) Login(username, password, captchaId, captchaCode string) (string, error) {
	// 验证验证码
	if !s.captchaSrv.Verify(captchaId, captchaCode) {
		return "", errors.New("验证码错误")
	}
	// 验证账号
	user := s.userRepo.GetByUsername(username, s.userRepo.WithApp(), s.userRepo.WithSupplier())
	if user.ShopUserId == 0 {
		return "", errors.New("账号不存在")
	}
	if utils.EncryptPassword(password) != user.Password {
		return "", errors.New("密码错误")
	}
	if user.IsDelete == 1 {
		return "", apperrors.NewWithReplace(constant.CodeAccountDeleted, "账号 %s 被删除，请联系管理员", []string{user.UserName})
	}
	if user.IsStatus == 1 {
		return "", errors.New("账号被禁用，请联系管理员")
	}
	if user.App == nil || user.App.IsDelete == 1 {
		return "", errors.New("未找到绑定的商家，请确认登录信息")
	}
	if user.App.IsRecycle != 0 {
		return "", errors.New("商家账号异常，请联系管理员")
	}

	// 判断权限
	permissions, err := s.roleAccessSrv.GetPermission(true, constant.CASHIER_ROUTE_NAME, user.ShopUserId)
	if len(permissions) == 0 {
		return "", errors.New("当前无权限，请联系管理员")
	}

	// 检查是否有未交班的收银员
	currentUser := s.userRepo.GetCurrentCashier(user.BindKey)
	if currentUser.ShopUserId != 0 && currentUser.ShopUserId != user.ShopUserId {
		return "", apperrors.NewWithReplace(constant.CodeUnhandShiftUserExists, "当前收银机上有未交班的账号，请联系 %s 完成交班后再登录", []string{currentUser.RealName})
	}

	// 是否是首次接班
	//isFirstLogin := user.CashierOnline == 0

	// 是否已在其他收银机登录
	if user.CashierOnline == 1 && user.BindKey != user.BindKey {
		cashierName := user.RealName
		if cashierName == "" {
			cashierName = user.UserName
		}
		return "", apperrors.NewWithReplace(constant.CodeUnhandShiftUserExists, "收银员 %s 已在其他收银机登录未交班，请先完成交班操作", []string{cashierName})
	}

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
	token, err := auth.GenerateToken(constant.SOURCE_CASHIER, user.ShopUserId, user.AppId, config.JWT.Secret, config.JWT.Expire)
	if err != nil {
		return "", errors.New("生成token失败")
	}
	return token, nil
}

// Logout 退出登录
func (s *CashierAuthService) Logout(cc *gin.Context) error {
	shopUser := s.userRepo.GetById(cc.GetUint("shopUserId"))
	return s.bindRecordSrv.Unbind(cc.GetUint("appId"), constant.SOURCE_CASHIER, shopUser.BindKey, shopUser.ShopUserId)
}

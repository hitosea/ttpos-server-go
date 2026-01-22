package helper

import (
	stdErrors "errors"
	"net/http"
	"regexp"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"
)

// isMySQLError 判断是否是数据库错误（MySQL/GORM 相关）
// 规则：
// 1. MySQL 错误码格式：Error + 四位数字
// 2. GORM 特定错误
// 3. 常见数据库错误关键词
func isMySQLError(err error) bool {
	if err == nil {
		return false
	}

	// 1. 检查 GORM 特定错误
	if stdErrors.Is(err, gorm.ErrRecordNotFound) ||
		stdErrors.Is(err, gorm.ErrInvalidTransaction) ||
		stdErrors.Is(err, gorm.ErrNotImplemented) ||
		stdErrors.Is(err, gorm.ErrMissingWhereClause) ||
		stdErrors.Is(err, gorm.ErrUnsupportedRelation) ||
		stdErrors.Is(err, gorm.ErrPrimaryKeyRequired) ||
		stdErrors.Is(err, gorm.ErrModelValueRequired) ||
		stdErrors.Is(err, gorm.ErrModelAccessibleFieldsRequired) ||
		stdErrors.Is(err, gorm.ErrSubQueryRequired) ||
		stdErrors.Is(err, gorm.ErrInvalidData) ||
		stdErrors.Is(err, gorm.ErrUnsupportedDriver) ||
		stdErrors.Is(err, gorm.ErrRegistered) ||
		stdErrors.Is(err, gorm.ErrInvalidField) ||
		stdErrors.Is(err, gorm.ErrEmptySlice) ||
		stdErrors.Is(err, gorm.ErrDryRunModeUnsupported) ||
		stdErrors.Is(err, gorm.ErrInvalidDB) ||
		stdErrors.Is(err, gorm.ErrInvalidValue) ||
		stdErrors.Is(err, gorm.ErrInvalidValueOfLength) ||
		stdErrors.Is(err, gorm.ErrPreloadNotAllowed) ||
		stdErrors.Is(err, gorm.ErrDuplicatedKey) ||
		stdErrors.Is(err, gorm.ErrForeignKeyViolated) ||
		stdErrors.Is(err, gorm.ErrCheckConstraintViolated) {
		return true
	}

	errMsg := err.Error()

	// 2. 检查 MySQL 错误码格式（主要规则）
	// 匹配 "Error " 后跟四位数字的模式（如 Error 1062, Error 1451）
	matched, _ := regexp.MatchString(`^Error \d{4}`, errMsg)
	if matched {
		return true
	}

	// 3. 检查常见数据库错误关键词（不区分大小写）
	// 注意：只保留网络层/连接层错误，MySQL 标准错误已被第2层覆盖
	lowerErrMsg := strings.ToLower(errMsg)
	dbKeywords := []string{
		"connection refused", // 网络连接被拒绝
		"connection reset",   // 网络连接重置
		"invalid connection", // 连接无效
		"sql:",               // Go database/sql 错误
	}

	for _, keyword := range dbKeywords {
		if strings.Contains(lowerErrMsg, keyword) {
			return true
		}
	}

	return false
}

// ErrorWithDetail 返回错误
func ErrorWithDetail(c *gin.Context, code int, err error) {
	// 记录完整错误到日志
	logger.Logger.Info("ErrorWithDetail", zap.String("url", c.Request.URL.String()), zap.String("error", err.Error()))

	if config.Server.Mode == constant.ServerModeRelease {
		// 只有Release模式才返回原始错误信息。如123123
		// 开发模式和测试模式都返回调用栈信息。如[v1/cashier/cashier_instant.go:384]h.orderService.GetOrderCartInfoByDeviceSn failed: [app/service/order.go:1837]deviceRepo.GetDevice failed: [app/repository/device.go:41]db.First failed: 123123
		err = pkgerrors.Cause(err)
		// Release 模式下，如果是 MySQL 错误，返回通用错误信息
		if isMySQLError(err) {
			err = stdErrors.New("查询数据失败")
		}
	}

	messages := []string{err.Error()}
	var appErr errors.AppError
	if pkgerrors.As(err, &appErr) {
		if appErr.GetCode() != constant.CodeFail {
			code = appErr.GetCode()
		}
		if len(appErr.Replace) > 0 {
			messages = append(messages, appErr.Replace...)
		}
	}
	Fail(c, code, messages...)
}

// ErrorWithMessage 返回错误
func ErrorWithMessage(c *gin.Context, code int, err error) {
	// 记录完整错误到日志
	logger.Logger.Info("ErrorWithMessage", zap.String("url", c.Request.URL.String()), zap.String("error", err.Error()))

	if config.Server.Mode == constant.ServerModeRelease {
		// 只有Release模式才返回原始错误信息。如123123
		// 开发模式和测试模式都返回调用栈信息。如[v1/cashier/cashier_instant.go:384]h.orderService.GetOrderCartInfoByDeviceSn failed: [app/service/order.go:1837]deviceRepo.GetDevice failed: [app/repository/device.go:41]db.First failed: 123123
		err = pkgerrors.Cause(err)
		// Release 模式下，如果是 MySQL 错误，返回通用错误信息
		if isMySQLError(err) {
			err = stdErrors.New("查询数据失败")
		}
	}

	messages := []string{err.Error()}
	var appErr errors.AppError
	if pkgerrors.As(err, &appErr) {
		if len(appErr.Replace) > 0 {
			messages = append(messages, appErr.Replace...)
		}
	}
	Fail(c, code, messages...)
}

// ErrorWithData 返回错误携带数据
func ErrorWithData(c *gin.Context, code int, data any, err error) {
	// 记录完整错误到日志
	logger.Logger.Info("ErrorWithData", zap.String("url", c.Request.URL.String()), zap.String("error", err.Error()))

	if config.Server.Mode == constant.ServerModeRelease {
		err = pkgerrors.Cause(err)
		// Release 模式下，如果是 MySQL 错误，返回通用错误信息
		if isMySQLError(err) {
			err = stdErrors.New("查询数据失败")
		}
	}

	messages := []string{err.Error()}
	var appErr errors.AppError
	if pkgerrors.As(err, &appErr) {
		code = appErr.GetCode()
		if len(appErr.Replace) > 0 {
			messages = append(messages, appErr.Replace...)
		}
	}
	FailWithData(c, code, data, nil, messages...)
}

// ErrorAutoWithData 返回错误携带数据
func ErrorAutoWithData(c *gin.Context, code int, err error) {
	// 记录完整错误到日志
	logger.Logger.Info("ErrorAutoWithData", zap.String("url", c.Request.URL.String()), zap.String("error", err.Error()))

	if config.Server.Mode == constant.ServerModeRelease {
		err = pkgerrors.Cause(err)
		// Release 模式下，如果是 MySQL 错误，返回通用错误信息
		if isMySQLError(err) {
			err = stdErrors.New("查询数据失败")
		}
	}

	messages := []string{err.Error()}
	var appErr errors.AppError
	if pkgerrors.As(err, &appErr) {
		code = appErr.GetCode()
		if len(appErr.Replace) > 0 {
			messages = append(messages, appErr.Replace...)
		}
	}
	FailWithData(c, code, appErr.GetData(), nil, messages...)
}

// Success 返回成功
func Success(c *gin.Context, data any, message ...string) {
	msg := "请求成功"
	if len(message) == 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0])
	} else if len(message) > 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0], message[1:]...)
	} else {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), msg)
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:    constant.CodeSuccess,
		Message: msg,
		Data:    data,
	})
}

// Fail 返回失败
func Fail(c *gin.Context, code int, message ...string) {
	msg := "fail"
	if len(message) == 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0])
	} else if len(message) > 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0], message[1:]...)
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:    code,
		Message: msg,
		Data:    gin.H{},
	})
}

// FailWithData 返回失败携带数据
func FailWithData(c *gin.Context, code int, data any, err error, message ...string) {
	msg := "fail"
	if err != nil {
		if config.Server.Mode == constant.ServerModeRelease {
			msg = pkgerrors.Cause(err).Error()
		} else {
			msg = err.Error()
		}
	}
	if len(message) == 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0])
	} else if len(message) > 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0], message[1:]...)
	} else {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), msg)
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}

// HandleValidationError 处理参数验证错误
func HandleValidationError(c *gin.Context, err error, obj any, messages map[string]string) {
	structFieldJsonTagMaps := utils.GetStructFieldsMapRecursive(obj)
	var ves validator.ValidationErrors
	ok := pkgerrors.As(err, &ves)
	if !ok {
		if config.Server.Mode == constant.ServerModeDebug {
			Fail(c, constant.CodeParamError, err.Error())
		} else {
			Fail(c, constant.CodeParamError, "参数错误")
		}
		return
	}
	if config.Server.Mode == constant.ServerModeDebug {
		// 获取环境配置
		Fail(c, constant.CodeParamError, err.Error())
		return
	}
	for _, ve := range ves {
		if jsonTag, jsonTagExists := structFieldJsonTagMaps[ve.StructField()]; jsonTagExists {
			if message, messageExists := messages[jsonTag+"."+ve.Tag()]; messageExists {
				Fail(c, constant.CodeParamError, message)
				return
			}
		}
	}
	Fail(c, constant.CodeParamError, "参数错误")
}

// GetCompanyUuid 获取公司ID
func GetCompanyUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.CompanyUuid)
}

// GetStaffUuid 获取员工Uuid
func GetStaffUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.StaffUuid)
}

// GetAssistantUuid 获取员工Uuid(助手端)
func GetAssistantUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.AssistantStaffUuid)
}

// GetSource 获取来源
func GetSource(c *gin.Context) string {
	return c.GetString(jwt.Source)
}

func GetDeskUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.DeskUuid)
}

func GetDeviceSn(c *gin.Context) string {
	return c.GetString(jwt.DeviceId)
}

func GetDeviceUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.DeviceUuid)
}

func GetRequestUuid(c *gin.Context) string {
	return c.GetString(jwt.RequestUuid)
}

func GetMemberUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.MemberUuid)
}

func GetMember(c *gin.Context) model.Member {
	if val, exists := c.Get(jwt.Member); exists {
		if member, ok := val.(model.Member); ok {
			return member
		}
	}
	return model.Member{}
}

// GetLanguage 获取语言
func GetLanguage(c *gin.Context) string {
	return i18n.GetAcceptLanguage(c)
}

// GetCompany 获取request上下文的company
func GetCompany(cc *gin.Context) model.Company {
	if val, exists := cc.Get(jwt.Company); exists {
		if company, ok := val.(model.Company); ok {
			return company
		}
	}
	return model.Company{}
}

// GetStaff 获取request上下文的staff
func GetStaff(cc *gin.Context) model.Staff {
	if val, exists := cc.Get(jwt.Staff); exists {
		if staff, ok := val.(model.Staff); ok {
			return staff
		}
	}
	return model.Staff{}
}

// GetCompanySetting 获取request上下文的companySetting
func GetCompanySetting(cc *gin.Context) model.CompanySetting {
	if val, exists := cc.Get(jwt.CompanySetting); exists {
		if companySetting, ok := val.(model.CompanySetting); ok {
			return companySetting
		}
	}
	return model.CompanySetting{}
}

func GetH5OrderUuid(cc *gin.Context) string {
	return cc.GetHeader("h5_order_uuid")
}

func GetContext(c *gin.Context) context.Context {
	return context.NewContext(
		context.WithGinContext(c.Copy()), // 在上下文中添加gin上下文
		// FIXME: 先注释-防止gin.Request.Context() 被取消导致上下文取消，导致上下文传递错误
		// context.WithContext(c.Request.Context()),         // 使用 gin.Request.Context() 以包含 otelgin 创建的 span
		context.WithLanguage(GetLanguage(c)),             // 在上下文中添加语言
		context.WithCompanyUuid(GetCompanyUuid(c)),       // 在上下文中添加公司uuid
		context.WithSource(GetSource(c)),                 //在上下文中添加来源
		context.WithCompany(GetCompany(c)),               // 在上下文中添加公司信息
		context.WithStaff(GetStaff(c)),                   // 在上下文中添加员工信息
		context.WithStaffUuid(GetStaffUuid(c)),           // 在上下文中添加员工Uuid
		context.WithMember(GetMember(c)),                 // 在上下文中添加会员信息
		context.WithMemberUuid(GetMemberUuid(c)),         // 在上下文中添加会员Uuid
		context.WithAssistantUuid(GetAssistantUuid(c)),   // 在上下文中添加员工Uuid(助手端)
		context.WithCompanySetting(GetCompanySetting(c)), // 在上下文中添加公司设置信息
		context.WithDeskUuid(GetDeskUuid(c)),             // 在上下文中添加桌台ID信息
		context.WithDeviceSn(GetDeviceSn(c)),             // 在上下文中添加设备SN信息
		context.WithDeviceUuid(GetDeviceUuid(c)),         // 在上下文中添加设备Uuid
		context.WithRequestUuid(GetRequestUuid(c)),       // 在上下文中请求uuid
		context.WithH5OrderUuid(GetH5OrderUuid(c)),       // 在上下文中添加H5订单ID
		context.WithLogger(func() *zap.Logger {
			if logger.Logger == nil {
				return zap.NewNop() // 避免未初始化日志导致nil空指针错误
			}
			return logger.Logger
		}()),
	)
}

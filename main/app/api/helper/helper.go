package helper

import (
	"net/http"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/dto"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

func ErrorWithDetail(c *gin.Context, code int, err error) {
	messages := []string{err.Error()}
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		code = appErr.GetCode()
		if len(appErr.Replace) > 0 {
			messages = append(messages, appErr.Replace...)
		}
	}
	Fail(c, code, messages...)
}

func Success(c *gin.Context, data interface{}, message ...string) {
	msg := "success"
	if len(message) == 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0])
	} else if len(message) > 1 {
		msg = i18n.Translate(i18n.GetAcceptLanguage(c), message[0], message[1:]...)
	}
	c.JSON(http.StatusOK, dto.Response{
		Code:    constant.CodeSuccess,
		Message: msg,
		Data:    data,
	})
}

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

// HandleValidationError 处理参数验证错误
func HandleValidationError(c *gin.Context, err error, obj any, messages map[string]string) {
	structFieldJsonTagMaps := utils.GetStructFieldsMapRecursive(obj)
	var ves validator.ValidationErrors
	ok := errors.As(err, &ves)
	if !ok {
		Fail(c, constant.CodeBadRequest, "参数错误")
		return
	}
	if config.Server.Mode == "debug" {
		// 获取环境配置
		Fail(c, constant.CodeBadRequest, err.Error())
		return
	}
	for _, ve := range ves {
		if jsonTag, jsonTagExists := structFieldJsonTagMaps[ve.StructField()]; jsonTagExists {
			if message, messageExists := messages[jsonTag+"."+ve.Tag()]; messageExists {
				Fail(c, constant.CodeBadRequest, message)
				return
			}
		}
	}
	Fail(c, constant.CodeBadRequest, "参数错误")
}

// GetCompanyUuid 获取公司ID
func GetCompanyUuid(c *gin.Context) uint64 {
	return c.GetUint64(jwt.CompanyUuid)
}

// GetSource 获取来源
func GetSource(c *gin.Context) uint {
	return c.GetUint(jwt.Source)
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

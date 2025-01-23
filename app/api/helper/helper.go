package helper

import (
	"net/http"
	"ttpos-server-go/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/utils"
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

// GetCompanyId 获取公司ID
func GetCompanyId(c *gin.Context) uint {
	token := c.GetHeader("token")
	claims, err := auth.ParseToken(token, config.JWT.Secret)
	if err != nil {
		return 0
	}
	companyId := claims.AppId
	return companyId
}

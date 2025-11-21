package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("auto_order_limit", autoOrderLimit)
		_ = v.RegisterValidation("strong_password", strongPassword)
		_ = v.RegisterValidation("permission_password", permissionPassword)
		_ = v.RegisterValidation("opening_hours", openingHours)
	}
}

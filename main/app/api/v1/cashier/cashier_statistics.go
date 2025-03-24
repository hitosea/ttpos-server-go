package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// statisticsHandler 营业数据相关控制器
type statisticsHandler struct {
	businessSrv service.IBusinessSrv
}

// Printer 打印
// @Summary 打印
// @Description 打印
// @Tags 收银端.营业数据
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.BusinessDataPrinterReq true "打印参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/statistics/printer [post]
func (h *statisticsHandler) Printer(c *gin.Context) {
	ctx := helper.GetContext(c)
	var printerReq req.BusinessDataPrinterReq
	if err := c.ShouldBindJSON(&printerReq); err != nil {
		helper.HandleValidationError(c, err, printerReq, nil)
		return
	}
	printerData, err := h.businessSrv.Printer(ctx, printerReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, printerData)
}

func RegisterStatisticsHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	businessSrv := service.NewBusinessSrv(dbm)

	wrapper := &statisticsHandler{
		businessSrv: businessSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/statistics/printer", wrapper.Printer) // 打印
	}
}

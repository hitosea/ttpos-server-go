package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// OrderImportHandler 订单导入处理程序
type OrderImportHandler struct {
	orderImportSrv service.IOrderImportSrv
}

// ImportOrder 处理订单导入
// @Summary 导入订单数据
// @Description 通过 Excel 文件导入订单数据
// @Tags 商家端.订单导入
// @Accept multipart/form-data
// @Produce json
// @Security JwtToken
// @Param file formData file true "Excel 文件"
// @Success 200 {object} map[string]interface{} "导入结果"
// @Failure 400 {object} nil "参数错误"
// @Failure 500 {object} nil "服务器错误"
// @Router /shop/order/import [post]
func (h *OrderImportHandler) ImportOrder(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "获取文件失败"))
		return
	}

	// 校验文件格式
	if file.Header.Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" &&
		file.Header.Get("Content-Type") != "application/octet-stream" {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("文件格式错误，仅支持 .xlsx 格式"))
		return
	}

	// 校验文件大小（最大 10MB）
	maxSize := int64(10 * 1024 * 1024)
	if file.Size > maxSize {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("文件大小不能超过 10MB"))
		return
	}

	// 调用 Service 导入订单
	successCount, failCount, failList, err := h.orderImportSrv.Import(ctx, file)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 返回结果
	result := map[string]interface{}{
		"success_count": successCount,
		"fail_count":    failCount,
		"fail_list":     failList,
	}

	helper.Success(c, result)
}

// RegisterOrderImportHandlers 注册订单导入路由
func RegisterOrderImportHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	settingSrv := setting.NewSrv(dbm, cache)
	localeSrv := service.NewLocaleSrv()
	translateSrv := service.NewTranslateSrv(dbm, cache)
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, service.NewCashBoxSrv(dbm), service.WithSmsSrv(dbm))
	productSrv := service.NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv)
	orderImportSrv := service.NewOrderImportSrv(dbm, orderSrv, productSrv, memberSrv)

	// 初始化处理器
	handler := &OrderImportHandler{
		orderImportSrv: orderImportSrv,
	}

	// 订单导入接口（需要认证）
	privateApi := router.Group("", middleware.Auth(service.NewAuthSrv(dbm, service.NewCaptchaSrv(cache), service.NewRoleAccessSrv(dbm), service.NewDeviceSrv(settingSrv, dbm), service.NewStaffShiftSrv(cache, dbm, service.NewCashBoxSrv(dbm), service.NewStatisticsSrv()), settingSrv), dbm))
	{
		privateApi.POST("/order/import", handler.ImportOrder)
	}
}


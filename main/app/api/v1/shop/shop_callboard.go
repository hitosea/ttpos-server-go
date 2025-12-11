package shop

import (
	"errors"

	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/callboard"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type CallBoardHandler struct {
	service       callboard.ICallBoardService
	uploadFileSrv service.IUploadFileSrv // 文件上传服务
}

// BindDevice 绑定设备
// @Summary 通过绑定码绑定设备
// @Description 商家输入6位绑定码绑定展示设备
// @Tags 商家端.叫号展示
// @Accept json
// @Produce json
// @Security Bearer
// @Param data body req.BindDeviceReq true "绑定请求"
// @Success 200 {object} dto.Response
// @Router /shop/callboard/device/bind [post]
func (h *CallBoardHandler) BindDevice(c *gin.Context) {
	var reqData req.BindDeviceReq
	if err := c.ShouldBindJSON(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()
	clientVersion := ctx.GetVersion()

	err := h.service.BindDevice(c.Request.Context(), companyUuid, reqData, clientVersion)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, nil, "绑定成功")
}

// UpdateBindInfo 更新绑定信息
// @Summary 更新信息
// @Description 更新已绑定的叫号展示设备信息
// @Tags 商家端.叫号展示
// @Accept json
// @Produce json
// @Security Bearer
// @Param data body req.UpdateBindInfoReq true "更新请求"
// @Success 200 {object} dto.Response
// @Router /shop/callboard/device/update [post]
func (h *CallBoardHandler) UpdateBindInfo(c *gin.Context) {
	var reqData req.UpdateBindInfoReq
	if err := c.ShouldBindJSON(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()
	clientVersion := ctx.GetVersion()

	err := h.service.UpdateBindInfo(c.Request.Context(), companyUuid, reqData, clientVersion)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, nil, "更新成功")
}

// GetDeviceList 获取设备列表
// @Summary 获取已绑定设备列表
// @Description 获取商家已绑定的叫号展示设备列表
// @Tags 商家端.叫号展示
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.Response{data=resp.DeviceListResp}
// @Router /shop/callboard/device/list [get]
func (h *CallBoardHandler) GetDeviceList(c *gin.Context) {
	var reqData req.GetDeviceListReq
	if err := c.ShouldBindQuery(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()

	resp, err := h.service.GetDeviceList(c.Request.Context(), companyUuid, reqData)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, resp)
}

// UnbindDevice 解绑设备
// @Summary 解绑展示设备
// @Description 解绑已绑定的叫号展示设备
// @Tags 商家端.叫号展示
// @Accept json
// @Produce json
// @Security Bearer
// @Param data body req.UnbindDeviceReq true "解绑请求"
// @Success 200 {object} dto.Response
// @Router /shop/callboard/device [delete]
func (h *CallBoardHandler) UnbindDevice(c *gin.Context) {
	var reqData req.UnbindDeviceReq
	if err := c.ShouldBindJSON(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	ctx := helper.GetContext(c)
	companyUuid := ctx.GetCompanyUuid()

	err := h.service.UnbindDevice(c.Request.Context(), companyUuid, reqData.Uuid)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, nil)
}

// UploadBackgroundImage 上传叫号系统背景图片
// @Summary 上传叫号系统背景图片
// @Description 上传叫号系统背景图片，支持 JPEG、PNG、WEBP 格式，大小限制 20MB
// @Tags 商家端.叫号展示
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "背景图片文件"
// @Success 200 {object} dto.Response{data=resp.UploadFileResp}
// @Router /shop/callboard/upload_background_image [post]
func (h *CallBoardHandler) UploadBackgroundImage(c *gin.Context) {
	ctx := helper.GetContext(c)
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 检查文件大小（20MB = 20 * 1024 * 1024 bytes）
	maxSizeBytes := int64(20 * 1024 * 1024)
	if file.Size > maxSizeBytes {
		helper.ErrorWithMessage(c, constant.CodeFail, errors.New("文件大小不能超过20MB"))
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	uploadFileResp, err := h.uploadFileSrv.UploadImage(ctx, fileReader, file.Filename, file.Size, 0, "callboardBackground")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, uploadFileResp)
}

// RegisterCallBoardHandlers 注册叫号展示相关路由
func RegisterCallBoardHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	callBoardSrv := callboard.NewCallBoardService(dbm, cache)
	handler := &CallBoardHandler{
		service:       callBoardSrv,
		uploadFileSrv: service.NewUploadFileSrv(dbm),
	}
	privateApi := router.Group("/callboard", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/device/bind", handler.BindDevice)
		privateApi.POST("/device/update", handler.UpdateBindInfo)
		privateApi.GET("/device/list", handler.GetDeviceList)
		privateApi.DELETE("/device", handler.UnbindDevice)
		privateApi.POST("/upload_background_image", handler.UploadBackgroundImage)
	}
}

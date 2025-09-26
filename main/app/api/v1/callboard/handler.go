package callboard

import (
	"errors"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service/callboard"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	ttcontext "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service callboard.ICallBoardService
}

// GetBindCode 获取绑定码
// @Summary 获取设备绑定码
// @Description 展示设备获取6位数字绑定码，有效期10分钟
// @Tags 叫号展示.设备端
// @Accept json
// @Produce json
// @Param data body req.GetBindCodeReq true "请求参数"
// @Success 200 {object} dto.Response{data=resp.BindCodeResp}
// @Router /callboard/device/bind_code [get]
func (h *Handler) GetBindCode(c *gin.Context) {
	var reqData req.GetBindCodeReq
	if err := c.ShouldBindQuery(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	resp, err := h.service.GetBindCode(c.Request.Context(), reqData)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, resp)
}

// GetBindInfo 获取绑定信息
// @Summary 获取设备绑定信息
// @Description 查询设备绑定状态，返回商户ID和设备密钥
// @Tags 叫号展示.设备端
// @Accept json
// @Produce json
// @Param device_id query string true "设备ID"
// @Success 200 {object} dto.Response{data=resp.BindInfoResp}
// @Router /callboard/device/bind_info [get]
func (h *Handler) GetBindInfo(c *gin.Context) {
	var reqData req.GetBindInfoReq
	if err := c.ShouldBindQuery(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}

	resp, err := h.service.GetBindInfo(c.Request.Context(), reqData)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	if resp == nil {
		helper.ErrorWithMessage(c, constant.CodeFail, errors.New("设备未绑定"))
		return
	}

	helper.Success(c, resp)
}

// GetQueueData 获取队列数据
// @Summary 获取排队队列数据
// @Description 获取等待队列和取餐队列数据
// @Tags 叫号展示.设备端
// @Accept json
// @Produce json
// @Param device_id query string true "设备ID"
// @Param limit query int true "限制数量"
// @Param update_time query int true "更新时间,unix时间戳"
// @Success 200 {object} dto.Response{data=resp.QueueDataResp}
// @Router /callboard/data [get]
func (h *Handler) GetQueueData(c *gin.Context) {
	var reqData req.GetQueueDataReq
	if err := c.ShouldBindQuery(&reqData); err != nil {
		helper.HandleValidationError(c, err, reqData, nil)
		return
	}
	companyUuid := ttcontext.GetCompanyUuid(c.Request.Context())
	resp, err := h.service.GetQueueData(c.Request.Context(), companyUuid, reqData)
	if err != nil {
		helper.ErrorWithMessage(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, resp)
}

// RegisterHandlers 注册路由
func RegisterHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	callBoardService := callboard.NewCallBoardService(dbm, cache)
	h := &Handler{
		service: callBoardService,
	}
	// 需要应用签名验证的接口
	{
		noAuthGroup := router.Group("")
		noAuthGroup.GET("/device/bind_code", h.GetBindCode)
		noAuthGroup.GET("/device/bind_info", h.GetBindInfo)
	}

	// 设备已绑定后的接口
	{
		deviceGroup := router.Group("", middleware.BindedDeviceAuth(callBoardService))
		deviceGroup.GET("/data", h.GetQueueData)
	}
}

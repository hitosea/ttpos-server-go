package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// SoldOutHandler 沽清相关控制器
type SoldOutHandler struct {
	soldOutSrv service.ISoldOutSrv
}

// SoldOutList 沽清售罄列表
// @Summary 沽清售罄列表
// @Description 沽清售罄列表
// @Tags 收银端.沽清
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Success 200 {object} dto.Response{data=resp.SoldOutPaginationResp}
// @Router /cashier/sold_out/list [get]
func (h *SoldOutHandler) SoldOutList(c *gin.Context) {
	// 绑定请求参数
	var soldOutListReq req.SoldOutListReq
	if err := c.ShouldBindQuery(&soldOutListReq); err != nil {
		helper.HandleValidationError(c, err, soldOutListReq, dto.PageReqMessage)
		return
	}
	// 获取沽清列表
	res, err := h.soldOutSrv.GetSoldOutList(helper.GetCompanyUuid(c), soldOutListReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddSoldOut 添加沽清商品
// @Summary 添加沽清商品
// @Description 添加沽清商品
// @Tags 收银端.沽清
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddSoldOutReq true "添加沽清商品参数"
// @Success 200 {object} dto.Response
// @Router /cashier/sold_out/add [post]
func (h *SoldOutHandler) AddSoldOut(c *gin.Context) {
	// 绑定请求参数
	var addSoldOutReq req.AddSoldOutReq
	if err := c.ShouldBindJSON(&addSoldOutReq); err != nil {
		helper.HandleValidationError(c, err, addSoldOutReq, nil)
		return
	}
	// 获取沽清列表
	err := h.soldOutSrv.AddSoldOut(helper.GetCompanyUuid(c), addSoldOutReq.SoldOutData)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "操作成功")
}

// CancelSoldOut 取消沽清商品
// @Summary 取消沽清商品
// @Description 取消沽清商品
// @Tags 收银端.沽清
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CancelSoldOutReq true "取消沽清商品参数"
// @Success 200 {object} dto.Response
// @Router /cashier/sold_out/cancel [post]
func (h *SoldOutHandler) CancelSoldOut(c *gin.Context) {
	// 绑定请求参数
	var cancelSoldOut req.CancelSoldOutReq
	if err := c.ShouldBindJSON(&cancelSoldOut); err != nil {
		helper.HandleValidationError(c, err, cancelSoldOut, nil)
		return
	}
	// 取消沽清商品
	err := h.soldOutSrv.CancelSoldOut(helper.GetCompanyUuid(c), cancelSoldOut.ProductBomUuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "取消成功")
}

// CancelAllSoldOut 全部取消沽清商品
// @Summary 全部取消沽清商品
// @Description 全部取消沽清商品
// @Tags 收银端.沽清
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /cashier/sold_out/cancel_all [post]
func (h *SoldOutHandler) CancelAllSoldOut(c *gin.Context) {
	// 取消沽清全部商品
	err := h.soldOutSrv.CancelAllSoldOut(helper.GetCompanyUuid(c))
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{}, "取消成功")
}

// GetSettings 获取商品沽清设置
// @Summary 获取商品沽清设置
// @Description 获取指定商品的沽清设置信息
// @Tags 收银端.沽清
// @Accept json
// @Produce json
// @Security JwtToken
// @Param product_package_uuid query uint64 true "商品包ID"
// @Success 200 {object} dto.Response{data=resp.SoldOutSettingsResp}
// @Router /cashier/sold_out/settings [get]
func (h *SoldOutHandler) GetSettings(c *gin.Context) {
	// 绑定请求参数
	var req req.GetSoldOutSettingsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	// 获取沽清设置
	respData, err := h.soldOutSrv.GetSettings(helper.GetCompanyUuid(c), &req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, respData)
}

func RegisterSoldOutHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 初始化处理器
	wrapper := SoldOutHandler{
		soldOutSrv: service.NewSoldOutSrv(dbm, service.NewLocaleSrv()),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/sold_out/list", wrapper.SoldOutList)             // 沽清售罄列表
		privateApi.POST("/sold_out/add", wrapper.AddSoldOut)              // 添加沽清商品
		privateApi.POST("/sold_out/cancel", wrapper.CancelSoldOut)        // 取消沽清商品
		privateApi.POST("/sold_out/cancel_all", wrapper.CancelAllSoldOut) // 取消全部沽清商品
		privateApi.GET("/sold_out/settings", wrapper.GetSettings)         // 获取商品沽清设置
	}
}

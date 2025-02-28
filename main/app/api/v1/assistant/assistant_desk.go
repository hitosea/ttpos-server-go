package assistant

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	Service service.IDeskSrv // 主服务
}

// GetDeskRegionAndType 处理获取桌台的区域和类型
// @Summary 获取桌台的区域和类型
// @Description 获取桌台的区域和类型
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "桌台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取桌台的区域和类型的逻辑
	res, err := h.Service.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {array} resp.DeskListWithPaginationResp "桌台列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var deskListReq req.DeskListReq
	if err := c.ShouldBindQuery(&deskListReq); err != nil {
		helper.HandleValidationError(c, err, deskListReq, dto.PageReqMessage)
		return
	}
	// 获取收银产品列表
	res, err := h.Service.GetDeskList(ctx, companyId, deskListReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, apperrors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取桌台详情
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 绑定请求参数
	var deskInfoReq req.DeskInfoReq
	if err := c.ShouldBindQuery(&deskInfoReq); err != nil {
		helper.HandleValidationError(c, err, deskInfoReq, nil)
		return
	}
	// 获取收银产品列表
	res, err := h.Service.GetDeskInfo(companyId, deskInfoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, apperrors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterDeskHandlers 注册收银产品路由
func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv)

	// 创建处理程序
	wrapper := DeskHandler{
		Service: service.NewDeskSrv(
			dbm,        // 数据库管理器
			localeSrv,  // 多语言服务
			orderSrv,   // 订单服务
			settingSrv, // 设置服务
			deviceSrv,  // 设备服务
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)
		privateApi.GET("/desk/list", wrapper.GetDeskList)
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)
	}
}

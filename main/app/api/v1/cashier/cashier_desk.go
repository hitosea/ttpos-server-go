package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
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

// GetCashierDeskRegionAndType 处理获取收银台的区域和类型
// @Summary 获取收银台的区域和类型
// @Description 获取收银台的区域和类型
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Success 200 {object} cashier_resp.DeskRegionAndTypeListWithPaginationResp "收银台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取收银台的区域和类型的逻辑
	res, err := h.Service.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetCashierDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.DeskListReq true "列表参数"
// @Success 200 {array} cashier_resp.DeskListWithPaginationResp "收银台列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.DeskListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取收银产品列表
	res, err := h.Service.GetDeskList(companyUuid, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetCashierDeskList 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.DeskInfoReq true "详情参数"
// @Success 200 {object} cashier_resp.DeskInfoResp "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	req := req.DeskInfoReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 获取收银产品列表
	res, err := h.Service.GetDeskInfo(companyUuid, req.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterProductHandlers 注册收银产品路由
func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	// 初始化处理器
	wrapper := DeskHandler{
		Service: service.NewDeskSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
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

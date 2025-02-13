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
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "收银台区域和类型列表"
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
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {array} resp.DeskListWithPaginationResp "收银台列表"
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
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.DeskInfoResp "桌台详情"
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

// CreateDeskOrder 处理创建桌台订单
// @Summary 创建桌台订单
// @Description 创建桌台订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data body req.DeskOrderCreateReq true "创建桌台订单参数"
// @Success 200 {object} resp.CreateDeskOrderResp "创建桌台订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/create [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.Service.CreateDeskOrder(helper.GetCompanyUuid(c), params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// IsCellCloseDesk 判断桌台是否可关闭
// @Summary 判断桌台是否可关闭
// @Description 判断桌台是否可关闭
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data query req.DeskInfoReq true "详情参数"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/is_cell_close [get]
func (h *DeskHandler) IsCellCloseDesk(c *gin.Context) {
	params := req.DeskInfoReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	if _, err := h.Service.IsCellCloseDesk(helper.GetCompanyUuid(c), params.Uuid); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// todo 获取已经送厨的商品 - 等王总写完拿来用

	// 返回结果
	helper.Success(c, gin.H{})
}

// CloseDesk 处理关闭桌台
// @Summary 关闭桌台
// @Description 关闭桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @param data query req.DeskCloseReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/close [post]
func (h *DeskHandler) CloseDesk(c *gin.Context) {
	// 绑定请求参数
	params := req.DeskCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	err := h.Service.CloseDesk(helper.GetCompanyUuid(c), params.Uuid, params.Reason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
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
	localeSrv := service.NewLocaleSrv()
	orderSrv := service.NewOrderSrv(dbm, localeSrv, cache)

	// 初始化处理器
	wrapper := DeskHandler{
		Service: service.NewDeskSrv(
			dbm,        // 数据库管理器
			localeSrv,  // 多语言服务
			orderSrv,   // 订单服务
			settingSrv, // 设置服务
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)
		privateApi.GET("/desk/list", wrapper.GetDeskList)
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)
		privateApi.GET("/desk/is_cell_close", wrapper.IsCellCloseDesk)
		privateApi.POST("/desk/close", wrapper.CloseDesk)
		privateApi.POST("/desk/order/create", wrapper.CreateDeskOrder)
	}
}

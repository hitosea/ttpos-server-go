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
	service      service.IDeskSrv // 主服务
	orderService service.IOrderSrv
}

// GetDeskRegionAndType 处理获取桌台的区域和类型
// @Summary 获取桌台的区域和类型
// @Description 获取桌台的区域和类型
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "桌台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取桌台的区域和类型的逻辑
	res, err := h.service.GetDeskRegionAndTypeList(companyId)
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
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, err := h.service.GetDeskList(companyUuid, req)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
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
	res, err := h.service.GetDeskInfo(companyUuid, req.Uuid)
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
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "创建桌台订单参数"
// @Success 200 {object} resp.CreateDeskOrderResp "创建桌台订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/create [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.service.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CloseDesk 处理关闭桌台
// @Summary 关闭桌台
// @Description 关闭桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskCloseReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/close [post]
func (h *DeskHandler) CloseDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.service.CloseDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CancelDeskOrder 处理取消桌台订单
// @Summary 取消桌台订单
// @Description 取消桌台订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cancel [post]
func (h *DeskHandler) CancelDeskOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	req := req.OrderCancelReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	//
	err := h.orderService.CancelOrder(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductDelete 处理删除桌台订单商品
// @Summary 删除桌台订单商品
// @Description 删除桌台订单商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/delete [delete]
func (h *DeskHandler) OrderProductDelete(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	staff := helper.GetStaff(c)
	source := helper.GetSource(c)
	// 绑定请求参数
	params := req.OrderProductDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductDelete(companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductChangePrice 处理桌台订单商品改价
// @Summary 桌台订单商品改价
// @Description 桌台订单商品改价
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/price [post]
func (h *DeskHandler) OrderProductChangePrice(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	staff := helper.GetStaff(c)
	source := helper.GetSource(c)
	// 绑定请求参数
	params := req.OrderProductChangePriceReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductChangePrice(companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderChangePopulation 处理桌台订单修改人数
// @Summary 桌台订单修改人数
// @Description 桌台订单修改人数
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/population [post]
func (h *DeskHandler) OrderChangePopulation(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	staff := helper.GetStaff(c)
	source := helper.GetSource(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderChangePopulation(companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/remark [post]
func (h *DeskHandler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	_, err := h.orderService.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// RegisterDeskHandlers 注册收银产品路由
func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv)

	// 初始化处理器
	wrapper := DeskHandler{
		service: service.NewDeskSrv(
			dbm,        // 数据库管理器
			localeSrv,  // 多语言服务
			orderSrv,   // 订单服务
			settingSrv, // 设置服务
		),
		orderService: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)
		privateApi.GET("/desk/list", wrapper.GetDeskList)
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)
		privateApi.POST("/desk/close", wrapper.CloseDesk)
		privateApi.POST("/desk/order/create", wrapper.CreateDeskOrder)
		privateApi.POST("/desk/order/cancel", wrapper.CancelDeskOrder)
		privateApi.DELETE("/desk/order/product/delete", wrapper.OrderProductDelete)
		privateApi.POST("/desk/order/product/price", wrapper.OrderProductChangePrice)
		privateApi.POST("/desk/order/population", wrapper.OrderChangePopulation)
		privateApi.POST("/desk/order/product/remark", wrapper.OrderProductRemark)
	}
}

package cashier

import (
	"errors"
	"strconv"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// InstantHandler 收银点餐处理程序
type InstantHandler struct {
	orderService service.IOrderSrv // 订单服务
}

// CreateInstantOrder 创建点餐订单
// @Summary 创建点餐订单
// @Description 创建点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.CreateOrderResp
// @Router /cashier/instant/order/create [post]
func (h *InstantHandler) CreateInstantOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 创建订单
	res, err := h.orderService.CreateInstantOrder(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("创建订单失败"))
		return
	}
	helper.Success(c, res)
}

// CancelOrder 处理取消点餐订单
// @Summary 取消点餐订单
// @Description 取消点餐订单
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil "取消点餐订单成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cancel [post]
func (h *InstantHandler) CancelOrder(c *gin.Context) {
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

// OrderProductDelete 处理删除点餐订单商品
// @Summary 删除点餐订单商品
// @Description 删除点餐订单商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/delete [delete]
func (h *InstantHandler) OrderProductDelete(c *gin.Context) {
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

// OrderProductChangePrice 处理删除点餐订单商品改价
// @Summary 点餐订单商品改价
// @Description 点餐订单商品改价
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/price [post]
func (h *InstantHandler) OrderProductChangePrice(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	source := helper.GetSource(c)
	staff := helper.GetStaff(c)
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

// OrderChangePopulation 处理点餐订单修改人数
// @Summary 点餐订单修改人数
// @Description 点餐订单修改人数
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/population [post]
func (h *InstantHandler) OrderChangePopulation(c *gin.Context) {
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

// OrderProductRemark 处理点餐订单商品备注
// @Summary 点餐订单商品备注
// @Description 点餐订单商品备注
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/product/remark [post]
func (h *InstantHandler) OrderProductRemark(c *gin.Context) {
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

// OrderCartInfo 查询点餐购物车信息
// @Summary 查询点餐购物车信息
// @Description 查询点餐购物车信息
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Param id path string true "账单ID"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/info [get]
func (h *InstantHandler) OrderCartInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	saleBillUuid, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	res, err := h.orderService.GetOrderCartInfo(ctx, saleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 收银端.点餐
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/instant/order/cart/product/add [post]
func (h *InstantHandler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 添加商品
	res, err := h.orderService.OrderCartProductAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterInstantHandlers 注册收银订单路由
func RegisterInstantHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv)

	// 创建收银产品处理程序
	wrapper := InstantHandler{
		orderService: orderSrv, // 订单服务
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/instant/order/create", wrapper.CreateInstantOrder)             // 创建点餐订单
		privateApi.POST("/instant/order/cancel", wrapper.CancelOrder)                    // 取消点餐订单
		privateApi.DELETE("/instant/order/product/delete", wrapper.OrderProductDelete)   // 删除点餐订单商品
		privateApi.POST("/instant/order/product/price", wrapper.OrderProductChangePrice) // 点餐订单商品改价
		privateApi.POST("/instant/order/population", wrapper.OrderChangePopulation)      // 点餐订单修改人数
		privateApi.POST("/instant/order/product/remark", wrapper.OrderProductRemark)     // 点餐订单商品备注
		privateApi.GET("/instant/order/cart/info", wrapper.OrderCartInfo)                // 查询点餐购物车信息
		privateApi.POST("/instant/order/cart/product/add", wrapper.OrderCartProductAdd)  // 向购物车添加商品
	}
}

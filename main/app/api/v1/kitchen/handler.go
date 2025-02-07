package kitchen

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

// PostCallHandle 处理呼叫
// @Summary 处理呼叫
// @Tags 呼叫
// @Accept json
// @Produce json
// @param data body req.ParamCustomerInfo true "id"
// @Success 200 {object} dto.Response
// @Router /kitchen/call/handle [post]
func (h *Handler) PostCallHandle(c *gin.Context) {
	// 处理呼叫的逻辑
}

// GetCallList 获取呼叫列表
// @Summary 获取呼叫列表
// @Tags 呼叫
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=resp.CallList}
// @Router /kitchen/call/list [get]
func (h *Handler) GetCallList(c *gin.Context) {
	// 获取呼叫列表的逻辑
	helper.Fail(c, constant.CodeBadRequest, "验证码签名8860不能为空")
}

// PostLogin 用户登录
// @Summary 用户登录
// @Tags 登录
// @Accept json
// @Produce json
// @param data body req.ParamLogin true "登陆参数"
// @Success 200 {object} dto.Response{data=resp.LoginInfo}
// @Router /kitchen/login [post]
func (h *Handler) PostLogin(c *gin.Context) {
	// 用户登录的逻辑
}

// GetProductionFinishedProductHistory 获取上菜历史
// @Summary 获取上菜历史
// @Tags 生产
// @Accept json
// @Produce json
// @param language query string true "语言"
// @param pageNum query int true "页码"
// @Success 200 {object} dto.Response{data=resp.ServedProductHistory}
// @Router /kitchen/production/finished_product/history [get]
func (h *Handler) GetProductionFinishedProductHistory(c *gin.Context) {
	// 获取成品历史的逻辑
}

// GetProductionFinishedProductLatest 获取最新成品
// @Summary 获取最新成品
// @Tags 生产
// @Accept json
// @Produce json
// @param language query string true "语言"
// @Success 200 {object} dto.Response{data=resp.ServedProductList}
// @Router /kitchen/production/finished_product/latest [get]
func (h *Handler) GetProductionFinishedProductLatest(c *gin.Context) {
	// 获取最新成品的逻辑
}

// GetProductionOrderCategory 获取生产订单分类
// @Summary 获取生产订单分类
// @Tags 生产
// @Accept json
// @Produce json
// @param categoryId query string true "分类ID。默认是0，全部"
// @param pageNum query int false "页码"
// @param language query string true "语言"
// @Success 200 {object} dto.Response{data=resp.ProductOrderListByCategory}
// @Router /kitchen/production/order/category [get]
func (h *Handler) GetProductionOrderCategory(c *gin.Context) {
	// 获取生产订单分类的逻辑
}

// GetProductionOrderList 获取生产订单列表
// @Summary 获取生产订单列表
// @Tags 生产
// @Accept json
// @Produce json
// @param pageNum query int true "页码"
// @param language query string true "语言"
// @Success 200 {object} dto.Response{data=resp.ProductOrderList}
// @Router /kitchen/production/order/list [get]
func (h *Handler) GetProductionOrderList(c *gin.Context) {
	// 获取生产订单列表的逻辑
}

// GetProductionOrderListByCategory 按分类获取生产订单列表
// @Summary 按分类获取生产订单列表
// @Tags 生产
// @Accept json
// @Produce json
// @param categoryId query string true "分类ID。默认是0，全部"
// @param pageNum query int true "页码"
// @param language query string true "语言"
// @Success 200 {object} dto.Response{data=resp.ProductOrderListByCategory}
// @Router /kitchen/production/order/list_by_category [get]
func (h *Handler) GetProductionOrderListByCategory(c *gin.Context) {
	// 按分类获取生产订单列表的逻辑
}

// GetSettingInfo 获取设置信息
// @Summary 获取设置信息
// @Tags 设置
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=resp.SettingInfo}
// @Router /kitchen/setting/info [get]
func (h *Handler) GetSettingInfo(c *gin.Context) {
	// 获取设置信息的逻辑
}

// PostSettingSave 保存设置
// @Summary 保存设置
// @Tags 设置
// @Accept json
// @Produce json
// @param data body req.KitchenSetting true "设置参数"
// @Success 200 {object} dto.Response{data=resp.SettingInfo}
// @Router /kitchen/setting/save [post]
func (h *Handler) PostSettingSave(c *gin.Context) {
	// 保存设置的逻辑
}

// PostSettingVerifyAdvancedPassword 验证高级密码
// @Summary 验证高级密码
// @Tags 设置
// @Accept json
// @Produce json
// @param data body req.ParamVerifyAdvancedPassword true "验证参数"
// @Success 200 {object} dto.Response
// @Router /kitchen/setting/verify_advanced_password [post]
func (h *Handler) PostSettingVerifyAdvancedPassword(c *gin.Context) {
	// 验证高级密码的逻辑
}

// RegisterHandlers 创建与 OpenAPI 规范匹配的 http.Handler。
func RegisterHandlers(router gin.IRouter) {
	wrapper := Handler{}

	router.POST("/login", wrapper.PostLogin) // 用户登录

	router.POST("/call/handle", wrapper.PostCallHandle)                                             // 处理呼叫
	router.GET("/call/list", wrapper.GetCallList)                                                   // 获取呼叫列表
	router.GET("/production/finished_product/history", wrapper.GetProductionFinishedProductHistory) // 获取上菜历史
	router.GET("/production/finished_product/latest", wrapper.GetProductionFinishedProductLatest)   // 获取最新成品
	router.GET("/production/order/category", wrapper.GetProductionOrderCategory)                    // 获取生产订单分类
	router.GET("/production/order/list", wrapper.GetProductionOrderList)                            // 获取生产订单列表
	router.GET("/production/order/list_by_category", wrapper.GetProductionOrderListByCategory)      // 按分类获取生产订单列表
	router.GET("/setting/info", wrapper.GetSettingInfo)                                             // 获取设置信息
	router.POST("/setting/save", wrapper.PostSettingSave)                                           // 保存设置
	router.POST("/setting/verify_advanced_password", wrapper.PostSettingVerifyAdvancedPassword)     // 验证高级密码
}

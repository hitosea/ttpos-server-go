package assistant

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	deskSrv  service.IDeskSrv
	orderSrv service.IOrderSrv
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
	res, err := h.deskSrv.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
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
	res, err := h.deskSrv.GetDeskList(ctx, companyId, deskListReq)
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
	res, err := h.deskSrv.GetDeskInfo(companyId, deskInfoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, apperrors.ErrInternal)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CreateDeskOrder 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/open [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {

	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.deskSrv.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		ctx.Log().Error("创建桌台订单失败", zap.Error(err))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/cart/product/add [post]
func (h *DeskHandler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 添加商品。 若没有点餐账单则新建一个
	res, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// ChangeDesk 处理切换桌台
// @Summary 切换桌台
// @Description 切换桌台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ChangeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/change [post]
func (h *DeskHandler) ChangeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.ChangeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, err := h.deskSrv.ChangeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// CompleteDesk 处理清台
// @Summary 清台
// @Description 清台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskInfoReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/complete [post]
func (h *DeskHandler) CompleteDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskJsonUuidReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.deskSrv.CompleteDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// MergeDesk 处理合并桌台
// @Summary 合并桌台
// @Description 合并桌台
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.MergeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/merge [post]
func (h *DeskHandler) MergeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MergeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, deskMergeCheckResp, err := h.deskSrv.MergeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, deskMergeCheckResp, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangeBuffet 处理桌台订单调整自助餐
// @Summary 桌台订单调整自助餐
// @Description 桌台订单调整自助餐
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangeBuffetReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/buffet [post]
func (h *DeskHandler) OrderChangeBuffet(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangeBuffet(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangePopulation 处理桌台订单修改人数
// @Summary 桌台订单修改人数
// @Description 桌台订单修改人数
// @Tags 点餐助手端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/order/population [post]
func (h *DeskHandler) OrderChangePopulation(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangePopulation(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
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
		deskSrv:  service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv),
		orderSrv: orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)    // 获取桌台的区域和类型
		privateApi.GET("/desk/list", wrapper.GetDeskList)                        // 获取桌台列表
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)                        // 获取桌台详情
		privateApi.POST("/desk/open", wrapper.CreateDeskOrder)                   // 创建桌台订单(开桌)
		privateApi.POST("/desk/order/population", wrapper.OrderChangePopulation) // 桌台订单修改人数
		privateApi.POST("/desk/order/buffet", wrapper.OrderChangeBuffet)         // 桌台订单调整自助餐
		privateApi.POST("/desk/change", wrapper.ChangeDesk)                      // 切换桌台（转台）
		privateApi.POST("/desk/complete", wrapper.CompleteDesk)                  // 完成桌台（清台）
		privateApi.POST("/desk/merge", wrapper.MergeDesk)                        // 合并桌台

		privateApi.POST("/desk/order/cart/product/add", wrapper.OrderCartProductAdd) // 向购物车添加商品
	}
}

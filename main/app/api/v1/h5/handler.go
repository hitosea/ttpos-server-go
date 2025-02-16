package h5

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// H5Handler 商家H5端处理程序
type H5Handler struct {
	service   service.IH5Srv // h5扫码服务
	deskSrv   service.IDeskSrv
	buffetSrv service.IBuffetSrv
}

// GetBaseInfo 获取桌码基础信息
// @Summary 桌码基础信息
// @Description 获取桌码基础信息，整个h5应用需要的基础信息
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.GetBaseInfoResponse{}
// @Router /index.php/scan/base.base/getInfo [post]
func (h H5Handler) GetBaseInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deskUuid uint64
	info, err := h.service.GetCompanyInfo(ctx, deskUuid)
	if err != nil {
		Fail(c, 500, "获取信息失败")
		return
	}
	Success(c, info)
}

// GetBaseInfo 获取自助餐套餐列表
// @Summary 自助餐套餐信息
// @Description 获取自助餐套餐列表
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.H5Response{data=resp.H5BuffetList}
// @Router /index.php/scan/order.Order/buffetList [post]
func (h *H5Handler) GetBuffetList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deskUuid uint64
	list, err := h.service.GetBuffetList(ctx, deskUuid)
	if err != nil {
		Fail(c, 500, "获取信息失败")
		return
	}
	Success(c, list)
}

// GetOpenDesk 开台
// @Summary 开台
// @Description 开台
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OpenDeskRequest true "创建扫码桌台订单参数"
// @Success 200 {object} resp.H5Response{data=resp.H5BuffetList}
// @Router /index.php/scan/order.Order/setTable [post]
func (h *H5Handler) GetOpenDesk(c *gin.Context) {
	ctx := helper.GetContext(c)

	var deskUuid uint64
	// 绑定请求参数
	params := req.OpenDeskRequest{}
	if err := c.ShouldBind(&params); err != nil {
		Fail(c, 500, "参数错误")
		return
	}

	err := h.service.OpenH5Desk(ctx, deskUuid, req.OpenDeskRequest{})
	if err != nil {
		Fail(c, 500, "获取信息失败")
		return
	}
	Success(c, nil)
}

// GetOpenDesk 获取商品分类
// @Summary 获取商品分类
// @Description 获取商品分类
// @Tags 扫码点餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.H5Response{data=resp.H5CategoryList}
// @Router /scan/product.category/index [post]
func (h *H5Handler) GetCategoryList(c *gin.Context) {
	ctx := helper.GetContext(c)
	list, err := h.service.GetCategoryList(ctx)
	if err != nil {
		Fail(c, 500, "获取信息失败")
		return
	}
	Success(c, list)
}

// RegisterH5Handlers 注册扫码h5路由
func RegisterH5Handlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	h5Srv := service.NewH5Srv(dbm, localeSrv, settingSrv)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv)
	deskSrv := service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv)
	buffetSrv := service.NewBuffetSrv(dbm, localeSrv)

	// 初始化处理器
	wrapper := H5Handler{
		service:   h5Srv,
		deskSrv:   deskSrv,
		buffetSrv: buffetSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.POST("/index.php/scan/base.base/getInfo", wrapper.GetBaseInfo)
		privateApi.POST("/index.php/scan/order.Order/buffetList", wrapper.GetBuffetList)
		privateApi.POST("/index.php/scan/order.Order/setTable", wrapper.GetOpenDesk)
	}
}

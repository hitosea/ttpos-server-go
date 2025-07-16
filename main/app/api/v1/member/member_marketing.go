package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MarketingHandler 营销活动控制器
type MarketingHandler struct {
	marketingActivitySrv service.IMarketingActivitySrv
}

// MarketingActivity 获取营销活动
// @Summary 获取营销活动
// @Description 获取营销活动
// @Tags 会员端.营销活动
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=member_resp.MemberMarketingActivityListResp}
// @Router /member/marketing_activity [get]
func (h *MarketingHandler) MarketingActivity(c *gin.Context) {
	ctx := helper.GetContext(c)
	marketingActivityResp, err := h.marketingActivitySrv.MarketingActivity(ctx)
	if err != nil {
		if marketingActivityResp != nil {
			helper.ErrorWithData(c, constant.CodeFail, marketingActivityResp, err)
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, marketingActivityResp)
}

// MarketingActivity 获取营销活动列表
// @Summary 获取营销活动列表
// @Description 获取营销活动列表
// @Tags 会员端.营销活动
// @Accept json
// @Produce json
// @Success 200 {object} dto.Response{data=member_resp.MemberMarketingActivityListResp}
// @Router /member/marketing_activity_list [get]
func (h *MarketingHandler) MarketingActivityList(c *gin.Context) {
	ctx := helper.GetContext(c)
	marketingActivityResp, err := h.marketingActivitySrv.MarketingActivityList(ctx)
	if err != nil {
		if marketingActivityResp != nil {
			helper.ErrorWithData(c, constant.CodeFail, marketingActivityResp, err)
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, marketingActivityResp)
}

func RegisterMarketingHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	marketingActivitySrv := service.NewMarketingActivitySrv(dbm, cache)
	wrapper := &MarketingHandler{
		marketingActivitySrv: marketingActivitySrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.GET("/marketing_activity", wrapper.MarketingActivity)          // 获取营销活动
		privateApi.GET("/marketing_activity_list", wrapper.MarketingActivityList) // 获取营销活动列表
	}
}

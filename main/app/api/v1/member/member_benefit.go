package member

import (
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

// BenefitHandler 营销活动控制器
type BenefitHandler struct {
	memberSrv service.IMemberSrv
}

// BenefitActivity 获取我的优惠券
// @Summary 获取我的优惠券
// @Description 获取我的优惠券
// @Tags 会员端.优惠券
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=member_resp.CouponListWithPaginationResp}
// @Router /member/coupon/list [get]
func (h *BenefitHandler) GetMemberCouponList(c *gin.Context) {
	ctx := helper.GetContext(c)
	couponListReq := req.CouponListReq{}
	couponListReq.MemberUuid = ctx.GetMemberUuid()
	couponResp, err := h.memberSrv.GetMemberCouponList(ctx, couponListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, couponResp)
}

// BenefitActivity 获取我的历史优惠券
// @Summary 获取我的历史优惠券
// @Description 获取我的历史优惠券
// @Tags 会员端.优惠券
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=member_resp.CouponListWithPaginationResp}
// @Router /member/coupon/history/list [get]
func (h *BenefitHandler) GetMemberCouponHistoryList(c *gin.Context) {
	ctx := helper.GetContext(c)
	couponListReq := req.CouponListReq{}
	couponListReq.MemberUuid = ctx.GetMemberUuid()
	couponListReq.IsHistory = 1
	couponResp, err := h.memberSrv.GetMemberCouponList(ctx, couponListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, couponResp)
}

func RegisterBenefitHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	memberSrv := service.NewMemberSrv(dbm)
	wrapper := &BenefitHandler{
		memberSrv: memberSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.GET("/coupon/list", wrapper.GetMemberCouponList)                // 获取我的优惠券
		privateApi.GET("/coupon/history/list", wrapper.GetMemberCouponHistoryList) // 获取我的历史优惠券
	}
}

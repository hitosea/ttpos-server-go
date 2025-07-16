package member

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证鉴权控制器
type AddressHandler struct {
	addressSrv service.IMemberAddressSrv
}

// AddressList 获取地址列表
// @Summary 获取地址列表
// @Description 获取地址列表
// @Tags 会员端.地址
// @Accept json
// @Produce json
// @param data query member_req.MemberAddressListReq true "详情参数"
// @Success 200 {object} dto.Response{data=member_resp.MemberAddressListResp}
// @Router /member/address [get]
func (h *AddressHandler) AddressList(c *gin.Context) {
	ctx := helper.GetContext(c)
	addressListReq := member_req.MemberAddressListReq{}
	if err := c.ShouldBindQuery(&addressListReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	addressListResp, err := h.addressSrv.GetAddressList(ctx, addressListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeUnauthorized, err)
		return
	}
	helper.Success(c, addressListResp)
}

// AddressAdd 添加地址
// @Summary 添加地址
// @Description 添加地址
// @Tags 会员端.地址
// @Accept json
// @Produce json
// @param data body member_req.MemberAddressAddReq true "详情参数"
// @Success 200 {object} dto.Response{}
// @Router /member/address/add [post]
func (h *AddressHandler) AddressAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	addressAddReq := member_req.MemberAddressAddReq{}
	if err := c.ShouldBindJSON(&addressAddReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.addressSrv.AddAddress(ctx, addressAddReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// AddressUpdate 更新地址
// @Summary 更新地址
// @Description 更新地址
// @Tags 会员端.地址
// @Accept json
// @Produce json
// @param data body member_req.MemberAddressUpdateReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.LoginResp}
// @Router /member/address/update [post]
func (h *AddressHandler) AddressUpdate(c *gin.Context) {
	ctx := helper.GetContext(c)
	addressUpdateReq := member_req.MemberAddressUpdateReq{}
	if err := c.ShouldBindJSON(&addressUpdateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.addressSrv.UpdateAddress(ctx, addressUpdateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeLoginFailed, err)
		return
	}
	helper.Success(c, gin.H{})
}

// @Summary 删除地址
// @Description 删除地址
// @Tags 会员端.地址
// @Accept json
// @Produce json
// @param data body member_req.MemberAddressDeleteReq true "详情参数"
// @Success 200 {object} dto.Response{}
// @Router /member/address/delete [delete]
func (h *AddressHandler) AddressDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	addressDeleteReq := member_req.MemberAddressDeleteReq{}
	if err := c.ShouldBindJSON(&addressDeleteReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.addressSrv.DeleteAddress(ctx, addressDeleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{})
}

// AddressAuth 认证地址手机号
// @Summary 认证地址手机号
// @Description 认证地址, 如果需要注册则注册会员并返回注册会员的token（前端需要更换token为当前token）
// @Tags 会员端.地址
// @Accept json
// @Produce json
// @param data body member_req.MemberAddressAuthReq true "详情参数"
// @Success 200 {object} dto.Response{data=member_resp.LoginResp}
// @Router /member/address/auth [post]
func (h *AddressHandler) AddressAuth(c *gin.Context) {
	ctx := helper.GetContext(c)
	addressAuthReq := member_req.MemberAddressAuthReq{}
	if err := c.ShouldBindJSON(&addressAuthReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	loginResp, err := h.addressSrv.AuthAddress(ctx, addressAuthReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, loginResp)
}

func RegisterAddressHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
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
	addressSrv := service.NewMemberAddressSrv(dbm, cache)
	wrapper := &AddressHandler{
		addressSrv: addressSrv,
	}
	// 需要认证
	privateApi := router.Group("", middleware.MemberAuth(authSrv, dbm))
	{
		privateApi.GET("/address", wrapper.AddressList)
		privateApi.POST("/address/add", wrapper.AddressAdd)
		privateApi.POST("/address/update", wrapper.AddressUpdate)
		privateApi.DELETE("/address/delete", wrapper.AddressDelete)
		privateApi.POST("/address/auth", wrapper.AddressAuth)
	}
}

package cashier

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant/jwt"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MemberHandler 结构体
type MemberHandler struct {
	authSrv   service.IAuthSrv
	memberSrv service.IMemberSrv
}

// GetMemberLevels 会员等级列表
// @Summary 会员等级列表
// @Description 会员等级列表
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{list=[]resp.MemberLevel}
// @Router /cashier/member/levels [get]
func (h *MemberHandler) GetMemberLevels(c *gin.Context) {
	helper.Success(c, gin.H{"list": h.memberSrv.GetLevels(c.GetUint64(jwt.CompanyUuid))})
}

// SearchMember 模糊搜索会员
// @Summary 模糊搜索会员
// @Description 模糊搜索会员
// @Tags 收银端.会员
// @Accept json
// @Produce json
// @Security JwtToken
// @param keyword query string false "关键字搜索：uuid\phone，前端处理前后空格"
// @Success 200 {object} dto.Response{list=[]resp.SearchMember}
// @Router /cashier/member/search [get]
func (h *MemberHandler) SearchMember(c *gin.Context) {
	helper.Success(c, gin.H{"list": h.memberSrv.SearchMember(c.GetUint64(jwt.CompanyUuid), c.Query("keyword"))})
}

func RegisterMemberHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	memberSrv := service.NewMemberSrv(dbm)

	wrapper := &MemberHandler{
		authSrv:   authSrv,
		memberSrv: memberSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/member/levels", wrapper.GetMemberLevels) // 获取会员等级列表
		privateApi.GET("/member/search", wrapper.SearchMember)    // 模糊搜索会员
	}
}

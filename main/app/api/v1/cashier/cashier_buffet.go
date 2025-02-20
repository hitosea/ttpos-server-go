package cashier

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

// BuffetHandler 自助餐处理程序
type BuffetHandler struct {
	Service service.IBuffetSrv // 主服务
}

// GetBuffetList 处理获取自助餐列表
// @Summary 获取自助餐列表
// @Description 获取自助餐列表
// @Tags 收银端.自助餐
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.BuffetListPaginationResp "自助餐列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/buffet/list [get]
func (h *BuffetHandler) GetBuffetList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 获取自助餐列表
	res, err := h.Service.GetBuffetList(companyUuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterBuffetHandlers 注册收银产品路由
func RegisterBuffetHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	bindRecordSrv := service.NewBindRecordSrv(settingSrv, dbm)
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, bindRecordSrv, staffShiftSrv, settingSrv)

	// 初始化处理器
	wrapper := BuffetHandler{
		Service: service.NewBuffetSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv))
	{
		privateApi.GET("/buffet/list", wrapper.GetBuffetList)
	}
}

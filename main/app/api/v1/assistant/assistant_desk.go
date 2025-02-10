package assistant

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req/cashier_req"
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	Service service.IDeskSrv // 桌台服务
}

// GetAssistantDeskRegionAndType 处理获取收银台的区域和类型
// @Summary 获取收银台的区域和类型
// @Description 获取收银台的区域和类型
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {object} cashier_resp.DeskRegionAndTypeListWithPaginationResp "收银台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	// 处理获取收银台的区域和类型的逻辑
	res, err := h.Service.GetDeskRegionAndTypeList(1)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetAssistantDeskList 处理获取收银台列表
// @Summary 获取收银台列表
// @Description 获取收银台列表
// @Tags 收银端
// @Accept json
// @Produce json
// @Success 200 {array} nil "收银台列表"
// @Failure 404 {object} nil "未找到"
// @Router /assistant/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	// 绑定请求参数
	req := cashier_req.DeskListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, dto.PageReqMessage)
		return
	}
	// 获取收银产品列表
	res, err := h.Service.GetDeskList(
		1,
		req,
	)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterProductHandlers 注册收银产品路由
func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager) {
	// 创建收银产品处理程序
	wrapper := DeskHandler{
		Service: service.NewDeskSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
		),
	}

	// 注册收银产品路由
	router.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)
	router.GET("/desk/list", wrapper.GetDeskList) // 获取收银产品列表
}

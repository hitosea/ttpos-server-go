package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
)

// MaterialHandler 物品处理程序
type MaterialHandler struct {
	materialSrv service.IMaterialSrv // 物品服务
}

// GetMaterialList 获取物品列表
// @Summary 获取物品列表
// @Description 获取物品列表
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "关键字"
// @Param category_uuid query int false "分类UUID"
// @Param status query int false "状态 0-全部 1-启用 2-停用"
// @Success 200 {object} material_resp.MaterialListWithPaginationResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/list [get]
func (h *MaterialHandler) GetMaterialList(c *gin.Context) {
	ctx := helper.GetContext(c)
	listReq := req.MaterialListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	listReq.CategoryUuids = listReq.GetCategoryUuids() // 获取有效分类UUID列表。 /api/v1/shop/material/list?category_uuids&keyword&page_no=1&page_size=1000&status 时，CategoryUuids会有一个0值，是无效的
	res, err := h.materialSrv.GetMaterialList(ctx, listReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetMaterialDetail 获取物品详情
// @Summary 获取物品详情
// @Description 获取物品详情
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string true "物品UUID"
// @Success 200 {object} material_resp.MaterialDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/detail [get]
func (h *MaterialHandler) GetMaterialDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.MaterialDetailReq{}

	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, dto.PageReqMessage)
		return
	}
	res, err := h.materialSrv.GetMaterialDetail(ctx, detailReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// AddMaterialCategory 创建物品类别
// @Summary 创建物品类别
// @Description 创建物品类别
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialCategoryAddReq true "物品类别添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/category/add [post]
func (h *MaterialHandler) AddMaterialCategory(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.MaterialCategoryAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	if err := addReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	err := h.materialSrv.AddMaterialCategory(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// GetMaterialCategoryList 获取物品类别列表
// @Summary 获取物品类别列表
// @Description 获取物品类别列表
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} material_resp.MaterialCategoryListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/category/list [get]
func (h *MaterialHandler) GetMaterialCategoryList(c *gin.Context) {
	ctx := helper.GetContext(c)
	listReq := req.MaterialCategoryListReq{}

	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	res, err := h.materialSrv.GetMaterialCategoryList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddMaterial 添加物品
// @Summary 添加物品
// @Description 添加物品
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialAddReq true "物品添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/add [post]
func (h *MaterialHandler) AddMaterial(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.MaterialAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	if err := addReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	err := h.materialSrv.AddMaterial(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// EditMaterial 编辑物品
// @Summary 编辑物品
// @Description 编辑物品
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialEditReq true "物品编辑请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/edit [post]
func (h *MaterialHandler) EditMaterial(c *gin.Context) {
	ctx := helper.GetContext(c)
	editReq := req.MaterialEditReq{}

	if err := c.ShouldBindJSON(&editReq); err != nil {
		helper.HandleValidationError(c, err, editReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.EditMaterial(ctx, editReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// UpdateMaterialStatusBatch 批量修改物品状态

// @Summary 批量修改物品状态
// @Description 批量修改物品状态
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialStatusReq true "物品状态修改请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/status/batch [post]
func (h *MaterialHandler) UpdateMaterialStatusBatch(c *gin.Context) {
	ctx := helper.GetContext(c)
	statusReq := req.MaterialStatusReq{}

	if err := c.ShouldBindJSON(&statusReq); err != nil {
		helper.HandleValidationError(c, err, statusReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.UpdateMaterialStatusBatch(ctx, statusReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// GetMaterialUnitList 获取物品的单位列表
// @Summary 获取物品的单位列表
// @Description 获取物品的单位列表
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string true "物品UUID"
// @Success 200 {object} material_resp.MaterialUnitListResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/unit/list [get]
func (h *MaterialHandler) GetMaterialUnitList(c *gin.Context) {
	ctx := helper.GetContext(c)
	unitListReq := req.MaterialUnitListReq{}

	if err := c.ShouldBindQuery(&unitListReq); err != nil {
		helper.HandleValidationError(c, err, unitListReq, dto.PageReqMessage)
		return
	}
	res, err := h.materialSrv.GetMaterialUnitList(ctx, unitListReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// AddProductBomCard 添加成本卡
// @Summary 添加成本卡
// @Description 添加成本卡
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductBomCardAddReq true "成本卡添加请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product_bom/card/add [post]
func (h *MaterialHandler) AddProductBomCard(c *gin.Context) {
	ctx := helper.GetContext(c)
	addReq := req.ProductBomCardAddReq{}

	if err := c.ShouldBindJSON(&addReq); err != nil {
		helper.HandleValidationError(c, err, addReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.AddProductBomCard(ctx, addReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "创建成功")
}

// GetProductBomCardDetail 规格商品成本卡详情
// @Summary 规格商品成本卡详情
// @Description 规格商品成本卡详情
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query string true "成本卡UUID"
// @Success 200 {object} material_resp.ProductBomCardDetailResp "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product_bom/card/detail [get]
func (h *MaterialHandler) GetProductBomCardDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	detailReq := req.ProductBomCardDetailReq{}

	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, dto.PageReqMessage)
		return
	}
	res, err := h.materialSrv.GetProductBomCardDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// UnlinkProductBomCard 解除成本卡关联
// @Summary 解除成本卡关联
// @Description 解除成本卡关联
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductBomCardUnlinkReq true "成本卡解除关联请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product_bom/card/unlink [post]
func (h *MaterialHandler) UnlinkProductBomCard(c *gin.Context) {
	ctx := helper.GetContext(c)
	unlinkReq := req.ProductBomCardUnlinkReq{}

	if err := c.ShouldBindJSON(&unlinkReq); err != nil {
		helper.HandleValidationError(c, err, unlinkReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.UnlinkProductBomCard(ctx, unlinkReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "解除成功")
}

// CopyProductBomCard 复制成本卡
// @Summary 复制成本卡
// @Description 复制成本卡
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductBomCardCopyReq true "成本卡复制请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product_bom/card/copy [post]
func (h *MaterialHandler) CopyProductBomCard(c *gin.Context) {
	ctx := helper.GetContext(c)
	copyReq := req.ProductBomCardCopyReq{}

	if err := c.ShouldBindJSON(&copyReq); err != nil {
		helper.HandleValidationError(c, err, copyReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.CopyProductBomCard(ctx, copyReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil, "关联成本卡成功")
}

// ImportProductBomCard 从菜品导入成本卡
// @Summary 从菜品导入成本卡
// @Description 从菜品导入成本卡
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.ProductBomCardImportReq true "成本卡导入请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/product_bom/card/import [post]
func (h *MaterialHandler) ImportProductBomCard(c *gin.Context) {
	ctx := helper.GetContext(c)
	importReq := req.ProductBomCardImportReq{}

	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, dto.PageReqMessage)
		return
	}
	if err := importReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeParamError, errors.WithMessage(err))
		return
	}
	err := h.materialSrv.ImportProductBomCard(ctx, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, nil)
}

// ImportMaterialList 获取导入物品列表
// @Summary 获取导入物品列表
// @Description 获取导入物品列表
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialImportListReq true "获取导入物品列表请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/import/list [get]
func (h *MaterialHandler) ImportMaterialList(c *gin.Context) {
	ctx := helper.GetContext(c)
	importReq := req.MaterialImportListReq{}

	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, dto.PageReqMessage)
		return
	}
	res, err := h.materialSrv.ImportMaterialList(ctx, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// ImportMaterial 导入物品
// @Summary 导入物品
// @Description 导入物品
// @Tags 商家端.物品管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.MaterialImportReq true "导入物品请求"
// @Success 200 {object} nil "成功"
// @Failure 400 {object} nil "错误请求"
// @Router /shop/material/import [post]
func (h *MaterialHandler) ImportMaterial(c *gin.Context) {
	ctx := helper.GetContext(c)
	importReq := req.MaterialImportReq{}
	if err := c.ShouldBindJSON(&importReq); err != nil {
		helper.HandleValidationError(c, err, importReq, dto.PageReqMessage)
		return
	}
	err := h.materialSrv.ImportMaterial(ctx, importReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{}, "导入成功")
}

func RegisterMaterialHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 创建物品处理程序
	wrapper := MaterialHandler{
		materialSrv: service.NewMaterialSrv(
			dbm,                    // 数据库管理器
			service.NewLocaleSrv(), // 多语言服务
			settingSrv,
		),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/material/list", wrapper.GetMaterialList)                    // 获取物品列表
		privateApi.GET("/material/detail", wrapper.GetMaterialDetail)                // 获取物品详情
		privateApi.POST("/material/category/add", wrapper.AddMaterialCategory)       // 创建物品类别
		privateApi.GET("/material/category/list", wrapper.GetMaterialCategoryList)   // 获取物品类别列表
		privateApi.POST("/material/add", wrapper.AddMaterial)                        // 添加物品
		privateApi.POST("/material/edit", wrapper.EditMaterial)                      // 编辑物品
		privateApi.POST("/material/status/batch", wrapper.UpdateMaterialStatusBatch) // 批量修改物品状态
		privateApi.GET("/material/unit/list", wrapper.GetMaterialUnitList)           // 获取物品的单位列表

		privateApi.POST("/product_bom/card/add", wrapper.AddProductBomCard)         // 添加成本卡
		privateApi.GET("/product_bom/card/detail", wrapper.GetProductBomCardDetail) // 规格商品成本卡详情
		privateApi.POST("/product_bom/card/unlink", wrapper.UnlinkProductBomCard)   // 解除成本卡关联
		privateApi.POST("/product_bom/card/copy", wrapper.CopyProductBomCard)       // 复制成本卡
		privateApi.POST("/product_bom/card/import", wrapper.ImportProductBomCard)   // 从菜品导入成本卡

		privateApi.GET("/material/import/list", wrapper.ImportMaterialList) // 导入物品列表
		privateApi.POST("/material/import", wrapper.ImportMaterial)         // 导入物品
	}
}

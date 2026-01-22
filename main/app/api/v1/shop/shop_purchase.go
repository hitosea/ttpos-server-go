package shop

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/purchase_order"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

// PurchaseHandler 采购控制器
type PurchaseHandler struct {
	authSrv                service.IAuthSrv
	purchaseOrderSrv       purchase_order.IPurchaseOrderSrv
	purchaseLimitSchemeSrv purchase_order.IPurchaseLimitSchemeSrv
	uploadFileSrv          service.IUploadFileSrv
	receiptFileSrv         service.IPurchaseReceiptFileSrv
}

// GetPurchaseOrderList 获取采购订单列表
// @Summary 获取采购订单列表
// @Description 分页获取采购订单列表
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderListReq true "采购订单列表请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderListResp} "成功"
// @Router /shop/purchase/order/list [get]
func (h *PurchaseHandler) GetPurchaseOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PurchaseOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetPurchaseOrderDetail 获取采购订单详情
// @Summary 获取采购订单详情
// @Description 根据UUID获取采购订单详情
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "采购订单UUID"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderDetailResp} "成功"
// @Router /shop/purchase/order/detail [get]
func (h *PurchaseHandler) GetPurchaseOrderDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.PurchaseOrderDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseOrderDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CreatePurchaseOrder 创建采购订单
// @Summary 创建采购订单
// @Description 创建新的采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderCreateReq true "创建采购订单请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/create [post]
func (h *PurchaseHandler) CreatePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PurchaseOrderCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.CreatePurchaseOrder(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// UpdatePurchaseOrder 更新采购订单
// @Summary 更新采购订单
// @Description 更新采购订单信息
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderUpdateReq true "更新采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/purchase/order/update [post]
func (h *PurchaseHandler) UpdatePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseOrderUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.purchaseOrderSrv.UpdatePurchaseOrder(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// UpdatePurchaseOrder 更新采购订单
// @Summary 更新采购订单物品单位
// @Description 更新采购订单物品单位
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderDetailReq true "更新采购订单物品单位请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseOrderDetailResp} "成功"
// @Router /shop/purchase/order/item/units/update [post]
func (h *PurchaseHandler) UpdatePurchaseOrderItemUnit(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseOrderDetailReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.UpdatePurchaseOrderItemUnit(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// DeletePurchaseOrder 删除采购订单
// @Summary 删除采购订单
// @Description 删除采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderDeleteReq true "删除采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/delete [delete]
func (h *PurchaseHandler) DeletePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.PurchaseOrderDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.purchaseOrderSrv.DeletePurchaseOrder(ctx, deleteReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// SubmitPurchaseOrder 提交采购订单
// @Summary 提交采购订单
// @Description 提交采购订单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderSubmitReq true "提交采购订单请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/submit [post]
func (h *PurchaseHandler) SubmitPurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var statusReq req.PurchaseOrderSubmitReq
	if err := c.ShouldBindJSON(&statusReq); err != nil {
		helper.HandleValidationError(c, err, statusReq, nil)
		return
	}

	err := h.purchaseOrderSrv.SubmitPurchaseOrder(ctx, statusReq)
	if err != nil {
		helper.ErrorAutoWithData(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{}, "提交成功")
}

// ApprovePurchaseOrder 审核采购订单
// @Summary 审核采购订单
// @Description 审核采购订单（通过或驳回）
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseOrderApproveReq true "审核采购订单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/order/approve [post]
func (h *PurchaseHandler) ApprovePurchaseOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var approveReq req.PurchaseOrderApproveReq
	if err := c.ShouldBindJSON(&approveReq); err != nil {
		helper.HandleValidationError(c, err, approveReq, nil)
		return
	}

	err := h.purchaseOrderSrv.ApprovePurchaseOrder(ctx, approveReq)
	if err != nil {
		helper.ErrorAutoWithData(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, gin.H{})
}

// CreatePurchaseReceipt 创建收货记录
// @Summary 创建收货记录
// @Description 创建采购收货记录
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptCreateReq true "创建收货记录请求参数"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptCreateResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/create [post]
func (h *PurchaseHandler) CreatePurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PurchaseReceiptCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	if ctx.Version(context.LT, "2.9.0") {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("您的软件版本过低，请升级后再试"))
		return
	} else if createReq.IsConfirm && len(createReq.FileUuids) == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("请上传相关附件后确定收货。"))
		return
	}
	resp, err := h.purchaseOrderSrv.CreatePurchaseReceiptOrder(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// UpdatePurchaseReceipt 更新收货记录
// @Summary 更新收货记录
// @Description 更新收货记录
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptOrderUpdateReq true "更新收货记录请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/update [post]
func (h *PurchaseHandler) UpdatePurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseReceiptOrderUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}
	if ctx.Version(context.LT, "2.9.0") {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("您的软件版本过低，请升级后再试"))
		return
	} else if updateReq.IsConfirm && len(updateReq.FileUuids) == 0 {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("请上传相关附件后确定收货。"))
		return
	}
	err := h.purchaseOrderSrv.UpdatePurchaseReceiptOrder(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// GetPurchaseReceiptList 获取收货记录列表
// @Summary 获取收货记录列表
// @Description 分页获取收货记录列表，支持按收货时间和创建时间筛选
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Param order_no query string false "订单号"
// @Param status_in query []int false "状态筛选"
// @Param receive_time_start query int64 false "收货时间开始（时间戳）"
// @Param receive_time_end query int64 false "收货时间结束（时间戳）"
// @Param create_time_start query int64 false "创建时间开始（时间戳）"
// @Param create_time_end query int64 false "创建时间结束（时间戳）"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptOrderListResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/list [get]
func (h *PurchaseHandler) GetPurchaseReceiptList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PurchaseReceiptOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseReceiptOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetPurchaseReceiptDetail 获取收货记录详情
// @Summary 获取收货记录详情
// @Description 根据UUID获取收货记录详情
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "收货记录UUID"
// @Success 200 {object} dto.Response{data=resp.PurchaseReceiptOrderDetailResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/detail [get]
func (h *PurchaseHandler) GetPurchaseReceiptDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.PurchaseReceiptOrderDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.purchaseOrderSrv.GetPurchaseReceiptOrderDetail(ctx, detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CancelPurchaseReceipt 取消收货单
// @Summary 取消收货单
// @Description 取消收货单
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseReceiptOrderCancelReq true "取消收货单请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/cancel [post]
func (h *PurchaseHandler) CancelPurchaseReceipt(c *gin.Context) {
	ctx := helper.GetContext(c)
	var cancelReq req.PurchaseReceiptOrderCancelReq
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		helper.HandleValidationError(c, err, cancelReq, nil)
		return
	}

	err := h.purchaseOrderSrv.CancelPurchaseReceiptOrder(ctx, cancelReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// UploadDocument 上传文档
// @Summary 上传文档
// @Description 上传文档附件（支持PDF、Word、Excel、图片等）
// @Tags 商家端.采购管理
// @Accept multipart/form-data
// @Produce json
// @Security JwtToken
// @Param file formData file true "文件"
// @Param group_id formData int false "分组ID"
// @Success 200 {object} dto.Response{data=resp.UploadFileResp} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/file/upload_document [post]
func (h *PurchaseHandler) UploadDocument(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "请选择要上传的文件"))
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err, "打开文件失败"))
		return
	}
	defer src.Close()

	// 获取分组ID
	groupId := uint64(0)
	if groupIdStr := c.PostForm("group_id"); groupIdStr != "" {
		if id, err := utils.StringToUint64(groupIdStr); err == nil {
			groupId = id
		}
	}

	// 上传文档
	resp, err := h.uploadFileSrv.UploadDocument(ctx, src, file.Filename, file.Size, groupId)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// DeleteReceiptFile 删除收货单附件
// @Summary 删除收货单附件
// @Description 删除收货单的指定附件（仅草稿状态）
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.DeleteReceiptFileReq true "删除附件请求参数"
// @Success 200 {object} dto.Response "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/receipt/file [delete]
func (h *PurchaseHandler) DeleteReceiptFile(c *gin.Context) {
	ctx := helper.GetContext(c)

	var deleteReq req.DeleteReceiptFileReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.receiptFileSrv.DeleteReceiptFile(ctx, deleteReq.FileUuid, deleteReq.ReceiptOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// GetLimitSchemeList 获取限购方案列表
// @Summary 获取限购方案列表
// @Description 分页获取限购方案列表
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param page_no query int true "页码"
// @Param page_size query int true "每页数量"
// @Param status query int false "状态：0=关闭，1=开启"
// @Param name query string false "方案名称（模糊搜索）"
// @Success 200 {object} dto.Response{data=resp.PurchaseLimitSchemeListResp} "成功"
// @Router /shop/purchase/limit/scheme/list [get]
func (h *PurchaseHandler) GetLimitSchemeList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.PurchaseLimitSchemeListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, nil)
		return
	}

	resp, err := h.purchaseLimitSchemeSrv.GetList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// GetLimitSchemeDetail 获取限购方案详情
// @Summary 获取限购方案详情
// @Description 根据UUID获取限购方案详情
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query int true "方案UUID"
// @Success 200 {object} dto.Response{data=resp.PurchaseLimitSchemeResp} "成功"
// @Router /shop/purchase/limit/scheme/detail [get]
func (h *PurchaseHandler) GetLimitSchemeDetail(c *gin.Context) {
	ctx := helper.GetContext(c)
	var detailReq req.PurchaseLimitSchemeDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.HandleValidationError(c, err, detailReq, nil)
		return
	}

	resp, err := h.purchaseLimitSchemeSrv.GetByUuid(ctx, detailReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, resp)
}

// CreateLimitScheme 创建限购方案
// @Summary 创建限购方案
// @Description 创建新的限购方案
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseLimitSchemeCreateReq true "创建限购方案请求参数"
// @Success 200 {object} dto.Response{data=map[string]interface{}} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/limit/scheme/create [post]
func (h *PurchaseHandler) CreateLimitScheme(c *gin.Context) {
	ctx := helper.GetContext(c)
	var createReq req.PurchaseLimitSchemeCreateReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		helper.HandleValidationError(c, err, createReq, nil)
		return
	}

	uuid, err := h.purchaseLimitSchemeSrv.Create(ctx, createReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{
		"uuid": uuid,
	})
}

// UpdateLimitScheme 更新限购方案
// @Summary 更新限购方案
// @Description 更新限购方案信息
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseLimitSchemeUpdateReq true "更新限购方案请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Router /shop/purchase/limit/scheme/update [post]
func (h *PurchaseHandler) UpdateLimitScheme(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.PurchaseLimitSchemeUpdateReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.HandleValidationError(c, err, updateReq, nil)
		return
	}

	err := h.purchaseLimitSchemeSrv.Update(ctx, updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

// DeleteLimitScheme 删除限购方案
// @Summary 删除限购方案
// @Description 软删除限购方案及其关联配置
// @Tags 商家端.采购管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.PurchaseLimitSchemeDeleteReq true "删除限购方案请求参数"
// @Success 200 {object} dto.Response{} "成功"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Router /shop/purchase/limit/scheme/delete [delete]
func (h *PurchaseHandler) DeleteLimitScheme(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReq req.PurchaseLimitSchemeDeleteReq
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		helper.HandleValidationError(c, err, deleteReq, nil)
		return
	}

	err := h.purchaseLimitSchemeSrv.Delete(ctx, deleteReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	helper.Success(c, gin.H{})
}

func RegisterPurchaseHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)

	// 采购服务
	purchaseOrderSrv := purchase_order.NewPurchaseOrderSrvImpl(dbm, settingSrv)
	purchaseLimitSchemeSrv := purchase_order.NewPurchaseLimitSchemeSrv(dbm)
	uploadFileSrv := service.NewUploadFileSrv(dbm)
	receiptFileSrv := service.NewPurchaseReceiptFileSrv(dbm)

	wrapper := &PurchaseHandler{
		authSrv:                authSrv,
		purchaseOrderSrv:       purchaseOrderSrv,
		purchaseLimitSchemeSrv: purchaseLimitSchemeSrv,
		uploadFileSrv:          uploadFileSrv,
		receiptFileSrv:         receiptFileSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.MinVersionCheck("2.9.0"), middleware.Auth(authSrv, dbm))
	{
		// 采购订单管理
		privateApi.GET("/purchase/order/list", wrapper.GetPurchaseOrderList)
		privateApi.GET("/purchase/order/detail", wrapper.GetPurchaseOrderDetail)
		privateApi.POST("/purchase/order/create", wrapper.CreatePurchaseOrder)
		privateApi.POST("/purchase/order/update", wrapper.UpdatePurchaseOrder)
		privateApi.DELETE("/purchase/order/delete", wrapper.DeletePurchaseOrder)
		privateApi.POST("/purchase/order/approve", wrapper.ApprovePurchaseOrder)
		privateApi.POST("/purchase/order/submit", wrapper.SubmitPurchaseOrder)

		// 收货管理
		privateApi.GET("/purchase/receipt/list", wrapper.GetPurchaseReceiptList)
		privateApi.POST("/purchase/receipt/create", wrapper.CreatePurchaseReceipt)
		privateApi.POST("/purchase/receipt/update", wrapper.UpdatePurchaseReceipt)
		privateApi.POST("/purchase/order/item/units/update", wrapper.UpdatePurchaseOrderItemUnit)
		privateApi.GET("/purchase/receipt/detail", wrapper.GetPurchaseReceiptDetail)
		privateApi.DELETE("/purchase/receipt/cancel", wrapper.CancelPurchaseReceipt)

		// 收货单附件管理
		privateApi.POST("/file/upload_document", wrapper.UploadDocument)
		privateApi.DELETE("/purchase/receipt/file", wrapper.DeleteReceiptFile)

		// 限购方案管理
		privateApi.GET("/purchase/limit/scheme/list", wrapper.GetLimitSchemeList)
		privateApi.GET("/purchase/limit/scheme/detail", wrapper.GetLimitSchemeDetail)
		privateApi.POST("/purchase/limit/scheme/create", wrapper.CreateLimitScheme)
		privateApi.POST("/purchase/limit/scheme/update", wrapper.UpdateLimitScheme)
		privateApi.DELETE("/purchase/limit/scheme/delete", wrapper.DeleteLimitScheme)
	}
}

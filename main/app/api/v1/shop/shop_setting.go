package shop

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/dto/resp/setting"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/message"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SettingHandler 设置控制器
type SettingHandler struct {
	dbm           *database.DBManager
	syncSrv       service.ISyncSrv
	settingSrv    settingSrv.ISrv
	otherSrv      service.IOtherSrv
	uploadFileSrv service.IUploadFileSrv
	dataManageSrv service.IDataManageSrv
	companySrv    service.ICompanySrv
}

// SaveStoreSetting 保存门店设置
// @Summary 保存门店设置
// @Description 保存门店设置
// @Tags 商家端.门店设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateStoreSetting true "更新门店设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/store [post]
func (h *SettingHandler) SaveStoreSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateStoreSetting req.UpdateStoreSetting
	if err := c.ShouldBindJSON(&updateStoreSetting); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.settingSrv.EditStoreSetting(ctx, updateStoreSetting)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 处理营业状态切换
	if updateStoreSetting.BusinessStatus != nil {
		if err := h.companySrv.UpdateBusinessStatus(ctx, req.UpdateBusinessStatusReq{
			Uuid:           ctx.GetCompanyUuid(),
			BusinessStatus: *updateStoreSetting.BusinessStatus,
		}); err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, err)
			return
		}
	}
	helper.Success(c, "保存成功")
}

// SaveBusinessSetting 保存业务设置
// @Summary 保存业务设置
// @Description 保存业务设置
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdateBusinessSetting true "更新业务设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/business [post]
func (h *SettingHandler) SaveBusinessSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateBusinessSetting req.UpdateBusinessSetting
	if err := c.ShouldBindJSON(&updateBusinessSetting); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.settingSrv.EditBusinessSetting(ctx, updateBusinessSetting)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功", "保存成功")
}

// GetBusinessSetting 获取业务设置
// @Summary 获取业务设置
// @Description 获取业务设置
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} setting.ShopBusiness
// @Router /shop/setting/business [get]
func (h *SettingHandler) GetBusinessSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	businessSetting, err := h.settingSrv.GetShopBusinessSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, businessSetting)
}

// GetFreeReason 获取免单原因
// @Summary 获取免单原因
// @Description 获取免单原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.GiftOrFreeOrderReasonResp
// @Router /shop/setting/free_reason [get]
func (h *SettingHandler) GetFreeReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	freeReason, err := h.otherSrv.GetGiftOrFreeReasonList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, resp.GiftOrFreeOrderReasonResp{
		List: freeReason.List,
	})
}

// AddFreeReason 新增免单原因
// @Summary 新增免单原因
// @Description 新增免单原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddFreeOrGiftReasonReq true "新增免单原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/free_reason/add [post]
func (h *SettingHandler) AddFreeReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addFreeReason req.AddFreeOrGiftReasonReq
	if err := c.ShouldBindJSON(&addFreeReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.AddFreeOrGiftReason(ctx, addFreeReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// EditFreeReason 编辑免单原因
// @Summary 编辑免单原因
// @Description 编辑免单原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.EditFreeOrGiftReasonReq true "编辑免单原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/free_reason/edit [post]
func (h *SettingHandler) EditFreeReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var editFreeReason req.EditFreeOrGiftReasonReq
	if err := c.ShouldBindJSON(&editFreeReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.EditFreeOrGiftReason(ctx, editFreeReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// DeleteFreeReason 删除免单原因
// @Summary 删除免单原因
// @Description 删除免单原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeleteFreeOrGiftReasonReq true "删除免单原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/free_reason [delete]
func (h *SettingHandler) DeleteFreeReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteFreeReason req.DeleteFreeOrGiftReasonReq
	if err := c.ShouldBindJSON(&deleteFreeReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.DeleteFreeOrGiftReason(ctx, deleteFreeReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "删除成功")
}

// GetReturnFoodReason 获取退菜原因
// @Summary 获取退菜原因
// @Description 获取退菜原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.ReturnFoodReasonResp
// @Router /shop/setting/return_food_reason [get]
func (h *SettingHandler) GetReturnFoodReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	returnFoodReason, err := h.otherSrv.GetReturnFoodReasonList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, resp.ReturnFoodReasonResp{
		List: returnFoodReason.List,
	})
}

// AddReturnFoodReason 新增退菜原因
// @Summary 新增退菜原因
// @Description 新增退菜原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.AddReturnFoodReasonReq true "新增退菜原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/return_food_reason/add [post]
func (h *SettingHandler) AddReturnFoodReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var addReturnFoodReason req.AddReturnFoodReasonReq
	if err := c.ShouldBindJSON(&addReturnFoodReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.AddReturnFoodReason(ctx, addReturnFoodReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// EditReturnFoodReason 编辑退菜原因
// @Summary 编辑退菜原因
// @Description 编辑退菜原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.EditReturnFoodReasonReq true "编辑退菜原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/return_food_reason/edit [post]
func (h *SettingHandler) EditReturnFoodReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var editReturnFoodReason req.EditReturnFoodReasonReq
	if err := c.ShouldBindJSON(&editReturnFoodReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.EditReturnFoodReason(ctx, editReturnFoodReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// DeleteReturnFoodReason 删除退菜原因
// @Summary 删除退菜原因
// @Description 删除退菜原因
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeleteReturnFoodReasonReq true "删除退菜原因"
// @Success 200 {object} dto.Response
// @Router /shop/setting/return_food_reason [delete]
func (h *SettingHandler) DeleteReturnFoodReason(c *gin.Context) {
	ctx := helper.GetContext(c)
	var deleteReturnFoodReason req.DeleteReturnFoodReasonReq
	if err := c.ShouldBindJSON(&deleteReturnFoodReason); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.otherSrv.DeleteReturnFoodReason(ctx, deleteReturnFoodReason)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "删除成功")
}

// GetOrderRemark 获取整单备注
// @Summary 获取整单备注
// @Description 获取整单备注列表
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Success 200 {object} resp.OrderRemarkResp
// @Security JwtToken
// @Router /shop/setting/order_remark [get]
func (h *SettingHandler) GetOrderRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	orderRemark, err := h.otherSrv.GetOrderRemarkList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, orderRemark)
}

// AddOrderRemark 新增整单备注
// @Summary 新增整单备注
// @Description 新增整单备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.AddOrderRemarkReq true "新增整单备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_remark/add [post]
func (h *SettingHandler) AddOrderRemark(c *gin.Context) {
	var addOrderRemark req.AddOrderRemarkReq
	if err := c.ShouldBindJSON(&addOrderRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.AddOrderRemark(ctx, addOrderRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "新增成功")
}

// EditOrderRemark 编辑整单备注
// @Summary 编辑整单备注
// @Description 编辑整单备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.EditOrderRemarkReq true "编辑整单备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_remark/edit [post]
func (h *SettingHandler) EditOrderRemark(c *gin.Context) {
	var editOrderRemark req.EditOrderRemarkReq
	if err := c.ShouldBindJSON(&editOrderRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.EditOrderRemark(ctx, editOrderRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "编辑成功")
}

// DeleteOrderRemark 删除整单备注
// @Summary 删除整单备注
// @Description 删除整单备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.DeleteOrderRemarkReq true "删除整单备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_remark [delete]
func (h *SettingHandler) DeleteOrderRemark(c *gin.Context) {
	var deleteOrderRemark req.DeleteOrderRemarkReq
	if err := c.ShouldBindJSON(&deleteOrderRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.DeleteOrderRemark(ctx, deleteOrderRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "删除成功")
}

// GetOrderItemRemark 获取单品备注
// @Summary 获取单品备注
// @Description 获取单品备注列表
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Success 200 {object} resp.OrderItemRemarkResp
// @Security JwtToken
// @Router /shop/setting/order_item_remark [get]
func (h *SettingHandler) GetOrderItemRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	orderItemRemark, err := h.otherSrv.GetOrderItemRemarkList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, orderItemRemark)
}

// AddOrderItemRemark 新增单品备注
// @Summary 新增单品备注
// @Description 新增单品备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.AddOrderItemRemarkReq true "新增单品备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_item_remark/add [post]
func (h *SettingHandler) AddOrderItemRemark(c *gin.Context) {
	var addOrderItemRemark req.AddOrderItemRemarkReq
	if err := c.ShouldBindJSON(&addOrderItemRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.AddOrderItemRemark(ctx, addOrderItemRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "新增成功")
}

// EditOrderItemRemark 编辑单品备注
// @Summary 编辑单品备注
// @Description 编辑单品备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.EditOrderItemRemarkReq true "编辑单品备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_item_remark/edit [post]
func (h *SettingHandler) EditOrderItemRemark(c *gin.Context) {
	var editOrderItemRemark req.EditOrderItemRemarkReq
	if err := c.ShouldBindJSON(&editOrderItemRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.EditOrderItemRemark(ctx, editOrderItemRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "编辑成功")
}

// DeleteOrderItemRemark 删除单品备注
// @Summary 删除单品备注
// @Description 删除单品备注
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param data body req.DeleteOrderItemRemarkReq true "删除单品备注"
// @Success 200 {object} dto.Response
// @Security JwtToken
// @Router /shop/setting/order_item_remark [delete]
func (h *SettingHandler) DeleteOrderItemRemark(c *gin.Context) {
	var deleteOrderItemRemark req.DeleteOrderItemRemarkReq
	if err := c.ShouldBindJSON(&deleteOrderItemRemark); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	ctx := helper.GetContext(c)
	err := h.otherSrv.DeleteOrderItemRemark(ctx, deleteOrderItemRemark)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "删除成功")
}

// GetMenuQrcode 获取电子菜单二维码
// @Summary 获取电子菜单二维码
// @Description 获取电子菜单二维码
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /shop/setting/menu_qrcode [get]
func (h *SettingHandler) GetMenuQrcode(c *gin.Context) {
	ctx := helper.GetContext(c)
	menuQrcode, err := h.settingSrv.GetMenuQrcode(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, gin.H{
		"url": menuQrcode,
	})
}

// GetMemberQrcode 获取会员端二维码
// @Summary 获取会员端二维码
// @Description 获取会员端二维码
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /shop/setting/member_qrcode [get]
func (h *SettingHandler) GetMemberQrcode(c *gin.Context) {
	ctx := helper.GetContext(c)
	helper.Success(c, gin.H{
		"url": viper.GetString("MEMBER_BASE_URL") + "/launch/" + strconv.FormatUint(ctx.GetCompanyUuid(), 10),
	})
}

// UploadLogo 上传logo
// @Summary 上传logo
// @Description 上传logo
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param file formData file true "上传logo"
// @Success 200 {object} dto.Response{data=resp.UploadFileResp}
// @Success 200 {object} dto.Response
// @Router /shop/setting/upload_logo [post]
func (h *SettingHandler) UploadLogo(c *gin.Context) {
	ctx := helper.GetContext(c)
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	fileReader, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	uploadFileResp, err := h.uploadFileSrv.UploadImage(ctx, fileReader, file.Filename, file.Size, 0, "logoUrl")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, uploadFileResp)
}

// Sync 获取总部最新数据
// @Summary 获取总部最新数据
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param request body req.SyncReq false "同步请求参数"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SyncResp}
// @Router /shop/setting/sync [post]
func (h *SettingHandler) Sync(c *gin.Context) {
	result, err := h.syncSrv.Sync(helper.GetContext(c), req.SyncReq{})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// GetSyncTaskList 获取同步任务列表
// @Summary 获取同步任务列表
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param page_no query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param status query int false "同步状态: 0-进行中, 1-已完成, 2-失败"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SyncTaskListPaginationResp}
// @Router /shop/setting/sync_task/list [get]
func (h *SettingHandler) GetSyncTaskList(c *gin.Context) {
	var listReq req.SyncTaskListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	result, err := h.syncSrv.GetTaskList(helper.GetContext(c), listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// GetSyncTaskDetail 获取同步任务详情
// @Summary 获取同步任务详情
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Param task_uuid query int true "任务UUID"
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SyncTaskDetailResp}
// @Router /shop/setting/sync_task/detail [get]
func (h *SettingHandler) GetSyncTaskDetail(c *gin.Context) {
	var detailReq req.SyncTaskDetailReq
	if err := c.ShouldBindQuery(&detailReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	result, err := h.syncSrv.GetTaskDetail(helper.GetContext(c), detailReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// GetHeadquartersDataList 获取总部可同步数据列表
// @Summary 获取总部可同步数据列表
// @Description 获取总部可同步数据列表（按种类分组，返回所有16种数据类型）
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.HeadquartersDataListResp}
// @Router /shop/setting/headquarters_data_list [get]
func (h *SettingHandler) GetHeadquartersDataList(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 不需要传递任何参数，查询所有类型
	result, err := h.syncSrv.GetHeadquartersDataList(ctx, req.GetHeadquartersDataListReq{})
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, result)
}

// GranularSync 颗粒化同步数据
// @Summary 颗粒化同步数据
// @Description 颗粒化同步数据（接收勾选的uuid列表，删除未勾选的，同步勾选的）
// @Tags 商家端.业务设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.GranularSyncReq true "请求参数"
// @Success 200 {object} dto.Response{data=resp.GranularSyncResp}
// @Router /shop/setting/granular_sync [post]
func (h *SettingHandler) GranularSync(c *gin.Context) {
	ctx := helper.GetContext(c)

	var syncReq req.GranularSyncReq
	if err := c.ShouldBindJSON(&syncReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	result, err := h.syncSrv.GranularSync(ctx, syncReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, result)
}

// GetPaymentMethodList 获取支付方式列表
// @Summary 获取支付方式列表
// @Description 获取支付方式列表
// @Tags 商家端.支付管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.PaymentMethodListResp}
// @Router /shop/setting/payment_method/list [get]
func (h *SettingHandler) GetPaymentMethodList(c *gin.Context) {
	ctx := helper.GetContext(c)
	paymentRepo := service.NewPaymentRepo(ctx, h.dbm)
	lianLianPayAvailable := true
	err := paymentRepo.ValidateConfigError(ctx.GetCompanyUuid())
	if err != nil {
		lianLianPayAvailable = false
	}
	result := h.settingSrv.GetPaymentMethodList(ctx, lianLianPayAvailable)
	helper.Success(c, result)
}

// GetCashierSetting 获取收银机设置
// @Summary 获取收银机设置
// @Description 获取收银机副屏设置，仅返回副屏相关配置（轮播内容、点餐时轮播内容、轮播间隔、展示模式）
// @Tags 商家端.收银机设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.CashierSecondaryScreenResp}
// @Router /shop/setting/cashier [get]
func (h *SettingHandler) GetCashierSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	cashierSetting, err := h.settingSrv.GetCashierSetting(ctx, nil)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 只返回副屏相关的字段
	resp := setting.CashierSecondaryScreenResp{
		Carousel:                cashierSetting.Carousel,
		NoOrderCarouselInterval: cashierSetting.NoOrderCarouselInterval,
		OrderDisplayMode:        cashierSetting.OrderDisplayMode,
		OrderCarousel:           cashierSetting.OrderCarousel,
		OrderCarouselInterval:   cashierSetting.OrderCarouselInterval,
	}
	helper.Success(c, resp)
}

// SaveCashierSetting 保存收银机设置
// @Summary 保存收银机设置
// @Description 保存收银机设置，包括副屏相关配置
// @Tags 商家端.收银机设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveCashierSettingReq true "保存收银机设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/cashier [post]
func (h *SettingHandler) SaveCashierSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var cashierSettingReq req.SaveCashierSettingReq
	if err := c.ShouldBindJSON(&cashierSettingReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 参数验证
	if err := cashierSettingReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 调用 Service 层保存
	err := h.settingSrv.EditCashierSetting(ctx, cashierSettingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// UpdatePrintSetting 更新打印设置
// @Summary 更新打印设置
// @Description 更新打印设置，包括自定义打印联数配置
// @Tags 商家端.打印设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.UpdatePrintSettingReq true "更新打印设置"
// @Success 200 {object} dto.Response{data=object{enable_custom_copies=string,checkout_slip_copies=int}}
// @Router /shop/setting/print_setting/update [post]
func (h *SettingHandler) UpdatePrintSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.UpdatePrintSettingReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 调用 Service 层更新
	err := h.settingSrv.UpdatePrintSetting(ctx, &updateReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "保存成功")
}

// GetPrintSetting 获取打印设置
// @Summary 获取打印设置
// @Description 获取打印设置，仅返回自定义打印联数相关配置
// @Tags 商家端.打印设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.PrintSettingResp}
// @Router /shop/setting/print_setting/get [get]
func (h *SettingHandler) GetPrintSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	printerSetting, err := h.settingSrv.GetPrinterSetting(ctx, nil)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 只返回自定义打印联数相关字段
	resp := setting.PrintSettingResp{
		EnableCustomCopies: printerSetting.EnableCustomCopies,
		CheckoutSlipCopies: printerSetting.CheckoutSlipCopies,
	}
	helper.Success(c, resp)
}

// UploadCashierCarousel 上传收银机轮播内容（图片/视频）
// @Summary 上传收银机轮播内容
// @Description 上传收银机副屏轮播内容，支持图片（JPG、JPEG、PNG、WEBP，<15MB）和视频（MP4，<30MB）
// @Tags 商家端.收银机设置
// @Accept multipart/form-data
// @Produce json
// @Security JwtToken
// @Param file formData file true "文件"
// @Param file_type formData string false "文件类型：image 或 video，不传则自动识别"
// @Param group_id formData int false "分组ID"
// @Success 200 {object} dto.Response{data=resp.UploadFileResp}
// @Router /shop/setting/cashier/carousel/upload [post]
func (h *SettingHandler) UploadCashierCarousel(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 打开文件
	fileReader, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	defer fileReader.Close()

	// 获取文件类型参数（可选，如果不传则根据文件扩展名自动识别）
	fileTypeParam := c.PostForm("file_type")

	// 获取文件扩展名
	fileName := file.Filename
	extension := ""
	if len(fileName) > 0 {
		dotIndex := -1
		for i := len(fileName) - 1; i >= 0; i-- {
			if fileName[i] == '.' {
				dotIndex = i
				break
			}
		}
		if dotIndex >= 0 && dotIndex < len(fileName)-1 {
			extension = fileName[dotIndex+1:]
		}
	}
	extension = strings.ToLower(extension)

	// 获取分组ID
	groupId := uint64(0)
	if groupIdStr := c.PostForm("group_id"); groupIdStr != "" {
		if id, err := strconv.ParseUint(groupIdStr, 10, 64); err == nil {
			groupId = id
		}
	}

	var uploadFileResp *resp.UploadFileResp

	// 根据文件类型或扩展名判断是图片还是视频
	if fileTypeParam == "image" || (fileTypeParam == "" && (extension == "jpg" || extension == "jpeg" || extension == "png" || extension == "webp")) {
		// 图片上传：JPG、JPEG、PNG、WEBP，<15MB
		allowedExts := []string{"jpg", "jpeg", "png", "webp"}
		isAllowed := false
		for _, ext := range allowedExts {
			if extension == ext {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("图片仅支持JPG、JPEG、PNG、WEBP格式"))
			return
		}

		// 检查文件大小：15MB
		maxSizeBytes := int64(15 * 1024 * 1024)
		if file.Size > maxSizeBytes {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("图片文件大小不能超过15MB"))
			return
		}

		uploadFileResp, err = h.uploadFileSrv.UploadImage(ctx, fileReader, fileName, file.Size, groupId, "shop")
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, err)
			return
		}
	} else if fileTypeParam == "video" || (fileTypeParam == "" && extension == "mp4") {
		// 视频上传：MP4，<30MB
		if extension != "mp4" {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("视频仅支持MP4格式"))
			return
		}

		// 检查文件大小：30MB
		maxSizeBytes := int64(30 * 1024 * 1024)
		if file.Size > maxSizeBytes {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("视频文件大小不能超过30MB"))
			return
		}

		uploadFileResp, err = h.uploadFileSrv.UploadVideo(ctx, fileReader, fileName, file.Size, groupId, 30)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, err)
			return
		}
	} else {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("不支持的文件格式，图片支持JPG、JPEG、PNG、WEBP，视频支持MP4"))
		return
	}

	helper.Success(c, uploadFileResp)
}

// GetDataManage 获取数据管理
// @Summary 获取数据管理
// @Description 获取数据管理
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.GetDataManageResp}
// @Router /shop/setting/data_manage [get]
func (h *SettingHandler) GetDataManage(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.dataManageSrv.GetDataManage(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// SetDataManage 设置数据管理
// @Summary 设置数据管理
// @Description 设置数据管理
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SetDataManageReq true "设置数据管理"
// @Success 200 {object} dto.Response
// @Router /shop/setting/data_manage/set [post]
func (h *SettingHandler) SetDataManage(c *gin.Context) {
	ctx := helper.GetContext(c)
	var setDataManageReq req.SetDataManageReq
	if err := c.ShouldBindJSON(&setDataManageReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.dataManageSrv.SetDataManage(ctx, setDataManageReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "设置成功")
}

// SaveDataManageStatus 保存数据管理开关状态
// @Summary 保存数据管理开关状态
// @Description 仅保存数据管理开关状态，不影响操作人员和订单选择
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveDataManageStatusReq true "保存数据管理状态"
// @Success 200 {object} dto.Response
// @Router /shop/setting/data_manage/save_status [post]
func (h *SettingHandler) SaveDataManageStatus(c *gin.Context) {
	ctx := helper.GetContext(c)
	var saveReq req.SaveDataManageStatusReq
	if err := c.ShouldBindJSON(&saveReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.dataManageSrv.SaveStatus(ctx, saveReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// SetDataManageStaff 设置数据管理操作人员
// @Summary 设置数据管理操作人员
// @Description 设置操作人员，立即生效
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SetDataManageStaffReq true "设置操作人员"
// @Success 200 {object} dto.Response
// @Router /shop/setting/data_manage/set_staff [post]
func (h *SettingHandler) SetDataManageStaff(c *gin.Context) {
	ctx := helper.GetContext(c)
	var staffReq req.SetDataManageStaffReq
	if err := c.ShouldBindJSON(&staffReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.dataManageSrv.SetStaff(ctx, staffReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "设置成功")
}

// GetDataManageOrderList 获取已选订单列表
// @Summary 获取已选订单列表
// @Description 获取数据管理中已选的订单列表，支持分页和搜索
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param page_no query int true "页码"
// @param page_size query int true "页面大小"
// @param order_no query string false "订单编号搜索"
// @param date_type query int false "时间类型,-1=全部、0=今天、1=昨天、2=本周"
// @param query_start_date query string false "查询开始日期"
// @param query_end_date query string false "查询结束日期"
// @param bill_type query int false "订单类型,-1=全部、0=餐单、1=外卖"
// @Success 200 {object} dto.Response{data=setting.DataManageOrderListResp}
// @Router /shop/setting/data_manage/order_list [get]
func (h *SettingHandler) GetDataManageOrderList(c *gin.Context) {
	ctx := helper.GetContext(c)
	var listReq req.GetDataManageOrderListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	result, err := h.dataManageSrv.GetOrderList(ctx, listReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// RestoreDataManageOrder 恢复（移除）单条已选订单
// @Summary 恢复单条已选订单
// @Description 从数据管理中移除单条已选订单
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.RestoreDataManageOrderReq true "恢复订单"
// @Success 200 {object} dto.Response
// @Router /shop/setting/data_manage/order_restore [post]
func (h *SettingHandler) RestoreDataManageOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var restoreReq req.RestoreDataManageOrderReq
	if err := c.ShouldBindJSON(&restoreReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.dataManageSrv.RestoreOrder(ctx, restoreReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "操作成功")
}

// GetDataManageOrderSelect 获取可选订单列表
// @Summary 获取可选订单列表
// @Description 获取可供选择的订单列表，支持筛选和分页，标记已选状态
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param page_no query int true "页码"
// @param page_size query int true "页面大小"
// @param order_no query string false "订单编号搜索"
// @param date_type query int false "时间类型,-1=全部、0=今天、1=昨天、2=本周"
// @param query_start_date query string false "查询开始日期"
// @param query_end_date query string false "查询结束日期"
// @param bill_type query int false "订单类型,-1=全部、0=餐单、1=外卖"
// @Success 200 {object} dto.Response{data=setting.DataManageOrderSelectResp}
// @Router /shop/setting/data_manage/order_select [get]
func (h *SettingHandler) GetDataManageOrderSelect(c *gin.Context) {
	ctx := helper.GetContext(c)
	var selectReq req.GetDataManageOrderSelectReq
	if err := c.ShouldBindQuery(&selectReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	result, err := h.dataManageSrv.GetOrderSelect(ctx, selectReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// SubmitDataManageOrder 提交订单选择
// @Summary 提交订单选择
// @Description 提交当前筛选范围内的订单选择结果，与筛选范围外的已选订单合并
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SubmitDataManageOrderReq true "提交订单选择"
// @Success 200 {object} dto.Response
// @Router /shop/setting/data_manage/order_submit [post]
func (h *SettingHandler) SubmitDataManageOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	var submitReq req.SubmitDataManageOrderReq
	if err := c.ShouldBindJSON(&submitReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.dataManageSrv.SubmitOrder(ctx, submitReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "操作成功")
}

// GetDataManageOrderSelectStats 获取可选订单统计预览
// @Summary 获取可选订单统计预览
// @Description 预览提交后的选中数量和实付金额，不持久化
// @Tags 商家端.数据管理
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.GetDataManageOrderSelectStatsReq true "统计预览参数"
// @Success 200 {object} dto.Response{data=setting.DataManageOrderSelectStatsResp}
// @Router /shop/setting/data_manage/order_select_stats [post]
func (h *SettingHandler) GetDataManageOrderSelectStats(c *gin.Context) {
	ctx := helper.GetContext(c)
	var statsReq req.GetDataManageOrderSelectStatsReq
	if err := c.ShouldBindJSON(&statsReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	resp, err := h.dataManageSrv.GetOrderSelectStats(ctx, statsReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, resp)
}

// GetKioskSetting 获取自助点餐机设置
// @Summary 获取自助点餐机设置
// @Description 获取自助点餐机设置
// @Tags 商家端.自助点餐机设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.KioskResp}
// @Router /shop/setting/kiosk [get]
func (h *SettingHandler) GetKioskSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	kioskSetting, err := h.settingSrv.GetKioskSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 只返回响应字段，不返回敏感信息
	resp := setting.KioskResp{
		AdvancedPassword:  kioskSetting.AdvancedPassword,
		CallWaiterEnabled: kioskSetting.CallWaiterEnabled,
		LanguageList:      kioskSetting.LanguageList,
		Language:          kioskSetting.Language,
		DefaultLanguage:   kioskSetting.DefaultLanguage,
		Carousel:          kioskSetting.Carousel,
	}
	helper.Success(c, resp)
}

// SaveKioskSetting 保存自助点餐机设置
// @Summary 保存自助点餐机设置
// @Description 保存自助点餐机设置
// @Tags 商家端.自助点餐机设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveKioskSettingReq true "保存自助点餐机设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/kiosk [post]
func (h *SettingHandler) SaveKioskSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var kioskSettingReq req.SaveKioskSettingReq
	if err := c.ShouldBindJSON(&kioskSettingReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 参数验证
	if err := kioskSettingReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 调用 Service 层保存
	err := h.settingSrv.EditKioskSetting(ctx, kioskSettingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// GetStoreScanOrderSetting 获取门店点餐配置
// @Summary 获取门店点餐配置
// @Description 获取门店点餐配置，包含业务开关、外送服务、到店自取等
// @Tags 商家端.门店点餐设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.StoreScanOrderSettingResp}
// @Router /shop/setting/store_scan_order [get]
func (h *SettingHandler) GetStoreScanOrderSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.settingSrv.GetStoreScanOrderSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// SaveStoreScanOrderSetting 保存门店点餐配置
// @Summary 保存门店点餐配置
// @Description 保存门店点餐配置，包含业务开关、外送服务、到店自取等
// @Tags 商家端.门店点餐设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveStoreScanOrderSettingReq true "保存门店点餐配置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/store_scan_order [post]
func (h *SettingHandler) SaveStoreScanOrderSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var settingReq req.SaveStoreScanOrderSettingReq
	if err := c.ShouldBindJSON(&settingReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	err := h.settingSrv.SaveStoreScanOrderSetting(ctx, settingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// GetTemplateStyleSetting 获取模板样式配置
// @Summary 获取模板样式配置
// @Description 获取模板样式配置，返回当前模板样式值（1/2/3）
// @Tags 商家端.模板样式设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.TemplateStyleSettingResp}
// @Router /shop/setting/template_style [get]
func (h *SettingHandler) GetTemplateStyleSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	result, err := h.settingSrv.GetTemplateStyleSetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, result)
}

// SaveTemplateStyleSetting 保存模板样式配置
// @Summary 保存模板样式配置
// @Description 保存模板样式配置，模板值为 1/2/3
// @Tags 商家端.模板样式设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveTemplateStyleSettingReq true "保存模板样式配置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/template_style [post]
func (h *SettingHandler) SaveTemplateStyleSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var settingReq req.SaveTemplateStyleSettingReq
	if err := c.ShouldBindJSON(&settingReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	err := h.settingSrv.SaveTemplateStyleSetting(ctx, settingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "保存成功")
}

// GetKitchenSetting 获取厨显设置
// @Summary 获取厨显设置
// @Description 获取厨显设置
// @Tags 商家端.厨显设置
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=setting.KitchenResp}
// @Router /shop/setting/kitchen [get]
func (h *SettingHandler) GetKitchenSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	companySetting, err := h.settingSrv.GetCompanySetting(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	languageList, err := h.settingSrv.GetStoreLanguageList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	kitchenSetting, err := h.settingSrv.GetKitchenSetting(ctx, companySetting, languageList)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, kitchenSetting.KitchenResp)
}

// SaveKitchenSetting 保存厨显设置
// @Summary 保存厨显设置
// @Description 保存厨显设置
// @Tags 商家端.厨显设置
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.SaveKitchenSettingReq true "保存厨显设置"
// @Success 200 {object} dto.Response
// @Router /shop/setting/kitchen [post]
func (h *SettingHandler) SaveKitchenSetting(c *gin.Context) {
	ctx := helper.GetContext(c)
	var kitchenSettingReq req.SaveKitchenSettingReq
	if err := c.ShouldBindJSON(&kitchenSettingReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 参数验证
	if err := kitchenSettingReq.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 调用 Service 层保存
	err := h.settingSrv.SaveKitchenSetting(ctx, kitchenSettingReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	helper.Success(c, "保存成功")
}

// UploadKioskCarousel 上传自助点餐机轮播内容（图片/视频）
// @Summary 上传自助点餐机轮播内容
// @Description 上传自助点餐机轮播内容，支持图片（JPG、JPEG、PNG、WEBP，<2MB）和视频（MP4，<10MB）
// @Tags 商家端.自助点餐机设置
// @Accept multipart/form-data
// @Produce json
// @Security JwtToken
// @Param file formData file true "文件"
// @Param file_type formData string false "文件类型：image 或 video，不传则自动识别"
// @Param group_id formData int false "分组ID"
// @Success 200 {object} dto.Response{data=resp.UploadFileResp}
// @Router /shop/setting/kiosk/carousel/upload [post]
func (h *SettingHandler) UploadKioskCarousel(c *gin.Context) {
	ctx := helper.GetContext(c)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}

	// 打开文件
	fileReader, err := file.Open()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	defer fileReader.Close()

	// 获取文件类型参数（可选，如果不传则根据文件扩展名自动识别）
	fileTypeParam := c.PostForm("file_type")

	// 获取文件扩展名
	fileName := file.Filename
	extension := ""
	if len(fileName) > 0 {
		dotIndex := -1
		for i := len(fileName) - 1; i >= 0; i-- {
			if fileName[i] == '.' {
				dotIndex = i
				break
			}
		}
		if dotIndex >= 0 && dotIndex < len(fileName)-1 {
			extension = fileName[dotIndex+1:]
		}
	}
	extension = strings.ToLower(extension)

	// 获取分组ID
	groupId := uint64(0)
	if groupIdStr := c.PostForm("group_id"); groupIdStr != "" {
		if id, err := strconv.ParseUint(groupIdStr, 10, 64); err == nil {
			groupId = id
		}
	}

	var uploadFileResp *resp.UploadFileResp

	// 根据文件类型或扩展名判断是图片还是视频
	if fileTypeParam == "image" || (fileTypeParam == "" && (extension == "jpg" || extension == "jpeg" || extension == "png" || extension == "webp")) {
		// 图片上传：JPG、JPEG、PNG、WEBP，<2MB
		allowedExts := []string{"jpg", "jpeg", "png", "webp"}
		isAllowed := slices.Contains(allowedExts, extension)
		if !isAllowed {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("图片仅支持JPG、JPEG、PNG、WEBP格式"))
			return
		}

		// 检查文件大小：2MB
		maxSizeBytes := int64(2 * 1024 * 1024)
		if file.Size > maxSizeBytes {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("图片文件大小不能超过2MB"))
			return
		}

		uploadFileResp, err = h.uploadFileSrv.UploadImage(ctx, fileReader, fileName, file.Size, groupId, "shop")
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, err)
			return
		}
	} else if fileTypeParam == "video" || (fileTypeParam == "" && extension == "mp4") {
		// 视频上传：MP4，<10MB
		if extension != "mp4" {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("视频仅支持MP4格式"))
			return
		}

		// 检查文件大小：10MB
		maxSizeBytes := int64(10 * 1024 * 1024)
		if file.Size > maxSizeBytes {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.New("视频文件大小不能超过10MB"))
			return
		}

		uploadFileResp, err = h.uploadFileSrv.UploadVideo(ctx, fileReader, fileName, file.Size, groupId, 10)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, err)
			return
		}
	} else {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.New("不支持的文件格式，图片支持JPG、JPEG、PNG、WEBP，视频支持MP4"))
		return
	}

	helper.Success(c, uploadFileResp)
}

// GetCompanyList 获取可见门店列表
// @Summary 获取可见门店列表
// @Description 获取可见门店列表
// @Tags 商家端.门店管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.SaasCompanyListResp}
// @Router /shop/company/list [get]
func (h *SettingHandler) GetCompanyList(c *gin.Context) {
	ctx := helper.GetContext(c)
	companyList, err := h.companySrv.GetCompanyList(ctx)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, companyList)
}

// GetCompanyInfo 获取门店信息
// @Summary 获取门店信息
// @Description 获取门店信息（仅总部可用）
// @Tags 商家端.门店管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param uuid query uint64 true "门店UUID"
// @Success 200 {object} dto.Response{data=resp.CompanyStoreResp}
// @Router /shop/company/info [get]
func (h *SettingHandler) GetCompanyInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	var getCompanyInfoReq req.GetCompanyInfoReq
	if err := c.ShouldBindQuery(&getCompanyInfoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	companyInfo, err := h.companySrv.GetCompanyInfo(ctx, getCompanyInfoReq.Uuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, companyInfo)
}

// UpdateCompanyInfo 修改门店信息
// @Summary 修改门店信息
// @Description 修改门店信息（仅总部可用）
// @Tags 商家端.门店管理
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data body req.UpdateCompanySettingReq true "更新门店信息"
// @Success 200 {object} dto.Response
// @Router /shop/company/update [post]
func (h *SettingHandler) UpdateCompanyInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	var updateReq req.UpdateCompanySettingReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	if err := h.companySrv.UpdateCompanyInfo(ctx, updateReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	helper.Success(c, "更新成功")
}

func RegisterSettingHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := settingSrv.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	otherSrv := service.NewOtherSrv(dbm, cache, settingSrv)
	translateSrv := service.NewTranslateSrv(dbm, cache)
	messageSrv := message.NewIMessageSrv(dbm)
	supplierSrv := service.NewSupplierSrv(dbm)
	productSrv := service.NewProductSrv(dbm, service.NewLocaleSrv(), settingSrv, cache, translateSrv)
	materialSrv := service.NewMaterialSrv(dbm, service.NewLocaleSrv(), settingSrv, translateSrv, messageSrv)
	warehouseSrv := service.NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	wrapper := &SettingHandler{
		dbm:           dbm,
		settingSrv:    settingSrv,
		otherSrv:      otherSrv,
		uploadFileSrv: service.NewUploadFileSrv(dbm),
		syncSrv:       service.NewSyncSrv(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv, paymentMethodSrv),
		dataManageSrv: service.NewDataManageSrv(dbm, settingSrv),
		companySrv:    service.NewCompanySrv(dbm, settingSrv),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/setting/store", wrapper.SaveStoreSetting)                        // 保存门店设置
		privateApi.POST("/setting/business", wrapper.SaveBusinessSetting)                  // 保存业务设置
		privateApi.GET("/setting/business", wrapper.GetBusinessSetting)                    // 获取业务设置
		privateApi.GET("/setting/cashier", wrapper.GetCashierSetting)                      // 获取收银机设置
		privateApi.POST("/setting/cashier", wrapper.SaveCashierSetting)                    // 保存收银机设置
		privateApi.POST("/setting/cashier/carousel/upload", wrapper.UploadCashierCarousel) // 上传收银机轮播内容
		privateApi.GET("/setting/print_setting/get", wrapper.GetPrintSetting)              // 获取打印设置
		privateApi.POST("/setting/print_setting/update", wrapper.UpdatePrintSetting)       // 更新打印设置
		privateApi.GET("/setting/kiosk", wrapper.GetKioskSetting)                          // 获取自助点餐机设置
		privateApi.POST("/setting/kiosk", wrapper.SaveKioskSetting)                        // 保存自助点餐机设置
		privateApi.POST("/setting/kiosk/carousel/upload", wrapper.UploadKioskCarousel)     // 上传自助点餐机轮播内容
		privateApi.GET("/setting/kitchen", wrapper.GetKitchenSetting)                      // 获取厨显设置
		privateApi.POST("/setting/kitchen", wrapper.SaveKitchenSetting)                    // 保存厨显设置
		privateApi.GET("/setting/store_scan_order", wrapper.GetStoreScanOrderSetting)      // 获取门店点餐配置
		privateApi.POST("/setting/store_scan_order", wrapper.SaveStoreScanOrderSetting)    // 保存门店点餐配置
		privateApi.GET("/setting/template_style", wrapper.GetTemplateStyleSetting)         // 获取模板样式配置
		privateApi.POST("/setting/template_style", wrapper.SaveTemplateStyleSetting)       // 保存模板样式配置

		privateApi.GET("/setting/free_reason", wrapper.GetFreeReason)                     // 获取免单原因
		privateApi.POST("/setting/free_reason/add", wrapper.AddFreeReason)                // 新增免单原因
		privateApi.POST("/setting/free_reason/edit", wrapper.EditFreeReason)              // 编辑免单原因
		privateApi.DELETE("/setting/free_reason", wrapper.DeleteFreeReason)               // 删除免单原因
		privateApi.GET("/setting/return_food_reason", wrapper.GetReturnFoodReason)        // 获取退菜原因
		privateApi.POST("/setting/return_food_reason/add", wrapper.AddReturnFoodReason)   // 新增退菜原因
		privateApi.POST("/setting/return_food_reason/edit", wrapper.EditReturnFoodReason) // 编辑退菜原因
		privateApi.DELETE("/setting/return_food_reason", wrapper.DeleteReturnFoodReason)  // 删除退菜原因
		privateApi.GET("/setting/order_remark", wrapper.GetOrderRemark)                   // 获取整单备注
		privateApi.POST("/setting/order_remark/add", wrapper.AddOrderRemark)              // 新增整单备注
		privateApi.POST("/setting/order_remark/edit", wrapper.EditOrderRemark)            // 编辑整单备注
		privateApi.DELETE("/setting/order_remark", wrapper.DeleteOrderRemark)             // 删除整单备注
		privateApi.GET("/setting/order_item_remark", wrapper.GetOrderItemRemark)          // 获取单品备注
		privateApi.POST("/setting/order_item_remark/add", wrapper.AddOrderItemRemark)     // 新增单品备注
		privateApi.POST("/setting/order_item_remark/edit", wrapper.EditOrderItemRemark)   // 编辑单品备注
		privateApi.DELETE("/setting/order_item_remark", wrapper.DeleteOrderItemRemark)    // 删除单品备注
		privateApi.GET("/setting/payment_method/list", wrapper.GetPaymentMethodList)      // 获取支付方式列表
		// 电子菜单二维码
		privateApi.GET("/setting/menu_qrcode", wrapper.GetMenuQrcode) // 获取电子菜单二维码
		// 会员端二维码
		privateApi.GET("/setting/member_qrcode", wrapper.GetMemberQrcode)                  // 获取会员端二维码
		privateApi.POST("/setting/upload_logo", wrapper.UploadLogo)                        // 上传logo
		privateApi.POST("/setting/sync", wrapper.Sync)                                     // 获取总部最新数据
		privateApi.GET("/setting/sync_task/list", wrapper.GetSyncTaskList)                 // 获取同步任务列表
		privateApi.GET("/setting/sync_task/detail", wrapper.GetSyncTaskDetail)             // 获取同步任务详情
		privateApi.GET("/setting/headquarters_data_list", wrapper.GetHeadquartersDataList) // 获取总部可同步数据列表
		privateApi.POST("/setting/granular_sync", wrapper.GranularSync)                    // 颗粒化同步数据
		// 数据管理
		privateApi.GET("/setting/data_manage", wrapper.GetDataManage)      // 获取数据管理
		privateApi.POST("/setting/data_manage/set", wrapper.SetDataManage) // 设置数据管理
		// 数据管理 - 拆分接口
		privateApi.POST("/setting/data_manage/save_status", wrapper.SaveDataManageStatus)                 // 保存数据管理开关状态
		privateApi.POST("/setting/data_manage/set_staff", wrapper.SetDataManageStaff)                     // 设置操作人员
		privateApi.GET("/setting/data_manage/order_list", wrapper.GetDataManageOrderList)                 // 获取已选订单列表
		privateApi.POST("/setting/data_manage/order_restore", wrapper.RestoreDataManageOrder)             // 恢复单条已选订单
		privateApi.GET("/setting/data_manage/order_select", wrapper.GetDataManageOrderSelect)             // 获取可选订单列表
		privateApi.POST("/setting/data_manage/order_submit", wrapper.SubmitDataManageOrder)               // 提交订单选择
		privateApi.POST("/setting/data_manage/order_select_stats", wrapper.GetDataManageOrderSelectStats) // 可选订单统计预览

		// 门店管理（总部功能）
		privateApi.GET("/company/list", wrapper.GetCompanyList)       // 获取总部下所有门店列表
		privateApi.GET("/company/info", wrapper.GetCompanyInfo)       // 获取门店信息
		privateApi.POST("/company/update", wrapper.UpdateCompanyInfo) // 修改门店信息
	}
}

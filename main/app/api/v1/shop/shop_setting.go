package shop

import (
	"errors"
	"strconv"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// AuthHandler 认证鉴权控制器
type SettingHandler struct {
	syncSrv       service.ISyncSrv
	settingSrv    setting.ISrv
	otherSrv      service.IOtherSrv
	uploadFileSrv service.IUploadFileSrv
	dataManageSrv service.IDataManageSrv
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
	result := h.settingSrv.GetPaymentMethodList(ctx)
	helper.Success(c, result)
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

func RegisterSettingHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
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
	wrapper := &SettingHandler{
		settingSrv:    settingSrv,
		otherSrv:      otherSrv,
		uploadFileSrv: service.NewUploadFileSrv(dbm),
		syncSrv:       service.NewSyncSrv(dbm, warehouseSrv, supplierSrv, productSrv, materialSrv),
		dataManageSrv: service.NewDataManageSrv(dbm),
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.POST("/setting/store", wrapper.SaveStoreSetting)                        // 保存门店设置
		privateApi.POST("/setting/business", wrapper.SaveBusinessSetting)                  // 保存业务设置
		privateApi.GET("/setting/business", wrapper.GetBusinessSetting)                    // 获取业务设置
		privateApi.POST("/setting/cashier", wrapper.SaveCashierSetting)                    // 保存收银机设置
		privateApi.POST("/setting/cashier/carousel/upload", wrapper.UploadCashierCarousel) // 上传收银机轮播内容

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
		privateApi.GET("/setting/payment_method/list", wrapper.GetPaymentMethodList)      // 获取支付方式列表
		// 电子菜单二维码
		privateApi.GET("/setting/menu_qrcode", wrapper.GetMenuQrcode) // 获取电子菜单二维码
		// 会员端二维码
		privateApi.GET("/setting/member_qrcode", wrapper.GetMemberQrcode)      // 获取会员端二维码
		privateApi.POST("/setting/upload_logo", wrapper.UploadLogo)            // 上传logo
		privateApi.POST("/setting/sync", wrapper.Sync)                         // 获取总部最新数据
		privateApi.GET("/setting/sync_task/list", wrapper.GetSyncTaskList)     // 获取同步任务列表
		privateApi.GET("/setting/sync_task/detail", wrapper.GetSyncTaskDetail) // 获取同步任务详情
		// 数据管理
		privateApi.GET("/setting/data_manage", wrapper.GetDataManage)      // 获取数据管理
		privateApi.POST("/setting/data_manage/set", wrapper.SetDataManage) // 设置数据管理
	}
}

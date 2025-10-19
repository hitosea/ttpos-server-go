package cashier

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/app/tasks"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/middleware"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// DeskHandler 桌台处理程序
type DeskHandler struct {
	deskSrv   service.IDeskSrv   // 主服务
	memberSrv service.IMemberSrv // 会员服务
	orderSrv  service.IOrderSrv
}

// GetDeskRegionAndType 处理获取桌台的区域和类型
// @Summary 获取桌台的区域和类型
// @Description 获取桌台的区域和类型
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} resp.DeskRegionAndTypeListWithPaginationResp "桌台区域和类型列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/region_and_type [get]
func (h *DeskHandler) GetDeskRegionAndType(c *gin.Context) {
	companyId := helper.GetCompanyUuid(c)
	// 处理获取桌台的区域和类型的逻辑
	res, err := h.deskSrv.GetDeskRegionAndTypeList(companyId)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskList 处理获取桌台列表
// @Summary 获取桌台列表
// @Description 获取桌台列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskListReq true "列表参数"
// @Success 200 {object} resp.DeskListWithPaginationResp "收银台列表"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/list [get]
func (h *DeskHandler) GetDeskList(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var listReq req.DeskListReq
	if err := c.ShouldBindQuery(&listReq); err != nil {
		helper.HandleValidationError(c, err, listReq, dto.PageReqMessage)
		return
	}
	// 获取收银产品列表
	res, err := h.deskSrv.GetDeskList(ctx, companyUuid, listReq)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// GetDeskInfo 处理获取收银台列表
// @Summary 获取桌台详情
// @Description 获取桌台详情
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskInfoReq true "详情参数"
// @Success 200 {object} resp.Desk "桌台详情"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/info [get]
func (h *DeskHandler) GetDeskInfo(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	// 绑定请求参数
	var infoReq req.DeskInfoReq
	if err := c.ShouldBindQuery(&infoReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 获取收银产品列表
	res, err := h.deskSrv.GetDeskInfo(companyUuid, infoReq.Uuid)
	// 处理错误
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// CreateDeskOrder 处理创建开台
// @Summary 开台
// @Description 开台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskOrderCreateReq true "开台参数"
// @Success 200 {object} resp.CreateDeskOrderResp "开台成功"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/open [post]
func (h *DeskHandler) CreateDeskOrder(c *gin.Context) {

	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskOrderCreateReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}

	// 创建桌台订单
	res, err := h.deskSrv.CreateDeskOrder(ctx, params)
	// 处理错误
	if err != nil {
		ctx.Log().Error("创建桌台订单失败", zap.Error(err))
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}

	// 返回结果
	helper.Success(c, res)
}

// CloseDesk 处理关闭桌台
// @Summary 关闭桌台
// @Description 关闭桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.DeskCloseReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/close [post]
func (h *DeskHandler) CloseDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskCloseReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.deskSrv.CloseDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// CompleteDesk 处理清台
// @Summary 清台
// @Description 清台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.DeskInfoReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/complete [post]
func (h *DeskHandler) CompleteDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.DeskJsonUuidReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	err := h.deskSrv.CompleteDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// ChangeDesk 处理切换桌台
// @Summary 切换桌台
// @Description 切换桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.ChangeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/change [post]
func (h *DeskHandler) ChangeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.ChangeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, err := h.deskSrv.ChangeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// MergeTable 处理合并桌台
// @Summary 合并桌台
// @Description 合并桌台
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.MergeDeskReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/merge [post]
func (h *DeskHandler) MergeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.MergeDeskReq{}
	if err := c.ShouldBind(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.DeskReqMessage)
		return
	}
	//
	info, deskMergeCheckResp, err := h.deskSrv.MergeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithData(c, constant.CodeFail, deskMergeCheckResp, err)
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// CancelDeskOrder 处理取消桌台订单
// @Summary 取消桌台订单
// @Description 取消桌台订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderCancelReq true "详情参数"
// @Success 200 {object} nil
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cancel [post]
func (h *DeskHandler) CancelDeskOrder(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	var cancelReq req.OrderCancelReq
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	//
	err := h.orderSrv.CancelOrder(ctx, cancelReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderProductDelete 处理删除桌台订单商品
// @Summary 删除桌台订单商品
// @Description 删除桌台订单商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductDeleteReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/delete [delete]
func (h *DeskHandler) OrderProductDelete(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	staff := helper.GetStaff(c)
	source := helper.GetSource(c)
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	shopCart, err := h.orderSrv.OrderProductDelete(ctx, companyUuid, staff.Uuid, source, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// OrderProductChangePrice 处理桌台订单商品改价
// @Summary 桌台订单商品改价
// @Description 桌台订单商品改价
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductChangePriceReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/price [post]
func (h *DeskHandler) OrderProductChangePrice(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductChangePriceReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderProductChangePrice(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderDiscount 处理桌台订单打折
// @Summary 桌台订单打折
// @Description 桌台订单打折
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountMethodReq true "打折参数，根据discount_method值(1:改价,2:打折,3:抹零)提供对应的额外字段"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/discount [post]
func (h *DeskHandler) OrderDiscount(c *gin.Context) {
	ctx := helper.GetContext(c)
	bodyBytes, _ := io.ReadAll(c.Request.Body) // Body只能读取一次，之后想再次从body中读取数据需要重新往body中写入数据
	// 绑定请求参数
	params := req.OrderDiscountMethodReq{}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新写入数据
	// 从body中读取数据
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新写入数据

	var shopCart *resp.ShopCart
	var err error
	// 改价
	if params.DiscountMethod == 1 {
		amountChangeReq := req.OrderAmountChangeReq{}
		if err := c.ShouldBindJSON(&amountChangeReq); err != nil {
			helper.HandleValidationError(c, err, amountChangeReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderAmountChange(ctx, amountChangeReq)
	}
	// 打折
	if params.DiscountMethod == 2 {
		discountReq := req.OrderDiscountReq{}
		if err := c.ShouldBindJSON(&discountReq); err != nil {
			helper.HandleValidationError(c, err, discountReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderDiscount(ctx, discountReq)
	}
	// 抹零
	if params.DiscountMethod == 3 {
		zeroRuleReq := req.OrderZeroRuleReq{}
		if err := c.ShouldBindJSON(&zeroRuleReq); err != nil {
			helper.HandleValidationError(c, err, zeroRuleReq, req.OrderReqMessage)
			return
		}
		shopCart, err = h.orderSrv.OrderZeroRule(ctx, zeroRuleReq)
	}
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, shopCart)
}

// @Summary 取消桌台订单所有优惠折扣，包括改价、打折、抹零
// @Description 取消桌台订单所有优惠折扣，包括改价、打折、抹零
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderDiscountCancelReq true "取消优惠折扣参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/discount/cancel [post]
func (h *DeskHandler) OrderDiscountCancel(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderDiscountCancelReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderDiscountCancel(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info, "操作成功")
}

// OrderChangePopulation 处理桌台订单修改人数
// @Summary 桌台订单修改人数
// @Description 桌台订单修改人数
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangePopulationReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/population [post]
func (h *DeskHandler) OrderChangePopulation(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangePopulationReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangePopulation(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangeBuffet 处理桌台订单调整自助餐
// @Summary 桌台订单调整自助餐
// @Description 桌台订单调整自助餐
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangeBuffetReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/buffet [post]
func (h *DeskHandler) OrderChangeBuffet(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangeBuffet(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderChangeBuffetClock 处理桌台订单自助餐加钟
// @Summary 桌台订单自助餐加钟
// @Description 桌台订单自助餐加钟
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderChangeBuffetClockReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/buffet/clock [post]
func (h *DeskHandler) OrderChangeBuffetClock(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetClockReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderChangeBuffetClock(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, err)
		return
	}
	// 返回结果
	helper.Success(c, info, "加钟成功")
}

// OrderChangeBuffetProductList 处理获取自助餐商品列表
// @Summary 获取自助餐商品列表
// @Description 获取自助餐商品列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderChangeBuffetProductListReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.BuffetProductList}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/buffet/product/list [get]
func (h *DeskHandler) GetDeskBuffetProductList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderChangeBuffetProductListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("获取自助餐商品列表", zap.Any("params", params))
	//
	info, err := h.orderSrv.OrderDeskBuffetProductList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderProductRemark 处理桌台订单商品备注
// @Summary 桌台订单商品备注
// @Description 桌台订单商品备注
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderProductRemarkReq true "详情参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/product/remark [post]
func (h *DeskHandler) OrderProductRemark(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderProductRemarkReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	//
	info, err := h.orderSrv.OrderProductRemark(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, info)
}

// OrderCartInfo 处理查询桌台购物车信息
// @Summary 查询桌台购物车信息
// @Description 查询桌台购物车信息
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Param data query req.OrderCartInfoReq true "请求参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/info [get]
func (h *DeskHandler) OrderCartInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartInfoReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 查询购物车信息。含H5订单的商品。使用场景“接单”-“进入桌台”
	if params.H5OrderUuid > constant.OptionalUuid {
		res, err := h.orderSrv.GetOrderCartInfo(ctx, params.SaleBillUuid, repository.WithH5OrderUuid(params.H5OrderUuid))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
			return
		}
		// 返回结果
		helper.Success(c, res)
	} else {
		// 查询购物车信息
		res, err := h.orderSrv.GetOrderCartInfo(ctx, params.SaleBillUuid)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
			return
		}
		// 返回结果
		helper.Success(c, res)
	}
}

// OrderCartProductAdd 向购物车添加商品
// @Summary 向购物车添加商品
// @Description 向购物车添加商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductAddReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/add [post]
func (h *DeskHandler) OrderCartProductAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 添加商品。 若没有点餐账单则新建一个
	res, err := h.orderSrv.InstantOrderCartProductAdd(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), errors.ErrProductPriceChanged.Error()) {
			helper.ErrorWithData(c, constant.CodeOrderCheckProductPriceChanged, res, fmt.Errorf("%s", i18n.Translate(ctx.GetLanguage(), err.Error())))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductPackageAdd 向购物车添加套餐
// @Summary 向购物车添加套餐
// @Description 向购物车添加套餐
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductPackageAddReq true "套餐参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product_package/add [post]
func (h *DeskHandler) OrderCartProductPackageAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductPackageAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 向购物车添加套餐
	res, err := h.orderSrv.OrderCartProductPackageAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductFlavorAndAttribute 查询购物车商品“规格/属性”
// @Summary 查询购物车商品“规格/属性”
// @Description 查询购物车商品“规格/属性”
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.OrderCartProductFlavorAndAttributeReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ProductFlavorAndAttributeRes}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/flavor_and_attribute [get]
func (h *DeskHandler) OrderCartProductFlavorAndAttribute(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductFlavorAndAttributeReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 查询购物车商品“规格/属性”
	res, err := h.orderSrv.OrderCartProductFlavorAndAttribute(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductFlavorAndAttributeChange 修改购物车商品“规格/属性”
// @Summary 修改购物车商品“规格/属性”
// @Description 修改购物车商品“规格/属性”
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductFlavorAndAttributeChangeReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/flavor_and_attribute [post]
func (h *DeskHandler) OrderCartProductFlavorAndAttributeChange(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductFlavorAndAttributeChangeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 修改购物车商品“规格/属性”
	res, err := h.orderSrv.OrderCartProductFlavorAndAttributeChange(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductNum 修改购物车商品数量
// @Summary 修改购物车某个商品的数量
// @Description 修改购物车商品数量
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductNumReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/num [post]
func (h *DeskHandler) OrderCartProductNum(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductNumReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("桌台页面修改购物车商品数量接口请求", zap.Any("params", params))
	// 修改购物车商品数量
	res, err := h.orderSrv.OrderCartProductNum(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCooking 送厨购物车商品
// @Summary 送厨购物车商品
// @Description 送厨购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductCookingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/cooking [post]
func (h *DeskHandler) OrderCartProductCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	ctx.Log().Debug("桌台页面送厨购物车商品接口请求", zap.Any("params", params))
	// 送厨购物车商品
	res, checkRes, err := h.orderSrv.InstantOrderCartProductCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	// 返回结果
	code := res.GetCode()
	helper.FailWithData(c, code, res, nil, constant.ParseCodeOrderCheck(code))
	// helper.Success(c, res)
}

// OrderCartProductReturning 退菜购物车商品
// @Summary 退菜购物车商品
// @Description 退菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/returning [post]
func (h *DeskHandler) OrderCartProductReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductReturningReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCancelReturning 取消退菜购物车商品
// @Summary 取消退菜购物车商品
// @Description 取消退菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/cancel_returning [post]
func (h *DeskHandler) OrderCartProductCancelReturning(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 退菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductCancelReturning(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductChangeDesk 转菜购物车商品
// @Summary 转菜购物车商品
// @Description 转菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductChangeDeskReq true "转菜购物车商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/change_desk [post]
func (h *DeskHandler) OrderCartProductChangeDesk(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductChangeDeskReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 转菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductChangeDesk(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("转菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductWrap 打包单商品
// @Summary 打包单商品
// @Description 打包单商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductWrapReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/wrap [post]
func (h *DeskHandler) OrderCartProductWrap(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductWrapReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 打包单商品
	res, err := h.orderSrv.OrderCartProductWrap(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("打包单商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductUnwrap 取消打包单商品
// @Summary 取消打包单商品
// @Description 取消打包单商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductWrapReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/unwrap [post]
func (h *DeskHandler) OrderCartProductUnwrap(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductUnwrapReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 取消打包单商品
	res, err := h.orderSrv.OrderCartProductUnwrap(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消打包单商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductGiving 赠菜购物车商品
// @Summary 赠菜购物车商品
// @Description 赠菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductGivingReq true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/giving [post]
func (h *DeskHandler) OrderCartProductGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductGivingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 赠菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductCancelGiving 取消赠菜购物车商品
// @Summary 取消赠菜购物车商品
// @Description 取消赠菜购物车商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProduct true "商品参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product/cancel_giving [post]
func (h *DeskHandler) OrderCartProductCancelGiving(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProduct{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 取消赠菜购物车商品
	res, err := h.orderSrv.InstantOrderCartProductCancelGiving(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("取消退菜购物车商品成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderMustPlanConfirm 确认必点商品
// @Summary 确认必点商品
// @Description 确认必点商品
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderMustPlanConfirmReq true "确认必点商品参数"
// @Success 200 {object} dto.Response{}
// @Router /cashier/desk/order/must_plan/confirm [post]
func (h *DeskHandler) OrderMustPlanConfirm(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面确认必点商品接口请求")

	params := req.InstantOrderMustPlanConfirmReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("确认必点商品", zap.Any("params", params))
	// 确认必点商品
	res, mustPlan, err := h.orderSrv.InstantOrderMustPlanConfirm(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if !res {
		helper.ErrorWithDetail(c, constant.CodeOrderCheckProductMust, errors.New(fmt.Sprintf("【%s】%s", mustPlan.Name, i18n.Translate(ctx.GetLanguage(), errors.ErrMustPlanNotComplete.Error()))))
		return
	}
	// 返回结果
	helper.Success(c, gin.H{})
}

// OrderCheck 订单检查
// @Summary 订单检查
// @Description 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @Success 200 {object} dto.Response{data=resp.OrderCheckRes}
// @Router /cashier/desk/order/check [get]
func (h *DeskHandler) OrderCheck(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面订单检查接口请求")

	params := req.InstantOrderCheckReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("订单检查", zap.Any("params", params))
	// 订单检查
	checkRes, err := h.orderSrv.OrderCheck(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if checkRes != nil {
		ctx.Log().Debug("送厨检查不通过", zap.Any("res", checkRes))
		helper.FailWithData(c, checkRes.Code, checkRes.OrderCheckRes, nil, constant.ParseCodeOrderCheck(checkRes.Code))
		return
	}
	ctx.Log().Debug("订单检查成功")
	// 返回结果
	helper.Success(c, resp.OrderCheckRes{})
}

// OrderPaymentInfo 获取结账页面信息
// @Summary 获取结账页面信息
// @Description 获取结账页面信息
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_bill_uuid query string true "销售账单UUID"
// @param sale_order_uuid query string true "销售订单UUID"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/payment/info [get]
func (h *DeskHandler) OrderPaymentInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面结账页面信息接口请求")

	params := &req.InstantOrderPaymentInfoReq{}
	if err := params.Parse(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("查询销售订单收银机结账页面信息", zap.Any("params", params))
	// 获取销售订单的付款信息
	res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, nil, params.SaleBillUuid, params.SaleOrderUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCoupon 选择或取消优惠券
// @Summary 选择或取消优惠券
// @Description 选择或取消优惠券
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCouponReq true "选择或取消优惠券参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/payment/coupon [post]
func (h *DeskHandler) OrderPaymentCoupon(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面选择或取消优惠券接口请求")

	var couponReq req.InstantOrderPaymentCouponReq
	if err := c.ShouldBindJSON(&couponReq); err != nil {
		helper.HandleValidationError(c, err, couponReq, nil)
		return
	}
	ctx.Log().Info("选择或取消优惠券", zap.Any("params", couponReq))
	// 选择或取消优惠券
	res, err := h.orderSrv.OrderPaymentCoupon(ctx, couponReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentInfo 设置订单的抵扣积分数量
// @Summary 设置订单的抵扣积分数量
// @Description 设置订单的抵扣积分数量
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentPointsReq true "设置订单的抵扣积分数量参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/payment/points [post]
func (h *DeskHandler) OrderPaymentPoints(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面设置订单的抵扣积分数量接口请求")

	params := req.InstantOrderPaymentPointsReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("设置订单的抵扣积分数量", zap.Any("params", params))
	// 设置订单的抵扣积分数量
	res, err := h.orderSrv.OrderPaymentPoints(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentQrcode 获取支付方式的二维码信息
// @Summary 获取支付方式的二维码信息
// @Description 获取支付方式的二维码信息
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data query req.InstantOrderPaymentQrcodeReq true "获取支付方式的二维码信息参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentQrcodeInfoResp}
// @Router /cashier/desk/order/payment/qrcode [get]
func (h *DeskHandler) OrderPaymentQrcodeInfo(c *gin.Context) {
	ctx := helper.GetContext(c)
	params := req.InstantOrderPaymentQrcodeReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	// 获取支付二维码
	res, err := h.orderSrv.InstantOrderPaymentQrcode(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCreate 创建一个支付单
// @Summary 创建一个支付单
// @Description 创建一个支付单
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCreateReq true "创建一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /cashier/desk/order/payment/create [post]
func (h *DeskHandler) OrderPaymentCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面创建一个支付单接口请求")

	params := req.InstantOrderPaymentCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("创建一个支付单", zap.Any("params", params))
	// 创建一个支付单
	res, err := h.orderSrv.InstantOrderPaymentCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("创建一个支付单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentCancel 撤销一个支付单
// @Summary 撤销一个支付单
// @Description 撤销一个支付单
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentCancelReq true "撤销一个支付单参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /cashier/desk/order/payment/cancel [post]
func (h *DeskHandler) OrderPaymentCancel(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面撤销一个支付单接口请求")

	params := req.InstantOrderPaymentCancelReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("撤销一个支付单", zap.Any("params", params))
	// 撤销一个支付单
	res, err := h.orderSrv.InstantOrderPaymentCancel(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("撤销一个支付单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentFinish 完成销售订单的付款结账
// @Summary 完成销售订单的付款结账
// @Description 完成销售订单的付款结账
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentFinishReq true "完成销售订单的付款结账参数"
// @Success 200 {object} dto.Response{data=resp.OrderFinishResp}
// @Router /cashier/desk/order/payment/finish [post]
func (h *DeskHandler) OrderPaymentFinish(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面销售订单的付款结账接口请求")

	params := req.InstantOrderPaymentFinishReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("桌台销售订单的付款结账", zap.Any("params", params))
	// 桌台销售订单的付款结账
	res, err := h.orderSrv.InstantOrderPaymentFinish(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "请刷新优惠券列表") {
			// 获取销售订单的付款信息
			res, err := h.orderSrv.InstantOrderPaymentInfo(ctx, nil, params.SaleBillUuid, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			// 返回结果
			helper.ErrorWithData(c, constant.CodeCouponInvalid, res, fmt.Errorf("%s", i18n.Translate(ctx.GetLanguage(), "优惠券信息变化，请重新确认。")))
			return
		}
		if strings.Contains(err.Error(), "物品库存不足") {
			ctx.Log().Error("桌台销售订单的付款结账失败", zap.Any("err", err))
			itemCode := ""
			re := regexp.MustCompile(`物品库存不足,(WPR\d+)`)
			matches := re.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				itemCode = matches[1]
			}
			productInfos, err := h.orderSrv.GetProductNameByItemCode(ctx, itemCode, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			productList := make([]resp.Product, 0)
			for _, productInfo := range productInfos {
				productList = append(productList, resp.Product{
					LocaleName: productInfo.ProductName,
				})
			}
			orderCheckRes := &resp.OrderCheckRes{
				Products: &resp.CartProductList{
					List: productList,
				},
			}
			helper.FailWithData(c, constant.CodeOrderCheckProductStockZero, orderCheckRes, nil, i18n.Translate(ctx.GetLanguage(), "以下商品库存不足，请删除后再下单"))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("桌台销售订单的付款结账成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderFree 免单
// @Summary 免单
// @Description 免单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderFreeReq true "免单参数"
// @Success 200 {object} dto.Response{data=resp.OrderFinishResp}
// @Router /cashier/desk/order/free [post]
func (h *DeskHandler) OrderFree(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面免单接口请求")

	params := req.InstantOrderFreeReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("桌台免单", zap.Any("params", params))
	// 桌台免单
	res, err := h.orderSrv.InstantOrderFree(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "物品库存不足") {
			ctx.Log().Error("桌台销售订单的付款结账失败", zap.Any("err", err))
			itemCode := ""
			re := regexp.MustCompile(`物品库存不足,(WPR\d+)`)
			matches := re.FindStringSubmatch(err.Error())
			if len(matches) > 1 {
				itemCode = matches[1]
			}
			productInfos, err := h.orderSrv.GetProductNameByItemCode(ctx, itemCode, params.SaleOrderUuid)
			if err != nil {
				helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
				return
			}
			productList := make([]resp.Product, 0)
			for _, productInfo := range productInfos {
				productList = append(productList, resp.Product{
					LocaleName: productInfo.ProductName,
				})
			}
			orderCheckRes := &resp.OrderCheckRes{
				Products: &resp.CartProductList{
					List: productList,
				},
			}
			helper.FailWithData(c, constant.CodeOrderCheckProductStockZero, orderCheckRes, nil, i18n.Translate(ctx.GetLanguage(), "以下商品库存不足，请删除后再下单"))
			return
		}
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("桌台免单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderPaymentZeroRule 设置结账抹零规则
// @Summary 设置结账抹零规则
// @Description 设置结账抹零规则
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderPaymentZeroRuleReq true "设置结账抹零规则参数"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp}
// @Router /cashier/desk/order/payment/zero_rule [post]
func (h *DeskHandler) OrderPaymentZeroRule(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面设置结账抹零规则接口请求")

	params := req.InstantOrderPaymentZeroRuleReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	if err := params.Validate(); err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Info("设置结账抹零规则", zap.Any("params", params))
	// 设置结账抹零规则
	res, err := h.orderSrv.InstantOrderPaymentZeroRule(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("设置结账抹零规则成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderCreate 创建一个销售订单
// @Summary 创建一个销售订单
// @Description 创建一个销售订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderCreateReq true "创建一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/create [post]
func (h *DeskHandler) OrderSaleOrderCreate(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面创建一个销售订单接口请求")

	params := req.InstantOrderSaleOrderCreateReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("创建一个销售订单", zap.Any("params", params))
	// 创建一个销售订单
	res, err := h.orderSrv.InstantOrderSaleOrderCreate(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("创建一个销售订单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderMoveProduct 从一个销售订单移动商品到另一个销售订单
// @Summary 从一个销售订单移动商品到另一个销售订单
// @Description 从一个销售订单移动商品到另一个销售订单
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderMoveProductReq true "从一个销售订单移动商品到另一个销售订单参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/move_product [post]
func (h *DeskHandler) OrderSaleOrderMoveProduct(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面从一个销售订单移动商品到另一个销售订单接口请求")

	params := req.InstantOrderSaleOrderMoveProductReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("从一个销售订单移动商品到另一个销售订单", zap.Any("params", params))
	// 从一个销售订单移动商品到另一个销售订单
	res, err := h.orderSrv.SaleOrderMoveProduct(ctx, params, false)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("从一个销售订单移动商品到另一个销售订单成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res, "拆单成功")
}

// OrderSaleOrderDelete 删除一个销售订单(删除拆单)
// @Summary 删除一个销售订单(删除拆单)
// @Description 删除一个销售订单(删除拆单)
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderDeleteReq true "删除一个销售订单(删除拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/delete [delete]
func (h *DeskHandler) OrderSaleOrderDelete(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面删除一个销售订单(删除拆单)接口请求")

	params := req.InstantOrderSaleOrderDeleteReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("删除一个销售订单(删除拆单)", zap.Any("params", params))
	// 删除一个销售订单(删除拆单)
	res, err := h.orderSrv.InstantOrderSaleOrderDelete(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("删除一个销售订单(删除拆单)成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// OrderSaleOrderDeleteAll 删除所有子销售订单(撤销拆单)
// @Summary 删除所有子销售订单(撤销拆单)
// @Description 删除所有子销售订单(撤销拆单)
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.InstantOrderSaleOrderDeleteAllReq true "删除所有子销售订单(撤销拆单)参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Router /cashier/desk/order/sale_order/delete_all [delete]
func (h *DeskHandler) OrderSaleOrderDeleteAll(c *gin.Context) {
	ctx := helper.GetContext(c)
	ctx.Log().Debug("收到桌台页面删除所有子销售订单(撤销拆单)接口请求")

	params := req.InstantOrderSaleOrderDeleteAllReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, nil)
		return
	}
	ctx.Log().Info("删除所有子销售订单(撤销拆单)", zap.Any("params", params))
	// 删除所有子销售订单(撤销拆单)
	res, err := h.orderSrv.InstantOrderSaleOrderDeleteAll(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	ctx.Log().Debug("删除所有子销售订单(撤销拆单)成功", zap.Any("res", res))
	// 返回结果
	helper.Success(c, res)
}

// GetMemberDiscount 获取订单会员优惠
// @Summary 获取订单会员优惠
// @Description 获取订单会员优惠
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param sale_order_uuid query integer true "销售订单uuid"
// @param sale_bill_uuid query integer true "销售账单uuid"
// @param member_uuid query integer true "会员Uuid"
// @Success 200 {object} dto.Response{data=resp.MemberDiscountResp}
// @Router /cashier/desk/order/member/discount [get]
func (h *DeskHandler) GetMemberDiscount(c *gin.Context) {
	var discountReq req.GetMemberDiscountReq
	if err := c.ShouldBindQuery(&discountReq); err != nil {
		helper.HandleValidationError(c, err, discountReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	ctx.Log().Info("获取会员优惠", zap.Any("params", discountReq))
	order, err := h.memberSrv.GetMemberDiscount(ctx, discountReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, order)
}

// OrderUseMember 确认使用会员优惠并验证密码
// @Summary 确认使用会员优惠并验证密码
// @Description 确认使用会员优惠并验证密码
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.CheckMemberPasswordReq true "确认使用会员优惠并验证密码"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /cashier/desk/order/member/confirm [post]
func (h *DeskHandler) OrderUseMember(c *gin.Context) {
	var passwordReq req.CheckMemberPasswordReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	res, isCustomAmountAndZero, err := h.orderSrv.OrderUseMember(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	if isCustomAmountAndZero {
		helper.FailWithData(c, constant.CodeSuccess, res, nil, "改价/抹零已失效，请重新进行改价/抹零操作")
		return
	}
	helper.Success(c, res)
}

// OrderMemberCancel 不使用此会员
// @Summary 不使用此会员
// @Description 不使用此会员
// @Tags 收银端.桌台.结账
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderMemberCancelReq true "不使用此会员"
// @Success 200 {object} dto.Response{data=resp.InstantOrderPaymentInfoResp} "结账页面信息"
// @Router /cashier/desk/order/member/cancel [delete]
func (h *DeskHandler) OrderMemberCancel(c *gin.Context) {
	var passwordReq req.OrderMemberCancelReq
	if err := c.ShouldBindJSON(&passwordReq); err != nil {
		helper.HandleValidationError(c, err, passwordReq, req.CheckMemberPasswordMessage)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderMemberCancel(ctx, passwordReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderPrint 打印小票
// @Summary 桌台订单打印小票
// @Description 桌台订单打印小票
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/desk/order/print [post]
func (h *DeskHandler) OrderPrint(c *gin.Context) {
	var printReq req.OrderPrintReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrint(ctx, printReq, true)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderPrint 打印发票
// @Summary 桌台订单打印发票
// @Description 桌台订单打印发票
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderPrintInvoiceReq true "参数"
// @Success 200 {object} dto.Response{data=resp.PrinterData} "打印数据"
// @Router /cashier/desk/order/print/invoice [post]
func (h *DeskHandler) OrderPrintInvoice(c *gin.Context) {
	var printReq req.OrderPrintInvoiceReq
	if err := c.ShouldBindJSON(&printReq); err != nil {
		helper.HandleValidationError(c, err, printReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.OrderPrintInvoice(ctx, printReq)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res, "发送成功")
}

// OrderUnlock 订单解锁
// @Summary 订单解锁
// @Description 订单解锁
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderUnlockReq true "订单解锁"
// @Success 200 {object} dto.Response
// @Router /cashier/desk/order/unlock [post]
func (h *DeskHandler) OrderUnlock(c *gin.Context) {
	var unlockReq req.OrderUnlockReq
	if err := c.ShouldBindJSON(&unlockReq); err != nil {
		helper.HandleValidationError(c, err, unlockReq, nil)
		return
	}
	ctx := helper.GetContext(c)
	err := h.orderSrv.OrderUnlock(ctx, unlockReq.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, gin.H{})
}

// GetOrderMemberList 获取订单会员列表
// @Summary 获取订单会员列表
// @Description 获取订单会员列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response{data=resp.InstantOrderMemberList}
// @Router /cashier/desk/order/member/list [get]
func (h *DeskHandler) GetOrderMemberList(c *gin.Context) {
	var req req.OrderGetOrderMemberListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	ctx := helper.GetContext(c)
	res, err := h.orderSrv.GetOrderMemberList(ctx, req.SaleBillUuid)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// GetDailySalesOutboundSummary 获取每日销售出库汇总
// @Summary 获取每日销售出库汇总
// @Description 获取每日销售出库汇总
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Router /cashier/desk/order/daily_sales_outbound_summary [get]
func (h *DeskHandler) GetDailySalesOutboundSummary(c *gin.Context) {
	companyUuid := helper.GetCompanyUuid(c)
	dbm := database.GetDBManager(config.Database)
	company, err := repository.NewCompanyRepo(dbm.GetDB(companyUuid)).GetCompany(repository.CommonRepo.WhereByUuid(companyUuid))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	tasks.NewDailySalesOutboundSummaryTask(dbm, cache.Global).ProcessCompany(&company)
	helper.Success(c, gin.H{})
}

// GetHeadquarterMaterialList 获取总部物品列表
// @Summary 获取总部物品列表
// @Description 获取总部物品列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @Success 200 {object} dto.Response
// @Router /cashier/desk/order/headquarter_material_list [get]
func (h *DeskHandler) GetHeadquarterMaterialList(c *gin.Context) {
	var req req.GetHeadquarterMaterialListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		helper.HandleValidationError(c, err, req, nil)
		return
	}
	ctx := helper.GetContext(c)
	erpSrv := erp.NewIErpSrv(database.GetDBManager(config.Database))
	res, err := erpSrv.GetHeadquarterMaterialList(ctx, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	helper.Success(c, res)
}

// OrderCartProductBatchCooking 获取分批送厨弹框的销售订单商品列表
// @Summary 获取分批送厨弹框的销售订单商品列表
// @Description 获取分批送厨弹框的销售订单商品列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.GetOrderCartProductBatchCookingListReq true "获取分批送厨弹框的销售订单商品列表"
// @Success 200 {object} dto.Response{data=resp.OrderCartProductBatchCookingRes}
// @Router /cashier/desk/order/cart/batch/cooking [get]
func (h *DeskHandler) OrderCartProductBatchCookingList(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.GetOrderCartProductBatchCookingListReq{}
	if err := c.ShouldBindQuery(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 获取分批送厨弹框的销售订单商品列表
	res, err := h.orderSrv.GetOrderCartProductBatchCookingList(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// OrderCartProductBatchCooking 分批送厨弹框的销售订单商品列表
// @Summary 分批送厨弹框的销售订单商品列表
// @Description 分批送厨弹框的销售订单商品列表
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductBatchCookingReq true "分批送厨弹框的销售订单商品列表"
// @Success 200 {object} dto.Response{data=resp.OrderCartProductBatchCooking}
// @Router /cashier/desk/order/cart/batch/cooking [post]
func (h *DeskHandler) OrderCartProductBatchCooking(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductBatchCookingReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 分批送厨弹框的销售订单商品列表
	res, err := h.orderSrv.OrderCartProductBatchCooking(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}

// RegisterDeskHandlers 注册收银产品路由
func RegisterDeskHandlers(router gin.IRouter, dbm *database.DBManager, cache cache.Cache) {
	// 初始化服务
	captchaSrv := service.NewCaptchaSrv(cache)
	settingSrv := setting.NewSrv(dbm, cache)
	roleAccessSrv := service.NewRoleAccessSrv(dbm)
	deviceSrv := service.NewDeviceSrv(settingSrv, dbm)
	cashBoxSrv := service.NewCashBoxSrv(dbm)
	statisticsSrv := service.NewStatisticsSrv()
	staffShiftSrv := service.NewStaffShiftSrv(cache, dbm, cashBoxSrv, statisticsSrv)
	authSrv := service.NewAuthSrv(dbm, captchaSrv, roleAccessSrv, deviceSrv, staffShiftSrv, settingSrv)
	localeSrv := service.NewLocaleSrv()
	mustPlanSrv := service.NewMustPlanSrv(dbm)
	paymentMethodSrv := service.NewPaymentMethodSrv(dbm, settingSrv)
	memberSrv := service.NewMemberSrv(dbm, cache)
	orderSrv := service.NewOrderSrv(dbm, localeSrv, settingSrv, mustPlanSrv, paymentMethodSrv, memberSrv, cashBoxSrv, service.WithSmsSrv(dbm))

	// 初始化处理器
	wrapper := DeskHandler{
		deskSrv:   service.NewDeskSrv(dbm, localeSrv, orderSrv, settingSrv, deviceSrv, mustPlanSrv),
		memberSrv: memberSrv,
		orderSrv:  orderSrv,
	}

	// 需要认证
	privateApi := router.Group("", middleware.Auth(authSrv, dbm))
	{
		privateApi.GET("/desk/region_and_type", wrapper.GetDeskRegionAndType)                                              // 获取桌台的区域和类型
		privateApi.GET("/desk/list", wrapper.GetDeskList)                                                                  // 获取桌台列表
		privateApi.GET("/desk/info", wrapper.GetDeskInfo)                                                                  // 获取桌台详情
		privateApi.POST("/desk/close", wrapper.CloseDesk)                                                                  // 关闭桌台
		privateApi.POST("/desk/complete", wrapper.CompleteDesk)                                                            // 完成桌台（清台）
		privateApi.POST("/desk/change", wrapper.ChangeDesk)                                                                // 切换桌台（转台）
		privateApi.POST("/desk/open", wrapper.CreateDeskOrder)                                                             // 创建桌台订单(开桌)
		privateApi.POST("/desk/merge", wrapper.MergeDesk)                                                                  // 合并桌台
		privateApi.POST("/desk/order/cancel", wrapper.CancelDeskOrder)                                                     // 取消桌台订单
		privateApi.DELETE("/desk/order/product/delete", wrapper.OrderProductDelete)                                        // 删除桌台订单商品
		privateApi.POST("/desk/order/product/price", wrapper.OrderProductChangePrice)                                      // 桌台订单商品改价
		privateApi.POST("/desk/order/discount", wrapper.OrderDiscount)                                                     // 桌台订单打折
		privateApi.POST("/desk/order/discount/cancel", wrapper.OrderDiscountCancel)                                        // 取消桌台订单所有优惠折扣，包括改价、打折、抹零
		privateApi.POST("/desk/order/population", wrapper.OrderChangePopulation)                                           // 桌台订单修改人数
		privateApi.POST("/desk/order/buffet", wrapper.OrderChangeBuffet)                                                   // 桌台订单调整自助餐
		privateApi.POST("/desk/order/buffet/clock", wrapper.OrderChangeBuffetClock)                                        // 桌台订单调整自助餐加钟
		privateApi.GET("/desk/order/buffet/product/list", wrapper.GetDeskBuffetProductList)                                // 获取自助餐商品列表
		privateApi.POST("/desk/order/product/remark", wrapper.OrderProductRemark)                                          // 桌台订单商品备注
		privateApi.GET("/desk/order/cart/info", wrapper.OrderCartInfo)                                                     // 查询点餐购物车信息
		privateApi.POST("/desk/order/cart/product/add", wrapper.OrderCartProductAdd)                                       // 向购物车添加商品
		privateApi.POST("/desk/order/cart/product_package/add", wrapper.OrderCartProductPackageAdd)                        // 向购物车添加套餐
		privateApi.GET("/desk/order/cart/product/flavor_and_attribute", wrapper.OrderCartProductFlavorAndAttribute)        // 查询购物车商品“规格/属性”
		privateApi.POST("/desk/order/cart/product/flavor_and_attribute", wrapper.OrderCartProductFlavorAndAttributeChange) // 修改购物车商品“规格/属性”
		privateApi.POST("/desk/order/cart/product/num", wrapper.OrderCartProductNum)                                       // 修改购物车商品数量
		privateApi.POST("/desk/order/cart/cooking", wrapper.OrderCartProductCooking)                                       // 送厨购物车商品
		privateApi.POST("/desk/order/cart/product/returning", wrapper.OrderCartProductReturning)                           // 退菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_returning", wrapper.OrderCartProductCancelReturning)              // 取消退菜购物车商品
		privateApi.POST("/desk/order/cart/product/change_desk", wrapper.OrderCartProductChangeDesk)                        // 转菜
		privateApi.POST("/desk/order/cart/product/wrap", wrapper.OrderCartProductWrap)                                     // 打包单商品
		privateApi.POST("/desk/order/cart/product/unwrap", wrapper.OrderCartProductUnwrap)                                 // 取消打包单商品
		privateApi.POST("/desk/order/cart/product/giving", wrapper.OrderCartProductGiving)                                 // 赠菜购物车商品
		privateApi.POST("/desk/order/cart/product/cancel_giving", wrapper.OrderCartProductCancelGiving)                    // 取消赠菜购物车商品
		privateApi.POST("/desk/order/must_plan/confirm", wrapper.OrderMustPlanConfirm)                                     // 确认必点商品
		privateApi.GET("/desk/order/check", wrapper.OrderCheck)                                                            // 订单检查。场景：1、点击结账按钮时，检查订单是否可以结账
		privateApi.GET("/desk/order/payment/info", wrapper.OrderPaymentInfo)                                               // 获取结账页面信息
		privateApi.POST("/desk/order/payment/coupon", wrapper.OrderPaymentCoupon)                                          // 选择或取消优惠券
		privateApi.POST("/desk/order/payment/points", wrapper.OrderPaymentPoints)                                          // 设置订单的抵扣积分数量
		privateApi.GET("/desk/order/payment/qrcode", wrapper.OrderPaymentQrcodeInfo)                                       // 获取支付方式的二维码信息
		privateApi.POST("/desk/order/payment/create", wrapper.OrderPaymentCreate)                                          // 创建一个支付单
		privateApi.POST("/desk/order/payment/cancel", wrapper.OrderPaymentCancel)                                          // 撤销一个支付单
		privateApi.POST("/desk/order/payment/finish", wrapper.OrderPaymentFinish)                                          // 完成销售订单的付款结账
		privateApi.POST("/desk/order/free", wrapper.OrderFree)                                                             // 免单
		privateApi.POST("/desk/order/payment/zero_rule", wrapper.OrderPaymentZeroRule)                                     // 设置结账抹零规则
		privateApi.POST("/desk/order/sale_order/create", wrapper.OrderSaleOrderCreate)                                     // 创建一个销售订单
		privateApi.POST("/desk/order/sale_order/move_product", wrapper.OrderSaleOrderMoveProduct)                          // 从一个销售订单移动商品到另一个销售订单
		privateApi.DELETE("/desk/order/sale_order/delete", wrapper.OrderSaleOrderDelete)                                   // 删除一个销售订单(删除拆单)
		privateApi.DELETE("/desk/order/sale_order/delete_all", wrapper.OrderSaleOrderDeleteAll)                            // 删除所有子销售订单(撤销拆单)
		privateApi.GET("/desk/order/member/discount", wrapper.GetMemberDiscount)                                           // 获取订单会员优惠
		privateApi.POST("/desk/order/member/confirm", wrapper.OrderUseMember)                                              // 确认使用会员优惠并验证密码
		privateApi.DELETE("/desk/order/member/cancel", wrapper.OrderMemberCancel)                                          // 不使用此会员
		privateApi.POST("/desk/order/print", wrapper.OrderPrint)                                                           // 打印小票
		privateApi.POST("/desk/order/print/invoice", wrapper.OrderPrintInvoice)                                            // 打印发票
		privateApi.POST("/desk/order/unlock", wrapper.OrderUnlock)                                                         // 订单解锁
		privateApi.GET("/desk/order/member/list", wrapper.GetOrderMemberList)                                              // 使用会员列表
		privateApi.GET("/desk/order/daily_sales_outbound_summary", wrapper.GetDailySalesOutboundSummary)                   // 获取每日销售出库汇总
		privateApi.GET("/desk/order/headquarter_material_list", wrapper.GetHeadquarterMaterialList)                        // 获取总部物品列表
		privateApi.GET("/desk/order/cart/batch/cooking", wrapper.OrderCartProductBatchCookingList)                         // 获取分批送厨弹框的销售订单商品列表
		privateApi.POST("/desk/order/cart/batch/cooking", wrapper.OrderCartProductBatchCooking)                            // 分批送厨
	}
}

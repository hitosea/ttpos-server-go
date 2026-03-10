package service

import (
	contexts "context"
	builtinerrors "errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/req/member_req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/takeout"
	"ttpos-server-go/config"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/validator"
	"ttpos-server-go/pkg/websocket"

	"github.com/hdt3213/delayqueue"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IMemberOrderSrv interface {
	// 会员端
	CreateMemberOrder(ctx context.Context, req req.CreateMemberOrderReq) (*resp.CreateMemberOrderResp, *resp.OrderCheckServiceRes, error) // 创建会员端外送订单
	CreateMemberDineInOrder(ctx context.Context, req req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error)               // 创建会员端堂食订单
	GetDineInOrderFormInfo(ctx context.Context, req req.GetDineInOrderFormInfoReq) (*resp.DineInOrderFormResp, error)                    // 获取堂食订单提交表单信息
	SetDineInOrderDiningMethod(ctx context.Context, req req.SetDineInOrderDiningMethodReq) error                                         // 设置堂食订单用餐方式
	PayDineInOrder(ctx context.Context, req req.PayDineInOrderReq) error                                                                 // 堂食订单提交支付
	GetDineInOrderPayInfo(ctx context.Context, req req.GetDineInOrderPayInfoReq) (*resp.DineInOrderPaymentInfoResp, error)               // 获取堂食订单支付信息
	GetDineInOrderPayStatus(ctx context.Context, req req.GetDineInOrderPayStatusReq) (*resp.DineInOrderPaymentStatusResp, error)         // 获取堂食订单支付状态
	GetMemberOrderFormInfo(ctx context.Context, req req.GetMemberOrderFormInfoReq) (*resp.CreateMemberOrderResp, error)                  // 获取订单提交表单信息
	SetMemberOrderAddress(ctx context.Context, req member_req.SetMemberOrderAddressReq) (*resp.CreateMemberOrderResp, error)             // 设置会员端订单地址
	PayMemberOrder(ctx context.Context, request member_req.PayMemberOrderReq) error                                                      // 会员端订单提交支付，状态变为待支付
	GetMemberOrderPayInfo(ctx context.Context, request member_req.GetMemberOrderPayInfoReq) (*resp.MemberOrderPaymentInfoResp, error)    // 会员端订单获取支付信息
	GetMemberOrderPayStatus(ctx context.Context, request member_req.GetMemberOrderPayStatusReq) (*resp.MemberOrderPaymentStatusResp, error) // 会员端订单获取支付状态
	GetMemberOrderList(ctx context.Context, req req.MemberOrderListReq) (*resp.GetMemberOrderListResp, error)                                // 查询收银机"外送"页面的订单列表
	GetMemberOrderDetail(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderDetailResp, error)                       // 查询会员端订单详情
	GetMemberOrderPaymentMethodList(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderPaymentMethodListResp, error) // 获取会员端订单支付方式列表
	MemberOrderCancel(ctx context.Context, req member_req.CancelOrderReq) error                                                              // 会员端订单取消
	GetRiderInfo(ctx context.Context, getRiderInfoReq member_req.GetRiderInfoReq) (*resp.MemberOrderCoordinates, error)                      // 获取骑手信息
	SendAuthCode(ctx context.Context, sendCodeReq req.MemberOrderSendAuthCodeReq) error                                                      // 发送认证验证码

	// 外送订单退款相关
	GetMemberOrderReturnInfo(ctx context.Context, req member_req.MemberOrderReturnInfoReq) (*resp.OrderReturnInfoResp, error) // 获取外送订单退款弹窗信息
	MemberOrderReturn(ctx context.Context, req req.OrderReturnReq) (error, int)                                               // 外送订单退款/部分退款
	MemberOrderReReturn(ctx context.Context, req req.OrderReReturnReq) (error, int)                                           // 外送订单重新退款

	// 收银机
	GetMemberCashierOrderList(ctx context.Context, req req.MemberOrderListReq) (*resp.GetMemberCashierOrderListResp, error)              // 查询收银机"外送"页面的订单列表
	GetMemberCashierOrderDetail(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderCashierDetailResp, error)     // 查询收银机"外送"页面的订单详情
	GetMemberOrderManageList(ctx context.Context, req req.MemberOrderManageListReq) (*resp.GetMemberOrderManageListResp, error)          // 查询收银机"外送"管理页面的订单列表
	GetMemberOrderManageDetail(ctx context.Context, req req.GetMemberOrderManageDetailReq) (*resp.GetMemberOrderManageDetailResp, error) // 查询收银机"外送"管理页面的订单详情
	MemberOrderCancelInCashier(ctx context.Context, request member_req.CancelOrderReq) error                                             // 收银端取消订单
	AcceptMemberSaleOrder(ctx context.Context, request req.AcceptOrderReq, options ...AcceptMemberSaleOrderFunc) error                   // 接单外送订单
	RejectMemberSaleOrder(ctx context.Context, req req.RejectOrderReq) error                                                             // 拒单外送订单
	CookFinishMemberSaleOrder(ctx context.Context, request req.CookFinishOrderReq) error                                                 // 备餐完成外送订单
	GetMemberCashierOrderSearch(ctx context.Context, req req.MemberOrderSearchReq) (*resp.GetMemberCashierOrderSearchResp, error)        // 搜索订单列表通过关键词

	// 系统
	MemberOrderSelectingTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error   // 外送订单选购超时自动取消订单
	MemberOrderPayTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error         // 外送订单支付超时自动取消订单
	MemberOrderRiderPickupTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error // 外送订单骑手接单超时自动取消订单

	// 完成外送订单
	CompleteMemberSaleOrder(ctx context.Context, memberSaleOrderUuid uint64) error // 完成外送订单

	// 会员端堂食订单管理
	GetMemberDineInOrderList(ctx context.Context, req req.MemberDineInOrderListReq) (*resp.GetMemberDineInOrderListResp, error)       // 获取会员端堂食订单列表
	GetMemberDineInOrderDetail(ctx context.Context, req req.GetMemberDineInOrderDetailReq) (*resp.GetMemberDineInOrderDetailResp, error) // 获取会员端堂食订单详情
}

func (s *orderSrv) CompleteMemberSaleOrder(ctx context.Context, memberSaleOrderUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)
	memberSaleOrder, err := s.getMemberSaleOrderAllInfo(ctx, memberSaleOrderUuid, nil)
	if err != nil {
		return errors.WithMessage(err)
	}
	saleBill := memberSaleOrder.SaleBill
	saleOrder := saleBill.GetFirstSaleOrder()

	// 发放积分
	// 计算本单获取的积分. 如果订单没有会员，则不计算
	if saleOrder.ConsumerUuid != 0 {
		// 计算积分
		// 根据订单类型（自助餐订单或非自助餐订单）选择积分策略（按比例或按人数）
		pointsRule, err := s.GetPointsRuleInfo(ctx, saleBill.IsBuffetSaleBill(), saleOrder.Member.MemberLevelUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		saleOrder.SetGiftPointsRate(int(saleBill.MealNum), *pointsRule)
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新销售订单
		if err := repository.NewSaleOrderRepo(tx).UpdateSaleOrderRecord(*saleOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 设置sort排序。 ！！！ 注意是修改sort字段为0，gorm默认不修改值为0的字段
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderSort(memberSaleOrderUuid, constant.MemberSaleOrderSortDefault); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	utils.Go(func() {
		s.bus.PublishRiderCompletedMemberSaleOrderEvent(event.RiderCompletedMemberSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				SaleBillUuid:        saleBill.Uuid,
				SaleOrderUuid:       saleOrder.Uuid,
				MemberSaleOrderUuid: memberSaleOrderUuid,
				MemberUuid:          saleOrder.ConsumerUuid,
			},
			MemberSaleOrderUuid: memberSaleOrderUuid,
			SaleBill:            saleBill,
		})
	})
	return nil
}

// CreateMemberOrder 创建会员端订单。实现思路：
// 1. 订单uuid为0时，新建订单
// 2. 订单uuid不为0且订单未取消时，更新订单商品。根据商品签名（规格id-属性id，属性Iid-加料id，加料id），识别出本次订单提交中"删除的订单商品"、"新增的订单商品"、"修改的订单商品"；
// 3. 订单uuid不为0且订单已取消时，新建订单
func (s *orderSrv) CreateMemberOrder(ctx context.Context, req req.CreateMemberOrderReq) (*resp.CreateMemberOrderResp, *resp.OrderCheckServiceRes, error) {
	// 获取数据库
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	if err := req.Validate(); err != nil {
		return nil, nil, errors.WithMessage(err)
	}

	if req.MemberSaleOrderUuid == 0 {
		// 新建订单
		var result *resp.CreateMemberOrderResp
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			ctxCopy := ctx.Copy()
			ctxCopy.SetDB(tx)
			res, err := s.createMemberOrder(ctxCopy, req)
			if err != nil {
				return errors.WithMessage(err)
			}
			result = res
			return nil
		}); err != nil {
			return nil, nil, errors.WithMessage(err)
		}

		// 	发布"创建外送订单"事件
		utils.Go(func() {
			s.bus.PublishCreateMemberSaleOrderEvent(event.CreateMemberSaleOrderPayload{
				BasePayload: event.BasePayload{
					Ctx:                 ctx,
					CompanyUuid:         ctx.GetCompanyUuid(),
					Source:              ctx.GetSource(),
					SaleBillUuid:        result.MemberSaleOrderInfo.SaleBillUuid,
					SaleOrderUuid:       result.MemberSaleOrderInfo.SaleOrderUuid,
					MemberSaleOrderUuid: result.MemberSaleOrderInfo.MemberSaleOrderUuid,
					MemberUuid:          ctx.GetMemberUuid(),
				},
			})
		})

		// 添加选购超时自动取消订单的延时队列任务
		if Queue.MemberOrderCancelQueue != nil {
			utils.Go(func() {
				memberSaleOrderUuidStr := strconv.FormatUint(result.MemberSaleOrderInfo.MemberSaleOrderUuid, 10)
				// 构建队列消息参数
				paramsJson := utils.ToJson(map[string]interface{}{
					"member_sale_order_uuid": result.MemberSaleOrderInfo.MemberSaleOrderUuid,
					"company_uuid":           ctx.GetCompanyUuid(),
					"cancel_scene":           constant.MemberSaleOrderSceneSelectingTimeout,
				})

				// 发送30分钟后自动取消订单的延时消息
				_, err := Queue.MemberOrderCancelQueue.SendDelayMsgV2(
					paramsJson,
					30*time.Minute,               // 30分钟后执行
					delayqueue.WithRetryCount(3), // 重试3次
				)
				if err != nil {
					ctx.Log().Error("添加选购超时自动取消订单任务失败",
						zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr),
						zap.Error(err))
				}
			})
		}

		return result, nil, nil
	} else {
		// 更新订单
		memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(req.MemberSaleOrderUuid)
		if err != nil {
			return nil, nil, errors.WithMessage(err)
		}
		// 判断订单状态
		if memberSaleOrder.IsCancel() {
			// 订单已取消，新建订单
			var result *resp.CreateMemberOrderResp
			if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
				ctxCopy := ctx.Copy()
				ctxCopy.SetDB(tx)
				res, err := s.createMemberOrder(ctxCopy, req)
				if err != nil {
					return errors.WithMessage(err)
				}
				result = res
				return nil
			}); err != nil {
				return nil, nil, errors.WithMessage(err)
			}
			return result, nil, nil
		}
		// 订单未取消，更新订单
		return s.updateMemberOrder(ctx, req, memberSaleOrder)
	}
}

// GetDineInOrderFormInfo 获取堂食订单提交表单信息
// 用于会员端堂食订单的提交页面，显示商品列表、金额信息、支付方式等
func (s *orderSrv) GetDineInOrderFormInfo(ctx context.Context, request req.GetDineInOrderFormInfoReq) (*resp.DineInOrderFormResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "订单不存在")
	}

	// 验证订单类型
	if !saleBill.IsInstantBill() {
		return nil, errors.New("不是堂食订单")
	}

	// 验证订单状态
	if saleBill.IsCanceled() {
		return nil, errors.New("订单已取消")
	}
	if saleBill.IsFinish() {
		return nil, errors.New("订单已完成")
	}

	// 获取指定的销售订单
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 重新计算金额并保存
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
		return nil, errors.WithMessage(err, "计算订单金额失败")
	}

	// 构建商品列表
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	products := make([]resp.DineInProduct, 0)
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		// 跳过已删除的商品
		if saleOrderProduct.DeleteTime != 0 {
			continue
		}
		products = append(products, resp.DineInProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.GetLocaleName(),
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(), // 包含规格+小料+属性
			Num:                  saleOrderProduct.Num,
			UnitPrice:            saleOrderProduct.OriginTotalPrice, // 折前单价，含税费
			Amount:               saleOrderProduct.GetTotalPriceOrigin(),
			Image: func() string {
				if saleOrderProduct.ImageFileUuid != 0 && saleOrderProduct.ImageFile != nil {
					return saleOrderProduct.ImageFile.GetUrl(baseUrl)
				}
				return ""
			}(),
		})
	}

	// 构建金额信息
	amountInfo := resp.DineInAmountInfo{
		ProductAmount:  saleOrder.ProductAmount,    // 商品金额
		TaxAmount:      saleOrder.TaxFee,           // 消费税
		ServiceAmount:  saleOrder.ServiceFee,       // 服务费
		MemberDiscount: saleOrder.MemberDiscountFee, // 会员折扣
		TotalAmount:    saleOrder.Amount,           // 应付金额（已包含税费和服务费）
	}

	// 获取支付方式列表
	paymentMethods, err := repository.NewPaymentMethodRepo(db).GetLianLianPayPaymentMethodList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取支付方式失败")
	}
	payList := make([]resp.PaymentMethodItem, 0)
	for _, paymentMethod := range paymentMethods {
		payList = append(payList, resp.PaymentMethodItem{
			Source:        paymentMethod.Source,
			SourceText:    constant.PaymentMethodSourceTextMap[paymentMethod.Source],
			Uuid:          paymentMethod.Uuid,
			PaymentName:   paymentMethod.GetPaymentName(),
			PaymentMethod: paymentMethod.GetName(),
			FeePercent:    paymentMethod.FeePercent,
			Logo: func() string {
				if paymentMethod.IsWechatPay() {
					return baseUrl + "/image/pay/wechat_pay.png"
				}
				if paymentMethod.IsAliPay() {
					return baseUrl + "/image/pay/alipay.png"
				}
				if paymentMethod.IsQrPromptPay() {
					return baseUrl + "/image/pay/qr_prompt_pay.png"
				}
				return ""
			}(),
			Code: paymentMethod.Code,
		})
	}

	return &resp.DineInOrderFormResp{
		SaleBillUuid:   saleBill.Uuid,
		SaleOrderUuid:  saleOrder.Uuid,
		DiningMethod:   saleBill.DiningMethod,
		ProductList:    resp.DineInProductList{List: products},
		AmountInfo:     amountInfo,
		PaymentMethods: resp.PaymentMethodList{List: payList},
		Remark:         saleBill.Remark,
	}, nil
}

// SetDineInOrderDiningMethod 设置堂食订单用餐方式
func (s *orderSrv) SetDineInOrderDiningMethod(ctx context.Context, request req.SetDineInOrderDiningMethodReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "订单不存在")
	}

	// 验证订单类型
	if !saleBill.IsInstantBill() {
		return errors.New("不是堂食订单")
	}

	// 验证订单状态
	if saleBill.IsCanceled() {
		return errors.New("订单已取消")
	}
	if saleBill.IsFinish() {
		return errors.New("订单已完成")
	}

	// 更新用餐方式
	saleBill.DiningMethod = request.DiningMethod

	// 保存更新
	if err := repository.NewSaleBillRepo(db).UpdateSaleBill(saleBill); err != nil {
		return errors.WithMessage(err, "更新订单失败")
	}

	return nil
}

// PayDineInOrder 堂食订单提交支付
func (s *orderSrv) PayDineInOrder(ctx context.Context, request req.PayDineInOrderReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "订单不存在")
	}

	// 验证订单类型
	if !saleBill.IsInstantBill() {
		return errors.New("不是堂食订单")
	}

	// 验证订单状态
	if saleBill.IsCanceled() {
		return errors.New("订单已取消")
	}
	if saleBill.IsFinish() {
		return errors.New("订单已完成")
	}

	// 获取指定的销售订单
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return errors.New("销售订单不存在")
	}

	// 验证订单金额
	if saleOrder.Amount < 1 {
		return errors.New("订单金额小于1，无法支付")
	}

	// 验证支付方式
	if request.PaymentMethodUuid == 0 {
		return errors.New("请选择支付方式")
	}

	// 更新订单备注
	if request.Remark != "" {
		saleBill.Remark = request.Remark
		// 保存更新
		if err := repository.NewSaleBillRepo(db).UpdateSaleBill(saleBill); err != nil {
			return errors.WithMessage(err, "更新订单失败")
		}
	}

	return nil
}

// GetDineInOrderPayInfo 获取堂食订单支付信息
func (s *orderSrv) GetDineInOrderPayInfo(ctx context.Context, request req.GetDineInOrderPayInfoReq) (*resp.DineInOrderPaymentInfoResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "订单不存在")
	}

	// 验证订单类型
	if !saleBill.IsInstantBill() {
		return nil, errors.New("不是堂食订单")
	}

	// 获取指定的销售订单
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 验证订单金额
	if saleOrder.Amount < 1 {
		return nil, errors.NewWithCode(constant.CodeOrderAmountLessThan1, "订单金额小于1，无法支付")
	}

	// 获取支付方式
	var paymentMethod *model.PaymentMethod
	if request.PaymentMethodUuid != 0 {
		paymentMethod, err = repository.NewPaymentMethodRepo(db).GetPaymentMethodByUuid(request.PaymentMethodUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if paymentMethod.Uuid == 0 {
			return nil, errors.New("支付方式不存在")
		}
	} else {
		return nil, errors.New("请选择支付方式")
	}

	// 验证支付方式是否可用
	if !paymentMethod.IsLianLianPay() {
		return nil, errors.New("支付方式不可用")
	}

	// 关联UUID使用销售订单UUID
	relatedUuid := saleOrder.Uuid

	// 查询或创建支付订单
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)
	paymentOrder, err := paymentOrderRepo.GetPaymentOrderInfo(
		repository.CommonRepo.WhereBySoftDelete(),
		paymentOrderRepo.WhereRelatedUuid(relatedUuid),
		paymentOrderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
		paymentOrderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
	)
	if err == nil && paymentOrder.Uuid != 0 {
		// 已有支付订单且已支付
		if paymentOrder.Status == constant.PaymentOrderStatusPaid {
			return &resp.DineInOrderPaymentInfoResp{
				SaleBillUuid:      saleBill.Uuid,
				SaleOrderUuid:     saleOrder.Uuid,
				PaymentOrderUuid:  paymentOrder.Uuid,
				PaymentMethodName: paymentMethod.PaymentName,
				Status:            paymentOrder.Status,
				PaymentAmount:     paymentOrder.PaymentAmount,
			}, nil
		}
	} else {
		// 创建支付订单
		createPaymentOrder, err := paymentOrderRepo.Create(model.PaymentOrder{
			PaymentMethodName: paymentMethod.PaymentName,
			PaymentMethodUuid: paymentMethod.Uuid,
			PaymentFeePercent: 0,
			RelatedType:       constant.PaymentOrderRelatedTypeSaleOrder,
			RelatedUuid:       relatedUuid,
			CurrencyUnit: func() string {
				currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
				if err != nil {
					return "THB"
				}
				return currencySetting.Unit
			}(),
			PaymentAmount:        saleOrder.Amount,
			PaymentCommissionFee: 0,
			Amount:               saleOrder.Amount,
			Status:               constant.PaymentOrderStatusUnPay,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "创建支付订单失败")
		}
		paymentOrder = &createPaymentOrder
	}

	// 判断当前是手机端还是PC端
	isOpenPc := !ctx.IsMobile() && viper.GetString("OPEN_MEMBER_PC_PAY") == "true"

	// 创建连连支付订单
	payment, err := NewPaymentRepo(ctx, s.dbm).CreatePayment(CreatePaymentReq{
		PaymentOrderUuid:  paymentOrder.Uuid,
		RelatedType:       constant.PaymentOrderRelatedTypeSaleOrder,
		RelatedUuid:       relatedUuid,
		PaymentMethodUuid: paymentMethod.Uuid,
		PaymentMethodCode: paymentMethod.Code,
		PaymentAmount:     saleOrder.Amount,
		CommissionFee:     0,
		PaymentMethod: func() string {
			if isOpenPc {
				return PaymentMethodWechatPay
			}
			return PaymentMethodH5Payment
		}(),
		RedirectUrl: func() string {
			if ctx.IsMobile() && paymentMethod.IsWechatPay() {
				return fmt.Sprintf("%s/company/%d/dine_in/payment_result/%d/", config.Server.MemberBaseUrl, ctx.GetCompanyUuid(), saleBill.Uuid)
			}
			return ""
		}(),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.DineInOrderPaymentInfoResp{
		SaleBillUuid:      saleBill.Uuid,
		SaleOrderUuid:     saleOrder.Uuid,
		PaymentOrderUuid:  payment.PaymentOrderUuid,
		PaymentMethodName: paymentMethod.PaymentName,
		IsWechatPay:       paymentMethod.IsWechatPay(),
		QrCode:            utils.IfString(isOpenPc || paymentMethod.IsQrPromptPay(), payment.LinkUrl, ""),
		LinkUrl:           utils.IfString(isOpenPc || paymentMethod.IsQrPromptPay(), "", payment.LinkUrl),
		Status:            payment.GetStatus(),
		PaymentAmount:     payment.OrderAmount,
	}, nil
}

// GetDineInOrderPayStatus 获取堂食订单支付状态
func (s *orderSrv) GetDineInOrderPayStatus(ctx context.Context, request req.GetDineInOrderPayStatusReq) (*resp.DineInOrderPaymentStatusResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	ctx.SetDB(db)

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "订单不存在")
	}

	// 获取指定的销售订单
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}

	// 获取支付状态
	status := uint(0)
	if saleOrder.Status == constant.SaleOrderStatusFinish {
		status = 1
	}

	return &resp.DineInOrderPaymentStatusResp{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: saleOrder.Uuid,
		Status:        status,
	}, nil
}

// GetMemberOrderFormInfo 获取订单提交表单信息
func (s *orderSrv) GetMemberOrderFormInfo(ctx context.Context, request req.GetMemberOrderFormInfoReq) (*resp.CreateMemberOrderResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	info, err := s.GetMemberOrderCheckoutInfo(ctx, req.GetMemberOrderCheckoutInfoReq{
		MemberSaleOrderUuid: request.MemberSaleOrderUuid,
	}, nil)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return info, nil
}

// SetMemberOrderAddress 设置会员端订单地址
func (s *orderSrv) SetMemberOrderAddress(ctx context.Context, request member_req.SetMemberOrderAddressReq) (*resp.CreateMemberOrderResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberAddress, err := repository.NewMemberAddressRepo(db).GetMemberAddressByUuid(request.MemberAddressUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	member := ctx.GetMember()
	if member.Uuid == 0 {
		return nil, errors.New("无法找到会员")
	}
	if memberAddress.MemberUuid != member.Uuid {
		return nil, errors.New("地址与会员不匹配")
	}

	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecordOnly(request.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if memberSaleOrder.MemberUuid != member.Uuid {
		return nil, errors.New("订单与会员不匹配")
	}
	addressChanged := memberSaleOrder.MemberAddressUuid != request.MemberAddressUuid

	memberSaleOrder.SetMemberAddress(*memberAddress)
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 返回会员端订单信息
	info, err := s.GetMemberOrderCheckoutInfo(ctx, req.GetMemberOrderCheckoutInfoReq{
		MemberSaleOrderUuid: request.MemberSaleOrderUuid,
		AddressChanged:      addressChanged,
	}, nil)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}

// PayMemberOrder 提交支付
func (s *orderSrv) PayMemberOrder(ctx context.Context, request member_req.PayMemberOrderReq) error {
	db := ctx.GetDB()

	// 保存订单备注信息
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	if memberSaleOrder.Amount < 1 {
		return errors.New("订单金额小于1，无法支付")
	}

	// 如果未设置地址，返回错误
	if memberSaleOrder.MemberAddressUuid == 0 {
		return errors.New("请先选择订单地址")
	}

	// 如果未选择支付方式，返回错误
	if request.PaymentMethodUuid == 0 {
		return errors.New("请选择支付方式")
	}

	companySetting := ctx.GetCompanySetting()
	deliveryConfig, err := companySetting.GetDeliveryConfig(constant.ProviderNameSkootar, memberSaleOrder.DeliveryDistance)
	if err != nil {
		return errors.WithMessage(err, "配送费配置失败")
	}
	if !deliveryConfig.IsInDeliveryRange {
		return errors.NewWithCode(constant.CodeOrderAddressNotInDeliveryRange, "订单地址不在配送范围内")
	}

	// 如果订单未计算距离费，则查询距离并计算距离费
	if !memberSaleOrder.GetIsDistanceCalculated() {
		// 查询距离
		distance, err := s.QueryDistance(ctx, memberSaleOrder)
		if err != nil {
			return errors.WithMessage(err)
		}
		memberSaleOrder.SetDeliveryDistance(distance)
		if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
			return errors.WithMessage(err)
		}
	}

	memberSaleOrder.Remark = request.Remark
	if err := memberSaleOrder.SetPendingPayment(request.PaymentMethodUuid); err != nil {
		return errors.WithMessage(err)
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 只有第一次提交支付时才生成订单流水号并记录提交支付时间
		if memberSaleOrder.SubmitPayTime == 0 {
			memberSaleOrder.SubmitPayTime = time.Now().Unix() // 记录提交支付时间戳
			// 创建会员端订单流水号
			timezone := ctx.GetCompanySetting().Timezone
			serialNo, err := s.createMemberSaleOrderSerialNo(db, timezone)
			if err != nil {
				ctx.Log().Error("会员端订单流水号生成失败", zap.Error(err))
				return errors.WithMessage(err, "会员端订单流水号生成失败")
			}
			memberSaleOrder.SerialNumber = serialNo
		}

		// 更新sale_bill的serial_no
		if err := repository.NewSaleBillRepo(db).UpdateSaleBillSerialNo(memberSaleOrder.SaleBillUuid, memberSaleOrder.SerialNumber); err != nil {
			return errors.WithMessage(err)
		}

		if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderPendingPayment(memberSaleOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 添加24小时后自动取消订单的延时队列任务
	if Queue.MemberOrderCancelQueue != nil {
		utils.Go(func() {
			memberSaleOrderUuidStr := strconv.FormatUint(request.MemberSaleOrderUuid, 10)
			// 构建队列消息参数
			paramsJson := utils.ToJson(map[string]interface{}{
				"member_sale_order_uuid": request.MemberSaleOrderUuid,
				"company_uuid":           ctx.GetCompanyUuid(),
				"cancel_scene":           constant.MemberSaleOrderScenePaymentTimeout,
			})

			// 发送24小时后自动取消订单的延时消息
			_, err := Queue.MemberOrderCancelQueue.SendDelayMsgV2(
				paramsJson,
				time.Duration(config.Server.PaymentTimeout)*time.Second, // 24小时后执行
				delayqueue.WithRetryCount(3),                            // 重试3次
			)
			if err != nil {
				ctx.Log().Error("添加24小时自动取消订单任务失败",
					zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr),
					zap.Error(err))
			}
		})
	}

	return nil
}

// GetMemberOrderPayInfo 查询支付信息
func (s *orderSrv) GetMemberOrderPayInfo(ctx context.Context, request member_req.GetMemberOrderPayInfoReq) (*resp.MemberOrderPaymentInfoResp, error) {
	db := ctx.GetDB()

	// 订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(request.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 判断订单金额是否为0
	if memberSaleOrder.Amount < 1 {
		return nil, errors.NewWithCode(constant.CodeOrderAmountLessThan1, "订单金额小于1，无法支付")
	}
	// 判断订单是否可以支付
	if memberSaleOrder.Status < constant.MemberSaleOrderStatusPendingPayment {
		return nil, errors.New("订单状态不可支付")
	}

	// 判断当前是否连连支付
	var paymentMethod *model.PaymentMethod
	if request.PaymentMethodUuid != 0 {
		paymentMethod, err = repository.NewPaymentMethodRepo(db).GetPaymentMethodByUuid(request.PaymentMethodUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		if paymentMethod.Uuid == 0 {
			return nil, errors.New("支付方式不存在")
		}
		if err := memberSaleOrder.SetPendingPayment(paymentMethod.Uuid); err != nil {
			return nil, errors.WithMessage(err)
		}
		if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderPendingPayment(memberSaleOrder); err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		paymentMethod = memberSaleOrder.PaymentMethod
		if paymentMethod == nil || paymentMethod.Uuid == 0 {
			return nil, errors.New("支付方式不存在")
		}
	}

	// 支付方式是否可用
	if !paymentMethod.IsLianLianPay() {
		return nil, errors.New("支付方式不可用")
	}

	// 获取销售订单关联UUId
	relatedUuid := memberSaleOrder.SaleBill.GetFirstSaleOrder().Uuid

	// 判断支付方式是否已支付
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)
	paymentOrder, err := paymentOrderRepo.GetPaymentOrderInfo(
		repository.CommonRepo.WhereBySoftDelete(),
		paymentOrderRepo.WhereRelatedUuid(relatedUuid),
		paymentOrderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
		paymentOrderRepo.WherePaymentMethodUuid(paymentMethod.Uuid),
	)
	if err == nil && paymentOrder.Uuid != 0 {
		if paymentOrder.Status == constant.PaymentOrderStatusPaid {
			return &resp.MemberOrderPaymentInfoResp{
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
				PaymentOrderUuid:    paymentOrder.Uuid,
				PaymentMethodName:   paymentMethod.PaymentName,
				Status:              paymentOrder.Status,
				PaymentAmount:       paymentOrder.PaymentAmount,
			}, nil
		}
	} else {
		// 添加支付方式
		createPaymentOrder, err := paymentOrderRepo.Create(model.PaymentOrder{
			PaymentMethodName: paymentMethod.PaymentName,
			PaymentMethodUuid: paymentMethod.Uuid,
			PaymentFeePercent: 0,
			RelatedType:       constant.PaymentOrderRelatedTypeSaleOrder,
			RelatedUuid:       relatedUuid,
			CurrencyUnit: func() string {
				currencySetting, err := s.settingSrv.GetCurrencySetting(ctx)
				if err != nil {
					return "THB"
				}
				return currencySetting.Unit
			}(),
			PaymentAmount:        memberSaleOrder.Amount,
			PaymentCommissionFee: 0,
			Amount:               memberSaleOrder.Amount,
			Status:               constant.PaymentOrderStatusUnPay,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "添加支付方式失败")
		}
		//
		paymentOrder = &createPaymentOrder
	}

	// 判断当前手机端访问还是PC端访问
	isOpenPc := !ctx.IsMobile() && viper.GetString("OPEN_MEMBER_PC_PAY") == "true"

	// 创建连连支付订单
	payment, err := NewPaymentRepo(ctx, s.dbm).CreatePayment(CreatePaymentReq{
		PaymentOrderUuid:    paymentOrder.Uuid,
		RelatedType:         constant.PaymentOrderRelatedTypeSaleOrder,
		RelatedUuid:         relatedUuid,
		PaymentMethodUuid:   paymentMethod.Uuid,
		PaymentMethodCode:   paymentMethod.Code,
		PaymentAmount:       memberSaleOrder.Amount,
		CommissionFee:       0,
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		PaymentMethod: func() string {
			if isOpenPc {
				return PaymentMethodWechatPay
			}
			return PaymentMethodH5Payment
		}(),
		RedirectUrl: func() string {
			if ctx.IsMobile() && paymentMethod.IsWechatPay() {
				return fmt.Sprintf("%s/company/%d/payment_result/%d/", config.Server.MemberBaseUrl, ctx.GetCompanyUuid(), memberSaleOrder.Uuid)
			}
			return ""
		}(),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.MemberOrderPaymentInfoResp{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		PaymentOrderUuid:    payment.PaymentOrderUuid,
		PaymentMethodName:   paymentMethod.PaymentName,
		IsWechatPay:         paymentMethod.IsWechatPay(),
		QrCode:              utils.IfString(isOpenPc || paymentMethod.IsQrPromptPay(), payment.LinkUrl, ""),
		LinkUrl:             utils.IfString(isOpenPc || paymentMethod.IsQrPromptPay(), "", payment.LinkUrl),
		Status:              payment.GetStatus(), // 支付单状态 支付状态, 0-未支付 1-已支付 (可选择轮询当前接口，获取支付状态)
		PaymentAmount:       payment.OrderAmount, // 支付金额
	}, nil
}

// GetMemberOrderPayStatus 获取支付状态
func (s *orderSrv) GetMemberOrderPayStatus(ctx context.Context, request member_req.GetMemberOrderPayStatusReq) (*resp.MemberOrderPaymentStatusResp, error) {
	db := ctx.GetDB()
	// 订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(request.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 判断订单是否已经支付
	return &resp.MemberOrderPaymentStatusResp{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		Status:              memberSaleOrder.GetPayStatus(),
	}, nil
}

// GetMemberOrderList 获取"外送"订单列表
func (s *orderSrv) GetMemberOrderList(ctx context.Context, req req.MemberOrderListReq) (*resp.GetMemberOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	memberSaleOrderRepo := repository.NewMemberSaleOrderRepo(db)
	memberSaleOrders, total, err := memberSaleOrderRepo.PaginateGetMemberSaleOrder(
		req.PageNo, req.PageSize,
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.WhereByNoSelectingTimeout(), // 不包含选购超时订单
		repository.CommonRepo.SortWithSort("desc"),
		repository.CommonRepo.SortWithSubmitPayTime("desc"),
		memberSaleOrderRepo.WithSaleBillSaleOrderProduct(),
		repository.CommonRepo.WhereByMemberUuid(ctx.GetMemberUuid()),
		memberSaleOrderRepo.WhereNotStatusIn([]uint{constant.MemberSaleOrderStatusSelecting}),
		memberSaleOrderRepo.WhereStatusIn(constant.GetMemberOrderStatusList(req.Status)),
		memberSaleOrderRepo.WhereKeyword(req.Keyword, ctx.GetLanguage()),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	memberOrders := make([]resp.MemberOrder, 0)
	for _, memberSaleOrder := range memberSaleOrders {
		// 支付超时自动取消订单
		if memberSaleOrder.GetRemainingPaymentTime() == 0 && memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingPayment {
			s.MemberOrderPayTimeoutAutoCancel(ctx, memberSaleOrder.Uuid)
			memberSaleOrder.Status = constant.MemberSaleOrderStatusCancelled
		}
		//
		memberOrders = append(memberOrders, resp.MemberOrder{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			CompanyName:         ctx.GetCompany().Name,
			SerialNumber:        memberSaleOrder.SerialNumber,
			Status:              memberSaleOrder.Status,
			Num:                 memberSaleOrder.ProductNum,
			Amount:              memberSaleOrder.Amount,
			ProductAmount:       memberSaleOrder.OriginProductAmount,
			Rider: resp.RiderInfo{
				Name:              memberSaleOrder.RiderName,
				Phone:             memberSaleOrder.RiderPhone,
				Location:          memberSaleOrder.Location,
				RemainingDistance: memberSaleOrder.RemainingDistance,
			},
			ProductList: func() []resp.MemberOrderProduct {
				products := make([]resp.MemberOrderProduct, 0)
				for _, saleOrderProduct := range memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts {
					products = append(products, resp.MemberOrderProduct{
						LocaleName: saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
						Num:        saleOrderProduct.Num,
						TotalPrice: saleOrderProduct.GetTotalPrice(),
						Image: func() string {
							if saleOrderProduct.ImageFile == nil {
								return ""
							}
							return saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
						}(),
					})
				}
				return products
			}(),
		})
	}

	return &resp.GetMemberOrderListResp{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
		List: memberOrders,
	}, nil
}

// GetMemberOrderDetail 获取会员端订单详情
func (s *orderSrv) GetMemberOrderDetail(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrder, err := getMemberOrderDetail(ctx, req.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	//
	products := make([]resp.MemberOrderProduct, 0)
	for _, saleOrderProduct := range memberSaleOrder.SaleBill.GetFirstSaleOrder().SaleOrderProducts {
		products = append(products, resp.MemberOrderProduct{
			LocaleName:          saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
			OriginTotalPrice:    saleOrderProduct.GetTotalPriceOrigin(),
			Image: func() string {
				if saleOrderProduct.ImageFile == nil {
					return ""
				}
				return saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
			}(),
		})
	}
	//
	var address resp.MemberOrderDetailAddress
	if memberSaleOrder.Address != nil {
		companySetting := ctx.GetCompanySetting()
		deliveryConfig, err := companySetting.GetDeliveryConfig(constant.ProviderNameSkootar, memberSaleOrder.DeliveryDistance)
		if err != nil {
			return nil, errors.WithMessage(err, "配送费配置失败")
		}
		// 如果距离为0，则认为不在配送范围内
		if memberSaleOrder.DeliveryDistance == 0 {
			deliveryConfig.IsInDeliveryRange = false
		}
		address = resp.MemberOrderDetailAddress{
			ContactName:       memberSaleOrder.ContactName,
			Phone:             memberSaleOrder.ContactPhone,
			PhonePrefix:       memberSaleOrder.ContactPhonePrefix,
			Address:           memberSaleOrder.ContactAddress + memberSaleOrder.ContactAddressDetail,
			IsInDeliveryRange: deliveryConfig.IsInDeliveryRange,
		}
	}
	//
	return &resp.GetMemberOrderDetailResp{
		MemberSaleOrderUuid:  memberSaleOrder.Uuid,
		OrderNo:              memberSaleOrder.OrderNo,
		CompanyName:          ctx.GetCompany().Name,
		PayTime:              memberSaleOrder.PayTime,
		FinishTime:           memberSaleOrder.FinishTime,
		CancelTime:           memberSaleOrder.CancelTime,
		RemainingPaymentTime: memberSaleOrder.GetRemainingPaymentTime(),
		CancelReason: func() string {
			if memberSaleOrder.IsSelfCancel() {
				return i18n.Translate(ctx.GetLanguage(), "自主取消") + " (" + memberSaleOrder.CancelReason + ")"
			}
			if memberSaleOrder.IsMerchantCancel() {
				return i18n.Translate(ctx.GetLanguage(), "商家取消") + " (" + memberSaleOrder.CancelReason + ")"
			}
			return i18n.Translate(ctx.GetLanguage(), memberSaleOrder.CancelReason)
		}(),
		CreateTime: memberSaleOrder.CreateTime,
		Status:     memberSaleOrder.Status,
		Remark:     memberSaleOrder.Remark,
		AmountInfo: resp.MemberOrderAmountInfo{
			Amount:            memberSaleOrder.Amount,
			MemberDiscountFee: memberSaleOrder.MemberDiscountFee,
		},
		ProductList: resp.MemberProductList{
			List:          products,
			ProductAmount: memberSaleOrder.OriginProductAmount,
		},
		AddressInfo: address,
		DeliveryConfig: resp.DeliveryResp{
			DeliveryDistance:   memberSaleOrder.DeliveryDistance,
			DeliveryFeeAmount:  memberSaleOrder.DeliveryFeeAmount,
			DeliveryFeeMinFee:  memberSaleOrder.DeliveryFeeMinFee,
			DeliveryFeeBaseFee: memberSaleOrder.DeliveryFeeBaseFee,
			DeliveryFeePerKm:   memberSaleOrder.DeliveryFeePerKm,
		},
		Rider: resp.RiderInfo{
			Name:              memberSaleOrder.RiderName,
			Phone:             memberSaleOrder.RiderPhone,
			Location:          memberSaleOrder.Location,
			RemainingDistance: memberSaleOrder.RemainingDistance,
			EstimatedTime:     memberSaleOrder.ExpectedFinishTime,
		},
	}, nil
}

// GetMemberOrderPaymentMethodList 获取会员端订单支付方式列表
func (s *orderSrv) GetMemberOrderPaymentMethodList(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderPaymentMethodListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(req.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 判断订单是否可支付
	if !memberSaleOrder.IsCanPaid() {
		return nil, errors.New("订单状态不可支付")
	}
	// 判断订单是否支付超时
	if memberSaleOrder.GetRemainingPaymentTime() == 0 {
		// 支付超时自动取消订单
		err := s.MemberOrderPayTimeoutAutoCancel(ctx, memberSaleOrder.Uuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		// 订单支付超时
		return nil, errors.New("订单支付超时")
	}

	// 获取支付方式
	paymentMethods, err := repository.NewPaymentMethodRepo(db).GetLianLianPayPaymentMethodList()
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	payList := make([]resp.PaymentMethodItem, 0)
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	for _, paymentMethod := range paymentMethods {
		payList = append(payList, resp.PaymentMethodItem{
			Source:        paymentMethod.Source,
			SourceText:    constant.PaymentMethodSourceTextMap[paymentMethod.Source],
			Uuid:          paymentMethod.Uuid,
			PaymentName:   paymentMethod.GetPaymentName(),
			PaymentMethod: paymentMethod.GetName(),
			FeePercent:    paymentMethod.FeePercent,
			Logo: func() string {
				if paymentMethod.IsWechatPay() {
					return baseUrl + "/image/pay/wechat_pay.png"
				}
				if paymentMethod.IsAliPay() {
					return baseUrl + "/image/pay/alipay.png"
				}
				if paymentMethod.IsQrPromptPay() {
					return baseUrl + "/image/pay/qr_prompt_pay.png"
				}
				return ""
			}(),
			Code: paymentMethod.Code,
		})
	}

	return &resp.GetMemberOrderPaymentMethodListResp{
		List:                 payList, // 支付方式列表
		RemainingPaymentTime: memberSaleOrder.GetRemainingPaymentTime(),
		Amount:               memberSaleOrder.Amount,
		PaymentMethodUuid:    memberSaleOrder.PaymentMethodUuid,
	}, nil
}

// MemberOrderCancel 会员端订单取消
func (s *orderSrv) MemberOrderCancel(ctx context.Context, request member_req.CancelOrderReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(request.MemberSaleOrderUuid)
		defer lock.NewSystemLock().UnlockUuid(request.MemberSaleOrderUuid)
		ctx.AddLock()
	}

	// 获取DB
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if !memberSaleOrder.IsCanCancel() {
		return errors.New("订单状态不可取消")
	}

	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(memberSaleOrder.SaleBill.Uuid)
		defer lock.NewSystemLock().UnlockUuid(memberSaleOrder.SaleBill.Uuid)
		ctx.AddLock()
	}

	// 获取销售账单Uuid
	saleBillUuid := memberSaleOrder.SaleBillUuid

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果发生恐慌，回滚事务
		}
	}()

	// 设置DB
	ctx.SetDB(tx)

	// 获取订单信息
	orderRepo := repository.NewOrderRepo(tx)
	billInfo, err := orderRepo.GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 获取销售订单
	saleOrder := billInfo.GetFirstSaleOrder()

	// 退回商品库存
	if err := s.returnInventory(ctx, billInfo); err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 取消订单
	err = orderRepo.CancelOrder(ctx, saleBillUuid, 0, request.CancelReason)
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 标记送厨单、送厨商品为删除
	productionRepo := repository.NewProductionRepo(tx)
	saleBillUuidOpt := productionRepo.WhereSaleBillUuid(billInfo.Uuid)
	err = productionRepo.UpdateOrder([]repository.DBOption{saleBillUuidOpt}, map[string]any{
		"delete_time": time.Now().Unix(),
	})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单失败"), err.Error())
	}

	// 修改送厨商品数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
	err = productionRepo.UpdateProduct([]repository.DBOption{saleBillUuidOpt}, map[string]any{"num": 0})
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除送厨单商品失败"), err.Error())
	}

	// 已经支付的-发起退款
	var returnOrder *model.ReturnOrder
	if memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingMerchantAccept {
		returnOrder, err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*saleOrder, MemberSaleOrderRefundReq{
			CancelReason: "客户取消订单",
		})
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
		// 退款金额
		memberSaleOrder.RefundAmount = returnOrder.RefundAmount
	}

	// 设置订单为“已取消”状态
	memberSaleOrder.SetCancel(request.CancelReason)

	// 更新订单状态
	if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		tx.Rollback()
		return errors.WithMessage(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithMessage(err)
	}

	// 发布“订单取消”操作事件
	utils.Go(func() {
		s.bus.PublishCancelMemberOrderEvent(event.CancelMemberOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				Source:              ctx.GetSource(),
				SaleBillUuid:        billInfo.Uuid,
				SaleOrderUuid:       saleOrder.Uuid,
				OperatorUuid:        0,
				MemberUuid:          memberSaleOrder.MemberUuid,
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
			},
			Data: event.CancelMemberOrderPayloadData{
				Type: "user_cancel",
				Refunds: func() []event.CancelMemberOrderPayloadDataRefund {
					refunds := make([]event.CancelMemberOrderPayloadDataRefund, 0)
					if returnOrder != nil {
						for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
							refunds = append(refunds, event.CancelMemberOrderPayloadDataRefund{
								Name:              returnOrderAmount.PaymentMethod.PaymentName,
								Code:              returnOrderAmount.PaymentMethod.Code,
								Amount:            returnOrderAmount.Amount,
								RefundStatus:      returnOrderAmount.RefundStatus,
								ReturnAmountUuid:  returnOrderAmount.Uuid,
								ReturnOrderUuid:   returnOrder.Uuid,
								PaymentOrderUuid:  returnOrderAmount.PaymentOrderUuid,
								PaymentMethodUuid: returnOrderAmount.PaymentMethodUuid,
							})
						}
					}
					return refunds
				}(),
			},
		})
	})

	// 发送短信通知
	utils.Go(func() {
		NewSMSSrv(s.dbm).SendDeliveryOrderCancelSMS(ctx, memberSaleOrder.ContactPhone, &sms.DeliveryOrderCancel{
			CancelScene: sms.TemplateDeliveryOrderCanceledBySelf,
			Company:     ctx.GetCompany().Name,
			OrderNo:     memberSaleOrder.OrderNo,
		})
	})

	// 成功后，推送到厨显端更新订单
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
			"update_time": time.Now().Unix(),
		})
	})

	//
	return nil
}

// MemberOrderCancelInCashier 收银端取消订单
// 收银端取消订单，退回商品库存
// 标记MemberSaleOrder为已取消
func (s *orderSrv) MemberOrderCancelInCashier(ctx context.Context, request member_req.CancelOrderReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(request.MemberSaleOrderUuid)
		defer s.lock.UnlockUuid(request.MemberSaleOrderUuid)
		ctx.AddLock()
	}

	// 获取DB
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	if !memberSaleOrder.IsCanCancelInCashier() {
		return errors.New("订单状态不可取消")
	}

	// 禁止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(memberSaleOrder.SaleBill.Uuid)
		defer s.lock.UnlockUuid(memberSaleOrder.SaleBill.Uuid)
		ctx.AddLock()
	}

	// 获取销售账单Uuid
	saleBillUuid := memberSaleOrder.SaleBill.Uuid
	// 获取订单信息
	orderRepo := repository.NewOrderRepo(db)
	billInfo, err := orderRepo.GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 获取销售订单
	saleOrder := billInfo.GetFirstSaleOrder()

	var returnOrder *model.ReturnOrder

	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 退回商品库存
		ctxCopy := ctx.Copy()
		ctxCopy.SetDB(tx) // 确保 returnInventory 使用事务
		if err := s.returnInventory(ctxCopy, billInfo); err != nil {
			return errors.WithMessage(err)
		}

		// 取消订单
		err = orderRepo.CancelOrder(ctxCopy, saleBillUuid, 0, request.CancelReason)
		if err != nil {
			return errors.WithMessage(err)
		}

		// 标记送厨单、送厨商品为删除
		productionRepo := repository.NewProductionRepo(tx)
		saleBillUuidOpt := productionRepo.WhereSaleBillUuid(billInfo.Uuid)
		err = productionRepo.UpdateOrder([]repository.DBOption{saleBillUuidOpt}, map[string]any{
			"delete_time": time.Now().Unix(),
		})
		if err != nil {
			return errors.WithMessage(builtinerrors.New("删除送厨单失败"), err.Error())
		}

		// 修改送厨商品数量为0，在确认整单退菜时、确认该菜品全退时，再标记为删除
		err = productionRepo.UpdateProduct([]repository.DBOption{saleBillUuidOpt}, map[string]any{"num": 0})
		if err != nil {
			return errors.WithMessage(builtinerrors.New("删除送厨单商品失败"), err.Error())
		}

		// 已经支付的-发起退款
		if memberSaleOrder.Status > constant.MemberSaleOrderStatusPendingPayment {
			returnOrder, err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*saleOrder, MemberSaleOrderRefundReq{
				CancelReason: "客户取消订单",
			})
			if err != nil {
				return errors.WithMessage(err)
			}
		}

		// 设置订单为“已取消”状态
		memberSaleOrder.RefundAmount = memberSaleOrder.Amount
		memberSaleOrder.SetCancelInCashier(request.CancelReason)

		// 更新订单状态
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	// 发布“整单取消”操作事件
	utils.Go(func() {

		// 取消订单
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				Source:              ctx.GetSource(),
				SaleBillUuid:        billInfo.Uuid,
				OperatorUuid:        int64(ctx.GetStaffUuid()),
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
				MemberUuid:          memberSaleOrder.MemberUuid,
			},
			Data: event.CancelMemberOrderPayloadData{
				Type: "shop_cancel",
				Refunds: func() []event.CancelMemberOrderPayloadDataRefund {
					refunds := make([]event.CancelMemberOrderPayloadDataRefund, 0)
					if returnOrder != nil {
						for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
							refunds = append(refunds, event.CancelMemberOrderPayloadDataRefund{
								Name:              returnOrderAmount.PaymentMethod.PaymentName,
								Code:              returnOrderAmount.PaymentMethod.Code,
								Amount:            returnOrderAmount.Amount,
								RefundStatus:      returnOrderAmount.RefundStatus,
								ReturnAmountUuid:  returnOrderAmount.Uuid,
								ReturnOrderUuid:   returnOrder.Uuid,
								PaymentOrderUuid:  returnOrderAmount.PaymentOrderUuid,
								PaymentMethodUuid: returnOrderAmount.PaymentMethodUuid,
							})
						}
					}
					return refunds
				}(),
			},
		})
	})

	// 发送短信通知
	utils.Go(func() {
		NewSMSSrv(s.dbm).SendDeliveryOrderCancelSMS(ctx, memberSaleOrder.ContactPhone, &sms.DeliveryOrderCancel{
			CancelScene: sms.TemplateDeliveryOrderCanceledByMerchant,
			Company:     ctx.GetCompany().Name,
			OrderNo:     memberSaleOrder.OrderNo,
		})
	})

	// 成功后，推送到厨显端更新订单
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
			"update_time": time.Now().Unix(),
		})
	})

	//
	return nil
}

// GetRiderInfo 获取骑手信息
func (s *orderSrv) GetRiderInfo(ctx context.Context, getRiderInfoReq member_req.GetRiderInfoReq) (*resp.MemberOrderCoordinates, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(getRiderInfoReq.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if !memberSaleOrder.IsRiderPickup() {
		return nil, errors.New("订单未骑手接单")
	}

	companySetting := ctx.GetCompanySetting()
	company := ctx.GetCompany()

	merchantLat, merchantLng := companySetting.GetCoordinates()
	customerLat, customerLng := memberSaleOrder.GetCustomerLocation()

	// 调用外送服务获取骑手位置接口，如果获取成功，则更新member_sale_order的location，否则使用member_sale_order的location
	var riderLat, riderLng string
	takeoutSrv := takeout.NewTakeoutSrv()
	// 获取骑手实时位置
	riderLocation, err := takeoutSrv.GetDriverInfo(contexts.Background(), &req.GetDriverInfoReq{
		ShopOrderUuid: strconv.FormatUint(memberSaleOrder.Uuid, 10),
	})
	if err == nil && riderLocation != nil {
		riderLat = fmt.Sprintf("%.6f", riderLocation.Lat)
		riderLng = fmt.Sprintf("%.6f", riderLocation.Lng)
		memberSaleOrder.Location = fmt.Sprintf("%s,%s", riderLat, riderLng)
		// 更新骑手位置到订单
		if err = repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
			logger.Logger.Error("更新骑手位置失败",
				zap.Error(err),
				zap.Uint64("member_sale_order_uuid", memberSaleOrder.Uuid),
			)
		}
	} else {
		riderLat, riderLng = memberSaleOrder.GetLocation()
	}

	return &resp.MemberOrderCoordinates{
		Merchant: resp.OrderCoordinate{
			Name:    company.Name,
			Address: companySetting.Address,
			Lat:     merchantLat,
			Lng:     merchantLng,
		},
		Customer: resp.OrderCoordinate{
			Name:    memberSaleOrder.ContactName,
			Address: memberSaleOrder.ContactAddressDetail,
			Lat:     customerLat,
			Lng:     customerLng,
		},
		DriverInfo: resp.DriverInfoResp{
			Name:          memberSaleOrder.RiderName,
			Phone:         memberSaleOrder.RiderPhone,
			Avatar:        memberSaleOrder.RiderAvatar,
			Rating:        memberSaleOrder.RiderRating,
			Lat:           riderLat,
			Lng:           riderLng,
			EstimatedTime: memberSaleOrder.ExpectedFinishTime,
		},
	}, nil

}

// SendAuthCode 发送认证验证码
func (s *orderSrv) SendAuthCode(ctx context.Context, req req.MemberOrderSendAuthCodeReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	companyUuid := ctx.GetCompanyUuid()
	// 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(req.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}
	// 获取商家信息
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(companyUuid)
	if err != nil || company.IsExpired() || company.IsDelete() {
		return errors.New("无法使用该功能，请联系商家")
	}
	if company.CompanySetting == nil || company.CompanySetting.IsOpenMember != 1 {
		return errors.New("商家会员服务已关闭")
	}
	// 获取验证码
	code, err := validator.GetCode(cache.Global, companyUuid, memberSaleOrder.ContactPhone)
	if err != nil {
		return err
	}
	// 发送验证码短信
	if err := s.smsSrv.SendMemberAuthOrderCodeSMS(ctx, memberSaleOrder.ContactPhone, &sms.MemberSendCodeRequest{
		Code: code,
	}); err != nil {
		return err
	}
	return nil
}

// GetMemberCashierOrderList 获取收银端"外送"订单列表
func (s *orderSrv) GetMemberCashierOrderList(ctx context.Context, request req.MemberOrderListReq) (*resp.GetMemberCashierOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	orderTotal := int64(0)
	var extra resp.ExtraMemberCashierOrderListMeta
	memberOrders := make([]resp.MemberCashierOrder, 0)
	if request.Keyword != "" {
		res, err := s.GetMemberCashierOrderSearch(ctx, req.MemberOrderSearchReq{Keyword: request.Keyword})
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		memberOrders = res.List
		orderTotal = int64(len(memberOrders))
		// 获取数量
		getOrderNum := func(status []uint) int64 {
			statusMap := make(map[uint]bool)
			for _, v := range status {
				statusMap[v] = true
			}
			num := int64(0)
			for _, v := range memberOrders {
				if _, ok := statusMap[v.Status]; ok {
					num++
				}
			}
			return num
		}
		extra = resp.ExtraMemberCashierOrderListMeta{
			UnacceptNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUnaccept)),
			AcceptNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusAccept)),
			UndeliveryNum: getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUndelivery)),
			DeliveryNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivery)),
			CompletedNum:  getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivered)),
			CancelNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusCancel)),
		}
	} else {
		memberSaleOrders, total, err := repository.NewMemberSaleOrderRepo(db).GetCashierMemberSaleOrderList(
			request.PageNo,
			request.PageSize,
			constant.GetStatusList(request.Status),
		)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		orderTotal = total
		for _, memberSaleOrder := range memberSaleOrders {
			// 支付超时自动取消订单
			if memberSaleOrder.GetRemainingPaymentTime() == 0 && memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingPayment {
				s.MemberOrderPayTimeoutAutoCancel(ctx, memberSaleOrder.Uuid)
				memberSaleOrder.Status = constant.MemberSaleOrderStatusCancelled
			}
			memberOrders = append(memberOrders, resp.MemberCashierOrder{
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
				SerialNumber:        memberSaleOrder.SerialNumber,
				Status:              memberSaleOrder.Status,
				StatusGroup:         constant.ParseToStatusGroup(memberSaleOrder.Status),
				Num:                 memberSaleOrder.ProductNum,
				ProductAmount:       memberSaleOrder.Amount,
			})
		}
		// 获取数量
		getOrderNum := func(status []uint) int64 {
			num, _ := repository.NewMemberSaleOrderRepo(db).GetOrderNum(status)
			return num
		}
		extra = resp.ExtraMemberCashierOrderListMeta{
			UnacceptNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUnaccept)),
			AcceptNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusAccept)),
			UndeliveryNum: getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUndelivery)),
			DeliveryNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivery)),
			CompletedNum:  getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivered)),
			CancelNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusCancel)),
		}
	}

	return &resp.GetMemberCashierOrderListResp{
		Meta: dto.PageResponse{
			PageNo:   request.PageNo,
			PageSize: request.PageSize,
			Total:    orderTotal,
		},
		Extra: extra,
		List:  memberOrders,
	}, nil
}

// GetMemberOrderDetail 获取收银端"外送"订单详情
func (s *orderSrv) GetMemberCashierOrderDetail(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderCashierDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrder, err := getMemberOrderDetail(ctx, req.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	products := make([]resp.MemberOrderProduct, 0)
	for _, saleOrderProduct := range memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts {
		products = append(products, resp.MemberOrderProduct{
			LocaleName:          saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
			OriginTotalPrice:    saleOrderProduct.GetTotalPriceOrigin(),
		})
	}

	var address resp.MemberOrderDetailAddress
	if memberSaleOrder.Address != nil {
		address = resp.MemberOrderDetailAddress{
			ContactName: memberSaleOrder.ContactName,
			Phone:       memberSaleOrder.ContactPhone,
			PhonePrefix: memberSaleOrder.ContactPhonePrefix,
			Address:     memberSaleOrder.ContactAddress + memberSaleOrder.ContactAddressDetail,
		}
	}
	return &resp.GetMemberOrderCashierDetailResp{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		PayTime:             memberSaleOrder.PayTime,
		FinishTime:          memberSaleOrder.FinishTime,
		CancelTime:          memberSaleOrder.CancelTime,
		CancelReason:        memberSaleOrder.CancelReason,
		Remark:              memberSaleOrder.Remark,
		AmountInfo: resp.MemberOrderAmountInfo{
			Amount: memberSaleOrder.Amount,
		},
		ProductList: resp.MemberProductList{
			List:          products,
			ProductAmount: memberSaleOrder.SaleBill.Amount, // 外送订单的商品金额等于SaleBill的amount，因为SaleBill的amount是包括了除配送费之外的所有金额
		},
		AddressInfo: address,
		Rider: resp.RiderInfo{
			Name:  memberSaleOrder.RiderName,
			Phone: memberSaleOrder.RiderPhone,
		},
	}, nil
}

// GetMemberOrderManageList 获取收银机订单管理外送订单列表
func (s *orderSrv) GetMemberOrderManageList(ctx context.Context, req req.MemberOrderManageListReq) (*resp.GetMemberOrderManageListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	var orderNo *string
	if req.OrderNo != "" {
		orderNo = &req.OrderNo
	}
	var serialNo *string
	if req.SerialNo != "" {
		serialNo = &req.SerialNo
	}
	memberSaleOrders, total, err := repository.NewMemberSaleOrderRepo(db).GetCashierMemberSaleOrderManageList(req.PageNo, req.PageSize, constant.GetStatusList(req.Status), repository.GetCashierMemberSaleOrderManageListReq{
		OrderNo:    orderNo,
		SerialNo:   serialNo,
		TimeFilter: req.GetTimeFilterParams(ctx.GetCompanySetting().Timezone),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	memberOrders := make([]resp.MemberOrderManage, 0)
	for _, memberSaleOrder := range memberSaleOrders {
		var payType string
		if memberSaleOrder.PaymentMethod != nil {
			payType = memberSaleOrder.PaymentMethod.PaymentName
		}
		// 支付超时自动取消订单
		if memberSaleOrder.GetRemainingPaymentTime() == 0 && memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingPayment {
			s.MemberOrderPayTimeoutAutoCancel(ctx, memberSaleOrder.Uuid)
			memberSaleOrder.Status = constant.MemberSaleOrderStatusCancelled
		}
		// 获取联系人信息
		var contact resp.ContactInfo
		if memberSaleOrder.Address != nil {
			contact = resp.ContactInfo{
				Name:  memberSaleOrder.ContactName,
				Phone: memberSaleOrder.ContactPhone,
			}
		}

		var rider resp.RiderInfo
		rider.Name = memberSaleOrder.RiderName
		rider.Phone = memberSaleOrder.RiderPhone
		if rider.Name == "" {
			rider.Name = "-"
		}
		if rider.Phone == "" {
			rider.Phone = "*"
		}

		memberOrders = append(memberOrders, resp.MemberOrderManage{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			SerialNumber:        memberSaleOrder.SerialNumber,
			OrderNo:             memberSaleOrder.OrderNo,
			Status:              memberSaleOrder.Status,
			StatusGroup:         constant.ParseToStatusGroup(memberSaleOrder.Status),
			CreateTime:          memberSaleOrder.CreateTime,
			PayTime:             memberSaleOrder.PayTime,
			OriginAmount:        memberSaleOrder.OriginAmountValue(),
			PayAmount:           memberSaleOrder.Amount,
			DeliveryFee:         memberSaleOrder.DeliveryFeeAmount,
			PayType:             payType,
			Contact:             contact,
			Rider:               rider,
			Extra: resp.OrderListsExtra{
				IsCellReject: func() bool {
					// 待商家接单时，可以拒单
					return memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingMerchantAccept
				}(),
				IsCellCancel: func() bool {
					// 商家备餐中，商家可以取消订单
					return memberSaleOrder.Status == constant.MemberSaleOrderStatusCooking
				}(),
				IsCellRefund: memberSaleOrder.IsCanRefund(),
				IsCellContactRider: func() bool {
					// 骑手配送中，可以联系骑手
					return memberSaleOrder.Status == constant.MemberSaleOrderStatusDelivering ||
						memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingRiderDelivery
				}(),
			},
		})
	}

	var opts []repository.DBOption
	if req.SerialNo != "" {
		opts = append(opts, repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySerialNumber(req.SerialNo)))
	}
	if req.OrderNo != "" {
		opts = append(opts, repository.CommonRepo.DBOption(repository.CommonRepo.WhereByOrderNo(req.OrderNo)))
	}

	getOrderNum := func(status string) int64 {
		opts = append(opts, repository.CommonRepo.DBOption(repository.CommonRepo.WhereByNoSelectingTimeout())) // 不包含选购超时订单
		num, _ := repository.NewMemberSaleOrderRepo(db).GetCashierMemberSaleOrderNum(constant.GetStatusList(status), req.GetTimeFilterParams(ctx.GetCompanySetting().Timezone), opts...)
		return num
	}

	unpaidNum := getOrderNum(constant.CashierSaleMemberOrderStatusUnpaid)         // 待付款
	unacceptNum := getOrderNum(constant.CashierMemberSaleOrderStatusUnaccept)     // 待接单
	acceptNum := getOrderNum(constant.CashierMemberSaleOrderStatusAccept)         // 备餐中
	undeliveryNum := getOrderNum(constant.CashierMemberSaleOrderStatusUndelivery) // 待配送
	deliveryNum := getOrderNum(constant.CashierMemberSaleOrderStatusDelivery)     // 配送中
	cancelNum := getOrderNum(constant.CashierMemberSaleOrderStatusCancel)         // 已取消
	completeNum := getOrderNum(constant.CashierMemberSaleOrderStatusDelivered)    // 已完成

	allNum := unpaidNum + unacceptNum + acceptNum + undeliveryNum + deliveryNum + cancelNum + completeNum

	return &resp.GetMemberOrderManageListResp{
		Meta: resp.OrderManageListMeta{
			PageResponse: dto.PageResponse{
				PageNo:   req.PageNo,
				PageSize: req.PageSize,
				Total:    total,
			},
			TotalNum:      allNum,
			UnpaidNum:     unpaidNum,
			UnacceptNum:   unacceptNum,
			AcceptNum:     acceptNum,
			UndeliveryNum: undeliveryNum,
			DeliveryNum:   deliveryNum,
			CompleteNum:   completeNum,
			CancelNum:     cancelNum,
		},
		List: memberOrders,
	}, nil
}

// GetMemberOrderManageDetail 获取收银机订单管理外送订单详情
func (s *orderSrv) GetMemberOrderManageDetail(ctx context.Context, req req.GetMemberOrderManageDetailReq) (*resp.GetMemberOrderManageDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrder, err := getMemberOrderDetail(ctx, req.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取订单商品列表
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	products := make([]resp.MemberOrderManageProduct, 0)
	for _, saleOrderProduct := range memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts {
		var imageUrl string
		if saleOrderProduct.ImageFile != nil {
			imageUrl = saleOrderProduct.ImageFile.GetUrl(baseUrl)
		}
		products = append(products, resp.MemberOrderManageProduct{
			LocaleName:          saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			ImageUrl:            imageUrl,
			OriginUnitPrice:     saleOrderProduct.OriginTotalPrice,
			UnitPrice:           saleOrderProduct.TotalPrice,
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
			OriginTotalPrice:    saleOrderProduct.GetTotalPriceOrigin(),
			RefundAmount:        saleOrderProduct.GetReturnPrice(),
		})
	}

	// 获取操作日志
	operationLogs, err := s.GetRecordList(ctx, memberSaleOrder.SaleBill.Uuid, 0)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	var payType string
	if memberSaleOrder.PaymentMethod != nil {
		payType = memberSaleOrder.PaymentMethod.PaymentName
	}
	var member resp.OrderMember
	if memberSaleOrder.SaleBill.SaleOrders[0].Member != nil {
		if !memberSaleOrder.SaleBill.SaleOrders[0].Member.IsVisitor {
			member = resp.OrderMember{
				ID:   memberSaleOrder.SaleBill.SaleOrders[0].Member.ID,
				Name: memberSaleOrder.SaleBill.SaleOrders[0].Member.Nickname,
			}
		}
	}
	return &resp.GetMemberOrderManageDetailResp{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		BillType:            2,
		SerialNo:            memberSaleOrder.SerialNumber,
		OrderNo:             memberSaleOrder.OrderNo,
		Status:              memberSaleOrder.Status,
		OriginAmount:        memberSaleOrder.OriginAmountValue(),
		PayAmount:           memberSaleOrder.Amount,
		RefundAmount:        memberSaleOrder.RefundAmount,
		MemberDiscount:      memberSaleOrder.MemberDiscountFee,
		DeliveryFee:         memberSaleOrder.DeliveryFeeAmount,
		PayType:             payType,
		PayTime:             memberSaleOrder.PayTime,
		CreateTime:          memberSaleOrder.CreateTime,
		FinshTime:           memberSaleOrder.FinishTime,
		CancelReason: func() string {
			if memberSaleOrder.IsSelfCancel() {
				return i18n.Translate(ctx.GetLanguage(), "自主取消") + " (" + memberSaleOrder.CancelReason + ")"
			}
			if memberSaleOrder.IsMerchantCancel() {
				return i18n.Translate(ctx.GetLanguage(), "商家取消") + " (" + memberSaleOrder.CancelReason + ")"
			}
			return i18n.Translate(ctx.GetLanguage(), memberSaleOrder.CancelReason)
		}(),
		CancelTime: memberSaleOrder.CancelTime,
		Remark:     memberSaleOrder.Remark,
		Member:     member,
		Cachier: resp.CachierInfo{
			Uuid: memberSaleOrder.SaleBill.CashierUuid,
			Name: memberSaleOrder.SaleBill.CashierName,
		},
		ProductList:  resp.MemberProductManageList{List: products},
		OperationLog: resp.OperationLog{List: operationLogs},
	}, nil
}

type AcceptMemberSaleOrderFunc func(opt *AcceptMemberSaleOrderOption)

type AcceptMemberSaleOrderOption struct {
	IsAutoAccept bool // 是否自动接单
}

func WithIsAutoAccept() AcceptMemberSaleOrderFunc {
	return func(opt *AcceptMemberSaleOrderOption) {
		opt.IsAutoAccept = true
	}
}

// AcceptMemberSaleOrder 接单外送订单
func (s *orderSrv) AcceptMemberSaleOrder(ctx context.Context, request req.AcceptOrderReq, options ...AcceptMemberSaleOrderFunc) error {
	opt := &AcceptMemberSaleOrderOption{}
	for _, option := range options {
		option(opt)
	}

	if ctx.GetDB() == nil {
		db := s.dbm.GetDB(ctx.GetDbId())
		ctx.SetDB(db)
	}
	db := ctx.GetDB()

	// 获取订单信息
	memberSaleOrder, err := getMemberOrderDetail(ctx, request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 接单
	memberSaleOrder.Accept()

	//  获取未送厨的商品列表
	unCookingSaleOrderProducts := memberSaleOrder.SaleBill.GetSaleOrderProductUnCooking()
	if len(unCookingSaleOrderProducts) > 0 {
		// 整单送厨
		ctx.SetScene(constant.SceneMemberOrder)
		_, checkRes, err := s.InstantOrderCartProductCooking(ctx, req.OrderCartProductCookingReq{
			SaleBillUuid:        memberSaleOrder.SaleBill.Uuid,
			IgnoreMust:          true,
			IsMemberOrderAccept: true,
		})
		if err != nil {
			return errors.WithMessage(err, "整单送厨失败")
		}
		if checkRes != nil {
			ctx.Log().Info("接单外送订单时送厨失败", zap.Any("checkRes", checkRes))
			return errors.New("商品库存不足")
		}
	}
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		res, err := createSkootarOrder(ctx, memberSaleOrder)
		if err != nil {
			if strings.Contains(err.Error(), "创建skootar外送订单失败") {
				return errors.WithMessage(errors.NewWithCode(constant.CodeTakeoutCreateOrderError, "外送订单创建失败"), err.Error())
			}
			return errors.WithMessage(errors.NewWithCode(constant.CodeOrderPrepareDataError, "发起外送订单失败"), err.Error())
		}
		memberSaleOrder.RelatedOrderNo = res.TakeoutRefNo
		memberSaleOrder.ExpectedFinishTime = res.FinishTime
		if opt.IsAutoAccept {
			memberSaleOrder.IsAutoAccept = constant.Yes          // 自动接单
			memberSaleOrder.PayTime = memberSaleOrder.AcceptTime // 支付时间等于接单时间
		}
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderAccept(*memberSaleOrder); err != nil {
			return errors.WithMessage(err, "更新外送订单失败")
		}
		return nil
	}); err != nil {
		ctx.Log().Error("接单失败", zap.Error(err))
		return errors.WithMessage(err, "更新外送订单失败")
	}

	// 添加骑手接单超时自动取消订单的延时队列任务
	if Queue.MemberOrderCancelQueue != nil {
		utils.Go(func() {
			memberSaleOrderUuidStr := strconv.FormatUint(request.MemberSaleOrderUuid, 10)
			// 构建队列消息参数
			paramsJson := utils.ToJson(map[string]interface{}{
				"member_sale_order_uuid": request.MemberSaleOrderUuid,
				"company_uuid":           ctx.GetCompanyUuid(),
				"cancel_scene":           constant.MemberSaleOrderSceneRiderPickupTimeout,
			})
			// 发送骑手接单超时时间后自动取消订单的延时消息
			_, err := Queue.MemberOrderCancelQueue.SendDelayMsgV2(
				paramsJson,
				time.Duration(memberSaleOrder.RiderAcceptTimeout)*time.Minute, // 骑手接单超时时间后执行
				delayqueue.WithRetryCount(3),                                  // 重试3次
			)
			if err != nil {
				ctx.Log().Error("添加骑手接单超时自动取消订单任务失败",
					zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr),
					zap.Error(err))
			}
		})
	}

	// 发布"外送接单"操作事件
	utils.Go(func() {
		// 设置商品状态为送厨状态
		memberSaleOrder.SaleBill.SetSaleOrderProductCooking()
		// 发布"外送接单"操作事件
		s.bus.PublishAcceptMemberSaleOrderEvent(event.AcceptMemberSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				Source:              ctx.GetSource(),
				SaleBillUuid:        memberSaleOrder.SaleBillUuid,
				SaleOrderUuid:       memberSaleOrder.SaleOrderUuid,
				OperatorUuid:        int64(ctx.GetStaffUuid()),
				Staff:               ctx.GetStaff(),
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
				MemberUuid:          memberSaleOrder.MemberUuid,
			},
			MemberSaleOrder: memberSaleOrder,
		})
	})

	return nil
}

func createSkootarOrder(ctx context.Context, memberSaleOrder *model.MemberSaleOrder) (*resp.CreateTakeoutOrderResp, error) {
	db := ctx.GetDB()
	lat, lng, _ := memberSaleOrder.Address.GetLocation()
	// 获取商家地址
	companySetting := ctx.GetCompanySetting()
	if companySetting.ID == 0 {
		// 如果ctx中没有商家设置，则从数据库中获取
		if company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(ctx.GetCompanyUuid()); err == nil {
			companySetting = *company.CompanySetting
			ctx.SetCompany(*company)
			ctx.SetCompanySetting(companySetting)
		}
	}
	latitude, longitude := companySetting.GetCoordinates()
	if latitude == "" || longitude == "" {
		return nil, errors.New("无法找到商家经纬度")
	}
	// 选择外送渠道
	memberSaleOrder.RelatedOrderType = constant.ProviderNameSkootar
	// 状态变更的回调地址
	callbackUrl := config.Server.Domain + "/api/v1/member/order/callback?company_uuid=" + fmt.Sprintf("%d", ctx.GetCompany().Uuid)

	takeoutSrv := takeout.NewTakeoutSrv()
	params := req.CreateTakeoutOrderReq{
		ProviderName:  memberSaleOrder.RelatedOrderType,
		Remark:        memberSaleOrder.Remark,
		CallbackUrl:   callbackUrl,
		ShopOrderUuid: fmt.Sprintf("%d", memberSaleOrder.Uuid),
		CustomerLocation: &req.TakeoutLocation{
			ContactName:  memberSaleOrder.ContactName,
			ContactPhone: memberSaleOrder.ContactPhone,
			TakeoutAddress: req.TakeoutAddress{
				AddressName: memberSaleOrder.ContactName,
				Address:     memberSaleOrder.ContactAddress,
				Lat:         lat,
				Lng:         lng,
			},
		},
		MerchantLocation: &req.TakeoutLocation{
			ContactName:  ctx.GetCompany().Name,
			ContactPhone: companySetting.LinkPhone,
			TakeoutAddress: req.TakeoutAddress{
				AddressName: ctx.GetCompany().Name,
				Address:     companySetting.Address,
				Lat:         latitude,
				Lng:         longitude,
			},
		},
	}
	startTime := time.Now()
	res, err := takeoutSrv.CreateOrder(contexts.Background(), &params)
	if err != nil {
		ctx.Log().Info("创建外送订单失败", zap.Error(err), zap.Duration("cost", time.Since(startTime)))
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeTakeoutCreateOrderError, "创建skootar外送订单失败"), err.Error())
	}
	ctx.Log().Info("创建外送订单成功", zap.String("takeout_ref_no", res.TakeoutRefNo), zap.Duration("cost", time.Since(startTime)))

	return res, nil
}

// RejectMemberSaleOrder 拒单外送订单
func (s *orderSrv) RejectMemberSaleOrder(ctx context.Context, request req.RejectOrderReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 获取订单信息
	memberSaleOrder, err := getMemberOrderDetail(ctx, request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 获取销售账单
	billInfo, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(memberSaleOrder.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 拒单
	memberSaleOrder.Reject()

	// 整单取消SaleBill
	if err := s.CancelOrder(ctx, req.OrderCancelReq{
		SaleBillUuid: memberSaleOrder.SaleBill.Uuid,
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 退款
	returnOrder, err := NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*billInfo.GetFirstSaleOrder(), MemberSaleOrderRefundReq{
		CancelReason: constant.MemberSaleOrderSceneReason[memberSaleOrder.CancelScene],
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	// 更新订单状态
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderReject(*memberSaleOrder); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 发布"外送拒单"操作事件
	utils.Go(func() {
		s.bus.PublishRejectMemberSaleOrderEvent(event.RejectMemberSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  memberSaleOrder.SaleBill.Uuid,
				SaleOrderUuid: memberSaleOrder.SaleBill.SaleOrders[0].Uuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
				Staff:         ctx.GetStaff(),
			},
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			MemberSaleOrder:     memberSaleOrder,
			Data: event.CancelMemberOrderPayloadData{
				Type: "reject_order",
				Refunds: func() []event.CancelMemberOrderPayloadDataRefund {
					refunds := make([]event.CancelMemberOrderPayloadDataRefund, 0)
					if returnOrder != nil {
						for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
							refunds = append(refunds, event.CancelMemberOrderPayloadDataRefund{
								Name:              returnOrderAmount.PaymentMethod.PaymentName,
								Code:              returnOrderAmount.PaymentMethod.Code,
								Amount:            returnOrderAmount.Amount,
								RefundStatus:      returnOrderAmount.RefundStatus,
								ReturnAmountUuid:  returnOrderAmount.Uuid,
								ReturnOrderUuid:   returnOrder.Uuid,
								PaymentOrderUuid:  returnOrderAmount.PaymentOrderUuid,
								PaymentMethodUuid: returnOrderAmount.PaymentMethodUuid,
							})
						}
					}
					return refunds
				}(),
			},
		})
	})

	// 发送短信通知
	utils.Go(func() {
		NewSMSSrv(s.dbm).SendDeliveryOrderCancelSMS(ctx, memberSaleOrder.ContactPhone, &sms.DeliveryOrderCancel{
			CancelScene: sms.TemplateDeliveryRejected,
			Company:     ctx.GetCompany().Name,
			OrderNo:     memberSaleOrder.OrderNo,
		})
	})

	return nil
}

// CookFinishMemberSaleOrder 备餐完成外送订单
func (s *orderSrv) CookFinishMemberSaleOrder(ctx context.Context, request req.CookFinishOrderReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrder, err := getMemberOrderDetail(ctx, request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 备餐完成
	memberSaleOrder.CookFinish()

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新备餐完成
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderCookFinish(*memberSaleOrder); err != nil {
			return errors.WithMessage(err)
		}
		// 确认订单
		takeoutSrv := takeout.NewTakeoutSrv()
		if err := takeoutSrv.ConfirmOrder(contexts.Background(), &req.ConfirmTakeoutOrderReq{
			ShopOrderUuid: fmt.Sprintf("%d", memberSaleOrder.Uuid),
		}); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 发布"外送备餐完成"操作事件
	utils.Go(func() {
		s.bus.PublishCookFinishMemberSaleOrderEvent(event.CookFinishMemberSaleOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:           ctx,
				CompanyUuid:   ctx.GetCompanyUuid(),
				Source:        ctx.GetSource(),
				SaleBillUuid:  memberSaleOrder.SaleBill.Uuid,
				SaleOrderUuid: memberSaleOrder.SaleBill.SaleOrders[0].Uuid,
				OperatorUuid:  int64(ctx.GetStaffUuid()),
				Staff:         ctx.GetStaff(),
			},
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			MemberSaleOrder:     memberSaleOrder,
		})
	})

	return nil
}

// GetMemberCashierOrderSearch 订单搜索
func (s *orderSrv) GetMemberCashierOrderSearch(ctx context.Context, req req.MemberOrderSearchReq) (*resp.GetMemberCashierOrderSearchResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrders, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderByContactNameAndContactPhoneSuffix(req.Keyword, req.Keyword)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	memberOrders := make([]resp.MemberCashierOrder, 0)
	for _, memberSaleOrder := range memberSaleOrders {
		// 支付超时自动取消订单
		if memberSaleOrder.GetRemainingPaymentTime() == 0 && memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingPayment {
			s.MemberOrderPayTimeoutAutoCancel(ctx, memberSaleOrder.Uuid)
			memberSaleOrder.Status = constant.MemberSaleOrderStatusCancelled
		}
		//
		memberOrders = append(memberOrders, resp.MemberCashierOrder{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			SerialNumber:        memberSaleOrder.SerialNumber,
			Status:              memberSaleOrder.Status,
			StatusGroup:         constant.ParseToStatusGroup(memberSaleOrder.Status),
			Num:                 memberSaleOrder.ProductNum,
			ProductAmount:       memberSaleOrder.Amount,
		})
	}

	return &resp.GetMemberCashierOrderSearchResp{
		List: memberOrders,
	}, nil
}

// GetMemberOrderReturnInfo 获取外送订单退款弹窗信息
func (s *orderSrv) GetMemberOrderReturnInfo(ctx context.Context, req member_req.MemberOrderReturnInfoReq) (*resp.OrderReturnInfoResp, error) {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(req.MemberSaleOrderUuid)
		defer lock.NewSystemLock().UnlockUuid(req.MemberSaleOrderUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取外送订单信息
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(req.MemberSaleOrderUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(memberSaleOrder.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetFirstSaleOrder()
	if saleOrder == nil {
		return nil, errors.New("找不到销售订单")
	}

	products := make([]resp.OrderReturnProduct, 0)

	// 获取销售订单的每个付款单的可退款金额
	// 要求排好序：退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	paymentRecords, currencyUnit := saleOrder.GetPaymentOrderCanReturnAmount()

	// 配送费
	deliveryFee := memberSaleOrder.DeliveryFeeAmount

	// 构建退款支付记录
	memberPaymentRecords := make([]resp.OrderReturnPaymentRecord, 0)
	for _, record := range paymentRecords {
		canReturnAmount := decimal.NewFromFloat(record.CanReturnAmount).Sub(decimal.NewFromFloat(deliveryFee)).Round(2).InexactFloat64()
		paymentAmount := decimal.NewFromFloat(record.PaymentAmount).Sub(decimal.NewFromFloat(deliveryFee)).Round(2).InexactFloat64()
		memberPaymentRecords = append(memberPaymentRecords, resp.OrderReturnPaymentRecord{
			PaymentMethodCode: record.PaymentMethodCode,
			PaymentOrderUuid:  record.PaymentOrderUuid,
			PaymentMethodName: record.PaymentMethodName,
			PaymentMethodUuid: record.PaymentMethodUuid,
			CurrencyUnit:      record.CurrencyUnit,
			PaymentAmount:     paymentAmount,
			CanReturnAmount:   canReturnAmount,
		})
	}

	// 获取商品列表
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		if saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsGiftProduct() || saleOrderProduct.Status == constant.OrderProductStatusUnSending {
			continue
		}
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.GetLocaleName(), // Requirement: story-main-product-attribute-snapshot-fix
			LocaleAttributeName:  saleOrderProduct.GetAttributeName(),
			Num:                  saleOrderProduct.GetCanReturnNum(), // 可退货数量=订单商品数量-已退货数量
			NumType:              saleOrderProduct.NumType,
			Price:                saleOrderProduct.TotalPrice,
			CanReturnAmount:      saleOrderProduct.GetCanReturnPrice(),
			CurrencyUnit:         currencyUnit,
		})
	}

	// 过滤掉单价为0的商品
	productList := make([]resp.OrderReturnProduct, 0)
	for _, product := range products {
		if product.Price == 0 {
			continue
		}
		productList = append(productList, product)
	}

	// 可退款金额
	canReturnAmount := saleOrder.GetCanReturnAmountWithDeliveryFee(deliveryFee)
	res := &resp.OrderReturnInfoResp{
		ManualReturnPoints: saleOrder.CanManualReturnPoints(), // 是否可以手动退款积分。订单是按比例赠送积分且未发生积分抵扣时，不自动退款。
		DeductiblePoints:   saleOrder.GetManualReturnPoints(), // 可扣除积分。订单赠送的积分-已经退回的积分
		CanReturnAmount:    canReturnAmount,                   // 可退款金额. 可退款金额=订单最终应收金额-配送费-已退款金额
		PaymentRecords:     memberPaymentRecords,
		Products:           productList,
	}

	return res, nil
}

// MemberOrderReturn 外送订单退款/部分退款
func (s *orderSrv) MemberOrderReturn(ctx context.Context, memberReturnReq req.OrderReturnReq) (error, int) {
	// 禁止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(memberReturnReq.MemberSaleOrderUuid)
		defer s.lock.UnlockUuid(memberReturnReq.MemberSaleOrderUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取外送订单信息
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberReturnReq.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 检查订单状态是否允许退款（外送订单已支付且未取消状态下可以退款）
	if memberSaleOrder.Status < constant.MemberSaleOrderStatusPendingMerchantAccept || memberSaleOrder.IsCancel() {
		return errors.New("当前订单状态不允许退款"), constant.CodeFail
	}

	// 获取销售账单信息
	orderRepo := repository.NewOrderRepo(db)
	saleBill, err := orderRepo.GetSaleBillAllInfo(memberSaleOrder.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err), constant.CodeFail
	}

	// 获取销售订单信息
	saleOrder := saleBill.GetFirstSaleOrder()
	if saleOrder == nil {
		return errors.WithMessage(errors.New("找不到销售订单")), constant.CodeFail
	}

	if memberReturnReq.Points > saleOrder.GetManualReturnPoints() {
		return errors.WithMessage(errors.New("退款积分不能大于最大可退积分")), constant.CodeFail
	}

	// 构建退款请求，转换为用餐订单退款请求格式
	orderReturnReq := req.OrderReturnReq{
		SaleBillUuid:  memberSaleOrder.SaleBillUuid,
		SaleOrderUuid: saleOrder.Uuid,
		BankCode:      memberReturnReq.BankCode,
		AccountNo:     memberReturnReq.AccountNo,
		AccountName:   memberReturnReq.AccountName,
		Points:        memberReturnReq.Points,
	}

	// 转换商品列表
	for _, product := range memberReturnReq.Products {
		orderReturnReq.Products = append(orderReturnReq.Products, req.OrderReturnProduct{
			SaleOrderProductUuid: product.SaleOrderProductUuid,
			Num:                  product.Num,
		})
	}

	// 调用原有的退款逻辑
	ctx.SetScene(constant.SceneMemberOrder)
	err, codeResult := s.ReturnOrder(ctx, orderReturnReq)
	if err != nil {
		return errors.WithMessage(err), codeResult
	}

	return nil, codeResult
}

// MemberOrderReReturn 外送订单重新退款
func (s *orderSrv) MemberOrderReReturn(ctx context.Context, memberReReturnReq req.OrderReReturnReq) (error, int) {
	// 构建重新退款请求，转换为用餐订单重新退款请求格式
	orderReReturnReq := req.OrderReReturnReq{
		ReturnOrderUuid:  memberReReturnReq.ReturnOrderUuid,
		ReturnAmountUuid: memberReReturnReq.ReturnAmountUuid,
		BankCode:         memberReReturnReq.BankCode,
		AccountNo:        memberReReturnReq.AccountNo,
		AccountName:      memberReReturnReq.AccountName,
	}

	// 调用原有的重新退款逻辑
	return s.ReReturnOrder(ctx, orderReReturnReq)
}

// MemberOrderPayTimeoutAutoCancel 外送订单支付超时自动取消订单
func (s *orderSrv) MemberOrderPayTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error {
	// 获取DB
	db := ctx.GetDB()

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	// 2. 检查订单状态是否可以取消 - 只有待支付状态的订单才能自动取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusPendingPayment {
		if memberSaleOrder.Status == constant.MemberSaleOrderStatusCancelled {
			return nil
		}
		return errors.New("订单状态不支持取消")
	}
	if memberSaleOrder.GetRemainingPaymentTime() > 5 {
		return errors.New("订单支付未超时")
	}

	// 3. 执行取消操作
	reason := constant.MemberSaleOrderSceneReason[constant.MemberSaleOrderScenePaymentTimeout]
	memberSaleOrder.SetCancel(reason)
	memberSaleOrder.CancelScene = constant.MemberSaleOrderScenePaymentTimeout
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		logger.Logger.Error("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	// 4. 取消订单（如果有关联的销售账单）
	if memberSaleOrder.SaleBillUuid > 0 {
		if err := repository.NewOrderRepo(db).CancelOrder(ctx, memberSaleOrder.SaleBillUuid, 0, reason); err != nil {
			logger.Logger.Error("取消销售账单失败",
				zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid),
				zap.Uint64("saleBillUuid", memberSaleOrder.SaleBillUuid),
				zap.Error(err))
			// 这里不返回false，因为会员订单状态已经更新成功了
		}
	}

	// 发布“订单取消”操作事件
	utils.Go(func() {
		event.NewSystemBus().PublishCancelMemberOrderEvent(event.CancelMemberOrderPayload{
			BasePayload: event.BasePayload{ // 基础信息
				Ctx:                 ctx,
				CompanyUuid:         ctx.GetCompanyUuid(),
				Source:              constant.SourceMember,
				SaleBillUuid:        memberSaleOrder.SaleBillUuid,
				SaleOrderUuid:       memberSaleOrder.Uuid,
				MemberUuid:          memberSaleOrder.MemberUuid,
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
			},
			Data: event.CancelMemberOrderPayloadData{
				Type: "timeout_cancel",
			},
		})
	})

	return nil
}

// MemberOrderSelectingTimeoutAutoCancel 外送订单选购超时自动取消订单
func (s *orderSrv) MemberOrderSelectingTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error {
	// 获取DB
	db := ctx.GetDB()

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	// 2. 检查订单状态是否可以取消 - 只有选购中状态的订单才能自动取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusSelecting {
		return errors.New("订单状态不支持取消")
	}

	// 3. 执行取消操作
	reason := constant.MemberSaleOrderSceneReason[constant.MemberSaleOrderSceneSelectingTimeout]
	memberSaleOrder.SetCancel(reason)
	memberSaleOrder.CancelScene = constant.MemberSaleOrderSceneSelectingTimeout
	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
		logger.Logger.Error("更新会员订单状态失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	return nil
}

// MemberOrderRiderPickupTimeoutAutoCancel 外送订单骑手接单超时自动取消订单
func (s *orderSrv) MemberOrderRiderPickupTimeoutAutoCancel(ctx context.Context, memberSaleOrderUuid uint64) error {
	// 获取DB
	db := ctx.GetDB()

	// 1. 获取会员端销售订单
	memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(memberSaleOrderUuid)
	if err != nil {
		logger.Logger.Error("获取会员端销售订单失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	// 2. 检查订单状态是否可以取消 - 只有待骑手接单状态的订单才能自动取消
	if memberSaleOrder.Status != constant.MemberSaleOrderStatusCooking &&
		memberSaleOrder.Status != constant.MemberSaleOrderStatusPendingRiderPickup {
		return errors.New("订单状态不支持取消")
	}

	// 3. 调用外送取消
	if err := takeout.NewTakeoutSrv().CancelOrder(contexts.Background(), &req.CancelTakeoutOrderReq{
		ShopOrderUuid: fmt.Sprintf("%d", memberSaleOrder.Uuid),
	}); err != nil {
		logger.Logger.Error("自动取消订单失败 - 取消外送订单失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	// 事务开始
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 获取订单信息
		orderRepo := repository.NewOrderRepo(tx)
		billInfo, err := orderRepo.GetSaleBillAllInfo(memberSaleOrder.SaleBillUuid)
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}

		// 获取销售订单
		saleOrder := billInfo.GetFirstSaleOrder()
		if saleOrder == nil {
			return errors.WithMessage(errors.New("找不到销售订单"))
		}

		// 4. 已经支付的-发起退款
		var returnOrder *model.ReturnOrder
		returnOrder, err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*saleOrder, MemberSaleOrderRefundReq{
			CancelReason: constant.MemberSaleOrderSceneReason[constant.MemberSaleOrderSceneRiderPickupTimeout],
		})
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}

		// 5. 执行取消操作
		reason := constant.MemberSaleOrderSceneReason[constant.MemberSaleOrderSceneRiderPickupTimeout]
		memberSaleOrder.SetCancel(reason)
		memberSaleOrder.CancelScene = constant.MemberSaleOrderSceneRiderPickupTimeout
		memberSaleOrder.RefundAmount = returnOrder.RefundAmount
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrder(*memberSaleOrder); err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}

		// 发布“订单取消”操作事件
		utils.Go(func() {
			s.bus.PublishCancelMemberOrderEvent(event.CancelMemberOrderPayload{
				BasePayload: event.BasePayload{
					Ctx:                 ctx,
					CompanyUuid:         ctx.GetCompanyUuid(),
					Source:              ctx.GetSource(),
					SaleBillUuid:        memberSaleOrder.SaleBillUuid,
					SaleOrderUuid:       memberSaleOrder.Uuid,
					OperatorUuid:        0,
					MemberUuid:          memberSaleOrder.MemberUuid,
					MemberSaleOrderUuid: memberSaleOrder.Uuid,
				},
				Data: event.CancelMemberOrderPayloadData{
					Type: "rider_pickup_timeout",
					Refunds: func() []event.CancelMemberOrderPayloadDataRefund {
						refunds := make([]event.CancelMemberOrderPayloadDataRefund, 0)
						if returnOrder != nil {
							for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
								refunds = append(refunds, event.CancelMemberOrderPayloadDataRefund{
									Name:              returnOrderAmount.PaymentMethod.PaymentName,
									Code:              returnOrderAmount.PaymentMethod.Code,
									Amount:            returnOrderAmount.Amount,
									RefundStatus:      returnOrderAmount.RefundStatus,
									ReturnAmountUuid:  returnOrderAmount.Uuid,
									ReturnOrderUuid:   returnOrder.Uuid,
									PaymentOrderUuid:  returnOrderAmount.PaymentOrderUuid,
									PaymentMethodUuid: returnOrderAmount.PaymentMethodUuid,
								})
							}
						}
						return refunds
					}(),
				},
			})
		})

		return nil
	}); err != nil {
		logger.Logger.Error("自动取消订单失败", zap.Uint64("memberSaleOrderUuid", memberSaleOrderUuid), zap.Error(err))
		return err
	}

	return nil
}

// CreateMemberDineInOrder 创建会员端堂食订单
// 1. sale_bill_uuid 为0时，新建订单并添加商品
// 2. sale_bill_uuid 不为0时，更新已有订单（识别新增、修改、删除的商品）
// 与外送订单的核心差异：BillType=Instant, DiningMethod=DineIn, 商品价格与收银机一致
func (s *orderSrv) CreateMemberDineInOrder(ctx context.Context, request req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error) {
	if err := request.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	if request.SaleBillUuid == 0 {
		// 新建订单
		return s.createDineInOrder(ctx, request)
	}
	// 更新已有订单
	return s.updateDineInOrder(ctx, request)
}

// createDineInOrder 创建堂食订单（内部方法）
func (s *orderSrv) createDineInOrder(ctx context.Context, request req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error) {
	// 创建订单
	instantOrder, err := s.createInstantOrder(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	targetSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(instantOrder.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "订单不存在")
	}

	// 添加所有商品
	if err := s.addDineInProducts(ctx, instantOrder.SaleBillUuid, instantOrder.SaleOrderUuid, request.Products, targetSaleBill); err != nil {
		return nil, err
	}

	return &resp.CreateInstantOrderResp{
		SaleBillUuid:  instantOrder.SaleBillUuid,
		SaleOrderUuid: instantOrder.SaleOrderUuid,
	}, nil
}

// updateDineInOrder 更新堂食订单（内部方法）
// 实现增量更新：识别新增、修改、删除的商品
func (s *orderSrv) updateDineInOrder(ctx context.Context, request req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	saleBillUuid := request.SaleBillUuid
	saleOrderUuid := request.SaleOrderUuid

	// 获取销售账单信息
	targetSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "订单不存在")
	}

	// 检查订单类型：只能更新堂食订单
	if !targetSaleBill.IsInstantBill() {
		return nil, errors.New("只能更新堂食订单")
	}

	// 检查订单状态：已取消或已完成的订单不允许更新
	if targetSaleBill.IsCanceled() {
		return nil, errors.New("订单已取消，无法更新")
	}
	if targetSaleBill.IsFinish() {
		return nil, errors.New("订单已完成，无法更新")
	}

	// 如果未指定 SaleOrderUuid，使用第一个销售订单
	if saleOrderUuid == 0 {
		if len(targetSaleBill.SaleOrders) == 0 {
			return nil, errors.New("订单数据异常：无销售订单")
		}
		saleOrderUuid = targetSaleBill.SaleOrders[0].Uuid
	}

	// 计算商品差异
	addProducts, deleteProducts, updateProducts, updateOlderProducts, err := s.diffDineInProducts(ctx, db, request.Products, targetSaleBill)
	if err != nil {
		return nil, err
	}

	// 执行删除
	for _, saleOrderProduct := range deleteProducts {
		saleOrderProduct.DeleteProduct()
	}

	// 执行修改
	for key, product := range updateProducts {
		saleOrderProduct := updateOlderProducts[key]
		saleOrderProduct.Num = product.Num
		saleOrderProduct.SetUpdate() // 标记该商品需要更新
	}

	// 执行新增
	if err := s.addDineInProducts(ctx, saleBillUuid, saleOrderUuid, s.toOrderProductAddReqSlice(addProducts), targetSaleBill); err != nil {
		return nil, err
	}

	return &resp.CreateInstantOrderResp{
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
	}, nil
}

// diffDineInProducts 计算堂食订单商品差异
// 返回：新增商品、删除商品、修改商品、修改前商品
func (s *orderSrv) diffDineInProducts(ctx context.Context, db *gorm.DB, products []req.OrderProductAddReq, saleBill *model.SaleBill) (
	addProducts map[string]req.OrderProductAddReq,
	deleteProducts map[string]*model.SaleOrderProduct,
	updateProducts map[string]req.OrderProductAddReq,
	updateOlderProducts map[string]*model.SaleOrderProduct,
	err error,
) {
	// 构建提交商品映射
	commitProductMap := make(map[string]req.OrderProductAddReq)
	for index := range products {
		product := products[index]
		productPackageAttributes, err := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(ctx.GetCompanyUuid(), product.AttributeUuidList)
		if err != nil {
			return nil, nil, nil, nil, errors.WithMessage(err)
		}
		attributeUuids := make([]uint64, 0)
		for _, attribute := range productPackageAttributes {
			attributeUuids = append(attributeUuids, attribute.AttributeUuid)
		}
		key := product.ProductKey(attributeUuids)
		commitProductMap[key] = product
	}

	// 构建已有商品映射（过滤已软删除的商品）
	olderProductMap := make(map[string]*model.SaleOrderProduct)
	if len(saleBill.SaleOrders) > 0 {
		for index := range saleBill.SaleOrders[0].SaleOrderProducts {
			product := saleBill.SaleOrders[0].SaleOrderProducts[index]
			// 跳过已删除的商品
			if product.DeleteTime != 0 {
				continue
			}
			key := product.ProductKey()
			olderProductMap[key] = product
		}
	}

	addProducts = make(map[string]req.OrderProductAddReq)
	deleteProducts = make(map[string]*model.SaleOrderProduct)
	updateProducts = make(map[string]req.OrderProductAddReq)
	updateOlderProducts = make(map[string]*model.SaleOrderProduct)

	// 提交中存在，订单中不存在 → 新增
	for key, product := range commitProductMap {
		if _, ok := olderProductMap[key]; !ok {
			addProducts[key] = product
		}
	}

	// 提交中不存在，订单中存在 → 删除
	for key, product := range olderProductMap {
		if _, ok := commitProductMap[key]; !ok {
			deleteProducts[key] = product
		}
	}

	// 提交和订单都存在 → 修改
	for key, product := range commitProductMap {
		if _, ok := olderProductMap[key]; ok {
			updateProducts[key] = product
			updateOlderProducts[key] = olderProductMap[key]
		}
	}

	return addProducts, deleteProducts, updateProducts, updateOlderProducts, nil
}

// addDineInProducts 添加堂食订单商品
// 注意：即使 products 为空，也需要调用 ActionAdd 来保存 targetSaleBill 中已标记为更新的对象（删除、修改）
func (s *orderSrv) addDineInProducts(ctx context.Context, saleBillUuid, saleOrderUuid uint64, products []req.OrderProductAddReq, targetSaleBill *model.SaleBill) error {
	productParams := make([]req.ProductParams, 0, len(products))
	for _, product := range products {
		param := req.ProductParams{
			FlavorProductBomUuid:            product.FlavorUuid,
			Num:                             product.Num,
			SauceProductBomUuidList:         product.SauceUuidList,
			ProductPackageAttributeUuidList: product.AttributeUuidList,
		}
		productParams = append(productParams, param)
	}

	params := req.ProductAddReq{
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
		Products:      productParams,
		IsMemberAdd:   false, // 堂食订单不应用外送折扣率
	}
	return s.ActionAdd(ctx, params, targetSaleBill)
}

// toOrderProductAddReqSlice 将 map 转换为 slice
func (s *orderSrv) toOrderProductAddReqSlice(products map[string]req.OrderProductAddReq) []req.OrderProductAddReq {
	result := make([]req.OrderProductAddReq, 0, len(products))
	for _, product := range products {
		result = append(result, product)
	}
	return result
}

// ==================== 会员端堂食订单管理 ====================

// GetMemberDineInOrderList 获取会员端堂食订单列表
func (s *orderSrv) GetMemberDineInOrderList(ctx context.Context, listReq req.MemberDineInOrderListReq) (*resp.GetMemberDineInOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBillRepo := repository.NewSaleBillRepo(db)
	h5OrderRepo := repository.NewH5OrderRepo(db)

	// 构建查询条件
	var dbOptions []repository.DBOption
	dbOptions = append(dbOptions,
		repository.CommonRepo.WhereBySoftDelete(),
		repository.CommonRepo.WhereByMemberUuid(ctx.GetMemberUuid()),
		repository.CommonRepo.WhereBySource(constant.SaleBillSourceMember),
		repository.CommonRepo.WhereByBillType(constant.SaleBillTypeInstant), // 堂食订单 BillType = 1
		repository.CommonRepo.SortWithCreateTime("desc"),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "SaleOrders",
				Args: []any{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
			repository.WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []any{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
			repository.WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			repository.WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ImageFile",
			},
		),
	)

	// 状态过滤
	billStatuses, _, isPaid := constant.GetMemberDineInOrderStatusFilter(listReq.Status)
	if billStatuses != nil {
		dbOptions = append(dbOptions, repository.CommonRepo.WhereByMultipleStatus(billStatuses))
	}

	// 分页查询
	saleBills, total, err := saleBillRepo.GetSaleBillListPage(listReq.PageNo, listReq.PageSize, dbOptions...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 构建响应
	list := make([]resp.MemberDineInOrder, 0, len(saleBills))
	for _, saleBill := range saleBills {
		// 过滤支付状态
		if isPaid != nil {
			billIsPaid := saleBill.IsExistPaid()
			if *isPaid != billIsPaid {
				continue
			}
		}

		// 获取 H5 订单状态
		var h5Status uint
		h5Order, _ := h5OrderRepo.GetH5Order(
			h5OrderRepo.WhereSaleBillUuid(saleBill.Uuid),
			h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
		)
		if h5Order != nil {
			h5Status = h5Order.Status
		}

		// 计算商品数量和金额
		var num float64
		var productAmount float64
		productList := make([]resp.MemberOrderProduct, 0)

		saleOrder := saleBill.GetFirstSaleOrder()
		if saleOrder != nil {
			for i, product := range saleOrder.SaleOrderProducts {
				if product.IsPackageSubProduct() {
					continue // 跳过套餐子商品
				}
				num += product.Num
				productAmount += product.GetTotalPrice()

				// 只返回前3个商品
				if i < 3 {
					productList = append(productList, resp.MemberOrderProduct{
						LocaleName: product.GetLocaleName(),
						Num:        product.Num,
						TotalPrice: product.GetTotalPrice(),
						Image: func() string {
							if product.ImageFile == nil {
								return ""
							}
							return product.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
						}(),
					})
				}
			}
		}

		// 计算订单状态
		statusInfo := s.getMemberDineInOrderStatusInfo(saleBill, h5Status, ctx.GetLanguage())

		list = append(list, resp.MemberDineInOrder{
			SaleBillUuid:  saleBill.Uuid,
			CompanyName:   ctx.GetCompany().Name,
			SerialNo:      saleBill.SerialNo,
			OrderNo:       saleBill.OrderNo,
			StatusInfo:    statusInfo,
			DiningMethod:  saleBill.DiningMethod,
			Num:           num,
			Amount:        saleBill.Amount,
			ProductAmount: productAmount,
			CreateTime:    saleBill.CreateTime,
			ProductList:   productList,
		})
	}

	return &resp.GetMemberDineInOrderListResp{
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
		List: list,
	}, nil
}

// GetMemberDineInOrderDetail 获取会员端堂食订单详情
func (s *orderSrv) GetMemberDineInOrderDetail(ctx context.Context, detailReq req.GetMemberDineInOrderDetailReq) (*resp.GetMemberDineInOrderDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	h5OrderRepo := repository.NewH5OrderRepo(db)
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)

	// 获取销售账单完整信息
	saleBill, err := orderRepo.GetSaleBillAllInfo(detailReq.SaleBillUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取订单失败")
	}

	// 验证是否是当前会员的订单（通过 SaleOrder.ConsumerUuid 验证）
	saleOrder := saleBill.GetFirstSaleOrder()
	if saleOrder == nil || saleOrder.ConsumerUuid != ctx.GetMemberUuid() {
		return nil, errors.New("无权查看此订单")
	}

	// 验证是否是堂食订单
	if saleBill.Source != constant.SaleBillSourceMember || saleBill.BillType != constant.SaleBillTypeInstant {
		return nil, errors.New("订单类型错误")
	}

	// 获取 H5 订单状态
	var h5Status uint
	h5Order, _ := h5OrderRepo.GetH5Order(
		h5OrderRepo.WhereSaleBillUuid(saleBill.Uuid),
		h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
	)
	if h5Order != nil {
		h5Status = h5Order.Status
	}

	// 获取商品列表
	productList := make([]resp.MemberOrderProduct, 0)
	for _, product := range saleOrder.SaleOrderProducts {
		if product.IsPackageSubProduct() {
			continue // 跳过套餐子商品
		}
		productList = append(productList, resp.MemberOrderProduct{
			LocaleName:          product.GetLocaleName(),
			LocaleAttributeName: product.GetAttributeName(),
			Num:                 product.Num,
			TotalPrice:          product.GetTotalPrice(),
			OriginTotalPrice:    product.GetTotalPriceOrigin(),
			Image: func() string {
				if product.ImageFile == nil {
					return ""
				}
				return product.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
			}(),
		})
	}

	// 获取支付信息（通过 SaleOrder 关联的 PaymentOrders 获取）
	var payTime int64
	var paidAmount float64
	var paymentMethodName string
	paymentOrders, _ := paymentOrderRepo.GetPaymentOrderList(
		paymentOrderRepo.WhereRelatedUuid(saleOrder.Uuid),
		paymentOrderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
		paymentOrderRepo.WhereStatus(constant.PaymentOrderStatusPaid),
		paymentOrderRepo.WithPaymentMethod(),
	)
	if len(paymentOrders) > 0 {
		// 获取第一个已支付的支付订单
		paymentOrder := paymentOrders[0]
		payTime = paymentOrder.CreateTime // PaymentOrder 使用 CreateTime 作为支付时间
		paidAmount = paymentOrder.Amount
		if paymentOrder.PaymentMethod != nil {
			paymentMethodName = paymentOrder.PaymentMethod.GetName()
		}
	}

	// 计算剩余支付时间
	var remainingPaymentTime int64
	if saleBill.Status == constant.SaleBillStatusPending && !saleBill.IsExistPaid() {
		// 假设支付超时时间为15分钟
		paymentTimeout := int64(15 * 60)
		remaining := paymentTimeout - (time.Now().Unix() - saleBill.CreateTime)
		if remaining > 0 {
			remainingPaymentTime = remaining
		}
	}

	// 获取商家信息
	company := ctx.GetCompany()
	var companyAddress, companyPhone string
	if company.CompanySetting != nil {
		companyAddress = company.CompanySetting.Address
		companyPhone = company.CompanySetting.LinkPhone
	}

	// 获取取消时间（从 SaleOrder 获取 DeleteTime 作为取消时间）
	var cancelTime int64
	if saleBill.Status == constant.SaleBillStatusCanceled {
		cancelTime = saleOrder.DeleteTime
	}

	return &resp.GetMemberDineInOrderDetailResp{
		SaleBillUuid:         saleBill.Uuid,
		CompanyName:          company.Name,
		CompanyAddress:       companyAddress,
		CompanyPhone:         companyPhone,
		SerialNo:             saleBill.SerialNo,
		OrderNo:              saleBill.OrderNo,
		StatusInfo:           s.getMemberDineInOrderStatusInfo(saleBill, h5Status, ctx.GetLanguage()),
		DiningMethod:         saleBill.DiningMethod,
		Remark:               saleBill.Remark,
		CreateTime:           saleBill.CreateTime,
		PayTime:              payTime,
		FinishTime:           saleBill.FinishTime,
		CancelTime:           cancelTime,
		RemainingPaymentTime: remainingPaymentTime,
		AmountInfo: resp.MemberDineInOrderAmountInfo{
			ProductAmount:     saleBill.ProductAmount,
			DiscountAmount:    saleBill.CustomDiscountFee + saleBill.MemberDiscountFee, // 折扣金额 = 自定义折扣 + 会员折扣
			ServiceFee:        saleBill.ServiceFee,
			TaxFee:            saleBill.TaxFee,
			PackageFee:        0, // SaleBill 没有 PackageFee 字段
			Amount:            saleBill.Amount,
			PaidAmount:        paidAmount,
			PaymentMethodName: paymentMethodName,
		},
		ProductList: resp.MemberProductList{
			List:          productList,
			ProductAmount: saleBill.ProductAmount,
		},
	}, nil
}

// getMemberDineInOrderStatusInfo 获取堂食订单状态信息
func (s *orderSrv) getMemberDineInOrderStatusInfo(saleBill *model.SaleBill, h5Status uint, language string) resp.MemberDineInOrderStatusInfo {
	var status string
	var statusText string

	switch saleBill.Status {
	case constant.SaleBillStatusCanceled:
		status = constant.MemberDineInOrderStatusCancelled
		statusText = i18n.Translate(language, "已取消")
	case constant.SaleBillStatusComplete:
		status = constant.MemberDineInOrderStatusCompleted
		statusText = i18n.Translate(language, "已完成")
	case constant.SaleBillStatusPending:
		if saleBill.IsExistPaid() {
			status = constant.MemberDineInOrderStatusInProgress
			statusText = i18n.Translate(language, "进行中")
		} else {
			status = constant.MemberDineInOrderStatusUnpaid
			statusText = i18n.Translate(language, "待支付")
		}
	}

	return resp.MemberDineInOrderStatusInfo{
		Status:       status,
		StatusText:   statusText,
		H5Status:     h5Status,
		H5StatusText: s.getH5OrderStatusText(h5Status, language),
	}
}

// getH5OrderStatusText 获取 H5 订单状态文字
func (s *orderSrv) getH5OrderStatusText(status uint, language string) string {
	switch status {
	case constant.H5OrderStatusOrder:
		return i18n.Translate(language, "待接单")
	case constant.H5OrderStatusAccepted:
		return i18n.Translate(language, "已接单")
	case constant.H5OrderStatusRejected:
		return i18n.Translate(language, "已拒单")
	default:
		return ""
	}
}


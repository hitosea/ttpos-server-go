package service

import (
	contexts "context"
	builtinerrors "errors"
	"fmt"
	"sort"
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
	CreateMemberOrder(ctx context.Context, req req.CreateMemberOrderReq) (*resp.CreateMemberOrderResp, *resp.OrderCheckServiceRes, error)    // 创建会员端外送订单
	CreateMemberDineInOrder(ctx context.Context, req req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error)                   // 创建会员端堂食订单
	SubmitMemberDineInOrder(ctx context.Context, req req.SubmitMemberDineInOrderReq) error                                                   // 先下单后付模式：提交堂食订单到收银机
	GetDineInOrderFormInfo(ctx context.Context, req req.GetDineInOrderFormInfoReq) (*resp.DineInOrderFormResp, error)                        // 获取堂食订单提交表单信息
	SetDineInOrderDiningMethod(ctx context.Context, req req.SetDineInOrderDiningMethodReq) error                                             // 设置堂食订单用餐方式
	PayDineInOrder(ctx context.Context, req req.PayDineInOrderReq) error                                                                     // 堂食订单提交支付
	GetDineInOrderPayInfo(ctx context.Context, req req.GetDineInOrderPayInfoReq) (*resp.DineInOrderPaymentInfoResp, error)                   // 获取堂食订单支付信息
	GetDineInOrderPayStatus(ctx context.Context, req req.GetDineInOrderPayStatusReq) (*resp.DineInOrderPaymentStatusResp, error)             // 获取堂食订单支付状态
	GetMemberOrderFormInfo(ctx context.Context, req req.GetMemberOrderFormInfoReq) (*resp.CreateMemberOrderResp, error)                      // 获取订单提交表单信息
	SetMemberOrderAddress(ctx context.Context, req member_req.SetMemberOrderAddressReq) (*resp.CreateMemberOrderResp, error)                 // 设置会员端订单地址
	PayMemberOrder(ctx context.Context, request member_req.PayMemberOrderReq) error                                                          // 会员端订单提交支付，状态变为待支付
	GetMemberOrderPayInfo(ctx context.Context, request member_req.GetMemberOrderPayInfoReq) (*resp.MemberOrderPaymentInfoResp, error)        // 会员端订单获取支付信息
	GetMemberOrderPayStatus(ctx context.Context, request member_req.GetMemberOrderPayStatusReq) (*resp.MemberOrderPaymentStatusResp, error)  // 会员端订单获取支付状态
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
	CheckDineInOrder(ctx context.Context, req req.CheckDineInOrderReq) (*resp.OrderCheckServiceRes, error)                               // 会员端堂食订单结算前检查
	GetMemberDineInOrderList(ctx context.Context, req req.MemberDineInOrderListReq) (*resp.GetMemberDineInOrderListResp, error)          // 获取会员端堂食订单列表
	GetMemberDineInOrderDetail(ctx context.Context, req req.GetMemberDineInOrderDetailReq) (*resp.GetMemberDineInOrderDetailResp, error) // 获取会员端堂食订单详情
	CancelMemberDineInOrder(ctx context.Context, req req.CancelMemberDineInOrderReq) error                                               // 取消会员端堂食订单（仅未付款）

	// 测试用接口
	MockPayDineInOrderCallback(ctx context.Context, req req.MockPayDineInOrderCallbackReq) error // 模拟堂食订单支付完成回调（仅用于测试）
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

// CheckDineInOrder 会员端堂食订单结算前检查
// 参考收银机 OrderCheck，在会员端点击"去结算"时校验：
// 消费税变动、服务费变动、库存、价格变动、规格、商品变动（被删除、下架）、限购等
func (s *orderSrv) CheckDineInOrder(ctx context.Context, request req.CheckDineInOrderReq) (*resp.OrderCheckServiceRes, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())

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

	ctx.Log().Debug("会员端堂食订单结算前检查")

	// 获取所有商品，用于检查限购和商品变动
	saleOrderProductAll := saleBill.GetSaleOrderProductAll()

	// 对商品进行结账检查: 检查商品是否删除、下架、库存是否充足、规格价格变动、小料的价格变动、超过限购
	// ignoreMust=true: 会员端堂食订单不检查必点商品
	checkServiceRes, errCheck := s.checkOrder(ctx, true, db, request.SaleBillUuid, 0, saleOrderProductAll, WithCheckTypeCheckout())
	if errCheck != nil {
		return nil, errors.WithMessage(errCheck, "订单检查失败")
	}
	if checkServiceRes != nil {
		return checkServiceRes, nil
	}

	// 检查含税未含税是否改变、服务费类型是否改变、服务费是否改变、服务费比例是否改变
	res, newSetting, err := s.checkSaleBillSettingChanged(ctx, saleBill)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	if res != nil {
		// 如果账单快照设置变化，更新销售订单的金额
		shopCartInfo, err := repository.NewOrderRepo(db).GetOrderCartInfo(request.SaleBillUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		freshSaleBill := shopCartInfo.SaleBill
		if err := s.CalcAndSaveSaleBill(ctx, db, freshSaleBill, model.WithLatestPrice(), model.WithSaleBillSetting(newSetting)); err != nil {
			return nil, errors.WithMessage(err, "更新订单金额失败")
		}
		return res, nil
	}

	return nil, nil
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
		// 跳过已删除的商品和套餐子商品（套餐子商品挂在套餐主商品下）
		if saleOrderProduct.DeleteTime != 0 || saleOrderProduct.IsPackageSubProduct() {
			continue
		}
		product := resp.DineInProduct{
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
			ProductType:        uint(saleOrderProduct.ProductType),
			PackageProductList: resp.PackageProductList{List: make([]resp.PackageProduct, 0)},
		}
		// 套餐商品：挂载子商品列表
		if saleOrderProduct.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(saleOrderProduct.Uuid)
			for _, sub := range subProducts {
				if sub.DeleteTime != 0 {
					continue
				}
				product.PackageProductList.List = append(product.PackageProductList.List, resp.PackageProduct{
					Uuid:                sub.Uuid,
					LocaleName:          sub.GetLocaleName(),
					LocaleAttributeName: sub.GetAttributeName(), // 包含规格+小料+属性
					Num:                 sub.Num,
					UnitNum:             sub.GetProductNum(),
					AddPrice:            sub.AddPrice,
				})
			}
		}
		products = append(products, product)
	}

	// 构建金额信息
	amountInfo := resp.DineInAmountInfo{
		ProductAmount:  saleOrder.ProductOriginalAmount, // 商品金额
		TaxAmount:      saleOrder.TaxFee,                // 消费税
		ServiceAmount:  saleOrder.ServiceFee,            // 服务费
		MemberDiscount: saleOrder.MemberDiscountFee,     // 会员折扣
		TotalAmount:    saleOrder.Amount,                // 应付金额（已包含税费和服务费）
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

	// 读取门店配置判断先下单后付模式
	isOrderFirstPayLater := false
	storeScanOrderSetting, settingErr := s.settingSrv.GetStoreScanOrderSetting(ctx)
	if settingErr == nil && storeScanOrderSetting.IsOrderFirstPayLater == constant.OrderFirstPayLaterYes {
		isOrderFirstPayLater = true
	}

	return &resp.DineInOrderFormResp{
		SaleBillUuid:         saleBill.Uuid,
		SaleOrderUuid:        saleOrder.Uuid,
		DiningMethod:         saleBill.DiningMethod,
		ProductList:          resp.DineInProductList{List: products},
		AmountInfo:           amountInfo,
		PaymentMethods:       resp.PaymentMethodList{List: payList},
		Remark:               saleBill.Remark,
		IsOrderFirstPayLater: isOrderFirstPayLater,
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

	// 设置用餐方式并更新商品税率
	saleBill.SetTakeoutSaleBill(request.DiningMethod)

	// 重新计算金额并保存
	if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
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
		return errors.NewWithCode(constant.CodeOrderAmountLessThan1, "订单金额小于1，无法支付")
	}

	// 验证支付方式
	if request.PaymentMethodUuid == 0 {
		return errors.New("请选择支付方式")
	}

	// 记录提交支付时间（仅首次提交时记录）
	if saleBill.SubmitPayTime == 0 {
		saleBill.SubmitPayTime = time.Now().Unix()
	}

	// 更新订单备注和提交支付时间
	if request.Remark != "" {
		saleBill.Remark = request.Remark
	}
	if err := repository.NewSaleBillRepo(db).UpdateSaleBill(saleBill); err != nil {
		return errors.WithMessage(err, "更新订单失败")
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

	// 设置订单为"已取消"状态
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

	// 发布"订单取消"操作事件
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

		// 设置订单为"已取消"状态
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

	// 发布"整单取消"操作事件
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

	// 会员外送订单接单成功后，同步到 ERP（有接单场景）
	// 异步推送，失败不影响接单结果；通过 ErpSyncStatus 幂等控制避免重复推送
	utils.Go(func() {
		s.SyncMemberOrderToErp(ctx, memberSaleOrder.SaleBill, db)
	})

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

	// 发布"订单取消"操作事件
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

		// 发布"订单取消"操作事件
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
	request.ApplyCompatibleFields()
	if err := request.Validate(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 检查是否在营业时间内
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "获取门店业务设置失败")
	}
	companySetting := ctx.GetCompanySetting()
	if !utils.SetTimezone(companySetting.Timezone).IsWithinOpeningHours(businessSetting.OpeningHours) {
		return nil, errors.New(i18n.Translate(ctx.GetLanguage(), "店铺休息中"))
	}

	var result *resp.CreateInstantOrderResp
	if request.SaleBillUuid == 0 {
		// 新建订单
		result, err = s.createDineInOrder(ctx, request)
	} else {
		// 更新已有订单
		result, err = s.updateDineInOrder(ctx, request)
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SubmitMemberDineInOrder 先下单后付模式：提交堂食订单到收银机
// 会员端调用 create 后可多次加购，最终通过 submit 提交生成 H5 订单
func (s *orderSrv) SubmitMemberDineInOrder(ctx context.Context, request req.SubmitMemberDineInOrderReq) error {
	// 分布式锁：防止并发双击创建重复 H5 订单
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	// 1. 检查门店配置：必须启用"先下单后付"模式
	storeScanOrderSetting, err := s.settingSrv.GetStoreScanOrderSetting(ctx)
	if err != nil {
		return errors.WithMessage(err, "获取门店配置失败")
	}
	if storeScanOrderSetting.IsOrderFirstPayLater != constant.OrderFirstPayLaterYes {
		return errors.New("当前模式不支持此操作")
	}

	// 2. 获取订单并校验状态
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "获取订单信息失败")
	}
	if saleBill.Status != constant.SaleBillStatusPending {
		return errors.New("订单状态不允许提交")
	}
	if saleBill.SubmitPayTime > 0 {
		return errors.New("订单已提交，请勿重复操作")
	}

	// 3. 标记 sale_bill 为先下单后付款
	saleBill.IsOrderFirstPayLater = 1
	if err := repository.NewSaleBillRepo(db).UpdateSaleBill(saleBill); err != nil {
		return errors.WithMessage(err, "更新订单标记失败")
	}

	// 4. 创建 H5 订单
	if err := s.createH5OrderForOrderFirst(ctx, request.SaleBillUuid, request.SaleOrderUuid); err != nil {
		return err
	}

	// 5. 尝试自动接单（失败不影响 submit 结果）
	s.autoAcceptOrderFirst(ctx, request.SaleBillUuid)

	return nil
}

// autoAcceptOrderFirst 先下单后付模式：尝试自动接单
// 复用与支付完成事件中 autoAcceptMemberDineInOrder 相同的判断逻辑：
// - 关闭 H5 接单功能时直接自动接单
// - 开启时根据接单设置（金额限额）判断
// 自动接单失败不影响 submit 结果，订单保持待接单状态等待手动处理
func (s *orderSrv) autoAcceptOrderFirst(ctx context.Context, saleBillUuid uint64) {
	db := s.dbm.GetDB(ctx.GetDbId())
	companySetting := ctx.GetCompanySetting()

	// 判断是否允许自动接单
	if !companySetting.GetIsOpenH5Order() {
		// 关闭 H5 接单功能：直接自动接单
		ctx.Log().Info("autoAcceptOrderFirst, h5 order disabled, auto accept directly",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Uint64("sale_bill_uuid", saleBillUuid))
	} else {
		// 开启 H5 接单功能：检查接单设置和金额限额
		acceptOrderSetting, err := s.settingSrv.GetAcceptOrderSetting(ctx)
		if err != nil {
			ctx.Log().Error("autoAcceptOrderFirst, GetAcceptOrderSetting failed",
				zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
				zap.Uint64("sale_bill_uuid", saleBillUuid),
				zap.Error(err))
			return
		}

		saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
		if err != nil {
			ctx.Log().Error("autoAcceptOrderFirst, GetSaleBillAllInfo failed",
				zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
				zap.Uint64("sale_bill_uuid", saleBillUuid),
				zap.Error(err))
			return
		}

		saleOrder := saleBill.GetFirstSaleOrder()
		if saleOrder == nil {
			return
		}
		totalPrice := saleBill.GetUnAcceptH5OrderProductTotalPrice(saleOrder.SaleOrderProducts)

		if !acceptOrderSetting.CanAutoOrder(totalPrice) {
			ctx.Log().Info("autoAcceptOrderFirst, auto accept not allowed",
				zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
				zap.Uint64("sale_bill_uuid", saleBillUuid),
				zap.Float64("totalPrice", totalPrice))
			return
		}
	}

	// 获取 H5 订单
	h5OrderRepo := repository.NewH5OrderRepo(db)
	h5Order, err := h5OrderRepo.GetH5Order(
		h5OrderRepo.WhereSaleBillUuid(saleBillUuid),
		h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
	)
	if err != nil || h5Order.Status != constant.H5OrderStatusOrder {
		ctx.Log().Warn("autoAcceptOrderFirst, h5 order not found or not pending",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Uint64("sale_bill_uuid", saleBillUuid),
			zap.Error(err))
		return
	}

	// 获取门店主收银设备 UUID，用于挂单归属
	deviceRepo := repository.NewDeviceRepo(db)
	mainDevice, err := deviceRepo.GetDevice(deviceRepo.WhereSource(constant.SourceCashier), deviceRepo.WhereMain())
	if err != nil {
		// 主设备不存在时尝试获取任意收银设备
		mainDevice, err = deviceRepo.GetDevice(deviceRepo.WhereSource(constant.SourceCashier))
		if err != nil {
			ctx.Log().Warn("autoAcceptOrderFirst, no cashier device found, skip auto accept",
				zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
				zap.Uint64("sale_bill_uuid", saleBillUuid))
			return
		}
	}
	ctx.SetDeviceUuid(mainDevice.Uuid)

	// 执行自动接单
	result, err := s.AcceptH5Order(ctx, h5Order.Uuid, true)
	if err != nil {
		ctx.Log().Error("autoAcceptOrderFirst, AcceptH5Order failed",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Uint64("h5_order_uuid", h5Order.Uuid),
			zap.Error(err))
		return
	}
	if result != nil {
		ctx.Log().Warn("autoAcceptOrderFirst, AcceptH5Order check failed",
			zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
			zap.Uint64("h5_order_uuid", h5Order.Uuid),
			zap.Any("result", result))
		return
	}

	ctx.Log().Info("autoAcceptOrderFirst, auto accept success",
		zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
		zap.Uint64("sale_bill_uuid", saleBillUuid),
		zap.Uint64("h5_order_uuid", h5Order.Uuid),
		zap.Uint64("device_uuid", mainDevice.Uuid))
}

// createH5OrderForOrderFirst 为"先下单后付"模式创建 H5 订单
// 复用支付完成事件中 createH5OrderForMemberDineIn 的核心逻辑，区别在于：
// 1. 不标记订单为已完成（SaleBill.Status 保持 Pending，因为还没付款）
// 2. 设置 submit_pay_time 使订单在列表中可见
// 自动接单由 autoAcceptOrderFirst 在外部处理
func (s *orderSrv) createH5OrderForOrderFirst(ctx context.Context, saleBillUuid uint64, saleOrderUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单信息（包含商品列表）
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "获取订单信息失败")
	}

	// 获取销售订单
	var saleOrder *model.SaleOrder
	for _, order := range saleBill.SaleOrders {
		if order.Uuid == saleOrderUuid {
			saleOrder = order
			break
		}
	}
	if saleOrder == nil {
		return errors.New("销售订单不存在")
	}

	// 获取销售订单商品
	saleOrderProducts := saleOrder.SaleOrderProducts
	if len(saleOrderProducts) == 0 {
		return errors.New("订单商品为空")
	}

	// 获取公司设置
	companySetting := ctx.GetCompanySetting()

	// 创建 h5_order 记录
	h5OrderUuid, _ := utils.GetID()
	now := time.Now().Unix()

	// 构建 h5_order_product 列表（仅包含主商品，套餐子商品随主商品一起处理）
	lang := ctx.GetLanguage()
	h5OrderProductList := make([]*model.H5OrderProduct, 0, len(saleOrderProducts))
	for _, product := range saleOrderProducts {
		// 跳过套餐子商品
		if product.PackageUuid > 0 {
			continue
		}
		h5OrderProductList = append(h5OrderProductList, &model.H5OrderProduct{
			// 快照信息
			Name:          product.MultiLanguageName.GetNameByLang(lang),
			Price:         product.GetFinalSalePrice(),
			SalePrice:     product.SalePrice,
			Num:           product.Num,
			AttributeText: product.GetAttributeNamesByLang(lang),
			Remark:        product.Remark,
			// 关联uuid
			SaleOrderProductUuid: product.Uuid,
			H5OrderUuid:          h5OrderUuid,
			SaleBillUuid:         saleBill.Uuid,
		})
	}

	h5Order := &model.H5Order{
		BaseModel: model.BaseModel{
			Uuid:       h5OrderUuid,
			CreateTime: now,
			UpdateTime: now,
		},
		DeskUuid:        0,                                // 会员端堂食订单无桌台
		SaleOrderUuid:   saleOrder.Uuid,                   // 销售订单uuid
		SaleBillUuid:    saleBill.Uuid,                    // 销售账单uuid
		DeskNo:          saleBill.SerialNo,                // 取餐号（使用 sale_bill.serial_no）
		Status:          constant.H5OrderStatusOrder,      // 状态：待接单
		OrderType:       constant.H5OrderTypeMemberDineIn, // 订单类型：会员端堂食订单
		IsAutoAccept:    0,                                // 非自动接单
		IsNeedAudit:     companySetting.IsOpenH5Order,     // 关闭扫码点餐接单则不需要审核，直接送厨
		OrderTime:       now,                              // 下单时间
		H5OrderProducts: h5OrderProductList,
	}

	// 保存到数据库：创建 H5 订单 + 更新 sale_order_product + 设置 submit_pay_time
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		h5OrderRepo := repository.NewH5OrderRepo(tx)
		saleOrderProductRepo := repository.NewSaleOrderProductRepo(tx)
		saleBillRepo := repository.NewSaleBillRepo(tx)

		// 创建 h5_order
		if _, err := h5OrderRepo.CreateH5Order(*h5Order); err != nil {
			return err
		}

		// 批量创建 h5_order_product
		for _, h5OrderProduct := range h5Order.H5OrderProducts {
			if _, err := h5OrderRepo.CreateH5OrderProduct(*h5OrderProduct); err != nil {
				return err
			}
		}

		// 批量更新 sale_order_product 的 h5_order_uuid 和 is_accept_order
		if err := saleOrderProductRepo.Update(map[string]any{
			"h5_order_uuid":   h5OrderUuid,
			"is_accept_order": constant.OrderProductIsAcceptOrderUnAccept,
		},
			repository.CommonRepo.WhereBySaleOrderUuid(saleOrder.Uuid),
			repository.CommonRepo.WhereBySoftDelete(),
		); err != nil {
			return err
		}

		// 设置 submit_pay_time，使订单在列表查询中可见（列表过滤条件：submit_pay_time > 0）
		saleBill.SubmitPayTime = now
		if err := saleBillRepo.UpdateSaleBill(saleBill); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err, "创建H5订单失败")
	}

	ctx.Log().Info("createH5OrderForOrderFirst, h5_order created",
		zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
		zap.Uint64("h5_order_uuid", h5OrderUuid),
		zap.Uint64("sale_bill_uuid", saleBillUuid),
		zap.String("desk_no", saleBill.SerialNo))

	return nil
}

// createDineInOrder 创建堂食订单（内部方法）
func (s *orderSrv) createDineInOrder(ctx context.Context, request req.CreateMemberDineInOrderReq) (*resp.CreateInstantOrderResp, error) {
	// 创建订单
	instantOrder, err := s.createInstantOrder(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取销售账单信息
	db := s.dbm.GetDBWithContext(ctx)
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
	db := s.dbm.GetDBWithContext(ctx)
	saleBillUuid := request.SaleBillUuid
	saleOrderUuid := request.SaleOrderUuid

	// 获取销售账单信息
	targetSaleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(saleBillUuid)
	if err != nil {
		// 订单不存在,则新建新订单
		// 新建订单
		return s.createDineInOrder(ctx, request)
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

	// 套餐修改 → 整体替换（删除旧套餐 + 新增新套餐）
	for key, product := range updateProducts {
		if product.ProductType == constant.ProductTypePackage {
			addProducts[key] = product
			deleteProducts[key] = updateOlderProducts[key]
			delete(updateProducts, key)
			delete(updateOlderProducts, key)
		}
	}

	// 执行删除
	for _, saleOrderProduct := range deleteProducts {
		saleOrderProduct.DeleteProduct()
		// 套餐主商品删除时，一并删除其子商品
		if saleOrderProduct.ProductType == constant.ProductTypePackage && len(targetSaleBill.SaleOrders) > 0 {
			for index := range targetSaleBill.SaleOrders[0].SaleOrderProducts {
				subProduct := targetSaleBill.SaleOrders[0].SaleOrderProducts[index]
				if subProduct.PackageUuid == saleOrderProduct.Uuid && subProduct.DeleteTime == 0 {
					subProduct.DeleteProduct()
				}
			}
		}
	}

	// 执行修改
	for key, product := range updateProducts {
		saleOrderProduct := updateOlderProducts[key]
		saleOrderProduct.Num = product.Num
		saleOrderProduct.SetUpdate() // 标记该商品需要更新
	}

	// 执行新增
	if err := s.addDineInProducts(ctx, saleBillUuid, saleOrderUuid, s.toProductParamsSlice(addProducts), targetSaleBill); err != nil {
		return nil, err
	}

	return &resp.CreateInstantOrderResp{
		SaleBillUuid:  saleBillUuid,
		SaleOrderUuid: saleOrderUuid,
	}, nil
}

// diffDineInProducts 计算堂食订单商品差异
// 返回：新增商品、删除商品、修改商品、修改前商品
func (s *orderSrv) diffDineInProducts(ctx context.Context, db *gorm.DB, products []req.ProductParams, saleBill *model.SaleBill) (
	addProducts map[string]req.ProductParams,
	deleteProducts map[string]*model.SaleOrderProduct,
	updateProducts map[string]req.ProductParams,
	updateOlderProducts map[string]*model.SaleOrderProduct,
	err error,
) {
	// 构建提交商品映射
	commitProductMap := make(map[string]req.ProductParams)
	for index := range products {
		product := products[index]
		key, keyErr := dineInProductKey(ctx, db, product)
		if keyErr != nil {
			err = errors.WithMessage(keyErr)
			return
		}
		commitProductMap[key] = product
	}

	// 构建已有商品映射（过滤已软删除的商品和套餐子商品）
	olderProductMap := make(map[string]*model.SaleOrderProduct)
	if len(saleBill.SaleOrders) > 0 {
		for index := range saleBill.SaleOrders[0].SaleOrderProducts {
			product := saleBill.SaleOrders[0].SaleOrderProducts[index]
			// 跳过已删除的商品
			if product.DeleteTime != 0 {
				continue
			}
			// 跳过套餐子商品（子商品由套餐主商品统一管理）
			if product.ProductType == constant.ProductTypePackageSubProduct {
				continue
			}
			key := product.ProductKey()
			// 套餐主商品用 product_package_uuid + 子商品签名生成 key，与请求侧格式对齐
			if product.ProductType == constant.ProductTypePackage {
				key = fmt.Sprintf("pkg:%d-%s", product.ProductPackageUuid, packageSubProductSignFromModel(product))
			}
			olderProductMap[key] = product
		}
	}

	addProducts = make(map[string]req.ProductParams)
	deleteProducts = make(map[string]*model.SaleOrderProduct)
	updateProducts = make(map[string]req.ProductParams)
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
func (s *orderSrv) addDineInProducts(ctx context.Context, saleBillUuid, saleOrderUuid uint64, products []req.ProductParams, targetSaleBill *model.SaleBill) error {
	db := s.dbm.GetDBWithContext(ctx)
	productParams := make([]req.ProductParams, 0, len(products))
	for _, product := range products {
		product.Operation = "add"
		// 套餐商品：通过 product_package_uuid 查 DB 获取套餐的 BOM UUID
		if product.ProductType == constant.ProductTypePackage {
			productPackage, err := repository.NewProductPackageRepo(db).GetProductPackage(
				repository.CommonRepo.WhereByUuid(product.ProductPackageUuid),
				repository.CommonRepo.WhereBySoftDelete(),
				repository.NewProductPackageRepo(db).WithProductBoms(
					repository.CommonRepo.WhereBySoftDelete(),
				),
			)
			if err != nil {
				return errors.WithMessage(err, "查询套餐信息失败")
			}
			productPackageFlavorBomUuid := func() uint64 {
				if len(productPackage.ProductBoms) == 0 {
					return 0
				}
				return productPackage.ProductBoms[0].Uuid
			}()
			if productPackageFlavorBomUuid == 0 {
				return errors.WithMessage(errors.New("套餐商品规格不存在"), "套餐商品规格不存在")
			}
			product.FlavorProductBomUuid = productPackageFlavorBomUuid

			// 转换子商品参数
			subProductParams := make([]req.ProductParams, 0, len(product.Products))
			for _, sub := range product.Products {
				subParam := req.ProductParams{
					FlavorProductBomUuid:            sub.FlavorUuid,
					Num:                             sub.Num,
					UnitNum:                         sub.UnitNum,
					ProductPackageAttributeUuidList: sub.AttributeUuidList,
					ProductPackageGroupUuid:         sub.ProductPackageGroupUuid,
					Operation:                       "add",
					AddPrice:                        sub.AddPrice,
				}
				subProductParams = append(subProductParams, subParam)
			}
			product.SetIsPackageProduct(subProductParams)
		}
		productParams = append(productParams, product)
	}

	params := req.ProductAddReq{
		SaleBillUuid:   saleBillUuid,
		SaleOrderUuid:  saleOrderUuid,
		Products:       productParams,
		IsMemberAdd:    false, // 堂食订单不应用外送折扣率
		IsMemberDineIn: true,  // 会员堂食下单. 标记商品都是结账减库存
	}
	return s.ActionAdd(ctx, params, targetSaleBill)
}

// toProductParamsSlice 将 map 转换为 slice
func (s *orderSrv) toProductParamsSlice(products map[string]req.ProductParams) []req.ProductParams {
	result := make([]req.ProductParams, 0, len(products))
	for _, product := range products {
		result = append(result, product)
	}
	return result
}

// packageSubProductSignFromModel 从已有的 SaleOrderProduct 生成套餐子商品签名
// 格式与 dineInProductKey() 套餐分支一致，用于 diffDineInProducts key 对齐
func packageSubProductSignFromModel(product *model.SaleOrderProduct) string {
	type subProduct struct {
		FlavorUuid              uint64   `json:"flavor_uuid"`
		AttributeUuid           []uint64 `json:"attribute_uuid"`
		ProductPackageGroupUuid uint64   `json:"product_package_group_uuid"`
		Num                     float64  `json:"num"`
		UnitNum                 float64  `json:"unit_num"`
	}
	subProducts := make([]subProduct, 0)
	utils.FromJson(product.PackageSubProductParams, &subProducts)
	sort.Slice(subProducts, func(i, j int) bool {
		if subProducts[i].ProductPackageGroupUuid != subProducts[j].ProductPackageGroupUuid {
			return subProducts[i].ProductPackageGroupUuid < subProducts[j].ProductPackageGroupUuid
		}
		return subProducts[i].FlavorUuid < subProducts[j].FlavorUuid
	})
	parts := make([]string, 0, len(subProducts))
	for _, sp := range subProducts {
		parts = append(parts, fmt.Sprintf("%d:%d:%.2f:%.2f", sp.ProductPackageGroupUuid, sp.FlavorUuid, sp.Num, sp.UnitNum))
	}
	return strings.Join(parts, "|")
}

// dineInProductKey 生成堂食订单商品的 diff key
// 普通商品：FlavorProductBomUuid-属性-加料
// 套餐商品：pkg:ProductPackageUuid-子商品签名
func dineInProductKey(ctx context.Context, db *gorm.DB, product req.ProductParams) (string, error) {
	if product.ProductType == constant.ProductTypePackage {
		// 套餐：用 product_package_uuid + 子商品签名
		type subKey struct {
			FlavorUuid              uint64
			ProductPackageGroupUuid uint64
			Num                     float64
			UnitNum                 float64
		}
		subKeys := make([]subKey, 0, len(product.Products))
		for _, sub := range product.Products {
			subKeys = append(subKeys, subKey{
				FlavorUuid:              sub.FlavorUuid,
				ProductPackageGroupUuid: sub.ProductPackageGroupUuid,
				Num:                     sub.Num,
				UnitNum:                 sub.UnitNum,
			})
		}
		sort.Slice(subKeys, func(i, j int) bool {
			if subKeys[i].ProductPackageGroupUuid != subKeys[j].ProductPackageGroupUuid {
				return subKeys[i].ProductPackageGroupUuid < subKeys[j].ProductPackageGroupUuid
			}
			return subKeys[i].FlavorUuid < subKeys[j].FlavorUuid
		})
		parts := make([]string, 0, len(subKeys))
		for _, sk := range subKeys {
			parts = append(parts, fmt.Sprintf("%d:%d:%.2f:%.2f", sk.ProductPackageGroupUuid, sk.FlavorUuid, sk.Num, sk.UnitNum))
		}
		return fmt.Sprintf("pkg:%d-%s", product.ProductPackageUuid, strings.Join(parts, "|")), nil
	}

	// 普通商品：FlavorProductBomUuid-属性-加料
	attributeUuidList := make([]uint64, 0)
	if len(product.ProductPackageAttributeUuidList) > 0 {
		productPackageAttributes, err := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(ctx.GetCompanyUuid(), product.ProductPackageAttributeUuidList)
		if err != nil {
			return "", errors.WithMessage(err)
		}
		for _, attribute := range productPackageAttributes {
			attributeUuidList = append(attributeUuidList, attribute.AttributeUuid)
		}
	}
	sort.Slice(attributeUuidList, func(i, j int) bool {
		return attributeUuidList[i] < attributeUuidList[j]
	})
	sauceList := make([]uint64, 0, len(product.SauceProductBomUuidList))
	sauceList = append(sauceList, product.SauceProductBomUuidList...)
	sort.Slice(sauceList, func(i, j int) bool {
		return sauceList[i] < sauceList[j]
	})

	attrStrs := make([]string, 0, len(attributeUuidList))
	for _, a := range attributeUuidList {
		attrStrs = append(attrStrs, fmt.Sprintf("%d", a))
	}
	sauceStrs := make([]string, 0, len(sauceList))
	for _, s := range sauceList {
		sauceStrs = append(sauceStrs, fmt.Sprintf("%d", s))
	}
	return fmt.Sprintf("%d-%s-%s", product.FlavorProductBomUuid, strings.Join(attrStrs, ","), strings.Join(sauceStrs, ",")), nil
}

// ==================== 会员端堂食订单管理 ====================

// GetMemberDineInOrderList 获取会员端堂食订单列表
func (s *orderSrv) GetMemberDineInOrderList(ctx context.Context, listReq req.MemberDineInOrderListReq) (*resp.GetMemberDineInOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	saleBillRepo := repository.NewSaleBillRepo(db)
	h5OrderRepo := repository.NewH5OrderRepo(db)
	productionRepo := repository.NewProductionRepo(db)

	// 构建查询条件
	billStatuses, h5OrderStatuses, isPaid := constant.GetMemberDineInOrderStatusFilter(listReq.Status)

	// "先下单后付"模式兼容：
	// - "进行中"查询需要同时包含 Pending 状态（先下单后付的订单 Status=Pending 但有 H5 订单）
	// - "待支付"查询也需要内存过滤掉已有 H5 订单的（先下单后付的不是"待支付"）
	if listReq.Status == constant.MemberDineInOrderStatusInProgress {
		billStatuses = append(billStatuses, constant.SaleBillStatusPending)
	}

	dbOptions := s.buildDineInOrderListQueryOptions(ctx, billStatuses)

	if listReq.Keyword != "" {
		// 关键字搜索（按菜名或订单号）
		dbOptions = append(dbOptions, saleBillRepo.WhereKeyword(listReq.Keyword, ctx.GetLanguage()))
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
		if isPaid != nil && *isPaid != saleBill.IsExistPaid() {
			// 如果需要过滤支付状态时,过滤掉已经支付的或者未支付的. 使用场景: 仅查询"待支付"的订单时需要
			continue
		}

		// 获取 H5 订单
		h5Order := s.getH5OrderForMemberDineIn(h5OrderRepo, saleBill.Uuid)

		// "先下单后付"模式内存过滤：通过 IsOrderFirstPayLater 标记区分
		// - "待支付"列表：排除先下单后付的订单（它们属于"进行中"）
		// - "进行中"列表：Pending 状态必须是先下单后付的订单才能归入进行中
		if listReq.Status == constant.MemberDineInOrderStatusUnpaid && saleBill.IsOrderFirstPayLater == constant.OrderFirstPayLaterYes && saleBill.Status == constant.SaleBillStatusPending {
			continue // 先下单后付的订单不显示在"待支付"列表
		}
		if listReq.Status == constant.MemberDineInOrderStatusInProgress && saleBill.Status == constant.SaleBillStatusPending && saleBill.IsOrderFirstPayLater != constant.OrderFirstPayLaterYes {
			continue // Pending 状态非先下单后付的是普通待支付订单，不属于"进行中"
		}

		// 获取生产单完成状态
		isProductionFinished, _ := productionRepo.IsProductionFinishedBySaleBillUuid(saleBill.Uuid)

		// H5 订单状态内存过滤（结合生产单完成状态）
		// 对于"进行中"+ Pending 状态的先下单后付订单，跳过 h5OrderStatuses 过滤（已在上面过滤）
		if saleBill.Status != constant.SaleBillStatusPending {
			if !s.filterDineInOrderByH5Status(h5Order, h5OrderStatuses, listReq.Status, isProductionFinished) {
				continue
			}
		} else if listReq.Status == constant.MemberDineInOrderStatusInProgress {
			// 先下单后付订单的进行中过滤：H5 订单必须是待接单或已接单状态
			if h5Order == nil {
				continue
			}
			if h5Order.Status != constant.H5OrderStatusOrder && h5Order.Status != constant.H5OrderStatusAccepted {
				continue
			}
			// 已接单且生产单全部完成的不在"进行中"
			if h5Order.Status == constant.H5OrderStatusAccepted && isProductionFinished {
				continue
			}
		}

		// 构建订单响应项
		orderItem := s.buildMemberDineInOrderItem(ctx, *saleBill, h5Order, isProductionFinished)
		list = append(list, orderItem)
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

// buildDineInOrderListQueryOptions 构建会员端堂食订单列表查询条件
func (s *orderSrv) buildDineInOrderListQueryOptions(ctx context.Context, billStatuses []uint) []repository.DBOption {
	dbOptions := []repository.DBOption{
		repository.CommonRepo.WhereBySoftDelete(),
		// 使用 consumer_uuid 字段过滤当前会员的订单（sale_bill 表使用 consumer_uuid 而非 member_uuid）
		func(db *gorm.DB) *gorm.DB {
			return db.Where("consumer_uuid = ?", ctx.GetMemberUuid())
		},
		repository.CommonRepo.WhereBySource(constant.SaleBillSourceMember),
		repository.CommonRepo.WhereByBillType(constant.SaleBillTypeInstant), // 堂食订单 BillType = 1
		// 只显示已提交支付的订单（调用过 /member/order/dine_in/pay 接口）
		func(db *gorm.DB) *gorm.DB {
			return db.Where("submit_pay_time > 0")
		},
		repository.CommonRepo.SortWithCreateTime("desc"),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "SaleOrders",
				Args: []any{
					repository.CommonRepo.DBOption(repository.CommonRepo.WhereBySoftDelete()),
				},
			},
			repository.WithPreload{
				Query: "SaleOrders.ReturnOrders",
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
	}

	// 状态过滤
	if billStatuses != nil {
		dbOptions = append(dbOptions, repository.CommonRepo.WhereByMultipleStatus(billStatuses))
	}

	return dbOptions
}

// getH5OrderForMemberDineIn 获取会员端堂食订单的 H5 订单
// 如果没有找到，返回 nil（区分未 submit 的订单）
func (s *orderSrv) getH5OrderForMemberDineIn(h5OrderRepo repository.IH5OrderRepo, saleBillUuid uint64) *model.H5Order {
	h5Order, err := h5OrderRepo.GetH5Order(
		h5OrderRepo.WhereSaleBillUuid(saleBillUuid),
		h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
	)
	if err != nil || h5Order.Uuid == 0 {
		return nil
	}
	return h5Order
}

// filterDineInOrderByH5Status H5 订单状态内存过滤（用于区分"进行中"和"已完成"）
// 返回 true 表示订单应该包含在结果中，false 表示应该过滤掉
// h5OrderStatuses: 允许的 H5 订单状态列表（"进行中"状态需要匹配）
// requestStatus: 请求的状态过滤参数
// isProductionFinished: 生产单是否全部完成
func (s *orderSrv) filterDineInOrderByH5Status(h5Order *model.H5Order, h5OrderStatuses []uint, requestStatus string, isProductionFinished bool) bool {
	// 获取 H5 订单状态
	var h5Status uint
	if h5Order != nil {
		h5Status = h5Order.Status
	}

	// "进行中"状态查询：需要 H5 订单存在且状态匹配
	if len(h5OrderStatuses) > 0 {
		if h5Order == nil {
			// 没有 H5 订单（未开启厨显），不符合"进行中"条件
			return false
		}
		// 检查 H5 订单状态是否在允许的状态列表中
		for _, allowedStatus := range h5OrderStatuses {
			if h5Status == allowedStatus {
				// 如果是已接单状态（备餐中），还需要检查生产单是否全部完成
				// 如果生产单全部完成，说明已经出餐，不应该在"进行中"列表中
				if h5Status == constant.H5OrderStatusAccepted && isProductionFinished {
					return false // 生产单已完成，不在"进行中"
				}
				return true
			}
		}
		return false
	}

	// "已完成"状态查询：排除"进行中"的订单
	if requestStatus == constant.MemberDineInOrderStatusCompleted {
		if h5Order != nil {
			// 待接单状态：一定是进行中，排除
			if h5Status == constant.H5OrderStatusOrder {
				return false
			}
			// 已接单状态（备餐中）：需要检查生产单是否完成
			if h5Status == constant.H5OrderStatusAccepted && !isProductionFinished {
				return false // 生产单未完成，还在"备餐中"，排除
			}
		}
		// H5 订单不存在或已拒单，属于"已完成"
	}

	return true
}

// buildMemberDineInOrderProductList 构建会员端堂食订单商品列表
// 返回：商品数量、商品金额、商品列表（最多3个）
func (s *orderSrv) buildMemberDineInOrderProductList(ctx context.Context, saleBill model.SaleBill) (float64, float64, []resp.MemberOrderProduct) {
	var num float64
	var productAmount float64
	productList := make([]resp.MemberOrderProduct, 0)
	displayCount := 0

	saleOrder := saleBill.GetFirstSaleOrder()
	if saleOrder == nil {
		return num, productAmount, productList
	}

	for _, product := range saleOrder.SaleOrderProducts {
		if product.IsPackageSubProduct() {
			continue // 跳过套餐子商品
		}
		num += product.Num
		productAmount += product.GetTotalPrice()

		// 只返回前3个商品
		if displayCount < 3 {
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
			displayCount++
		}
	}

	return num, productAmount, productList
}

// buildMemberDineInOrderItem 构建单个会员端堂食订单响应项
func (s *orderSrv) buildMemberDineInOrderItem(ctx context.Context, saleBill model.SaleBill, h5Order *model.H5Order, isProductionFinished bool) resp.MemberDineInOrder {
	// 构建商品列表
	num, productAmount, productList := s.buildMemberDineInOrderProductList(ctx, saleBill)

	// 计算订单状态
	statusInfo := s.getMemberDineInOrderStatusInfo(&saleBill, h5Order, isProductionFinished, ctx.GetLanguage())

	// 计算应付金额：全部退款显示0，部分退款显示扣减后金额
	refundAmount := saleBill.GetTotalRefundAmount()
	amount := saleBill.Amount - refundAmount
	if amount < 0 {
		amount = 0
	}

	saleOrderUuid := uint64(0)
	if saleOrder := saleBill.GetFirstSaleOrder(); saleOrder != nil {
		saleOrderUuid = saleOrder.Uuid
	}

	return resp.MemberDineInOrder{
		SaleBillUuid:  saleBill.Uuid,
		SaleOrderUuid: saleOrderUuid,
		CompanyName:   ctx.GetCompany().Name,
		SerialNo:      saleBill.SerialNo,
		OrderNo:       saleBill.OrderNo,
		StatusInfo:    statusInfo,
		DiningMethod:  saleBill.DiningMethod,
		Num:           num,
		Amount:        amount,
		ProductAmount: productAmount,
		CreateTime:    saleBill.CreateTime,
		SubmitPayTime: saleBill.SubmitPayTime,
		ProductList:   productList,
	}
}

// GetMemberDineInOrderDetail 获取会员端堂食订单详情
func (s *orderSrv) GetMemberDineInOrderDetail(ctx context.Context, detailReq req.GetMemberDineInOrderDetailReq) (*resp.GetMemberDineInOrderDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	orderRepo := repository.NewOrderRepo(db)
	h5OrderRepo := repository.NewH5OrderRepo(db)
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)
	productionRepo := repository.NewProductionRepo(db)

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

	// 获取 H5 订单
	h5Order, _ := h5OrderRepo.GetH5Order(
		h5OrderRepo.WhereSaleBillUuid(saleBill.Uuid),
		h5OrderRepo.WhereOrderType(constant.H5OrderTypeMemberDineIn),
	)

	// 获取生产单完成状态
	// 注意：当没有生产单时（如先下单后付的订单尚未送厨），IsProductionFinishedBySaleBillUuid 返回 true（count==0）
	// 但此时订单并非真正完成，需要检查是否存在生产单来区分"从未送厨"和"全部完成"
	isProductionFinished, _ := productionRepo.IsProductionFinishedBySaleBillUuid(saleBill.Uuid)
	if isProductionFinished {
		hasProduction, _ := productionRepo.HasProductionOrderBySaleBillUuid(saleBill.Uuid)
		if !hasProduction {
			isProductionFinished = false
		}
	}

	// 获取商品列表（包含退款信息）
	baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
	productList := make([]resp.MemberDineInOrderProduct, 0)
	for _, product := range saleOrder.SaleOrderProducts {
		if product.IsPackageSubProduct() {
			continue // 跳过套餐子商品（套餐子商品挂在套餐主商品下）
		}

		// 计算商品退款金额
		var productRefundAmount float64
		for _, returnOrderProduct := range product.ReturnOrderProducts {
			productRefundAmount += returnOrderProduct.ProductTotalAmount
		}

		memberProduct := resp.MemberDineInOrderProduct{
			LocaleName:          product.GetLocaleName(),
			LocaleAttributeName: product.GetAttributeName(),
			Num:                 product.Num,
			Price:               product.SalePrice,      // 折前单价
			TotalPrice:          product.GetSalePrice(), // 折前总价 = 单价 * 数量
			Image: func() string {
				if product.ImageFile == nil {
					return ""
				}
				return product.ImageFile.GetUrl(baseUrl)
			}(),
			RefundAmount:       productRefundAmount,
			ProductType:        uint(product.ProductType),
			PackageProductList: resp.PackageProductList{List: make([]resp.PackageProduct, 0)},
		}

		// 套餐商品：挂载子商品列表
		if product.IsPackageProduct() {
			subProducts := saleOrder.GetPackageSubProductList(product.Uuid)
			for _, sub := range subProducts {
				if sub.DeleteTime != 0 {
					continue
				}
				memberProduct.PackageProductList.List = append(memberProduct.PackageProductList.List, resp.PackageProduct{
					Uuid:                sub.Uuid,
					LocaleName:          sub.GetLocaleName(),
					LocaleAttributeName: sub.GetAttributeName(),
					Num:                 sub.Num,
					UnitNum:             sub.GetProductNum(),
					AddPrice:            sub.AddPrice,
				})
			}
		}

		productList = append(productList, memberProduct)
	}

	// 计算订单总退款金额
	refundAmount := saleBill.GetTotalRefundAmount()

	// 计算应付金额：全部退款显示0，部分退款显示扣减后金额
	amount := saleBill.Amount - refundAmount
	if amount < 0 {
		amount = 0
	}

	// 获取支付信息（通过 SaleOrder 关联的 PaymentOrders 获取）
	var payTime int64
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
		if paymentOrder.PaymentMethod != nil {
			paymentMethodName = paymentOrder.PaymentMethod.GetName()
		}
	}

	// 判断是否是"先下单后付"模式的订单（Status=Pending 但有 H5 订单）
	isOrderFirstPayLater := saleBill.Status == constant.SaleBillStatusPending && h5Order != nil

	// 计算剩余支付时间（"先下单后付"模式的订单不显示支付倒计时）
	var remainingPaymentTime int64
	if saleBill.Status == constant.SaleBillStatusPending && !saleBill.IsExistPaid() && saleBill.SubmitPayTime > 0 && !isOrderFirstPayLater {
		// 假设支付超时时间为15分钟
		paymentTimeout := int64(15 * 60)
		remaining := paymentTimeout - (time.Now().Unix() - saleBill.SubmitPayTime)
		if remaining > 0 {
			remainingPaymentTime = remaining
		}
	}

	// 获取商家信息
	company := ctx.GetCompany()

	// 获取取消时间（从 SaleOrder 获取 DeleteTime 作为取消时间）
	var cancelTime int64
	if saleBill.Status == constant.SaleBillStatusCanceled {
		cancelTime = saleOrder.DeleteTime
	}

	// 待支付状态下获取支付方式列表（用于发起支付）
	// "先下单后付"模式的订单不显示支付方式列表（不需要立即支付）
	paymentMethodList := resp.PaymentMethodList{List: make([]resp.PaymentMethodItem, 0)}
	if saleBill.Status == constant.SaleBillStatusPending && !saleBill.IsExistPaid() && !isOrderFirstPayLater {
		paymentMethods, _ := repository.NewPaymentMethodRepo(db).GetLianLianPayPaymentMethodList()
		for _, paymentMethod := range paymentMethods {
			paymentMethodList.List = append(paymentMethodList.List, resp.PaymentMethodItem{
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
	}

	return &resp.GetMemberDineInOrderDetailResp{
		SaleBillUuid:         saleBill.Uuid,
		SaleOrderUuid:        saleOrder.Uuid,
		CompanyName:          company.Name,
		SerialNo:             saleBill.SerialNo,
		OrderNo:              saleBill.OrderNo,
		StatusInfo:           s.getMemberDineInOrderStatusInfo(saleBill, h5Order, isProductionFinished, ctx.GetLanguage()),
		DiningMethod:         saleBill.DiningMethod,
		Remark:               saleBill.Remark,
		CreateTime:           saleBill.CreateTime,
		SubmitPayTime:        saleBill.SubmitPayTime,
		PayTime:              payTime,
		CancelTime:           cancelTime,
		RemainingPaymentTime: remainingPaymentTime,
		RefundAmount:         refundAmount,
		IsOrderFirstPayLater: saleBill.IsOrderFirstPayLater == 1,
		AmountInfo: resp.MemberDineInOrderAmountInfo{
			DiscountAmount:    saleBill.CustomDiscountFee,
			ServiceFee:        saleBill.ServiceFee,
			TaxFee:            saleBill.TaxFee,
			Amount:            amount,
			PaymentMethodName: paymentMethodName,
		},
		ProductList: resp.MemberDineInOrderProductList{
			List: productList,
		},
		PaymentMethods: paymentMethodList,
	}, nil
}

// CancelMemberDineInOrder 取消会员端堂食订单（未付款订单或先下单后付的待接单订单可取消）
func (s *orderSrv) CancelMemberDineInOrder(ctx context.Context, cancelReq req.CancelMemberDineInOrderReq) error {
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(cancelReq.SaleBillUuid)
		defer lock.NewSystemLock().UnlockUuid(cancelReq.SaleBillUuid)
		ctx.AddLock()
	}

	// 获取DB
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取销售账单完整信息
	orderRepo := repository.NewOrderRepo(db)
	billInfo, err := orderRepo.GetSaleBillAllInfo(cancelReq.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 验证是否是会员端堂食订单
	if billInfo.Source != constant.SaleBillSourceMember || billInfo.BillType != constant.SaleBillTypeInstant {
		return errors.New("订单类型错误")
	}

	// 验证订单归属
	saleOrder := billInfo.GetFirstSaleOrder()
	if saleOrder == nil || saleOrder.ConsumerUuid != ctx.GetMemberUuid() {
		return errors.New("无权操作此订单")
	}

	// 验证订单是未付款状态
	if billInfo.Status != constant.SaleBillStatusPending {
		return errors.New("订单状态不可取消")
	}
	if billInfo.IsExistPaid() {
		return errors.New("订单已支付，不可取消")
	}

	// 检查是否是"先下单后付"模式的订单（有 H5 订单）
	h5OrderRepo := repository.NewH5OrderRepo(db)
	h5Order := s.getH5OrderForMemberDineIn(h5OrderRepo, billInfo.Uuid)

	if h5Order != nil {
		// "先下单后付"模式：只有待接单状态(H5OrderStatusOrder)的订单可以取消
		if h5Order.Status != constant.H5OrderStatusOrder {
			return errors.New("订单已被接单，不可取消")
		}

		// 取消订单并同时软删除关联的 H5 订单
		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			// 取消 SaleBill 和 SaleOrder
			if err := repository.NewOrderRepo(tx).CancelOrderWithoutTx(cancelReq.SaleBillUuid); err != nil {
				return err
			}
			// 软删除 H5 订单
			if err := repository.NewH5OrderRepo(tx).DeleteH5Order(h5Order.Uuid); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return errors.WithMessage(err)
		}
	} else {
		// 普通模式：直接取消订单
		// 取消订单（SaleBill.Status 0→2, SaleOrder.Status 0→2）
		// 未付款的堂食订单没有扣库存也没有送厨，无需退库存和删除送厨单
		err = repository.NewOrderRepo(db).CancelOrderWithoutTx(cancelReq.SaleBillUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
	}

	// 异步: 发布整单取消事件
	utils.Go(func() {
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: billInfo.Uuid,
				OperatorUuid: 0,
				MemberUuid:   ctx.GetMemberUuid(),
			},
		})
	})

	ctx.Log().Info("会员端堂食订单取消成功",
		zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
		zap.Uint64("sale_bill_uuid", cancelReq.SaleBillUuid),
		zap.Uint64("member_uuid", ctx.GetMemberUuid()),
	)

	return nil
}

// getMemberDineInOrderStatusInfo 获取堂食订单状态信息
// h5Order: H5订单，可为nil
// isProductionFinished: 生产单是否全部完成
func (s *orderSrv) getMemberDineInOrderStatusInfo(saleBill *model.SaleBill, h5Order *model.H5Order, isProductionFinished bool, language string) resp.MemberDineInOrderStatusInfo {
	var status string
	var statusText string

	switch saleBill.Status {
	case constant.SaleBillStatusCanceled:
		// 已取消：需要区分是主动取消还是被拒单
		if h5Order != nil && h5Order.Status == constant.H5OrderStatusRejected {
			status = constant.MemberDineInDetailStatusRejected
			statusText = i18n.Translate(language, "已拒单")
		} else {
			status = constant.MemberDineInDetailStatusCancelled
			statusText = i18n.Translate(language, "已取消")
		}
	case constant.SaleBillStatusComplete:
		// 已完成：需要区分 待接单、备餐中、已完成、部分退款、全部退款
		refundAmount := saleBill.GetTotalRefundAmount()
		if refundAmount > 0 {
			// 有退款
			if refundAmount >= saleBill.Amount {
				status = constant.MemberDineInDetailStatusFullRefund
				statusText = i18n.Translate(language, "全部退款")
			} else {
				status = constant.MemberDineInDetailStatusPartialRefund
				statusText = i18n.Translate(language, "部分退款")
			}
		} else if h5Order != nil {
			// 有 H5 订单，根据 H5 订单状态和生产单状态判断
			switch h5Order.Status {
			case constant.H5OrderStatusOrder:
				status = constant.MemberDineInDetailStatusPending
				statusText = i18n.Translate(language, "待接单")
			case constant.H5OrderStatusAccepted:
				// 已接单：根据生产单状态判断是备餐中还是已完成
				if isProductionFinished {
					status = constant.MemberDineInDetailStatusCompleted
					statusText = i18n.Translate(language, "已完成")
				} else {
					status = constant.MemberDineInDetailStatusPreparing
					statusText = i18n.Translate(language, "备餐中")
				}
			case constant.H5OrderStatusRejected:
				status = constant.MemberDineInDetailStatusRejected
				statusText = i18n.Translate(language, "已拒单")
			default:
				status = constant.MemberDineInDetailStatusCompleted
				statusText = i18n.Translate(language, "已完成")
			}
		}
	case constant.SaleBillStatusPending:
		// 待处理状态：通过 IsOrderFirstPayLater 标记区分两种模式
		// 不能仅依赖 h5Order != nil，因为普通模式支付过程中也可能创建 H5 订单
		if saleBill.IsOrderFirstPayLater == constant.OrderFirstPayLaterYes && h5Order != nil {
			// "先下单后付"模式：submit 时已标记 IsOrderFirstPayLater=1，根据 H5 订单状态判断
			switch h5Order.Status {
			case constant.H5OrderStatusOrder:
				status = constant.MemberDineInDetailStatusPending
				statusText = i18n.Translate(language, "待接单")
			case constant.H5OrderStatusAccepted:
				// 已接单：根据生产单状态判断是备餐中还是已完成
				if isProductionFinished {
					status = constant.MemberDineInDetailStatusCompleted
					statusText = i18n.Translate(language, "已完成")
				} else {
					status = constant.MemberDineInDetailStatusPreparing
					statusText = i18n.Translate(language, "备餐中")
				}
			case constant.H5OrderStatusRejected:
				status = constant.MemberDineInDetailStatusRejected
				statusText = i18n.Translate(language, "已拒单")
			default:
				status = constant.MemberDineInDetailStatusPending
				statusText = i18n.Translate(language, "待接单")
			}
		} else {
			// 普通模式：待支付
			status = constant.MemberDineInDetailStatusUnpaid
			statusText = i18n.Translate(language, "待支付")
		}
	}

	return resp.MemberDineInOrderStatusInfo{
		Status:     status,
		StatusText: statusText,
	}
}

// MockPayDineInOrderCallback 模拟堂食订单支付完成回调（仅用于测试）
// 该方法用于在支付服务尚未就绪时，手动触发支付完成的后续流程
func (s *orderSrv) MockPayDineInOrderCallback(ctx context.Context, request req.MockPayDineInOrderCallbackReq) error {
	// 仅在调试模式下允许使用
	if config.Server.Mode != "debug" && config.Server.Mode != "test" {
		return errors.New("该接口仅在调试/测试模式下可用")
	}

	db := ctx.GetDB()

	// 获取销售账单信息
	saleBill, err := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if err != nil {
		return errors.WithMessage(err, "获取订单信息失败")
	}

	// 验证订单来源
	if saleBill.Source != constant.SaleBillSourceMember {
		return errors.New("该订单不是会员端堂食订单")
	}

	// 验证订单状态（必须是待支付状态）
	if saleBill.Status != constant.SaleBillStatusPending {
		return errors.New("订单状态不是待支付，无法模拟支付回调")
	}

	// 获取第一个 SaleOrder
	saleOrder := saleBill.GetFirstSaleOrder()
	if saleOrder == nil {
		return errors.New("未找到销售订单")
	}

	// 模拟更新支付相关字段
	paymentAmount := saleBill.Amount

	// 查找默认支付方式：优先使用 Cash（code=40），如果没有则使用第一个有 erpnext_payment 的支付方式
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable))
	var mockPaymentMethod *model.PaymentMethod
	for _, pm := range paymentMethods {
		if pm.Code == 40 { // Cash
			mockPaymentMethod = pm
			break
		}
	}
	if mockPaymentMethod == nil {
		for _, pm := range paymentMethods {
			if pm.ErpnextPayment != "" {
				mockPaymentMethod = pm
				break
			}
		}
	}
	if mockPaymentMethod == nil {
		return errors.New("未找到可用的支付方式")
	}

	// 更新数据库：创建 PaymentOrder + 更新 SaleBill
	var paymentOrderUuid uint64
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新账单支付金额
		if err := tx.Model(&model.SaleBill{}).Where("uuid = ?", saleBill.Uuid).Updates(map[string]any{
			"payment_amount": paymentAmount,
		}).Error; err != nil {
			return err
		}
		// 创建 PaymentOrder（模拟真实支付回调的行为）
		paymentOrder, err := repository.NewPaymentOrderRepo(tx).Create(model.PaymentOrder{
			PaymentMethodName: mockPaymentMethod.PaymentName,
			PaymentMethodUuid: mockPaymentMethod.Uuid,
			PaymentFeePercent: 0,
			RelatedType:       constant.PaymentOrderRelatedTypeSaleOrder,
			RelatedUuid:       saleOrder.Uuid,
			CurrencyUnit:      "THB",
			PaymentAmount:     paymentAmount,
			Amount:            paymentAmount,
			TransactionNumber: fmt.Sprintf("MOCK-%d", saleOrder.Uuid),
			Status:            constant.PaymentOrderStatusPaid,
		})
		if err != nil {
			return err
		}
		paymentOrderUuid = paymentOrder.Uuid
		return nil
	}); err != nil {
		return errors.WithMessage(err, "更新订单支付信息失败")
	}

	// 发布支付完成事件，触发后续流程（创建 h5_order、自动接单、标记完成等）
	event.NewSystemBus().PublishPayFinishMemberDineInOrderEvent(event.PayFinishMemberDineInOrderPayload{
		BasePayload: event.BasePayload{
			Ctx:           ctx,
			CompanyUuid:   ctx.GetCompanyUuid(),
			Source:        constant.SourceMember,
			SaleBillUuid:  saleBill.Uuid,
			SaleOrderUuid: saleOrder.Uuid,
			MemberUuid:    ctx.GetMemberUuid(),
		},
		PaymentOrderUuid: paymentOrderUuid,
	})

	logger.Logger.Info("MockPayDineInOrderCallback, 模拟支付回调成功",
		zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
		zap.Uint64("sale_bill_uuid", saleBill.Uuid),
		zap.Uint64("sale_order_uuid", saleOrder.Uuid))

	return nil
}

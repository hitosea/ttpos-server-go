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
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/sms"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/hdt3213/delayqueue"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IMemberOrderSrv interface {
	// 会员端
	CreateMemberOrder(ctx context.Context, req req.CreateMemberOrderReq) (*resp.CreateMemberOrderResp, *resp.OrderCheckServiceRes, error)    // 创建会员端订单
	GetMemberOrdeFormInfo(ctx context.Context, req req.GetMemberOrdeFormInfoReq) (*resp.CreateMemberOrderResp, error)                        // 获取订单提交表单信息
	SetMemberOrderAddress(ctx context.Context, req member_req.SetMemberOrderAddressReq) (*resp.CreateMemberOrderResp, error)                 // 设置会员端订单地址
	PayMemberOrder(ctx context.Context, request member_req.PayMemberOrderReq) error                                                          // 会员端订单提交支付，状态变为待支付
	GetMemberOrderPayInfo(ctx context.Context, request member_req.GetMemberOrderPayInfoReq) (*resp.MemberOrderPaymentInfoResp, error)        // 会员端订单获取支付信息
	GetMemberOrderPayStatus(ctx context.Context, request member_req.GetMemberOrderPayStatusReq) (*resp.MemberOrderPaymentStatusResp, error)  // 会员端订单获取支付状态
	GetMemberOrderList(ctx context.Context, req req.MemberOrderListReq) (*resp.GetMemberOrderListResp, error)                                // 查询收银机"外送"页面的订单列表
	GetMemberOrderDetail(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderDetailResp, error)                       // 查询会员端订单详情
	GetMemberOrderPaymentMethodList(ctx context.Context, req req.GetMemberOrderDetailReq) (*resp.GetMemberOrderPaymentMethodListResp, error) // 获取会员端订单支付方式列表
	MemberOrderCancel(ctx context.Context, req member_req.CancelOrderReq) error                                                              // 会员端订单取消
	GetRiderInfo(ctx context.Context, getRiderInfoReq member_req.GetRiderInfoReq) (*resp.MemberOrderCoordinates, error)                      // 获取骑手信息

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
	AcceptMemberSaleOrder(ctx context.Context, req req.AcceptOrderReq) error                                                             // 接单外送订单
	RejectMemberSaleOrder(ctx context.Context, req req.RejectOrderReq) error                                                             // 拒单外送订单
	CookFinishMemberSaleOrder(ctx context.Context, request req.CookFinishOrderReq) error                                                 // 备餐完成外送订单
	GetMemberCashierOrderSearch(ctx context.Context, req req.MemberOrderSearchReq) (*resp.GetMemberCashierOrderSearchResp, error)        // 搜索订单列表通过关键词
	//
	PaidMemberOrder(ctx context.Context, request member_req.PaidMemberOrderReq) error // 会员端订单支付成功. TODO 用于测试，提测前删掉
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
		go func() {
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
		}()

		//FIXME N分钟后取消订单
		Queue.TakeoutCancelQueue.SendDelayMsgV2(
			strconv.FormatUint(result.MemberSaleOrderInfo.MemberSaleOrderUuid, 10),
			30*time.Minute,
			delayqueue.WithRetryCount(3),
		)

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

// GetMemberOrdeFormInfo 获取订单提交表单信息
func (s *orderSrv) GetMemberOrdeFormInfo(ctx context.Context, request req.GetMemberOrdeFormInfoReq) (*resp.CreateMemberOrderResp, error) {
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

	memberSaleOrder.MemberAddressUuid = request.MemberAddressUuid
	memberSaleOrder.ContactLocation = memberAddress.Location
	memberSaleOrder.ContactAddress = memberAddress.Address
	memberSaleOrder.ContactAddressDetail = memberAddress.Street
	memberSaleOrder.ContactName = memberAddress.Name
	memberSaleOrder.ContactPhone = memberAddress.Phone
	memberSaleOrder.ContactPhonePrefix = memberAddress.Country
	memberSaleOrder.ContactGender = memberAddress.Gender

	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderAddress(*memberSaleOrder); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 返回会员端订单信息
	info, err := s.GetMemberOrderCheckoutInfo(ctx, req.GetMemberOrderCheckoutInfoReq{
		MemberSaleOrderUuid: request.MemberSaleOrderUuid,
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

	// 如果未设置地址，返回错误
	if memberSaleOrder.MemberAddressUuid == 0 {
		return errors.New("请先选择订单地址")
	}

	memberSaleOrder.Remark = request.Remark
	memberSaleOrder.SetPendingPayment(request.PaymentMethodUuid)

	// TODO 记录会员折扣、商品数量、商品金额、订单总金额

	if err := repository.NewMemberSaleOrderRepo(db).UpdateMemberSaleOrderPendingPayment(memberSaleOrder); err != nil {
		return errors.WithMessage(err)
	}

	// 添加24小时后自动取消订单的延时队列任务
	go func() {
		memberSaleOrderUuidStr := strconv.FormatUint(request.MemberSaleOrderUuid, 10)
		if Queue.MemberOrderCancelQueue != nil {
			// 构建队列消息参数
			paramsJson := utils.ToJson(map[string]interface{}{
				"member_sale_order_uuid": request.MemberSaleOrderUuid,
				"company_uuid":           ctx.GetCompanyUuid(),
				"reason":                 "支付超时",
			})
			// 发送24小时后自动取消订单的延时消息，重试1次
			_, err := Queue.MemberOrderCancelQueue.SendDelayMsgV2(
				paramsJson,
				24*time.Hour,                 // 24小时后执行
				delayqueue.WithRetryCount(2), // 重试2次
			)
			if err != nil {
				ctx.Log().Error("添加24小时自动取消订单任务失败",
					zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr),
					zap.Error(err))
			}
		} else {
			ctx.Log().Error("MemberOrderCancelQueue队列未初始化",
				zap.String("memberSaleOrderUuid", memberSaleOrderUuidStr))
		}
	}()

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
	if !memberSaleOrder.IsCanPaid() {
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
		memberSaleOrder.SetPendingPayment(paymentMethod.Uuid)
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
		paymentOrderRepo.WhereRelatedType(constant.PaymentOrderRelatedTypeMemberOrder),
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
			RelatedType:       constant.PaymentOrderRelatedTypeMemberOrder,
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

	// 创建连连支付订单
	payment, err := NewPaymentRepo(ctx, s.dbm).CreatePayment(CreatePaymentReq{
		PaymentOrderUuid:  paymentOrder.Uuid,
		RelatedType:       constant.PaymentOrderRelatedTypeMemberOrder,
		RelatedUuid:       relatedUuid,
		PaymentMethodUuid: paymentMethod.Uuid,
		PaymentMethodCode: paymentMethod.Code,
		PaymentAmount:     memberSaleOrder.Amount,
		CommissionFee:     0,
		PaymentMethod:     PaymentMethodH5Payment,
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &resp.MemberOrderPaymentInfoResp{
		MemberSaleOrderUuid: memberSaleOrder.Uuid,
		PaymentOrderUuid:    payment.PaymentOrderUuid,
		PaymentMethodName:   paymentMethod.PaymentName,
		QrCode:              utils.IfString(paymentMethod.IsQrPromptPay(), payment.LinkUrl, ""),
		LinkUrl:             utils.IfString(paymentMethod.IsQrPromptPay(), "", payment.LinkUrl),
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
		memberSaleOrderRepo.WithSaleBillSaleOrderProduct(),
		memberSaleOrderRepo.WhereNotStatusIn([]uint{constant.MemberSaleOrderStatusSelecting}),
		memberSaleOrderRepo.WhereStatusIn(constant.GetMemberOrderStatusList(req.Status)),
		memberSaleOrderRepo.WhereKeyword(req.Keyword, ctx.GetLanguage()),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	memberOrders := make([]resp.MemberOrder, 0)
	for _, memberSaleOrder := range memberSaleOrders {
		memberOrders = append(memberOrders, resp.MemberOrder{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			CompanyName:         ctx.GetCompany().Name,
			SerialNumber:        memberSaleOrder.SerialNumber,
			Status:              memberSaleOrder.Status,
			Num:                 memberSaleOrder.ProductNum,
			ProductAmount:       memberSaleOrder.ProductAmount,
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
						LocaleName: saleOrderProduct.MultiLanguageName.GetNames(),
						Num:        saleOrderProduct.Num,
						TotalPrice: saleOrderProduct.GetTotalPrice(),
						Image:      saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
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
	for _, saleOrderProduct := range memberSaleOrder.SaleBill.SaleOrders[0].SaleOrderProducts {
		products = append(products, resp.MemberOrderProduct{
			LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
			OriginTotalPrice:    saleOrderProduct.GetTotalProductPrice(),
			Image:               saleOrderProduct.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request)),
		})
	}
	//
	var address resp.MemberOrderDetailAddress
	if memberSaleOrder.Address != nil {
		address = resp.MemberOrderDetailAddress{
			ContactName: memberSaleOrder.ContactName,
			Phone:       memberSaleOrder.ContactPhone,
			PhonePrefix: memberSaleOrder.ContactPhonePrefix,
			Address:     memberSaleOrder.ContactAddress + memberSaleOrder.ContactAddressDetail,
		}
	}
	//
	return &resp.GetMemberOrderDetailResp{
		MemberSaleOrderUuid:  memberSaleOrder.Uuid,
		CompanyName:          ctx.GetCompany().Name,
		PayTime:              memberSaleOrder.PayTime,
		FinishTime:           memberSaleOrder.FinishTime,
		CancelTime:           memberSaleOrder.CancelTime,
		RemainingPaymentTime: memberSaleOrder.GetRemainingPaymentTime(),
		CancelReason:         memberSaleOrder.CancelReason,
		CreateTime:           memberSaleOrder.CreateTime,
		Status:               memberSaleOrder.Status,
		Remark:               memberSaleOrder.Remark,
		AmountInfo: resp.MemberOrderAmountInfo{
			Amount:            memberSaleOrder.Amount,
			MemberDiscountFee: memberSaleOrder.MemberDiscountFee,
		},
		ProductList: resp.MemberProductList{
			List:          products,
			ProductAmount: memberSaleOrder.ProductAmount,
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
			Name:  memberSaleOrder.RiderName,
			Phone: memberSaleOrder.RiderPhone,
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
	if !memberSaleOrder.IsCanPaid() {
		return nil, errors.New("订单状态不可支付")
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
				var logoUrl string
				if paymentMethod.LogoFile != nil {
					logoUrl = paymentMethod.LogoFile.GetUrl(baseUrl)
				}
				if logoUrl == "" && paymentMethod.DefaultImg != "" {
					logoUrl = strings.TrimRight(baseUrl, "/") + paymentMethod.DefaultImg
				}
				return logoUrl
			}(),
			Qrcode: func() string {
				if paymentMethod.QrcodeFile != nil {
					return paymentMethod.QrcodeFile.GetUrl(baseUrl)
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

	// 取消订单后，删除所有销售订单商品
	err = repository.NewSaleOrderProductRepo(tx).DeleteSaleOrderProductBySaleBillUuid(billInfo.Uuid)
	if err != nil {
		tx.Rollback()
		return errors.WithMessage(builtinerrors.New("删除销售订单商品失败"), err.Error())
	}

	// 已经支付的-发起退款
	if memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingMerchantAccept {
		err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*saleOrder, MemberSaleOrderRefundReq{
			CancelReason: "客户取消订单",
		})
		if err != nil {
			tx.Rollback()
			return errors.WithMessage(err)
		}
	}

	// 设置订单为“已取消”状态
	memberSaleOrder.RefundAmount = memberSaleOrder.Amount
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

	// 发布“整单取消”操作事件
	go func() {
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{ // 整单取消
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: billInfo.Uuid,
				OperatorUuid: int64(ctx.GetMemberUuid()),
			},
		})
	}()

	// 发送短信通知
	go NewSMSSrv(s.dbm).SendDeliveryOrderBySelfCancelSMS(ctx, memberSaleOrder.ContactPhone, &sms.DeliveryOrderCancelBySelfRequest{
		Company: ctx.GetCompany().Name,
		OrderNo: memberSaleOrder.OrderNo,
	})

	// 成功后，推送到厨显端更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
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

	err = repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 退回商品库存
		if err := s.returnInventory(ctx.Copy(), billInfo); err != nil {
			return errors.WithMessage(err)
		}

		// 取消订单
		err = orderRepo.CancelOrder(ctx.Copy(), saleBillUuid, 0, request.CancelReason)
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

		// 取消订单后，删除所有销售订单商品
		err = repository.NewSaleOrderProductRepo(tx).DeleteSaleOrderProductBySaleBillUuid(billInfo.Uuid)
		if err != nil {
			return errors.WithMessage(builtinerrors.New("删除销售订单商品失败"), err.Error())
		}

		// 已经支付的-发起退款
		if memberSaleOrder.Status == constant.MemberSaleOrderStatusPendingMerchantAccept {
			err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*saleOrder, MemberSaleOrderRefundReq{
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
	go func() {
		s.bus.PublishCancelOrderEvent(event.CancelOrderPayload{
			BasePayload: event.BasePayload{ // 整单取消
				Ctx:          ctx,
				CompanyUuid:  ctx.GetCompanyUuid(),
				Source:       ctx.GetSource(),
				SaleBillUuid: billInfo.Uuid,
				OperatorUuid: int64(ctx.GetMemberUuid()),
			},
		})
	}()

	// 成功后，推送到厨显端更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
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
	riderLat, riderLng := memberSaleOrder.GetLocation()

	return &resp.MemberOrderCoordinates{
		Merchant: resp.OrderCoordinate{
			Name:    company.Name,
			Address: companySetting.Address,
			Lat:     merchantLat,
			Lng:     merchantLng,
		},
		Customer: resp.OrderCoordinate{
			Name:    memberSaleOrder.ContactName,
			Address: memberSaleOrder.ContactAddress + "(" + memberSaleOrder.ContactAddressDetail + ")",
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

// GetMemberCashierOrderList 获取收银端"外送"订单列表
func (s *orderSrv) GetMemberCashierOrderList(ctx context.Context, req req.MemberOrderListReq) (*resp.GetMemberCashierOrderListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrders, total, err := repository.NewMemberSaleOrderRepo(db).GetCashierMemberSaleOrderList(
		req.PageNo,
		req.PageSize,
		constant.GetStatusList(req.Status),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	memberOrders := make([]resp.MemberCashierOrder, 0)
	for _, memberSaleOrder := range memberSaleOrders {
		memberOrders = append(memberOrders, resp.MemberCashierOrder{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			SerialNumber:        memberSaleOrder.SerialNumber,
			Status:              memberSaleOrder.Status,
			StatusGroup:         constant.ParseToStatusGroup(memberSaleOrder.Status),
			Num:                 memberSaleOrder.ProductNum,
			ProductAmount:       memberSaleOrder.ProductAmount,
		})
	}

	// 获取数量
	getOrderNum := func(status []uint) int64 {
		num, _ := repository.NewMemberSaleOrderRepo(db).GetOrderNum(status)
		return num
	}

	return &resp.GetMemberCashierOrderListResp{
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
		Extra: resp.ExtraMemberCashierOrderListMeta{
			UnacceptNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUnaccept)),
			AcceptNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusAccept)),
			UndeliveryNum: getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusUndelivery)),
			DeliveryNum:   getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivery)),
			CompletedNum:  getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusDelivered)),
			CancelNum:     getOrderNum(constant.GetStatusList(constant.CashierMemberSaleOrderStatusCancel)),
		},
		List: memberOrders,
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
			LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
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
			ProductAmount: memberSaleOrder.ProductAmount,
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
				IsCellRefund: func() bool {
					// 订单完成后，用户可以申请退款。TODO 且未全部退款时
					return memberSaleOrder.Status == constant.MemberSaleOrderStatusCompleted
				}(),
			},
		})
	}

	getOrderNum := func(memberOrders []resp.MemberOrderManage, statusGroup string) int64 {
		num := int64(0)
		for _, memberOrder := range memberOrders {
			if memberOrder.StatusGroup == statusGroup {
				num++
			}
		}
		return num
	}

	unpaidNum := getOrderNum(memberOrders, "unpaid")
	unacceptNum := getOrderNum(memberOrders, "unaccept")
	acceptNum := getOrderNum(memberOrders, "accept")
	undeliveryNum := getOrderNum(memberOrders, "undelivery")
	deliveryNum := getOrderNum(memberOrders, "delivery")
	completeNum := getOrderNum(memberOrders, "completed")
	cancelNum := getOrderNum(memberOrders, "cancel")

	allNum := unpaidNum + unacceptNum + undeliveryNum + deliveryNum + completeNum + cancelNum

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
			LocaleName:          saleOrderProduct.MultiLanguageName.GetNames(),
			LocaleAttributeName: saleOrderProduct.GetAttributeName(),
			ImageUrl:            imageUrl,
			OriginUnitPrice:     saleOrderProduct.OriginTotalPrice,
			UnitPrice:           saleOrderProduct.TotalPrice,
			Num:                 saleOrderProduct.Num,
			TotalPrice:          saleOrderProduct.GetTotalPrice(),
			OriginTotalPrice:    saleOrderProduct.GetTotalProductPrice(),
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
		member = resp.OrderMember{
			ID:   memberSaleOrder.SaleBill.SaleOrders[0].Member.ID,
			Name: memberSaleOrder.SaleBill.SaleOrders[0].Member.Nickname,
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
		PayType:             payType,
		PayTime:             memberSaleOrder.PayTime,
		CreateTime:          memberSaleOrder.CreateTime,
		FinshTime:           memberSaleOrder.FinishTime,
		CancelReason:        memberSaleOrder.CancelReason,
		CancelTime:          memberSaleOrder.CancelTime,
		Remark:              memberSaleOrder.Remark,
		Member:              member,
		Cachier: resp.CachierInfo{
			Uuid: memberSaleOrder.SaleBill.CashierUuid,
			Name: memberSaleOrder.SaleBill.CashierName,
		},
		ProductList:  resp.MemberProductManageList{List: products},
		OperationLog: resp.OperationLog{List: operationLogs},
	}, nil
}

// AcceptMemberSaleOrder 接单外送订单
func (s *orderSrv) AcceptMemberSaleOrder(ctx context.Context, request req.AcceptOrderReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	memberSaleOrder, err := getMemberOrderDetail(ctx, request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 接单
	memberSaleOrder.Accept()

	// 获取未送厨的商品列表
	unCookingSaleOrderProducts := memberSaleOrder.SaleBill.GetSaleOrderProductUnCooking()
	if len(unCookingSaleOrderProducts) > 0 {
		// 整单送厨
		_, checkRes, err := s.InstantOrderCartProductCooking(ctx, req.OrderCartProductCookingReq{
			SaleBillUuid:        memberSaleOrder.SaleBill.Uuid,
			IgnoreMust:          true,
			IsMemberOrderAccept: true,
		})
		if err != nil {
			return errors.WithMessage(err, "整单送厨失败")
		}
		if checkRes != nil {
			return errors.New("整单送厨失败")
		}
	}
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderAccept(*memberSaleOrder)
		lat, lng, _ := memberSaleOrder.Address.GetLocation()
		// 获取商家地址
		companySetting := ctx.GetCompanySetting()
		latitude, longitude := companySetting.GetCoordinates()
		if latitude == "" || longitude == "" {
			return errors.New("无法找到商家经纬度")
		}
		// 选择外送渠道
		memberSaleOrder.RelatedOrderType = constant.ProviderNameSkootar
		// 状态变更的回调地址
		callbackUrl := config.TakeOutRpcConf.CallbackEndpoint + "/api/v1/member/order/callback?company_uuid=" + fmt.Sprintf("%d", ctx.GetCompany().Uuid)

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
			ctx.Log().Error("创建外送订单失败", zap.Error(err))
			ctx.Log().Info("创建外送订单失败", zap.String("takeout_ref_no", res.TakeoutRefNo), zap.Duration("cost", time.Since(startTime)))
			return errors.WithMessage(errors.NewWithCode(constant.CodeTakeoutCreateOrderError, "创建外送订单失败"), err.Error())
		}
		ctx.Log().Info("创建外送订单成功", zap.String("takeout_ref_no", res.TakeoutRefNo), zap.Duration("cost", time.Since(startTime)))
		memberSaleOrder.RelatedOrderNo = res.TakeoutRefNo
		memberSaleOrder.ExpectedFinishTime = res.FinishTime
		if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderProviderInfo(*memberSaleOrder); err != nil {
			return errors.WithMessage(err, "更新外送订单失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err, "更新外送订单失败")
	}

	// 发布"外送接单"操作事件
	go func() {
		s.bus.PublishAcceptMemberSaleOrderEvent(event.AcceptMemberSaleOrderPayload{
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
	}()

	return nil
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
	s.CancelOrder(ctx, req.OrderCancelReq{
		SaleBillUuid: memberSaleOrder.SaleBill.Uuid,
	})

	// 退款
	err = NewPaymentRepo(ctx, s.dbm).MemberSaleOrderRefund(*billInfo.GetFirstSaleOrder(), MemberSaleOrderRefundReq{
		CancelReason: "商家拒单",
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	// 更新订单状态
	repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderReject(*memberSaleOrder)
		return nil
	})

	// 发布"外送拒单"操作事件
	go func() {
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
		})
	}()

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
	go func() {
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
	}()

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
		memberOrders = append(memberOrders, resp.MemberCashierOrder{
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			SerialNumber:        memberSaleOrder.SerialNumber,
			Status:              memberSaleOrder.Status,
			Num:                 memberSaleOrder.ProductNum,
			ProductAmount:       memberSaleOrder.ProductAmount,
		})
	}

	return &resp.GetMemberCashierOrderSearchResp{
		List: memberOrders,
	}, nil
}

// PaidMemberOrder 会员端订单支付成功
func (s *orderSrv) PaidMemberOrder(ctx context.Context, request member_req.PaidMemberOrderReq) error {
	ctx.SetDB(s.dbm.GetDB(ctx.GetDbId()))
	memberSaleOrder, err := getMemberOrderDetail(ctx, request.MemberSaleOrderUuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 更新订单状态
	s.bus.PublishPayFinishMemberSaleOrderEvent(event.PayFinishMemberSaleOrderPayload{
		BasePayload: event.BasePayload{
			Ctx:                 ctx,
			CompanyUuid:         ctx.GetCompanyUuid(),
			Source:              ctx.GetSource(),
			SaleBillUuid:        memberSaleOrder.SaleBill.Uuid,
			SaleOrderUuid:       memberSaleOrder.SaleBill.SaleOrders[0].Uuid,
			MemberSaleOrderUuid: memberSaleOrder.Uuid,
			MemberUuid:          ctx.GetMemberUuid(),
		},
		MemberSaleOrderUuid: request.MemberSaleOrderUuid,
	})
	return nil
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

	// 构建退款支付记录
	memberPaymentRecords := make([]resp.OrderReturnPaymentRecord, 0)
	for _, record := range paymentRecords {
		memberPaymentRecords = append(memberPaymentRecords, resp.OrderReturnPaymentRecord{
			PaymentMethodCode: record.PaymentMethodCode,
			PaymentOrderUuid:  record.PaymentOrderUuid,
			PaymentMethodName: record.PaymentMethodName,
			PaymentMethodUuid: record.PaymentMethodUuid,
			CurrencyUnit:      record.CurrencyUnit,
			PaymentAmount:     record.PaymentAmount,
			CanReturnAmount:   record.CanReturnAmount,
		})
	}

	// 获取商品列表
	for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
		if saleOrderProduct.IsCancelProduct() || saleOrderProduct.IsGiftProduct() || saleOrderProduct.Status == constant.OrderProductStatusUnSending {
			continue
		}
		products = append(products, resp.OrderReturnProduct{
			SaleOrderProductUuid: saleOrderProduct.Uuid,
			LocaleName:           saleOrderProduct.MultiLanguageName.GetNames(),
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
	canReturnAmount := saleOrder.GetCanReturnAmount()
	res := &resp.OrderReturnInfoResp{
		ManualReturnPoints: saleOrder.CanManualReturnPoints(), // 是否可以手动退款积分。订单是按比例赠送积分且未发生积分抵扣时，不自动退款。
		DeductiblePoints:   saleOrder.GetManualReturnPoints(), // 可扣除积分。订单赠送的积分-已经退回的积分
		CanReturnAmount:    canReturnAmount,                   // 可退款金额. 可退款金额=订单最终应收金额-已退款金额
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

package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/saas"
	"ttpos-server-go/config"
	contexts "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/encrypt"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/shopspring/decimal"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/spf13/viper"
)

// RequestTimeOut 请求超时
const RequestTimeOut = 10

const (
	PaymentMethodH5Payment = "H5_PAYMENT"   // 微信 - H5支付
	PaymentMethodWechatPay = "DYNAMIC_CODE" // 微信 - 动态码支付
)

type CreatePaymentReq struct {
	RelatedType         int
	RelatedUuid         uint64
	PaymentMethodUuid   uint64
	PaymentMethodCode   int
	PaymentAmount       float64
	CommissionFee       float64
	PaymentMethod       string
	PaymentOrderUuid    uint64
	MemberSaleOrderUuid uint64
	RedirectUrl         string // 跳转地址
}

// LianLianPaymentResp 连连支付仓库
type LianLianPaymentResp struct {
	CallbackUrl     string `json:"callback_url"`      // 回调地址
	CreatedAt       string `json:"created_at"`        // 创建时间
	MerchantOrderNo string `json:"merchant_order_no"` // 商户订单号
	PayAt           string `json:"pay_at"`            // 支付时间
	PayTypeDesc     string `json:"pay_type_desc"`     // 支付类型描述
	PaymentId       string `json:"payment_id"`        // 支付ID
	Order           struct {
		CreateTime    string `json:"create_time"`    // 创建时间
		OrderAmount   string `json:"order_amount"`   // 订单金额
		OrderCurrency string `json:"order_currency"` // 订单币种
		OrderId       string `json:"order_id"`       // 订单号
		OrderStatus   string `json:"order_status"`   // 订单状态  WP-未支付，RP-已经支付
		// 微信/支付宝
		LinkUrl string `json:"link_url"` // 链接地址 - 二维码内容
		// 支付宝
		DueDate         string `json:"due_date"`          // 支付宝 - 到期时间
		OrderCreateTime string `json:"order_create_time"` // 支付宝 - 订单创建时间
		// LIANLIAN_ALI_OFFLINE_PAY
		MerchantId      string `json:"merchant_id"`        // 商户ID
		MerchantOrderId string `json:"merchant_order_id"`  // 商户订单id
		QrCode          string `json:"qr_code"`            // 二维码 - base64
		QrCodeExpireSec string `json:"qr_code_expire_sec"` // 二维码过期秒 480
	} `json:"order"`
}

// PaymentRepo 连连支付仓库
type PaymentRepo struct {
	dbm                    *database.DBManager
	ctx                    contexts.Context
	payServiceIp           string
	payServiceUrl          string
	payServiceRsaPublicKey string
	payServiceAuthUri      string
	payCallbackUrl         string // 支付回调地址
	refundCallbackUrl      string // 退款回调地址
	orderCurrency          string // 订单币种 [THB, USD, JPY, CNY]
}

// NewPaymentRepo 创建连连支付仓库
func NewPaymentRepo(ctx contexts.Context, dbm *database.DBManager) *PaymentRepo {
	return &PaymentRepo{
		ctx:                    ctx,
		dbm:                    dbm,
		payServiceIp:           viper.GetString("PAY_SERVICE_IP"),
		payServiceUrl:          viper.GetString("PAY_SERVICE_URL"),
		payServiceRsaPublicKey: viper.GetString("PAY_SERVICE_RSA_PUBLIC_KEY"),
		payServiceAuthUri:      viper.GetString("PAY_SERVICE_AUTH_URI"),
		payCallbackUrl: func() string {
			if viper.GetString("PAY_SERVICE_LIANLIAN_CALLBACK_URL") == "" {
				if config.Server.Domain != "" {
					return config.Server.Domain + "/api/v1/passport/lianlian/callback"
				} else {
					return ""
				}
			}
			return viper.GetString("PAY_SERVICE_LIANLIAN_CALLBACK_URL")
		}(),
		refundCallbackUrl: func() string {
			if viper.GetString("PAY_SERVICE_LIANLIAN_REFUND_CALLBACK_URL") == "" {
				if config.Server.Domain != "" {
					return config.Server.Domain + "/api/v1/passport/lianlian/refund/callback"
				} else {
					return ""
				}
			}
			return viper.GetString("PAY_SERVICE_LIANLIAN_REFUND_CALLBACK_URL")
		}(),
		orderCurrency: "THB",
	}
}

// CreatePayment 创建支付
func (p *PaymentRepo) CreatePayment(req CreatePaymentReq) (*model.LlPaymentOrder, error) {
	paymentApp, err := p.validateConfig(p.ctx.GetCompanyUuid())
	if err != nil {
		return nil, err
	}

	// 创建支付订单仓库
	llPaymentOrderRepo := repository.NewLlPaymentOrderRepo(p.dbm.GetDB(p.ctx.GetDbId()))
	merchantUserId := fmt.Sprintf("%d", p.ctx.GetCompanyUuid())
	merchantOrderNo := utils.GenerateMerchantOrderNo("PS")

	// 获取支付配置
	url, orderType, err := p.getPaymentInfo(req.PaymentMethodCode)
	if err != nil {
		return nil, err
	}

	// 判断是否已存在有效待支付二维码
	order, err := p.GetValidPaymentOrderByUuid(req.RelatedType, req.RelatedUuid, orderType, merchantUserId, req.PaymentAmount)
	if err != nil {
		return nil, err
	}
	if order != nil {
		return order, nil
	}

	// 组装请求数据
	jsonStr := fmt.Sprintf("{\"shop_supplier_id\":%v,\"merchant_order_no\":\"%v\",\"order_amount\":\"%v\",\"order_currency\":\"%v\",\"order_desc\":\"%v\",\"full_name\":\"%v\",\"merchant_user_id\":%v,\"callback_url\":\"%v\",\"payment_method\":\"%v\",\"redirect_url\":\"%v\"}",
		p.ctx.GetCompanyUuid(),
		merchantOrderNo,
		req.PaymentAmount,
		p.orderCurrency,
		"CHECK_OUT",
		"CASHIER",
		p.ctx.GetCompanyUuid(),
		strings.ReplaceAll(p.payCallbackUrl, "/", "\\/"),
		req.PaymentMethod,
		strings.ReplaceAll(req.RedirectUrl, "/", "\\/"),
	)

	// 请求支付
	response, err := p.postRequest(url, jsonStr, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
		"sign":         p.requestSign(paymentApp.LlSignSalt, jsonStr),
		"client-ip":    p.ctx.GetRemoteIp(),
	}, RequestTimeOut)
	if err != nil {
		return nil, errors.WithMessage(errors.NewWithCode(constant.CodeOrderPayError, "请求支付失败"), err.Error())
	}

	// 返回支付结果
	var resp LianLianPaymentResp
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	// 然后将 JSON 字符串解析到结构体中
	if err := json.Unmarshal(responseJSON, &resp); err != nil {
		return nil, err
	}
	// 当是微信或者支付宝支付的时候，需要把LinkUrl转换成二维码
	if req.PaymentMethodCode == constant.PaymentMethodCodeLianLianWechatPay || req.PaymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
		if req.PaymentMethod == PaymentMethodH5Payment {
			if req.PaymentMethodCode == constant.PaymentMethodCodeLianLianWechatPay {
				resp.Order.QrCode = resp.Order.LinkUrl
			} else {
				resp.Order.QrCode = "alipays://platformapi/startapp?saId=10000007&qrcode=" + resp.Order.LinkUrl
			}
		} else {
			if resp.Order.QrCode == "" {
				qr, err := qrcode.New(resp.Order.LinkUrl, qrcode.Medium)
				if err != nil {
					return nil, err
				}
				// 生成PNG格式的二维码图片
				png, err := qr.PNG(256) // 生成256x256大小的PNG图片
				if err != nil {
					return nil, err
				}
				// 转换为base64
				resp.Order.QrCode = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
			}
		}
		if req.PaymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
			resp.Order.CreateTime = resp.Order.OrderCreateTime
		}
	} else {
		// PromptPay直接使用返回的QR code
		resp.Order.QrCode = "data:image/png;base64," + resp.Order.QrCode
	}

	// 生成雪花ID
	uuid := req.PaymentOrderUuid
	if uuid == 0 {
		uuid, err = utils.GetID()
		if err != nil {
			return nil, errors.New("生成雪花ID失败")
		}
	}

	// 创建支付订单
	paymentOrder := &model.LlPaymentOrder{
		PaymentOrderUuid:    uuid,
		PaymentMethodUuid:   req.PaymentMethodUuid,
		RelatedType:         req.RelatedType,
		RelatedUuid:         req.RelatedUuid,
		MerchantOrderId:     merchantOrderNo,
		MerchantId:          resp.Order.MerchantId,
		OrderId:             resp.Order.OrderId,
		OrderType:           resp.PayTypeDesc,
		OrderStatus:         resp.Order.OrderStatus,
		OrderAmount:         req.PaymentAmount,
		OrderCurrency:       resp.Order.OrderCurrency,
		FullName:            "CASHIER",
		OrderDesc:           resp.PayTypeDesc,
		LinkUrl:             resp.Order.QrCode,
		MerchantUserId:      merchantUserId,
		LlCreateTime:        resp.Order.CreateTime,
		CommissionFee:       req.CommissionFee,
		MemberSaleOrderUuid: req.MemberSaleOrderUuid,
	}
	// 设置创建时间
	paymentOrder.CreateTime = time.Now().Unix()
	// 设置过期时间
	paymentOrder.ExpiredTime = paymentOrder.CreateTime + paymentOrder.GetAliveTime()
	// 保存支付订单
	if _, err := llPaymentOrderRepo.Create(*paymentOrder); err != nil {
		return nil, err
	}
	// 返回支付订单
	return paymentOrder, nil
}

// GetValidPaymentOrderByUuid 获取有效的支付订单
func (p *PaymentRepo) GetValidPaymentOrderByUuid(
	relatedType int,
	relatedUuid uint64,
	orderType string,
	merchantUserId string,
	paymentAmount float64,
) (*model.LlPaymentOrder, error) {
	llPaymentOrderRepo := repository.NewLlPaymentOrderRepo(p.dbm.GetDB(p.ctx.GetDbId()))
	// 判断是否已存在有效待支付二维码
	order, err := llPaymentOrderRepo.GetPaymentOrder(
		repository.CommonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			db = db.Where("related_uuid = ?", relatedUuid)
			db = db.Where("related_type = ?", relatedType)
			db = db.Where("merchant_user_id = ?", merchantUserId)
			db = db.Where("order_type = ?", orderType)
			db = db.Where("order_amount = ?", paymentAmount)
			db = db.Where("order_currency = ?", p.orderCurrency)
			db = db.Where("(expired_time > ? or pay_time != ?)", time.Now().Unix()+5, 0)
			return db.Order("id desc")
		},
	)
	if err != nil {
		return nil, err
	}
	if order.Uuid > 0 {
		return &order, nil
	}
	return nil, nil
}

// PaymentServiceRefundReq 连连支付仓库
type PaymentServiceRefundReq struct {
	PaymentOrderUuid      uint64  // 支付订单UUID
	RelatedType           int     // 支付订单类型
	MerchantRefundOrderNo string  // 商户退款订单号
	RefundAmount          float64 // 退款金额
	RefundOrderId         string  // 退款ID
	BankCode              string  // 银行代码
	AccountNo             string  // 账号
	AccountName           string  // 账号名称
	RefundRequestIndex    int     `json:"refund_request_index"` // 退款请求次数索引
}
type LianLianPaymentRefundResp struct {
	MerchantId       string `json:"merchant_id"`        // 商户ID
	RefundOrderId    string `json:"refund_order_id"`    // 退款订单ID
	MerchantRefundId string `json:"merchant_refund_id"` // 商户退款ID
	RefundAmount     string `json:"refund_amount"`      // 退款金额
	RefundCurrency   string `json:"refund_currency"`    // 退款币种
	RefundStatus     string `json:"refund_status"`      // 退款状态
	CreateTime       string `json:"ll_create_time"`     // 创建时间
}

func (p *PaymentRepo) handleRefundError(err error, serviceRefundReq *PaymentServiceRefundReq) (*LianLianPaymentRefundResp, error) {
	if serviceRefundReq.RefundRequestIndex == 0 {
		serviceRefundReq.RefundRequestIndex = 1
		return p.Refund(*serviceRefundReq)
	}
	return nil, fmt.Errorf("退款失败: %w", err)
}
func (p *PaymentRepo) Refund(serviceRefundReq PaymentServiceRefundReq) (*LianLianPaymentRefundResp, error) {
	paymentApp, err := p.validateConfig(p.ctx.GetCompanyUuid())
	if err != nil {
		return nil, err
	}
	// 获取支付订单信息
	order, err := repository.NewLlPaymentOrderRepo(p.dbm.GetDB(p.ctx.GetDbId())).GetPaymentOrder(
		repository.CommonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			db = db.Where("payment_order_uuid = ?", serviceRefundReq.PaymentOrderUuid)
			db = db.Where("related_type = ?", serviceRefundReq.RelatedType)
			return db.Order("id desc")
		},
	)
	if err != nil {
		return nil, err
	}
	if order.Uuid == 0 {
		return nil, errors.New("支付订单不存在")
	}
	if order.PayTime == 0 {
		return nil, errors.New("支付订单未支付")
	}
	// 组装请求数据
	jsonStr := fmt.Sprintf("{\"shop_supplier_id\":\"%v\",\"merchant_order_no\":\"%v\",\"merchant_refund_id\":\"%v\",\"refund_amount\":\"%v\",\"refund_currency\":\"%v\",\"refund_reason\":\"%v\",\"callback_url\":\"%v\",\"merchant_refund_order_no\":\"%v\",\"bank_code\":\"%v\",\"account_no\":\"%v\",\"account_name\":\"%v\"}",
		p.ctx.GetCompanyUuid(),
		order.MerchantOrderId,
		serviceRefundReq.RefundOrderId,
		serviceRefundReq.RefundAmount,
		p.orderCurrency,
		"其他",
		strings.ReplaceAll(p.refundCallbackUrl+"?merchant_refund_id="+serviceRefundReq.MerchantRefundOrderNo, "/", "\\/"),
		serviceRefundReq.MerchantRefundOrderNo,
		serviceRefundReq.BankCode,
		serviceRefundReq.AccountNo,
		serviceRefundReq.AccountName,
	)
	// 计算签名并请求
	response, err := p.postRequest(p.payServiceUrl+"/api/receipts/lianlianRefund", jsonStr, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
		"sign":         p.requestSign(paymentApp.LlSignSalt, jsonStr),
	}, RequestTimeOut)
	if err != nil {
		return p.handleRefundError(err, &serviceRefundReq)
	}
	// 返回结果
	var resp LianLianPaymentRefundResp
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return p.handleRefundError(err, &serviceRefundReq)
	}
	// 然后将 JSON 字符串解析到结构体中
	if err := json.Unmarshal(responseJSON, &resp); err != nil {
		return p.handleRefundError(err, &serviceRefundReq)
	}
	// 返回支付订单
	return &resp, nil
}

// HandleCallback 处理支付回调
func (p *PaymentRepo) HandleCallback(sign string, callbackReq req.LianLianCallbackRequest) error {
	// 验证签名
	err := p.validateSign(sign, callbackReq)
	if err != nil {
		return err
	}
	// 验证支付状态
	if callbackReq.PayStatus != 1 {
		fmt.Println("支付状态不正确")
		return errors.New("支付状态不正确")
	}
	// 转换商户ID
	companyUuid, err := strconv.ParseUint(callbackReq.CompanyUuid, 10, 64)
	if err != nil {
		fmt.Println("商户ID格式错误", err)
		return fmt.Errorf("商户ID格式错误: %v", err)
	}
	// 设置数据库
	db := p.dbm.GetDB(companyUuid)
	p.ctx.SetDB(db)
	p.ctx.SetCompanyUuid(companyUuid)

	// 获取支付订单
	order, err := repository.NewLlPaymentOrderRepo(db).GetPaymentOrder(
		repository.CommonRepo.WhereBySoftDelete(),
		func(db *gorm.DB) *gorm.DB {
			db = db.Where("merchant_user_id = ?", callbackReq.MerchantUserId)
			db = db.Where("order_amount = ?", callbackReq.OrderAmount)
			db = db.Where("pay_time = ?", "")
			db = db.Where("merchant_order_id = ?", callbackReq.MerchantOrderNo)
			db = db.Preload("PaymentMethod")
			return db
		},
	)
	if err != nil {
		return err
	}
	if order.Uuid == 0 {
		fmt.Println("支付订单不存在")
		return errors.New("支付订单不存在")
	}
	// 将PayAt转换为时间戳
	payAt, err := time.Parse("2006-01-02 15:04:05", callbackReq.PayAt)
	if err != nil {
		fmt.Println("支付时间格式错误", err)
		return fmt.Errorf("支付时间格式错误: %v", err)
	}

	// 将 paymentAmount 和 fee/100 转换为 decimal
	commissionFee := decimal.NewFromFloat(order.CommissionFee)
	orderAmount := decimal.NewFromFloat(order.OrderAmount)
	paymentFeePercent := commissionFee.Div(orderAmount).Truncate(3).Round(2).InexactFloat64()

	// 获取支付订单
	paymentOrderId := uint(0)
	paymentOrderUuid := order.PaymentOrderUuid
	paymentOrderRepo := repository.NewPaymentOrderRepo(db)
	paymentOrder, _ := paymentOrderRepo.GetPaymentOrder(
		paymentOrderRepo.WhereRelatedUuid(order.RelatedUuid),
		paymentOrderRepo.WherePaymentMethodUuid(order.PaymentMethodUuid),
	)
	if paymentOrder.Uuid > 0 {
		paymentOrderId = paymentOrder.ID
		paymentOrderUuid = paymentOrder.Uuid
	}

	// 更新支付订单状态
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 更新连连支付订单
		err := repository.NewLlPaymentOrderRepo(tx).Update(order.Uuid, map[string]interface{}{
			"order_status":       "PS",
			"pay_time":           payAt.Unix(),
			"payment_order_uuid": paymentOrderUuid,
		})
		if err != nil {
			return errors.WithMessage(err)
		}

		// 创建或更新支付单
		if paymentOrderId > 0 {
			paymentOrder.SetUpdate()
			paymentOrder.TransactionNumber = callbackReq.PaymentId
			paymentOrder.Status = constant.PaymentOrderStatusPaid
		} else {
			paymentOrder = model.PaymentOrder{
				BaseModel: model.BaseModel{
					ID:   paymentOrderId,
					Uuid: paymentOrderUuid,
				},
				PaymentMethodName: func() string {
					if order.PaymentMethod == nil {
						return ""
					}
					return order.PaymentMethod.Name
				}(),
				PaymentFeePercent:    paymentFeePercent,
				PaymentMethodUuid:    order.PaymentMethodUuid,
				RelatedType:          order.RelatedType,
				RelatedUuid:          order.RelatedUuid,
				CurrencyUnit:         order.OrderCurrency,
				PaymentAmount:        order.OrderAmount - order.CommissionFee,
				PaymentCommissionFee: order.CommissionFee,
				Amount:               order.OrderAmount,
				TransactionNumber:    callbackReq.PaymentId,
				Status:               constant.PaymentOrderStatusPaid,
			}
		}
		if err := repository.NewPaymentOrderRepo(tx).UpdateOrCreatePaymentOrderRecord(paymentOrder); err != nil {
			return err
		}

		// 会员端订单支付成功
		if order.MemberSaleOrderUuid > 0 {
			memberSaleOrder, err := repository.NewMemberSaleOrderRepo(db).GetMemberSaleOrderRecord(order.MemberSaleOrderUuid)
			if err != nil {
				return err
			}

			// 如果submit_pay_time为0，说明还没有设置过提交支付时间，此时设置为支付成功时间
			if memberSaleOrder.SubmitPayTime == 0 {
				if err := repository.NewMemberSaleOrderRepo(tx).UpdateMemberSaleOrderSubmitPayTime(memberSaleOrder.Uuid, payAt.Unix()); err != nil {
					return err
				}
			}

			// 更新订单状态
			event.NewSystemBus().PublishPayFinishMemberSaleOrderEvent(event.PayFinishMemberSaleOrderPayload{
				BasePayload: event.BasePayload{
					Ctx:                 p.ctx,
					CompanyUuid:         p.ctx.GetCompanyUuid(),
					Source:              constant.SourceMember,
					SaleBillUuid:        memberSaleOrder.SaleBillUuid,
					SaleOrderUuid:       memberSaleOrder.SaleOrderUuid,
					MemberSaleOrderUuid: memberSaleOrder.Uuid,
					MemberUuid:          memberSaleOrder.MemberUuid,
				},
				MemberSaleOrderUuid: memberSaleOrder.Uuid,
			})
		} else if order.RelatedType == constant.PaymentOrderRelatedTypeSaleOrder {
			// 非会员端外送订单支付成功（可能是会员端堂食订单或 Kiosk 自助点餐机订单）
			// 先查询销售订单获取 SaleBillUuid
			saleOrder, err := repository.NewSaleOrderRepo(db).GetSaleOrderByUuid(order.RelatedUuid)
			if err == nil && saleOrder != nil {
				// 查询销售账单
				saleBill, err := repository.NewOrderRepo(db).GetSaleBill(
					repository.CommonRepo.WhereByUuid(saleOrder.SaleBillUuid),
				)
				if err == nil {
					// 根据订单来源分发不同的事件
					switch saleBill.Source {
					case constant.SaleBillSourceMember:
						// 会员端堂食订单支付成功，发布事件
						event.NewSystemBus().PublishPayFinishMemberDineInOrderEvent(event.PayFinishMemberDineInOrderPayload{
							BasePayload: event.BasePayload{
								Ctx:           p.ctx,
								CompanyUuid:   p.ctx.GetCompanyUuid(),
								Source:        constant.SourceMember,
								SaleBillUuid:  saleBill.Uuid,
								SaleOrderUuid: order.RelatedUuid,
							},
							PaymentOrderUuid: paymentOrderUuid,
						})
						// case constant.SaleBillSourceKiosk:
						// 	// Kiosk 订单支付成功，发布事件触发送厨
						// 	event.NewSystemBus().PublishPayFinishKioskOrderEvent(event.PayFinishKioskOrderPayload{
						// 		BasePayload: event.BasePayload{
						// 			Ctx:           p.ctx,
						// 			CompanyUuid:   p.ctx.GetCompanyUuid(),
						// 			Source:        constant.SourceKiosk,
						// 			SaleBillUuid:  saleBill.Uuid,
						// 			SaleOrderUuid: order.RelatedUuid,
						// 		},
						// 		PaymentOrderUuid: paymentOrderUuid,
						// 	})
					}
				}
			}
		}

		return nil
	}); err != nil {
		fmt.Println("更新支付订单失败 - 03", err)
		logger.Logger.Error("更新支付订单失败 - 03", zap.Error(err))
		return err
	}
	return nil
}

// HandleRefundCallback 处理退款回调
func (p *PaymentRepo) HandleRefundCallback(sign string, callbackReq req.LianLianRefundCallbackRequest) error {
	// 验证签名
	err := p.validateRefundSign(sign, callbackReq)
	if err != nil {
		fmt.Println("验证签名失败", err)
		logger.Logger.Error("验证签名失败", zap.Error(err))
		return err
	}
	// 转换商户ID
	companyUuid, err := strconv.ParseUint(callbackReq.CompanyUuid, 10, 64)
	if err != nil {
		fmt.Println("商户ID格式错误", err)
		logger.Logger.Error("商户ID格式错误", zap.Error(err))
		return fmt.Errorf("商户ID格式错误: %v", err)
	}
	db := p.dbm.GetDB(companyUuid)
	//
	returnOrderRepo := repository.NewReturnOrderRepo(db)
	orderAmount, err := returnOrderRepo.GetReturnOrderAmount(
		returnOrderRepo.WithReturnOrder(),
		returnOrderRepo.WhereMerchantRefundOrderNo(callbackReq.MerchantRefundId),
	)
	if err != nil {
		fmt.Println("获取退货金额失败", err)
		logger.Logger.Error("获取退货金额失败", zap.Error(err))
		return err
	}
	// 验证退款状态
	if orderAmount.RefundStatus == 1 || orderAmount.RefundStatus == 2 {
		return nil
	}
	// 更新退款状态
	if callbackReq.RefundStatus == "RS" {
		orderAmount.RefundStatus = 1
	} else {
		orderAmount.RefundStatus = 2
	}
	err = returnOrderRepo.UpdateReturnOrderAmount([]repository.DBOption{returnOrderRepo.WhereUuid(orderAmount.Uuid)}, orderAmount)
	if err != nil {
		fmt.Println("更新退款状态失败", err)
		logger.Logger.Error("更新退款状态失败", zap.Error(err))
		return err
	}

	// Desk - AfterUpdate 更新桌台后的逻辑 - 推送桌台更新
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.UPDATE_REFUND_STATE, map[string]interface{}{
			"uuid":        orderAmount.Uuid,
			"update_time": orderAmount.BaseModel.UpdateTime,
		})
	})

	return nil
}

// getPaymentInfo 获取支付配置
func (p *PaymentRepo) getPaymentInfo(paymentMethodCode int) (string, string, error) {
	// 根据支付方式调用不同的支付接口
	var url string
	var orderType string
	// 支付方式
	if paymentMethodCode == constant.PaymentMethodCodeLianLianWechatPay {
		url = p.payServiceUrl + "/api/receipts/lianlianWechatPay"
		orderType = "LIANLIAN_WECHAT"
	} else if paymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
		url = p.payServiceUrl + "/api/receipts/lianlianAliOfflinePay"
		orderType = "LIANLIAN_ALI_OFFLINE_PAY"
	} else if paymentMethodCode == constant.PaymentMethodCodeLianLianQRPromptPay {
		url = p.payServiceUrl + "/api/receipts/lianlianQrPromptPay"
		orderType = "LIANLIAN_QR_PROMPT_PAY"
	} else {
		return "", "", errors.New("不支持的支付方式")
	}
	return url, orderType, nil
}

// validateConfigError 验证支付配置错误
func (p *PaymentRepo) ValidateConfigError(companyUuid uint64) error {
	_, paymentAppErr := p.validateConfig(companyUuid)
	if paymentAppErr != nil {
		return paymentAppErr
	}
	return nil
}

// validateConfig 验证支付配置
func (p *PaymentRepo) validateConfig(companyUuid uint64) (*model.PaymentApp, error) {
	paymentApp, paymentAppErr := saas.NewPaymentAppRepo(p.dbm.GetDB(0)).GetPaymentAppCompanyUuid(companyUuid)
	// 检查支付配置
	if paymentAppErr != nil || paymentApp == nil || paymentApp.ID == 0 {
		return nil, errors.New("未配置支付信息")
	}
	if p.payServiceUrl == "" {
		return nil, errors.New("未配置PAY_SERVICE_URL")
	}
	if p.payCallbackUrl == "" {
		return nil, errors.New("未配置PAY_SERVICE_LIANLIAN_CALLBACK_URL")
	}
	return paymentApp, nil
}

// ValidateSign 验证支付签名
func (p *PaymentRepo) validateSign(sign string, req req.LianLianCallbackRequest) error {
	companyUuid, err := strconv.ParseUint(req.CompanyUuid, 10, 64)
	if err != nil {
		return fmt.Errorf("商户ID格式错误: %v", err)
	}
	paymentApp, err := p.validateConfig(companyUuid)
	if err != nil {
		return err
	}
	// 组装请求数据
	jsonStr := fmt.Sprintf("{\"shop_supplier_id\":\"%v\",\"merchant_order_no\":\"%v\",\"merchant_user_id\":\"%v\",\"pay_type_desc\":\"%v\",\"pay_status\":%v,\"payment_id\":\"%v\",\"order_amount\":\"%v\",\"order_currency\":\"%v\",\"pay_at\":\"%v\"}",
		req.CompanyUuid,
		req.MerchantOrderNo,
		req.MerchantUserId,
		req.PayTypeDesc,
		req.PayStatus,
		req.PaymentId,
		req.OrderAmount,
		req.OrderCurrency,
		req.PayAt,
	)
	// 验证支付签名
	if p.requestSign(paymentApp.LlSignSalt, jsonStr) != sign {
		return errors.New("支付签名验证失败")
	}
	return nil
}

// validateRefundSign 验证退款签名
func (p *PaymentRepo) validateRefundSign(sign string, req req.LianLianRefundCallbackRequest) error {
	companyUuid, err := strconv.ParseUint(req.CompanyUuid, 10, 64)
	if err != nil {
		return fmt.Errorf("商户ID格式错误: %v", err)
	}
	paymentApp, err := p.validateConfig(companyUuid)
	if err != nil {
		return err
	}
	// 组装请求数据
	jsonStr := fmt.Sprintf("{\"shop_supplier_id\":\"%v\",\"refund_status\":\"%v\",\"refund_order_id\":\"%v\"}",
		req.CompanyUuid,
		req.RefundStatus,
		req.RefundOrderId,
	)
	// 验证支付签名
	if p.requestSign(paymentApp.LlSignSalt, jsonStr) != sign {
		return errors.New("支付签名验证失败")
	}
	return nil
}

// paySign 计算支付签名
func (p *PaymentRepo) requestSign(signSalt string, jsonStr string) string {
	// 将盐值和 JSON 字符串组合
	combinedStr := signSalt + jsonStr
	// 计算 SHA-256 哈希
	hash := sha256.Sum256([]byte(combinedStr))
	// 将哈希值转换为字符串
	hashStr := hex.EncodeToString(hash[:])
	// 截取前 32 个字符，等同于 PHP 的 substr(..., 0, 32)
	return hashStr[:32]
}

// postRequest 发送HTTP POST请求
func (p *PaymentRepo) postRequest(url string, jsonData string, headers map[string]string, timeout int) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("请求失败:支付服务异常")
	}

	// 解析响应
	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		logger.Logger.Error("postRequest-json.Unmarshal", zap.Error(err))
		logger.Logger.Error("postRequest-json.Unmarshal", zap.String("body", string(body)))
		return nil, errors.New("解析响应失败")
	}
	// 检查返回结果
	code, ok := responseMap["code"].(float64)
	if !ok || code != 1 {
		msg, _ := responseMap["msg"].(string)
		if msg == "" {
			msg = "请求失败"
		}
		return nil, errors.New(msg)
	}

	// 获取响应数据
	responseData, ok := responseMap["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("响应数据格式错误")
	}

	return responseData, nil
}

// MemberSaleOrderRefund 会员端销售单发起退款
type MemberSaleOrderRefundReq struct {
	CancelReason string `json:"cancel_reason"` // 取消原因
	BankCode     string `json:"bank_code"`     // 银行代码 - 暂时不用
	AccountNo    string `json:"account_no"`    // 账号 - 暂时不用
	AccountName  string `json:"account_name"`  // 账户名称- 暂时不用
}

func (p *PaymentRepo) MemberSaleOrderRefund(saleOrder model.SaleOrder, req MemberSaleOrderRefundReq) (*model.ReturnOrder, error) {
	db := p.ctx.GetDB()
	// 创建退货单
	returnOrder := model.ReturnOrder{
		BaseModel: model.BaseModel{
			Uuid: func() uint64 {
				id, _ := utils.GetID()
				return id
			}(),
		},
		RelatedOrderType: constant.ReturnOrderRelatedOrderTypeSaleOrder,
		RelatedOrderUuid: saleOrder.Uuid,
		RelatedOrderNo:   saleOrder.OrderNo,
		ReturnType:       constant.ReturnOrderRefundTypeTotal,
		RefundAmount:     saleOrder.Amount,
		RefundReason:     req.CancelReason,
		ReturnOrderAmounts: func() []model.ReturnOrderAmount {
			var returnOrderAmounts []model.ReturnOrderAmount
			for _, paymentOrder := range saleOrder.PaymentOrders {
				if paymentOrder.IsDelete() {
					continue
				}
				returnOrderAmounts = append(returnOrderAmounts, model.ReturnOrderAmount{
					BaseModel: model.BaseModel{
						Uuid: func() uint64 {
							id, _ := utils.GetID()
							return id
						}(),
					},
					PaymentMethodUuid:     paymentOrder.PaymentMethodUuid,
					Amount:                paymentOrder.Amount,
					PaymentOrderUuid:      paymentOrder.Uuid,
					MerchantRefundOrderNo: utils.GenerateMerchantOrderNo("RE"),
					PaymentMethod:         &model.PaymentMethod{Code: paymentOrder.PaymentMethod.Code, PaymentName: paymentOrder.PaymentMethodName},
				})
			}
			return returnOrderAmounts
		}(),
		BankCode:    req.BankCode,
		AccountNo:   req.AccountNo,
		AccountName: req.AccountName,
	}
	returnOrderUuid, err := repository.NewReturnOrderRepo(db).CreateReturnOrderRecord(returnOrder)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	// 发起退款
	for _, returnOrderAmount := range returnOrder.ReturnOrderAmounts {
		refund, err := p.Refund(PaymentServiceRefundReq{
			RelatedType:           constant.PaymentOrderRelatedTypeSaleOrder, // 相关类型
			PaymentOrderUuid:      returnOrderAmount.PaymentOrderUuid,        // 支付订单UUID
			MerchantRefundOrderNo: returnOrderAmount.MerchantRefundOrderNo,   // 商户退款订单号
			RefundAmount:          returnOrderAmount.Amount,                  // 退款金额
			RefundOrderId:         returnOrderAmount.LlReturnOrderid,         // 退款ID
			BankCode:              returnOrder.BankCode,
			AccountNo:             returnOrder.AccountNo,
			AccountName:           returnOrder.AccountName,
		})
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		// 创建退款金额
		returnOrderAmount.ReturnOrderUuid = returnOrderUuid
		returnOrderAmount.LlReturnOrderid = refund.RefundOrderId
		if err = repository.NewReturnOrderRepo(db).CreateReturnOrderAmount([]model.ReturnOrderAmount{returnOrderAmount}); err != nil {
			return nil, errors.WithMessage(err)
		}
	}
	return &returnOrder, nil
}

// GetSignSaltParams 获取签名盐参数
type GetSignSaltParams struct {
	LlWhiteIp            string `json:"ll_white_ip"`
	LlMerchantId         string `json:"ll_merchant_id"`
	LlPublicKey          string `json:"ll_public_key"`
	LlMerchantPrivateKey string `json:"ll_merchant_private_key"`
	LlToken              string `json:"ll_token"`
	LlStoreId            string `json:"ll_store_id"`
	ShopSupplierId       uint64 `json:"shop_supplier_id"`
}

// GetSignSaltResponse 获取签名盐响应
type GetSignSaltResponse struct {
	SignSalt string `json:"sign_salt"`
}

// GetSignSalt 获取签名盐
func (p *PaymentRepo) GetSignSalt(params GetSignSaltParams) (string, error) {
	if p.payServiceIp == "" || p.payServiceUrl == "" || p.payServiceRsaPublicKey == "" {
		return "", errors.New("支付服务环境配置错误")
	}
	if params.LlWhiteIp != p.payServiceIp {
		return "", errors.New("白名单IP不匹配")
	}
	plaintext, err := json.Marshal(params)
	if err != nil {
		logger.Logger.Error("GetSignSalt-json.Marshal", zap.Error(err))
		return "", errors.WithMessage(err, "参数序列化失败")
	}

	rsaPublicKey, err := encrypt.GetPublicKey(p.payServiceRsaPublicKey)
	if err != nil {
		logger.Logger.Error("GetSignSalt-encrypt.GetPublicKey", zap.Error(err))
		return "", errors.WithMessage(err, "支付服务环境配置错误")
	}

	encryptedData, err := encrypt.PubEncrypt(string(plaintext), rsaPublicKey)
	if err != nil {
		logger.Logger.Error("GetSignSalt-encrypt.PubEncrypt", zap.Error(err))
		return "", errors.WithMessage(err, "加密数据失败")
	}

	type RequestData struct {
		EncryptData string `json:"encrypt_data"`
	}
	requestData := RequestData{
		EncryptData: encryptedData,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		logger.Logger.Error("GetSignSalt-json.Marshal", zap.Error(err))
		return "", errors.WithMessage(err, "参数序列化失败")
	}

	response, err := p.postRequest(p.payServiceUrl+p.payServiceAuthUri, string(jsonData), map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}, RequestTimeOut)
	if err != nil {
		logger.Logger.Error("GetSignSalt-postRequest", zap.Error(err))
		return "", errors.WithMessage(err, "请求支付服务失败")
	}

	// 解析响应
	var resp GetSignSaltResponse
	if err := utils.MapToStruct(response, &resp); err != nil {
		logger.Logger.Error("GetSignSalt-MapToStruct", zap.Error(err))
		return "", errors.WithMessage(err, "响应数据格式错误")
	}

	return resp.SignSalt, nil
}

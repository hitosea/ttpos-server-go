package lianlian

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/admin"
	contexts "ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"

	"github.com/spf13/viper"
)

// 请求超时
const REQUEST_TIME_OUT = 10

// 支付方式
const (
	LIANLIAN_WECHAT_PAY      = "/api/receipts/lianlianWechatPay"
	LIANLIAN_PROMPT_PAY      = "/api/receipts/lianlianQrPromptPay"
	LIANLIAN_ALI_OFFLINE_PAY = "/api/receipts/lianlianAliOfflinePay"
)

// 支付类型
const (
	CASHIER_ORDER  = 1
	RECHARGE_ORDER = 2
)

// PaymentRepo 连连支付仓库
type PaymentRepo struct {
	dbm                           *database.DBManager
	ctx                           contexts.Context
	payServiceUrl                 string
	payServiceLianlianCallbackUrl string
}

// LianLianPaymentRepo 连连支付仓库
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
		// 二维码有效期 微信(90111)-60分 支付宝(90222)-15分 promptPay(90333)-8分
		// $alive_time = [
		//     '90111' =>  60 * 60,
		//     '90222' =>  60 * 15,
		//     '90333' =>  60 * 8,
		// ];
	} `json:"order"`
}

// NewPaymentRepo 创建连连支付仓库
func NewPaymentRepo(ctx contexts.Context, dbm *database.DBManager) *PaymentRepo {
	return &PaymentRepo{
		ctx:                           ctx,
		dbm:                           dbm,
		payServiceUrl:                 viper.GetString("PAY_SERVICE_URL"),
		payServiceLianlianCallbackUrl: viper.GetString("PAY_SERVICE_LIANLIAN_CALLBACK_URL"),
	}
}

// CreatePayment 创建支付
func (p *PaymentRepo) CreatePayment(
	relatedType int,
	relatedUuid uint64,
	paymentMethodCode int,
	paymentAmount float64,
	commissionFee float64,
) (*model.LlPaymentOrder, error) {
	paymentApp, err := p.ValidateConfig()
	if err != nil {
		return nil, err
	}

	// 创建支付订单仓库
	llPaymentOrderRepo := repository.NewLlPaymentOrderRepo(p.dbm.GetDB(p.ctx.GetDbId()))
	merchantUserId := fmt.Sprintf("%d", p.ctx.GetStaffUuid())
	merchantOrderNo := p.GenerateMerchantOrderNo()

	// 获取支付配置
	url, orderType, err := p.getPaymentInfo(paymentMethodCode)
	if err != nil {
		return nil, err
	}

	// 判断是否已存在有效待支付二维码
	order, err := p.GetValidPaymentOrderByUuid(relatedType, relatedUuid, orderType, merchantUserId, paymentAmount)
	if err != nil {
		return nil, err
	}
	if order != nil {
		return order, nil
	}

	// 组装请求数据
	jsonStr := fmt.Sprintf("{\"shop_supplier_id\":%v,\"merchant_order_no\":\"%v\",\"order_amount\":\"%v\",\"order_currency\":\"%v\",\"order_desc\":\"%v\",\"full_name\":\"%v\",\"merchant_user_id\":%v,\"callback_url\":\"%v\"}",
		p.ctx.GetCompanyUuid(),
		merchantOrderNo,
		paymentAmount,
		"THB",
		"CHECK_OUT",
		"CASHIER",
		p.ctx.GetStaffUuid(),
		strings.ReplaceAll(p.payServiceLianlianCallbackUrl, "/", "\\/"),
	)

	// 计算签名
	response, err := p.postRequest(url, jsonStr, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
		"sign":         p.paySign(paymentApp.LlSignSalt, jsonStr),
	}, REQUEST_TIME_OUT)
	if err != nil {
		return nil, err
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
	if paymentMethodCode == constant.PaymentMethodCodeLianLianWechatPay || paymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
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
		if paymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
			resp.Order.CreateTime = resp.Order.OrderCreateTime
		}
	} else {
		// PromptPay直接使用返回的QR code
		resp.Order.QrCode = "data:image/png;base64," + resp.Order.QrCode
	}

	// 生成雪花ID
	uuid, err := utils.GetID()
	if err != nil {
		return nil, errors.New("生成雪花ID失败")
	}

	// 创建支付订单
	paymentOrder := &model.LlPaymentOrder{
		PaymentOrderUuid: uuid,
		RelatedType:      relatedType,
		RelatedUuid:      relatedUuid,
		MerchantOrderId:  merchantOrderNo,
		MerchantId:       resp.Order.MerchantId,
		OrderId:          resp.Order.OrderId,
		OrderType:        resp.PayTypeDesc,
		OrderStatus:      resp.Order.OrderStatus,
		OrderAmount:      paymentAmount,
		OrderCurrency:    resp.Order.OrderCurrency,
		FullName:         "CASHIER",
		OrderDesc:        resp.PayTypeDesc,
		LinkUrl:          resp.Order.QrCode,
		MerchantUserId:   merchantUserId,
		LlCreateTime:     resp.Order.CreateTime,
		CommissionFee:    commissionFee,
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

// GetValidPaymentOrder 获取有效的支付订单
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
			db = db.Where("order_currency = ?", "THB")
			db = db.Where("pay_time = ?", constant.No)
			db = db.Where("expired_time > ?", time.Now().Unix()+5)
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

// getPaymentInfo 获取支付配置
func (p *PaymentRepo) getPaymentInfo(paymentMethodCode int) (string, string, error) {
	// 根据支付方式调用不同的支付接口
	var url string
	var orderType string
	if paymentMethodCode == constant.PaymentMethodCodeLianLianWechatPay {
		url = p.payServiceUrl + LIANLIAN_WECHAT_PAY
		orderType = "LIANLIAN_WECHAT"
	} else if paymentMethodCode == constant.PaymentMethodCodeLianLianAliPay {
		url = p.payServiceUrl + LIANLIAN_ALI_OFFLINE_PAY
		orderType = "LIANLIAN_ALI_OFFLINE_PAY"
	} else if paymentMethodCode == constant.PaymentMethodCodeLianLianQRPromptPay {
		url = p.payServiceUrl + LIANLIAN_PROMPT_PAY
		orderType = "LIANLIAN_QR_PROMPT_PAY"
	} else {
		return "", "", errors.New("不支持的支付方式")
	}
	return url, orderType, nil
}

// Validate 验证支付配置
func (p *PaymentRepo) ValidateConfig() (*model.PaymentApp, error) {
	db := p.dbm.GetDB(0)
	paymentApp, paymentAppErr := admin.NewPaymentAppRepo(db).GetPaymentAppCompanyUuid(p.ctx.GetCompanyUuid())
	// 检查支付配置
	if paymentAppErr != nil || paymentApp == nil || paymentApp.ID == 0 {
		return nil, errors.New("未配置支付信息")
	}
	if p.payServiceUrl == "" {
		return nil, errors.New("未配置PAY_SERVICE_URL")
	}
	if p.payServiceLianlianCallbackUrl == "" {
		return nil, errors.New("未配置PAY_SERVICE_LIANLIAN_CALLBACK_URL")
	}
	return paymentApp, nil
}

// paySign 计算支付签名
func (p *PaymentRepo) paySign(signSalt string, jsonStr string) string {
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

/**
* 请求支付订单号
* @return string
 */
func (p *PaymentRepo) GenerateMerchantOrderNo() string {
	prefix := "PS"
	datePart := time.Now().Format("20060102150405")
	randomPart := fmt.Sprintf("%08d", rand.Intn(100000000))
	return prefix + datePart + randomPart
}

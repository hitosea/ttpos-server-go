<?php

namespace app\common\service\payment;

use app\common\model\order\UserRechargeOrder;
use app\common\model\order\UserRechargeOrderPayType;
use help\HttpHelp;
use app\common\model\order\Order;
use app\common\model\store\PayType;
use app\common\service\BaseService;
use app\common\model\pay\PaymentApp;
use app\common\model\settings\Setting;
use app\common\exception\BaseException;
use app\common\model\order\OrderPayType;
use app\common\enum\order\OrderStatusEnum;
use app\common\model\payment\PaymentOrder;

/**
 * 支付服务
 */
class PaymentService extends BaseService
{

    public $order;
    public int $order_type; // 订单支付类型 1-点餐订单 2-会员充值订单

    // 请求超时
    const REQUEST_TIME_OUT = 10;
    // 支付方式
    const LIANLIAN_WECHAT_PAY = '/api/receipts/lianlianWechatPay';
    const LIANLIAN_PROMPT_PAY = '/api/receipts/lianlianQrPromptPay';
    const LIANLIAN_ALI_OFFLINE_PAY = '/api/receipts/lianlianAliOfflinePay';
    // 支付类型
    const CASHIER_ORDER = 1;
    const RECHARGE_ORDER = 2;

    public function __construct($order_id, $pay_type_value, $type = self::CASHIER_ORDER, $payPice = 0)
    {
        // 收银订单支付
        if ($type == self::CASHIER_ORDER) {
            $detail = Order::detail([
                ['order_id', '=', $order_id],
                ['order_status', '=', OrderStatusEnum::NORMAL]
            ]);
            if (!$detail) {
                throw new BaseException(['msg' => '订单状态变动，请重新查看']);
            }
            // 配置
            if (!self::checkSignSalt($detail->shop_supplier_id)) {
                throw new BaseException(['msg' => '未配置PAY_SERVICE_SIGN_SALT']);
            }
            if (!self::checkServiceUrl()) {
                throw new BaseException(['msg' => '未配置PAY_SERVICE_URL']);
            }
            if (in_array($pay_type_value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY])) {
                if (!env('PAY_SERVICE_LIANLIAN_CALLBACK_URL')) {
                    throw new BaseException(['msg' => 'env未配置PAY_SERVICE_LIANLIAN_CALLBACK_URL']);
                }
                if (OrderPayType::where('order_id', $order_id)->where('value', $pay_type_value)->find()) {
                    throw new BaseException(['msg' => '当前支付已完成，请选择其他方式支付']);
                }
            }
            $opt_price = OrderPayType::where('order_id', $detail->order_id)->sum('price');
            if ($opt_price >= $detail->pay_price) {
                throw new BaseException(['msg' => '订单已足额']);
            }
            if (($opt_price + $payPice) > $detail->pay_price) {
                throw new BaseException(['msg' => '超出订单剩余可支付金额']);
            }
            $this->order = $detail;
            $this->order_type = self::CASHIER_ORDER;
        }
        // 会员充值订单支付
        else {
            $detail = UserRechargeOrder::detail($order_id);
            if (!$detail) {
                throw new BaseException(['msg' => '订单状态变动，请重新查看']);
            }
            // 配置
            if (!self::checkSignSalt($detail->shop_supplier_id)) {
                throw new BaseException(['msg' => '未配置PAY_SERVICE_SIGN_SALT']);
            }
            if (!self::checkServiceUrl()) {
                throw new BaseException(['msg' => '未配置PAY_SERVICE_URL']);
            }
            if (in_array($pay_type_value, [PayType::SOURCE_LIANLIAN_WECHAT_PAY, PayType::SOURCE_LIANLIAN_ALI_PAY, PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY])) {
                if (!env('PAY_SERVICE_LIANLIAN_CALLBACK_URL')) {
                    throw new BaseException(['msg' => 'env未配置PAY_SERVICE_LIANLIAN_CALLBACK_URL']);
                }
                if (UserRechargeOrderPayType::where('order_id', $order_id)->where('value', $pay_type_value)->find()) {
                    throw new BaseException(['msg' => '当前支付已完成，请选择其他方式支付']);
                }
            }
            $opt_price = UserRechargeOrderPayType::where('order_id', $detail->order_id)->sum('price');
            if ($opt_price >= $detail->pay_price) {
                throw new BaseException(['msg' => '订单已足额']);
            }
            if (($opt_price + $payPice) > $detail->pay_price) {
                throw new BaseException(['msg' => '超出订单剩余可支付金额']);
            }
            $this->order = $detail;
            $this->order_type = self::RECHARGE_ORDER;
        }

    }

    /**
     * 检查PAY_SERVICE_SIGN_SALT
     */
    public static function checkSignSalt($shop_supplier_id = 0)
    {
        // 不能合并，以saas端为准
        return (new PaymentApp([], 0))->value('ll_sign_salt') ?: env('PAY_SERVICE_SIGN_SALT') ?: '';
    }

    /**
     * 检查PAY_SERVICE_URL
     */
    public static function checkServiceUrl()
    {
        // 不能合并，以saas端为准
        return env('PAY_SERVICE_URL') ?? '';
    }

    /**
     * LianLian 微信支付
     * @param $data
     * @return array|false
     */
    public function LianLianWechatPay($data)
    {
        $domainName = self::checkServiceUrl();
        $data['callback_url'] = env('PAY_SERVICE_LIANLIAN_CALLBACK_URL');
        $data['order_currency'] = "THB";
        $request_data = $this->assembleLianLianWechatPayRequest($data);
        $url = $domainName . self::LIANLIAN_WECHAT_PAY;
        //
        $response = HttpHelp::postRequest($url, json_encode($request_data, true), [
            'Content-Type: application/json; charset=utf-8',
            'sign: ' . $this->paySign($request_data),
        ], self::REQUEST_TIME_OUT);
        if (!$response) {
            $this->error = '请求失败:支付服务异常';
            return false;
        }
        $response_arr = json_decode($response, true);
        if ($response_arr['code'] != 1) {
            $this->error = isset($response_arr['msg']) ? $response_arr['msg'] : '请求失败:';
            return false;
        }
        //
        $paymentOrder = $this->addLianLianWechatPay($data, $response_arr['data']);

        return $this->assembleLianLianPayResponse($paymentOrder, $response_arr['data']);
    }

    // LianLian PromptPay
    public function LianLianPromptPay($data)
    {
        $domainName = self::checkServiceUrl();
        $data['callback_url'] = env('PAY_SERVICE_LIANLIAN_CALLBACK_URL');
        $data['order_currency'] = "THB";
        $request_data = $this->assembleLianLianWechatPayRequest($data);
        $url = $domainName . self::LIANLIAN_PROMPT_PAY;
        //
        $response = HttpHelp::postRequest($url, json_encode($request_data, true), [
            'Content-Type: application/json; charset=utf-8',
            'sign: ' . $this->paySign($request_data),
        ], self::REQUEST_TIME_OUT);
        if (!$response) {
            $this->error = '请求失败:支付服务异常';
            return false;
        }
        $response_arr = json_decode($response, true);
        if ($response_arr['code'] != 1) {
            $this->error = isset($response_arr['msg']) ? $response_arr['msg'] : '请求失败:';
            return false;
        }
        //
        $paymentOrder = $this->addLianLianPromptPay($data, $response_arr['data']);
        //
        return $this->assembleLianLianPayResponse($paymentOrder, $response_arr['data']);
    }

    /**
     * LianLian 微信支付
     * @param $data
     * @return array|false
     */
    public function LianLianAliOfflinePay($data)
    {
        $domainName = self::checkServiceUrl();
        $data['callback_url'] = env('PAY_SERVICE_LIANLIAN_CALLBACK_URL');
        $data['order_currency'] = "THB";
        $request_data = $this->assembleLianLianWechatPayRequest($data);
        $url = $domainName . self::LIANLIAN_ALI_OFFLINE_PAY;
        //
        $response = HttpHelp::postRequest($url, json_encode($request_data, true), [
            'Content-Type: application/json; charset=utf-8',
            'sign: ' . $this->paySign($request_data),
        ], self::REQUEST_TIME_OUT);
        if (!$response) {
            $this->error = '请求失败:支付服务异常';
            return false;
        }
        $response_arr = json_decode($response, true);
        if ($response_arr['code'] != 1) {
            $this->error = isset($response_arr['msg']) ? $response_arr['msg'] : '请求失败:';
            return false;
        }
        //
        $paymentOrder = $this->addLianLianAliOfflinePay($data, $response_arr['data']);

        return $this->assembleLianLianPayResponse($paymentOrder, $response_arr['data']);
    }

    /**
     * lianlian微信请求数据
     * @param $data
     * @return array
     */
    public function assembleLianLianWechatPayRequest($data)
    {
        return [
            "shop_supplier_id" => $data['shop_supplier_id'],
            "merchant_order_no" => $data['merchant_order_no'],
            "order_amount" => $data['order_amount'],
            "order_currency" => $data['order_currency'],
            "order_desc" => $data['order_desc'],
            "full_name" => $data['full_name'],
            "merchant_user_id" =>  $data['merchant_user_id'],
            "callback_url" => $data['callback_url'],
        ];
    }

    /**
     * lianlian 支付请求返回数据
     * @param $data
     * @return array
     */
    public function assembleLianLianPayResponse(PaymentOrder $paymentOrder, $data)
    {
        return [
            "payment_order_id" => $paymentOrder->id,
            "order_id" => $this->order->order_id,
            "pay_type_desc" => isset($data['pay_type_desc']) ? $data['pay_type_desc'] : '',
            "link_url" => isset($data['order']['link_url']) ? $data['order']['link_url'] : '',
            "qr_code" => isset($data['order']['qr_code']) ? $data['order']['qr_code'] : '',
            "qr_code_expire_sec" => isset($data['order']['qr_code_expire_sec']) ? $data['order']['qr_code_expire_sec'] : 0,
        ];
    }

    /**
     * 支付服务回调参数验签
     * @param array $notify_data
     * @param $sign
     * @return bool
     */
    public static function checkCallbackSign(array $notify_data, $sign)
    {
        $raw_string = json_encode($notify_data, JSON_UNESCAPED_UNICODE);
        $shop_supplier_id = $notify_data['shop_supplier_id'] ?? 0;
        $sign_salt = self::checkSignSalt($shop_supplier_id);
        $raw_hash = substr(hash('sha256', $sign_salt . $raw_string), 0, 32);
        return $raw_hash == $sign;
    }

    /**
     * 发起支付请求签名
     * @param array $notify_data
     * @param $sign
     * @return string
     */
    public function paySign(array $request_data)
    {
        $raw_string = json_encode($request_data, JSON_UNESCAPED_UNICODE);
        $shop_supplier_id = $request_data['shop_supplier_id'] ?? 0;
        $sign_salt = self::checkSignSalt($shop_supplier_id);
        return substr(hash('sha256', $sign_salt . $raw_string), 0, 32);
    }

    /**
     * 请求支付订单号
     * @return string
     */
    public function generateMerchantOrderNo()
    {
        $prefix = 'PS';
        $datePart = date('YmdHis'); // 获取当前日期
        $randomPart = str_pad(rand(0, 99999999), 8, '0', STR_PAD_LEFT); // 生成一个8位的随机数
        return $prefix . $datePart . $randomPart;
    }

    /**
     * 添加wechatPay订单记录
     */
    public function addLianLianWechatPay(array $request_params, array $response_params)
    {
        $paymentOrder = new PaymentOrder;
        $saveArr = [
            'merchant_order_no' => $request_params['merchant_order_no'],
            'merchant_user_id' => $request_params['merchant_user_id'],
            'order_id' => $this->order_type == self::CASHIER_ORDER ? $this->order->order_id : 0,
            'recharge_order_id' => $this->order_type == self::RECHARGE_ORDER ? $this->order->id : 0,
            'order_status' => PaymentOrder::PENDING,
            'source' => PayType::SOURCE_LIANLIANPAY,
            'source_pay_method' => PayType::SOURCE_LIANLIAN_PAY_METHOD[PayType::SOURCE_LIANLIAN_WECHAT_PAY],
            'order_amount' => $request_params['order_amount'],
            'order_currency' => $request_params['order_currency'],
            'full_name' => $request_params['full_name'],
            'order_desc' => $request_params['order_desc'],
            'callback_url' => $request_params['callback_url'],
            'request_params' => json_encode($request_params, JSON_UNESCAPED_UNICODE),
            'response' => json_encode($response_params, JSON_UNESCAPED_UNICODE),
            'link_url' => $response_params['order']['link_url'],
            'pay_type_value' => $request_params['pay_type_value'],
            'pay_type_price' => $request_params['pay_type_price'],
            'pay_type_fee' => $request_params['pay_type_fee'],
            'pay_type_fee_money' => $request_params['pay_type_fee_money'],
        ];
        $paymentOrder->save($saveArr);
        return $paymentOrder;
    }

    /**
     * 添加PromptPay订单记录
     */
    public function addLianLianPromptPay(array $request_params, array $response_params)
    {
        $paymentOrder = new PaymentOrder;
        $saveArr = [
            'merchant_order_no' => $request_params['merchant_order_no'],
            'merchant_user_id' => $request_params['merchant_user_id'],
            'order_id' => $this->order_type == self::CASHIER_ORDER ? $this->order->order_id : 0,
            'recharge_order_id' => $this->order_type == self::RECHARGE_ORDER ? $this->order->id : 0,
            'order_status' => PaymentOrder::PENDING,
            'source' => PayType::SOURCE_LIANLIANPAY,
            'source_pay_method' => PayType::SOURCE_LIANLIAN_PAY_METHOD[PayType::SOURCE_LIANLIAN_QR_PROMPT_PAY],
            'order_amount' => $request_params['order_amount'],
            'order_currency' => $request_params['order_currency'],
            'full_name' => $request_params['full_name'],
            'order_desc' => $request_params['order_desc'],
            'callback_url' => $request_params['callback_url'],
            'request_params' => json_encode($request_params, JSON_UNESCAPED_UNICODE),
            'response' => json_encode($response_params, JSON_UNESCAPED_UNICODE),
            'qr_code' => $response_params['order']['qr_code'],
            'qr_code_expire_sec' => $response_params['order']['qr_code_expire_sec'],
            'pay_type_value' => $request_params['pay_type_value'],
            'pay_type_price' => $request_params['pay_type_price'],
            'pay_type_fee' => $request_params['pay_type_fee'],
            'pay_type_fee_money' => $request_params['pay_type_fee_money'],
        ];
        $paymentOrder->save($saveArr);
        return $paymentOrder;
    }

    /**
     * 添加AliPay订单记录
     */
    public function addLianLianAliOfflinePay(array $request_params, array $response_params)
    {
        $paymentOrder = new PaymentOrder;
        $saveArr = [
            'merchant_order_no' => $request_params['merchant_order_no'],
            'merchant_user_id' => $request_params['merchant_user_id'],
            'order_id' => $this->order_type == self::CASHIER_ORDER ? $this->order->order_id : 0,
            'recharge_order_id' => $this->order_type == self::RECHARGE_ORDER ? $this->order->id : 0,
            'order_status' => PaymentOrder::PENDING,
            'source' => PayType::SOURCE_LIANLIANPAY,
            'source_pay_method' => PayType::SOURCE_LIANLIAN_PAY_METHOD[PayType::SOURCE_LIANLIAN_ALI_PAY],
            'order_amount' => $request_params['order_amount'],
            'order_currency' => $request_params['order_currency'],
            'full_name' => $request_params['full_name'],
            'order_desc' => $request_params['order_desc'],
            'callback_url' => $request_params['callback_url'],
            'request_params' => json_encode($request_params, JSON_UNESCAPED_UNICODE),
            'response' => json_encode($response_params, JSON_UNESCAPED_UNICODE),
            'link_url' => $response_params['order']['link_url'],
            'pay_type_value' => $request_params['pay_type_value'],
            'pay_type_price' => $request_params['pay_type_price'],
            'pay_type_fee' => $request_params['pay_type_fee'],
            'pay_type_fee_money' => $request_params['pay_type_fee_money'],
        ];
        $paymentOrder->save($saveArr);
        return $paymentOrder;
    }

    /**
     * 获取有效可支付订单
     * @param $order_id
     * @param $merchant_user_id
     * @param $pay_type_value
     * @param $pay_type_price
     * @return array|mixed
     */
    public static function alivePaymentOrder($order_id, $merchant_user_id, $pay_type_value, $pay_type_price, $pay_type_fee)
    {
        // 二维码有效期 微信(90111)-60分 支付宝(90222)-15分 promptPay(90333)-8分
        $alive_time = [
            '90111' =>  60 * 60,
            '90222' =>  60 * 15,
            '90333' =>  60 * 8,
        ];
        $now_expire_time = time() - $alive_time[$pay_type_value];

        return PaymentOrder::where('order_id', $order_id)
            ->where('merchant_user_id', $merchant_user_id)
            ->where('pay_type_value', $pay_type_value)
            ->where('pay_type_price', $pay_type_price)
            ->where('pay_type_fee', $pay_type_fee)
            ->where('create_time', '>', $now_expire_time)
            ->order('id', 'desc')
            ->find();
    }

    /**
     * 组装支付订单格式
     * @param PaymentOrder $paymentOrder
     * @return string[]
     */
    public static function assemblePaymentOrder(PaymentOrder $paymentOrder)
    {
        return [
            "link_url" => $paymentOrder->link_url,
            "order_id" => $paymentOrder->order_id,
            "pay_type_desc" => $paymentOrder->source_pay_method,
            "payment_order_id" => $paymentOrder->id,
            "qr_code" => $paymentOrder->qr_code,
            "qr_code_expire_sec" => $paymentOrder->qr_code_expire_sec,
        ];
    }
}

<?php

namespace app\common\service\payment;

use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\payment\PaymentOrder;
use app\common\service\BaseService;
use app\common\model\pay\PaymentApp;
use app\common\exception\BaseException;
use help\HttpHelp;
use think\facade\Log;

/**
 * 退款服务
 */
class RefundService extends BaseService
{

    // 请求超时
    const REQUEST_TIME_OUT = 10;
    // 接口
    const LIANLIAN_REFUND = '/api/receipts/lianlianRefund';

    public function __construct($shop_supplier_id = 0)
    {
        $shop_supplier_id = $shop_supplier_id == 0 ? request()->shopSupplierId : $shop_supplier_id;
        // 配置
        if (!self::checkSignSalt($shop_supplier_id)) {
            Log::write('未配置PAY_SERVICE_SIGN_SALT');
            throw new BaseException(['msg' => '未配置PAY_SERVICE_SIGN_SALT']);
        }
        if (!self::checkServiceUrl()) {
            Log::write('未配置PAY_SERVICE_URL');
            throw new BaseException(['msg' => '未配置PAY_SERVICE_URL']);
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
     * @param $payment_order_id
     * @param $refund_money
     * @param $merchant_refund_id           // LianLian的唯一退款id
     * @param $merchant_refund_order_no     // 用于查询支付服务退单记录
     * @param $bank_code
     * @param $account_no
     * @param $account_name
     * @return array|false
     */
    public function refund($payment_order_id, $refund_money, $merchant_refund_id = '', $merchant_refund_order_no = '', $bank_code = '', $account_no = '', $account_name = '')
    {
        $paymentOrder = PaymentOrder::where('id', $payment_order_id)->find();
        if (!$paymentOrder) {
            throw new BaseException(['msg' => '支付记录不存在']);
        }
        if ($refund_money > $paymentOrder->order_amount) {
            throw new BaseException(['msg' => '退款金额不能大于支付金额']);
        }
        if (!$callbackUrl = env('PAY_SERVICE_LIANLIAN_REFUND_CALLBACK_URL')) {
            throw new BaseException(['msg' => '未配置 PAY_SERVICE_LIANLIAN_REFUND_CALLBACK_URL']);
        }
        //
        $requestData = [
            'shop_supplier_id' => $paymentOrder->shop_supplier_id,
            'merchant_order_no' => $paymentOrder->merchant_order_no,
            'merchant_refund_id' => $merchant_refund_id,                // lianlian的唯一退款id
            'merchant_refund_order_no' => $merchant_refund_order_no,    // 用于查询支付服务退单记录
            'refund_amount' => $refund_money,
            'refund_currency' => $paymentOrder->order_currency,
            'callback_url' => $callbackUrl . '?payment_order_id=' . $payment_order_id,
        ];
        //
        if ($paymentOrder->pay_type_value == OrderPayTypeEnum::LIANLIAN_QR_PROMPT_PAY) {
            $requestData['bank_code'] = $bank_code;
            $requestData['account_no'] = $account_no;
            $requestData['account_name'] = $account_name;
        }
        //
        return $this->LianLianRefund($requestData);
    }

    /**
     * LianLian 退款
     * @param $data
     *  shop_supplier_id
        merchant_order_no
        refund_amount
        refund_currency
        refund_reason
        callback_url
     *
     * @return array|false
     */
    public function LianLianRefund($data)
    {
        $domainName = self::checkServiceUrl();
        $data['order_currency'] = "THB";
        $request_data = $this->assembleLianLianRefundRequest($data);
        $url = $domainName . self::LIANLIAN_REFUND;
        //
        $response = HttpHelp::postRequest($url, json_encode($request_data, true), [
            'Content-Type: application/json; charset=utf-8',
            'sign: ' . $this->requestSign($request_data),
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
        return $this->assembleLianLianRefundResponse($response_arr['data']);
    }

    /**
     * lianlian退款请求数据
     * @param $data
     * @return array
     */
    public function assembleLianLianRefundRequest($data)
    {
        //  注意数组参数顺序要和支付服务一致，不然影响支验签
        $reData = [
            "shop_supplier_id" => $data['shop_supplier_id'],
            "merchant_order_no" => $data['merchant_order_no'],
            "merchant_refund_id" => $data['merchant_refund_id'],
            "refund_amount" => $data['refund_amount'],
            "refund_currency" => $data['refund_currency'],
            "refund_reason" => __('其他'),
            "callback_url" => $data['callback_url'],
            "merchant_refund_order_no" => $data['merchant_refund_order_no'],
        ];
        if (isset($data['bank_code'])) {
            $reData['bank_code'] = $data['bank_code'];
            $reData['account_no'] = $data['account_no'];
            $reData['account_name'] = $data['account_name'];
        }
        return $reData;
    }

    /**
     * lianlian 退款请求返回数据
     * @param $data
     * @return array
     */
    public function assembleLianLianRefundResponse($data)
    {
        return [
            'merchant_id' => $data['merchant_id'] ?? '',
            'refund_order_id' => $data['refund_order_id'] ?? '',
            'merchant_refund_id' => $data['merchant_refund_id'] ?? '',
            'refund_amount' => $data['refund_amount'] ?? '',
            'refund_currency' => $data['refund_currency'] ?? '',
            'refund_status' => $data['refund_status'] ?? '',
            'll_create_time' => $data['ll_create_time'] ?? '',
        ];
    }

    /**
     * 发起请求签名
     * @param array $notify_data
     * @param $sign
     * @return string
     */
    public function requestSign(array $request_data)
    {
        $raw_string = json_encode($request_data, JSON_UNESCAPED_UNICODE);
        $shop_supplier_id = $request_data['shop_supplier_id'] ?? 0;
        $sign_salt = self::checkSignSalt($shop_supplier_id);
        return substr(hash('sha256', $sign_salt . $raw_string), 0, 32);
    }

    /**
     * 请求支付退款单号
     * @return string
     */
    public function generateMerchantRefundOrderNo()
    {
        $prefix = 'RS';
        $datePart = date('YmdHis'); // 获取当前日期
        $randomPart = str_pad(rand(0, 99999999), 8, '0', STR_PAD_LEFT); // 生成一个8位的随机数
        return $prefix . $datePart . $randomPart;
    }
}

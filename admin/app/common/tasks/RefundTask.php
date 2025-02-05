<?php

namespace app\common\tasks;

use app\common\model\order\OrderRefundDestination;
use app\common\service\payment\RefundService;
use think\facade\Log;


/**
 * 退款服务
 * 使用：Task::deliver(new RefundTask($payment_order_id, $refund_money, $order_refund_destination_id));
 */
class RefundTask extends Task
{
    protected $shop_supplier_id;

    protected $payment_order_id;
    protected $refund_money;
    protected $order_refund_destination_id;
    protected $merchant_refund_order_no;

    protected $bank_code;
    protected $account_no;
    protected $account_name;

    public function __construct($shop_supplier_id, $payment_order_id, $refund_money, $order_refund_destination_id, $merchant_refund_order_no, $bank_code = '', $account_no = '', $account_name = '')
    {
        $this->shop_supplier_id = $shop_supplier_id;
        $this->payment_order_id = $payment_order_id;
        $this->refund_money = $refund_money;
        $this->order_refund_destination_id = $order_refund_destination_id;
        $this->merchant_refund_order_no = $merchant_refund_order_no;
        $this->bank_code = $bank_code;
        $this->account_no = $account_no;
        $this->account_name = $account_name;
    }
    
    public function start()
    {
        $refundService = (new RefundService($this->shop_supplier_id));
        $result = $refundService->refund($this->payment_order_id, $this->refund_money, '', $this->merchant_refund_order_no, $this->bank_code, $this->account_no, $this->account_name);
        if (!$result) {
            $err_msg = $refundService->getError() ?: '请求支付服务失败';
            $log = "payment_order_id[{$this->payment_order_id}]请求退款失败:{$err_msg}";
            // 记录 refund_order_id
            OrderRefundDestination::where('id', $this->order_refund_destination_id)->update(['refund_order_id' => '']);
            Log::write($log);
            return false;
        } else {
            // 记录 refund_order_id
            OrderRefundDestination::where('id', $this->order_refund_destination_id)->update(['refund_order_id' => $result['refund_order_id']]);
        }

    }

    public function end()
    {

    }
}

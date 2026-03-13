<?php

namespace app\common\model\shop;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\shop\UserShiftLog;
use help\HttpHelp;

/**
 * 用户交班快照
 */
class UserShiftSnapshot extends BaseModel
{

    protected $name = 'staff_shift_snapshot';
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;
    protected $deleteTime = 'delete_time';

    /**
     * 获取快照
     */
    public function getSnapshot($log)
    {
        $detail = $this->where('shift_log_uuid', $log['uuid'])->find();
        if ($detail) {
            $content = $detail['content'];
            $content = json_decode($content, true);
            //
            return $content;
        } else {
            return $this->createSnapshot($log['id']);
        }
    }

    /**
     * 创建快照
     */
    public function createSnapshot($shiftLogId)
    {
        $model = new UserShiftLog;
        $detail = $model->detail($shiftLogId);
        if (is_object($detail)) {
            $detail = $detail->toArray();
        }

        // 请求营业数据
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/business', [
            'duty_no' => $detail['shift_no'],
            'query_start_time' => $detail['start_time'],
            'query_end_time' => $detail['end_time'],
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json',
            'Client-Version: 199.99.99',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $this->error = $res['message'] ?? '请求失败';
            return false;
        }
        $businessData = $res['data'];

        // 请求班次退款金额
        $res = HttpHelp::getRequest('http://nginx/api/v1/shop/statistics/shift_refund_amount', [
            'staff_uuid' => $detail['staff_uuid'],
            'query_start_time' => $detail['start_time'],
            'query_end_time' => $detail['end_time'],
        ], [
            'Authorization: Bearer ' . request()->header('token'),
            'Accept-Language: ' . request()->header('language'),
            'Content-Type: application/json',
        ]);
        if (!$res) {
            $this->error = '请求失败';
            return false;
        }
        $res = json_decode($res, true);
        if (($res['code'] ?? -1) != 0) {
            $this->error = $res['message'] ?? '请求失败';
            return false;
        }
        $refundData = $res['data'];

        $detail['order'] = [
            'received_price' => $businessData['total_received_price'], // 实收金额
            'service_money' => $businessData['total_service_money'], // 服务费
            'pay_fee_money' => $businessData['total_pay_fee_money'], // 支付手续费
            'consumption_tax_money' => $businessData['total_tax_money'], // 税费
            'product_num' => $businessData['total_product_num'], // 商品数量
            'discount_money' => $businessData['total_discount_money'], // 优惠折扣
            'user_discount_money' => $businessData['total_user_discount_money'], // 会员折扣
            'total_give_product_price' => $businessData['total_give_product_price'] ?? 0, // 赠送商品金额
            'refund_money' => $refundData['refund_amount'], // 退款
            'recharge_amount' => $businessData['member_data']['recharge_amount'], // 充值金额
            'gift_money' => $businessData['member_data']['gift_money'], // 赠送金额
            'gift_points' => $businessData['member_data']['gift_points'], // 赠送积分
            'total_order_num' => $businessData['total_order_num'], // 合计-所有订单数
            'total_table_num' => $businessData['total_table_num'], // 合计-桌数
            'total_people_num' => $businessData['total_people_num'], // 合计-人数
            'min_order_price' => $businessData['min_order_price'], // 合计-最小订单金额
            'max_order_price' => $businessData['max_order_price'], // 合计-最大订单金额
            'avg_order_price' => $businessData['avg_order_price'], // 合计-平均订单金额
            'table_order_num' => $businessData['all_table_order_num'], // 桌台方式-订单数（桌数）
            'table_people_num' => $businessData['all_table_people_num'], // 桌台方式-人数
            'table_min_order_price' => $businessData['all_table_min_order_price'], // 桌台方式-最小订单金额
            'table_max_order_price' => $businessData['all_table_max_order_price'], // 桌台方式-最大订单金额
            'table_avg_order_price' => $businessData['all_table_avg_order_price'], // 桌台方式-平均订单金额
            'table_people_avg' => $businessData['all_table_people_avg'], // 桌台方式-人均
            'cashier_order_num' => $businessData['all_cashier_order_num'], // 点餐方式-订单数
            'cashier_min_order_price' => $businessData['all_cashier_min_order_price'], // 点餐方式-最小订单金额
            'cashier_max_order_price' => $businessData['all_cashier_max_order_price'], // 点餐方式-最大订单金额
            'cashier_avg_order_price' => $businessData['all_cashier_avg_order_price'], // 点餐方式-平均订单金额
            'percentage_list' => [], // 税类
            'incomes' => [], // 支付收入
            'peak_hour_list' => [], // 高峰期
        ];
        foreach ($businessData['percentage_list'] as $percentageItem) {
            $detail['order']['percentage_list'][] = [
                'consumption_tax' => $percentageItem['consumption_tax'], // 税类-消费税
                'tax_rate' => $percentageItem['tax_rate'], // 税类-税率
                'total_price' => $percentageItem['total_price'] // 税类-总计
            ];
        }
        foreach ($businessData['payment_method_incomes'] as $paymentMethodIncome) {
            $detail['order']['incomes'][] = [
                'pay_type_name' => $paymentMethodIncome['name'], // 支付收入-名称
                'price' => $paymentMethodIncome['amount'], // 支付收入-金额
            ];
        }
        foreach ($businessData['peak_hour_list'] as $peakHourItem) {
            $detail['order']['peak_hour_list'][] = [
                'time_period' => $peakHourItem['time_period'], // 高峰期-时间
                'num' => $peakHourItem['num'], // 高峰期-订单数
                'amount' => $peakHourItem['amount'], // 高峰期-订单金额
            ];
        }
        // 异常信息
        $detail['abnormal'] = [
            'refund_product_times' => $businessData['abnormal_data']['refund_product_times'], // 退菜次数
            'refund_times' => $businessData['abnormal_data']['refund_times'], // 退款次数
            'reverse_settle_times' => $businessData['abnormal_data']['reverse_settle_times'], // 反结账次数
            'product_free_times' => $businessData['abnormal_data']['product_free_times'], // 赠菜次数
            'free_order_times' => $businessData['abnormal_data']['free_order_times'], // 免单次数
            'product_move_times' => $businessData['abnormal_data']['product_move_times'], // 转菜次数
            'change_price_times' => $businessData['abnormal_data']['change_price_times'], // 单品改价次数
            'change_order_price_times' => $businessData['abnormal_data']['change_order_price_times'], // 整单改价次数
            'discount_order_times' => $businessData['abnormal_data']['discount_order_times'], // 整单折扣次数
            'round_order_times' => $businessData['abnormal_data']['round_order_times'], // 整单抹零次数
        ];
        // 销售信息
        $salesInfo = [];
        foreach ($businessData['category_list'] as $categoryItem) {
            $salesInfo[] = [
                'name_text' => $categoryItem['name'], // 分类
                'sales' => $categoryItem['sales_num'], // 销售数量
                'prices' => $categoryItem['prices'], // 销售金额
            ];
        }
        $detail['salesInfo'] = $salesInfo;
        //
        $content = json_encode($detail);

        $this->save([
            'shift_log_uuid' => $detail['uuid'],
            'content' => $content,
        ]);

        return $detail;
    }

    /**
     * 计算营业收入
     */
    public function calcBusinessIncome($detail) {
        if (isset($detail['order'])) {
            $balance = 0;
            if (isset($detail['order']['incomes'])) {
                foreach ($detail['order']['incomes'] as $income) {
                    if (isset($income['pay_type']) && $income['pay_type'] == 10) {
                        $balance = helper::bcadd($balance, $income['price'], 2);
                    }
                }
            }
            // 营业收入 = 实收金额-税费 + 现金收入
            return helper::bcadd(helper::bcsub($detail['order']['received_price'], $detail['order']['consumption_tax_money'], 2), $balance, 2);
        }

        return 0;
    }
}

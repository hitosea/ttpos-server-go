<?php

namespace app\common\model\user;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\enum\order\MemberSaleOrderStatusEnum;

/**
 * 会员外送订单
 */
class MemberSaleOrder extends BaseModel
{
    use SoftDelete;
    protected $name = 'member_sale_order';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    public function getList($param)
    {
        $channel = $param['channel'] ?? '';
        $startTime = $param['start_time'] ?? '';
        $endTime = $param['end_time'] ?? '';
 
        return $this->when($channel, function ($q) use ($channel) {
            $q->where('channel', '=', $channel);
        })->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
            $q->whereBetween('create_time', [$startTime, $endTime]);
        })->where('status', '=', MemberSaleOrderStatusEnum::FINISHED)->order(["create_time" => 'desc'])->paginate($param);
    }

    public function getMonthStatistics($param)
    {
        $startTime = $param['start_time'] ?? '';
        $endTime = $param['end_time'] ?? '';

        // 统计订单数、配送费、基础服务费、起步配送费、距离单价
        $list = $this->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
            $q->whereBetween('create_time', [$startTime, $endTime]);
        })->where('status', '=', MemberSaleOrderStatusEnum::FINISHED)->select();

        $data = [
            'order_count' => 0,
            'delivery_fee_amount' => 0,
        ];
        $channelData = [];
        foreach ($list as $item) {
            $data['order_count'] += 1;
            $data['delivery_fee_amount'] += $item['delivery_fee_amount'];
            if (!isset($channelData[$item['related_order_type']])) {
                $channelData[$item['related_order_type']] = 0;
            }
            $channelData[$item['related_order_type']] += $item['delivery_fee_amount'];
        }

        $newChannelData = [];
        foreach ($channelData as $channel => $amount) {
            $newChannelData[] = [
                'channel' => $channel,
                'amount' => $amount,
            ];
        }
        $data['channel_data'] = $newChannelData;

        return $data;
    }
}

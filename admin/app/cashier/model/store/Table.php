<?php


namespace app\cashier\model\store;

use app\common\model_old\order\Order;
use app\common\enum\order\OrderStatusEnum;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model_old\store\Table as TableModel;

/**
 * 桌位模型
 */
class Table extends TableModel
{
    /**
     * 获取列表数据
     * @param $area_id
     * @param $type_id
     * @param $user
     * @param int $status 桌台状态 10-未开台 30-已开台
     * @param int $occ_status 占用状态 1-自助餐 2-非自助餐 3-待清台
     * @param int $is_lock 桌台锁单状态选择 0-未选 1-已选
     */
    public function getList($area_id, $type_id, $user, $status = 0, $occ_status = 0, $is_lock = 0)
    {
        $model = $this;
        $query = $model->with(['underwayOrder'])
            ->when(!empty($area_id), function ($query) use ($area_id) {
                return $query->where('area_id', '=', $area_id);
            })
            ->when(!empty($type_id), function ($query) use ($type_id) {
                return $query->where('type_id', '=', $type_id);
            });

        // 获取总列表
        $total_list = $query->order('sort asc')->select()->toArray();

        // 初始化计数器
        $total_num = count($total_list);
        $available_num = 0;
        $occ_buffet_num = 0;
        $occ_not_buffet_num = 0;
        $occ_wait_num = 0;
        $lock_num = 0;  // 锁单数量

        // 过滤和统计
        $status = (int)$status;
        $list = array_filter($total_list, function ($table) use ($status, &$available_num, &$occ_buffet_num, &$occ_not_buffet_num, &$occ_wait_num, &$lock_num) {
            if ($table['status'] == 10) {
                $available_num++;
            } elseif ($table['status'] == 30) {
                if ($table['underwayOrder']) {
                    if ($table['underwayOrder']['is_buffet'] == 1) {
                        $occ_buffet_num++;
                    } else {
                        $occ_not_buffet_num++;
                    }
                    if ($table['underwayOrder']['is_lock'] == 1) {
                        $lock_num++;
                    }
                } else {
                    $occ_wait_num++;
                }
            }
            if ($status == 0) {
                return true;
            }
            return $table['status'] == $status;
        });

        // 计算每桌的应收金额
        $orderModel = new Order();
        foreach ($list as &$table) {
            $table['pay_price'] = 0;
            $table['table_remark'] = '';
            if ($table['underwayOrder']) {
                /** @var Order $order */
                $order = $orderModel->where('order_id', $table['underwayOrder']['order_id'])->find();
                if ($order) {
                    $table['pay_price'] = $order->getTablePrice();
                    $table['table_remark'] = $order->table_remark;
                }
            }
        }

        // 过滤列表 占用状态 1-自助餐 2-非自助餐 3-待清台
        $list = array_values(array_filter($list, function ($table) use ($status, $occ_status, $is_lock) {
            if ($is_lock) {
                // 过滤待清台
                if ($table['status'] == 30 && $table['underwayOrder'] && $table['underwayOrder']['is_lock']) {
                    return true;
                }
                return false;
            }
            if (!$occ_status) {
                // 过滤待清台
                if ($status != 0 && $table['status'] == 30 && !$table['underwayOrder']) {
                    return false;
                }
                return true;
            }
            $has_order = isset($table['underwayOrder']);
            $is_buffet = $table['underwayOrder']['is_buffet'] ?? null;
            $status = $table['status'] ?? null;
            switch ($occ_status) {
                case 1:
                    return $has_order && $status == 30 && $is_buffet == 1;
                case 2:
                    return $has_order && $status == 30 && $is_buffet == 0;
                case 3:
                    return !$has_order && $status == 30;
                default:
                    return false;
            }
        }));

        return compact('total_num', 'available_num', 'occ_buffet_num', 'occ_not_buffet_num', 'occ_wait_num', 'lock_num', 'list');
    }

    /**
     * 获取空闲桌台列表数据
     */
    public function getEnableList($area_id, $type_id)
    {
        $model = $this;
        $list = $model->when(!empty($area_id), function ($query) use ($area_id) {
            return $query->where('area_id', '=', $area_id);
        })
            ->when(!empty($type_id), function ($query) use ($type_id) {
                return $query->where('type_id', '=', $type_id);
            })
            ->where('status', '=', 10)
            ->order(['sort' => 'asc', 'create_time' => 'desc'])
            ->select();

        return $list;
    }


    private function formatPayEndTime($leftTime)
    {
        if ($leftTime <= 0) {
            return '';
        }

        $str = '';
        //        $day = floor($leftTime / 86400);
        $hour = floor($leftTime / 3600);
        $min = floor(($leftTime - $hour * 3600) / 60);
        $sec = floor($leftTime - $hour * 3600 - $min * 60);
        //        if ($day > 0) $str .= $day . '天';
        if ($hour > 0)
            $str .= $hour . '小时';
        if ($min > 0)
            $str .= $min . '分';
        if ($sec > 0)
            $str .= $sec . '秒';
        return $str;
    }

    // 开台
    public static function open($table_id)
    {
        return self::where('table_id', '=', $table_id)->update(['status' => 30]);
    }

    // 关台
    public static function close($table_id)
    {
        Order::where('order_status', OrderStatusEnum::NORMAL)
            ->where('pay_status', OrderPayStatusEnum::PENDING)
            ->where('table_id', $table_id)
            ->update(['order_status' => OrderStatusEnum::CANCELLED]);
        return self::where('table_id', '=', $table_id)->update(['status' => 10]);
    }
}

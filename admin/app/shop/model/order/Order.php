<?php

namespace app\shop\model\order;

use app\shop\service\order\ExportService;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\order\Order as OrderModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 订单模型
 */
class Order extends OrderModel
{

    /**
     * 主订单字段
     */
    protected static $mainFields = [
        'order_id', 'order_name', 'parent_id', 'merge_parent_id', 'refund_money', 'is_free', 'pay_price', 'order_status',
        'pay_time', 'is_settled', 'cashier_id', 'is_merge', 'table_no', 'call_no', 'order_no', 'order_price',
        'merge_parent_id', 'order_source', 'user_id', 'create_time', 'pay_status', 'is_lock', 'lock_time', 'actual_price',
        'free_pay_price',
    ];

    /**
     * 子订单字段
     */
    protected static $subFields = [
        'order_id', 'order_name', 'parent_id', 'merge_parent_id', 'refund_money', 'is_free', 'pay_price', 'order_status',
        'pay_time', 'is_settled', 'cashier_id', 'is_merge', 'table_no', 'call_no', 'order_no', 'order_price',
        'merge_parent_id', 'order_source', 'user_id', 'create_time', 'pay_status', 'is_lock', 'lock_time', 'actual_price',
        'free_pay_price',
    ];

    /**
     * 订单列表
     */
    public function getList($dataType, $data = null)
    {
        $model = $this;
        // 检索查询条件
        $model = $model->setWhere($model, $data);
        // 获取数据列表
        return $model->field(self::$mainFields)->with([
            'user' => function ($query) {
                $query->field(['user_id', 'nickName']);
            },
            'payType',
            'refundType',
            'invoiceInfo',
            'mergeList'  => function ($query) {
                $query->field(['order_id', 'merge_parent_id', 'table_no', 'call_no']);
            },
            'parentOrder.payType',
            'subOrder' => function ($query) {
                $query->field(self::$subFields)->with([
                    'payType',
                    'user' => function ($query) {
                        $query->field(['user_id', 'nickName']);
                    },
                ])->order('order_id', 'asc');
            }
        ])->order(['create_time' => 'desc'])->where($this->transferDataType($dataType))->where(function ($q) {
            $q->where('is_merge', 0)
                ->whereOr(function ($q) {
                    $q->where('is_merge', 1)->where('pay_status', OrderPayStatusEnum::SUCCESS);
                });
        })->paginate($data);
    }

    /**
     * 获取订单总数
     */
    public function getCount($type, $data)
    {
        $model = $this;
        // 检索查询条件
        $model = $model->setWhere($model, $data);
        // 获取数据列表
        return $model->alias('order')
            ->where($this->transferDataType($type))->where(function ($q) {
                $q->where('is_merge', 0)
                    ->whereOr(function ($q) {
                        $q->where('is_merge', 1)->where('pay_status', OrderPayStatusEnum::SUCCESS);
                    });
            })
            ->count();
    }

    /**
     * 订单列表(全部)
     */
    public function getListAll($dataType, $query = [])
    {
        $model = $this;
        // 检索查询条件
        $model = $model->setWhere($model, $query);
        // 获取数据列表
        return $model->with([
            'address',
            'user',
            'extract',
            'cashier',
            'payType',
            'product' => function ($query) {
                $query->withTrashed()->where('is_send_kitchen', 1)->with(['image']);
            },
            'buffet' => function ($query) {
                $query->with(['buffetCustomerType']);
            },
        ])
            ->alias('order')
            ->field('order.*')
            ->where($this->transferDataType($dataType))
            ->where(function ($q) {
                $q->where('is_merge', 0)
                    ->whereOr(function ($q) {
                        $q->where('is_merge', 1)->where('pay_status', OrderPayStatusEnum::SUCCESS);
                    });
            })
            ->where('order.is_delete', '=', 0)
            ->limit(2000)
            ->order(['order.create_time' => 'desc'])
            ->select();
    }

    /**
     * 订单导出
     */
    public function exportList($dataType, $query)
    {
        // 获取订单列表
        try {
            $list = $this->getListAll($dataType, $query);
            if (count($list) > 1000) {
                $this->error = '请选择具体时间段，最多可导出1000条以下的数据';
                return false;
            }
            if (($query['request_type'] ?? '') == 1) {
                return true;
            }
        } catch (\Throwable $th) {
            $this->error = '请选择具体时间段，最多可导出1000条以下的数据';
            return false;
        }
        // 导出excel文件
        return (new Exportservice)->orderList($list);
    }

    /**
     * 设置检索查询条件
     */
    private function setWhere($model, $data)
    {
        // 时间类型 0-全都 1-今天 2-昨天 3-周
        if (isset($data['parent_id'])) {
            $model = $model->where('parent_id', '=', $data['parent_id']);
        }
        $startTime = 0;
        $endTime = 0;
        if (isset($data['time_type']) && $data['time_type']) {
            switch ($data['time_type'] ?? 1) {
                case '1': //今天
                    $startTime = strtotime(date('Y-m-d'));
                    $endTime = $startTime + 86399;
                    break;
                case '2': //昨天
                    $startTime = strtotime("-1 days", strtotime(date('Y-m-d')));
                    $endTime = $startTime + 86399;
                    break;
                case '3': //本周
                    $startTime = strtotime('monday this week'); // 本周第一天的时间戳
                    $endTime = strtotime('sunday this week +23 hours +59 minutes +59 seconds'); // 本周最后一天的时间戳
                    break;
            }
        }

        // 收银类型 0-全都 10-桌台 20-收银
        $orderSource = isset($data['order_source']) ? intval($data['order_source']) : 0;
        if ($orderSource) {
            $model = $model->where('order_source', '=', $data['order_source']);
        }
        // 订单号
        if (isset($data['order_no']) && $data['order_no'] != '') {
            $model = $model->like('order_no', trim($data['order_no']));
        }
        // 配送方式(10外卖配送 20上门取30打包带走40店内就餐
        $styleId = isset($data['style_id']) ? intval($data['style_id']) : 0;
        if ($styleId) {
            $model = $model->where('delivery_type', '=', $data['style_id']);
        }
        // 用餐方式0外卖1店内
        if (isset($data['order_type'])) {
            $model = $model->where('order_type', '=', $data['order_type']);
        }
        // 搜索时间段
        if (isset($data['date']) && is_array($data['date']) && isset($data['date'][0]) && isset($data['date'][1])) {
            $model = $model->where('create_time', 'between', [strtotime($data['date'][0]), strtotime($data['date'][1]) + 86399]);
        } else if (isset($data['time_mode']) && is_array($data['time_mode']) && count($data['time_mode']) == 1 && $data['time']) {
            $time_mode = $data['time_mode'][0];
            $timeField = $time_mode == 1 ? 'pay_time' : 'create_time';
            $data[$timeField] = $data['time'] ?? '';
            //
            $timeFields = ['create_time', 'pay_time'];
            foreach ($timeFields as $field) {
                if (isset($data[$field]) && is_array($data[$field])) {
                    $startTime = isset($data[$field][0]) && $data[$field][0] ? strtotime($data[$field][0]) : null;
                    $endTime = isset($data[$field][1]) && $data[$field][1] ? strtotime($data[$field][1]) + 86399 : null;
                    // 开始时间 + 结束时间
                    if ($startTime && $endTime) {
                        $model = $model->where($field, 'between', [$startTime, $endTime]);
                    } elseif ($startTime) {
                        // 只有开始时间
                        $model = $model->where($field, '>', $startTime);
                    } elseif ($endTime) {
                        // 只有结束时间
                        $model = $model->where($field, '<', $endTime)->where($field, '>', 0);
                    }
                }
            }
        } else if (isset($data['time_mode']) && is_array($data['time_mode']) && count($data['time_mode']) == 2 && $data['time']) {
            $startTime = isset($data['time'][0]) && $data['time'][0] ? strtotime($data['time'][0]) : null;
            $endTime = isset($data['time'][1]) && $data['time'][1] ? strtotime($data['time'][1]) + 86399 : null;
            // 开始时间 + 结束时间
            if ($startTime && $endTime) {
                $model = $model->where(function ($query) use ($startTime, $endTime) {
                    $query->where('create_time', 'between', [$startTime, $endTime]);
                    $query->whereOr('pay_time', 'between', [$startTime, $endTime]);
                });
            } else if ($startTime) {
                // 只有开始时间
                $model = $model->where(function ($query) use ($startTime, $endTime) {
                    $query->where('create_time', '>', $startTime);
                    $query->whereOr('pay_time', '>', $startTime);
                });
            } else if ($endTime) {
                // 只有结束时间
                $model = $model->where(function ($query) use ($endTime) {
                    $query->where(function ($subQuery) use ($endTime) {
                        $subQuery->where('create_time', '<', $endTime)
                            ->where('create_time', '>', 0);
                    });
                    $query->whereOr(function ($subQuery) use ($endTime) {
                        $subQuery->where('pay_time', '<', $endTime)
                            ->where('pay_time', '>', 0);
                    });
                });
            }
        } else if ($startTime && $endTime) {
            // 没有时间范围才按 time_type 查询
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }
        // 已送厨
        return $model;
    }

    /**
     * 转义数据类型条件
     */
    private function transferDataType($dataType)
    {
        $filter = [];
        // 订单数据类型
        switch ($dataType) {
            case 'all':
                $filter[] = ['extra_times', '>', 0];
                break;
            case 'payment';
                // $filter[] = ['pay_status', '=', OrderPayStatusEnum::PENDING];
                $filter[] = ['order_status', '=', 10];
                $filter[] = ['extra_times', '>', 0];
                break;
            case 'process';
                $filter[] = ['pay_status', '=', OrderPayStatusEnum::SUCCESS];
                $filter[] = ['order_status', '=', 10];
                $filter[] = ['extra_times', '>', 0];
                break;
            case 'complete';
                $filter[] = ['pay_status', '=', OrderPayStatusEnum::SUCCESS];
                $filter[] = ['order_status', '=', 30];
                $filter[] = ['extra_times', '>', 0];
                break;
            case 'cancel';
                $filter[] = ['order_status', '=', 20];
                $filter[] = ['extra_times', '>', 0];
                break;
        }
        return $filter;
    }

    /**
     * 发送订单
     * @return bool
     */
    public function sendOrder($order_id)
    {
        $deliver = SettingModel::getSupplierItem('deliver', $this['supplier']['shop_supplier_id']);
        if ($this['order_status']['value'] != 10 || $this['deliver_status'] != 0) {
            $this->error = '订单已发送或已完成';
            return false;
        }
        // 开启事务
        $this->startTrans();
        try {
            $this->addOrder($deliver);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 获取待处理订单
     */
    public function getReviewOrderTotal($shop_supplier_id = 0)
    {
        $model = $this;
        $filter['pay_status'] = OrderPayStatusEnum::SUCCESS;
        $filter['delivery_status'] = 10;
        $filter['order_status'] = 10;
        return $model->where($filter)->count();
    }

    /**
     * 获取某天的总销售额
     * 结束时间不传则查一天
     */
    public function getOrderTotalPrice($startDate, $endDate, $shop_supplier_id = 0)
    {
        $model = $this;
        $startDate && $model = $model->where('pay_time', '>=', strtotime($startDate));
        if (is_null($endDate) && $startDate) {
            $model = $model->where('pay_time', '<', strtotime($startDate) + 86400);
        } else if ($endDate) {
            $model = $model->where('pay_time', '<', strtotime($endDate) + 86400);
        }
        return $model->where('pay_status', '=', 20)
            ->where('order_status', '<>', 20)
            ->where('is_merge', '=', 0)
            ->value('sum(pay_price - refund_money)');
    }

    /**
     * 获取某天的客单价
     * 结束时间不传则查一天
     */
    public function getOrderPerPrice($startDate, $endDate = null)
    {
        $model = $this;
        $model = $model->where('pay_time', '>=', strtotime($startDate));
        if (is_null($endDate)) {
            $model = $model->where('pay_time', '<', strtotime($startDate) + 86400);
        } else {
            $model = $model->where('pay_time', '<', strtotime($endDate) + 86400);
        }
        return $model->where('pay_status', '=', 20)
            ->where('order_status', '<>', 20)
            ->avg('pay_price');
    }

    /**
     * 获取某天的下单用户数
     */
    public function getPayOrderUserTotal($day, $shop_supplier_id = 0)
    {
        $model = $this;
        $startTime = strtotime($day);
        $userIds = $model->distinct(true)
            ->where('pay_time', '>=', $startTime)
            ->where('pay_time', '<', $startTime + 86400)
            ->where('pay_status', '=', 20)
            ->column('user_id');
        return count($userIds);
    }

    /**
     * 获取平台的总销售额
     */
    public function getTotalMoney($type, $is_settled = -1)
    {
        $model = $this;
        $model = $model->where('pay_status', '=', 20)->where('order_status', '<>', 20);
        if ($is_settled == 0) {
            $model = $model->where('is_settled', '=', 0);
        }
        if ($type == 'all') {
            return $model->sum('pay_price');
        } else if ($type == 'supplier') {
            return ($model->sum('pay_price')) - ($model->sum('refund_money'));
        } else if ($type == 'sys') {
            return $model->sum('sys_money');
        }
        return 0;
    }
}

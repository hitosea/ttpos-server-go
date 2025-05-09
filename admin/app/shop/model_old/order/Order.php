<?php

namespace app\shop\model_old\order;

use app\shop\service\order\ExportServiceOld;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\enum\settings\DeliveryTypeEnum;
use app\common\model_old\order\OrderProductFree;
use app\common\model_old\order\OrderOperationLog;
use app\common\model_old\order\Order as OrderModel;
use app\common\model_old\settings\Setting as SettingModel;

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
     * 获取订单
     */
    public static function getLists($data)
    {
        // 订单列表
        $model = new self([], request()->appId);
        // 时间模式
        if (!isset($data['time_mode']) || !is_array($data['time_mode'])) {
            $data['time_mode'] = [0]; // 默认开台时间
        }
        //
        $result = $model->getList($data['dataType'] ?? 'all', $data);
        foreach ($result as $key => $item) {
            // 是否显示退款按钮 1-显示 0-隐藏
            /** @var OrderModel $item */
            [$result[$key]['is_refund_button'], $result[$key]['is_cancel_button']] = $item->getButtonStatus($item);
            if ($item['subOrder']) {
                foreach ($item['subOrder'] as $subKey => $subItem) {
                    /** @var OrderModel $subItem */
                    [$result[$key]['subOrder'][$subKey]['is_refund_button'], $result[$key]['subOrder'][$subKey]['is_cancel_button']] = $subItem->getButtonStatus($subItem);
                }
            }
            // 拆单主单支付方式去重
            if ($item['parent_id'] == 0 && count($item['subOrder']) > 0) {
                $payTypes = $item['payType']->toArray();
                $uniquePayTypes = [];
                foreach ($payTypes as $payType) {
                    $uniquePayTypes[$payType['value']] = $payType;
                }
                $item['payType'] = new \think\Collection(array_values($uniquePayTypes));
            }
        }
        $result = $result->toArray();
        // 
        $list = [];
        foreach ($result['data'] as $key => $item) {
            $consumer_uuids = [];
            if ($item['user']['user_id'] ?? '') {
                $consumer_uuids[] = $item['user']['user_id'];
            }
            // 
            $subOrder = [];
            foreach ($item['subOrder'] ?? [] as $key => $subItem) {
                if (isset($subItem['payType']) && $subItem['payType'] instanceof \think\Collection) {
                    $subItemPayType = $subItem['payType']->toArray();
                } else {
                    $subItemPayType = $subItem['payType'];
                } 
                if ($subItem['user']['user_id'] ?? '') {
                    $consumer_uuids[] = $subItem['user']['user_id'];
                }
                // 
                $status = $subItem['order_status']['value'] == 10 ? 0 : ($subItem['order_status']['value'] == 30 ? 1 : 2);
                // 
                $subOrder[] = [
                    'bill_type' => $subItem['order_source'] == 10 ? 0 : 1,
                    'consumer_uuids' => ($subItem['user']['user_id'] ?? '') . '',
                    'extra' => [
                        'is_cell_cancel' => $subItem['is_cancel_button'] ? true : false,
                        'is_cell_delete' => $status == 2,
                        'is_cell_invoice' => $status == 1,
                        'is_cell_print' => $status == 1,
                        'is_cell_refund' => $subItem['is_refund_button'] ? true : false,
                        'is_cell_reverse_settle' => false,
                    ],
                    'finish_time' => $subItem['pay_time_text'],
                    'is_split' => false,
                    'order_amount' => $subItem['order_price'],
                    'order_no' => $subItem['order_no'],
                    'pay_type_name' => implode(',', array_column($subItemPayType, 'name')),
                    'payment_amount' => $subItem['actual_receive_price'],
                    'sale_bill_uuid' => $subItem['order_id'],
                    'sale_order_uuid' => $subItem['order_id'],
                    'serial_no' => ($subItem['table_no'] ?: $subItem['call_no']) . '-' . ($key + 1),
                    'status' => $status,
                ];
            }
            //
            if (isset($item['payType']) && $item['payType'] instanceof \think\Collection) {
                $payType = $item['payType']->toArray();
            } else {
                $payType = $item['payType'];
            } 
            // 
            $status = $item['order_status']['value'] == 10 ? 0 : ($item['order_status']['value'] == 30 ? 1 : 2);
            $list[] = [
                'bill_type' => $item['order_source'] == 10 ? 0 : 1,
                'consumer_uuids' => implode(',', array_unique($consumer_uuids)),
                'extra' => [
                    'is_cell_cancel' => $item['is_cancel_button'] ? true : false,
                    'is_cell_delete' => $status == 2,
                    'is_cell_invoice' => $status == 1,
                    'is_cell_print' => $status == 1,
                    'is_cell_refund' => $item['is_refund_button'] ? true : false,
                    'is_cell_reverse_settle' => false,
                ],
                'finish_time' => $item['pay_time_text'],
                'is_split' => count($item['subOrder']) > 0,
                'order_amount' => $item['order_price'],
                'order_no' => $item['order_no'],
                'pay_type_name' => isset($item['payType']) && count($item['payType']) > 0 ? implode(',', array_column($payType, 'name')) : '',
                'payment_amount' => $item['actual_receive_price'],
                'sale_bill_uuid' => $item['order_id'],
                'sale_order_uuid' => $item['order_id'],
                'sale_orders' => $subOrder,
                'serial_no' => $item['table_no'] ?: $item['call_no'],
                'status' => $item['order_status']['value'] == 10 ? 0 : ($item['order_status']['value'] == 30 ? 1 : 2),
            ];
        }
        // 
        $meta = [
            'page_no' => $result['current_page'],
            'page_size' => $result['per_page'],
            'total' => $result['total'],
            'total_num' => $model->getCount('all', $data),
            'unpaid_num' => $model->getCount('payment', $data),
            'complete_num' => $model->getCount('complete', $data),
            'cancel_num' => $model->getCount('cancel', $data),
        ];
        // 
        $ex_style = DeliveryTypeEnum::store();
        // 
        return compact('list', 'ex_style', 'meta');
    }

    // 订单详情
    public static function details($order_id)
    {
        $model = new self([], request()->appId);
        /** @var OrderModel $detail */
        $detail = $model->detailWithTrashed($order_id, null, ["'' as free_tag_text"]);
        if (isset($detail['pay_time']) && $detail['pay_time'] > 0) {
            $detail['pay_time'] = date('Y-m-d H:i:s', $detail['pay_time']);
        }
        if (isset($detail['delivery_time']) && $detail['delivery_time'] != '') {
            $detail['delivery_time'] = date('Y-m-d H:i:s', $detail['delivery_time']);
        }
        // 
        $extra = [
            'is_cell_cancel' => false,
            'is_cell_delete' => $detail['order_status']['value'] == 20,
            'is_cell_invoice' => $detail['order_status']['value'] == 10,
            'is_cell_print' => $detail['order_status']['value'] == 10,
            'is_cell_refund' => false,
            'is_cell_reverse_settle' => false,
        ];
        //
        if ($detail) {
            $orderProductIds = array_column($detail['product']->toArray(), 'order_product_id');
            $orderProductFrees = OrderProductFree::where('order_product_id', 'in', $orderProductIds)->select()->toArray();
            foreach ($detail['product'] as &$orderProduct) {
                $orderProduct->free_tag_text = $orderProduct->getFreeTagText($orderProductFrees);
            }
            // 是否显示退款按钮 1-显示 0-隐藏
            $buttonStatus = $detail->getButtonStatus($detail);
            $extra['is_cell_refund'] = $buttonStatus[0] ? true : false;
            $extra['is_cell_cancel'] = $buttonStatus[1] ? true : false;
            // 拆单主单支付方式去重
            if ($detail['parent_id'] == 0 && count($detail['subOrder']) > 0) {
                $payTypes = $detail['payType']->toArray();
                $uniquePayTypes = [];
                foreach ($payTypes as $payType) {
                    $uniquePayTypes[$payType['value']] = $payType;
                }
                $detail['payType'] = new \think\Collection(array_values($uniquePayTypes));
            }
        }

        // 
        $member_names = [];
        $member_uuids = [];
        if ($detail['user']['user_id'] ?? '') {
            $member_uuids[] = $detail['user']['user_id'];
            $member_names[] = $detail['user']['nickName'];
        }
        // 
        $subOrder = [];
        if (count($detail['subOrder']) == 0) {
            $detail['subOrder'][] = $detail;
        }
        foreach ($detail['subOrder'] ?? [] as $key => $subItem) {
            if (isset($subItem['payType']) && $subItem['payType'] instanceof \think\Collection) {
                $subItemPayType = $subItem['payType']->toArray();
            } else {
                $subItemPayType = $subItem['payType'];
            } 
            if ($subItem['user']['user_id'] ?? '') {
                $member_uuids[] = $subItem['user']['user_id'];
                $member_names[] = $subItem['user']['nickName'];
            }
            // 
            $buffetNames = [];
            foreach ($subItem['buffet'] as &$buffet) {
                $buffetNames[] = extractLanguage($buffet['name']);
            }
            $payTypes = [];
            foreach ($subItem['payType']->toArray() as &$payType) {
                $payTypes[] = [
                    'code' => $payType['value'],
                    'currency_unit' => '',
                    'payment_amount' => $payType['price'],
                    'payment_type_name' => $payType['name'],
                    'status' => $payType['pay_status'],
                    'status_reason' => '',
                    'source' => $payType['source'],
                    'source_text' => $payType['source_text'],
                ];
            }
            // 
            $status = $subItem['order_status']['value'] == 10 ? 0 : ($subItem['order_status']['value'] == 30 ? 1 : 2);
            // 
            $products = [];
            //
            foreach ($subItem['buffetCustomerType'] as $product) {
                $names = json_decode($product->getData('buffet_name') ?? '', JSON_UNESCAPED_UNICODE);
                $products[] = [
                    'uuid' => $product['buffet_id'],
                    'locale_name' => [
                        'zh' => $names['zh'] ?? '',
                        'th' => $names['th'] ?? '',
                        'en' => $names['en'] ?? '',
                        'zhtw' => $names['zhtw'] ?? '',
                        'ja' => $names['ja'] ?? '',
                        'ko' => $names['ko'] ?? '',
                        'my' => $names['my'] ?? '',
                        'tr' => $names['tr'] ?? '',
                    ],
                    'locale_attribute_name' => (object)[],
                    'price' => floatval($product['price']),
                    'num' => $product['num'],
                    'sale_price' => floatval($product['total_consumption_tax_order_price']),
                    'total_price' => floatval($product['total_consumption_tax_pay_price']),
                    'refund_amount' => floatval($product['refund_money']),
                    'status' => 0,
                    'remark' => "",
                    'is_gift' => false,
                    'is_buffet' => false,
                    'is_buffet_customer' => true,
                    'is_delay' => false,
                    'is_must' => false,
                    'gift_reason' => '',
                    'image_url' => '',
                    'refund_reason' => '',
                ];
            }
            // 
            foreach ($subItem['delay'] ?? [] as $product) {
                $products[] = [
                    'uuid' => $product['id'],
                    'locale_name' => [
                        'zh' => $product['name'] ?? '',
                        'th' => $product['name'] ?? '',
                        'en' => $product['name'] ?? '',
                        'zhtw' => $product['name'] ?? '',
                        'ja' => $product['name'] ?? '',
                        'ko' => $product['name'] ?? '',
                        'my' => $product['name'] ?? '',
                        'tr' => $product['name'] ?? '',
                    ],
                    'locale_attribute_name' => (object)[],
                    'price' => floatval($product['price']),
                    'num' => $product['num'],
                    'sale_price' => floatval($product['total_product_price']),
                    'total_price' => floatval($product['total_price']),
                    'refund_amount' => floatval($product['refund_money']),
                    'status' => 0,
                    'remark' => "",
                    'is_gift' => false,
                    'is_buffet' => false,
                    'is_buffet_customer' => false,
                    'is_delay' => true,
                    'is_must' => false,
                    'gift_reason' => '',
                    'image_url' => '',
                    'refund_reason' => '',
                ];
            }
            // 
            foreach ($subItem['product'] as $product) {
                $names = json_decode($product->getData('product_name') ?? '', JSON_UNESCAPED_UNICODE);
                $attributes = json_decode($product->getData('product_attr') ?? '', JSON_UNESCAPED_UNICODE);
                $products[] = [
                    'uuid' => $product['product_id'],
                    'locale_name' => [
                        'zh' => $names['zh'] ?? '',
                        'th' => $names['th'] ?? '',
                        'en' => $names['en'] ?? '',
                        'zhtw' => $names['zhtw'] ?? '',
                        'ja' => $names['ja'] ?? '',
                        'ko' => $names['ko'] ?? '',
                        'my' => $names['my'] ?? '',
                        'tr' => $names['tr'] ?? '',
                    ],
                    'locale_attribute_name' => empty($attributes) ? (object)[] : [
                        'zh' => $attributes['zh'] ?? '',
                        'th' => $attributes['th'] ?? '',
                        'en' => $attributes['en'] ?? '',
                        'zhtw' => $attributes['zhtw'] ?? '',
                        'ja' => $attributes['ja'] ?? '',
                        'ko' => $attributes['ko'] ?? '',
                        'my' => $attributes['my'] ?? '',
                        'tr' => $attributes['tr'] ?? '',
                    ],
                    'price' => floatval($product['product_price']),
                    'num' => $product['total_num'],
                    'sale_price' => floatval($product['total_consumption_tax_order_price']),
                    'total_price' => floatval($product['total_consumption_tax_pay_price']),
                    'refund_amount' => floatval($product['refund_money']),
                    'status' => $product['is_return'] ? 1 : 0,
                    'remark' => $product['free_remark'],
                    'is_gift' => $product['is_free'] ? true : false,
                    'is_buffet' => $product['is_buffet_product'] ? true : false,
                    'is_buffet_customer' => false,
                    'is_delay' => false,
                    'is_must' => false,
                    'gift_reason' => $product['free_remark'],
                    'image_url' => $product['image']['file_path'] ?? '',
                    'refund_reason' => $product['productReturn']['reason'] ?? '',
                ];
            }
            // 
            $subOrder[] = [
                'bill_type' => $subItem['order_source'] == 10 ? 0 : 1,
                'dining_method' => $subItem['order_type'],
                'finish_time' => $subItem['pay_time'],
                'free_reason' => [
                    'zh' => $subItem['free_remark'],
                    'th' => $subItem['free_remark'],
                    'en' => $subItem['free_remark'],
                    'zhtw' => $subItem['free_remark'],
                    'ja' => $subItem['free_remark'],
                    'ko' => $subItem['free_remark'],
                    'my' => $subItem['free_remark'],
                    'tr' => $subItem['free_remark'],
                ],
                'is_free' => $subItem['is_free'] == 1 ? true : false,
                'member_uuid' => $subItem['user']['user_id'] ?? 0,
                'member_name' => $subItem['user']['nickName'] ?? '',
                'order_amount' => floatval($subItem['order_price']),
                'order_no' => $subItem['order_no'],
                'pay_type_name' => implode(',', array_column($subItemPayType, 'name')),
                'payment_amount' => floatval($subItem['actual_receive_price']),
                'products' => $products,
                'refund_amount' => floatval($subItem['refund_money']),
                'sale_order_uuid' => $subItem['order_id'],
                'serial_no' => ($subItem['table_no'] ?: $subItem['call_no']) . '-' . ($key + 1),
                'status' => $status,
            ];
        }

        // 
        $buffetNames = [];
        foreach ($detail['buffet'] as &$buffet) {
            $buffetNames[] = extractLanguage($buffet['name']);
        }
        $payTypes = [];
        foreach ($detail['payType']->toArray() as &$payType) {
            $payTypes[] = [
                'code' => $payType['value'],
                'currency_unit' => '',
                'payment_amount' => $payType['price'],
                'payment_type_name' => $payType['name'],
                'status' => $payType['pay_status'],
                'status_reason' => '',
                'source' => $payType['source'],
                'source_text' => $payType['source_text'],
                'sale_bill_uuid' => $detail['order_id'],
            ];
        }

        // 操作日志
        $operationOrderId = $detail['parent_id'] > 0 ? $detail['parent_id'] : $order_id;
        $operationLog = OrderOperationLog::getLogList($operationOrderId);
        $operation_log = ['list'=>[]];
        foreach ($operationLog as $log) {
            $operation_log['list'][] = [
                "uuid" => 0,
                "user_name" => $log['user_name'],
                "user_email" => $log['user_email'],
                "source" => $log['source'],
                "create_time" => $log['create_time'],
                "description" => $log['description'],
                // "pay_type" => $log['pay_type'],
                "pay_type" => [],
                "refund_type" => (int)$log['refund_type']
            ];
        }

        // 格式转换
        $detail = [
            'bill_type' => $detail['order_source'] == 10 ? 0 : 1,
            'buffet_names' => implode(',', $buffetNames),
            'cancel_reason' => $detail['cancel_remark'],
            'cashier_name' => $detail['cashier']['real_name'] ??  $detail['cashier']['user_name'] ?? '-',
            'dining_method' => $detail['order_type'],
            'finish_time' => $detail['pay_time'],
            'create_time' => $detail['create_time'],
            'is_buffet' => $detail['is_buffet'] == 1 ? true : false,
            'is_split' => count($detail['subOrder']) > 0 ? true : false,
            'member_names' => implode(',', $member_names),
            'member_uuids' => implode(',', $member_uuids),
            'order_amount' => $detail['order_price'],
            'order_no' => $detail['order_no'],
            'pay_types' => $payTypes,
            'payment_amount' => $detail['actual_receive_price'],
            'refund_amount' => $detail['refund_money'],
            'remark' => $detail['free_remark'] ?: $detail['table_remark'],
            'sale_bill_uuid' => $detail['order_id'],
            'sale_orders' => $subOrder,
            'serial_no' => ($detail['table_no'] ?: $detail['call_no']),
            'status' => $detail['order_status']['value'] == 10 ? 0 : ($detail['order_status']['value'] == 30 ? 1 : 2),
        ];
        //
        return compact('detail', 'extra', 'operation_log');
    }

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
        $query['shop_supplier_id'] = request()->appId;
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
        return (new ExportServiceOld)->orderList($list);
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
        if (isset($data['order_type']) && $data['order_type'] != '-1') {
            $model = $model->where('order_type', '=', $data['order_type']);
        }
        // 店铺ID
        if (isset($data['shop_supplier_id']) && $data['shop_supplier_id']) {
            $model = $model->where('shop_supplier_id', '=', $data['shop_supplier_id']);
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
                    $endTime = isset($data[$field][1]) && $data[$field][1] ? strtotime($data[$field][1]) + (strstr($data['time'][1], ':') ? 0 : 86399) : null;
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
            $endTime = isset($data['time'][1]) && $data['time'][1] ? strtotime($data['time'][1]) + (strstr($data['time'][1], ':') ? 0 : 86399) : null;
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
        if ($shop_supplier_id) {
            $model = $model->where('shop_supplier_id', '=', $shop_supplier_id);
        }
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
        if ($shop_supplier_id) {
            $model = $model->where('shop_supplier_id', '=', $shop_supplier_id);
        }
        return $model->where('pay_status', '=', 20)
            ->where('order_status', '<>', 20)
            ->where('is_delete', '=', 0)
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
            ->where('is_delete', '=', 0)
            ->avg('pay_price');
    }

    /**
     * 获取某天的下单用户数
     */
    public function getPayOrderUserTotal($day, $shop_supplier_id = 0)
    {
        $model = $this;
        $startTime = strtotime($day);
        if ($shop_supplier_id) {
            $model = $model->where('shop_supplier_id', '=', $shop_supplier_id);
        }
        $userIds = $model->distinct(true)
            ->where('pay_time', '>=', $startTime)
            ->where('pay_time', '<', $startTime + 86400)
            ->where('pay_status', '=', 20)
            ->where('is_delete', '=', 0)
            ->column('user_id');
        return count($userIds);
    }

    /**
     * 获取平台的总销售额
     */
    public function getTotalMoney($type, $is_settled = -1)
    {
        $model = $this;
        $model = $model->where('pay_status', '=', 20)
            ->where('order_status', '<>', 20)
            ->where('is_delete', '=', 0);
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

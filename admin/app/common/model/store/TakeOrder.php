<?php


namespace app\common\model\store;

use app\common\enum\order\OrderErrorEnum;
use app\common\library\helper;
use app\common\model\buffet\BuffetProduct;
use app\common\model\order\OrderBuffet;
use app\common\model\product\ProductSku;
use app\common\model\shop\User;
use app\common\model\BaseModel;
use app\common\model\order\Order;
use app\common\model\order\OrderProduct;
use app\common\model\order\OrderOperationLog;

/**
 * 接单模型
 */
class TakeOrder extends BaseModel
{
     /**
     * 追加字段
     * @var string[]
     */
    protected $append = [
        'reject_time_text',
    ];

    /**
     * 接单实例
     * @param $id
     * @return array|mixed
     */
    public static function getRejectTimeTextAttr($value, $data)
    {
        return ($data['reject_time'] ?? '') ? date('Y-m-d h:i:s', $data['reject_time']) : '';
    }

    /**
     * 接单实例
     * @param $id
     * @return array|mixed
     */
    public static function detail($id)
    {
        return self::alias('to')->field(['to.*', 'o.order_id', 'o.table_id', 'o.table_no', 'o.is_buffet'])
            ->leftJoin('order o', "to.order_id = o.order_id")
            ->where('id', $id)
            ->find();
    }

    /**
     * 获取接单列表
     * @param $params
     * @return array
     */
    public function getList($params)
    {
        $area_id = isset($params['area_id']) ? $params['area_id'] : 0;
        $status = isset($params['status']) ? $params['status'] : 0;
        //
        $model =  $this->alias('to')
            ->leftJoin('table t', 'to.table_id = t.table_id')
            ->field(['to.*', 't.table_no'])
            ->when($area_id > 0, function ($q) use ($area_id) {
                $q->where('t.area_id', '=', $area_id);
            })
            ->when($status != -1, function ($q) use ($status) {
                $q->when($status == 0, function ($q) use ($status) {
                    $q->where('to.status', '=', $status);
                });
                $q->when($status == 1, function ($q) use ($status) {
                    $q->where(function ($q) {
                        $q->where('to.status', '=', -1)->whereOr('to.status', '=', 1);
                    });
                });
            });
        //
        $time = time() - 60 * 60 * 24 * 7;
        // 已处理未按最新处理的在最上方, 已处理页面此处只保留7天
        if ($status == 1) {
            $list = $model->where('to.create_time', '>', $time)
                ->order('to.create_time', 'desc')
                ->paginate($params)
                ->toArray();
        } else {
            $list = $model->order('to.create_time', 'asc')
                ->paginate($params)
                ->toArray();
        }
        //
        $area_list = $status == -1 ? [] : TableArea::getSucList();
        $notice = self::where('status', '=', 0)->count();   // 待接单数量
        $done = self::where('create_time', '>', $time)->where(function ($q) {
            $q->where('status', '=', 1)->whereOr('status', '=', -1);
        })->count();
        //
        return [
            'notice' => $notice,
            'done' => $done,
            'list' => $list,
            'area_list' => $area_list,
            'time' => date("Y-m-d H:i:s"),
        ];
    }

    /**
     * 接单详情
     * @return array
     */
    public function getDetail()
    {
        // 本批新增商品
        $newProductList = OrderProduct::field(['product_name', 'total_num', 'total_price', 'batch_no'])
            ->where('order_id', $this->order_id)
            ->where('batch_no', $this->batch_no)
            ->withTrashed()
            ->select()->toArray();
        // 该单已接单商品(只要扫码端来源的)
        $takeProductList = OrderProduct::field(['product_name', 'total_num', 'total_price'])
            ->where('order_id', $this->order_id)
            ->where('is_send_kitchen', '=', 1)
            ->where('add_source', '=', Order::SCAN_PRODUCT_SOURCE)
            ->select()->toArray();
        //
        $new_total_price = array_sum(array_column($newProductList, 'total_price'));
        $take_total_price = array_sum(array_column($takeProductList, 'total_price'));
        // 新增和已送厨商品金额总计
        $this['pay_price'] = helper::bcadd($new_total_price, $take_total_price);
        if ($this->cashier_id == -1) {
            $cashier = [
                'shop_user_id' => 0,
                'user_name' => __('系统自动接单'),
                'real_name' => __('系统自动接单'),
            ];
        } else {
            $cashier = User::field(['shop_user_id', 'user_name', 'real_name'])->where('shop_user_id', $this->cashier_id)->find();
        }

        // 操作日志
        $operationLog = OrderOperationLog::getLogList($this->order_id, $this->batch_no);

        return [
            'detail' => $this,
            'cashier' => $cashier,
            'newProductList' => $newProductList,
            'takeProductList' => $takeProductList,
            'operationLog' => $operationLog,
        ];
    }

    /**
     * @param $order
     * @param $price
     * @param $num
     * @param $status
     * @param $batch_no
     * @param $take_time
     * @param $source
     * @param $is_auto  // 是否自动接单
     * @return void
     */
    public function add($order, $price, $num, $status, $batch_no, $take_time = 0, $source = 'scan', $is_auto = false)
    {
        $cashier_id = request()->cashier_id ?? 0;
        $updateData = [
            'order_id' => $order->order_id,
            'table_id' => $order->table_id,
            'batch_no' => $batch_no,
            'num' => $num,
            'price' => $price,
            'status' => $status,
            'take_time' => $take_time,
            'source' => $source,
            'cashier_id' => $is_auto ? -1 : $cashier_id,
        ];
        $this->save($updateData);
    }

    /**
     * 接单
     * @return bool
     */
    public function accept()
    {
        if ($this->status != 0) {
            $this->error = '当前状态不可操作';
            return false;
        }
        $this->startTrans();
        try {
            // 检查商品价格和后台是否一致
            $orderProductList = OrderProduct::with(['productSku'])->where('batch_no', $this->batch_no)->select();
            $diffProductArr = [];
            if (!empty($orderProductList)) {
                $order = Order::where('order_id', $this->order_id)->find();
                // 订单自助餐商品价格为0
                $buffet_product_ids = [];
                if ($order->is_buffet) {
                    $buffet_ids = (new OrderBuffet)->where('order_id', $order->order_id)->column('buffet_id');
                    $buffet_product_ids = (new BuffetProduct)->whereIn('buffet_id', $buffet_ids)->column('product_id');
                }
                //
                $product_sku_ids = [];
                foreach ($orderProductList as $op_item) {
                    $product_sku_ids[] = $op_item->product_sku_id;
                }
                //
                $productSkuList = ProductSku::whereIn('product_sku_id', $product_sku_ids)->select()->toArray();
                foreach ($productSkuList as $ps_item) {
                    $sku_price_arr[$ps_item['product_sku_id']] = $ps_item['product_price'];
                }
                // 对比并更新
                foreach ($orderProductList as $op_item) {
                    if (in_array($op_item->product_id, $buffet_product_ids)) {
                        $product_price = 0;
                        $sku_price_arr[$op_item->product_sku_id] = 0;
                    } else {
                        $product_price = floatval(helper::bcsub($op_item->product_price, $op_item->feed_price));
                    }
                    if ($op_item->is_change_price == 0 && $product_price != $sku_price_arr[$op_item->product_sku_id]) {
                        $diffProductArr[] = $op_item->product_name_text . "（{$op_item->productSku->spec_name_text}）";
                        $update_price = helper::bcadd($sku_price_arr[$op_item->product_sku_id], $op_item->feed_price);
                        $op_item->where('order_product_id', $op_item->order_product_id)->update(['product_price' => $update_price]);
                    }
                }
            }
            if (!empty($diffProductArr)) {
                (new Order)->reloadPrice($this->order_id);
                $this->commit();
                $this->error ='以下商品价格有变动，请核对后再下单';
                $this->errorData = $diffProductArr;
                $this->errorCode = OrderErrorEnum::DIFF_PRICE_PRODUCT;
                return false;
            }

            // 订单送厨
            $model = new OrderProduct();
            if (!$model->sendKitchen($this->order_id, 'accept', true, 40, OrderProduct::CASHIER_SEND_KITCHEN, Order::CASHIER_PRODUCT_SOURCE, 0, $this->batch_no)) {
                $this->error = $model->getError() ?: '送厨失败';
                $this->errorData = $model->getErrorData();
                $this->errorCode = $model->getErrorCode();
                return false;
            }
            // 更新记录状态
            $updateData = [
                'status' => 1,
                'take_time' => time(),
                'cashier_id' => request()->cashier_id ?? 0,
            ];
            $this->save($updateData);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 拒单
     * @return bool
     */
    public function reject()
    {
        if ($this->status != 0) {
            $this->error = '当前状态不可操作';
            return false;
        }
        $this->startTrans();
        try {
            // 订单商品拒单
            OrderProduct::where('batch_no', $this->batch_no)->update(['is_reject' => 1]);
            // 更新记录状态
            $updateData = [
                'status' => -1,
                'reject_time' => time(),
                'cashier_id' => request()->cashier_id ?? 0,
            ];
            $this->save($updateData);
            OrderOperationLog::createLog($this->order_id, OrderOperationLog::ACTION_ORDER_REJECT, [], $this->batch_no);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 生成下单批次号
     * @return string
     */
    public static function generateBatchNo()
    {
        return time() . sprintf("%06d", mt_rand(0, 999999));
    }

    /**
     * 最新十条待接单消息
     * @return mixed
     */
    public static function getNewTakeOrder()
    {
        return (new TakeOrder)->alias('to')
            ->leftJoin('table t', 'to.table_id = t.table_id')
            ->field(['to.*', 't.table_no'])
            ->where('to.status', 0)
            ->order('create_time', 'desc')
            ->limit(10)
            ->select()
            ->toArray();
    }

    /**
     * 待接单数量
     * @return int
     */
    public static function getTakeOrderNotice()
    {
        return self::where('status', '=', 0)->count();
    }
}

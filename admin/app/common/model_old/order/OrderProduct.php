<?php

namespace app\common\model_old\order;

use help\QueueHelp;
use think\facade\Db;
use think\facade\Env;
use app\common\library\helper;
use app\common\model_old\BaseModel;
use app\common\model_old\store\FreeTag;
use think\model\concern\SoftDelete;
use app\common\enum\http\StatusCode;
use app\common\model_old\store\TakeOrder;
use app\common\exception\BaseException;
use app\common\model_old\supplier\Printing;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderStatusEnum;
use app\common\model_old\buffet\BuffetProduct;
use app\common\model_old\order\OrderProductFree;
use app\common\model_old\order\Order as OrderModel;
use app\common\enum\product\DeductStockTypeEnum;
use app\common\service\order\OrderPrinterService;
use app\common\model_old\product\Product as ProductModel;
use app\common\model_old\settings\Setting as SettingModel;
use app\common\service\product\factory\ProductFactory;
use app\common\model_old\product\ProductSku as ProductSkuModel;
use app\common\model_old\order\OrderProduct as OrderProductModel;

/**
 * 订单商品模型
 */
class OrderProduct extends BaseModel
{
    use SoftDelete;
    protected $name = 'order_product';
    protected $pk = 'order_product_id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    protected $append = [
        'product_name_text',
        'kitchen_status', // 菜品状态 0-制作中 1-已完成
        'consumption_tax_pay_price',
        'total_consumption_tax_pay_price',
        'total_consumption_tax_order_price',
    ];

    // 送厨来源 1-收银 2-平板 3-扫码
    const CASHIER_SEND_KITCHEN = 1;
    const TABLET_SEND_KITCHEN = 2;
    const SCAN_SEND_KITCHEN = 3;
    // 添加来源 1-收银 2-平板 3-扫码
    const CASHIER_ADD_PRODUCT = 1;
    const TABLET_ADD_PRODUCT = 2;
    const SCAN_ADD_PRODUCT = 3;

    protected $printOrder;
    protected $sourceProductList;

    /**
     * 获取打印订单
     */
    public function getPrintOrder()
    {
        return $this->printOrder;
    }

    /**
     * 菜品状态
     */
    public function getKitchenStatusAttr($value, $data)
    {
        try {
            return $data['is_send_kitchen'] == 1 && $data['finish_num'] == $data['total_num'] ? 1 : 0;
        } catch (\Throwable $th) {
            return 0;
        }
    }

    /**
     * 商品属性组
     */
    public function getAttrIdsAttr($value, $data)
    {
        if (isset($data['attr_ids']) && !empty($data['attr_ids'])) {
            return json_decode($data['attr_ids'], true);
        }
        return $value;
    }

    /**
     * 商品加料组
     */
    public function getFeedIdsAttr($value, $data)
    {
        if (isset($data['feed_ids']) && !empty($data['feed_ids'])) {
            return json_decode($data['feed_ids'], true);
        }
        return $value;
    }

    /**
     * 免赠标签
     */
    public function orderProductFree()
    {
        return $this->hasMany(OrderProductFree::class, 'order_product_id', 'order_product_id');
    }

    /**
     * 免单原因
     */
    public function getFreeTagText($orderProductFree)
    {
        $free_tag_arr = [];
        if ($orderProductFree) {
            foreach ($orderProductFree as $item) {
                if ($item['order_product_id'] == $this->order_product_id) {
                    $free_tag_arr[] = extractLanguage($item['free_tag'] ?? '');
                }
            }
        }
        $free_tag_text = implode('、', $free_tag_arr);
        if ($free_tag_text) {
            if (isset($this['free_remark']) && $this['free_remark']) {
                $free_tag_text = $free_tag_text . '、' . $this['free_remark'];
            }
        } else if (isset($this['free_remark']) && $this['free_remark']) {
            $free_tag_text = $this['free_remark'];
        }
        return $free_tag_text;
    }

    /**
     * 获取商品数据 (可指定某天)
     */
    public function getProductNameTextAttr($value, $data)
    {
        return extractLanguage($data['product_name'] ?? '');
    }

    /**
     * 商品价格
     */
    public function getProductPriceAttr($value)
    {
        return floatval($value);
    }

    /**
     * 属性
     */
    public function getProductAttrAttr($value, $isPrintTmp)
    {
        $separator = ';';
        if ($isPrintTmp === true) {
            $lang = checkDetect();
            if ($lang == 'zh' || $lang == 'zhtw') {
                $separator = ' ； ';
            } else if ($lang == 'my') {
                $separator = ' | ';
            } else {
                $separator = ' ; ';
            }
        }
        $value = trim($value, ';'); // 先去除前后分号
        $values = explode("};{", $value);
        foreach ($values as $key => $part) {
            $partStr = '{' . trim($part, '{}') . '}';
            json_decode($partStr, true);
            $values[$key] = (json_last_error() === JSON_ERROR_NONE) ? extractLanguage($partStr) : __($part);
        }
        return implode($separator, array_filter($values)) ?: '';
    }

    /**
     * 商品最终[单价]应付（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getConsumptionTaxPayPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_pay_price'])
            && isset($data['total_num'])
            && isset($data['product_consumption_tax'])
            && isset($data['product_service_fee'])
            && isset($data['product_service_consumption_tax'])
        ) {
            $unit_pay_price = helper::bcdiv($data['total_pay_price'], $data['total_num'] ?: 1);                                     // 商品应付单价
            $unit_product_consumption_tax = helper::bcdiv($data['product_consumption_tax'], $data['total_num']);                                // 商品消费税单价
            $unit_product_service_fee = helper::bcdiv($data['product_service_fee'], $data['total_num']);                                        // 商品服务费单价
            $unit_product_service_consumption_tax = helper::bcdiv($data['product_service_consumption_tax'], $data['total_num']);                // 商品服务费消费税单价
            $product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($unit_pay_price, $unit_product_consumption_tax) : $unit_pay_price;     // 含税单价
            $product_price = helper::bcadd($product_price, helper::bcadd($unit_product_service_fee, $unit_product_service_consumption_tax));    // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
            $consumption_tax_pay_price = $product_price;
            return floatval($consumption_tax_pay_price);
        }
        return 0;
    }

    /**
     * 商品最终[总价]应付（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getTotalConsumptionTaxPayPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_pay_price'])
            && isset($data['total_num'])
            && isset($data['product_consumption_tax'])
            && isset($data['product_service_fee'])
            && isset($data['product_service_consumption_tax'])
        ) {
            $unit_pay_price = helper::bcdiv($data['total_pay_price'], $data['total_num'] ?: 1);                                     // 商品应付单价
            $unit_product_consumption_tax = helper::bcdiv($data['product_consumption_tax'], $data['total_num']);                                // 商品消费税单价
            $unit_product_service_fee = helper::bcdiv($data['product_service_fee'], $data['total_num']);                                        // 商品服务费单价
            $unit_product_service_consumption_tax = helper::bcdiv($data['product_service_consumption_tax'], $data['total_num']);                // 商品服务费消费税单价
            $product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($unit_pay_price, $unit_product_consumption_tax) : $unit_pay_price;     // 含税单价
            $product_price = helper::bcadd($product_price, helper::bcadd($unit_product_service_fee, $unit_product_service_consumption_tax));    // 最终单价 = 含税单价 + 商品服务费单价 + 商品服务费消费税单价
            $consumption_tax_pay_price = helper::bcmul($product_price, $data['total_num']);                                                     // 最终总价
            return floatval($consumption_tax_pay_price);
        }
        return 0;
    }

    /**
     * 商品最终[总价]原价（商品消费税+商品服务费+商品服务费消费税）
     * @param $value
     * @param $data
     * @return float|int
     */
    public function getTotalConsumptionTaxOrderPriceAttr($value, $data = [])
    {
        if (
            isset($data['tax_calc_type'])
            && isset($data['total_product_price'])
            && isset($data['product_original_consumption_tax'])
            && isset($data['product_original_service_fee'])
            && isset($data['product_original_service_consumption_tax'])
        ) {
            $total_product_price = $data['tax_calc_type'] == 2 ? helper::bcadd($data['total_product_price'], $data['product_original_consumption_tax']) : $data['total_product_price'];     // 商品含税价;
            $total_final_price = helper::bcadd($total_product_price, helper::bcadd($data['product_original_service_fee'], $data['product_original_service_consumption_tax']));              // 最终总价
            return floatval($total_final_price);
        }
        return 0;
    }

    /**
     * 订单商品列表
     * @return \think\model\relation\BelongsTo
     */
    public function image()
    {
        return $this->belongsTo('app\\common\\model\\file\\UploadFile', 'image_id', 'file_id');
    }

    /**
     * 关联商品表
     * @return \think\model\relation\BelongsTo
     */
    public function product()
    {
        return $this->belongsTo('app\\common\\model\\product\\Product');
    }

    /**
     * 关联商品sku表
     * @return \think\model\relation\BelongsTo
     */
    public function sku()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductSku', 'spec_sku_id', 'spec_sku_id');
    }

    /**
     * 关联商品ProductSku表
     * @return \think\model\relation\BelongsTo
     */
    public function productSku()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductSku', 'product_sku_id', 'product_sku_id');
    }

    /**
     * 关联订单主表
     * @return \think\model\relation\BelongsTo
     */
    public function orderM()
    {
        if ($this->sub_order_id > 0) {
            return $this->belongsTo('Order', 'sub_order_id', 'order_id');
        }
        return $this->belongsTo('Order', 'order_id', 'order_id');
    }

    /**
     * 关联分销商
     * @return \think\model\relation\BelongsTo
     */
    public function agent()
    {
        return $this->belongsTo('app\\common\\model\\agent\\Apply', 'agent_user_id', 'user_id');
    }

    /**
     * 关联商品退菜表
     * @return \think\model\relation\BelongsTo
     */
    public function productReturn()
    {
        return $this->belongsTo('app\\common\\model\\order\\OrderProductReturn', 'order_product_id', 'order_product_id');
    }

    /**
     * 订单商品详情
     * @param $where
     * @return array|\think\Model|null
     * @throws \think\db\exception\DataNotFoundException
     * @throws \think\db\exception\DbException
     * @throws \think\db\exception\ModelNotFoundException
     */
    public static function detail($where, $with = ['image'])
    {
        return static::with($with)->find($where);
    }

    /**
     * 获取商品数据 (可指定某天)
     */
    public function getProductData($startDate, $endDate, $type, $shop_supplier_id)
    {
        $model = $this;
        if ($shop_supplier_id > 0) {
            $model = $model->where('order.shop_supplier_id', '=', $shop_supplier_id);
        }
        $model = $model->alias('order_product')
            ->join('order order', 'order_product.order_id = order.order_id', 'left');

        $model = $model->where('order.create_time', '>=', strtotime($startDate));
        if (is_null($endDate)) {
            $model = $model->where('order.create_time', '<', strtotime($startDate) + 86400);
        } else {
            $model = $model->where('order.create_time', '<', strtotime($endDate) + 86400);
        }

        if ($type == 'no_pay') {
            // 未支付
            return $model->where('order.pay_status', '=', 10)->sum('order_product.total_num');
        } else if ($type == 'pay') {
            // 已支付
            return $model->where('order.pay_status', '=', 20)->sum('order_product.total_num');
        }
        return 0;
    }

    /**
     * 判断订单未送厨商品是否存在
     */
    public function isExist($data)
    {
        $model = $this;
        $order_product_id = $model->where('is_send_kitchen', '=', 0)
            ->where('order_id', '=', $data['order_id'])
            ->where('product_id', '=', $data['product_id'])
            ->where('product_sku_id', '=', $data['product_sku_id'])
            ->where('product_attr', '=', $data['describe'])
            ->value('order_product_id');
        return $order_product_id;
    }

    /**
     * 判断商品库存
     * @param $product_num
     * @param $product_source
     * @return int
     * @throws \think\db\exception\DbException
     */
    public function getStockState($product_num, $product_source = Order::CASHIER_PRODUCT_SOURCE)
    {
        $orderProductNum = !$this->order_id ? 0 : OrderProductModel::where('order_id', $this->order_id)
            // 下单减库存
            ->when($this->deduct_stock_type == DeductStockTypeEnum::CREATE, function ($q) use ($product_source) {
                $q->where('is_send_kitchen', 0)->where('add_source', $product_source);  // TODO 未送厨商品区分终端来源
            })
            //
            ->where('order_product_id', '<>', $this->order_product_id)
            ->where('product_id', '=', $this->product_id)
            ->where('product_sku_id', '=', $this->product_sku_id)
            ->sum('total_num');
        //
        return (new ProductSkuModel)->where('product_id', '=', $this->product_id)
            ->where('product_sku_id', '=', $this->product_sku_id)
            ->where('stock_num', '>=', $orderProductNum + $product_num - 1)
            ->count();
    }

    /**
     * 判断商品是否下架
     * @param $product_id
     * @return int
     * @throws \think\db\exception\DbException
     */
    public function productState($product_id)
    {
        return (new ProductModel)->where('product_id', '=', $product_id)
            ->where('product_status', '=', 10)
            ->where('is_delete', '=', 0)
            ->count();
    }

    /**
     * 加减更改商品数量
     * @param $param
     * @param $product_source
     * @return bool
     */
    public function sub($param, $product_source = Order::CASHIER_PRODUCT_SOURCE, $source = 'cashier')
    {
        // 设置数据
        $settingData = isset($param['setting_data']) ? $param['setting_data'] : [];

        /** @var Order $orderM */
        $orderM = $this->orderM()->field(['order_id', 'app_id'])->find();
        if (!$orderM) {
            return $this->handleError('订单不存在');
        }

        // 禁止并发操作
        $queue = new QueueHelp('ORDER_ALL_' . $orderM->app_id . '_' . $orderM->order_id);
        $queue->while();

        /** @var Order $order */
        $order = $orderM->where('order_id', '=', $orderM->order_id)->find();
        if (!$order) {
            return $this->handleError('订单不存在', 0, $queue);
        }

        // 检查自助餐商品可添加状态
        if ($order['is_buffet'] == 1 && $order['buffet_expired_time'] != -1 && $order['buffet_expired_time'] < time() && $param['type'] != 'down') {
            // 自助餐设置
            $buffetSetting = $settingData ? $settingData[SettingEnum::BUFFET]['values'] : SettingModel::getSupplierItem(SettingEnum::BUFFET, $order['shop_supplier_id'] ?? 0, $order['app_id'] ?? 0);
            if ($buffetSetting['is_buy_continue'] != 1) {
                return $this->handleError('用餐时间已到，无法继续下单', 0, $queue);
            }
            if ($buffetSetting['is_buy_continue'] == 1 && $this['is_buffet_product'] == 1) {
                return $this->handleError('自助餐时间已到达，自助餐商品不可继续下单', 0, $queue);
            }
        }

        // 当前产品
        $orderProduct = $this;
        // 添加/减少数量
        $param['product_num'] = $param['type'] == 'up' ? $this['total_num'] + 1 : $this['total_num'] - 1;
        // 主订单ID
        $mainOrderId = $order->parent_id ?: $order->order_id;
        // 添加 - 拆单1的ID
        $firstSubOrderId = 0;
        // 是否复制
        $isCopy = false;

        /**
        * 当来源不是收银机的时候，判断是否存在拆单 - 往拆单1加减
        */
        if ($source == 'assistant' || $product_source != Order::CASHIER_PRODUCT_SOURCE) {
            // 取得 拆单1 的数据，没有就是不存在拆单
            if (($firstSubOrder = $order->where('parent_id', $mainOrderId)->order('order_id')->find())) {
                // 验证 拆单1 状态是否可操作
                /** @var Order $firstSubOrder */
                $firstSubOrderError = $firstSubOrder->validateOrderActionableStatus();
                if ($param['type'] == 'up' && $firstSubOrderError) {
                    return $this->handleError($firstSubOrderError, 0, $queue);
                }
                // 给值 拆单1的ID
                $firstSubOrderId = $firstSubOrder->order_id;
                // 判断拆单1 是否存在未送厨商品，存在的话就做加减法，不存在就复制一条到拆单1
                if (!$firstSubOrderError && $this->sub_order_id != $firstSubOrder->order_id) {
                    $firstOrderProduct = $this->where('main_order_product_id', $this->main_order_product_id)
                        ->where('sub_order_id', $firstSubOrder->order_id)
                        // 同价格同属性同规格同加料同备注的情况下，才合并，不同的话就要分开显示
                        ->where('is_free', $this->is_free)
                        ->where('remark', $this->remark)
                        ->where('product_price', $this->product_price)
                        //
                        ->order('is_send_kitchen')
                        ->find();
                    if ($firstOrderProduct && !$orderProduct->is_send_kitchen) {
                        if ($param['type'] == 'up') {
                            $param['product_num'] = $firstOrderProduct->total_num + 1;
                        }
                        if ($param['type'] == 'down') {
                            $param['product_num'] = $firstOrderProduct->total_num - 1;
                        }
                        $orderProduct = $firstOrderProduct;
                    } else if($param['type'] == 'up') {
                        $isCopy = true;
                    }
                }
            }
        }

        /**
         * 添加
        */
        if ($param['type'] == 'up') {
            // 验证订单状态是否可操作
            if ($error = $order->validateOrderActionableStatus()) {
                return $this->handleError($error, 0, $queue);
            }
            // 判断商品是否下架
            $product = ProductModel::with(['sku.material', 'feed'])
                ->where('product_id', $this->product_id)
                ->where('product_status', '=', 10)
                ->where('is_delete', '=', 0)
                ->find()?->append([]);
            if (!$product && $param['type'] != 'down') {
                $this->errorData = ['product_id' => $this->product_id];
                return $this->handleError('商品已下架', StatusCode::PRODUCT_ERROR_NOT_EXIST, $queue);
            }
            // 判断限购
            $feedIds = is_array($orderProduct->feed_ids) ? $orderProduct->feed_ids : json_decode($orderProduct->feed_ids);
            if ($order->validatePurchaseLimit($mainOrderId, $product_source, $product, $orderProduct['product_sku_id'], 1, $order['meal_num'], $feedIds) === false) {
                return $this->handleError($order->getError(), 0, $queue);
            }
        }

        /**
         * 减少数量不能大于商品数量
         */
        if ($param['type'] == 'down') {
            if ($param['product_num'] > $orderProduct['total_num']) {
                return $this->handleError('减少数量不能大于商品数量', 0, $queue);
            }
        }

        //
        if ($orderProduct->is_send_kitchen) {
            return $this->handleError('商品已送厨，不可操作', 0, $queue);
        }

        // 禁止并发操作
        $this->startTrans();
        try {
            if ($param['product_num'] <= 0) {
                $orderProduct->force()->delete();
            } else {
                $orderProduct->save(['total_num' => $param['product_num']]);
            }

            // 转移
            if ($param['type'] == 'up' && $firstSubOrderId && $isCopy) {
                $order->addToSubOrder([
                    'main_order_id' => $mainOrderId,
                    'from_order_id' => $this->sub_order_id,
                    'to_order_id' => $firstSubOrderId,
                    'product_list' => [[
                        'order_product_id' => $this->order_product_id,
                        'num' => 1
                    ]]
                ]);
            }

            //
            (new Order)->reloadPrice($firstSubOrderId ?: $order['order_id']);

            //
            $this->commit();
            $queue->release();
            return true;
        } catch (BaseException $e) {
            $this->rollback();
            return $this->handleError($e->getMessage(), 0, $queue);
        }
    }

    /**
     * 删除未送厨商品
     * @param $order_product_id
     * @return bool
     */
    public function delProduct($order_product_id)
    {
        $this->startTrans();
        try {
            $orderProductIds = is_array($order_product_id) ? $order_product_id : [$order_product_id];
            //
            $models = $this->where('order_product_id', 'in', $orderProductIds)->select();
            //
            $allOrderProducts = [];
            $mainOrderProducts = [];
            $appName = app('http')->getName();
            if ($appName != 'cashier') {
                foreach ($models as $model) {
                    $key = md5($model['main_order_product_id'] . $model['is_free'] . $model['remark'] . $model['product_price']);
                    if (!isset($mainOrderProducts[$key])) {
                        if ($model['main_order_product_id'] == 0) {
                            $allOrderProducts[$model->order_product_id] = $model;
                        } else {
                            $mainOrderProducts[$key] = $key;
                            $mainOrderProducts = $this->where('main_order_product_id', $model['main_order_product_id'])
                                ->where('is_free', $model['is_free'])
                                ->where('remark',  $model['remark'])
                                ->where('product_price',  $model['product_price'])
                                ->select();
                            foreach ($mainOrderProducts as $mainOrderProduct) {
                                $allOrderProducts[$mainOrderProduct->order_product_id] = $mainOrderProduct;
                            }
                        }
                    }
                    if (!isset($allOrderProducts[$mainOrderProduct->order_product_id])) {
                        $allOrderProducts[$model->order_product_id] = $model;
                    }
                }
            } else {
                $allOrderProducts = $models;
            }
            //
            foreach ($allOrderProducts as $model) {
                if (!$model) {
                    $this->error = '记录不存在';
                    return false;
                }
                //
                $order_id = $model['sub_order_id'] ?: $model['order_id'];
                // 检查订单状态
                $detail = OrderModel::detail([
                    ['order_id', '=', $order_id],
                    ['order_status', '=', OrderStatusEnum::NORMAL]
                ]);
                //
                if (!$detail) {
                    $this->rollback();
                    $this->error = '当前订单不可修改';
                    return false;
                }
                //
                if ($error = $detail->validateOrderActionableStatus()) {
                    $this->rollback();
                    $this->error = $error;
                    return false;
                }
                if ($model->is_send_kitchen == 1 && $model->is_return == 0) {
                    $this->rollback();
                    $this->error = '商品已送厨，禁止删除';
                    return false;
                }
                $model->force()->delete();
                //
                (new OrderModel)->reloadPrice($order_id);
            }
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 删除退菜商品
     * @param $order_product_id
     * @return bool
     */
    public function delReturnProduct($order_product_id)
    {
        $this->startTrans();
        try {
            $orderProductIds = is_array($order_product_id) ? $order_product_id : [$order_product_id];
            $models = $this->where('order_product_id', 'in', $orderProductIds)->select();
            foreach ($models as $model) {
                $order_id = $model['order_id'];
                if (!$model) {
                    $this->error = '记录不存在';
                    return false;
                }
                // 检查订单状态
                $detail = OrderModel::detail([
                    ['order_id', '=', $order_id],
                    ['order_status', '=', OrderStatusEnum::NORMAL]
                ]);
                if (!$detail) {
                    $this->rollback();
                    $this->error = '当前订单不可修改';
                    return false;
                }
                if ($detail->is_lock == 1) {
                    $this->rollback();
                    $this->error = '订单已被锁定，请解锁后重新操作';
                    return false;
                }
                $model->force()->delete();
                // 收银台订单副表为空删除主订单
                if (self::where('order_id', '=', $order_id)->count() == 0) {
                    $order = OrderModel::where('order_id', '=', $order_id)->find();
                    if ($order['table_id'] == 0) {
                        $order->force()->delete();
                    } else {
                        (new OrderModel)->reloadPrice($order_id);
                    }
                } else {
                    (new OrderModel)->reloadPrice($order_id);
                }
            }
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 未送厨商品备注
     * @param $order_product_id
     * @param $remark
     * @param $sub_order_id
     * @return false
     * @throws \think\db\exception\DataNotFoundException
     * @throws \think\db\exception\DbException
     * @throws \think\db\exception\ModelNotFoundException
     */
    public function updateKitchenRemark($order_product_id, $remark, $sub_order_id = 0)
    {
        $orderProduct = $this->where('order_product_id', '=', $order_product_id)->find();
        if (empty($orderProduct)) {
            $this->error = '商品不存在';
            return false;
        }
        if ($orderProduct->orderM()->value('is_lock') == 1) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }
        $orderId = $orderProduct->order_id;
        //
        $orderProduct->remark = $remark;
        $orderProduct->save();
        //
        return $sub_order_id > 0 ? $sub_order_id : $orderId;
    }

    /**
     * 收银端列表商品改价
     * @param $order_product_id
     * @param $money
     * @return false|mixed
     */
    public function changePrice($order_product_id, $money, $sub_order_id = 0)
    {
        $this->startTrans();
        try {
            if ($money < 0 || $money > 1000000) {
                $this->error = "价格错误";
                return false;
            }
            $p = OrderProduct::where('order_product_id', '=', $order_product_id)->find();
            if (!$p) {
                $this->error = "商品不存在";
                return false;
            }
            if ($p->orderM()->value('is_lock') == 1) {
                $this->error = '订单已被锁定，请解锁后重新操作';
                return false;
            }
            // 改价
            $p->product_price = $money;
            $p->is_change_price = 1;
            $p->total_price = helper::bcmul($money, $p->total_num);
            if ($p->save()) {
                // 更新
                (new OrderModel)->reloadPrice($p['order_id']);
                // 兼容拆单订单信息
                $splitOrder = $p->orderM()->field(['order_id', 'parent_id', 'order_name'])->find();
                // 添加操作记录
                OrderOperationLog::createLog($p['order_id'], OrderOperationLog::ACTION_CHANGE_PRICE, [
                    'order_product_id' => $p->order_product_id,
                    'product_id' => $p->product_id,
                    'product_name' => $p->product_name,
                    'product_attr' => $p->getData('product_attr'),
                    'total_num' => $p->total_num,
                    'price' => $money,
                    'parent_id' => $splitOrder->parent_id,         // 拆单主单ID
                    'order_name' => $splitOrder->order_name,       // 订单名称
                ], '改价');
                //
                $this->commit();
                return $sub_order_id > 0 ? $sub_order_id : $p['order_id'];
            } else {
                $this->error = "商品不存在";
                return false;
            }
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 平板端提交前端购物车数据下单
     * @param $add_product_arr
     * @param $table_info
     * @param $order_id
     * @param $type
     * @param $is_print
     * @param $delivery
     * @param $send_kitchen_source
     * @param $product_source
     * @param $ignore_must // 忽略必点商品 0-否 1-是
     * @return bool
     */
    public function addAndSendKitchen($add_product_arr, $table_info, $order_id, $type, $is_print, $delivery, $send_kitchen_source, $product_source, $ignore_must = 0)
    {
        $tableId = $table_info['table_id'];
        $device_id = '';
        // 禁止并发操作
        $queue = new QueueHelp("ORDERS_ADD_TO_ORDER::{$tableId}:" . ($tableId == 0 ? $device_id : '-') . '::' . (request()->shopSupplierId ?: request()->appId ?: 0));
        $queue->while();
        //
        $this->startTrans();
        try {
            // 添加商品
            $difProductArr = [];
            // 获取方案所有
            $order = Order::where('order_id', $order_id)->find();
            $schemeMustProductIds = $order ? $order->getSchemeMustProductIds() : [];
            //
            $orderModel = new OrderModel();
            //
            foreach ($add_product_arr as $add_product) {
                $add_product['remark'] = $add_product['remark'] ?? '';
                if ($add_product['remark'] && mb_strlen($add_product['remark']) > 50) {
                    $queue->release();
                    $this->rollback();
                    $this->error = '备注超过50个字符';
                    return false;
                }
                $isDiff = $orderModel->addToTableOrder($add_product, $order_id, $schemeMustProductIds, $product_source);
                if ($isDiff === false) {
                    $queue->release();
                    $this->error = $orderModel->getError();
                    $this->errorData = $orderModel->getErrorData();
                    $this->errorCode = $orderModel->getErrorCode();
                    $this->rollback();
                    return false;
                }
                if ($isDiff) {
                    $difProductArr[] = $isDiff;
                }
            }
            if (count($difProductArr) > 0) {
                $queue->release();
                $this->rollback();
                $this->errorData = $difProductArr;
                $this->errorCode = OrderErrorEnum::TABLET_ORDER_PRICE_CHANGE;
                $this->error = '以下商品价格有变动，请核对后再下单';
                return false;
            }

            /**
             * 判断是否存在拆单 - 判断
             */
            $sub_order_id = Order::where('parent_id', $order_id)->order('order_id')->value('order_id') ?: 0;
            if ($sub_order_id) {
                if ($error = Order::where('order_id', $sub_order_id)->find()?->validateOrderActionableStatus()) {
                    return $this->handleError($error, 0);
                }
            }

            // 送厨
            if ($this->sendKitchen($order_id, $type, $is_print, $delivery, $send_kitchen_source, $product_source, $ignore_must)) {
                $this->commit();
                $queue->release();
                return true;
            } else {
                $queue->release();
                $this->errorData = $this->getErrorData();
                $this->errorCode = $this->getErrorCode();
                $this->error = $this->getError();
                return false;
            }
        } catch (BaseException $e) {
            $queue->release();
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 设置产品列表值
     * @param $sourceProductList
     * @return self
     */
    public function setSourceProductList($sourceProductList)
    {
        $this->sourceProductList = $sourceProductList;
        return $this;
    }

    /**
     * 送厨前验证 - 库存，限购等相关
     * @param OrderModel $order
     * @param $type
     * @param $sourceProductList
     */
    public function sendKitchenBeforeVerify(OrderModel $order, $type, $sourceProductList)
    {
        $productSource = $sourceProductList['productSource'];
        $orderProductList = $sourceProductList['orderProductList'];
        $allProductList = $sourceProductList['allProductList'];

        // 授权信息
        $license = request()->licenses;
        if (!$license) {
            $this->error = '授权无效';
            $this->errorCode = OrderErrorEnum::STOCK_ERROR;
            return false;
        }

        // 检查订单是否被锁定
        if ($type != 'payment' && $order->is_lock == 1) {
            $this->error = '订单已被锁定，请解锁后重新操作';
            return false;
        }

        // 过滤得到未送厨的产品(过滤未接单的商品)
        $unSendKitchenProduct = array_filter($orderProductList, function ($orderProduct) {
            return $orderProduct['is_send_kitchen'] == 0 && empty($orderProduct['batch_no']);
        });

        // 检查待送厨商品状态
        if ($order['is_buffet'] == 1 && $type != 'payment') {
            // 自助餐设置
            $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $order['shop_supplier_id'], $order['app_id']);
            $buffet_remaining_time = Order::getBuffetRemainingTime($order['buffet_expired_time']);
            // 检查非自助餐商品超时
            foreach ($unSendKitchenProduct as $orderProduct) {
                if ($orderProduct['is_buffet_product'] != 1 && $buffet_remaining_time <= 0 && $order['buffet_expired_time'] != -1 && $buffetSetting['is_buy_continue'] != 1) {
                    $this->errorCode = OrderErrorEnum::BUFFET_SEND_TIME_OUT;
                    $this->error = '点餐时间已到，无法继续下单';
                    return false;
                }
            }
            // 检查自助餐商品超时
            foreach ($unSendKitchenProduct as $orderProduct) {
                if ($orderProduct['is_buffet_product'] == 1 && $buffet_remaining_time <= 0 && $order['buffet_expired_time'] != -1) {
                    $this->errorCode = OrderErrorEnum::BUFFET_TIME_OUT;
                    $this->error = '自助餐时间已到达，自助餐商品不可继续下单';
                    return false;
                }
            }
        }

        // 检查限购
        $orderProductTotalNumArray = [];
        foreach ($orderProductList as $orderProduct) {
            // 1.0.9 不受并桌过来的商品数量影响
            if ($orderProduct['merge_from_table_id'] == 0) {
                $orderProductTotalNumArray[$orderProduct['product_id']] = [
                    'product_id' => $orderProduct['product_id'],
                    'product_sku_id' => $orderProduct['product_sku_id'],
                    'total_num' => ($orderProductTotalNumArray[$orderProduct['product_id']]['total_num'] ?? 0) + $orderProduct['total_num']
                ];
            }
        }
        $buffetIds = (new OrderBuffet)->where('order_id', '=', $order->order_id)->column('buffet_id');
        $buffetProductAll = (new BuffetProduct)->where('buffet_id', 'in', $buffetIds)->select()->toArray();
        $buffetProductLimitNums = [];
        foreach ($buffetProductAll as $buffetProduct) {
            // 自助餐限制数量改为累加
            if (isset($buffetProductLimitNums[$buffetProduct['product_id']]) && $buffetProductLimitNums[$buffetProduct['product_id']] != 0) {
                if ($buffetProduct['limit_num'] == 0) {
                    $buffetProductLimitNums[$buffetProduct['product_id']] = $buffetProduct['limit_num'];
                } else {
                    $buffetProductLimitNums[$buffetProduct['product_id']] = $buffetProduct['limit_num'] + $buffetProductLimitNums[$buffetProduct['product_id']];
                }
            } else if (!isset($buffetProductLimitNums[$buffetProduct['product_id']])) {
                $buffetProductLimitNums[$buffetProduct['product_id']] = $buffetProduct['limit_num'];
            }
        }
        $out_limit_num = [];
        foreach ($unSendKitchenProduct as $orderProduct) {
            if ($orderProduct['is_buffet_product'] == 1) {
                $limitNum = $buffetProductLimitNums[$orderProduct['product_id']] * $order['meal_num'];
            } else {
                $limitNum = $allProductList[$orderProduct['product_id']]['limit_num'] ?: 0;
            }
            $totalNum = $orderProductTotalNumArray[$orderProduct['product_id']]['total_num'] ?? 1;
            if ($limitNum && $totalNum > $limitNum) {
                $orderProduct['tablet_product_name_text'] = ProductSkuModel::getNameById($orderProduct['product_sku_id']);
                $out_limit_num[] = $orderProduct;
            }
        }
        if (count($out_limit_num) > 0) {
            $this->error = "以下商品超出限购数量，请在限购数量内下单";
            $this->errorData = $out_limit_num;
            $this->errorCode = OrderErrorEnum::OUT_LIMIT_NUM;
            return false;
        }

        // 检查加料库存
        $outFeedStockProductArr = $order->checkOrderFeedIsFull($productSource, $sourceProductList);
        if (count($outFeedStockProductArr) > 0) {
            $this->error = '商品加料库存不足';
            $this->errorData = $outFeedStockProductArr;
            $this->errorCode = OrderErrorEnum::OUT_SEND_MATERIAL_NUM;
            return false;
        }

        // 检查材料库存
        if (($license['sale'] ?? 0) == 1) {
            $outProductArr = $order->checkOrderProductIsFull($productSource, $sourceProductList);
            if (count($outProductArr) > 0) {
                $this->error = '商品材料库存不足';
                $this->errorData = $outProductArr;
                $this->errorCode = OrderErrorEnum::OUT_SEND_MATERIAL_NUM;
                return false;
            }
        }

        // 结账送厨判断
        if ($order['is_buffet'] == 1 && $type == 'payment') {
            // 自助餐设置
            $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $order['shop_supplier_id'], $order['app_id']);
            $buffetRemainingTime = Order::getBuffetRemainingTime($order['buffet_expired_time']);
            // 检查非自助餐商品超时
            $productList = [];
            $buffetProductList = [];
            foreach ($unSendKitchenProduct as $orderProduct) {
                if ($orderProduct['is_buffet_product'] != 1 && $buffetRemainingTime <= 0 && $order['buffet_expired_time'] != -1 && $buffetSetting['is_buy_continue'] != 1) {
                    $productList[] = $orderProduct;
                } else if ($orderProduct['is_buffet_product'] == 1 && $buffetRemainingTime <= 0 && $order['buffet_expired_time'] != -1) {
                    $buffetProductList[] = $orderProduct;
                }
            }
            $noticeList = array_merge($productList, $buffetProductList);
            if (count($noticeList) > 0) {
                $this->error = $buffetSetting['is_buy_continue'] != 1 ? '用餐时间已到，请先删除未送厨商品' : '自助餐时间已到达，请先删除未送厨商品';
                $this->errorData = $noticeList;
                $this->errorCode = OrderErrorEnum::OUT_LIMIT_TIME;
                return false;
            }
        }

        //
        return true;
    }

    /**
     * 送厨
     * @param $order_id
     * @param $type                 // 触发送厨类型 kitchen-送厨 payment-结算 accept-接单
     * @param $is_print
     * @param $delivery             // 30-打包 40-堂食
     * @param $sendKitchenSource    // 从哪里面送厨 送厨来源 1-收银(助手) 2-平板 3-扫码
     * @param $productSource        // 商品来源 1-收银 2-平板 3-扫码
     * @param $ignore_must          // 忽略必点商品 0-否 1-是
     * @param $batch_no             // 下单批次号
     */
    public function sendKitchen($order_id, $type = 'kitchen', $isPrint = true, $delivery = 40, $sendKitchenSource = 1, $productSource = Order::CASHIER_PRODUCT_SOURCE, $ignore_must = 0, $batch_no = '')
    {
        Db::connect()->execute("SET SESSION sql_mode = ''");
        //
        if ($type != 'payment') {
            $queue = new QueueHelp('ORDER_ALL_' . request()->appId . '_' . $order_id);
            $queue->while();
        }
        /** @var OrderModel $order */
        $order = (new OrderModel)->where('order_id', $order_id)->find();
        if (!$order) {
            $this->error = "订单不存在";
            if ($type != 'payment') {
                $queue->release();
            }
            return false;
        }
        //
        if ($error = $order->validateOrderActionableStatus($type)) {
            $this->error = $error;
            if ($type != 'payment') {
                $queue->release();
            }
            return false;
        }
        // 送厨必点商品检查
        if ($type == 'kitchen' && $ignore_must == 0 && !$order->checkSchemeMustProduct()) {
            if ($productSource == Order::SCAN_PRODUCT_SOURCE) {
                $this->error = '已下单和本次要下单的商品未选择必点商品，确定要继续下单吗？';
            } else if ($sendKitchenSource == OrderProduct::CASHIER_SEND_KITCHEN) {
                $this->error = '已送厨和本次要送厨的商品未选择必点商品，确定要继续送厨吗？';
            } else {
                $this->error = $order->getError();
            }
            $this->errorData = $order->getErrorData();
            $this->errorCode = $order->getErrorCode();
            if ($type != 'payment') {
                $queue->release();
            }
            return false;
        }
        //
        $order_create_time = $order['create_time'];

        // 得到产品源数据
        $sourceProductList = $this->sourceProductList ?: $order->getOrderSourceProductList($productSource, $type, $batch_no);
        $orderProductList = $sourceProductList['orderProductList'];
        $allProductList = $sourceProductList['allProductList'];
        $allProductSkuList = $sourceProductList['allProductSkuList'];

        // 得到未送厨的产品
        $unSendKitchenProduct = array_values(array_filter($orderProductList, function ($orderProduct) {
            return $orderProduct['is_send_kitchen'] == 0;
        }));

        // 判断是否下架
        foreach ($unSendKitchenProduct as $orderProduct) {
            // 判断商品是否下架
            foreach ($allProductList as $product) {
                if ($product['is_delete'] == 1 || $product['product_status']['value'] == 20){
                    if ($orderProduct['product_id'] == $product['product_id']) {
                        $this->error = __('商品') .' '. $product['product_name_text'] . ' ' . __('已下架，请选择其他商品');
                        $this->errorData = ['product_id' => $product['product_id']];
                        $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST;
                        return false;
                    }
                }
            }
            // 判断规格是否下架
            if (!isset($allProductSkuList[$orderProduct['product_sku_id']])) {
                $this->error = __('规格') .' '. $orderProduct['product_name_text'] . '-' . $orderProduct['product_attr'] . ' ' . __('已下架，请选择其他规格');
                $this->errorData = ['product_id' => $orderProduct['product_id'], 'product_sku_id' => $orderProduct['product_sku_id']];
                $this->errorCode = StatusCode::PRODUCT_ERROR_NOT_EXIST_SKU;
                return false;
            }
        }

        // 送厨前验证 - 库存，限购等相关
        if (!$this->sendKitchenBeforeVerify($order, $type, $sourceProductList)) {
            if ($type != 'payment') {
                $queue->release();
            }
            return false;
        }
        //
        $isAutoOrder = false;
        //
        $this->startTrans();
        try {
            // 只检查未送厨的
            $unSendSourceProductList = $sourceProductList;
            $unSendSourceProductList['orderProductList'] = array_values(array_filter($unSendSourceProductList['orderProductList'], function ($orderProduct) {
                return $orderProduct['is_send_kitchen'] == 0 && ($orderProduct['is_return'] ?? 0) == 0 && ($orderProduct['is_reject'] ?? 0) == 0;
            }));
            //
            if ($order->parent_id == 0) {
                OrderModel::where('order_id', $order->order_id)
                    ->whereOr('parent_id', $order->order_id)
                    ->inc('extra_times', 1)
                    ->update();
            } else {
                OrderModel::where('order_id', $order->order_id)
                    ->whereOr('order_id', $order->parent_id)
                    ->inc('extra_times', 1)
                    ->update();
            }
            // 得到送厨时间
            $sendKitchenTime = time();
            $stock_update = true; //
            // 处理未送厨商品
            $is_accept_scan_order = isset(request()->licenses['is_accept_scan_order']) ? request()->licenses['is_accept_scan_order'] : 0;
            if ($sendKitchenSource == OrderProduct::SCAN_ADD_PRODUCT && $is_accept_scan_order == 1) {
                $cashierSetting = SettingModel::getSupplierItem(SettingEnum::CASHIER, $this['shop_supplier_id']);
                /** 1.0.9 扫码下单需要受到商家接单功能影响 **/
                $batch_time = time(); // 下单时间
                $batch_no = TakeOrder::generateBatchNo(); // 扫码端下单批号
                // 只处理未接单商品
                $unSendKitchenProduct = array_values(array_filter($unSendKitchenProduct, function ($orderProduct) {
                    return $orderProduct['batch_no'] == '';
                }));
                if (empty($unSendKitchenProduct)) {
                    $this->errorCode = OrderErrorEnum::SCAN_EMPTY_CART;
                    $this->error = '下单商品不能为空';
                    if ($type != 'payment') {
                        $queue->release();
                    }
                    return false;
                }
                $num = 0;
                $price = 0;
                foreach ($unSendKitchenProduct as $order_product) {
                    $num += $order_product->total_num;
                    $price += $order_product->total_price;
                }
                /**
                 * is_accept_scan_order 收银端是否开启扫码点餐接单（不开启下单直接送厨） 0-不开启 1-开启
                 * is_auto_order 自动接单是否开启
                 */
                if ($cashierSetting['is_auto_order'] == 1 && $price < $cashierSetting['auto_order_limit']) {
                    $take_time = $batch_time;
                    // 开启
                    $updateData = [
                        'batch_time' => $batch_time,
                        'batch_no' => $batch_no,
                        'is_send_kitchen' => 1,
                        'send_kitchen_time' => $sendKitchenTime,
                        'send_kitchen_source' => $sendKitchenSource,
                        'merge_from_table_id' => 0, // 从本桌送厨的不再标记并桌来源
                    ];
                    // 接单记录
                    (new TakeOrder)->add($order, $price, $num, 1, $batch_no, $take_time, 'scan', true);
                    //
                    $isAutoOrder = true;
                } else {
                    // 未开启
                    $updateData = [
                        'batch_time' => $batch_time,
                        'batch_no' => $batch_no,
                    ];
                    // 接单记录
                    (new TakeOrder)->add($order, $price, $num, 0, $batch_no);
                    //
                    $isAutoOrder = $isPrint = false;
                    $stock_update = false;  // 未送厨不更新库存
                }
            } else {
                if ($type == 'accept') {
                    $updateData = [
                        'is_send_kitchen' => 1,
                        'send_kitchen_time' => $sendKitchenTime,
                        'send_kitchen_source' => $sendKitchenSource,
                        'merge_from_table_id' => 0, // 从本桌送厨的不再标记并桌来源
                    ];
                } else {
                    $updateData = [
                        'batch_time' => 1,
                        'is_send_kitchen' => 1,
                        'send_kitchen_time' => $sendKitchenTime,
                        'send_kitchen_source' => $sendKitchenSource,
                        'merge_from_table_id' => 0, // 从本桌送厨的不再标记并桌来源
                    ];
                }
            }
            // 更新库存
            if ($stock_update) {
                $res = ProductFactory::getFactory($order['order_source'])->updateOrderProductStock($unSendSourceProductList, 'dec');
                if ($res !== true) {
                    $this->errorCode = OrderErrorEnum::OUT_PRODUCT_STOCK;
                    $this->error = "以下商品库存不足，请删除后再下单";
                    $this->errorData = $res;
                    $this->rollback();
                    if ($type != 'payment') {
                        $queue->release();
                    }
                    return false;
                }
            }
            // 更新前打印数据
            $order['product'] = $unSendKitchenProduct;
            $this->printOrder = $order;
            //
            $buffetIds = (new OrderBuffet)->where('order_id', $order_id)->column('buffet_id') ?: [];
            $buffetProductAll = (new BuffetProduct)
                ->distinct(true)
                ->where('buffet_id', 'in', $buffetIds)
                ->where('product_id', 'in', array_column($unSendKitchenProduct, 'product_id'))
                ->select()
                ->toArray();
            $buffetProductArray = [];
            foreach ($buffetProductAll as $buffetProduct) {
                if (!isset($buffetProductArray[$buffetProduct['product_id']]) || $buffetProduct['is_show_kitchen'] < $buffetProductArray[$buffetProduct['product_id']]['is_show_kitchen']) {
                    $buffetProductArray[$buffetProduct['product_id']] = $buffetProduct;
                }
            }
            $showKitchenProductIds = [];
            $notShowKitchenProductIds = [];
            foreach ($unSendKitchenProduct as $orderProduct) {
                if (isset($buffetProductArray[$orderProduct['product_id']])) {
                    $orderProduct['is_show_kitchen'] = $buffetProductArray[$orderProduct['product_id']]['is_show_kitchen'];
                } else {
                    $orderProduct['is_show_kitchen'] = $allProductList[$orderProduct['product_id']]['is_show_kitchen'];
                }
                //
                if ($orderProduct['is_show_kitchen'] == 1) {
                    $showKitchenProductIds[] = $orderProduct->order_product_id;
                } else {
                    $notShowKitchenProductIds[] = $orderProduct->order_product_id;
                }
            }
            self::where('order_product_id', 'in', $showKitchenProductIds)->update($updateData);
            self::where('order_product_id', 'in', $notShowKitchenProductIds)->update(array_merge($updateData, [
                'finish_time' => $sendKitchenTime,
                'finish_num' => Db::raw('total_num')
            ]));

            //
            OrderModel::where('order_id', $order->order_id)->whereOr('parent_id', $order->order_id)->update([
                'delivery_type' => $delivery,   // 打包状态
                'is_must_notice' => 0,          // 下过一次单不在弹出必点页
            ]);
            //
            OrderProductModel::where('order_id', $order_id)->save(['delivery' => $delivery]);
            //
            (new Order())->reloadPrice($order_id);
            // 非结账送厨下单合单重新生成
            if ($order['merge_id'] && $type != 'payment') {
                $param = [
                    'app_id' => $order['app_id'],
                    'shop_supplier_id' => $order['shop_supplier_id'],
                    'settle_device_id' => $order['device_id'],
                    'device_id' => $order['device_id'],
                ];
                $order->generateMasterMergeOrder($order['merge_id'], $param);
            }
            // 添加操作记录
            if ($unSendKitchenProduct) {
                if (!empty($batch_no)) {
                    // 收银触发的送厨需要处理待接订单状态
                    if ($sendKitchenSource == OrderProduct::CASHIER_ADD_PRODUCT) {
                        $update_total_price = 0;
                        foreach ($unSendKitchenProduct as $orderProduct) {
                            $update_total_price += $orderProduct['total_price'];
                        }
                        TakeOrder::where('batch_no', $batch_no)->update(['price' => $update_total_price, 'status' => 1, 'take_time' => time()]);
                    }
                    if ($is_accept_scan_order && ($isAutoOrder || $sendKitchenSource == OrderProduct::CASHIER_ADD_PRODUCT)) {
                        OrderOperationLog::createLog($order_id, OrderOperationLog::ACTION_ORDER_TAKING, ['is_auto_order' => $isAutoOrder], $batch_no);
                    }
                }
                //
                if ($isPrint) {
                    OrderOperationLog::createLog($order_id, OrderOperationLog::ACTION_SEND_KITCHEN, array_map(function ($product) use ($isAutoOrder) {
                        return [
                            'order_product_id' => $product['order_product_id'],
                            'product_id' => $product['product_id'],
                            'product_name' => $product['product_name'],
                            'product_attr' => $product->getData('product_attr'),
                            'total_num' => $product['total_num'],
                            'is_auto_order' => $isAutoOrder,
                        ];
                    }, $unSendKitchenProduct), $batch_no);
                }
            }
            //
            $this->commit();
            if ($type != 'payment') {
                $queue->release();
            }
            //
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            if ($type != 'payment') {
                $queue->release();
            }
            return false;
        }
        // 菜品打印
        if ($isPrint) {
            // 送厨更新取单号
            if ($order->table_id == 0 && $order->call_no == '') {
                $order->call_no = $order->getTableNumber($order_create_time);
                $order->save();
            }
            (new OrderPrinterService)->printProductTicket($this->printOrder, Printing::PRINT_TYPE_KITCHEN);
        }
        //
        if ($type != 'payment') {
            $queue->release();
        }
        return true;
    }

    /**
     * 助手、平板端、扫码 订单送厨商品按送厨时间分组
     * @param $order_id
     * @param $buffet_customer_list
     * @param $delay_list
     * @return array
     */
    public static function getGroupByTime($order_id, $buffet_customer_list = [], $delay_list = [], $order_product_list = [], $is_scan = 0)
    {
        $prefix = Env::get('DB_PREFIX');
        $orderProductList = $order_product_list ?: OrderProduct::alias('op')
            ->leftJoin('product p', 'op.product_id = p.product_id')
            ->leftJoin(
                "(
                    select 
                        tos.order_id, 
                        tos.batch_no, 
                        tos.reject_time, 
                        count(1) as reject_count 
                    from (
                        SELECT tos.order_id, tos.batch_no, tos.reject_time
                        FROM {$prefix}take_order AS tos
                    ) tos
                    group by tos.order_id, tos.batch_no
                ) tos",
                "op.order_id = tos.order_id and op.batch_no = tos.batch_no and op.is_reject = 1"
            )
            ->leftJoin(
                "(
                    select * from (
                        SELECT bp.product_id, bp.is_show_kitchen, bp.buffet_id
                        FROM {$prefix}buffet_product AS bp
                        where bp.buffet_id IN (SELECT buffet_id FROM {$prefix}order_buffet WHERE order_id = {$order_id})
                        ORDER BY bp.is_show_kitchen
                        LIMIT 99999
                    ) bp
                    group by bp.product_id
                ) bp",
                "p.product_id = bp.product_id"
            )
            ->where('op.order_id', '=', $order_id)
            ->when($is_scan, function ($q) {
                $q->where(function ($q) {
                    $q->where('op.is_send_kitchen', '=', 1)->whereOr('op.batch_no', '<>', '');  // 1.0.9 扫码端需要显示（送厨的 + 接单功能已下单(但未送厨的) ）
                });
            })
            ->when(!$is_scan, function ($q) {
                $q->where(function ($q) {
                    $q->where('op.is_send_kitchen', '=', 1);
                });
            })
            ->group('main_order_product_id, is_free, remark, product_price, product_attr')
            ->field('
                op.*,
                p.product_id,
                IFNULL(bp.is_show_kitchen, p.is_show_kitchen) as is_show_kitchen,
                tos.reject_time,
                sum(op.total_num) as total_num,
                sum(op.finish_num) as finish_num,
                sum(op.total_price) as total_price,
                sum(op.total_product_price) as total_product_price,
                sum(op.refund_money) as refund_money,
                sum(op.refund_num) as refund_num,
                sum(op.tax_rate) as tax_rate,
                sum(op.consumption_tax) as consumption_tax
            ')
            ->select();
        $result = [];
        // 自助餐顾客
        foreach ($buffet_customer_list as $buffet) {
            $addTime = strtotime($buffet['create_time']);
            if (!isset($result[$addTime])) {
                $result[$addTime] = [];
            }
            $result[$addTime]['type'] = 'buffet';
            $result[$addTime]['plist'][] = $buffet;
            $result[$addTime]['timestamp'] = $addTime;
            $result[$addTime]['date'] = date('H:i:s', strtotime($buffet['create_time']));
        }
        // 加钟
        foreach ($delay_list as $delay) {
            $addTime = strtotime($delay['create_time']);
            if (!isset($result[$addTime])) {
                $result[$addTime] = [];
            }
            $result[$addTime]['type'] = 'delay';
            $result[$addTime]['plist'][] = $delay;
            $result[$addTime]['timestamp'] = $addTime;
            $result[$addTime]['date'] = date('H:i:s', strtotime($delay['create_time']));
        }
        // 商品
        foreach ($orderProductList as $orderProduct) {
            $is_reject = isset($orderProduct['is_reject']) ? $orderProduct['is_reject'] : 0;
            $batch_time = isset($orderProduct['batch_time']) ? $orderProduct['batch_time'] : 0;
            $batch_no = isset($orderProduct['batch_no']) ? $orderProduct['batch_no'] : '';
            if (!empty($batch_no) && $batch_time > 1) {
                $sendKitchenTime = $batch_time;
            } else {
                $sendKitchenTime = $orderProduct['send_kitchen_time'] > 0 ? $orderProduct['send_kitchen_time'] : $batch_time;
            }
            if (!isset($result[$sendKitchenTime])) {
                $result[$sendKitchenTime] = [];
            }
            if (isset($orderProduct['is_show_kitchen'])) {
                $orderProduct['is_show_kitchen']  = ($orderProduct['finish_time'] ?? 0) > 0 ? 2 : $orderProduct['is_show_kitchen'];
            }
            $result[$sendKitchenTime]['type'] = 'product';
            $result[$sendKitchenTime]['plist'][] = $orderProduct;
            $result[$sendKitchenTime]['timestamp'] = $sendKitchenTime;
            $result[$sendKitchenTime]['date'] = date('H:i:s', $sendKitchenTime);
            $result[$sendKitchenTime]['take_date'] = $orderProduct['send_kitchen_time'] > 0 ? date('H:i:s', $orderProduct['send_kitchen_time']) : 0;
            if ($is_reject == 1) {
                $result[$sendKitchenTime]['is_take'] = -1;
                $result[$sendKitchenTime]['take_date'] = date('H:i:s', $orderProduct['reject_time']);   // TODO update_time可能不准确
            } else {
                $result[$sendKitchenTime]['is_take'] = $batch_time ? 1 : 0;
            }
            // 是否是扫码端的
            $result[$sendKitchenTime]['is_scan'] = !empty($batch_no) ? 1 : 0;
        }
        return $result;
    }

    /**
     * 数组同键名值相加合并
     * @param $a
     * @param $b
     * @param $num_key
     * @return array
     */
    public static function mergeConsumedArr($a, $b, $num_key = 'consumed')
    {
        $c = [];
        foreach ($a as $key => $value) {
            // 如果 $b 中存在相同的键，则将对应位置的值相加
            if (isset($b[$key])) {
                $c[$key] = [$num_key => $value[$num_key] + $b[$key][$num_key]];
            } else {
                $c[$key] = $value; // 如果 $b 中不存在相同的键，则直接复制 $a 中的值到 $c
            }
        }
        // 将 $b 中存在但 $a 中不存在的元素复制到 $c
        foreach ($b as $key => $value) {
            if (!isset($c[$key])) {
                $c[$key] = $value;
            }
        }
        return $c;
    }

    /**
     * 获取未送厨商品列表
     * @param $order_id
     * @param $source
     * @return array
     */
    public static function getUnSendArrList($order_id, $source = self::CASHIER_ADD_PRODUCT)
    {
        return (new OrderProduct)
            ->where('order_id', '=', $order_id)
            ->where('is_send_kitchen', '=', 0)
            ->where('batch_no', '=', '')
            ->where('add_source', '=', $source)
            ->select()->toArray();
    }

    /**
     * 赠菜
     * @param $free_tag_id
     * @param $free_tag
     * @param $free_remark
     * @return bool
     */
    public function toFree($free_tag_ids = [], $free_remark = '')
    {
        $this->startTrans();
        try {
            // 订单商品免单表记录
            if ($free_tag_ids) {
                (new OrderProductFree)->where('order_product_id', $this->order_product_id)->delete();
                $freeTagList = FreeTag::whereIn('id', $free_tag_ids)->select()->toArray();
                if ($freeTagList) {
                    $saveAllArr = [];
                    foreach ($freeTagList as $item) {
                        $saveAllArr[] = [
                            'free_tag_id' => $item['id'],
                            'free_tag' => $item['free_tag'],
                            'order_id' => $this->order_id,
                            'order_product_id' => $this->order_product_id,
                            'product_id' => $this->product_id,
                            'product_sku_id' => $this->product_sku_id,
                            'app_id' => $this->app_id,
                        ];
                    }
                    (new OrderProductFree)->saveAll($saveAllArr);
                }
            }
            // 订单商品记录
            $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $this['shop_supplier_id'] ?? 0, $this['app_id'] ?? 0);
            $saveArr = [
                'is_free' => $store['gift_method'] == '10' ? 1 : 2,
                'free_remark' => $free_remark,
            ];
            if (!$this->save($saveArr)) {
                $this->rollback();
                return false;
            }
            (new OrderModel)->reloadPrice($this->order_id);
            $this->commit();
            return true;
        } catch (BaseException $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 转移商品到子单
     * @param $id 商品ID
     * @param $fromOrderId 原订单ID
     * @param $toOrderId 目标订单ID
     * @param $num 数量
     * @return bool
     */
    public static function addToSubOrder($id, $fromOrderId, $toOrderId, $num)
    {
        $fromOrderProduct = self::where([
            'order_product_id' => $id,
            'sub_order_id' => $fromOrderId
        ])->find();

        if ($fromOrderProduct->total_num == $num) {
            // 转移全部数量
            $fromOrderProduct->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderProduct, $num);
            $fromOrderProduct->force()->delete();
        } else {
            // 转移部分数量
            // 从原子单中扣除部分数量
            $fromTotalNum = helper::bcsub((string)$fromOrderProduct->total_num, (string)$num);
            $fromTotalPrice = helper::bcmul((string)$fromOrderProduct->product_price, $fromTotalNum);
            $fromOrderProduct->total_num = (int)$fromTotalNum;
            $fromOrderProduct->total_price = (float)$fromTotalPrice;
            $fromOrderProduct->total_pay_price = (float)$fromTotalPrice;
            $fromOrderProduct->save();
            // 添加到目标子单
            $fromOrderProduct->sub_order_id = $toOrderId;
            self::addOrIncNum($fromOrderProduct, $num);
        }
    }

    /**
     * 添加商品或增加商品数量
     * @param $productData
     * @param int $num
     * @return bool
     */
    public static function addOrIncNum($orderProduct, $num = 1)
    {
        // 是否存在该商品
        /** @var OrderProduct $exist_product */
        $exist_product = (new OrderProduct)
            ->where('order_id', $orderProduct->order_id)
            ->where('main_order_product_id', $orderProduct->main_order_product_id)
            ->where('sub_order_id', $orderProduct->sub_order_id)
            ->where('product_attr', $orderProduct->getData('product_attr'))
            ->where('product_sku_id', $orderProduct->product_sku_id)
            ->where('add_source', $orderProduct->add_source)
            ->where('scheme_id', $orderProduct->scheme_id)
            ->where('is_free', $orderProduct->is_free)
            ->where('is_change_price', $orderProduct->is_change_price)
            ->where('product_price', $orderProduct->product_price)
            ->where('remark', $orderProduct->remark)
            ->where('is_send_kitchen', $orderProduct->is_send_kitchen)
            ->where('is_return', $orderProduct->is_return)
            ->where('batch_no', '')
            ->find();
        if ($exist_product && !$orderProduct->is_move) {
            $totalNum = helper::bcadd((string)$exist_product->total_num, (string)$num);
            $exist_product->save([
                'total_num' => (int)$totalNum,
                'total_price' => (float)helper::bcmul((string)$exist_product->product_price, $totalNum),
                'total_pay_price' => (float)helper::bcmul((string)$exist_product->product_price, $totalNum),
            ]);
        } else {
            // 保存商品
            $newOrderProductModel = new self();
            $newOrderProductModel->save([
                'order_id' => $orderProduct->order_id,
                'main_order_product_id' => $orderProduct->main_order_product_id,
                'app_id' => $orderProduct->app_id,
                'product_id' => $orderProduct->product_id,
                'product_name' => $orderProduct->product_name,
                'image_id' => $orderProduct->image_id,
                'deduct_stock_type' => $orderProduct->deduct_stock_type,
                'spec_type' => $orderProduct->spec_type,
                'content' => $orderProduct->content,
                'product_sku_id' => $orderProduct->product_sku_id,
                'product_attr' => $orderProduct->getData('product_attr'),
                'product_price' => $orderProduct->product_price,
                'line_price' => $orderProduct->line_price,
                'total_num' => $num,
                'total_price' => (float)helper::bcmul((string)$orderProduct->product_price, (string)$num),
                'total_pay_price' => (float)helper::bcmul((string)$orderProduct->product_price, (string)$num),
                'is_buffet_product' => $orderProduct->is_buffet_product,
                'feed_price' => $orderProduct->feed_price,
                'feed_uuids' => $orderProduct->getData('feed_uuids'),
                'attr_ids' => $orderProduct->getData('attr_ids'),
                'feed_ids' => $orderProduct->getData('feed_ids'),
                'add_source' => $orderProduct->add_source,
                'kitchen_is_open' => $orderProduct->kitchen_is_open,
                'is_send_kitchen' => $orderProduct->is_send_kitchen,
                'send_kitchen_time' => $orderProduct->send_kitchen_time,
                'is_free' => $orderProduct->is_free,
                'free_remark' => $orderProduct->free_remark,
                'is_move' => $orderProduct->is_move,
                'move_from_table_id' => $orderProduct->move_from_table_id,
                'move_from_order_id' => $orderProduct->move_from_order_id,
                'remark' => $orderProduct->remark,
                'is_change_price' => $orderProduct->is_change_price,
                'is_require' => $orderProduct->is_require,
                'is_return' => $orderProduct->is_return,
                'scheme_id' => $orderProduct->scheme_id,
                'sub_order_id' => $orderProduct->sub_order_id,
            ]);
        }
    }
}

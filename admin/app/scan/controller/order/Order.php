<?php

namespace app\scan\controller\order;

use help\QueueHelp;
use app\common\library\helper;
use app\common\model\buffet\Buffet;
use app\scan\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\order\OrderBuffet;
use app\common\model\order\OrderProduct;
use app\shop\model\buffet\BuffetProduct;
use app\common\enum\order\OrderErrorEnum;
use app\common\enum\settings\SettingEnum;
use app\scan\model\order\Order as OrderModel;
use app\scan\model\store\Table as TableModel;
use app\common\model\order\Order as OrderAlias;
use app\common\model\settings\Setting as SettingModel;
use app\cashier\service\order\settled\CashierOrderSettledService;

/**
 * 点餐
 * @Apidoc\Group("order")
 * @Apidoc\Sort(1)
 */
class Order extends Controller
{
    /**
     * @Apidoc\Title("桌台开台")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/setTable")
     * @Apidoc\Param("meal_num", type="int", require=true, desc="就餐人数")
     * @Apidoc\Param("is_buffet", type="int", require=true, desc="是否自助餐 0-否 1-是")
     * @Apidoc\Param("buffet_ids", type="array", require=false, desc="自助餐ID")
     * @Apidoc\Param("buffet_customer_type", type="array", require=false, desc="自助餐顾客类型信息")
     * @Apidoc\Returned()
     */
    public function setTable()
    {
        // 立即购买：获取订单商品列表
        $params = $this->postData();
        $params['eat_type'] = 10;
        $params['order_source'] = 10;
        $shopSupplierId = $this->table['shop_supplier_id'] ?? 0;
        $appId = $this->table['app_id'] ?? 0;
        $tableId = $this->table['table_id'] ?: 0;
        if ($params['is_buffet'] != 1 && ($params['meal_num'] > 999 || $params['meal_num'] < 1)) {
            return $this->renderError('请输入1-999的人数');
        }
        $params['table_id'] = $tableId;

        // 禁止并发操作 - 开台/转台
        $queue = new QueueHelp('TABLE_ORDER_ALL_' . $appId . '_' . $tableId);
        $queue->while();

        $table = TableModel::detail($tableId);
        if (!$table) {
            $queue->release();
            return $this->renderError('桌台不存在');
        }
        if ($table['status'] == 30) {
            $queue->release();
            return $this->renderError('桌台已开台');
        }
        // 自助餐
        if (($params['is_buffet'] ?? 0) == 1) {
            // 自助餐设置
            $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $shopSupplierId, $appId);
            if ($buffetSetting['is_open'] != 1) {
                $queue->release();
                return $this->renderError('未开启自助餐');
            }
            if (empty($params['buffet_ids'])) {
                $queue->release();
                return $this->renderError('请选择自助餐');
            }
            if (empty($params['buffet_customer_type'])) {
                $queue->release();
                return $this->renderError('请选择输入顾客人数');
            }
        }

        // 实例化订单service
        $user = [
            'cashier_id' => 0,
            'user_name' => '',
            'account' => '',
            'mobile' => NULL,
            'shop_supplier_id' => $this->table['shop_supplier_id'],
            'name' => '',
        ];
        $orderService = new CashierOrderSettledService($user, [], $params);
        // 订单信息初始化
        $orderInfo = $orderService->settlement();
        if ($orderService->hasError()) {
            $queue->release();
            return $this->renderError($orderService->getError());
        }
        //
        $orderInfo['device_id'] = $this->table['device_id'] ?? '';
        $orderInfo['add_source'] = OrderAlias::SCAN_PRODUCT_SOURCE;
        // 创建订单
        $order_id = $orderService->createOrder($orderInfo, $params['table_id']);
        if (!$order_id) {
            $queue->release();
            return $this->renderError($orderService->getError() ?: '订单创建失败');
        }
        // 返回结算信息
        $queue->release();
        return $this->renderSuccess('开台成功', ['order_id' => $order_id]);
    }

    /**
     * @Apidoc\Title("桌台未下单商品")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/getUnSendKitchen")
     * @Apidoc\Returned("list",type="array",ref="app\tablet\model\order\Order\getUnSendKitchen")
     */
    public function getUnSendKitchen()
    {
        $table_id = $this->table['table_id'] ?? 0;
        $model = new OrderModel();
        $detail = $model->getUnSendKitchen($table_id, OrderAlias::SCAN_PRODUCT_SOURCE);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $unSendList = $detail->unSendKitchenProduct ?? [];
        return $this->renderSuccess('', $unSendList);
    }

    /**
     * @Apidoc\Title("桌台已下单商品")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/getSendKitchen")
     * @Apidoc\Returned("list",type="array",ref="app\tablet\model\order\Order\getSendKitchen")
     */
    public function getSendKitchen()
    {
        $table_id = $this->table['table_id'] ?? 0;
        $detail = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $model = new OrderModel();
        $order = $model->getSendAndBatchKitchen($table_id);

        $list = [];
        $total_price = 0;

        if ($order) {
            $delay = $order['delay'] ?? [];
            $buffetCustomerType = $order['buffetCustomerType'] ?? [];

            $buffetCustomerTotalPrice = helper::getArrayColumnSum($order['buffetCustomerType'], 'total_price');
            $delayTotalPrice = helper::getArrayColumnSum($order['delay'], 'total_price');
            $sendKitchenProductTotalPrice = helper::getArrayColumnSum($order['sendAndBatchKitchenProduct'], 'total_product_price');
            $total_price = helper::bcadd(helper::bcadd($buffetCustomerTotalPrice, $delayTotalPrice), $sendKitchenProductTotalPrice);
            $list = OrderProduct::getGroupByTime($order['order_id'], $buffetCustomerType, $delay, [], 1);
            array_multisort(array_column($list, 'timestamp'), SORT_DESC, $list); //SORT_DESC降序，SORT_ASC升序
        }
        return $this->renderSuccess('请求成功', compact('list', 'total_price'));
    }

    /**
     * @Apidoc\Title("修改商品数量")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/sub")
     * @Apidoc\Param("order_product_id", type="int", require=true, desc="订单商品ID")
     * @Apidoc\Param("product_num", type="int", require=true, desc="商品数量")
     * @Apidoc\Param("type", type="string", require=true, desc="up-添加 down-减少")
     * @Apidoc\Returned()
     */
    public function sub($order_product_id)
    {
        $data = $this->postData();
        $sub_type = $data['type'] ?? 'up';
        $table_id = $this->table['table_id'] ?? 0;
        //
        $detail = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        /** @var OrderProduct $model */
        $model = OrderProduct::detail($order_product_id);
        if (!$model) {
            return $this->renderError('商品不存在', [], OrderErrorEnum::TABLET_SEND_PRODUCT_NOT_FOUND);
        }
        $product_id = $model->product_id;
        $order_id = $detail['order_id'];
        //
        // 后台配置
        $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $this->table['shop_supplier_id'] ?? 0, $this->table['app_id'] ?? 0);
        $this->table['setting_data'][SettingEnum::BUFFET]['values'] = $buffetSetting;
        $tabletSetting = $this->table['setting_data'][SettingEnum::TABLET]['values'] ?? [];
        //
        if ($detail['is_buffet'] == 1 && $sub_type == 'up' && $detail['buffet_expired_time'] != -1) {
            // 订单自助餐商品
            $buffet_ids = (new OrderBuffet)->where('order_id', $order_id)->column('buffet_id');
            $buffet_product_ids = (new BuffetProduct)->whereIn('buffet_id', $buffet_ids)->column('product_id');
            // 自助餐设置
            if ($model->is_buffet_product != 1) {
                $remain_continue_time_second = $buffetSetting['remain_continue_time'] ? $buffetSetting['remain_continue_time'] * 60 : 0;
                if ($buffetSetting['is_remain_continue'] == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                    return $this->renderError('点餐时间已到，无法继续下单');
                }
            }
            // 自助餐的点餐限制
            [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time] = OrderModel::getBuffetRemain($order_id, $buffet_ids);
            $remain_continue_time_second = $remain_continue_time * 60;
            if ($is_remain_continue == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                return $this->renderError('点餐时间已到，无法继续下单');
            }
        }
        //
        if ($sub_type == 'up') {
            // 下单限制
            // 自助餐
            if ($detail['is_buffet'] == 1 && $tabletSetting['is_buffet_order_limit'] == 1) {
                // 数量限制
                $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                $unSendNum++;
                if ($tabletSetting['buffet_order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['buffet_order_limit']['limit_num']) {
                    return $this->renderError('数量限制', ['value' => floatval($tabletSetting['buffet_order_limit']['limit_num'])], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                }
            }
            // 非自助餐
            if ($detail['is_buffet'] != 1 && $tabletSetting['is_order_limit'] == 1) {
                // 数量限制
                $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                $unSendNum++;
                if ($tabletSetting['order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['order_limit']['limit_num']) {
                    return $this->renderError('数量限制', ['value' => $tabletSetting['order_limit']['limit_num']], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                }
            }
        }
        //
        $data['setting_data'] = $this->table['setting_data'];
        if ($model->sub($data, OrderAlias::SCAN_PRODUCT_SOURCE)) {
            (new OrderModel())->reloadPrice($model['order_id']);
            return $this->renderSuccess('操作成功', $detail->getScanOrderInfo(1, 0, OrderProduct::SCAN_ADD_PRODUCT, $this->table['setting_data']));
        }
        return $this->renderError($model->getError() ?: '操作失败', $model->getErrorData(), $model->getErrorCode());
    }

    /**
     * @Apidoc\Title("添加商品")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/add")
     * @Apidoc\Param("product_id", type="int", require=true, desc="商品ID")
     * @Apidoc\Param("product_num", type="int", require=true, desc="商品数量")
     * @Apidoc\Param("product_sku_id", type="int", require=true, desc="商品SKU ID")
     * @Apidoc\Param("attr", type="int", require=false, desc="商品属性，如果有必填")
     * @Apidoc\Param("feed", type="string", require=false, desc="加料")
     * @Apidoc\Param("describe", type="string", require=false, desc="描述，拼接商品的规格，属性加料。如：小份;蒜蓉;番茄,茄子;")
     * @Apidoc\Param("is_buffet", type="int", require=true, desc="是否自助餐商品 0-否 1-是")
     * @Apidoc\Param("type", type="string", require=true, desc="up-添加 down-减少")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $data = $this->postData();
        $table_id = $this->table['table_id'] ?? 0;
        $sub_type = $data['type'] ?? 'up';
        $detail = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $order_id = $detail['order_id'];
        //
        $productSkuId = intval($data['product_sku_id'] ?? 0);
        $describe = $data['describe'] ?? '';
        $scheme_id = $data['scheme_id'] ?? 0;
        // 后台配置
        $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $this->table['shop_supplier_id'] ?? 0, $this->table['app_id'] ?? 0);
        $this->table['setting_data'][SettingEnum::BUFFET]['values'] = $buffetSetting;
        $tabletSetting = $this->table['setting_data'][SettingEnum::TABLET]['values'] ?? [];
        // 用餐时间检查
        if ($detail['is_buffet'] == 1 && $detail['buffet_expired_time'] != -1 && $buffetSetting['is_buy_continue'] != 1 && $detail['buffet_expired_time'] < time()) {
            return $this->renderError('用餐时间已到，无法继续下单');
        }
        // 是否存在该商品
        /** @var OrderProduct $exist_product */
        $exist_product = (new OrderProduct)
            ->where('order_id', $order_id)
            ->where('product_attr', $describe)
            ->where('product_sku_id', $productSkuId)
            ->where('add_source', OrderProduct::SCAN_ADD_PRODUCT)
            ->where('batch_time', 0)    // 扫码端未下单的
            ->where('scheme_id', $scheme_id)    // 扫码端未下单的
            ->where('is_send_kitchen', 0)
            ->where('remark', '')
            ->find();

        // 存在修改数量
        if ($exist_product) {
            $product_id = $exist_product['product_id'];
            if ($detail['is_buffet'] == 1 && $sub_type == 'up' && $detail['buffet_expired_time'] != -1) {
                // 订单自助餐商品
                $buffet_ids = (new OrderBuffet)->where('order_id', $order_id)->column('buffet_id');
                $buffet_product_ids = (new BuffetProduct)->whereIn('buffet_id', $buffet_ids)->column('product_id');
                // 自助餐设置
                $remain_continue_time_second = $buffetSetting['remain_continue_time'] ? $buffetSetting['remain_continue_time'] * 60 : 0;
                if ($buffetSetting['is_remain_continue'] == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                    return $this->renderError('点餐时间已到，无法继续下单');
                }
                // 自助餐的点餐限制
                [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time] = OrderModel::getBuffetRemain($order_id, $buffet_ids);
                $remain_continue_time_second = $remain_continue_time * 60;
                if ($is_remain_continue == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                    return $this->renderError('点餐时间已到，无法继续下单');
                }
            }

            if ($sub_type == 'up') {
                // 下单限制
                // 自助餐
                if ($detail['is_buffet'] == 1 && $tabletSetting['is_buffet_order_limit'] == 1) {
                    // 数量限制
                    $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                    $unSendNum++;
                    if ($tabletSetting['buffet_order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['buffet_order_limit']['limit_num']) {
                        return $this->renderError('数量限制', ['value' => floatval($tabletSetting['buffet_order_limit']['limit_num'])], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                    }
                }
                // 非自助餐
                if ($detail['is_buffet'] != 1 && $tabletSetting['is_order_limit'] == 1) {
                    // 数量限制
                    $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                    $unSendNum++;
                    if ($tabletSetting['order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['order_limit']['limit_num']) {
                        return $this->renderError('数量限制', ['value' => $tabletSetting['order_limit']['limit_num']], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                    }
                }
            }

            $productNum =  $sub_type == 'up' ? $exist_product['total_num'] + 1 : $exist_product['total_num'] - 1;
            $param = [
                'type' => $sub_type,
                'product_num' => $productNum,
                'setting_data' => $this->table['setting_data']
            ];
            if ($exist_product->sub($param, OrderAlias::SCAN_PRODUCT_SOURCE)) {
                (new OrderModel())->reloadPrice($exist_product['order_id']);
                $o = OrderModel::detail($order_id);
                return $this->renderSuccess('操作成功', $o->getScanOrderInfo(0, 0, OrderProduct::SCAN_ADD_PRODUCT, $this->table['setting_data']));
            }
            return $this->renderError($exist_product->getError() ?: '操作失败', $exist_product->getErrorData());
        } else {
            if ($sub_type != 'up') {
                return $this->renderSuccess('操作成功', $detail->getScanOrderInfo(0, 0, OrderProduct::SCAN_ADD_PRODUCT, $this->table['setting_data']));
            }
            // 不存在新加一条
            unset($data['table_id']);
            $data['order_id'] = $order_id;
            $data['eat_type'] = 10;
            $data['add_source'] = OrderProduct::SCAN_ADD_PRODUCT;    // 添加来源 1-收银 2-桌台 3-扫码
            $product_id = $data['product_id'] ?? 0;

            if ($detail['is_buffet'] == 1 && $detail['buffet_expired_time'] != -1) {
                // 订单自助餐商品
                $buffet_ids = (new OrderBuffet)->where('order_id', $order_id)->column('buffet_id');
                $buffet_product_ids = (new BuffetProduct)->whereIn('buffet_id', $buffet_ids)->column('product_id');
                // 全局的自助餐设置
                if ($data['is_buffet'] != 1) {
                    $remain_continue_time_second = $buffetSetting['remain_continue_time'] ? $buffetSetting['remain_continue_time'] * 60 : 0;
                    if ($buffetSetting['is_remain_continue'] == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                        return $this->renderError('点餐时间已到，无法继续下单');
                    }
                }
                // 自助餐的点餐限制
                [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time] = OrderModel::getBuffetRemain($order_id, $buffet_ids);
                $remain_continue_time_second = $remain_continue_time * 60;
                if ($is_remain_continue == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second && in_array($product_id, $buffet_product_ids)) {
                    return $this->renderError('点餐时间已到，无法继续下单');
                }
            }

            // 下单限制
            // 自助餐
            if ($detail['is_buffet'] == 1 && $tabletSetting['is_buffet_order_limit'] == 1) {
                // 数量限制
                $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                $unSendNum++;
                if ($tabletSetting['buffet_order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['buffet_order_limit']['limit_num']) {
                    return $this->renderError('数量限制', ['value' => floatval($tabletSetting['buffet_order_limit']['limit_num'])], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                }
            }
            // 非自助餐
            if ($detail['is_buffet'] != 1 && $tabletSetting['is_order_limit'] == 1) {
                // 数量限制
                $unSendNum = OrderProduct::where('order_id', '=', $order_id)->where('add_source', OrderProduct::TABLET_ADD_PRODUCT)->where('is_send_kitchen', '=', 0)->sum('total_num'); // 未送厨商品数量
                $unSendNum++;
                if ($tabletSetting['order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['order_limit']['limit_num']) {
                    return $this->renderError('数量限制', ['value' => $tabletSetting['order_limit']['limit_num']], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
                }
            }

            $model = new OrderModel();
            $data['product_num'] = 1;
            $order_id = $model->addToOrder($data, $this->table, '', OrderAlias::SCAN_PRODUCT_SOURCE);
            if ($order_id > 0) {
                $o = OrderModel::detail($order_id);
                return $this->renderSuccess('添加商品成功', $o->getScanOrderInfo(0, 0, OrderProduct::SCAN_ADD_PRODUCT, $this->table['setting_data']));
            }
            return $this->renderError($model->getError() ?: '添加商品失败', $model->getErrorData(), $model->getErrorCode());
        }
    }

    /**
     * @Apidoc\Title("桌台下单(送厨)")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/sendKitchen")
     * @Apidoc\Param("add_product_arr", type="array", require=false, desc="添加要送厨商品")
     * @Apidoc\Param("ignore_must", type="int", require=false, desc="忽略必点商品 0-否 1-是")
     * @Apidoc\Returned()
     */
    public function sendKitchen()
    {
        $param = $this->postData();
        $ignore_must = isset($param['ignore_must']) ? $param['ignore_must'] : 0;
        //
        $table_id = $this->table['table_id'] ?? 0;
        $detail = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $order_id = $detail['order_id'];
        //
        $add_product_arr = OrderProduct::getUnSendArrList($order_id, OrderProduct::SCAN_ADD_PRODUCT);

        // 自助餐用餐时间限制下单
        if ($detail['is_buffet'] == 1 && $detail['buffet_expired_time'] != -1) {
            // 自助餐设置
            [$is_remain_continue, $remain_continue_notice_time, $remain_continue_time] = OrderModel::getBuffetRemain($detail['order_id']);
            $remain_continue_time_second = $remain_continue_time * 60;
            //
            $buffetSetting = SettingModel::getSupplierItem(SettingEnum::BUFFET, $detail['shop_supplier_id'] ?? 0, $detail['app_id'] ?? 0);
            // 用餐时间检查
            if ($buffetSetting['is_buy_continue'] != 1 && $detail['buffet_expired_time'] < time()) {
                return $this->renderError('用餐时间已到，无法继续下单');
            }
            // 自助餐商品检查
            foreach ($add_product_arr as $product) {
                if ($product['is_buffet_product'] == 1 && $is_remain_continue == 1 && $detail['buffet_remaining_time'] < $remain_continue_time_second) {
                    return $this->renderError('自助餐时间已到达，自助餐商品不可继续下单');
                }
            }
        }
        /**
         *  H5下单冷却、数量限制
         */
        $tabletSetting = SettingModel::getSupplierItem(SettingEnum::H5, $this->table['shop_supplier_id'] ?? 0, $this->table['app_id'] ?? 0);
        // 自助餐
        if ($detail['is_buffet'] == 1 && $tabletSetting['is_buffet_order_limit'] == 1) {
            // 送厨冷却
            if ($tabletSetting['buffet_order_limit']['is_limit_time'] == 1) {
                $lastTimestamp = OrderProduct::where('order_id', '=', $order_id)
                    ->where('is_send_kitchen', '=', 1)
                    ->where('send_kitchen_source', '=', OrderProduct::SCAN_SEND_KITCHEN)
                    ->order('send_kitchen_time desc')->value('send_kitchen_time') ?? 0;
                //
                $lastBatchTimestamp = OrderProduct::where('order_id', '=', $order_id)
                    ->where('is_send_kitchen', '=', 0)
                    ->where('batch_no', '<>', '')
                    ->where('send_kitchen_source', '=', 1)
                    ->order('batch_time desc')->value('batch_time') ?? 0;
                $lastTimestamp = max($lastTimestamp, $lastBatchTimestamp);
                $pastTimeSecond = (time() - $lastTimestamp);  //
                $limitTimeSecond = helper::bcmul($tabletSetting['buffet_order_limit']['limit_time'], 60);
                $coolingTimeSecond = helper::bcsub($limitTimeSecond, $pastTimeSecond);
                if ($lastTimestamp != 0 && $coolingTimeSecond > 0) {
                    return $this->renderError('时间限制', ['value' => floatval($coolingTimeSecond)], OrderErrorEnum::TABLET_SEND_TIME_LIMIT);
                }
            }
            // 数量限制
            $unSendNum = array_sum(array_column($add_product_arr, 'total_num'));
            if ($tabletSetting['buffet_order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['buffet_order_limit']['limit_num']) {
                return $this->renderError('数量限制', ['value' => floatval($tabletSetting['buffet_order_limit']['limit_num'])], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
            }
        }
        // 非自助餐
        if ($detail['is_buffet'] != 1 && $tabletSetting['is_order_limit'] == 1) {
            // 送厨冷却
            if ($tabletSetting['order_limit']['is_limit_time'] == 1) {
                $lastTimestamp = OrderProduct::where('order_id', '=', $order_id)
                    ->where('is_send_kitchen', '=', 1)
                    ->where('send_kitchen_source', '=', OrderProduct::SCAN_SEND_KITCHEN)
                    ->order('send_kitchen_time desc')->value('send_kitchen_time') ?? 0;
                //
                $lastBatchTimestamp = OrderProduct::where('order_id', '=', $order_id)
                    ->where('is_send_kitchen', '=', 0)
                    ->where('batch_no', '<>', '')
                    ->where('send_kitchen_source', '=', 1)
                    ->order('batch_time desc')->value('batch_time') ?? 0;
                $lastTimestamp = max($lastTimestamp, $lastBatchTimestamp);
                $pastTimeSecond = (time() - $lastTimestamp);  //
                $limitTimeSecond = helper::bcmul($tabletSetting['order_limit']['limit_time'], 60);
                $coolingTimeSecond = helper::bcsub($limitTimeSecond, $pastTimeSecond);
                if ($lastTimestamp != 0 && $coolingTimeSecond > 0) {
                    return $this->renderError('时间限制', ['value' => floatval($coolingTimeSecond)], OrderErrorEnum::TABLET_SEND_TIME_LIMIT);
                }
            }
            // 数量限制
            $unSendNum = array_sum(array_column($add_product_arr, 'total_num'));
            if ($tabletSetting['order_limit']['is_limit_num'] == 1 && $unSendNum > $tabletSetting['order_limit']['limit_num']) {
                return $this->renderError('数量限制', ['value' => $tabletSetting['order_limit']['limit_num']], OrderErrorEnum::TABLET_SEND_NUM_LIMIT);
            }
        }

        $model = new OrderProduct();
        if ($model->addAndSendKitchen([], $this->table, $order_id, 'kitchen', true, 40, OrderProduct::SCAN_SEND_KITCHEN, OrderAlias::SCAN_PRODUCT_SOURCE, $ignore_must)) {
            return $this->renderSuccess('下单成功');
        }
        return $this->renderError($model->getError() ?: '下单失败', $model->getErrorData(), $model->getErrorCode());
    }

    /**
     * @Apidoc\Title("删除商品")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/delProduct")
     * @Apidoc\Param("order_product_id", type="int|array", require=true, desc="订单商品ID, 多个传数组: [1,2]")
     * @Apidoc\Returned()
     */
    public function delProduct($order_product_id)
    {
        $table_id = $this->table['table_id'] ?? 0;
        $order = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$order) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $model = new OrderProduct();
        if ($model->delProduct($order_product_id)) {
            return $this->renderSuccess('删除成功');
        };
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("桌台-自助餐列表")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/buffetList")
     * @Apidoc\Returned()
     */
    public function buffetList()
    {
        $list = Buffet::getList();
        return $this->renderSuccess('', $list);
    }

    /**
     * @Apidoc\Title("确认必点商品")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.order/confirmMust")
     * @Apidoc\Param("order_id", type="int", require=true, desc="订单ID")
     * @Apidoc\Returned()
     */
    public function confirmMust()
    {
        $param = $this->postData();
        $order_id = $param['order_id'] ?? 0;
        $orderModel = (new OrderModel);
        /** @var OrderModel $order */
        $order = $orderModel->underwayDetail($order_id, []);
        if (!$order) {
            return $this->renderError($orderModel->getError() ?: '操作失败');
        }
        if (!$order->confirmMust()) {
            return $this->renderError($order->getError() ?: '操作失败', $order->getErrorData(), $order->getErrorCode());
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("获取桌台订单必点方案及商品基础信息")
     * @Apidoc\Method("GET")
     * @Apidoc\Url ("/index.php/scan/order.order/getScheme")
     * @Apidoc\Param("order_id", type="int", require=true, desc="订单ID")
     * @Apidoc\Returned()
     */
    public function getScheme()
    {
        $param = $this->getData();
        $order_id = $param['order_id'] ?? 0;
        $orderModel = (new OrderModel);
        /** @var OrderModel $order */
        $order = $orderModel->underwayDetail($order_id, []);
        if (!$order) {
            return $this->renderError($orderModel->getError() ?: '操作失败');
        }
        return $this->renderSuccess('请求成功', $order->getSchemeBaseProductList());
    }

    /**
     * @Apidoc\Title("修改备注")
     * @Apidoc\Method("POST")
     * @Apidoc\Url ("/index.php/scan/order.Order/remark")
     * @Apidoc\Param("order_product_id", type="int", require=true, desc="订单商品ID")
     * @Apidoc\Param("remark", type="string", require=false, desc="备注，最多50个字")
     * @Apidoc\Returned()
     */
    public function remark()
    {
        $param = $this->postData();
        $table_id = $this->table['table_id'] ?? 0;
        $orderProductId = $param['order_product_id'] ?? 0;
        $remark = $param['remark'] ?? '';

        if (!$orderProductId) {
            return $this->renderError('订单商品ID不能为空');
        }
        if ($remark && mb_strlen($remark) > 50) {
            return $this->renderError('备注最多50个字');
        }
        $order = OrderModel::getScanTableUnderwayOrder($table_id);
        if (!$order) {
            return $this->renderError('桌台已关闭', [], -4);
        }

        $model = new OrderProduct();
        $orderId = $model->updateKitchenRemark($orderProductId, $remark);
        if ($orderId > 0) {
            return $this->renderSuccess('备注成功');
        }
        return $this->renderError($model->getError() ?: '备注失败');
    }
}

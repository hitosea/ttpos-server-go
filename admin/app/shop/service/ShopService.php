<?php

namespace app\shop\service;

use app\common\library\helper;
use app\shop\model\order\Order;
use app\shop\model\product\Product;
use app\common\model\erp\ErpPurchaseOrder;
use app\common\repositories\OrderBusinessDataRepository;

/**
 * 商城模型
 */
class ShopService
{
    // 商品模型
    private $ProductModel;
    // 订单模型
    private $OrderModel;
    // 采购单模型
    private $ErpPurchaseOrderModel;
    // 错误信息
    public $error = '';

    /**
     * 构造方法
     */
    public function __construct()
    {
        /* 初始化模型 */
        $this->ProductModel = new Product();
        $this->OrderModel = new Order();
        $this->ErpPurchaseOrderModel = new ErpPurchaseOrder();
    }

    /**
     * 后台首页数据
     */
    public function getHomeData($user)
    {
        $today = date('Y-m-d');
        $yesterday = date('Y-m-d', strtotime('-1 day'));
        $shop_supplier_id = 0;
        if ($user['user_type'] == 1) {
            $shop_supplier_id = $user['shop_supplier_id'];
        }

        // 营业数据
        $params['shop_supplier_id'] = $shop_supplier_id;
        $businessData = (new OrderBusinessDataRepository($this->OrderModel, $params, 0))->setSource('HomeData')->getBusinessData();
        if (empty($businessData)) {
            $this->error = '请求失败';
            return false;
        }
        $todayBusinessData = (new OrderBusinessDataRepository($this->OrderModel, $params, 1))->setSource('HomeData')->getBusinessData();
        if (empty($todayBusinessData)) {
            $this->error = '请求失败';
            return false;
        }
        $ytdBusinessData = (new OrderBusinessDataRepository($this->OrderModel, $params, 2))->setSource('HomeData')->getBusinessData();
        if (empty($ytdBusinessData)) {
            $this->error = '请求失败';
            return false;
        }
        // 库存告急
        $stockProductList = (new Product())->getList([
            "status" => -1,
            "product_type" => 1,
            "product_name" => "",
            "category_id" => 0,
            "page" => 1,
            "list_rows" => 99999,
            "type" => "",
            "stock" => 10,
            "material_type" => 10
        ]);
        $purchaseApply = ErpPurchaseOrder::where('status', 0)->count() ?: 0;
        // 当天汇总
        $data = [
            'top_data' => [
                // 商品总量
                'product_total' => 0,
                // 用户总量
                'user_total' => $businessData['member_data']['user_count'],
                // 订单总量
                'order_total' => $businessData['total_order_num'],
                // 店铺总量
                'supplier_total' => 0,
                // 实收金额
                'total_money' => $businessData['total_received_price'],
                // 优惠折扣
                'discount_money' => $businessData['total_discount_money'],
                // 会员折扣
                'user_discount_money' => $businessData['total_user_discount_money'],
                // 退款金额
                'refund_money' => $businessData['total_refund_money'],
            ],
            'wait_data' => [
                // 订单
                'order' => [
                    'disposal' => 0,
                ],
                // 库存
                'stock' => [
                    'product' => $stockProductList['total'],
                ],
                // 采购单
                'purchase' => [
                    'apply' => $purchaseApply,
                ],
            ],
            'today_data' => [
                // 销售额(元)
                'order_total_price' => [
                    'tday' => $todayBusinessData['total_received_price'],
                    'ytd' => $ytdBusinessData['total_received_price']
                ],
                // 支付订单数
                'order_total' => [
                    'tday' => $todayBusinessData['total_order_num'],
                    'ytd' => $ytdBusinessData['total_order_num']
                ],
                // 新增用户数
                'new_user_total' => [
                    'tday' => $todayBusinessData['member_data']['user_count'],
                    'ytd' => $ytdBusinessData['member_data']['user_count']
                ],
                // 新供应商数
                'new_supplier_total' => [
                    'tday' => 0,
                    'ytd' => 0
                ],
                // 下单用户数
                'order_user_total' => [
                    'tday' => $this->getPayOrderUserTotal($today, $shop_supplier_id),
                    'ytd' => $this->getPayOrderUserTotal($yesterday, $shop_supplier_id)
                ],
                // 优惠折扣(元)
                'discount_money' => [
                    'tday' => $todayBusinessData['total_discount_money'],
                    'ytd' => $ytdBusinessData['total_discount_money'],
                ],
                // 会员折扣(元)
                'user_discount_money' => [
                    'tday' => $todayBusinessData['total_user_discount_money'],
                    'ytd' => $ytdBusinessData['total_user_discount_money'],
                ],
                // 退款金额(元)
                'order_refund_money' => [
                    'tday' => $todayBusinessData['total_refund_money'],
                    'ytd' => $ytdBusinessData['total_refund_money']
                ],
            ],
            'product_data' => [
                // 销量排行
                'salesNumRank' => $this->OrderModel->getProductRank(0, -1, $shop_supplier_id),
                // 销售额排行
                'salesMoneyRank' => $this->OrderModel->getProductRank(1, -1, $shop_supplier_id),
            ],
        ];
        return $data;
    }

    /**
     * 获取采购单数量
     */
    public function getPurchaseOrderCount($shop_supplier_id)
    {
        // 待审核
        return $this->ErpPurchaseOrderModel->getPurchaseOrderCount(ErpPurchaseOrder::STATUS_WAIT, $shop_supplier_id);
    }

    /**
     * 获取预计收入
     */
    private function getOrderIncome($day, $shop_supplier_id = 0)
    {
        return number_format($this->OrderModel->getOrderData($day, null, 'income_price', $shop_supplier_id));
    }

    /**
     * 获取商品总量
     */
    private function getProductTotal($where = [])
    {
        return number_format($this->ProductModel->getProductTotal($where), 0, '.', '');
    }

    /**
     * 获取商品总量
     */
    private function getSupplierProductTotal($shop_supplier_id, $product_type, $product_status = 0)
    {
        return number_format($this->ProductModel->getSupplierProductTotal($shop_supplier_id, $product_type, $product_status), 0, '.', '');
    }

    /**
     * 获取某天的下单用户数
     */
    private function getPayOrderUserTotal($day, $shop_supplier_id = 0)
    {
        return 0;
    }

    /**
     * 商品数据
     */
    public function getProductData($param)
    {
        $data = [
            // 商品总量
            'product_total' => $this->getSupplierProductTotal($param['shop_supplier_id'], $param['product_type']),
            // 上架商品总量
            'up_total' => $this->getSupplierProductTotal($param['shop_supplier_id'], $param['product_type'], 10),
            // 下架商品总量
            'down_total' => $this->getSupplierProductTotal($param['shop_supplier_id'], $param['product_type'], 20),
        ];
        return $data;
    }

    /**
     * 商品数据
     */
    public function getProductRank($data, $type)
    {
        return $this->ProductModel->getProductRank($data, $type);
    }
}

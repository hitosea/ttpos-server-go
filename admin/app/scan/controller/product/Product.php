<?php

namespace app\scan\controller\product;

use app\common\model\order\Order;
use hg\apidoc\annotation as Apidoc;
use app\scan\controller\Controller;
use app\scan\model\product\Product as ProductModel;

/**
 * 商品
 * @Apidoc\Group("product")
 */
class Product extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/scan/product.product/index")
     * @Apidoc\Param("category_id", type="int", require=true, desc="商品分类ID")
     * @Apidoc\Param("search", type="string", require=false, default="", desc="搜索关键字")
     * @Apidoc\Param("is_special", type="int", require=false, default="", desc="是否特色分类 0-否 1-是")
     * @Apidoc\Param("table_id", type="int", require=false, desc="桌台ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list",type="array")
     */
    public function index()
    {
        // 获取全部商品列表
        $param = $this->postData();
        $model = new ProductModel;
        $table_id = $this->table['table_id'] ?? 0;
        $order = Order::getScanTableUnderwayOrder($table_id);
        if (!$order) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $param['order_id'] = $order['order_id'];
        $param['product_source'] = Order::SCAN_PRODUCT_SOURCE;
        $list = $model->list(array_merge(['shop_supplier_id' => $this->table['shop_supplier_id']], $param));
        if ($order) {
            $buffetProductArr = Order::getOrderBuffetProductArr($order['order_id']);    // 自助餐商品
            $list = Order::handleBuffetProductIndex($list, $buffetProductArr, $order['meal_num']);
        }

        return $this->renderSuccess('', compact('list'));
    }


    /**
     * @Apidoc\Title("详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/scan/product.product/detail")
     * @Apidoc\Param("product_id", type="int", require=false, default="0", desc="商品id")
     * @Apidoc\Returned("list",type="array",ref="app\common\model\product\Product\detail")
     */
    public function detail($product_id)
    {
        $table_id = $this->table['table_id'] ?? 0;
        $order = Order::getScanTableUnderwayOrder($table_id);
        if (!$order) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        // 商品详情
        $detail = ProductModel::detail($product_id);
        if (!$detail) {
            return $this->renderError('商品不存在');
        }

        $buffetProductArr = Order::getOrderBuffetProductArr($order['order_id']);
        $detail = Order::handleBuffetProductDetail($detail, $buffetProductArr);

        return $this->renderSuccess('', $detail);
    }
}

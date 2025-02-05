<?php

namespace app\scan\controller\product;

use app\common\model\order\Order;
use app\scan\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\scan\model\product\Category as CategoryModel;

/**
 * 分类
 * @Apidoc\Group("product")
 */
class Category extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/scan/product.category/index")
     * @Apidoc\Returned("list",type="array",ref="app\scan\model\product\Category\getCashierALL")
     */
    public function index()
    {
        $table_id = $this->table['table_id'] ?? 0;
        $order = Order::getScanTableUnderwayOrder($table_id);
        if (!$order) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $model = new CategoryModel;

        $list = $model->getOrderProductCountByCategory(1, $order->order_id, Order::SCAN_PRODUCT_SOURCE);
        return $this->renderSuccess('', compact('list'));
    }
}

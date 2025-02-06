<?php

namespace app\menu\controller\product;

use app\common\model\order\Order;
use app\menu\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\menu\model\product\Category as CategoryModel;

/**
 * 分类
 * @Apidoc\Group("product")
 */
class Category extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/menu/product.category/index")
     * @Apidoc\Returned("list",type="array",ref="app\menu\model\product\Category\getCashierALL")
     */
    public function index()
    {
        $sid = $this->table['shop_supplier_id'] ?? 0;
        $model = new CategoryModel;
        $list = $model->getProductCountByCategory(1, $sid, 0, 1, true);
        return $this->renderSuccess('', compact('list'));
    }
}

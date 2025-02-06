<?php

namespace app\menu\controller\product;

use hg\apidoc\annotation as Apidoc;
use app\menu\controller\Controller;
use app\menu\model\product\Product as ProductModel;

/**
 * 商品
 * @Apidoc\Group("product")
 */
class Product extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/menu/product.product/index")
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
        $list = $model->list(array_merge(['shop_supplier_id' => $this->table['shop_supplier_id']], $param));
        return $this->renderSuccess('', compact('list'));
    }
}

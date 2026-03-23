<?php

namespace app\shop\controller\store;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\store\TableArea;
use app\common\model\product\Category;
use app\shop\model\order\Order as OrderModel;
use app\shop\service\statistics\ProductService;

/**
 * 商品销售统计（v1.0.8）
 * @Apidoc\Group("statistics")
 * @Apidoc\Sort(8)
 */
class Product extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.product/index")
     * @Apidoc\Param("time_type", type="int", require=false, default="", desc="时间类型 1-今天 2-昨天 3-本周 0-全部")
     * @Apidoc\Param("date", type="array", require=false, default="", desc="起始日期")
     * @Apidoc\Param("area_id", type="int", require=false, default="", desc="区域ID")
     * @Apidoc\Param("category_id", type="int", require=false, default="", desc="商品分类ID")
     * @Apidoc\Param("product_name", type="int", require=false, default="", desc="商品名称")
     * @Apidoc\Param("product_name", type="int", require=false, default="", desc="商品名称")
     * @Apidoc\Param("sort_field", type="string", require=false, default="", desc="排序字段 product_num-销售数量 sales_price-销售额")
     * @Apidoc\Param("sort_type", type="string", require=false, default="", desc="排序类型 asc-正序 desc-倒序")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="订单概况", children={
     *    @Apidoc\Param("product_name",type="int",default=0,desc="商品名称"),
     *    @Apidoc\Param("category_name",type="int",default=0,desc="商品分类"),
     *    @Apidoc\Param("product_num",type="int",default=0,desc="销售数量"),
     *    @Apidoc\Param("sales_price",type="int",default=0,desc="销售额"),
     *    @Apidoc\Param("free_product_num",type="int",default=0,desc="赠菜"),
     *    @Apidoc\Param("received_price",type="int",default=0,desc="实收金额"),
     *    @Apidoc\Param("business_price",type="int",default=0,desc="营业收入"),
     * })
     * @Apidoc\Returned("area_list", type="array", desc="区域列表", children={
     *    @Apidoc\Returned("area_name", type="string", desc="区域名称（多语言）"),
     *    @Apidoc\Returned("area_name_text", type="string", desc="区域名名称"),
     *    @Apidoc\Returned("area_id", type="int", desc="销量"),
     * })
     * @Apidoc\Returned("category_list", type="array", desc="商品分类", children={
     *    @Apidoc\Returned("category_id", type="int", desc="销量"),
     *    @Apidoc\Returned("category_name_text", type="string", desc="分类名称"),
     *    @Apidoc\Returned("category_name", type="string", desc="分类名称（多语言）"),
     * })
     */
    public function index()
    {
        return $this->renderError('This API service is currently under maintenance. Please download and use the TTPOS Shop client.');
        $data = $this->postData() ?: [];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?? 0;
        $model = new OrderModel();
        $list = $model->productSales($data);
        // 区域列表
        $area_list = TableArea::getAllList($shop_supplier_id);
        // 商品分类列表
        $category_list = Category::getCacheTree(1, 0, $this->store);
        //
        return $this->renderSuccess('', compact('list', 'area_list', 'category_list'));
    }

    /**
     *
     * @Apidoc\Title("导出")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.product/export")
     * @Apidoc\Param("time_type", type="int", require=false, default="", desc="时间类型 1-今天 2-昨天 3-本周 0-全部")
     * @Apidoc\Param("date", type="array", require=false, default="", desc="起始日期")
     * @Apidoc\Param("area_id", type="int", require=false, default="", desc="区域ID")
     * @Apidoc\Param("category_id", type="int", require=false, default="", desc="商品分类ID")
     * @Apidoc\Param("product_name", type="int", require=false, default="", desc="商品名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned()
     */
    public function export()
    {
        $data = $this->postData();
        $data['list_rows'] = 1000;
        $model = new OrderModel;
        $list = $model->productSales($data);
        if (empty($list['data'])) {
            return $this->renderError('导出数据为空');
        }
        (new ProductService())->productExport($list['data']);
        return $this->renderSuccess('');
    }
}

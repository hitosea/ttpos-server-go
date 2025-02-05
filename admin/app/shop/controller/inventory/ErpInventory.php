<?php

namespace app\shop\controller\inventory;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\erp\ErpMonthlyStatistics;
use app\shop\model\product\Category as CategoryModel;
use app\shop\model\product\ProductSku as ProductSkuModel;

/**
 * 库存管理
 * @Apidoc\Group("inventory")
 * @Apidoc\Sort(10)
 */
class ErpInventory extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpInventory/list")
     * @Apidoc\Param("material_type", type="int", require=false, desc="商品类型 类型 10-成品 20-材料")
     * @Apidoc\Param("product_status", type="int", require=false, desc="商品状态 产品状态(10上架 20下架 )")
     * @Apidoc\Param("category_id", type="int", default=0, require=false, desc="分类id")
     * @Apidoc\Param("keyword", type="string", require=false, desc="商品名称/条形码")
     * @Apidoc\Param("stock_num", type="int", require=false, desc="库存数量")
     * @Apidoc\Param("sort", type="string", require=false, desc="库存排序 asc-升序 desc-降序")
     * @Apidoc\Param("filter_having_material", type="int", default=0, require=false, desc="是否过滤有材料的商品 1-是 0-否")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\ProductSku\getSkuProductList", desc="库存列表",  children={
     *     @Apidoc\Returned("history_purchase_num", type="int", desc="历史进货数"),
     *     @Apidoc\Returned("history_loss_num", type="int", desc="历史报损数"),
     *     @Apidoc\Returned("product_sales", type="int", desc="历史销售数"),
     *  })
     * )
     */
    public function list()
    {
        $data = $this->postData();
        $filterHavingMaterial = isset($data['filter_having_material']) ? $data['filter_having_material'] : 0;
        $list = (new ProductSkuModel)->getSkuProductList($data, $filterHavingMaterial);
        $category = CategoryModel::getCacheTree(1, 0, $this->store);
        return $this->renderSuccess('', compact('list', 'category'));
    }

    /**
     * @Apidoc\Title("月度报表信息")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpInventory/monthlyStatistics")
     * @Apidoc\Param("date", type="int", require=false, desc="月份 2024-05")
     * @Apidoc\Returned()
     */
    public function monthlyStatistics()
    {
        $params = $this->postData();
        $month_start_stock = ErpMonthlyStatistics::getMonthStartStock($params); //  月初原有库存
        $month_end_stock = ErpMonthlyStatistics::getMonthEndStock($params); //  月末剩余库存
        $month_entry_stock = ErpMonthlyStatistics::getMonthEntry($params); // 月入库数量
        $month_exit_stock = ErpMonthlyStatistics::getMonthExit($params);                  //  月出库数量
        $month_damaged_num = ErpMonthlyStatistics::getMonthDamagedNum($params);           //  损耗数量
        $month_damaged_percent = ErpMonthlyStatistics::getMonthDamagedPercent($month_damaged_num, $month_entry_stock, $month_start_stock); //  损耗比例

        $data = [
            'info' => [
                'month_start_stock' => $month_start_stock,           //  月初原有库存
                'month_end_stock' => $month_end_stock,               //  月末剩余库存
                'month_entry_stock' => $month_entry_stock,           //  月入库数量
                'month_exit_stock' => $month_exit_stock,             //  月出库数量
                'month_damaged_num' => $month_damaged_num,           //  损耗数量
                'month_damaged_percent' => $month_damaged_percent,   //  损耗比例
            ],
            'damaged_list' => ErpMonthlyStatistics::getMonthProductDamagedList($params),            //  月初损耗排行
            'unsalable_list' => ErpMonthlyStatistics::getMonthProductUnsalableList($params),        //  月末滞销商品排行
        ];
        return $this->renderSuccess('', $data);
    }
}

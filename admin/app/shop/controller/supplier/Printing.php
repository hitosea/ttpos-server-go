<?php

namespace app\shop\controller\supplier;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\product\Label as LabelModel;
use app\shop\model\settings\Printer as PrinterModel;
use app\shop\model\product\Category as CategoryModel;
use app\shop\model\supplier\Printing as PrintingModel;

/**
 * 商品打印
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(6)
 */
class Printing extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.Printing/index")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\supplier\Printing\getList", desc="列表")
     */
    public function index()
    {
        // 供应商列表
        $model = new PrintingModel;
        $list = $model->getLists($this->postData());
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加(get-获取信息/post-添加）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.Printing/add")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="名称")
     * @Apidoc\Param("printer_id", type="array", require=true, default="", desc="打印机ids(v1.0.6)")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启 0-关闭 1-开启")
     * @Apidoc\Param("product_type", type="int", require=true, default=0, desc="用餐类型 0-外卖 1-店内")
     * @Apidoc\Param("print_type", type="int", require=true, default=0, desc="打印模式 10-付款打印 20-下单打印")
     * @Apidoc\Param("category_id", type="array", require=true, default="", desc="商品分组，菜品id")
     * @Apidoc\Param("area_id", type="array", require=true, default="", desc="区域分组，区域id")
     * @Apidoc\Param("is_open_one_food", type="int", require=true, default=0, desc="是否开启一菜一单 0-关闭 1-开启")
     * @Apidoc\Param("type", type="int", require=true, default=0, desc="打印类型 10-小票打印 20-标签打印")
     * @Apidoc\Param("print_method", type="int", require=true, default=0, desc="打印方式 10-整单打印 20-商品分组打印 30-按标签打印 40-按一菜一标签打印")
     * @Apidoc\Param("print_select", type="int", require=true, default="", desc="打印选择 1-合并打印 2-分开打印(v1.0.6)")
     * @Apidoc\Param("product_method", type="array", require=true, default="", desc="商品方式 1-按商品分类 2-按打印标签（v1.0.8）")
     * @Apidoc\Param("product_ids", type="array", require=true, default="", desc="商品ids（v1.0.8）")
     * @Apidoc\Returned("printerList", type="array", ref="app\shop\model\settings\Printer\getNoteAll", desc="打印机列表")
     * @Apidoc\Returned("printerTagList", type="array", ref="app\shop\model\settings\Printer\getTagAll", desc="标签打印机列表")
     * @Apidoc\Returned("storeList", type="array", ref="app\shop\model\product\Category\getAllCategory", desc="店内商品分类")
     * @Apidoc\Returned("labelList", type="array", ref="app\shop\model\product\Label\getAllList", desc="打印标签")
     */
    public function add()
    {
        $model = new PrintingModel;
        if ($this->request->isGet()) {
            $printerList = PrinterModel::getNoteAll($this->store['user']['shop_supplier_id']);
            $printerTagList = PrinterModel::getTagAll($this->store['user']['shop_supplier_id']);
            $storeList = (new CategoryModel)->getAllCategory(1, $this->store['user']['shop_supplier_id'], 0, 0, true);
            $labelList = (new LabelModel)->getAllList($this->store['user']['shop_supplier_id']);
            return $this->renderSuccess('', compact('printerList', 'printerTagList', 'storeList', 'labelList'));
        }
        //
        $param = $this->postData();
        if (mb_strlen($param['name'] ?? '') > 50) {
            return $this->renderError('名称限制输入50字符');
        }
        $print_method = $param['print_method'] ?? 0;
        if (!in_array($param['print_select'], [1, 2]) && $print_method == 40) {
            return $this->renderError('请正确选择合并或分开打印');
        }
        $product_method = $param['product_method'] ?? 0;
        if (!in_array($product_method, [1, 2])) {
            return $this->renderError('请正确选择商品方式');
        }
        // 新增记录
        if ($model->add($param)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑(get-获取信息/post-添加）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.Printing/edit")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="id")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="名称")
     * @Apidoc\Param("printer_id", type="array", require=true, default="", desc="打印机ids(v1.0.6)")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启 0-关闭 1-开启")
     * @Apidoc\Param("product_type", type="int", require=true, default=0, desc="用餐类型 0-外卖 1-店内")
     * @Apidoc\Param("print_type", type="int", require=true, default=0, desc="打印模式 10-付款打印 20-下单打印")
     * @Apidoc\Param("category_id", type="array", require=true, default="", desc="商品分组，菜品id")
     * @Apidoc\Param("type", type="int", require=true, default=0, desc="打印类型 10-小票打印 20-标签打印")
     * @Apidoc\Param("print_method", type="int", require=true, default=0, desc="打印方式 10-整单打印 20-商品分组打印 30-按标签打印")
     * @Apidoc\Param("print_select", type="int", require=true, default="", desc="打印选择 1-合并打印 2-分开打印(v1.0.6)")
     * @Apidoc\Param("product_method", type="array", require=true, default="", desc="商品方式 1-按商品分类 2-按打印标签（v1.0.8）")
     * @Apidoc\Param("product_ids", type="array", require=true, default="", desc="商品ids（v1.0.8）")
     * @Apidoc\Returned("printerList", type="array", ref="app\shop\model\settings\Printer\getNoteAll", desc="打印机列表")
     * @Apidoc\Returned("printerTagList", type="array", ref="app\shop\model\settings\Printer\getTagAll", desc="标签打印机列表")
     * @Apidoc\Returned("storeList", type="array", ref="app\shop\model\product\Category\getAllCategory", desc="店内商品分类")
     * @Apidoc\Returned("labelList", type="array", ref="app\shop\model\product\Label\getAllList", desc="打印标签")
     */
    public function edit($id)
    {
        $param = $this->postData();
        /** @var PrintingModel $model */
        $model = PrintingModel::detail($id, [
            'printingItem',
            'printingRegion',
            'printingProductItem' => [
                'product' => [ 'category' ]
            ],
        ]);
        if ($this->request->isGet()) {
            $model = $model->buildDetailData();
            // 获取打印机列表
            $printerList = PrinterModel::getNoteAll($this->store['user']['shop_supplier_id']);
            // 获取标签打印机列表
            $printerTagList = PrinterModel::getTagAll($this->store['user']['shop_supplier_id']);
            // 店内商品分类
            $storeList = (new CategoryModel)->getAllCategory(1, $this->store['user']['shop_supplier_id'], 0, 0, true);
            // 打印标签
            $labelList = (new LabelModel)->getAllList($this->store['user']['shop_supplier_id']);
            return $this->renderSuccess('', compact('model', 'printerList', 'printerTagList', 'storeList', 'labelList'));
        }
        //
        if (mb_strlen($param['name'] ?? '') > 50) {
            return $this->renderError('名称限制输入50字符');
        }
        $print_method = $param['print_method'] ?? 0;
        if (!in_array($param['print_select'], [1, 2]) && $print_method == 40) {
            return $this->renderError('请正确选择合并或分开打印');
        }
        $product_method = $param['product_method'] ?? 0;
        if (!in_array($product_method, [1, 2])) {
            return $this->renderError('请正确选择商品方式');
        }
        //
        /** @var PrintingModel $model*/
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.Printing/delete")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="id")
     * @Apidoc\Returned()
     */
    public function delete($id)
    {
        // 详情
        $model = PrintingModel::detail($id, [
            'printingItem',
            'printingRegion',
            'printingProductItem',
        ]);
        /** @var PrintingModel $model */
        if (!$model->setDelete()) {
            return $this->renderError('删除失败');
        }
        return $this->renderSuccess($model->getError() ?: '删除成功');
    }

    /**
     * @Apidoc\Title("状态")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.Printing/setStatus")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="id")
     * @Apidoc\Param("status", type="int", require=true, default=0, desc="状态 0-关闭 1-开启")
     * @Apidoc\Returned()
     */
    public function setStatus($id, $status)
    {
        // 详情
        $model = PrintingModel::detail($id);
        /** @var PrintingModel $model */
        if (!$model->setStatus($status)) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess($model->getError() ?: '操作成功');
    }
}

<?php

namespace app\shop\controller\product\expand;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\product\Unit as UnitModel;

/**
 * 单位库
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Unit extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.unit/index")
     * @Apidoc\Param("unit_name", type="string", require=false, default="", desc="单位名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Unit\getList", desc="列表")
     */
    public function index()
    {
        $model = new UnitModel;
        $list  = $model->getList($this->postData(), $this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.unit/add")
     * @Apidoc\Param("unit_name", type="string", require=true, desc="单位名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new UnitModel();
        if ($data = $model->add($this->postData(), $this->store['user']['shop_supplier_id'])) {
            return $this->renderSuccess('添加成功', $data);
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.unit/edit")
     * @Apidoc\Param("unit_id", type="int", require=true, desc="单位id")
     * @Apidoc\Param("unit_name", type="string", require=true, desc="单位名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function edit($unit_id)
    {
        /** @var UnitModel $model */
        $model = UnitModel::detail($unit_id);
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.unit/delete")
     * @Apidoc\Param("unit_id", type="string", require=true, desc="单位id，多个逗号隔开")
     * @Apidoc\Returned()
     */
    public function delete($unit_id)
    {
        $model = new UnitModel;
        if (!$model->setDelete($unit_id)) {
            return $this->renderError($model->getError() ?: '删除失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("关联菜品（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.unit/relatedProduct")
     * @Apidoc\Param("unit_id", type="int", require=true, desc="标签id")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="产品ids")
     * @Apidoc\Returned()
     */
    public function relatedProduct()
    {
        $data        = $this->postData();
        $unit_id     = $data['unit_id'] ?? 0;
        $product_ids = $data['product_ids'] ?? [];
        //
        /** @var UnitModel $model */
        $model = UnitModel::detail($unit_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->relatedProduct($unit_id, $product_ids)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }
}

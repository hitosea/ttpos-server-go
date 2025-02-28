<?php

namespace app\shop\controller\product\expand;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\product\Spec as SpecModel;

/**
 * 商品规格
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Spec extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/index")
     * @Apidoc\Param("spec_name", type="string", require=false, default="", desc="规格名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Spec\getList", desc="列表")
     */
    public function index()
    {
        $model = new SpecModel;
        $list  = $model->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加规格")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/add")
     * @Apidoc\Param("spec_name", type="string", require=true, desc="规格名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new SpecModel();
        if ($data = $model->add($this->postData())) {
            return $this->renderSuccess('添加成功', $data ?: []);
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑规格")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/edit")
     * @Apidoc\Param("spec_id", type="int", require=true, desc="规格id")
     * @Apidoc\Param("spec_name", type="string", require=true, desc="规格名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function edit($spec_id)
    {
        /** @var SpecModel $model */
        $model = SpecModel::detail($spec_id);
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除规格")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/delete")
     * @Apidoc\Param("spec_id", type="string", require=true, desc="规格id，多个逗号隔开")
     * @Apidoc\Returned()
     */
    public function delete($spec_id)
    {
        $model = new SpecModel;
        if (!$model->setDelete($spec_id)) {
            return $this->renderError($model->getError() ?: '删除失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("关联菜品（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/relatedProduct")
     * @Apidoc\Param("spec_id", type="int", require=true, desc="标签id")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="产品ids")
     * @Apidoc\Returned()
     */
    public function relatedProduct()
    {
        $data        = $this->postData();
        $spec_id     = $data['spec_id'] ?? 0;
        $product_ids = $data['product_ids'] ?? [];
        //
        /** @var SpecModel $model */
        $model = SpecModel::detail($spec_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->relatedProduct($spec_id, $product_ids)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("修改价格（v1.0.8）")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/batchPrice")
     * @Apidoc\Param("spec_id", type="int", require=true, desc="规格ID")
     * @Apidoc\Param("products", type="array", require=true, desc="产品信息数组", children={
     *     @Apidoc\Param("product_id", type="int", require=true, desc="产品ID"),
     *     @Apidoc\Param("product_sku_id", type="int", require=true, desc="产品规格ID"),
     *     @Apidoc\Param("product_price", type="float", require=true, desc="产品规格价格"),
     * })
     * @Apidoc\Returned()
     */
    public function batchPrice()
    {
        $data    = $this->postData();
        $spec_id = $data['spec_id'] ?? 0;
        /** @var SpecModel $model */
        $model = SpecModel::detail($spec_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->batchPrice($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("规格产品列表（v1.0.8）")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.spec/skuProduct")
     * @Apidoc\Param("spec_id", type="int", require=true, desc="规格ID")
     * @Apidoc\Returned()
     */
    public function skuProduct()
    {
        $data    = $this->postData();
        $spec_id = $data['spec_id'] ?? 0;
        if (!$spec_id) {
            return $this->renderError('参数错误');
        }
        /** @var SpecModel $model */
        $model = new SpecModel;
        $data = $model->skuProduct($spec_id);
        return $this->renderSuccess('', $data);
    }
}

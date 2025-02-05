<?php

namespace app\shop\controller\product\expand;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\product\Attribute as AttributeModel;


/**
 * 商品属性
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Attr extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.attr/index")
     * @Apidoc\Param("type", type="int", require=false, default=1, desc="类型 1-属性组 2-属性值")
     * @Apidoc\Param("parent_ids", type="int", require=false, default=0, desc="父级ID，英文逗号隔开，用于左边筛选")
     * @Apidoc\Param("attribute_name", type="string", require=false, default="", desc="属性名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Attribute\getList", desc="列表")
     */
    public function index()
    {
        $model = new AttributeModel;
        $list  = $model->getList($this->postData(), $this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.attr/add")
     * @Apidoc\Param("parent_id", type="int", require=true, default=0, desc="父级ID 0-顶级")
     * @Apidoc\Param("attribute_name", type="string", require=true, desc="属性名称")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new AttributeModel();
        if ($model->add($this->postData(), $this->store['user']['shop_supplier_id'])) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.attr/edit")
     * @Apidoc\Param("attribute_id", type="int", require=true, desc="属性id")
     * @Apidoc\Param("attribute_name", type="string", require=true, desc="属性名称")
     * @Apidoc\Returned()
     */
    public function edit($attribute_id)
    {
        /** @var AttributeModel $model */
        $model = AttributeModel::detail($attribute_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.attr/delete")
     * @Apidoc\Param("attribute_id", type="string", require=true, desc="属性id，多个逗号隔开")
     * @Apidoc\Returned()
     */
    public function delete($attribute_id)
    {
        $model = new AttributeModel;
        if (!$model->setDelete($attribute_id)) {
            return $this->renderError($model->getError() ?: '删除失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("关联菜品（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.attr/relatedProduct")
     * @Apidoc\Param("attribute_id", type="int", require=true, desc="标签id")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="产品ids")
     * @Apidoc\Returned()
     */
    public function relatedProduct()
    {
        $data         = $this->postData();
        $attribute_id = $data['attribute_id'] ?? 0;
        $product_ids  = $data['product_ids'] ?? [];
        //
        /** @var AttributeModel $model */
        $model = AttributeModel::detail($attribute_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->relatedProduct($attribute_id, $product_ids)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }
}

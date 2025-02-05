<?php

namespace app\shop\controller\product\expand;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\product\Label as LabelModel;

/**
 * 打印标签
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Label extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.label/index")
     * @Apidoc\Param("label_name", type="string", require=false, default="", desc="标签名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Label\getList", desc="列表")
     */
    public function index()
    {
        $model = new LabelModel;
        $list = $model->getList($this->postData(), $this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.label/add")
     * @Apidoc\Param("label_name", type="string", require=true, desc="标签名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new LabelModel();
        $data = $this->postData();
        if ($data['label_name'] == '') {
            return $this->renderError('标签名称不能为空');
        }
        if ($model->add($data, $this->store['user']['shop_supplier_id'])) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.label/edit")
     * @Apidoc\Param("label_id", type="int", require=true, desc="标签id")
     * @Apidoc\Param("label_name", type="string", require=true, desc="标签名称")
     * @Apidoc\Param("sort", type="int", require=false, default="100", desc="排序")
     * @Apidoc\Returned()
     */
    public function edit($label_id)
    {
        $data = $this->postData();
        if ($data['label_name'] == '') {
            return $this->renderError('标签名称不能为空');
        }
        /** @var LabelModel $model */
        $model = LabelModel::detail($label_id);
        if ($model->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.label/delete")
     * @Apidoc\Param("label_id", type="string", require=true, desc="标签id，多个逗号隔开")
     * @Apidoc\Returned()
     */
    public function delete($label_id)
    {
        $model = new LabelModel;
        if (!$model->setDelete($label_id)) {
            return $this->renderError($model->getError() ?: '删除失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("关联菜品（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/product.expand.label/relatedProduct")
     * @Apidoc\Param("label_id", type="int", require=true, desc="标签id")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="产品ids")
     * @Apidoc\Returned()
     */
    public function relatedProduct()
    {
        $data = $this->postData();
        $label_id = $data['label_id'] ?? 0;
        $product_ids = $data['product_ids'] ?? [];
        //
        /** @var LabelModel $model */
        $model = LabelModel::detail($label_id);
        if (!$model) {
            return $this->renderError('数据不存在');
        }
        if ($model->relatedProduct($label_id, $product_ids)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }
}

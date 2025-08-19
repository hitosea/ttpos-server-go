<?php

namespace app\shop\controller\supplier;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\settings\PrinterTemplate as PrinterTemplateModel;

/**
 * 打印模板
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(7)
 */
class PrinterTemplate extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.PrinterTemplate/list")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned()
     */
    public function list()
    {
        $model = new PrinterTemplateModel;
        $list = $model->getList($this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('', compact('list'));
    }


    /**
     * @Apidoc\Title("设置模板")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/supplier.PrinterTemplate/setTemplate")
     * @Apidoc\Param("id", type="int", require=true, default="", desc="id")
     * @Apidoc\Param("template", type="int", require=true, default="", desc="模板编号")
     * @Apidoc\Returned()
     */
    public function setTemplate()
    {
        $id = $this->postData()['id'];
        $data = $this->postData();
        $model = PrinterTemplateModel::detail($id);
        /** @var PrinterTemplateModel $model */
        if (!$model) {
            return $this->renderError('模板不存在');
        }
        if (!$model->setTemplate($data)) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess($model->getError() ?: '操作成功');
    }
}

<?php

namespace app\shop\controller\inventory;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\erp\ErpDamagedProductRecord as ErpDamagedProductRecordModel;

/**
 * 报损记录
 * @Apidoc\Group("inventory")
 * @Apidoc\Sort(10)
 */
class ErpDamagedProductRecord extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpDamagedProductRecord/list")
     * @Apidoc\Param("type", type="int", require=false, desc="损坏类型 1-丢失 2-损坏")
     * @Apidoc\Param("date", type="array", require=false, desc="起始日期")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\erp\ErpDamagedProductRecord\getList")
     */
    public function list()
    {
        $model = new ErpDamagedProductRecordModel;
        $list = $model->getList($this->postData());
        $chart_list = $model->getChartList($this->postData());
        return $this->renderSuccess('', compact('list', 'chart_list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpDamagedProductRecord/add")
     * @Apidoc\Param("product_id", type="int", require=true, desc="商品ID")
     * @Apidoc\Param("product_sku_id", type="int", require=true, desc="商品规格ID")
     * @Apidoc\Param("type", type="int", require=true, desc="损坏类型 1-丢失 2-损坏")
     * @Apidoc\Param("num", type="int", require=true, desc="报损数量")
     * @Apidoc\Param("remark", type="int", default=0, require=false, desc="备注")
     * @Apidoc\Param("operator_id", type="int", require=true, desc="操作人ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new ErpDamagedProductRecordModel;
        $param = $this->postData();
        $param['operator_id'] = $this->store['user']['shop_user_id'];
        $param['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        if ($model->add($param)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model?->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpDamagedProductRecord/delete")
     * @Apidoc\Param("id", type="int", require=true, desc="ID")
     * @Apidoc\Returned()
     */
    public function delete($id)
    {
        $detail = (new ErpDamagedProductRecordModel)->detail($id ?? 0);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail->delete()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($detail?->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("更改")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpDamagedProductRecord/update")
     * @Apidoc\Param("id", type="int", require=true, desc="ID")
     * @Apidoc\Param("product_id", type="int", require=false, desc="商品ID")
     * @Apidoc\Param("num", type="int", require=false, desc="报损数量")
     * @Apidoc\Param("remark", type="string", require=false, desc="备注")
     * @Apidoc\Returned()
     */
    public function update()
    {
        $param = $this->postData();
        $param['operator_id'] = $this->store['user']['shop_user_id'];
        $param['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        $model = new ErpDamagedProductRecordModel;
        if ($model->edit($param)) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }

    /**
     * @Apidoc\Title("审核")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/inventory.ErpDamagedProductRecord/review")
     * @Apidoc\Param("id", type="int", require=true, desc="ID")
     * @Apidoc\Param("review_status", type="int", require=true, desc="审核结果 1-通过 2-拒绝")
     * @Apidoc\Param("refused", type="int", require=true, desc="拒绝原因")
     * @Apidoc\Returned()
     */
    public function review()
    {
        $params = $this->postData();
        $detail = (new ErpDamagedProductRecordModel)->detail($params['id'] ?? 0);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail->review($params)) {
            return $this->renderSuccess('审核成功');
        }
        return $this->renderError($detail->getError() ?: '审核失败');
    }
}

<?php

namespace app\shop\controller\purchase;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\erp\ErpWarehouseForm;

/**
 * 入库记录
 * @Apidoc\Group("purchase")
 * @Apidoc\Sort(20)
 */
class ErpInventoryRecordIn extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpInventoryRecordIn/list")
     * @Apidoc\Param("type", type="int", require=false, desc="操作类型 10-采购入库 20-调整入库 21-添加入库 30-销售出库 40-调整出库 41-删除出库")
     * @Apidoc\Param("date", type="array", require=true, default="", desc="起始时间 date[0]开始时间 date[1]结束时间")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\erp\ErpInventoryRecord\getList")
     */
    public function list()
    {
        $param = $this->postData();
        $list = (new ErpWarehouseForm)->getList(array_merge($param));
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("撤销")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpInventoryRecordIn/cancel")
     * @Apidoc\Param("erp_inventory_id", type="int", require=true, desc="库存id")
     * @Apidoc\Returned()
     */
    public function cancel($erp_inventory_id)
    {
        $detail = (new ErpWarehouseForm)->detail($erp_inventory_id ?? 0);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail?->cancel()) {
            return $this->renderSuccess('撤销成功');
        }
        return $this->renderError($detail?->getError() ?: '撤销失败');
    }

    /**
     * @Apidoc\Title("删除库存")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpInventoryRecordIn/delete")
     * @Apidoc\Param("erp_inventory_id", type="int", require=true, desc="库存id")
     * @Apidoc\Returned()
     */
    public function delete($erp_inventory_id)
    {
        $detail = (new ErpWarehouseForm)->detail($erp_inventory_id ?? 0);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail?->del()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($detail?->getError() ?: '删除失败');
    }
}

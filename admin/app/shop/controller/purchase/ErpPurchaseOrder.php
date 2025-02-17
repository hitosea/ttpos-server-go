<?php

namespace app\shop\controller\purchase;

use app\common\model\shop\User;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\User as UserModel;
use app\shop\model\erp\ErpSupplier as ErpSupplierModel;
use app\shop\model\erp\ErpPurchaseOrder as ErpPurchaseOrderModel;

/**
 * 采购单
 * @Apidoc\Group("purchase")
 * @Apidoc\Sort(10)
 */
class ErpPurchaseOrder extends Controller
{
    /**
     * @Apidoc\Title("采购单列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/list")
     * @Apidoc\Param("keyword", type="string", require=false, desc="采购单名称/申请人")
     * @Apidoc\Param("erp_supplier_id", type="int", require=false, desc="erp供应商id")
     * @Apidoc\Param("purchaser_id", type="int", require=false, desc="采购员id")
     * @Apidoc\Param("status", type="int", require=false, desc="状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库")
     * @Apidoc\Param("date", type="array", require=true, default="", desc="起始时间 date[0]开始时间 date[1]结束时间")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\erp\ErpPurchaseOrder\getList", children={
     *      @Apidoc\Returned("erp_supplier_names", type="int", desc="采购商名称"),
     *      @Apidoc\Returned("erp_purchase_names", type="int", desc="采购员名称"),
     *      @Apidoc\Returned("applicant_name", type="decimal", desc="申请人名称"),
     * })
     */
    public function list()
    {
        $list = (new ErpPurchaseOrderModel)->getList($this->postData());
        $supplier_list = (new UserModel)->getList();
        $erp_supplier_list = (new ErpSupplierModel)->getSelectList();
        return $this->renderSuccess('', compact('list', 'supplier_list', 'erp_supplier_list'));
    }

    /**
     * @Apidoc\Title("采购单详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/api/shop/purchase.ErpPurchaseOrder/detail")
     * @Apidoc\Param("purchase_order_id", type="int", require=true, desc="采购单id")
     * @Apidoc\Returned("detail", type="object", ref="app\shop\model\erp\ErpPurchaseOrder\detail", children={
     *      @Apidoc\Returned("supplier_name", type="int", desc="店铺名称"),
     *      @Apidoc\Returned("applicant_name", type="int", desc="申请人名称"),
     *      @Apidoc\Returned("details", type="array", ref="app\common\model\erp\ErpPurchaseDetail\lists", desc="产品-列表", children={
     *          @Apidoc\Returned("product_type", type="int", desc="产品类型 10-成品 20-材料"),
     *          @Apidoc\Returned("product_name", type="int", desc="产品名称"),
     *          @Apidoc\Returned("product_stock", type="int", desc="产品库存"),
     *          @Apidoc\Returned("productImage", type="int", desc="产品图片"),
     *          @Apidoc\Returned("product_sku_name", type="int", desc="规格"),
     *          @Apidoc\Returned("erp_supplier_name", type="decimal", desc="供应商名称"),
     *      }),
     *      @Apidoc\Returned("logs", type="array", ref="app\shop\model\erp\ErpPurchaseOperationLog\lists", desc="日志-列表")
     * })
     */
    public function detail($purchase_order_id)
    {
        $detail = (new ErpPurchaseOrderModel)->detail($purchase_order_id ?? 0);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        return $this->renderSuccess('', compact('detail'));
    }

    /**
     * @Apidoc\Title("添加采购单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/add")
     * @Apidoc\Param("name", type="string", require=true, desc="采购单名称")
     * @Apidoc\Param("applicant_id", type="int", require=true, desc="申请人id")
     * @Apidoc\Param("arrival_time", type="string", require=true, desc="到货时间")
     * @Apidoc\Param("type", type="int", require=true, desc="采购方式 10-总部采购 20-自行采购")
     * @Apidoc\Param("purchase_detail", type="array", require=true, desc="采购单明细", children={
     *      @Apidoc\Returned("product_sku_id", type="int", desc="商品规格id"),
     *      @Apidoc\Returned("product_id", type="int", desc="商品id"),
     *      @Apidoc\Returned("estimate_purchase_price", type="decimal", desc="预计采购价格"),
     *      @Apidoc\Returned("estimate_purchase_num", type="int", desc="预计采购数量"),
     * })
     * @Apidoc\Param("remark", type="string", require=false, desc="备注")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $data = $this->postData();
        $model = new ErpPurchaseOrderModel;
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        $data['shop_user_id'] = $this->store['user']['shop_user_id'];

        if (empty($data['name'])) {
            return $this->renderError('采购名称不能为空');
        }
        if (mb_strlen($data['name']) > 100) {
            return $this->renderError('采购名称最大100个字符');
        }
        if (empty($data['applicant_id'])) {
            return $this->renderError('申请人不能为空');
        }
        if (!User::checkUserExist($data['applicant_id'])) {
            return $this->renderError('申请人不存在，请重新选择');
        }
        if (empty($data['remark'])) {
            return $this->renderError('备注不能为空');
        }
        if (mb_strlen($data['remark']) > 100) {
            return $this->renderError('备注最大100个字符');
        }

        if ($model?->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model?->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑采购单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/edit")
     * @Apidoc\Param("purchase_order_id", type="int", require=true, desc="采购单id")
     * @Apidoc\Param("name", type="string", require=true, desc="采购单名称")
     * @Apidoc\Param("applicant_id", type="int", require=true, desc="申请人id")
     * @Apidoc\Param("arrival_time", type="string", require=true, desc="到货时间")
     * @Apidoc\Param("type", type="int", require=true, desc="采购方式 10-总部采购 20-自行采购")
     * @Apidoc\Param("purchase_detail", type="array", require=true, desc="采购单明细", children={
     *      @Apidoc\Returned("product_sku_id", type="int", desc="商品规格id"),
     *      @Apidoc\Returned("product_id", type="int", desc="商品id"),
     *      @Apidoc\Returned("estimate_purchase_price", type="decimal", desc="预计采购价格"),
     *      @Apidoc\Returned("estimate_purchase_num", type="int", desc="预计采购数量"),
     * })
     * @Apidoc\Param("remark", type="string", require=false, desc="备注")
     * @Apidoc\Returned()
     */
    public function edit()
    {
        $data = $this->postData();
        $purchase_order_id = $data['purchase_order_id'] ?? 0;
        if (empty($purchase_order_id)) {
            return $this->renderError('采购单ID不能为空');
        }
        if (empty($data['name'])) {
            return $this->renderError('采购名称不能为空');
        }
        if (mb_strlen($data['name']) > 100) {
            return $this->renderError('采购名称最大100个字符');
        }
        if (empty($data['applicant_id'])) {
            return $this->renderError('申请人不能为空');
        }
        if (!User::checkUserExist($data['applicant_id'])) {
            return $this->renderError('申请人不存在，请重新选择');
        }
        if (empty($data['remark'])) {
            return $this->renderError('备注不能为空');
        }
        if (mb_strlen($data['remark']) > 100) {
            return $this->renderError('备注最大100个字符');
        }
        $detail = (new ErpPurchaseOrderModel)->detail($purchase_order_id);
        if (!$detail) {
            return $this->renderError('采购单ID不存在');
        }
        $data['shop_user_id'] = $this->store['user']['shop_user_id'];
        if ($detail?->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($detail?->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("调整采购单数据")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/adjust")
     * @Apidoc\Param("purchase_order_id", type="int", require=true, desc="采购单id")
     * @Apidoc\Param("purchase_detail", type="array", require=true, desc="采购单明细", children={
     *      @Apidoc\Returned("purchase_detail_id", type="int", desc="明细id"),
     *      @Apidoc\Returned("actual_purchase_price", type="decimal", desc="实际采购价格"),
     *      @Apidoc\Returned("actual_purchase_num", type="int", desc="实际采购数量"),
     * })
     * @Apidoc\Returned()
     */
    public function adjust()
    {
        $data = $this->postData();
        $purchase_order_id = $data['purchase_order_id'] ?? 0;
        if (empty($purchase_order_id)) {
            return $this->renderError('采购单ID不能为空');
        }
        $detail = (new ErpPurchaseOrderModel)->detail($purchase_order_id);
        if (!$detail) {
            return $this->renderError('采购单ID不存在');
        }
        $data['shop_user_id'] = $this->store['user']['shop_user_id'];
        if ($detail?->adjust($data,  $this->store['user'])) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($detail?->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("操作采购单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/operate")
     * @Apidoc\Param("purchase_order_id", type="int", require=true, desc="采购单id")
     * @Apidoc\Param("status", type="int", require=true, desc="操作状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库")
     * @Apidoc\Param("remark", type="string", require=false, desc="备注/驳回理由")
     * @Apidoc\Returned()
     */
    public function operate()
    {
        $data = $this->postData();
        $purchase_order_id = $data['purchase_order_id'] ?? 0;
        if (empty($purchase_order_id)) {
            return $this->renderError('采购单ID不能为空');
        }
        $detail = (new ErpPurchaseOrderModel)->detail($purchase_order_id);
        if (!$detail) {
            return $this->renderError('采购单ID不存在');
        }
        $data['shop_user_id'] = $this->store['user']['shop_user_id'];
        if ($detail?->operate($data)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($detail?->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("删除采购单")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpPurchaseOrder/delete")
     * @Apidoc\Param("purchase_order_id", type="int", require=true, desc="采购单id")
     * @Apidoc\Returned()
     */
    public function delete()
    {
        $data = $this->postData();
        $purchase_order_id = $data['purchase_order_id'] ?? 0;
        if (empty($purchase_order_id)) {
            return $this->renderError('采购单ID不能为空');
        }
        $detail = (new ErpPurchaseOrderModel)->simpleDetail($purchase_order_id);
        if (!$detail) {
            return $this->renderError('采购单ID不存在');
        }
        if ($detail?->del()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($detail?->getError() ?: '删除失败');
    }
}

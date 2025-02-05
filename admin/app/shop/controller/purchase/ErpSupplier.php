<?php

namespace app\shop\controller\purchase;

use app\common\model\shop\User;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\erp\ErpSupplier as ErpSupplierModel;

/**
 * 供应商管理
 * @Apidoc\Group("purchase")
 * @Apidoc\Sort(30)
 */
class ErpSupplier extends Controller
{
    /**
     * @Apidoc\Title("供应商列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpSupplier/list")
     * @Apidoc\Param("name", type="string", require=false, desc="供应商名称")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\erp\ErpSupplier\getList")
     */
    public function list()
    {
        $list = (new ErpSupplierModel)->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("供应商详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpSupplier/detail")
     * @Apidoc\Param("erp_supplier_id", type="int", require=true, desc="供应商id")
     * @Apidoc\Returned("detail", type="object", ref="app\shop\model\erp\ErpSupplier\detail")
     */
    public function detail()
    {
        $data = $this->postData();
        $erp_supplier_id = $data['erp_supplier_id'] ?? 0;
        if (empty($erp_supplier_id)) {
            return $this->renderError('供应商ID不能为空');
        }
        $detail = (new ErpSupplierModel)->detail($erp_supplier_id);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        return $this->renderSuccess('', compact('detail'));
    }

    /**
     * @Apidoc\Title("添加供应商")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpSupplier/add")
     * @Apidoc\Param("name", type="string", require=true, desc="供应商名称")
     * @Apidoc\Param("address", type="string", require=true, desc="供应商地址")
     * @Apidoc\Param("contact_person", type="string", require=true, desc="联系人")
     * @Apidoc\Param("contact_phone", type="string", require=true, desc="联系电话")
     * @Apidoc\Param("position", type="string", require=true, desc="职位")
     * @Apidoc\Param("purchaser_id", type="int", require=true, desc="采购员id")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $data = $this->postData();
        $model = new ErpSupplierModel;
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        // 空检验
        if ($data['name'] == '') {
            return $this->renderError('请输入供应商名称');
        }
        if ($data['address'] == '') {
            return $this->renderError('请输入供应商地址');
        }
        if ($data['purchaser_id'] == '') {
            return $this->renderError('请输入采购员ID');
        }
        if (!User::checkUserExist($data['purchaser_id'])) {
            return $this->renderError('采购负责人不存在');
        }
        // 供应商名称限制最大100字符
        if (mb_strlen($data['name']) > 100) {
            return $this->renderError('供应商名称限制最大100字符');
        }
        if (mb_strlen($data['address']) > 100) {
            return $this->renderError('供应商地址限制最大100字符');
        }
        if (mb_strlen($data['contact_person']) > 100) {
            return $this->renderError('联系人限制最大100字符');
        }
        if (mb_strlen($data['contact_phone']) > 50) {
            return $this->renderError('联系电话限制最大100字符');
        }
        if (mb_strlen($data['position']) > 100) {
            return $this->renderError('职位限制最大100字符');
        }
        if ($model?->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model?->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑供应商")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpSupplier/edit")
     * @Apidoc\Param("erp_supplier_id", type="int", require=true, desc="供应商id")
     * @Apidoc\Param("name", type="string", require=true, desc="供应商名称")
     * @Apidoc\Param("address", type="string", require=true, desc="供应商地址")
     * @Apidoc\Param("contact_person", type="string", require=true, desc="联系人")
     * @Apidoc\Param("contact_phone", type="string", require=true, desc="联系电话")
     * @Apidoc\Param("position", type="string", require=true, desc="职位")
     * @Apidoc\Param("purchaser_id", type="int", require=true, desc="采购员id")
     * @Apidoc\Returned()
     */
    public function edit()
    {
        $data = $this->postData();
        $erp_supplier_id = $data['erp_supplier_id'] ?? 0;
        if (empty($erp_supplier_id)) {
            return $this->renderError('供应商ID不能为空');
        }
        // 空检验
        if ($data['name'] == '') {
            return $this->renderError('请输入供应商名称');
        }
        if ($data['address'] == '') {
            return $this->renderError('请输入供应商地址');
        }
        if ($data['purchaser_id'] == '') {
            return $this->renderError('请输入采购员ID');
        }
        if (!User::checkUserExist($data['purchaser_id'])) {
            return $this->renderError('采购负责人不存在');
        }
        // 供应商名称限制最大100字符
        if (mb_strlen($data['name']) > 100) {
            return $this->renderError('供应商名称限制最大100字符');
        }
        if (mb_strlen($data['address']) > 100) {
            return $this->renderError('供应商地址限制最大100字符');
        }
        if (mb_strlen($data['contact_person']) > 100) {
            return $this->renderError('联系人限制最大100字符');
        }
        if (mb_strlen($data['contact_phone']) > 50) {
            return $this->renderError('联系电话限制最大100字符');
        }
        if (mb_strlen($data['position']) > 100) {
            return $this->renderError('职位限制最大100字符');
        }
        $detail = (new ErpSupplierModel)->detail($erp_supplier_id);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail?->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($detail?->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除供应商")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/purchase.ErpSupplier/delete")
     * @Apidoc\Param("erp_supplier_id", type="int", require=true, desc="供应商id")
     * @Apidoc\Returned()
     */
    public function delete($erp_supplier_id)
    {
        $data = $this->postData();
        $erp_supplier_id = $data['erp_supplier_id'] ?? 0;
        if (empty($erp_supplier_id)) {
            return $this->renderError('供应商ID不能为空');
        }
        $detail = (new ErpSupplierModel)->detail($erp_supplier_id);
        if (!$detail) {
            return $this->renderError('数据不存在');
        }
        if ($detail?->del()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($detail?->getError() ?: '删除失败');
    }
}

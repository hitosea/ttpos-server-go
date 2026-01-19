<?php

namespace app\admin\controller;

use help\ImgHelp;
use app\admin\validate\AppValidate;
use hg\apidoc\annotation as Apidoc;
use app\common\model\pay\PaymentApp;
use app\admin\model\app\App as AppModel;
use app\admin\validate\PaymentAppValidate;
use app\common\enum\settings\LanguageEnum;
use app\common\enum\settings\TimeZoneEnum;
use app\common\model\supplier\Supplier as SupplierModel;

/**
 * 商家管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(3)
 */
class Shop extends Controller
{
    /**
     * @Apidoc\Title("商家列表")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/index")
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称/ID")
     * @Apidoc\Param("shop_type", type="int", default=0, desc="商家类型 0全部, 1连锁店, 2单店")
     * @Apidoc\Param("status", type="int", default=0, desc="状态 0全部, 1启用, 2禁用")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\admin\model\app\App\getList")
     */
    public function index()
    {
        $model = new AppModel;
        $list = $model->getList($this->postData())?->toArray();
        foreach ($list['data'] as $key => &$item) {
            $list['data'][$key]['logo'] = ImgHelp::addImageDomain($item['logo'] ?? '');
        }
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("商家下拉选择列表")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/selectList")
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称/ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", children={
     *      @Apidoc\Property("id",type="int", desc="id"),
     *      @Apidoc\Property("name",type="string", desc="商家名称"),
     * })
     */
    public function selectList()
    {
        $model = new AppModel;
        $list = $model->getSelectList($this->postData());
        //
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加商家")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/add")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="商家名称")
     * @Apidoc\Param("level", type="int", require=true, default=1, desc="商家等级: 从1开始")
     * @Apidoc\Param("parent_id", type="int", require=true, default=0, desc="上级商家ID")
     * @Apidoc\Param("store_type", type="int", require=true, default=20, desc="经营模式: 20直营店, 10加盟店")
     * @Apidoc\Param("category_set", type="int", require=true, default="", desc="商品分类设置10同步 主店20分店创建")
     * @Apidoc\Param("logo", type="string", require=true, default="", desc="Logo")
     * @Apidoc\Param("sale_stock", type="int", require=true, default="", desc="进销存: 0不开启, 1开启")
     * @Apidoc\Param("reserve", type="int", require=true, default="", desc="预订: 0不开启, 1开启")
     * @Apidoc\Param("auth_day", type="int", require=true, default="", desc="授权时间(天) 0为永不过期")
     * @Apidoc\Param("auth_start_time", type="string", require=true, default="", desc="授权开始时间 yyyy-mm-dd HH:mm")
     * @Apidoc\Param("link_phone", type="string", require=true, default="", desc="联系电话")
     * @Apidoc\Param("cash_limit", type="int", require=true, default="", desc="收银机上限")
     * @Apidoc\Param("kitchen_limit", type="int", require=true, default="", desc="厨显上限")
     * @Apidoc\Param("tablet_limit", type="int", require=true, default="", desc="平板上限")
     * @Apidoc\Param("user_name", type="string", require=true, default="", desc="超管用户名")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="超管密码")
     * @Apidoc\Param("address", type="string", require=true, default="", desc="所在地")
     * @Apidoc\Param("mac_addr", type="string", require=true, default="", desc="mac地址")
     * @Apidoc\Param("serial_number", type="string", require=true, default="", desc="服务序列号")
     * @Apidoc\Param("deploy_mode", type="string", require=true, default="", desc="部署方式 0局域网部署, 1云部署")
     * @Apidoc\Param("assistant_limit", type="string", require=true, default="", desc="点餐助手上限")
     * @Apidoc\Param("chain_number", type="string", require=true, default="", desc="连锁编号")
     * @Apidoc\Param("is_open_member", type="int", require=true, default="", desc="是否开启会员: 0不开启, 1开启")
     * @Apidoc\Param("is_open_tablet", type="int", require=true, default="", desc="是否开启平板: 0不开启, 1开启")
     * @Apidoc\Param("is_open_scan", type="int", require=true, default="", desc="是否开启扫码H5: 0不开启, 1开启")
     * @Apidoc\Param("is_open_assistant", type="int", require=true, default="", desc="是否开启点餐助手: 0不开启, 1开启")
     * @Apidoc\Param("is_open_kitchen_kds", type="int", require=true, default="", desc="是否开启后厨KDS: 0不开启, 1开启")
     * @Apidoc\Param("is_open_buffet", type="int", require=true, default="", desc="是否开启自助餐: 0不开启, 1开启")
     * @Apidoc\Param("is_accept_scan_order", type="int", require=true, default="", desc="是否开启扫码点餐接单: 0不开启, 1开启（v1.0.9）")
     * @Apidoc\Param("is_open_local_print", type="int", require=true, default="", desc="是否开启本地打印服务: 0不开启, 1开启（v1.1.0）")
     * @Apidoc\Param("is_open_advanced_ticket_print", type="int", require=true, default="", desc="是否开启高级票据打印模板: 0不开启, 1开启（v2.8.0）")
     * @Apidoc\Param("table_limit", type="int", require=true, default="", desc="桌台上限 -1等于无限制")
     * @Apidoc\Param("printer_limit", type="int", require=true, default="", desc="打印机上限 -1等于无限制")
     * @Apidoc\Param("languages", type="array", require=true, default="", desc="支持语言")
     * @Apidoc\Param("timezone", type="string", require=true, default="", desc="时区")
     * @Apidoc\Param("enable_table_map", type="int", require=false, default=0, desc="是否启用桌台地图能力: 0不开启, 1开启")
     * @Apidoc\Param("enable_data_management", type="int", require=false, default=0, desc="是否启用数据管理能力: 0不开启, 1开启")
     * @Apidoc\Param("enable_kiosk", type="int", require=false, default=0, desc="是否启用自助点餐机: 0不开启, 1开启")
     * @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")
     */
    public function add(AppValidate $validate)
    {
        $param = $validate->goCheck('add');
        $param['logo'] = ImgHelp::removeImageDomain($param['logo'] ?? '');
        $param['deploy_mode'] = 1;
        /** @var AppModel $model */
        $model = new AppModel;
        if ($model->add($param)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("语言数据")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/langData")
     * @Apidoc\Param("keyword", type="string", default="", desc="商家名称/ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("language_list", type="array", desc="语言列表",  children={
     *      @Apidoc\Property("key",type="int", desc="id"),
     *      @Apidoc\Property("value",type="string", desc="名称"),
     * })
     * @Apidoc\Returned("time_zone_list", type="array", desc="时区列表", children={
     *      @Apidoc\Property("key",type="int", desc="id"),
     *      @Apidoc\Property("value",type="string", desc="名称"),
     * })
     */
    public function langData()
    {
        $model = new AppModel;
        $language_list = LanguageEnum::data(2);
        $time_zone_list = TimeZoneEnum::data(2);
        return $this->renderSuccess('', compact('language_list', 'time_zone_list'));
    }

    /**
     * @Apidoc\Title("编辑商家")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url("/api/admin/shop/edit")
     * @Apidoc\Param("app_id", type="int", require=true, default="", desc="主键")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="商家名称")
     * @Apidoc\Param("level", type="int", require=true, default=1, desc="商家等级: 从1开始")
     * @Apidoc\Param("parent_id", type="int", require=true, default=0, desc="上级商家ID")
     * @Apidoc\Param("store_type", type="int", require=true, default=20, desc="经营模式: 20直营店, 10加盟店")
     * @Apidoc\Param("category_set", type="int", require=true, default="", desc="商品分类设置10同步 主店20分店创建")
     * @Apidoc\Param("logo", type="string", require=true, default="", desc="Logo")
     * @Apidoc\Param("sale_stock", type="int", require=true, default="", desc="进销存: 0不开启, 1开启")
     * @Apidoc\Param("reserve", type="int", require=true, default="", desc="预订: 0不开启, 1开启")
     * @Apidoc\Param("auth_day", type="int", require=true, default="", desc="授权时间(天) 0为永不过期")
     * @Apidoc\Param("auth_start_time", type="string", require=true, default="", desc="授权开始时间 yyyy-mm-dd HH:mm")
     * @Apidoc\Param("link_phone", type="string", require=true, default="", desc="联系电话")
     * @Apidoc\Param("cash_limit", type="int", require=true, default="", desc="收银机上限")
     * @Apidoc\Param("kitchen_limit", type="int", require=true, default="", desc="厨显上限")
     * @Apidoc\Param("tablet_limit", type="int", require=true, default="", desc="平板上限")
     * @Apidoc\Param("user_name", type="string", require=true, default="", desc="超管用户名")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="超管密码")
     * @Apidoc\Param("address", type="string", require=true, default="", desc="所在地")
     * @Apidoc\Param("mac_addr", type="string", require=true, default="", desc="mac地址")
     * @Apidoc\Param("serial_number", type="string", require=true, default="", desc="服务序列号")
     * @Apidoc\Param("deploy_mode", type="string", require=true, default="", desc="部署方式 0局域网部署, 1云部署")
     * @Apidoc\Param("assistant_limit", type="string", require=true, default="", desc="点餐助手上限")
     * @Apidoc\Param("chain_number", type="string", require=true, default="", desc="连锁编号")
     * @Apidoc\Param("is_open_member", type="int", require=true, default="", desc="是否开启会员: 0不开启, 1开启")
     * @Apidoc\Param("is_open_tablet", type="int", require=true, default="", desc="是否开启平板: 0不开启, 1开启")
     * @Apidoc\Param("is_open_scan", type="int", require=true, default="", desc="是否开启扫码H5: 0不开启, 1开启")
     * @Apidoc\Param("is_open_assistant", type="int", require=true, default="", desc="是否开启点餐助手: 0不开启, 1开启")
     * @Apidoc\Param("is_open_kitchen_kds", type="int", require=true, default="", desc="是否开启后厨KDS: 0不开启, 1开启")
     * @Apidoc\Param("is_open_buffet", type="int", require=true, default="", desc="是否开启自助餐: 0不开启, 1开启")
     * @Apidoc\Param("is_accept_scan_order", type="int", require=true, default="", desc="是否开启扫码点餐接单: 0不开启, 1开启（v1.0.9）")
     * @Apidoc\Param("is_open_local_print", type="int", require=true, default="", desc="是否开启本地打印服务: 0不开启, 1开启（v1.1.0）")
     * @Apidoc\Param("is_open_advanced_ticket_print", type="int", require=true, default="", desc="是否开启高级票据打印模板: 0不开启, 1开启（v2.8.0）")
     * @Apidoc\Param("table_limit", type="int", require=true, default="", desc="桌台上限 -1等于无限制")
     * @Apidoc\Param("printer_limit", type="int", require=true, default="", desc="打印机上限 -1等于无限制")
     * @Apidoc\Param("languages", type="array", require=true, default="", desc="支持语言")
     * @Apidoc\Param("timezone", type="string", require=true, default="", desc="时区")
     * @Apidoc\Param("enable_table_map", type="int", require=false, default=0, desc="是否启用桌台地图能力: 0不开启, 1开启")
     * @Apidoc\Param("enable_data_management", type="int", require=false, default=0, desc="是否启用数据管理能力: 0不开启, 1开启")
     * @Apidoc\Param("enable_kiosk", type="int", require=false, default=0, desc="是否启用自助点餐机: 0不开启, 1开启")
     * @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")
     * @Apidoc\Returned("app_id",type="int", desc="id")
     */
    public function edit()
    {
        if ($this->request->isGet()) {
            $data['app_id'] = $this->request->get('app_id');
            $model = new AppModel;
            $model = $model->getList($data)?->toArray()['data'][0] ?? [];
            return $this->renderSuccess('', compact('model'));
        }
        $validate = new AppValidate();
        $param = $validate->goCheck('edit');
        $param['deploy_mode'] = 1;
        /** @var AppModel $model */
        $model = AppModel::detail($param['app_id']);
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('修改成功', ['app_id' => $param['app_id']]);
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }

    /**
     * @Apidoc\Title("删除商家")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/delete")
     * @Apidoc\Param("app_id", type="int", require=true, default="", desc="主键")
     */
    public function delete(AppValidate $validate)
    {
        $param = $validate->goCheck('id');
        /** @var AppModel $model */
        $model = AppModel::detail($param['app_id']);
        if (!$model->setDelete()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("启用禁用状态")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/updateStatus")
     * @Apidoc\Param("app_id", type="int", require=true, default="", desc="主键")
     */
    public function updateStatus(AppValidate $validate)
    {
        $param = $validate->goCheck('id');
        /** @var AppModel $model */
        $model = AppModel::detail($param['app_id']);
        if (!$model->updateStatus()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("启用禁用进销存")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/updateSaleStock")
     * @Apidoc\Param("app_id", type="int", require=true, default="", desc="主键")
     */
    public function updateSaleStock(AppValidate $validate)
    {
        $param = $validate->goCheck('id');
        /** @var SupplierModel $model */
        $model = SupplierModel::where('company_uuid', $param['app_id'])->find();
        if (!$model->updateSaleStock()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("启用禁用预定")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/updateReserve")
     * @Apidoc\Param("app_id", type="int", require=true, default="", desc="主键")
     */
    public function updateReserve(AppValidate $validate)
    {
        $param = $validate->goCheck('id');
        /** @var SupplierModel $model */
        $model = SupplierModel::where('company_uuid', $param['app_id'])->find();
        if (!$model->updateReserve()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("商家支付设置")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/shop/payment")
     * @Apidoc\Param("shop_supplier_id", type="int", require=true,  desc="商家ID")
     * @Apidoc\Param("app_id", type="int", require=true,  desc="应用ID")
     * @Apidoc\Param("ll_merchant_id", type="string", require=true,  desc="商户号")
     * @Apidoc\Param("ll_public_key", type="string", require=true,  desc="LianLianpay公钥")
     * @Apidoc\Param("ll_merchant_private_key", type="string", require=true,  desc="商户私钥")
     * @Apidoc\Param("ll_token", type="string", require=true,  desc="商户Token")
     * @Apidoc\Param("ll_store_id", type="string", require=true,  desc="商户站点ID")
     */
    public function payment(PaymentAppValidate $validate)
    {
        $model = new PaymentApp;
        if ($this->request->isGet()) {
            $param = $validate->goCheck('id');
            $detail = $model::detail($param['shop_supplier_id']);
            return $this->renderSuccess('操作成功', $detail);
        }
        $param = $validate->goCheck('add');
        if ($model->payment($param)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }
}

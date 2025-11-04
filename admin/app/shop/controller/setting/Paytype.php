<?php

namespace app\shop\controller\setting;

use help\ImgHelp;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\model\store\PayType as PayTypeModel;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 支付管理
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(10)
 */
class Paytype extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.Paytype/index")
     * @Apidoc\Returned("", type="array", ref="app\common\model\store\PayType\lists", desc="列表")
     */
    public function index()
    {
        $app_id = $this->store['app']['app_id'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        $list = PayTypeModel::list($shop_supplier_id, $app_id);
        foreach ($list as $key => $item) {
            $item['img'] = $item['logoFile'] ? $item['logoFile']['file_path'] : $item['default_img'];
            $item['qrcode'] = $item['qrcodeFile'] ? $item['qrcodeFile']['file_path'] : '';
            $list[$key] = $item;
        }
        // 支付方式
        $payment = SettingModel::getSupplierItem(SettingEnum::PAYMENT, $shop_supplier_id);
        return $this->renderSuccess('', compact('list', 'payment'));
    }

    /**
     * @Apidoc\Title("原始支付列表（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.Paytype/defaultPay")
     * @Apidoc\Returned("", type="array", ref="app\common\model\store\PayType\lists", desc="列表")
     */
    public function defaultPay()
    {
        $app_id = $this->store['app']['app_id'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        $list =  array_values(array_filter(OrderPayTypeEnum::data(), function ($item) {
            switch ($item['value']) {
                case OrderPayTypeEnum::BALANCE:
                    return false;
                case OrderPayTypeEnum::CASH:
                    return false;
                case OrderPayTypeEnum::WECHAT:
                    return false;
                case OrderPayTypeEnum::ALIPAY:
                    return false;
                case OrderPayTypeEnum::POS:
                    return false;
                case OrderPayTypeEnum::FREE_PAY:
                    return false;
            }
            return true;
        }));
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.paytype/add")
     * @Apidoc\Param("add_method", type="int", require=true, default=1, desc="添加方式（v1.0.8） 0-系统默认 1-直接添加（对象） 2-快捷添加（数组对象）")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="支付方式")
     * @Apidoc\Param("remark", type="string", require=true, default="", desc="名称")
     * @Apidoc\Param("is_show_checkout", type="array", require=false, default="", desc="结账显示 10-收银机 20-点餐助手（v1.0.9）")
     * @Apidoc\Param("is_show_recharge", type="array", require=false, default="", desc="充值显示 10-收银机 20-点餐助手（v1.0.9）")
     * @Apidoc\Param("fee", type="string", require=true, default="", desc="支付手续费0-100")
     * @Apidoc\Param("status", type="int", require=true, default=1, desc="状态 0禁用 1启用")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Param("img", type="string", require=true, default="", desc="图片")
     * @Apidoc\Param("qrcode", type="string", require=true, default=1, desc="二维码")
     * @Apidoc\Param("params", type="array", require=true, default="", desc="数组对象参数，快捷添加时使用（v1.0.8）")
     * @Apidoc\Returned("", type="array", ref="app\common\model\store\PayType\lists", desc="列表")
     */
    public function add()
    {
        $param = $this->postData();
        $param['app_id'] = $this->store['app']['app_id'];
        $param['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];

        // 检查添加方式
        $addMethod = $param['add_method'] ?? 1;
        if (!in_array($addMethod, [1, 2])) {
            return $this->renderError('添加方式不正确');
        }

        $model = new PayTypeModel;

        // 处理直接添加
        if ($addMethod == 1) {
            if (!$param = $model->validateData($param)) {
                return $this->renderError($model->getError() ?: '数据验证失败');
            }
            if ($model->save($param)) {
                return $this->renderSuccess('操作成功');
            }
            return $this->renderError($model->getError() ?: '操作失败');
        }

        // 处理快捷添加
        if ($addMethod == 2) {
            $items = $param['params'] ?? [];
            if (!is_array($items) || empty($items)) {
                return $this->renderError('参数错误');
            }
            // 验证所有数据
            $validatedItems = [];
            foreach ($items as $item) {
                $item['app_id'] = $param['app_id'];
                $item['shop_supplier_id'] = $param['shop_supplier_id'];
                if ($callback_item = $model->validateData($item)) {
                    $validatedItems[] = $callback_item;
                } else {
                    return $this->renderError($model->getError() ?: '数据验证失败');
                }
            }
            // 兼容快捷添加多个code一致
            foreach ($validatedItems as $key => $item) {
                $validatedItems[$key]['code'] = $item['code'] + $key * 100;
                if ($item['logo_file_uuid'] == 0) {
                    $validatedItems[$key]['default_img'] = $item['logo_url'];
                }
            }
            if ($model->saveAll($validatedItems)) {
                return $this->renderSuccess('操作成功');
            }
            return $this->renderError($model->getError() ?: '操作失败');
        }
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.paytype/edit")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="支付方式的ID")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="支付方式")
     * @Apidoc\Param("remark", type="string", require=true, default="", desc="名称")
     * @Apidoc\Param("is_show_checkout", type="array", require=false, default="", desc="结账显示 10-收银机 20-点餐助手（v1.0.9）")
     * @Apidoc\Param("is_show_recharge", type="array", require=false, default="", desc="充值显示 10-收银机 20-点餐助手（v1.0.9）")
     * @Apidoc\Param("fee", type="string", require=true, default="", desc="支付手续费0-100")
     * @Apidoc\Param("status", type="int", require=true, default=1, desc="状态 0禁用 1启用")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Param("img", type="string", require=true, default=1, desc="图片")
     * @Apidoc\Param("qrcode", type="string", require=true, default=1, desc="二维码")
     * @Apidoc\Returned("", type="array", ref="app\common\model\store\PayType\lists", desc="列表")
     */
    public function edit()
    {
        $param = $this->postData();
        $app_id = $this->store['app']['app_id'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        //
        $id = $param['id'] ?? '';
        $remark = $param['remark'] ?? '';
        if (!$id || !$orderPayType = PayTypeModel::where('id', $id)->find()) {
            return $this->renderError('无效的参数');
        }
        // 来源判断
        if ($orderPayType['source'] == PayTypeModel::SOURCE_DEFAULT) {
            if (!$remark) {
                return $this->renderError('请输入名称');
            }
            if (!($param['name'] ?? '')) {
                return $this->renderError('请输入支付方式');
            }
            if (!($param['img'] ?? '')) {
                return $this->renderError('请输入图片');
            }
        }
        $param['fee'] = $param['fee'] ?? 0;
        if ($param['fee'] < 0 || $param['fee'] > 100) {
            return $this->renderError('支付手续费范围0-100');
        }
        $sort = $param['sort'] ?? '';
        if ($sort == '' || !is_numeric($sort)) {
            return $this->renderError('请输入排序');
        }
        // 至少需要开启一种支付方式
        if (isset($param['status'])) {
            if (!in_array($param['status'], [0, 1])) {
                return $this->renderError('无效的status');
            }
            $list = PayTypeModel::list($shop_supplier_id, $app_id);
            $openCount = count(array_filter($list, function ($item) {
                return isset($item['status']) && $item['status'] == 1;
            }));
            // 支付方式
            $payment = SettingModel::getSupplierItem(SettingEnum::PAYMENT, $shop_supplier_id);
            $is_cash = $payment['is_cash'] ?? 0;
            $is_balance = $payment['is_balance'] ?? 0;
            if ($openCount == 1 && $param['status'] == 0 && $is_cash == 0 && $is_balance == 0) {
                return $this->renderError('至少需要开启一种支付方式');
            }
        }
        //
        $saveData = [
            'status' => $param['status'],
            'payment_name' => $remark,
            'fee_percent' => $param['fee'],
            'sort' => $sort,
            'fee' => $param['fee'],
            'name' => $param['name'],
        ];
        if (isset($param['is_show_checkout'])) {
            $isShowCheckout = $param['is_show_checkout'] ?: [];
            $saveData['is_show_cashier'] = in_array(PayTypeModel::CASHIER_SHOW_VALUE, $isShowCheckout) ? 1 : 0;
            $saveData['is_show_assistant'] = in_array(PayTypeModel::ASSISTANT_SHOW_VALUE, $isShowCheckout) ? 1 : 0;
        }
        if (isset($param['is_show_recharge'])) {
            $isShowRecharge = $param['is_show_recharge'] ?: [];
            $saveData['is_show_member_recharge'] = in_array(PayTypeModel::CASHIER_SHOW_VALUE, $isShowRecharge) ? 1 : 0;
        }
        if (isset($param['img']['file_id'])) {
            $saveData['logo_file_uuid'] = $param['img']['file_id'];
        }
        if (isset($param['img']['file_path'])) {
            $saveData['default_img'] = ImgHelp::removeImageDomain($param['img']['file_path']);
        }
        if (isset($param['qrcode']['file_id'])) {
            $saveData['qrcode_file_uuid'] = $param['qrcode']['file_id'];
        }
        if (isset($param['qrcode']['file_path'])) {
            $saveData['qrcode_url'] = ImgHelp::removeImageDomain($param['qrcode']['file_path']);
        }
        $orderPayType->save($saveData);
        //
        return $this->renderSuccess('操作成功', $orderPayType);
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.paytype/delete")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="支付方式的ID")
     */
    public function delete()
    {
        $param = $this->postData();
        $app_id = $this->store['app']['app_id'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        //
        if (!($param['id'] ?? '') || !$orderPayType = PayTypeModel::where('id', $param['id'])->find()) {
            return $this->renderError('无效的参数');
        }
        // 来源判断
        if ($orderPayType['source'] != PayTypeModel::SOURCE_DEFAULT) {
            return $this->renderError('该支付方式不允许删除');
        }
        // 至少需要开启一种支付方式
        $list = PayTypeModel::list($shop_supplier_id, $app_id);
        $openCount = count(array_filter($list, function ($item) {
            return isset($item['status']) && $item['status'] == 1;
        }));
        // 支付方式
        $payment = SettingModel::getSupplierItem(SettingEnum::PAYMENT, $shop_supplier_id);
        $is_cash = $payment['is_cash'] ?? 0;
        $is_balance = $payment['is_balance'] ?? 0;
        if ($openCount == 1 && $orderPayType['status'] == 1 && $is_cash == 0 && $is_balance == 0) {
            return $this->renderError('至少需要开启一种支付方式');
        }
        //
        $orderPayType->delete();
        //
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("支付方式设置")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/setting.paytype/editPayment")
     * @Apidoc\Param("is_cash", type="int", require=true, default=0, desc="是否开启现金支付 0-关闭 1-开启")
     * @Apidoc\Param("is_balance", type="int", require=true, default=0, desc="是否开启余额支付 0-关闭 1-开启")
     * @Apidoc\Param("is_other", type="int", require=true, default=0, desc="是否开启其他方式支付 0-关闭 1-开启")
     */
    public function editPayment()
    {
        $model = new SettingModel;
        $data = $this->request->param();
        $licenses = $this->request->licenses;
        $app_id = $this->store['app']['app_id'];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];

        $arr = [
            'is_cash' => $data['is_cash'] ?? 0,
            'is_balance' => $data['is_balance'] ?? 0,
            'is_other' => $data['is_other'] ?? 0,
        ];

        // 至少需要开启一种支付方式
        $list = PayTypeModel::list($shop_supplier_id, $app_id);
        $openCount = count(array_filter($list, function ($item) {
            return isset($item['status']) && $item['status'] == 1;
        }));

        // 检查是否至少有一种支付方式是开启的
        $is_cash = $arr['is_cash'] ?? 0;
        $is_balance = $arr['is_balance'] ?? 0;
        $is_other = $arr['is_other'] ?? 0;
        if ($is_cash == 0 && $is_balance == 0 && ($is_other == 0 || ($is_other == 1 && $openCount == 0))) {
            return $this->renderError('至少需要开启一种支付方式');
        }

        // 总权限 - 不开启会员
        if (($licenses['is_open_member'] ?? 0) == 0 && $arr['is_balance'] == 1) {
            return $this->renderError('未开启会员功能，不能使用余额');
        }

        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::PAYMENT, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }
}

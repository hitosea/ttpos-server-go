<?php

namespace app\shop\controller\setting;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\tax\TaxCategory;
use app\common\model\product\ProductTax;
use app\common\enum\settings\SettingEnum;
use app\common\model\buffet\Buffet;
use app\common\model\product\Product;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 相关设置
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(5)
 */
class Supplier extends Controller
{
    /**
     * @Apidoc\Title("货币单位(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Supplier/currencyUnit")
     * @Apidoc\Param("unit", type="string", require=true, default="", desc="主货币单位")
     * @Apidoc\Param("print_unit", type="string", require=true, default="", desc="主货币单位 - 打印专用（v1.0.6）")
     * @Apidoc\Param("unit_position", type="int", require=true, default=0, desc="主货币显示位置 0-金额前 1-金额后")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启副货币单位")
     * @Apidoc\Param("vice_unit", type="string", require=true, default="", desc="副货币单位")
     * @Apidoc\Param("vice_unit_position", type="int", require=true, default=0, desc="副货币显示位置 0-金额前 1-金额后")
     * @Apidoc\Param("unit_rate", type="string", require=true, default="", desc="主副货币转换比例")
     * @Apidoc\Returned()
     */
    public function currencyUnit()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::CURRENCY);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $unit_rate = $data['unit_rate'] ?? '';
        if (empty($data['unit'])) {
            return $this->renderError('主货币单位不能为空');
        }
        if (empty($data['print_unit'])) {
            return $this->renderError('打印主货币单位不能为空');
        }
        if (!empty($data['is_open']) && (empty($data['vice_unit']) || empty($unit_rate))) {
            return $this->renderError('副货币单位和主副货币汇率不能为空');
        }
        if (!empty($data['is_open']) && !is_numeric($unit_rate)) {
            return $this->renderError('主副货币汇率必须为数字');
        }

        $arr = [
            'unit' => $data['unit'], // 主货币单位
            'print_unit' => $data['print_unit'], // 主货币单位 - 打印专用
            // 主货币显示位置 0-金额前 1-金额后
            'unit_position' => $data['unit_position'] ?? 0,
            'is_open' => $data['is_open'], // 是否开启副货币单位
        ];
        if (!empty($data['is_open'])) {
            // 副货币单位
            $arr['vice_unit'] = $data['vice_unit'];
            // 副货币显示位置 0-金额前 1-金额后
            $arr['vice_unit_position'] = $data['vice_unit_position'] ?? 0;
            // 主副货币转换比例
            $arr['unit_rate'] = max(0, $unit_rate);
        } else {
            $old = SettingModel::getSupplierItem(SettingEnum::CURRENCY, $this->store['user']['shop_supplier_id']);
            $arr['vice_unit'] = $old['vice_unit'] ?? '';
            $arr['vice_unit_position'] = $old['vice_unit_position'] ?? 0;
            $arr['unit_rate'] = $old['unit_rate'] ?? 0;
        }
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::CURRENCY, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("税率管理(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Supplier/taxRate")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启副货币单位")
     * @Apidoc\Param("tax_rate", type="string", require=true, default="", desc="税率（v1.0.4.1已废除）")
     * @Apidoc\Param("calc_type", type="string", require=true, default="", desc="计算类型 商品已含税价-1 商品未含税价-2")
     * @Apidoc\Param("add_tax_category", type="array", require=true, desc="税类 名称 - 税率", children={
     *     @Apidoc\Returned("id", type="int", desc="税类id"),
     *     @Apidoc\Returned("name", type="string", desc="名称"),
     *     @Apidoc\Returned("tax_rate", type="string", desc="税率"),
     *     @Apidoc\Returned("action", type="string", desc="操作结果 delete-删除 edit-编辑 add-新增"),
     * })
     * @Apidoc\Returned()
     */
    public function taxRate()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::TAX_RATE);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $is_open = $data['is_open'] ?? 0;
        // 计算类型
        if ($is_open && !in_array($data['calc_type'], ['1', '2'])) {
            return $this->renderError('计算类型参数错误');
        }
        // 税类
        if ($is_open) {
            if (!isset($data['add_tax_category']) || empty($data['add_tax_category'])) {
                return $this->renderError('税类设置不能为空');
            }
            foreach ($data['add_tax_category'] as $item) {
                $taxCategoryModel = new TaxCategory;
                $name = $item['name'] ?? '';
                $taxRate = $item['tax_rate'] ?? '';
                if ($name == '') {
                    return $this->renderError('请输入税率名称');
                }
                if (mb_strlen($name) > 50) {
                    return $this->renderError('税率名称不能超过50字符');
                }
                $taxRate = round($taxRate, 2);
                $taxRate = max(0, $taxRate);
                if ($taxRate < 0 || $taxRate > 100) {
                    return $this->renderError('请输入0-100的税率');
                }
                // 如果是新增
                if (isset($item['action']) && $item['action'] == 'add') {
                    $item['app_id'] = $this->store['app']['app_id'];
                    unset($item['action']);
                    $item['uuid'] = createUuid();
                    $taxCategoryModel->save($item);
                    continue;
                }
                // 如果是删除 - 保留数据，只是修改状态
                if (isset($item['action']) && $item['action'] == 'delete') {
                    if (isset($item['id'])) {
                        $productUse = Product::where('dine_tax_uuid', $item['id'])
                            ->whereOr('takeout_tax_uuid', $item['id'])
                            ->count();
                        if ($productUse > 0) {
                            return $this->renderError('税类正在使用，不能删除');
                        }
                        $buffetUse = Buffet::where('tax_uuid', $item['id'])->count();
                        if ($buffetUse > 0) {
                            return $this->renderError('税类正在使用，不能删除');
                        }
                        $taxCategoryModel->destroy(['uuid' => $item['id']]);
                    }
                    continue;
                }
                // 如果是编辑
                if (isset($item['action']) && $item['action'] == 'edit') {
                    if (isset($item['id'])) {
                        unset($item['action']);
                        $taxCategoryModel->update([
                            'name' => $item['name'],
                            'tax_rate' => $item['tax_rate'],
                        ], ['uuid' => $item['id']]);
                    }
                    continue;
                }
            }
        }
        // 如果开启后之前没有过税类的商品、自助餐默认选择第一个税类
        if ($is_open) {
            (new ProductTax)->getProductDefaultTaxCategory();
        }
        $arr = [
            'is_open' => $data['is_open'], // 是否开启税率
            'calc_type' => $data['calc_type'], // 计算类型 商品已含税价-1 商品未含税价-2
            'add_tax_category' => [],
        ];
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::TAX_RATE, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("服务费(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Supplier/serviceCharge")
     * @Apidoc\Param("is_open", type="int", require=true, default=0, desc="是否开启副货币单位")
     * @Apidoc\Param("charge_type", type="string", require=true, default="", desc="服务费类型 1-固定金额 2-百分比")
     * @Apidoc\Param("service_charge", type="string", require=true, default="", desc="服务费金额")
     * @Apidoc\Param("service_charge_rate", type="string", require=true, default="", desc="服务费率")
     * @Apidoc\Param("is_open_tax", type="string", require=true, default="", desc="税费 1-收取税费 0-不收取税费")
     * @Apidoc\Returned()
     */
    public function serviceCharge()
    {
        if ($this->request->isGet()) {
            return $this->fetchData(SettingEnum::SERVICE_CHARGE);
        }
        $model = new SettingModel;
        $data = $this->request->param();
        //
        $charge_type = $data['charge_type'] ?? '';
        $service_charge = $data['service_charge'] ?? 0;
        $service_charge_rate = $data['service_charge_rate'] ?? 0;
        $is_open_tax = $data['is_open_tax'] ?? '';
        $apply_scope = $data['apply_scope'] ?? '1';
        $apply_scope_ordering = $data['apply_scope_ordering'] ?? '0';
        $apply_scope_table = $data['apply_scope_table'] ?? '0';
        $apply_scope_table_list = $data['apply_scope_table_list'] ?? [];
        $service_fee_base = $data['service_fee_base'] ?? '0';
        if (!in_array($charge_type, ['1', '2'])) {
            return $this->renderError('服务费类型错误');
        }
        if (!empty($data['is_open']) && $charge_type == '1' && $service_charge === '') {
            return $this->renderError('服务费不能为空');
        }
        if (!empty($data['is_open']) && $charge_type == '1' && !is_numeric($service_charge)) {
            return $this->renderError('服务费必须为数字');
        }
        if (!empty($data['is_open']) && $charge_type == '1' && ($service_charge < 0 || $service_charge > 1000000)) {
            return $this->renderError('请输入0-1000000的服务费');
        }
        if ($charge_type == '1' && strpos($service_charge, '.') !== false && strlen(substr(strrchr($service_charge, "."), 1)) > 2) {
            return $this->renderError('小数部分不能超2位');
        }
        if (!empty($data['is_open']) && $charge_type == '2' && $data['service_charge_rate'] === '') {
            return $this->renderError('服务费率不能为空');
        }
        if (!empty($data['is_open']) && $charge_type == '2' && !is_numeric($service_charge_rate)) {
            return $this->renderError('服务费率必须为数字');
        }
        if (!empty($data['is_open']) && $charge_type == '2' && ($service_charge_rate < 0 || $service_charge_rate > 100)) {
            return $this->renderError('请输入0-100的服务费率');
        }
        if (!empty($data['is_open']) && $charge_type == '2' && (strpos($service_charge_rate, '.') !== false && strlen(substr(strrchr($service_charge_rate, "."), 1)) > 2)) {
            return $this->renderError('小数部分不能超2位');
        }
        if (!in_array($is_open_tax, ['1', '0'])) {
            return $this->renderError('税费参数错误');
        }

        if ($apply_scope == '2') {
            if ($apply_scope_ordering == '0' && $apply_scope_table == '0') {
                return $this->renderError('请选择服务费应用范围');
            }
            if ($apply_scope_table == '1' && empty($apply_scope_table_list)) {
                return $this->renderError('请选择应用服务费的桌台');
            }
        }

        // 台位id列表, 字符串转数字
        $apply_scope_table_list = array_map(function ($item) {
            return intval($item);
        }, $apply_scope_table_list);

        $arr = [
            'is_open' => $data['is_open'], // 是否开启服务费
            'charge_type' => $charge_type, // 服务费类型 1-固定金额 2-百分比
            'apply_scope' => $apply_scope, // 适用范围 1-全部 2-部分
            'apply_scope_ordering' => $apply_scope_ordering, // 适用范围-点餐 0-关闭 1-开启
            'apply_scope_table' => $apply_scope_table, // 适用范围-桌台 0-关闭 1-开启
            'apply_scope_table_list' => $apply_scope_table_list, // 台位uuid列表
            'service_fee_base' => "$service_fee_base", // 服务费基础 "0"-商品惠后价 "1"-商品价格合计
        ];

        if ($charge_type == '1') {
            $arr['service_charge'] = max(0, $service_charge). ""; // 服务费
        } else if ($charge_type == '2') {
            $arr['service_charge_rate'] = max(0, $service_charge_rate) . ""; // 服务费率
            $arr['is_open_tax'] = $is_open_tax; // 税费 1-收取税费 0-不收取税费
        }

        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        if ($model->edit(SettingEnum::SERVICE_CHARGE, $arr, $shop_supplier_id)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * 获取配置值
     */
    public function fetchData($key)
    {
        if ($key == '') {
            return $this->renderError('缺少参数');
        }
        $ret = SettingModel::getSupplierItem($key, $this->store['user']['shop_supplier_id']);
        if ($key == SettingEnum::TAX_RATE) {
            // 获取税类列表
            $taxCategoryList = (new TaxCategory)->getList();
            foreach ($taxCategoryList as $item) {
                $item['id'] = $item['uuid'];
            }
            $ret['add_tax_category'] = $taxCategoryList;
        }
        $vars['values'] = $ret;
        return $this->renderSuccess('', compact('vars'));
    }
}

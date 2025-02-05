<?php

namespace app\shop\controller\setting;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\shop\model\settings\Printer as PrinterModel;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 打印设置
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(6)
 */
class Printing extends Controller
{
    /**
     * @Apidoc\Title("打印设置(get-获取/post-设置)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printing/index")
     * @Apidoc\Param("room_open", type="int", require=true, default="", desc="是否开启打印 0-关闭 1-开启")
     * @Apidoc\Param("room_printer_id", type="int", require=true, default="", desc="打印机id")
     * @Apidoc\Param("buyer_open", type="int", require=true, default="", desc="顾客是否开启打印 0-关闭 1-开启")
     * @Apidoc\Param("buyer_printer_id", type="int", require=true, default="", desc="顾客打印机id")
     * @Apidoc\Param("seller_open", type="int", require=true, default="", desc="商家是否开启打印 0-关闭 1-开启")
     * @Apidoc\Param("seller_printer_id", type="int", require=true, default="", desc="商家打印机id")
     * @Apidoc\Param("language_method", type="int", require=true, default="", desc="语言方式（收银） 1-单语言 2-多语言（v1.0.8）")
     * @Apidoc\Param("default_language", type="int", require=true, default="", desc="打印语言（收银）")
     * @Apidoc\Param("print_method", type="int", require=true, default="", desc="打印方式（收银） 1-文本打印 2-图片打印（v1.0.7）")
     * @Apidoc\Param("kitchen_language", type="int", require=true, default="", desc="v1.0.4.1打印语言（送厨）")
     * @Apidoc\Param("kitchen_print_method", type="int", require=true, default="", desc="打印方式（送厨） 1-文本打印 2-图片打印（v1.0.7）")
     * @Apidoc\Param("consumption_tax", type="int", require=true, default="", desc="v1.0.4.1消费税 1显示全部类型 2仅显示商品已含税 3仅显示商品未含税 4全部不显示")
     * @Apidoc\Param("calendar_list", type="array", require=true, default="", desc="日历列表（v1.0.5）", children={
     *   @Apidoc\Param("key", type="int", require=true, default="", desc="键"),
     *   @Apidoc\Param("name", type="string", require=true, default="", desc="名称"),
     * })
     * @Apidoc\Param("default_calendar", type="int", require=true, default="", desc="默认日历（v1.0.5）")
     * @Apidoc\Returned()
     */
    public function index()
    {
        if ($this->request->isGet()) {
            return $this->fetchData();
        }

        $postData = $this->postData();
        $printerSettings = SettingModel::getSupplierItem(SettingEnum::PRINTER, $this->store['user']['shop_supplier_id']);
        // 收银机设置
        if (isset($postData['cashier_open']) && isset($postData['cashier_printer'])) {
            $printerSettings['cashier_open'] = $postData['cashier_open'];
            $printerSettings['cashier_printer'] = $postData['cashier_printer'];
        }
        // 打印语言（收银）
        if (isset($postData['language_method'])) {
            $printerSettings['language_method'] = $postData['language_method'] ?? 1;
        }
        if (isset($postData['default_language'])) {
            $printerSettings['default_language'] = $postData['default_language'];
            $printerSettings['print_method'] = $postData['print_method'] ?? 1;
        }
        // 打印语言（送厨）
        if (isset($postData['kitchen_language'])) {
            $printerSettings['kitchen_language'] = $postData['kitchen_language'];
            $printerSettings['kitchen_print_method'] = $postData['kitchen_print_method'] ?? 1;
        }
        // 消费税
        if (isset($postData['consumption_tax'])) {
            $printerSettings['consumption_tax'] = $postData['consumption_tax'];
        }
        // 自助餐标识
        if (isset($postData['buffet_sign_open'])) {
            $printerSettings['buffet_sign_open'] = $postData['buffet_sign_open'];
        }
        if (isset($postData['monetary_unit_open'])) {
            $printerSettings['monetary_unit_open'] = $postData['monetary_unit_open'];
        }
        // 默认日历
        if (isset($postData['default_calendar'])) {
            $printerSettings['default_calendar'] = $postData['default_calendar'];
        }
        if (isset($printerSettings['language_list'])) {
            unset($printerSettings['language_list']);
        }
        if (isset($printerSettings['calendar_list'])) {
            unset($printerSettings['calendar_list']);
        }

        $model = new SettingModel;
        if ($model->edit(SettingEnum::PRINTER, $printerSettings, $this->store['user']['shop_supplier_id'])) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * 获取打印配置
     */
    public function fetchData()
    {
        //
        $vars['printerList'] = PrinterModel::getAll($this->store['user']['shop_supplier_id']);
        //
        $vars['cashierList'] = BindRecord::alias('a')->where('source', BindRecord::SOURCE_CASHIER)
            // 云部署方式 - 不显示网页端
            ->when(env('IS_CLOUD_DEPLOY', false), function ($q) {
                $q->where('platform', 'Android');
                // ->whereIn('brand', BindRecord::BRANDS_ALL)
            })
            ->field("a.key as cashier_key")
            ->field("CONCAT(if(remark='', '-', remark), ' (', a.key, ')') cashier_name")
            ->order('id')
            ->select()
            ->toArray();
        //
        $vars['values'] = SettingModel::getSupplierItem(SettingEnum::PRINTER, $this->store['user']['shop_supplier_id']);
        //
        return $this->renderSuccess('', $vars);
    }
}

<?php

namespace app\shop\controller\setting;

use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\BindRecord;
use app\common\service\websocket\Websocket;
use app\common\model\settings\LanPrinterScan;
use app\shop\model\settings\Printer as PrinterModel;


/**
 * 打印机管理
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(6)
 */
class Printer extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printer/index")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\settings\Printer\getList", desc="打印机列表")
     */
    public function index()
    {
        $model = new PrinterModel;
        $list = $model->getList($this->postData(), $this->store['user']['shop_supplier_id']);
        // 推送配置更新
        Websocket::pushClient(
            request()->appId, 
            Websocket::SOURCE_CASHIER, 
            Websocket::SOURCE_All, 
            Websocket::GET_LAN_PRINTER, 
            0,
            []
        );
        // 
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * 打印机类型
     */
    public function type()
    {
        $model = new PrinterModel;
        $printerType = $model::getPrinterTypeList();
        return $this->renderSuccess('', compact('printerType'));
    }

    /**
     * @Apidoc\Title("添加（get-获取类型/post-添加）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printer/add")
     * @Apidoc\Param("printer_name", type="string", require=true, default="", desc="打印机名称")
     * @Apidoc\Param("printer_type", type="string", require=true, default="", desc="打印机类型")
     * @Apidoc\Param("printer_config", type="string", require=true, default="", desc="打印机配，json格式")
     * @Apidoc\Param("print_times", type="int", require=true, default=0, desc="打印联数")
     * @Apidoc\Param("width", type="int", require=false, default=80, desc="纸张宽度（mm）")
     * @Apidoc\Param("enable_status_check", type="int", require=false, default=0, desc="是否启用状态检查 0-关闭 1-开启")
     * @Apidoc\Param("enable_sound", type="int", require=false, default=1, desc="是否启用打印提示音 0-关闭 1-开启")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Returned()
     */
    public function add()
    {
        if ($this->request->isGet()) {
            $cashierList = BindRecord::getCashierList();
            $printerType = PrinterModel::getPrinterTypeList();
            $lanPrinter = LanPrinterScan::field('ip', 'label')->where('status', 1)->select();
            $lanPrinter = array_map(function($item) {
                return [
                    'value' => $item['ip'],
                    'label' => $item['ip'] . ' (' . __('来自收银机扫描') . ')',
                ];
            }, $lanPrinter->toArray());
            return  $this->renderSuccess('', compact('cashierList', 'printerType', 'lanPrinter'));
        }
        // 新增记录
        $model = new PrinterModel;
        $data = $this->postData();
        //
        if (empty($data['printer_name'])) {
            return $this->renderError('请输入打印机名称');
        }
        if (mb_strlen($data['printer_name'] ?? '') > 50) {
            return $this->renderError('名称限制输入50字符');
        }
        if (empty($data['printer_type'])) {
            return $this->renderError('请选择打印机类型');
        }
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        if ($model->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("打印机详情")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printer/detail")
     * @Apidoc\Param("printer_id", type="int", require=true, default=0, desc="打印机id")
     * @Apidoc\Returned("detail", type="array", ref="app\shop\model\settings\Printer\detail", desc="打印机详情")
     * @Apidoc\Returned("printerType", type="array", ref="app\shop\model\settings\Printer\getPrinterTypeList", desc="打印机类型")
     */
    public function detail($printer_id)
    {
        $detail = PrinterModel::detail($printer_id);
        $detail['printer_type'] = [
            'text' => $detail['printerType']['name_text'] ?? '',
            'value' => $detail['printerType']['key'] ?? '',
        ];
        $printerType = PrinterModel::getPrinterTypeList();
        $cashierList = BindRecord::getCashierList();
        return $this->renderSuccess('', compact('detail', 'printerType', 'cashierList'));
    }

    /**
     * @Apidoc\Title("编辑（get-获取类型/post-编辑）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printer/edit")
     * @Apidoc\Param("printer_id", type="int", require=true, default=0, desc="打印机id")
     * @Apidoc\Param("printer_name", type="string", require=true, default="", desc="打印机名称")
     * @Apidoc\Param("printer_type", type="string", require=true, default="", desc="打印机类型")
     * @Apidoc\Param("printer_config", type="string", require=true, default="", desc="打印机配，json格式")
     * @Apidoc\Param("print_times", type="int", require=true, default=0, desc="打印联数")
     * @Apidoc\Param("width", type="int", require=false, default=80, desc="纸张宽度（mm）")
     * @Apidoc\Param("enable_status_check", type="int", require=false, default=0, desc="是否启用状态检查 0-关闭 1-开启")
     * @Apidoc\Param("enable_sound", type="int", require=false, default=1, desc="是否启用打印提示音 0-关闭 1-开启")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Returned()
     */
    public function edit($printer_id)
    {
        if ($this->request->isGet()) {
            return $this->detail($printer_id);
        }
        /** @var PrinterModel $model */
        $model = PrinterModel::detail($printer_id);
        // 更新记录
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除记录")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.Printer/delete")
     * @Apidoc\Param("printer_id", type="int", require=true, default=0, desc="打印机id")
     * @Apidoc\Returned()
     */
    public function delete($printer_id)
    {
        /** @var PrinterModel $model */
        $model = PrinterModel::detail($printer_id);
        if ($model->setDelete()) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }
}

<?php

namespace app\scan\controller\table;

use hg\apidoc\annotation as Apidoc;
use app\scan\controller\Controller;
use app\common\enum\settings\SettingEnum;
use app\scan\model\order\Order as OrderModel;
use app\scan\model\store\Table as TableModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 桌台相关
 */
class Table extends Controller
{

    /**
     * @Apidoc\Title("获取桌台信息")
     * @Apidoc\Desc("获取桌台信息")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/scan/table.table/getInfo")
     */
    public function getInfo()
    {
        $table_id = $this->table['table_id'] ?? 0;
        $table = TableModel::detail($table_id);
        if (!$table || $table['is_bind'] == 0) {
            return $this->renderError('桌台未绑定', [], -3);
        }
        // H5设置
        $tablet = SettingModel::getSupplierItem(SettingEnum::H5, $this->table['shop_supplier_id'] ?? 0, $this->table['app_id'] ?? 0);
        unset($tablet['advanced_password']);
        $table['tablet'] = $tablet;
        // 桌台当前进行中订单
        $table['order'] = OrderModel::getTableUnderwayOrder($table_id);
        return $this->renderSuccess('桌台信息', $table);
    }

    /**
     * @Apidoc\Title("桌台订单信息")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/scan/table.table/getOrderInfo")
     * @Apidoc\Param("need_unsend_list", type="int", require=false, desc="返回未送列表  0-不需要 1-需要")
     * @Apidoc\Param("need_send_list", type="int", require=false, desc="返回已经列表 0-不需要 1-需要")
     */
    public function getOrderInfo()
    {
        //
        $data = $this->postData();
        $need_unsend_list = isset($data['need_unsend_list']) ?  $data['need_unsend_list'] : 0;
        $need_send_list = isset($data['need_send_list']) ?  $data['need_send_list'] : 0;
        // Tid
        $tableId = $this->table['table_id'] ?? 0;
        /** @var OrderModel $detail */
        $detail = OrderModel::getScanTableUnderwayOrder($tableId);
        if (!$detail) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        $reData = $detail->getScanOrderInfo($need_unsend_list, $need_send_list);
        //
        return $this->renderSuccess('请求成功', $reData);
    }
}

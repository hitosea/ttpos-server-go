<?php

namespace app\scan\controller\call;

use app\common\model\store\Table;
use hg\apidoc\annotation as Apidoc;
use app\scan\controller\Controller;
use app\common\model\call\Call as CallModel;

/**
 * 呼叫相关
 */
class Call extends Controller
{
    /**
     * @Apidoc\Title("呼叫")
     * @Apidoc\Desc("呼叫")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/scan/call.call/call")
     * @Apidoc\Param("call_type", type="int", require=true, default="1", desc="呼叫类型(1服务员,2收款)")
     */
    public function call($call_type)
    {
        $table_id = $this->table['table_id'] ?? 0;
        $shop_supplier_id = $this->table['shop_supplier_id'] ?? 0;
        $table = Table::detail($table_id);
        if (!$table) {
            return $this->renderError('桌台不存在');
        }
        $tableNo = Table::where('table_id', $table_id)->value('table_no');
        if (!$tableNo) {
            return $this->renderError('桌台已关闭', [], -4);
        }
        //
        (new CallModel)->makeCall($table_id, $tableNo, $call_type, 0, $shop_supplier_id);
        // 
        return $this->renderSuccess();
    }
}

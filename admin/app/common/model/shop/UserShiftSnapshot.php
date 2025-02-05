<?php

namespace app\common\model\shop;

use app\common\model\BaseModel;
use app\common\model\order\Order;
use app\common\model\shop\UserShiftLog;

/**
 * 用户交班快照
 */
class UserShiftSnapshot extends BaseModel
{

    protected $name = 'shop_user_shift_snapshot';
    protected $pk = 'id';

    /**
     * 获取快照
     */
    public function getSnapshot($id)
    {
        $detail = $this->find($id);
        $content = $detail['content'];
        $content = json_decode($content, true);
        //
        return $content;
    }

    /**
     * 创建快照
     */
    public function createSnapshot($shiftLogId, $abnormal = [])
    {
        $model = new UserShiftLog;
        $detail = $model->detail($shiftLogId);
        // 当班订单数据统计
        $params = [
            'cashier_id' => $detail['shift_user_id'],
            'shop_supplier_id' => $detail['shop_supplier_id'],
            'date' => [$detail['shift_start_time'], $detail['shift_end_time']],
        ];
        $detail['order'] = (new Order)->storeOverview($params);
        // 销售信息
        $detail['salesInfo'] = $model->getSalesInfo($detail['shift_user_id'], $detail['shop_supplier_id'], $detail['shift_start_time'], $detail['shift_end_time']);
        // 异常信息
        if ($abnormal) {
            $detail['abnormal'] = $abnormal;
        }
        //
        $content = json_encode($detail);
        $this->save([
            'shift_log_id' => $shiftLogId,
            'content' => $content,
            'shop_supplier_id' => $detail['shop_supplier_id'],
            'app_id' => $detail['app_id'],
        ]);
    }
}

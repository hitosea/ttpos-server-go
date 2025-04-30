<?php

namespace app\common\model_old\erp;

use help\StringHelp;
use app\common\model_old\BaseModel;
use app\common\model_old\shop\User;

/**
 * 采购单操作日志模型
 */
class ErpPurchaseOperationLog extends BaseModel
{
    protected $name = 'erp_purchase_operation_log';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 操作状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库
     */
    const statusOperation = [
        10 => '添加',
        20 => '驳回',
        30 => '通过',
        40 => '已采购',
        50 => '入库',
    ];

    //
    public static function onBeforeInsert($model)
    {
        if (!isset($model['id'])) {
            $model['id'] = StringHelp::uuid();
        }
        return $model;
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_id', 'shop_user_id')->field(['shop_user_id', 'user_name', 'real_name']);
    }
}

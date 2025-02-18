<?php

namespace app\common\model\erp;

use help\StringHelp;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;

/**
 * 采购单操作日志模型
 */
class ErpPurchaseOperationLog extends BaseModel
{
    use SoftDelete;
    protected $name = 'purchase_form_log';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

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

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_uuid', 'uuid')->field(['uuid', 'username', 'real_name']);
    }
}

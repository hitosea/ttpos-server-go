<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFullReductionActivityToSaleBillTable extends Migrator
{
    /**
     * 添加满减活动抵扣金额字段到 ttpos_sale_bill 表
     */
    public function change()
    {
        $table = $this->table('sale_bill');

        // 检查字段是否已存在
        if (!$table->hasColumn('activity_amount')) {
            $table->addColumn('activity_amount', 'decimal', [
                'precision' => 22,
                'scale' => 4,
                'null' => false,
                'default' => 0,
                'comment' => '满减活动抵扣金额（所有sale_order的满减扣减金额总和）',
                'after' => 'gift_amount'
            ])->update();
        }
    }
}


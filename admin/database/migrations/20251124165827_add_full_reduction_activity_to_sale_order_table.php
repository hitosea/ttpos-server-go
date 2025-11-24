<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFullReductionActivityToSaleOrderTable extends Migrator
{
    /**
     * 添加满减活动相关字段到 ttpos_sale_order 表
     */
    public function change()
    {
        $table = $this->table('sale_order');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('activity_amount')) {
            $table->addColumn('activity_amount', 'decimal', [
                'precision' => 22,
                'scale' => 4,
                'null' => false,
                'default' => 0,
                'comment' => '满减活动抵扣金额（结账完成后记录）',
                'after' => 'coupon_amount'
            ])->update();
        }

        if (!$table->hasColumn('full_reduction_activity_uuid')) {
            $table->addColumn('full_reduction_activity_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '订单使用的满减活动UUID',
                'after' => 'activity_amount'
            ])->update();
        }

        if (!$table->hasColumn('full_reduction_activity_message')) {
            $table->addColumn('full_reduction_activity_message', 'string', [
                'limit' => 255,
                'null' => false,
                'default' => '',
                'comment' => '满减规则信息（如"满200减20"）',
                'after' => 'full_reduction_activity_uuid'
            ])->update();
        }
    }
}


<?php

use think\migration\Migrator;

class AddPlatformOrderStateToTakeoutOrderTable extends Migrator
{
    /**
     * 添加 platform_order_state 字段到外卖订单表
     * 用于存储平台原始订单状态（如 Grab 的 "FOOD_PREPARING" / "COMPLETED" 等）
     */
    public function change()
    {
        $table = $this->table('takeout_order');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('platform_order_state')) {
            $table->addColumn('platform_order_state', 'string', ['limit' => 50, 'default' => '', 'comment' => '平台订单状态原始值 (Grab: state)', 'after' => 'platform_order_id'])->update();
        }
    }
}


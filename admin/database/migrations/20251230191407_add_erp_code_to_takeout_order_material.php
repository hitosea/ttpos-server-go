<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddErpCodeToTakeoutOrderMaterial extends Migrator
{
    /**
     * 为 ttpos_takeout_order_material 表添加 erp_code 字段
     * 
     * 用途：存储原料的 ERP 编码，用于与 ERP 系统对接
     * 
     * 涉及表：
     * 1. ttpos_takeout_order_material - 添加 erp_code, takeout_order_item_uuid, takeout_order_item_modifier_uuid 字段
     * 
     * 版本兼容性：向前兼容，旧数据字段默认为空字符串或0
     */
    public function change()
    {
        $this->addErpCodeField();
    }
    
    /**
     * 为外卖订单原料表添加 ERP 编码字段
     */
    private function addErpCodeField()
    {
        if (!$this->hasTable('takeout_order_material')) {
            return;
        }
        
        $table = $this->table('takeout_order_material');
        
        if (!$table->hasColumn('erp_code')) {
            $table->addColumn('erp_code', 'string', [
                'limit' => 50,
                'null' => false,
                'default' => '',
                'comment' => 'ERP编码(来自Material.Code)',
                'after' => 'material_uuid'
            ]);
        }

        // 添加外卖订单商品UUID
        if (!$table->hasColumn('takeout_order_item_uuid')) {
            $table->addColumn('takeout_order_item_uuid', 'biginteger', [
                'limit' => 20,
                'null' => false,
                'default' => 0,
                'comment' => '外卖订单商品UUID(关联ttpos_takeout_order_item.uuid)',
                'after' => 'takeout_order_uuid'
            ]);
        }
        
        // 添加外卖订单商品修饰符UUID
        if (!$table->hasColumn('takeout_order_item_modifier_uuid')) {
            $table->addColumn('takeout_order_item_modifier_uuid', 'biginteger', [
                'limit' => 20,
                'null' => false,
                'default' => 0,
                'comment' => '外卖订单商品修饰符UUID(关联ttpos_takeout_order_item_modifier.uuid)',
                'after' => 'takeout_order_item_uuid'
            ]);
        }
        
        // 删除 staff_shift_log_uuid 字段
        if ($table->hasColumn('staff_shift_log_uuid')) {
            $table->removeColumn('staff_shift_log_uuid');
        }

        $table->update();
    }
    
}


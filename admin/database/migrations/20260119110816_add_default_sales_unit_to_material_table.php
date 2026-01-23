<?php

use think\migration\Migrator;

/**
 * 为物品表添加默认销售单位字段
 * 
 * 在 ttpos_material 表中添加 default_sales_unit 字段，用于存储 ERPNext 同步的默认销售单位ID
 * 关联 ttpos_material_unit.uuid
 */
class AddDefaultSalesUnitToMaterialTable extends Migrator
{
    /**
     * Change Method.
     *
     * 为物品表添加默认销售单位字段
     * 
     * 在 ttpos_material 表中添加 default_sales_unit 字段，用于存储 ERPNext 同步的默认销售单位
     * 关联 ttpos_material_unit对应的uom_code
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('material')) {
            return;
        }

        // 获取表对象
        $table = $this->table('material');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('default_sales_unit_uuid')) {
            // 添加 default_sales_unit 字段
            // 添加到 cost_unit_uuid 字段之后
            $table->addColumn('default_sales_unit_uuid', 'biginteger', [
                'limit' => 20,
                'null' => false,
                'default' => 0,
                'comment' => '默认销售单位ID（MaterialUnit UUID），关联 ttpos_material_unit.uuid',
                'after' => 'cost_unit_uuid'
            ]);
        }
        
        // 添加索引（用于查询优化）
        if (!$table->hasIndexByName('idx_default_sales_unit_uuid')) {
            $table->addIndex('default_sales_unit_uuid', ['name' => 'idx_default_sales_unit_uuid']);
        }
        
        $table->update();
    }
}

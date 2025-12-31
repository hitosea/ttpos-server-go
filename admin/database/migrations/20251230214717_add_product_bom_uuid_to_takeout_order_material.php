<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddProductBomUuidToTakeoutOrderMaterial extends Migrator
{
    /**
     * 添加 product_bom_uuid 字段到 ttpos_takeout_order_material 表
     * 用于记录原料消耗对应的 BOM UUID，便于准确聚合原料到具体的 BOM
     */
    public function change()
    {
        if (!$this->hasTable('takeout_order_material')) {
            return;
        }
        $table = $this->table('takeout_order_material');

        // 添加字段
        if (!$table->hasColumn('material_name')) {  
            $table->addColumn('material_name', 'text', [
                'null' => true,
                'comment' => '原料名称(来自Material.Name)',
                'after' => 'material_uuid'
            ]);
        }

        // 检查字段是否已存在
        if (!$table->hasColumn('product_bom_uuid')) {
            $table->addColumn('product_bom_uuid', 'biginteger', [
                'signed' => false,
                'default' => 0,
                'comment' => 'BOM UUID(关联ttpos_product_bom.uuid)',
                'after' => 'takeout_order_item_modifier_uuid'
            ]);
        }
        
        // 添加索引以提升查询性能
        if (!$table->hasIndex(['product_bom_uuid'])) {
            $table->addIndex(['product_bom_uuid'], ['name' => 'idx_product_bom_uuid']);
        }

        $table->update();
    }
}


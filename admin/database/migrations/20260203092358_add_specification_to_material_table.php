<?php

use think\migration\Migrator;

/**
 * 为物品表添加规格字段
 *
 * 在 ttpos_material 表中添加：
 * - specification: 规格，来自 ERPNext Item 的 Specification 字段
 */
class AddSpecificationToMaterialTable extends Migrator
{
    /**
     * Change Method.
     *
     * 为物品表添加规格字段
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('material')) {
            return;
        }

        // 获取表对象
        $table = $this->table('material');

        // 添加 specification 字段 - 规格
        if (!$table->hasColumn('specification')) {
            $table->addColumn('specification', 'string', [
                'limit' => 500,
                'null' => false,
                'default' => '',
                'comment' => '规格，来自 ERPNext Item 的 Specification 字段',
                'after' => 'supplier_erp_code'
            ]);
        }

        $table->update();
    }
}

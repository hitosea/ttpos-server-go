<?php

use think\migration\Migrator;


class CreateWarehouseMonthlyMaterialFormTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        if (!$this->hasTable('warehouse_monthly_material_form')) {
            $table = $this->table('warehouse_monthly_material_form', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '月度报表表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '月度报表ID']);
            $table->addColumn('year', 'integer', ['default' => 0, 'comment' => '年']);
            $table->addColumn('month', 'integer', ['default' => 0, 'comment' => '月']);
            $table->addColumn('scene', 'integer', ['default' => 0, 'comment' => '记录类型,0-月初 1-月末']);
            $table->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物料ID']);
            $table->addColumn('stock', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => 0.0000, 'comment' => '库存']);
            $table->addColumn('create_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)']);
            $table->addColumn('update_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)']);
            $table->addColumn('delete_time', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)']);
            $table->addIndex('uuid', ['unique' => true]);
            $table->create();
        }
    }
}

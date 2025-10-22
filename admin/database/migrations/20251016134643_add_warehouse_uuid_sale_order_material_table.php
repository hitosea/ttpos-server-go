<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseUuidSaleOrderMaterialTable extends Migrator
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
        // 创建销售订单原料表，包含以上字段
        if ($this->hasTable('sale_order_material')) {
            $table = $this->table('sale_order_material');
          if (!$table->hasColumn('warehouse_uuid')) {
            $table->addColumn('warehouse_uuid', 'biginteger', ['default' => 0, 'comment' => '仓库ID', 'after' => 'material_uuid'])
                ->update();
          }
          if (!$table->hasColumn('is_summarized')) {
            $table->addColumn('is_summarized', 'integer', ['default' => 0, 'comment' => '是否已经统计,0-未统计 1-已统计', 'after' => 'warehouse_uuid'])
                ->update();
          }
        }
    }
}

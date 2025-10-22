<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddOpeningHourToWarehouseInOutLogTable extends Migrator
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
        // 检查表是否存在
        if ($this->hasTable('warehouse_in_out_log')) {
            $table = $this->table('warehouse_in_out_log');
            
            // 检查opening_hours字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('opening_hours')) {
                $table->addColumn('opening_hours', 'string', ['limit' => 255, 'default' => '', 'comment' => '营业时段,仅用于Scene销售出库的场景', 'after' => 'other_org_name']);
            }
            
            $table->update();
        }
    }
}

<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddShiftUuidToWarehouseOutFormItemTable extends Migrator
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
        if ($this->hasTable('warehouse_out_form_item')) {
            $table = $this->table('warehouse_out_form_item');
            
            // 检查staff_shift_log_uuid字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('staff_shift_log_uuid')) {
                $table->addColumn('staff_shift_log_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工交班记录ID', 'after' => 'package_uuid']);
            }
            $table->update();
        }
    }
}

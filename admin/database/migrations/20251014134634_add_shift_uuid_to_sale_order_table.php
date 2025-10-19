<?php
/**
 * 添加对方机构相关字段到仓库出入库记录表
 */

use think\migration\Migrator;

class AddShiftUuidToSaleOrderTable extends Migrator
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
        if ($this->hasTable('sale_order')) {
            $table = $this->table('sale_order');
            
            // 检查staff_shift_log_uuid字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('staff_shift_log_uuid')) {
                $table->addColumn('staff_shift_log_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工交班记录ID', 'after' => 'sale_bill_uuid']);
            }
            $table->update();
        }

          // 检查表是否存在
          if ($this->hasTable('refund_order')) {
            $table = $this->table('refund_order');
            
            // 检查staff_shift_log_uuid字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('staff_shift_log_uuid')) {
                $table->addColumn('staff_shift_log_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工交班记录ID', 'after' => 'erp_invoice_name']);
            }
            $table->update();
        }

             // 检查表是否存在
             if ($this->hasTable('return_order')) {
                $table = $this->table('return_order');
                
                // 检查staff_shift_log_uuid字段是否不存在，如果不存在则添加
                if (!$table->hasColumn('staff_shift_log_uuid')) {
                    $table->addColumn('staff_shift_log_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '员工交班记录ID', 'after' => 'duty_no']);
                }
                $table->update();
            }
    }
}

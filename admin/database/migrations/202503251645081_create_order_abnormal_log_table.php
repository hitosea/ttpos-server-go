<?php

use think\migration\Migrator;
use Phinx\Db\Adapter\MysqlAdapter;

class CreateOrderAbnormalLogTable extends Migrator
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

        if ($this->hasTable('sale_bill_operation_record')) {
            // 改表名为 sale_order_operation_record
            $table = $this->table('sale_bill_operation_record', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '桌台账单操作记录']);
            $table->rename('sale_order_operation_record');
            $table->save();
        }

        if (!$this->hasTable('sale_order_abnormal_log')) {
            $table = $this->table('sale_order_abnormal_log', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '订单异常日志表']);
            $table->addColumn('sale_bill_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售账单ID']);
            $table->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '销售订单ID']);
            $table->addColumn('duty_no', 'string', ['limit' => 64, 'default' => '', 'comment' => '当班编号']);
            $table->addColumn('source', 'string', ['limit' => 150, 'null' => false, 'default' => '', 'comment' => '来源 order-订单 recharge-充值']);
            $table->addColumn('action', 'string', ['limit' => 150, 'null' => false, 'default' => '', 'comment' => '行为']);
            $table->addColumn('sub_action', 'string', ['limit' => 150, 'null' => false, 'default' => '', 'comment' => '自定义子行为']);
            $table->addColumn('sign', 'string', ['limit' => 255,'default' => '', 'comment' => '操作签名']);
            $table->addColumn('remark', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '备注']);
            $table->addColumn('cashier_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收银员ID']);
            $table->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间']);
            $table->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间']);
            $table->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间']);
            $table->create();
        }
    }
}

<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateStatisticsPaymentTable extends Migrator
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
        if (!$this->hasTable('statistics_payment')) {
            $table = $this->table('statistics_payment', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '支付统计表'
            ]);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => 'uuid'])
                ->addColumn('sale_bill_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售账单uuid'])
                ->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售订单uuid'])
                ->addColumn('duty_no', 'string', ['null' => false, 'default' => '', 'comment' => '当班编号'])
                ->addColumn('desk_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '桌台uuid'])
                ->addColumn('payment_method_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '支付方式uuid'])
                ->addColumn('payment_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '支付金额'])
                ->addColumn('refund_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款金额'])
                ->addColumn('complete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '完成时间'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
}

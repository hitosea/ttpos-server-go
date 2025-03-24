<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateSaleOrderPeakTimeTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     */
    public function change()
    {
        if (!$this->hasTable('sale_order_peak_time')) {
            $table = $this->table('sale_order_peak_time', [
                'id' => false,
                'primary_key' => 'id',
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '订单高峰时间表'
            ]);
            $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => 'UUID'])
                ->addColumn('date', 'integer', ['default' => 0, 'comment' => '日期（天）'])
                ->addColumn('hour', 'integer', ['default' => 0, 'comment' => '小时'])
                ->addColumn('num', 'integer', ['default' => 0, 'comment' => '订单数'])
                ->addColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0.00, 'comment' => '订单金额'])
                ->addColumn('cashier_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '收银员ID'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addIndex(['cashier_uuid'], ['name' => 'idx_cashier_uuid'])
                ->addIndex(['uuid'], ['name' => 'unique_uuid', 'unique' => true])
                ->create();
        }
    }
}
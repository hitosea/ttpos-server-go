<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateStatisticsDelay extends Migrator
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
        if (!$this->hasTable('statistics_delay')) {
            $table = $this->table('statistics_delay', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '加钟统计表'
            ]);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => 'uuid'])
                ->addColumn('sale_bill_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售账单uuid'])
                ->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售订单uuid'])
                ->addColumn('duty_no', 'string', ['null' => false, 'default' => '', 'comment' => '当班编号'])
                ->addColumn('desk_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '桌台uuid'])
                ->addColumn('buffet_delay_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '自助餐加钟价格ID'])
                ->addColumn('product_price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '价格,下单时固定不受后台改变，结账时再检查是否改变'])
                ->addColumn('product_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '商品数量'])
                ->addColumn('tax_rate', 'decimal', ['precision' => 14, 'scale' => 4, 'null' => false, 'default' => 0.00, 'comment' => '税率'])
                ->addColumn('tax_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '税费'])
                ->addColumn('service_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务费'])
                ->addColumn('service_tax', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务税费'])
                ->addColumn('give_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '赠菜数量'])
                ->addColumn('free_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '免单数量'])
                ->addColumn('refund_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '退款数量'])
                ->addColumn('complete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '完成时间'])
                ->addColumn('refund_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '退款时间'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true])
                ->addIndex(['sale_bill_uuid'])
                ->addIndex(['duty_no'])
                ->addIndex(['desk_uuid'])
                ->addIndex(['complete_time'])
                ->addIndex(['refund_time'])
                ->create();
        }
    }
}

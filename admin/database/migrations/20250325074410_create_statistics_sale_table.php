<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateStatisticsSaleTable extends Migrator
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
        if (!$this->hasTable('statistics_sale')) {
            $table = $this->table('statistics_sale', [
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '销售统计表'
            ]);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => 'uuid'])
                ->addColumn('sale_bill_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售账单uuid'])
                ->addColumn('sale_order_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '销售订单uuid'])
                ->addColumn('duty_no', 'string', ['null' => false, 'default' => '', 'comment' => '当班编号'])
                ->addColumn('desk_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '桌台uuid'])
                ->addColumn('meal_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '用餐人数'])
                ->addColumn('product_price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '商品原价: 不含税'])
                ->addColumn('product_sale_price', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '商品销售价'])
                ->addColumn('product_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '商品数量'])
                ->addColumn('product_tax', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '商品税'])
                ->addColumn('service_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务费'])
                ->addColumn('service_tax', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务税'])
                ->addColumn('discount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '优惠折扣'])
                ->addColumn('discount_member', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '会员折扣'])
                ->addColumn('gift_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '赠菜金额'])
                ->addColumn('gift_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '赠菜数量'])
                ->addColumn('free_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '免单金额'])
                ->addColumn('free_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '免单数量'])
                ->addColumn('payment_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '支付金额'])
                ->addColumn('payment_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '支付手续费'])
                ->addColumn('payment_balance', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '支付余额'])
                ->addColumn('refund_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款金额'])
                ->addColumn('complete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '完成时间'])
                ->addColumn('refund_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '退款时间'])
                ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
}

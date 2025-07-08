<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateNumFields extends Migrator
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
        $table = $this->table('statistics_customer_type');
        if ($table->hasColumn('product_num')) {
            $table->changeColumn('product_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '商品数量', 'after' => 'product_sale_price']);
        }
        if ($table->hasColumn('give_num')) {
            $table->changeColumn('give_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '赠菜数量', 'after' => 'service_tax']);
        }
        if ($table->hasColumn('free_num')) {
            $table->changeColumn('free_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '免单数量', 'after' => 'give_num']);
        }
        if ($table->hasColumn('refund_num')) {
            $table->changeColumn('refund_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款数量', 'after' => 'free_num']);
        }
        $table->update();

        $table = $this->table('statistics_delay');
        if ($table->hasColumn('product_num')) {
            $table->changeColumn('product_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '商品数量', 'after' => 'product_price']);
        }
        if ($table->hasColumn('give_num')) {
            $table->changeColumn('give_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '赠菜数量', 'after' => 'service_tax']);
        }
        if ($table->hasColumn('free_num')) {
            $table->changeColumn('free_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '免单数量', 'after' => 'give_num']);
        }
        if ($table->hasColumn('refund_num')) {
            $table->changeColumn('refund_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款数量', 'after' => 'free_num']);
        }
        $table->update();
    }
}

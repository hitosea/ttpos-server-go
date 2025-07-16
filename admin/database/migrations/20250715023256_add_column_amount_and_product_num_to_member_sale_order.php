<?php

use think\facade\Db;
use think\migration\Migrator;

class AddColumnAmountAndProductNumToMemberSaleOrder extends Migrator
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
        // 添加字段
        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('serial_number')) {
            $table->addColumn('serial_number', 'varchar', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '订单流水号', 'after' => 'status']);
            $table->update();
        }
        if (!$table->hasColumn('cancel_reason')) {
            $table->addColumn('cancel_reason', 'varchar', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '取消原因', 'after' => 'remark']);
            $table->update();
        }
        if (!$table->hasColumn('product_num')) {
            $table->addColumn('product_num', 'decimal', ['limit' => 12, 'precision' => 2, 'null' => false, 'default' => 0, 'comment' => '商品数量.订单中商品的总数量，商品A数量2，商品B数量1，则商品数量为3', 'after' => 'payment_method_uuid']);
            $table->update();
        }
        if (!$table->hasColumn('product_amount')) {
            $table->addColumn('product_amount', 'decimal', ['limit' => 12, 'precision' => 2, 'null' => false, 'default' => 0, 'comment' => '商品金额', 'after' => 'product_num']);
            $table->update();
        }
        if (!$table->hasColumn('member_discount_fee')) {
            $table->addColumn('member_discount_fee', 'decimal', ['limit' => 12, 'precision' => 2, 'null' => false, 'default' => 0, 'comment' => '会员折扣', 'after' => 'product_amount']);
            $table->update();
        }
        if (!$table->hasColumn('amount')) {
            $table->addColumn('amount', 'decimal', ['limit' => 12, 'precision' => 2, 'null' => false, 'default' => 0, 'comment' => '订单总金额.商品金额-会员折扣+配送费', 'after' => 'member_discount_fee']);
            $table->update();
        }

        // 添加字段
         $table = $this->table('member_sale_order_address');
         if (!$table->hasColumn('phone_prefix')) {
             $table->addColumn('phone_prefix', 'varchar', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系电话前缀', 'after' => 'contact_phone']);
             $table->update();
         }
    }
}
